package liveview

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestLifecycleStageString(t *testing.T) {
	tests := []struct {
		stage    LifecycleStage
		expected string
	}{
		{StageCreated, "Created"},
		{StageInitializing, "Initializing"},
		{StageMounting, "Mounting"},
		{StageMounted, "Mounted"},
		{StageUpdating, "Updating"},
		{StageUnmounting, "Unmounting"},
		{StageUnmounted, "Unmounted"},
		{StageError, "Error"},
		{LifecycleStage(999), "Unknown"},
	}
	
	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.stage.String(); got != tt.expected {
				t.Errorf("Stage.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestNewLifecycleManager(t *testing.T) {
	lm := NewLifecycleManager("test-component")
	
	if lm == nil {
		t.Fatal("NewLifecycleManager returned nil")
	}
	
	if lm.componentID != "test-component" {
		t.Errorf("Expected componentID 'test-component', got %s", lm.componentID)
	}
	
	if lm.GetStage() != StageCreated {
		t.Errorf("Expected initial stage to be Created, got %s", lm.GetStage())
	}
	
	if lm.hooks == nil {
		t.Error("Expected hooks to be initialized")
	}
	
	if len(lm.history) != 0 {
		t.Errorf("Expected empty history, got %d items", len(lm.history))
	}
}

func TestLifecycleMetadata(t *testing.T) {
	lm := NewLifecycleManager("test")
	
	// Set metadata
	lm.SetMetadata("key1", "value1")
	lm.SetMetadata("key2", 42)
	
	// Get metadata
	if val := lm.GetMetadata("key1"); val != "value1" {
		t.Errorf("Expected metadata key1='value1', got %v", val)
	}
	
	if val := lm.GetMetadata("key2"); val != 42 {
		t.Errorf("Expected metadata key2=42, got %v", val)
	}
	
	if val := lm.GetMetadata("nonexistent"); val != nil {
		t.Errorf("Expected nil for nonexistent key, got %v", val)
	}
}

func TestLifecycleCreate(t *testing.T) {
	lm := NewLifecycleManager("test")
	
	var beforeCreateCalled, createdCalled bool
	
	lm.SetHooks(&LifecycleHooks{
		OnBeforeCreate: func() error {
			beforeCreateCalled = true
			return nil
		},
		OnCreated: func() error {
			createdCalled = true
			return nil
		},
	})
	
	err := lm.Create()
	if err != nil {
		t.Errorf("Create() returned error: %v", err)
	}
	
	if !beforeCreateCalled {
		t.Error("OnBeforeCreate hook was not called")
	}
	
	if !createdCalled {
		t.Error("OnCreated hook was not called")
	}
	
	if lm.GetStage() != StageInitializing {
		t.Errorf("Expected stage to be Initializing after Create, got %s", lm.GetStage())
	}
}

func TestLifecycleMount(t *testing.T) {
	lm := NewLifecycleManager("test")
	
	// Must create before mount
	err := lm.Create()
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}
	
	var beforeMountCalled, mountedCalled bool
	
	lm.SetHooks(&LifecycleHooks{
		OnBeforeMount: func() error {
			beforeMountCalled = true
			return nil
		},
		OnMounted: func() error {
			mountedCalled = true
			return nil
		},
	})
	
	err = lm.Mount()
	if err != nil {
		t.Errorf("Mount() returned error: %v", err)
	}
	
	if !beforeMountCalled {
		t.Error("OnBeforeMount hook was not called")
	}
	
	if !mountedCalled {
		t.Error("OnMounted hook was not called")
	}
	
	if lm.GetStage() != StageMounted {
		t.Errorf("Expected stage to be Mounted, got %s", lm.GetStage())
	}
}

func TestLifecycleUpdate(t *testing.T) {
	lm := NewLifecycleManager("test")
	
	// Must create and mount before update
	lm.Create()
	lm.Mount()
	
	var beforeUpdateCalled, updatedCalled bool
	var capturedOldData, capturedNewData interface{}
	
	lm.SetHooks(&LifecycleHooks{
		OnBeforeUpdate: func(oldData, newData interface{}) error {
			beforeUpdateCalled = true
			capturedOldData = oldData
			capturedNewData = newData
			return nil
		},
		OnUpdated: func() error {
			updatedCalled = true
			return nil
		},
	})
	
	oldData := "old"
	newData := "new"
	
	err := lm.Update(oldData, newData)
	if err != nil {
		t.Errorf("Update() returned error: %v", err)
	}
	
	if !beforeUpdateCalled {
		t.Error("OnBeforeUpdate hook was not called")
	}
	
	if !updatedCalled {
		t.Error("OnUpdated hook was not called")
	}
	
	if capturedOldData != oldData {
		t.Errorf("Expected oldData '%v', got '%v'", oldData, capturedOldData)
	}
	
	if capturedNewData != newData {
		t.Errorf("Expected newData '%v', got '%v'", newData, capturedNewData)
	}
	
	if lm.GetStage() != StageMounted {
		t.Errorf("Expected stage to be Mounted after update, got %s", lm.GetStage())
	}
}

func TestLifecycleUnmount(t *testing.T) {
	lm := NewLifecycleManager("test")
	
	// Setup full lifecycle
	lm.Create()
	lm.Mount()
	
	var beforeUnmountCalled, unmountedCalled bool
	
	lm.SetHooks(&LifecycleHooks{
		OnBeforeUnmount: func() error {
			beforeUnmountCalled = true
			return nil
		},
		OnUnmounted: func() error {
			unmountedCalled = true
			return nil
		},
	})
	
	err := lm.Unmount()
	if err != nil {
		t.Errorf("Unmount() returned error: %v", err)
	}
	
	if !beforeUnmountCalled {
		t.Error("OnBeforeUnmount hook was not called")
	}
	
	if !unmountedCalled {
		t.Error("OnUnmounted hook was not called")
	}
	
	if lm.GetStage() != StageUnmounted {
		t.Errorf("Expected stage to be Unmounted, got %s", lm.GetStage())
	}
	
	// Unmounting again should not error
	err = lm.Unmount()
	if err != nil {
		t.Errorf("Second Unmount() should not error, got: %v", err)
	}
}

func TestLifecycleErrorHandling(t *testing.T) {
	lm := NewLifecycleManager("test")
	
	testError := errors.New("test error")
	var errorHandled bool
	var capturedStage LifecycleStage
	var capturedError error
	
	lm.SetHooks(&LifecycleHooks{
		OnBeforeCreate: func() error {
			return testError
		},
		OnError: func(stage LifecycleStage, err error) error {
			errorHandled = true
			capturedStage = stage
			capturedError = err
			return nil
		},
	})
	
	err := lm.Create()
	if err == nil {
		t.Error("Expected error from Create()")
	}
	
	if !errorHandled {
		t.Error("OnError hook was not called")
	}
	
	if capturedStage != StageCreated {
		t.Errorf("Expected error stage Created, got %s", capturedStage)
	}
	
	if capturedError != testError {
		t.Errorf("Expected error %v, got %v", testError, capturedError)
	}
	
	if lm.GetStage() != StageError {
		t.Errorf("Expected stage to be Error, got %s", lm.GetStage())
	}
}

func TestLifecycleStateTransitions(t *testing.T) {
	lm := NewLifecycleManager("test")
	
	var transitions []string
	
	lm.SetHooks(&LifecycleHooks{
		OnStateChange: func(oldStage, newStage LifecycleStage) {
			transitions = append(transitions, fmt.Sprintf("%s->%s", oldStage, newStage))
		},
	})
	
	// Run through lifecycle
	lm.Create()
	lm.Mount()
	lm.Update("old", "new")
	lm.Unmount()
	
	expectedTransitions := []string{
		"Created->Initializing",
		"Initializing->Mounting",
		"Mounting->Mounted",
		"Mounted->Updating",
		"Updating->Mounted",
		"Mounted->Unmounting",
		"Unmounting->Unmounted",
	}
	
	if len(transitions) != len(expectedTransitions) {
		t.Fatalf("Expected %d transitions, got %d", len(expectedTransitions), len(transitions))
	}
	
	for i, expected := range expectedTransitions {
		if transitions[i] != expected {
			t.Errorf("Transition %d: expected %s, got %s", i, expected, transitions[i])
		}
	}
}

func TestLifecycleHistory(t *testing.T) {
	lm := NewLifecycleManager("test")
	lm.maxHistory = 3 // Set small history for testing
	
	// Run through lifecycle
	lm.Create()
	lm.Mount()
	lm.Update("old", "new")
	lm.Update("new", "newer")
	lm.Unmount()
	
	history := lm.GetHistory()
	
	// Should only have last 3 transitions due to maxHistory
	if len(history) != 3 {
		t.Fatalf("Expected 3 history items (maxHistory), got %d", len(history))
	}
	
	// Check that transitions have timestamps and durations
	for i, transition := range history {
		if transition.Timestamp.IsZero() {
			t.Errorf("History item %d has zero timestamp", i)
		}
		if transition.Duration < 0 {
			t.Errorf("History item %d has negative duration", i)
		}
	}
}

func TestInvalidTransitions(t *testing.T) {
	tests := []struct {
		name        string
		setupStage  func(*LifecycleManager)
		operation   func(*LifecycleManager) error
		shouldError bool
	}{
		{
			name: "Mount from Created",
			setupStage: func(lm *LifecycleManager) {
				// Leave at Created stage
			},
			operation: func(lm *LifecycleManager) error {
				return lm.Mount()
			},
			shouldError: true,
		},
		{
			name: "Update from Created",
			setupStage: func(lm *LifecycleManager) {
				// Leave at Created stage
			},
			operation: func(lm *LifecycleManager) error {
				return lm.Update("old", "new")
			},
			shouldError: true,
		},
		{
			name: "Update from Mounted",
			setupStage: func(lm *LifecycleManager) {
				lm.Create()
				lm.Mount()
			},
			operation: func(lm *LifecycleManager) error {
				return lm.Update("old", "new")
			},
			shouldError: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lm := NewLifecycleManager("test")
			tt.setupStage(lm)
			
			err := tt.operation(lm)
			
			if tt.shouldError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.shouldError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

// TestLifecycleAwareComponent tests a component that implements LifecycleAware
type TestLifecycleAwareComponent struct {
	Driver               *ComponentDriver[*TestLifecycleAwareComponent]
	beforeCreateCalled   bool
	createdCalled        bool
	beforeMountCalled    bool
	mountedCalled        bool
	beforeUpdateCalled   bool
	updatedCalled        bool
	beforeUnmountCalled  bool
	unmountedCalled      bool
}

func (c *TestLifecycleAwareComponent) GetDriver() LiveDriver {
	return c.Driver
}

func (c *TestLifecycleAwareComponent) GetTemplate() string {
	return "<div>Test</div>"
}

func (c *TestLifecycleAwareComponent) Start() {}

func (c *TestLifecycleAwareComponent) OnBeforeCreate() error {
	c.beforeCreateCalled = true
	return nil
}

func (c *TestLifecycleAwareComponent) OnCreated() error {
	c.createdCalled = true
	return nil
}

func (c *TestLifecycleAwareComponent) OnBeforeMount() error {
	c.beforeMountCalled = true
	return nil
}

func (c *TestLifecycleAwareComponent) OnMounted() error {
	c.mountedCalled = true
	return nil
}

func (c *TestLifecycleAwareComponent) OnBeforeUpdate(oldData, newData interface{}) error {
	c.beforeUpdateCalled = true
	return nil
}

func (c *TestLifecycleAwareComponent) OnUpdated() error {
	c.updatedCalled = true
	return nil
}

func (c *TestLifecycleAwareComponent) OnBeforeUnmount() error {
	c.beforeUnmountCalled = true
	return nil
}

func (c *TestLifecycleAwareComponent) OnUnmounted() error {
	c.unmountedCalled = true
	return nil
}

func TestAutoLifecycleHooks(t *testing.T) {
	component := &TestLifecycleAwareComponent{}
	hooks := AutoLifecycleHooks(component)
	
	// Test that hooks are properly connected
	hooks.OnBeforeCreate()
	if !component.beforeCreateCalled {
		t.Error("OnBeforeCreate hook not connected")
	}
	
	hooks.OnCreated()
	if !component.createdCalled {
		t.Error("OnCreated hook not connected")
	}
	
	hooks.OnBeforeMount()
	if !component.beforeMountCalled {
		t.Error("OnBeforeMount hook not connected")
	}
	
	hooks.OnMounted()
	if !component.mountedCalled {
		t.Error("OnMounted hook not connected")
	}
	
	hooks.OnBeforeUpdate(nil, nil)
	if !component.beforeUpdateCalled {
		t.Error("OnBeforeUpdate hook not connected")
	}
	
	hooks.OnUpdated()
	if !component.updatedCalled {
		t.Error("OnUpdated hook not connected")
	}
	
	hooks.OnBeforeUnmount()
	if !component.beforeUnmountCalled {
		t.Error("OnBeforeUnmount hook not connected")
	}
	
	hooks.OnUnmounted()
	if !component.unmountedCalled {
		t.Error("OnUnmounted hook not connected")
	}
}

func TestWrapWithLifecycle(t *testing.T) {
	component := &TestComponent{}
	wrapped := WrapWithLifecycle(component, "test-wrapped")
	
	if wrapped == nil {
		t.Fatal("WrapWithLifecycle returned nil")
	}
	
	if wrapped.Component != component {
		t.Error("Wrapped component doesn't match original")
	}
	
	if wrapped.Lifecycle == nil {
		t.Fatal("Lifecycle manager is nil")
	}
	
	if wrapped.Lifecycle.componentID != "test-wrapped" {
		t.Errorf("Expected componentID 'test-wrapped', got %s", wrapped.Lifecycle.componentID)
	}
	
	// Test delegation
	if wrapped.GetTemplate() != component.GetTemplate() {
		t.Error("GetTemplate delegation failed")
	}
}

func TestWithLifecycleContext(t *testing.T) {
	lm := NewLifecycleManager("test")
	parentCtx := context.Background()
	
	ctx, cancel := WithLifecycleContext(parentCtx, lm)
	defer cancel()
	
	// Context should not be cancelled initially
	select {
	case <-ctx.Done():
		t.Error("Context cancelled prematurely")
	default:
		// Expected
	}
	
	// Setup lifecycle
	lm.Create()
	lm.Mount()
	
	// Unmount should cancel context
	lm.Unmount()
	
	// Give context time to cancel
	time.Sleep(10 * time.Millisecond)
	
	select {
	case <-ctx.Done():
		// Expected - context should be cancelled
	default:
		t.Error("Context not cancelled after unmount")
	}
}

func TestConcurrentLifecycleOperations(t *testing.T) {
	lm := NewLifecycleManager("test")
	
	// Setup lifecycle
	lm.Create()
	lm.Mount()
	
	var wg sync.WaitGroup
	numGoroutines := 100
	
	// Concurrent updates
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			lm.Update(id, id+1)
			lm.SetMetadata(fmt.Sprintf("key%d", id), id)
		}(i)
	}
	
	wg.Wait()
	
	// Should still be in valid state
	stage := lm.GetStage()
	if stage != StageMounted && stage != StageUpdating {
		t.Errorf("Unexpected stage after concurrent operations: %s", stage)
	}
	
	// Check some metadata was set
	if lm.GetMetadata("key0") == nil {
		t.Error("Metadata not set during concurrent operations")
	}
}

func TestLifecycleComponentIntegration(t *testing.T) {
	component := &TestComponent{}
	wrapped := WrapWithLifecycle(component, "integration-test")
	
	// Set hooks to track lifecycle
	var lifecycleEvents []string
	
	wrapped.Lifecycle.SetHooks(&LifecycleHooks{
		OnBeforeCreate: func() error {
			lifecycleEvents = append(lifecycleEvents, "before-create")
			return nil
		},
		OnCreated: func() error {
			lifecycleEvents = append(lifecycleEvents, "created")
			return nil
		},
		OnBeforeMount: func() error {
			lifecycleEvents = append(lifecycleEvents, "before-mount")
			return nil
		},
		OnMounted: func() error {
			lifecycleEvents = append(lifecycleEvents, "mounted")
			return nil
		},
	})
	
	// Start should trigger create and mount
	wrapped.Start()
	
	expectedEvents := []string{"before-create", "created", "before-mount", "mounted"}
	
	if len(lifecycleEvents) != len(expectedEvents) {
		t.Fatalf("Expected %d lifecycle events, got %d", len(expectedEvents), len(lifecycleEvents))
	}
	
	for i, expected := range expectedEvents {
		if lifecycleEvents[i] != expected {
			t.Errorf("Event %d: expected %s, got %s", i, expected, lifecycleEvents[i])
		}
	}
}