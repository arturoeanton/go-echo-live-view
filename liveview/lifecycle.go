package liveview

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// LifecycleStage represents the current stage of a component's lifecycle
type LifecycleStage int

const (
	// StageCreated indicates the component has been created but not initialized
	StageCreated LifecycleStage = iota
	// StageInitializing indicates the component is being initialized
	StageInitializing
	// StageMounting indicates the component is being mounted to the DOM
	StageMounting
	// StageMounted indicates the component has been mounted and is active
	StageMounted
	// StageUpdating indicates the component is being updated
	StageUpdating
	// StageUnmounting indicates the component is being unmounted
	StageUnmounting
	// StageUnmounted indicates the component has been unmounted
	StageUnmounted
	// StageError indicates the component encountered an error
	StageError
)

// String returns the string representation of a lifecycle stage
func (s LifecycleStage) String() string {
	stages := []string{
		"Created",
		"Initializing",
		"Mounting",
		"Mounted",
		"Updating",
		"Unmounting",
		"Unmounted",
		"Error",
	}
	if int(s) < len(stages) {
		return stages[s]
	}
	return "Unknown"
}

// LifecycleHooks defines hooks that can be registered for component lifecycle events
type LifecycleHooks struct {
	// OnBeforeCreate is called before the component is created
	OnBeforeCreate func() error
	
	// OnCreated is called after the component is created
	OnCreated func() error
	
	// OnBeforeMount is called before the component is mounted
	OnBeforeMount func() error
	
	// OnMounted is called after the component is mounted
	OnMounted func() error
	
	// OnBeforeUpdate is called before the component is updated
	OnBeforeUpdate func(oldData, newData interface{}) error
	
	// OnUpdated is called after the component is updated
	OnUpdated func() error
	
	// OnBeforeUnmount is called before the component is unmounted
	OnBeforeUnmount func() error
	
	// OnUnmounted is called after the component is unmounted
	OnUnmounted func() error
	
	// OnError is called when an error occurs during lifecycle
	OnError func(stage LifecycleStage, err error) error
	
	// OnStateChange is called when the lifecycle stage changes
	OnStateChange func(oldStage, newStage LifecycleStage)
}

// LifecycleManager manages component lifecycle
type LifecycleManager struct {
	mu          sync.RWMutex
	stage       LifecycleStage
	hooks       *LifecycleHooks
	componentID string
	history     []LifecycleTransition
	maxHistory  int
	metadata    map[string]interface{}
}

// LifecycleTransition records a transition between lifecycle stages
type LifecycleTransition struct {
	From      LifecycleStage
	To        LifecycleStage
	Timestamp time.Time
	Duration  time.Duration
	Error     error
}

// NewLifecycleManager creates a new lifecycle manager
func NewLifecycleManager(componentID string) *LifecycleManager {
	return &LifecycleManager{
		stage:       StageCreated,
		componentID: componentID,
		hooks:       &LifecycleHooks{},
		history:     make([]LifecycleTransition, 0),
		maxHistory:  100,
		metadata:    make(map[string]interface{}),
	}
}

// SetHooks sets the lifecycle hooks
func (lm *LifecycleManager) SetHooks(hooks *LifecycleHooks) {
	lm.mu.Lock()
	defer lm.mu.Unlock()
	lm.hooks = hooks
}

// GetStage returns the current lifecycle stage
func (lm *LifecycleManager) GetStage() LifecycleStage {
	lm.mu.RLock()
	defer lm.mu.RUnlock()
	return lm.stage
}

// SetMetadata stores metadata associated with the component
func (lm *LifecycleManager) SetMetadata(key string, value interface{}) {
	lm.mu.Lock()
	defer lm.mu.Unlock()
	lm.metadata[key] = value
}

// GetMetadata retrieves metadata associated with the component
func (lm *LifecycleManager) GetMetadata(key string) interface{} {
	lm.mu.RLock()
	defer lm.mu.RUnlock()
	return lm.metadata[key]
}

// transitionTo transitions to a new lifecycle stage
func (lm *LifecycleManager) transitionTo(newStage LifecycleStage) error {
	lm.mu.Lock()
	oldStage := lm.stage
	startTime := time.Now()
	
	// Record transition
	transition := LifecycleTransition{
		From:      oldStage,
		To:        newStage,
		Timestamp: startTime,
	}
	
	lm.stage = newStage
	lm.mu.Unlock()
	
	// Call state change hook
	if lm.hooks != nil && lm.hooks.OnStateChange != nil {
		lm.hooks.OnStateChange(oldStage, newStage)
	}
	
	// Record transition history
	lm.mu.Lock()
	transition.Duration = time.Since(startTime)
	if len(lm.history) >= lm.maxHistory {
		lm.history = lm.history[1:]
	}
	lm.history = append(lm.history, transition)
	lm.mu.Unlock()
	
	Debug("Component %s transitioned from %s to %s", lm.componentID, oldStage, newStage)
	return nil
}

// handleError handles lifecycle errors
func (lm *LifecycleManager) handleError(stage LifecycleStage, err error) error {
	if err == nil {
		return nil
	}
	
	lm.mu.Lock()
	lm.stage = StageError
	lm.mu.Unlock()
	
	if lm.hooks != nil && lm.hooks.OnError != nil {
		if hookErr := lm.hooks.OnError(stage, err); hookErr != nil {
			return fmt.Errorf("error in OnError hook: %w", hookErr)
		}
	}
	
	return err
}

// Create executes the create lifecycle phase
func (lm *LifecycleManager) Create() error {
	if lm.hooks != nil && lm.hooks.OnBeforeCreate != nil {
		if err := lm.hooks.OnBeforeCreate(); err != nil {
			return lm.handleError(StageCreated, err)
		}
	}
	
	if err := lm.transitionTo(StageInitializing); err != nil {
		return lm.handleError(StageInitializing, err)
	}
	
	if lm.hooks != nil && lm.hooks.OnCreated != nil {
		if err := lm.hooks.OnCreated(); err != nil {
			return lm.handleError(StageInitializing, err)
		}
	}
	
	return nil
}

// Mount executes the mount lifecycle phase
func (lm *LifecycleManager) Mount() error {
	currentStage := lm.GetStage()
	if currentStage != StageInitializing && currentStage != StageUnmounted {
		return fmt.Errorf("cannot mount from stage %s", currentStage)
	}
	
	if lm.hooks != nil && lm.hooks.OnBeforeMount != nil {
		if err := lm.hooks.OnBeforeMount(); err != nil {
			return lm.handleError(StageMounting, err)
		}
	}
	
	if err := lm.transitionTo(StageMounting); err != nil {
		return lm.handleError(StageMounting, err)
	}
	
	// Simulate mounting process
	if err := lm.transitionTo(StageMounted); err != nil {
		return lm.handleError(StageMounted, err)
	}
	
	if lm.hooks != nil && lm.hooks.OnMounted != nil {
		if err := lm.hooks.OnMounted(); err != nil {
			return lm.handleError(StageMounted, err)
		}
	}
	
	return nil
}

// Update executes the update lifecycle phase
func (lm *LifecycleManager) Update(oldData, newData interface{}) error {
	currentStage := lm.GetStage()
	if currentStage != StageMounted && currentStage != StageUpdating {
		return fmt.Errorf("cannot update from stage %s", currentStage)
	}
	
	if lm.hooks != nil && lm.hooks.OnBeforeUpdate != nil {
		if err := lm.hooks.OnBeforeUpdate(oldData, newData); err != nil {
			return lm.handleError(StageUpdating, err)
		}
	}
	
	if err := lm.transitionTo(StageUpdating); err != nil {
		return lm.handleError(StageUpdating, err)
	}
	
	// Simulate update process
	if err := lm.transitionTo(StageMounted); err != nil {
		return lm.handleError(StageMounted, err)
	}
	
	if lm.hooks != nil && lm.hooks.OnUpdated != nil {
		if err := lm.hooks.OnUpdated(); err != nil {
			return lm.handleError(StageMounted, err)
		}
	}
	
	return nil
}

// Unmount executes the unmount lifecycle phase
func (lm *LifecycleManager) Unmount() error {
	currentStage := lm.GetStage()
	if currentStage == StageUnmounted || currentStage == StageUnmounting {
		return nil // Already unmounted or unmounting
	}
	
	if lm.hooks != nil && lm.hooks.OnBeforeUnmount != nil {
		if err := lm.hooks.OnBeforeUnmount(); err != nil {
			return lm.handleError(StageUnmounting, err)
		}
	}
	
	if err := lm.transitionTo(StageUnmounting); err != nil {
		return lm.handleError(StageUnmounting, err)
	}
	
	// Simulate unmounting process
	if err := lm.transitionTo(StageUnmounted); err != nil {
		return lm.handleError(StageUnmounted, err)
	}
	
	if lm.hooks != nil && lm.hooks.OnUnmounted != nil {
		if err := lm.hooks.OnUnmounted(); err != nil {
			return lm.handleError(StageUnmounted, err)
		}
	}
	
	return nil
}

// GetHistory returns the lifecycle transition history
func (lm *LifecycleManager) GetHistory() []LifecycleTransition {
	lm.mu.RLock()
	defer lm.mu.RUnlock()
	
	result := make([]LifecycleTransition, len(lm.history))
	copy(result, lm.history)
	return result
}

// ComponentWithLifecycle extends a component with lifecycle management
type ComponentWithLifecycle interface {
	Component
	GetLifecycleManager() *LifecycleManager
}

// LifecycleComponent wraps a component with lifecycle management
type LifecycleComponent[T Component] struct {
	Component T
	Lifecycle *LifecycleManager
}

// GetLifecycleManager returns the lifecycle manager
func (lc *LifecycleComponent[T]) GetLifecycleManager() *LifecycleManager {
	return lc.Lifecycle
}

// GetTemplate delegates to the wrapped component
func (lc *LifecycleComponent[T]) GetTemplate() string {
	return lc.Component.GetTemplate()
}

// GetDriver delegates to the wrapped component
func (lc *LifecycleComponent[T]) GetDriver() LiveDriver {
	return lc.Component.GetDriver()
}

// Start initializes the component with lifecycle
func (lc *LifecycleComponent[T]) Start() {
	// Create lifecycle
	if err := lc.Lifecycle.Create(); err != nil {
		Debug("Lifecycle create error: %v", err)
	}
	
	// Start the wrapped component
	lc.Component.Start()
	
	// Mount lifecycle
	if err := lc.Lifecycle.Mount(); err != nil {
		Debug("Lifecycle mount error: %v", err)
	}
}

// WrapWithLifecycle wraps a component with lifecycle management
func WrapWithLifecycle[T Component](component T, componentID string) *LifecycleComponent[T] {
	return &LifecycleComponent[T]{
		Component: component,
		Lifecycle: NewLifecycleManager(componentID),
	}
}

// AttachLifecycleToDriver attaches lifecycle management to a component driver
func AttachLifecycleToDriver[T Component](driver *ComponentDriver[T]) *ComponentDriver[T] {
	lifecycle := NewLifecycleManager(driver.GetIDComponet())
	
	// Store lifecycle in driver's data
	driver.SetData(map[string]interface{}{
		"lifecycle": lifecycle,
	})
	
	// Hook into commit to trigger updates
	originalCommit := driver.Commit
	driver.lifecycleCommit = func() {
		oldData := driver.GetData()
		lifecycle.Update(oldData, driver.GetData())
		originalCommit()
	}
	
	return driver
}

// GetLifecycleFromDriver retrieves the lifecycle manager from a driver
func GetLifecycleFromDriver[T Component](driver *ComponentDriver[T]) *LifecycleManager {
	data := driver.GetData()
	if dataMap, ok := data.(map[string]interface{}); ok {
		if lifecycle, ok := dataMap["lifecycle"].(*LifecycleManager); ok {
			return lifecycle
		}
	}
	return nil
}

// LifecycleAware interface for components that want to manage their own lifecycle
type LifecycleAware interface {
	// OnBeforeCreate is called before the component is created
	OnBeforeCreate() error
	
	// OnCreated is called after the component is created
	OnCreated() error
	
	// OnBeforeMount is called before the component is mounted
	OnBeforeMount() error
	
	// OnMounted is called after the component is mounted
	OnMounted() error
	
	// OnBeforeUpdate is called before the component is updated
	OnBeforeUpdate(oldData, newData interface{}) error
	
	// OnUpdated is called after the component is updated
	OnUpdated() error
	
	// OnBeforeUnmount is called before the component is unmounted
	OnBeforeUnmount() error
	
	// OnUnmounted is called after the component is unmounted
	OnUnmounted() error
}

// AutoLifecycleHooks creates lifecycle hooks from a LifecycleAware component
func AutoLifecycleHooks(component LifecycleAware) *LifecycleHooks {
	return &LifecycleHooks{
		OnBeforeCreate:  component.OnBeforeCreate,
		OnCreated:       component.OnCreated,
		OnBeforeMount:   component.OnBeforeMount,
		OnMounted:       component.OnMounted,
		OnBeforeUpdate:  component.OnBeforeUpdate,
		OnUpdated:       component.OnUpdated,
		OnBeforeUnmount: component.OnBeforeUnmount,
		OnUnmounted:     component.OnUnmounted,
	}
}

// WithLifecycleContext creates a context that is cancelled when component unmounts
func WithLifecycleContext(ctx context.Context, lifecycle *LifecycleManager) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(ctx)
	
	// Cancel context on unmount
	oldOnBeforeUnmount := lifecycle.hooks.OnBeforeUnmount
	lifecycle.hooks.OnBeforeUnmount = func() error {
		cancel()
		if oldOnBeforeUnmount != nil {
			return oldOnBeforeUnmount()
		}
		return nil
	}
	
	return ctx, cancel
}