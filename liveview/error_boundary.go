package liveview

import (
	"fmt"
	"runtime/debug"
	"sync"
	"time"
)

// ErrorBoundary provides error handling for components
type ErrorBoundary struct {
	mu               sync.RWMutex
	errors           []ComponentError
	fallbackRenderer FallbackRenderer
	errorHandler     ErrorHandler
	maxErrors        int
	recoveryEnabled  bool
}

// ComponentError represents an error that occurred in a component
type ComponentError struct {
	ComponentID string
	Error       error
	StackTrace  string
	Timestamp   int64
	Recovered   bool
}

// ErrorHandler is a function that handles component errors
type ErrorHandler func(err ComponentError) error

// FallbackRenderer renders fallback UI when an error occurs
type FallbackRenderer func(componentID string, err error) string

// DefaultErrorBoundary creates a new error boundary with default settings
func DefaultErrorBoundary() *ErrorBoundary {
	return &ErrorBoundary{
		errors:          make([]ComponentError, 0),
		maxErrors:       100,
		recoveryEnabled: true,
		fallbackRenderer: func(componentID string, err error) string {
			return fmt.Sprintf(`
				<div class="error-boundary" style="border: 2px solid red; padding: 20px; margin: 10px; background: #fee;">
					<h3 style="color: red;">Component Error</h3>
					<p>Component ID: %s</p>
					<p>Error: %v</p>
					<p style="color: gray; font-size: 0.9em;">This component has been disabled due to an error.</p>
				</div>
			`, componentID, err)
		},
		errorHandler: func(err ComponentError) error {
			Debug("Component error in %s: %v", err.ComponentID, err.Error)
			return nil
		},
	}
}

// NewErrorBoundary creates a new error boundary with custom settings
func NewErrorBoundary(maxErrors int, recoveryEnabled bool) *ErrorBoundary {
	eb := DefaultErrorBoundary()
	eb.maxErrors = maxErrors
	eb.recoveryEnabled = recoveryEnabled
	return eb
}

// SetFallbackRenderer sets a custom fallback renderer
func (eb *ErrorBoundary) SetFallbackRenderer(renderer FallbackRenderer) {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	eb.fallbackRenderer = renderer
}

// SetErrorHandler sets a custom error handler
func (eb *ErrorBoundary) SetErrorHandler(handler ErrorHandler) {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	eb.errorHandler = handler
}

// CatchError catches and handles a component error
func (eb *ErrorBoundary) CatchError(componentID string, err error) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	// Create error record
	componentErr := ComponentError{
		ComponentID: componentID,
		Error:       err,
		StackTrace:  string(debug.Stack()),
		Timestamp:   getCurrentTimestamp(),
		Recovered:   eb.recoveryEnabled,
	}

	// Add to errors list (with limit)
	if len(eb.errors) >= eb.maxErrors {
		eb.errors = eb.errors[1:] // Remove oldest error
	}
	eb.errors = append(eb.errors, componentErr)

	// Call error handler
	if eb.errorHandler != nil {
		if handlerErr := eb.errorHandler(componentErr); handlerErr != nil {
			Debug("Error handler failed: %v", handlerErr)
		}
	}
}

// GetErrors returns all recorded errors
func (eb *ErrorBoundary) GetErrors() []ComponentError {
	eb.mu.RLock()
	defer eb.mu.RUnlock()
	
	result := make([]ComponentError, len(eb.errors))
	copy(result, eb.errors)
	return result
}

// GetErrorCount returns the number of errors
func (eb *ErrorBoundary) GetErrorCount() int {
	eb.mu.RLock()
	defer eb.mu.RUnlock()
	return len(eb.errors)
}

// ClearErrors clears all recorded errors
func (eb *ErrorBoundary) ClearErrors() {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	eb.errors = make([]ComponentError, 0)
}

// RenderFallback renders fallback UI for a failed component
func (eb *ErrorBoundary) RenderFallback(componentID string, err error) string {
	eb.mu.RLock()
	defer eb.mu.RUnlock()
	
	if eb.fallbackRenderer != nil {
		return eb.fallbackRenderer(componentID, err)
	}
	return fmt.Sprintf("<div>Component %s failed: %v</div>", componentID, err)
}

// IsRecoveryEnabled checks if recovery is enabled
func (eb *ErrorBoundary) IsRecoveryEnabled() bool {
	eb.mu.RLock()
	defer eb.mu.RUnlock()
	return eb.recoveryEnabled
}

// ComponentWithErrorBoundary wraps a component with error boundary protection
type ComponentWithErrorBoundary[T Component] struct {
	Component     T
	ErrorBoundary *ErrorBoundary
	Driver        *ComponentDriver[T]
}

// SafeExecute executes a function with error recovery
func (eb *ErrorBoundary) SafeExecute(componentID string, fn func() error) (err error) {
	if !eb.recoveryEnabled {
		return fn()
	}

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic recovered: %v", r)
			eb.CatchError(componentID, err)
		}
	}()

	err = fn()
	if err != nil {
		eb.CatchError(componentID, err)
	}
	return err
}

// SafeRender renders a component with error boundary protection
func (eb *ErrorBoundary) SafeRender(componentID string, renderFn func() string) string {
	if !eb.recoveryEnabled {
		return renderFn()
	}

	var result string
	var renderErr error

	func() {
		defer func() {
			if r := recover(); r != nil {
				renderErr = fmt.Errorf("render panic: %v", r)
			}
		}()
		result = renderFn()
	}()

	if renderErr != nil {
		eb.CatchError(componentID, renderErr)
		return eb.RenderFallback(componentID, renderErr)
	}

	return result
}

// WrapDriverWithErrorBoundary wraps a component driver with error boundary protection
func WrapDriverWithErrorBoundary[T Component](driver *ComponentDriver[T], eb *ErrorBoundary) *ComponentDriver[T] {
	// Store reference to error boundary
	driver.errorBoundary = eb
	return driver
}

// ComponentDriverWithErrorBoundary extends ComponentDriver with error boundary support
func (cw *ComponentDriver[T]) WithErrorBoundary(eb *ErrorBoundary) *ComponentDriver[T] {
	cw.errorBoundary = eb
	return cw
}

// SafeCommit performs a commit with error recovery
func (cw *ComponentDriver[T]) SafeCommit() {
	if cw.errorBoundary != nil {
		err := cw.errorBoundary.SafeExecute(cw.GetIDComponet(), func() error {
			cw.Commit()
			return nil
		})
		if err != nil {
			// Send error message to client
			cw.channel <- map[string]interface{}{
				"type":  "error",
				"id":    cw.GetIDComponet(),
				"value": cw.errorBoundary.RenderFallback(cw.GetIDComponet(), err),
			}
		}
	} else {
		cw.Commit()
	}
}

// SafeExecuteEvent executes an event with error recovery
func (cw *ComponentDriver[T]) SafeExecuteEvent(name string, data interface{}) {
	if cw.errorBoundary != nil {
		err := cw.errorBoundary.SafeExecute(cw.GetIDComponet(), func() error {
			cw.ExecuteEvent(name, data)
			return nil
		})
		if err != nil {
			// Send error message to client
			cw.channel <- map[string]interface{}{
				"type":  "error",
				"id":    cw.GetIDComponet(),
				"value": cw.errorBoundary.RenderFallback(cw.GetIDComponet(), err),
			}
		}
	} else {
		cw.ExecuteEvent(name, data)
	}
}

// ErrorStats provides statistics about errors
type ErrorStats struct {
	TotalErrors      int
	RecoveredErrors  int
	ComponentErrors  map[string]int
	RecentErrors     []ComponentError
}

// GetErrorStats returns error statistics
func (eb *ErrorBoundary) GetErrorStats() ErrorStats {
	eb.mu.RLock()
	defer eb.mu.RUnlock()
	
	stats := ErrorStats{
		TotalErrors:     len(eb.errors),
		ComponentErrors: make(map[string]int),
		RecentErrors:    make([]ComponentError, 0),
	}
	
	for _, err := range eb.errors {
		if err.Recovered {
			stats.RecoveredErrors++
		}
		stats.ComponentErrors[err.ComponentID]++
	}
	
	// Get last 10 errors
	start := len(eb.errors) - 10
	if start < 0 {
		start = 0
	}
	stats.RecentErrors = eb.errors[start:]
	
	return stats
}

// ErrorBoundaryConfig provides configuration for error boundaries
type ErrorBoundaryConfig struct {
	MaxErrors            int
	RecoveryEnabled      bool
	LogErrors            bool
	SendErrorsToClient   bool
	CustomFallbackHTML   string
	ErrorWebhookURL      string
	RetryAttempts        int
	RetryDelayMs         int
}

// DefaultErrorBoundaryConfig returns default configuration
func DefaultErrorBoundaryConfig() *ErrorBoundaryConfig {
	return &ErrorBoundaryConfig{
		MaxErrors:          100,
		RecoveryEnabled:    true,
		LogErrors:          true,
		SendErrorsToClient: false,
		RetryAttempts:      0,
		RetryDelayMs:       1000,
	}
}

// ConfigureErrorBoundary creates an error boundary from configuration
func ConfigureErrorBoundary(config *ErrorBoundaryConfig) *ErrorBoundary {
	eb := NewErrorBoundary(config.MaxErrors, config.RecoveryEnabled)
	
	if config.CustomFallbackHTML != "" {
		eb.SetFallbackRenderer(func(componentID string, err error) string {
			return config.CustomFallbackHTML
		})
	}
	
	if config.LogErrors {
		originalHandler := eb.errorHandler
		eb.SetErrorHandler(func(err ComponentError) error {
			Debug("[ErrorBoundary] Component %s error: %v", err.ComponentID, err.Error)
			if originalHandler != nil {
				return originalHandler(err)
			}
			return nil
		})
	}
	
	return eb
}

// timeNow is a variable function for testing
var timeNow = time.Now

// getCurrentTimestamp returns current Unix timestamp in milliseconds
func getCurrentTimestamp() int64 {
	return timeNow().UnixNano() / 1e6
}