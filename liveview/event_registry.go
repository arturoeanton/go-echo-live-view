package liveview

import (
	"context"
	"fmt"
	"math/rand"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"time"
)

// EventHandler represents a function that handles events
type EventHandler func(ctx context.Context, event *Event) error

// Event represents an event with data and metadata
type Event struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Target    string                 `json:"target"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
	Source    string                 `json:"source"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// EventRegistry manages event handlers with advanced features
type EventRegistry struct {
	mu              sync.RWMutex
	handlers        map[string][]HandlerEntry
	globalHandlers  []HandlerEntry
	middleware      []EventMiddleware
	config          *EventRegistryConfig
	metrics         *EventMetrics
	preventDefaults map[string]bool
	stopPropagation map[string]bool
}

// HandlerEntry represents a registered event handler
type HandlerEntry struct {
	ID          string
	Handler     EventHandler
	Priority    int
	Once        bool
	Filter      EventFilter
	Namespace   string
	Description string
	executed    bool
	mu          sync.Mutex
}

// EventFilter filters events before handling
type EventFilter func(event *Event) bool

// EventMiddleware processes events before handlers
type EventMiddleware func(next EventHandler) EventHandler

// EventRegistryConfig configures the event registry
type EventRegistryConfig struct {
	MaxHandlersPerEvent int
	EnableMetrics       bool
	EnableWildcards     bool
	DefaultTimeout      time.Duration
	MaxEventDataSize    int
	EnableNamespaces    bool
}

// EventMetrics tracks event handling metrics
type EventMetrics struct {
	mu            sync.RWMutex
	totalEvents   int64
	handledEvents int64
	failedEvents  int64
	eventCounts   map[string]int64
	avgDuration   time.Duration
	lastError     error
	lastErrorTime time.Time
}

// DefaultEventRegistryConfig returns default configuration
func DefaultEventRegistryConfig() *EventRegistryConfig {
	return &EventRegistryConfig{
		MaxHandlersPerEvent: 100,
		EnableMetrics:       true,
		EnableWildcards:     true,
		DefaultTimeout:      5 * time.Second,
		MaxEventDataSize:    1024 * 1024, // 1MB
		EnableNamespaces:    true,
	}
}

// NewEventRegistry creates a new event registry
func NewEventRegistry(config *EventRegistryConfig) *EventRegistry {
	if config == nil {
		config = DefaultEventRegistryConfig()
	}

	er := &EventRegistry{
		handlers:        make(map[string][]HandlerEntry),
		globalHandlers:  make([]HandlerEntry, 0),
		middleware:      make([]EventMiddleware, 0),
		config:          config,
		preventDefaults: make(map[string]bool),
		stopPropagation: make(map[string]bool),
	}

	if config.EnableMetrics {
		er.metrics = &EventMetrics{
			eventCounts: make(map[string]int64),
		}
	}

	return er
}

// On registers an event handler
func (er *EventRegistry) On(eventType string, handler EventHandler, options ...HandlerOption) (string, error) {
	er.mu.Lock()
	defer er.mu.Unlock()

	// Apply options
	entry := HandlerEntry{
		ID:      generateHandlerID(),
		Handler: handler,
	}

	for _, opt := range options {
		opt(&entry)
	}

	// Check max handlers limit
	if len(er.handlers[eventType]) >= er.config.MaxHandlersPerEvent {
		return "", fmt.Errorf("max handlers limit reached for event %s", eventType)
	}

	// Add to handlers
	if eventType == "*" {
		er.globalHandlers = append(er.globalHandlers, entry)
	} else {
		if er.handlers[eventType] == nil {
			er.handlers[eventType] = make([]HandlerEntry, 0)
		}
		er.handlers[eventType] = append(er.handlers[eventType], entry)
		
		// Sort by priority
		er.sortHandlersByPriority(eventType)
	}

	Debug("Event handler %s registered for %s", entry.ID, eventType)
	return entry.ID, nil
}

// Once registers a one-time event handler
func (er *EventRegistry) Once(eventType string, handler EventHandler, options ...HandlerOption) (string, error) {
	options = append(options, WithOnce())
	return er.On(eventType, handler, options...)
}

// Off removes an event handler
func (er *EventRegistry) Off(handlerID string) error {
	er.mu.Lock()
	defer er.mu.Unlock()

	found := false

	// Remove from specific event handlers
	for eventType, handlers := range er.handlers {
		for i, entry := range handlers {
			if entry.ID == handlerID {
				er.handlers[eventType] = append(handlers[:i], handlers[i+1:]...)
				found = true
				break
			}
		}
	}

	// Remove from global handlers
	for i, entry := range er.globalHandlers {
		if entry.ID == handlerID {
			er.globalHandlers = append(er.globalHandlers[:i], er.globalHandlers[i+1:]...)
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("handler %s not found", handlerID)
	}

	Debug("Event handler %s removed", handlerID)
	return nil
}

// OffAll removes all handlers for an event type
func (er *EventRegistry) OffAll(eventType string) {
	er.mu.Lock()
	defer er.mu.Unlock()

	if eventType == "*" {
		er.globalHandlers = make([]HandlerEntry, 0)
	} else {
		delete(er.handlers, eventType)
	}

	Debug("All handlers removed for %s", eventType)
}

// Emit triggers an event
func (er *EventRegistry) Emit(eventType string, data map[string]interface{}) error {
	event := &Event{
		ID:        generateEventID(),
		Type:      eventType,
		Data:      data,
		Timestamp: time.Now(),
		Metadata:  make(map[string]interface{}),
	}

	return er.EmitEvent(event)
}

// EmitEvent triggers a pre-built event
func (er *EventRegistry) EmitEvent(event *Event) error {
	if event == nil {
		return fmt.Errorf("event cannot be nil")
	}

	// Update metrics
	if er.metrics != nil {
		er.updateMetrics(event.Type, true)
	}

	// Get handlers
	handlers := er.getHandlers(event.Type)

	// Apply middleware
	for _, handler := range handlers {
		finalHandler := handler.Handler
		
		// Apply middleware in reverse order
		for j := len(er.middleware) - 1; j >= 0; j-- {
			finalHandler = er.middleware[j](finalHandler)
		}

		// Check if propagation was stopped
		if er.isPropagationStopped(event.ID) {
			break
		}

		// Apply filter
		if handler.Filter != nil && !handler.Filter(event) {
			continue
		}

		// Execute handler with timeout
		ctx, cancel := context.WithTimeout(context.Background(), er.config.DefaultTimeout)
		defer cancel()

		err := er.executeHandler(ctx, handler, event, finalHandler)
		if err != nil {
			if er.metrics != nil {
				er.metrics.mu.Lock()
				er.metrics.failedEvents++
				er.metrics.lastError = err
				er.metrics.lastErrorTime = time.Now()
				er.metrics.mu.Unlock()
			}
			Debug("Event handler error for %s: %v", event.Type, err)
		}
	}

	// Clean up propagation flags
	er.cleanupEventFlags(event.ID)

	return nil
}

// executeHandler executes a single handler
func (er *EventRegistry) executeHandler(ctx context.Context, entry *HandlerEntry, event *Event, handler EventHandler) error {
	// Check if once and already executed
	entry.mu.Lock()
	if entry.Once && entry.executed {
		entry.mu.Unlock()
		return nil
	}
	if entry.Once {
		entry.executed = true
	}
	entry.mu.Unlock()

	// Execute handler
	return handler(ctx, event)
}

// getHandlers returns all handlers for an event type
func (er *EventRegistry) getHandlers(eventType string) []*HandlerEntry {
	er.mu.RLock()
	defer er.mu.RUnlock()

	handlers := make([]*HandlerEntry, 0)

	// Add specific handlers
	if specific, exists := er.handlers[eventType]; exists {
		for i := range specific {
			handlers = append(handlers, &specific[i])
		}
	}

	// Add wildcard handlers
	if er.config.EnableWildcards {
		for pattern, wildcardHandlers := range er.handlers {
			if pattern != eventType && matchesWildcard(pattern, eventType) {
				for i := range wildcardHandlers {
					handlers = append(handlers, &wildcardHandlers[i])
				}
			}
		}
	}

	// Add global handlers
	for i := range er.globalHandlers {
		handlers = append(handlers, &er.globalHandlers[i])
	}

	return handlers
}

// Use adds middleware to the event processing chain
func (er *EventRegistry) Use(middleware EventMiddleware) {
	er.mu.Lock()
	defer er.mu.Unlock()

	er.middleware = append(er.middleware, middleware)
}

// PreventDefault prevents default behavior for an event
func (er *EventRegistry) PreventDefault(eventID string) {
	er.mu.Lock()
	defer er.mu.Unlock()

	er.preventDefaults[eventID] = true
}

// StopPropagation stops event propagation
func (er *EventRegistry) StopPropagation(eventID string) {
	er.mu.Lock()
	defer er.mu.Unlock()

	er.stopPropagation[eventID] = true
}

// isDefaultPrevented checks if default is prevented
func (er *EventRegistry) isDefaultPrevented(eventID string) bool {
	er.mu.RLock()
	defer er.mu.RUnlock()

	return er.preventDefaults[eventID]
}

// isPropagationStopped checks if propagation is stopped
func (er *EventRegistry) isPropagationStopped(eventID string) bool {
	er.mu.RLock()
	defer er.mu.RUnlock()

	return er.stopPropagation[eventID]
}

// cleanupEventFlags cleans up event flags
func (er *EventRegistry) cleanupEventFlags(eventID string) {
	er.mu.Lock()
	defer er.mu.Unlock()

	delete(er.preventDefaults, eventID)
	delete(er.stopPropagation, eventID)
}

// sortHandlersByPriority sorts handlers by priority
func (er *EventRegistry) sortHandlersByPriority(eventType string) {
	handlers := er.handlers[eventType]
	
	// Simple bubble sort for small arrays
	for i := 0; i < len(handlers)-1; i++ {
		for j := 0; j < len(handlers)-i-1; j++ {
			if handlers[j].Priority < handlers[j+1].Priority {
				handlers[j], handlers[j+1] = handlers[j+1], handlers[j]
			}
		}
	}
	
	er.handlers[eventType] = handlers
}

// updateMetrics updates event metrics
func (er *EventRegistry) updateMetrics(eventType string, success bool) {
	if er.metrics == nil {
		return
	}

	er.metrics.mu.Lock()
	defer er.metrics.mu.Unlock()

	er.metrics.totalEvents++
	er.metrics.eventCounts[eventType]++
	er.metrics.handledEvents++
}

// GetMetrics returns current metrics
func (er *EventRegistry) GetMetrics() *EventMetrics {
	if er.metrics == nil {
		return nil
	}

	er.metrics.mu.RLock()
	defer er.metrics.mu.RUnlock()

	// Return a copy
	return &EventMetrics{
		totalEvents:   er.metrics.totalEvents,
		handledEvents: er.metrics.handledEvents,
		failedEvents:  er.metrics.failedEvents,
		avgDuration:   er.metrics.avgDuration,
		lastError:     er.metrics.lastError,
		lastErrorTime: er.metrics.lastErrorTime,
	}
}

// Clear removes all handlers and resets the registry
func (er *EventRegistry) Clear() {
	er.mu.Lock()
	defer er.mu.Unlock()

	er.handlers = make(map[string][]HandlerEntry)
	er.globalHandlers = make([]HandlerEntry, 0)
	er.middleware = make([]EventMiddleware, 0)
	er.preventDefaults = make(map[string]bool)
	er.stopPropagation = make(map[string]bool)

	if er.metrics != nil {
		er.metrics = &EventMetrics{
			eventCounts: make(map[string]int64),
		}
	}
}

// HandlerOption configures a handler entry
type HandlerOption func(*HandlerEntry)

// WithPriority sets handler priority
func WithPriority(priority int) HandlerOption {
	return func(entry *HandlerEntry) {
		entry.Priority = priority
	}
}

// WithOnce marks handler as one-time
func WithOnce() HandlerOption {
	return func(entry *HandlerEntry) {
		entry.Once = true
	}
}

// WithFilter sets an event filter
func WithFilter(filter EventFilter) HandlerOption {
	return func(entry *HandlerEntry) {
		entry.Filter = filter
	}
}

// WithNamespace sets handler namespace
func WithNamespace(namespace string) HandlerOption {
	return func(entry *HandlerEntry) {
		entry.Namespace = namespace
	}
}

// WithDescription sets handler description
func WithDescription(description string) HandlerOption {
	return func(entry *HandlerEntry) {
		entry.Description = description
	}
}

// matchesWildcard checks if a pattern matches an event type
func matchesWildcard(pattern, eventType string) bool {
	if pattern == "*" {
		return true
	}

	// Convert wildcard pattern to regex
	regexPattern := strings.ReplaceAll(pattern, ".", "\\.")
	regexPattern = strings.ReplaceAll(regexPattern, "*", ".*")
	regexPattern = "^" + regexPattern + "$"

	matched, err := regexp.MatchString(regexPattern, eventType)
	if err != nil {
		return false
	}

	return matched
}

// generateHandlerID generates a unique handler ID
func generateHandlerID() string {
	return fmt.Sprintf("handler_%d_%d", time.Now().UnixNano(), randInt(1000000))
}

// generateEventID generates a unique event ID
func generateEventID() string {
	return fmt.Sprintf("event_%d_%d", time.Now().UnixNano(), randInt(1000000))
}

// randInt generates a random integer up to max
func randInt(max int) int {
	return rand.Intn(max)
}

// EventBus provides a global event bus
type EventBus struct {
	registry *EventRegistry
	mu       sync.RWMutex
}

// GlobalEventBus is the global event bus instance
var GlobalEventBus = &EventBus{
	registry: NewEventRegistry(nil),
}

// On registers a global event handler
func (eb *EventBus) On(eventType string, handler EventHandler, options ...HandlerOption) (string, error) {
	return eb.registry.On(eventType, handler, options...)
}

// Off removes a global event handler
func (eb *EventBus) Off(handlerID string) error {
	return eb.registry.Off(handlerID)
}

// Emit emits a global event
func (eb *EventBus) Emit(eventType string, data map[string]interface{}) error {
	return eb.registry.Emit(eventType, data)
}

// DelegatingEventHandler creates a handler that delegates to a method
func DelegatingEventHandler(target interface{}, methodName string) EventHandler {
	return func(ctx context.Context, event *Event) error {
		targetValue := reflect.ValueOf(target)
		method := targetValue.MethodByName(methodName)
		
		if !method.IsValid() {
			return fmt.Errorf("method %s not found", methodName)
		}

		// Call with different signatures
		methodType := method.Type()
		
		switch methodType.NumIn() {
		case 0:
			// No arguments
			method.Call(nil)
		case 1:
			// Just event
			if methodType.In(0) == reflect.TypeOf(event) {
				method.Call([]reflect.Value{reflect.ValueOf(event)})
			} else if methodType.In(0) == reflect.TypeOf(ctx) {
				method.Call([]reflect.Value{reflect.ValueOf(ctx)})
			}
		case 2:
			// Context and event
			if methodType.In(0) == reflect.TypeOf(ctx) && methodType.In(1) == reflect.TypeOf(event) {
				method.Call([]reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(event)})
			}
		default:
			return fmt.Errorf("unsupported method signature for %s", methodName)
		}

		return nil
	}
}

// ChainEventHandlers chains multiple handlers into one
func ChainEventHandlers(handlers ...EventHandler) EventHandler {
	return func(ctx context.Context, event *Event) error {
		for _, handler := range handlers {
			if err := handler(ctx, event); err != nil {
				return err
			}
		}
		return nil
	}
}

// ConditionalEventHandler creates a conditional handler
func ConditionalEventHandler(condition func(*Event) bool, handler EventHandler) EventHandler {
	return func(ctx context.Context, event *Event) error {
		if condition(event) {
			return handler(ctx, event)
		}
		return nil
	}
}

// ThrottledEventHandler creates a throttled handler
func ThrottledEventHandler(handler EventHandler, duration time.Duration) EventHandler {
	var lastExecution time.Time
	var mu sync.Mutex

	return func(ctx context.Context, event *Event) error {
		mu.Lock()
		defer mu.Unlock()

		now := time.Now()
		if now.Sub(lastExecution) < duration {
			return nil // Skip execution
		}

		lastExecution = now
		return handler(ctx, event)
	}
}

// DebouncedEventHandler creates a debounced handler
func DebouncedEventHandler(handler EventHandler, delay time.Duration) EventHandler {
	var timer *time.Timer
	var mu sync.Mutex

	return func(ctx context.Context, event *Event) error {
		mu.Lock()
		defer mu.Unlock()

		if timer != nil {
			timer.Stop()
		}

		timer = time.AfterFunc(delay, func() {
			handler(ctx, event)
		})

		return nil
	}
}