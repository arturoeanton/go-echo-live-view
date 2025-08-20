package liveview

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewEventRegistry(t *testing.T) {
	er := NewEventRegistry(nil)
	
	if er == nil {
		t.Fatal("NewEventRegistry returned nil")
	}
	
	if er.config == nil {
		t.Error("Expected default config to be set")
	}
	
	if len(er.handlers) != 0 {
		t.Error("Expected empty handlers map")
	}
	
	if er.metrics == nil {
		t.Error("Expected metrics to be initialized with default config")
	}
}

func TestEventRegistryOn(t *testing.T) {
	er := NewEventRegistry(nil)
	
	called := false
	handler := func(ctx context.Context, event *Event) error {
		called = true
		return nil
	}
	
	// Register handler
	id, err := er.On("test.event", handler)
	if err != nil {
		t.Errorf("On() error = %v", err)
	}
	
	if id == "" {
		t.Error("Expected non-empty handler ID")
	}
	
	// Emit event
	er.Emit("test.event", nil)
	
	if !called {
		t.Error("Handler was not called")
	}
}

func TestEventRegistryOnce(t *testing.T) {
	er := NewEventRegistry(nil)
	
	callCount := 0
	handler := func(ctx context.Context, event *Event) error {
		callCount++
		return nil
	}
	
	// Register once handler
	_, err := er.Once("test.event", handler)
	if err != nil {
		t.Errorf("Once() error = %v", err)
	}
	
	// Emit event multiple times
	er.Emit("test.event", nil)
	er.Emit("test.event", nil)
	er.Emit("test.event", nil)
	
	if callCount != 1 {
		t.Errorf("Expected handler to be called once, called %d times", callCount)
	}
}

func TestEventRegistryOff(t *testing.T) {
	er := NewEventRegistry(nil)
	
	called := false
	handler := func(ctx context.Context, event *Event) error {
		called = true
		return nil
	}
	
	// Register handler
	id, _ := er.On("test.event", handler)
	
	// Remove handler
	err := er.Off(id)
	if err != nil {
		t.Errorf("Off() error = %v", err)
	}
	
	// Emit event
	er.Emit("test.event", nil)
	
	if called {
		t.Error("Handler was called after removal")
	}
	
	// Try to remove non-existent handler
	err = er.Off("non-existent")
	if err == nil {
		t.Error("Expected error when removing non-existent handler")
	}
}

func TestEventRegistryOffAll(t *testing.T) {
	er := NewEventRegistry(nil)
	
	callCount := 0
	handler := func(ctx context.Context, event *Event) error {
		callCount++
		return nil
	}
	
	// Register multiple handlers
	er.On("test.event", handler)
	er.On("test.event", handler)
	er.On("test.event", handler)
	
	// Remove all handlers
	er.OffAll("test.event")
	
	// Emit event
	er.Emit("test.event", nil)
	
	if callCount != 0 {
		t.Errorf("Handlers were called after OffAll, count: %d", callCount)
	}
}

func TestEventRegistryPriority(t *testing.T) {
	er := NewEventRegistry(nil)
	
	var order []int
	
	// Register handlers with different priorities
	er.On("test.event", func(ctx context.Context, event *Event) error {
		order = append(order, 1)
		return nil
	}, WithPriority(1))
	
	er.On("test.event", func(ctx context.Context, event *Event) error {
		order = append(order, 3)
		return nil
	}, WithPriority(3))
	
	er.On("test.event", func(ctx context.Context, event *Event) error {
		order = append(order, 2)
		return nil
	}, WithPriority(2))
	
	// Emit event
	er.Emit("test.event", nil)
	
	// Check execution order (highest priority first)
	expected := []int{3, 2, 1}
	if len(order) != len(expected) {
		t.Fatalf("Expected %d handlers, got %d", len(expected), len(order))
	}
	
	for i, v := range expected {
		if order[i] != v {
			t.Errorf("Expected order[%d] = %d, got %d", i, v, order[i])
		}
	}
}

func TestEventRegistryFilter(t *testing.T) {
	er := NewEventRegistry(nil)
	
	called := false
	handler := func(ctx context.Context, event *Event) error {
		called = true
		return nil
	}
	
	// Register handler with filter
	filter := func(event *Event) bool {
		if event.Data == nil {
			return false
		}
		value, ok := event.Data["allow"]
		return ok && value == true
	}
	
	er.On("test.event", handler, WithFilter(filter))
	
	// Emit event that doesn't pass filter
	er.Emit("test.event", map[string]interface{}{"allow": false})
	
	if called {
		t.Error("Handler was called when filter should have blocked it")
	}
	
	// Emit event that passes filter
	called = false
	er.Emit("test.event", map[string]interface{}{"allow": true})
	
	if !called {
		t.Error("Handler was not called when filter should have passed")
	}
}

func TestEventRegistryWildcards(t *testing.T) {
	er := NewEventRegistry(nil)
	
	var capturedEvents []string
	handler := func(ctx context.Context, event *Event) error {
		capturedEvents = append(capturedEvents, event.Type)
		return nil
	}
	
	// Register wildcard handlers
	er.On("user.*", handler)
	er.On("*.created", handler)
	er.On("*", handler) // Global handler
	
	// Emit various events
	er.Emit("user.login", nil)
	er.Emit("user.logout", nil)
	er.Emit("post.created", nil)
	er.Emit("comment.updated", nil)
	
	// Check captured events - we expect at least 7:
	// - 2 from user.* (user.login, user.logout)
	// - 1 from *.created (post.created)  
	// - 4 from * (all events)
	expectedCount := 7
	if len(capturedEvents) < expectedCount {
		t.Errorf("Expected at least %d captures, got %d", expectedCount, len(capturedEvents))
	}
}

func TestEventRegistryMiddleware(t *testing.T) {
	er := NewEventRegistry(nil)
	
	var executionOrder []string
	
	// Add middleware
	middleware1 := func(next EventHandler) EventHandler {
		return func(ctx context.Context, event *Event) error {
			executionOrder = append(executionOrder, "middleware1-before")
			err := next(ctx, event)
			executionOrder = append(executionOrder, "middleware1-after")
			return err
		}
	}
	
	middleware2 := func(next EventHandler) EventHandler {
		return func(ctx context.Context, event *Event) error {
			executionOrder = append(executionOrder, "middleware2-before")
			err := next(ctx, event)
			executionOrder = append(executionOrder, "middleware2-after")
			return err
		}
	}
	
	er.Use(middleware1)
	er.Use(middleware2)
	
	// Register handler
	er.On("test.event", func(ctx context.Context, event *Event) error {
		executionOrder = append(executionOrder, "handler")
		return nil
	})
	
	// Emit event
	er.Emit("test.event", nil)
	
	// Check execution order
	expected := []string{
		"middleware1-before",
		"middleware2-before",
		"handler",
		"middleware2-after",
		"middleware1-after",
	}
	
	if len(executionOrder) != len(expected) {
		t.Fatalf("Expected %d executions, got %d", len(expected), len(executionOrder))
	}
	
	for i, exp := range expected {
		if executionOrder[i] != exp {
			t.Errorf("Execution[%d]: expected %s, got %s", i, exp, executionOrder[i])
		}
	}
}

func TestEventRegistryStopPropagation(t *testing.T) {
	er := NewEventRegistry(nil)
	
	var handlersExecuted []int
	
	// Register multiple handlers
	er.On("test.event", func(ctx context.Context, event *Event) error {
		handlersExecuted = append(handlersExecuted, 1)
		er.StopPropagation(event.ID)
		return nil
	}, WithPriority(3))
	
	er.On("test.event", func(ctx context.Context, event *Event) error {
		handlersExecuted = append(handlersExecuted, 2)
		return nil
	}, WithPriority(2))
	
	er.On("test.event", func(ctx context.Context, event *Event) error {
		handlersExecuted = append(handlersExecuted, 3)
		return nil
	}, WithPriority(1))
	
	// Emit event
	er.Emit("test.event", nil)
	
	// Only first handler should execute
	if len(handlersExecuted) != 1 {
		t.Errorf("Expected 1 handler to execute, got %d", len(handlersExecuted))
	}
	
	if handlersExecuted[0] != 1 {
		t.Errorf("Expected handler 1 to execute, got %d", handlersExecuted[0])
	}
}

func TestEventRegistryMetrics(t *testing.T) {
	er := NewEventRegistry(nil)
	
	// Register handlers
	er.On("success.event", func(ctx context.Context, event *Event) error {
		return nil
	})
	
	er.On("error.event", func(ctx context.Context, event *Event) error {
		return errors.New("test error")
	})
	
	// Emit events
	er.Emit("success.event", nil)
	er.Emit("success.event", nil)
	er.Emit("error.event", nil)
	
	// Get metrics
	metrics := er.GetMetrics()
	
	if metrics == nil {
		t.Fatal("Metrics should not be nil")
	}
	
	if metrics.totalEvents != 3 {
		t.Errorf("Expected 3 total events, got %d", metrics.totalEvents)
	}
	
	if metrics.handledEvents != 3 {
		t.Errorf("Expected 3 handled events, got %d", metrics.handledEvents)
	}
}

func TestEventRegistryMaxHandlers(t *testing.T) {
	config := &EventRegistryConfig{
		MaxHandlersPerEvent: 2,
		EnableMetrics:       false,
	}
	er := NewEventRegistry(config)
	
	handler := func(ctx context.Context, event *Event) error {
		return nil
	}
	
	// Register up to limit
	_, err := er.On("test.event", handler)
	if err != nil {
		t.Errorf("First handler registration failed: %v", err)
	}
	
	_, err = er.On("test.event", handler)
	if err != nil {
		t.Errorf("Second handler registration failed: %v", err)
	}
	
	// Try to exceed limit
	_, err = er.On("test.event", handler)
	if err == nil {
		t.Error("Expected error when exceeding max handlers limit")
	}
}

func TestGlobalEventBus(t *testing.T) {
	called := false
	handler := func(ctx context.Context, event *Event) error {
		called = true
		return nil
	}
	
	// Register on global bus
	id, err := GlobalEventBus.On("global.event", handler)
	if err != nil {
		t.Errorf("Global On() error = %v", err)
	}
	
	// Emit on global bus
	GlobalEventBus.Emit("global.event", nil)
	
	if !called {
		t.Error("Global handler was not called")
	}
	
	// Clean up
	GlobalEventBus.Off(id)
}

func TestDelegatingEventHandler(t *testing.T) {
	type TestTarget struct {
		called     bool
		eventData  *Event
		contextVal context.Context
	}
	
	target := &TestTarget{}
	
	// Method with just event
	handler := DelegatingEventHandler(target, "HandleEvent")
	
	// This will fail since method doesn't exist
	err := handler(context.Background(), &Event{Type: "test"})
	if err == nil {
		t.Error("Expected error for non-existent method")
	}
}

func TestChainEventHandlers(t *testing.T) {
	var order []int
	
	handler1 := func(ctx context.Context, event *Event) error {
		order = append(order, 1)
		return nil
	}
	
	handler2 := func(ctx context.Context, event *Event) error {
		order = append(order, 2)
		return nil
	}
	
	handler3 := func(ctx context.Context, event *Event) error {
		order = append(order, 3)
		return nil
	}
	
	chained := ChainEventHandlers(handler1, handler2, handler3)
	
	err := chained(context.Background(), &Event{})
	if err != nil {
		t.Errorf("ChainEventHandlers error = %v", err)
	}
	
	expected := []int{1, 2, 3}
	for i, v := range expected {
		if order[i] != v {
			t.Errorf("Expected order[%d] = %d, got %d", i, v, order[i])
		}
	}
}

func TestChainEventHandlersWithError(t *testing.T) {
	var order []int
	
	handler1 := func(ctx context.Context, event *Event) error {
		order = append(order, 1)
		return nil
	}
	
	handler2 := func(ctx context.Context, event *Event) error {
		order = append(order, 2)
		return errors.New("stop here")
	}
	
	handler3 := func(ctx context.Context, event *Event) error {
		order = append(order, 3)
		return nil
	}
	
	chained := ChainEventHandlers(handler1, handler2, handler3)
	
	err := chained(context.Background(), &Event{})
	if err == nil {
		t.Error("Expected error from chain")
	}
	
	// Only first two should execute
	if len(order) != 2 {
		t.Errorf("Expected 2 handlers to execute, got %d", len(order))
	}
}

func TestConditionalEventHandler(t *testing.T) {
	called := false
	handler := func(ctx context.Context, event *Event) error {
		called = true
		return nil
	}
	
	// Condition that checks event type
	condition := func(event *Event) bool {
		return event.Type == "allowed"
	}
	
	conditional := ConditionalEventHandler(condition, handler)
	
	// Test with non-matching condition
	conditional(context.Background(), &Event{Type: "denied"})
	if called {
		t.Error("Handler called when condition was false")
	}
	
	// Test with matching condition
	conditional(context.Background(), &Event{Type: "allowed"})
	if !called {
		t.Error("Handler not called when condition was true")
	}
}

func TestThrottledEventHandler(t *testing.T) {
	callCount := 0
	handler := func(ctx context.Context, event *Event) error {
		callCount++
		return nil
	}
	
	throttled := ThrottledEventHandler(handler, 50*time.Millisecond)
	
	// Call multiple times quickly
	for i := 0; i < 5; i++ {
		throttled(context.Background(), &Event{})
		time.Sleep(10 * time.Millisecond)
	}
	
	// Should only be called once due to throttling
	if callCount != 1 {
		t.Errorf("Expected 1 call due to throttling, got %d", callCount)
	}
	
	// Wait for throttle period to pass
	time.Sleep(60 * time.Millisecond)
	
	// Should be able to call again
	throttled(context.Background(), &Event{})
	
	if callCount != 2 {
		t.Errorf("Expected 2 calls after throttle period, got %d", callCount)
	}
}

func TestDebouncedEventHandler(t *testing.T) {
	var callCount int32
	handler := func(ctx context.Context, event *Event) error {
		atomic.AddInt32(&callCount, 1)
		return nil
	}
	
	debounced := DebouncedEventHandler(handler, 50*time.Millisecond)
	
	// Call multiple times quickly
	for i := 0; i < 5; i++ {
		debounced(context.Background(), &Event{})
		time.Sleep(10 * time.Millisecond)
	}
	
	// Handler should not have been called yet
	if atomic.LoadInt32(&callCount) != 0 {
		t.Errorf("Handler called before debounce period, count: %d", callCount)
	}
	
	// Wait for debounce period
	time.Sleep(100 * time.Millisecond)
	
	// Should be called once after debounce
	if atomic.LoadInt32(&callCount) != 1 {
		t.Errorf("Expected 1 call after debounce, got %d", callCount)
	}
}

func TestEventRegistryClear(t *testing.T) {
	er := NewEventRegistry(nil)
	
	// Add handlers and middleware
	er.On("event1", func(ctx context.Context, event *Event) error { return nil })
	er.On("event2", func(ctx context.Context, event *Event) error { return nil })
	er.Use(func(next EventHandler) EventHandler { return next })
	
	// Clear registry
	er.Clear()
	
	// Check everything is cleared
	if len(er.handlers) != 0 {
		t.Error("Handlers not cleared")
	}
	
	if len(er.middleware) != 0 {
		t.Error("Middleware not cleared")
	}
	
	if len(er.globalHandlers) != 0 {
		t.Error("Global handlers not cleared")
	}
}

func TestConcurrentEventOperations(t *testing.T) {
	er := NewEventRegistry(nil)
	
	var wg sync.WaitGroup
	var callCount int32
	
	// Register handlers concurrently
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			eventType := fmt.Sprintf("event%d", id%10)
			er.On(eventType, func(ctx context.Context, event *Event) error {
				atomic.AddInt32(&callCount, 1)
				return nil
			})
		}(i)
	}
	
	wg.Wait()
	
	// Emit events concurrently
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			eventType := fmt.Sprintf("event%d", id%10)
			er.Emit(eventType, nil)
		}(i)
	}
	
	wg.Wait()
	
	// Should have received many calls
	if atomic.LoadInt32(&callCount) == 0 {
		t.Error("No handlers were called")
	}
}

func TestEventWithNamespace(t *testing.T) {
	er := NewEventRegistry(nil)
	
	var capturedNamespaces []string
	
	// Register handlers with namespaces
	er.On("test.event", func(ctx context.Context, event *Event) error {
		capturedNamespaces = append(capturedNamespaces, "app")
		return nil
	}, WithNamespace("app"))
	
	er.On("test.event", func(ctx context.Context, event *Event) error {
		capturedNamespaces = append(capturedNamespaces, "module")
		return nil
	}, WithNamespace("module"))
	
	// Emit event
	er.Emit("test.event", nil)
	
	// Both handlers should execute
	if len(capturedNamespaces) != 2 {
		t.Errorf("Expected 2 namespaces, got %d", len(capturedNamespaces))
	}
}