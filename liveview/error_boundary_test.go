package liveview

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestDefaultErrorBoundary(t *testing.T) {
	eb := DefaultErrorBoundary()
	
	if eb == nil {
		t.Fatal("DefaultErrorBoundary returned nil")
	}
	
	if eb.maxErrors != 100 {
		t.Errorf("Expected maxErrors to be 100, got %d", eb.maxErrors)
	}
	
	if !eb.recoveryEnabled {
		t.Error("Expected recoveryEnabled to be true")
	}
	
	if eb.fallbackRenderer == nil {
		t.Error("Expected fallbackRenderer to be set")
	}
	
	if eb.errorHandler == nil {
		t.Error("Expected errorHandler to be set")
	}
}

func TestNewErrorBoundary(t *testing.T) {
	eb := NewErrorBoundary(50, false)
	
	if eb.maxErrors != 50 {
		t.Errorf("Expected maxErrors to be 50, got %d", eb.maxErrors)
	}
	
	if eb.recoveryEnabled {
		t.Error("Expected recoveryEnabled to be false")
	}
}

func TestCatchError(t *testing.T) {
	eb := DefaultErrorBoundary()
	testErr := errors.New("test error")
	
	eb.CatchError("component1", testErr)
	
	errors := eb.GetErrors()
	if len(errors) != 1 {
		t.Fatalf("Expected 1 error, got %d", len(errors))
	}
	
	if errors[0].ComponentID != "component1" {
		t.Errorf("Expected ComponentID 'component1', got %s", errors[0].ComponentID)
	}
	
	if errors[0].Error.Error() != "test error" {
		t.Errorf("Expected error message 'test error', got %s", errors[0].Error.Error())
	}
	
	if errors[0].StackTrace == "" {
		t.Error("Expected StackTrace to be set")
	}
	
	if !errors[0].Recovered {
		t.Error("Expected Recovered to be true")
	}
}

func TestMaxErrors(t *testing.T) {
	eb := NewErrorBoundary(3, true)
	
	// Add 5 errors
	for i := 0; i < 5; i++ {
		eb.CatchError(fmt.Sprintf("component%d", i), fmt.Errorf("error %d", i))
	}
	
	errors := eb.GetErrors()
	if len(errors) != 3 {
		t.Fatalf("Expected 3 errors (max limit), got %d", len(errors))
	}
	
	// Check that we have the last 3 errors
	for i := 0; i < 3; i++ {
		expectedID := fmt.Sprintf("component%d", i+2)
		if errors[i].ComponentID != expectedID {
			t.Errorf("Expected ComponentID %s, got %s", expectedID, errors[i].ComponentID)
		}
	}
}

func TestClearErrors(t *testing.T) {
	eb := DefaultErrorBoundary()
	
	eb.CatchError("component1", errors.New("error1"))
	eb.CatchError("component2", errors.New("error2"))
	
	if eb.GetErrorCount() != 2 {
		t.Errorf("Expected 2 errors, got %d", eb.GetErrorCount())
	}
	
	eb.ClearErrors()
	
	if eb.GetErrorCount() != 0 {
		t.Errorf("Expected 0 errors after clear, got %d", eb.GetErrorCount())
	}
}

func TestRenderFallback(t *testing.T) {
	eb := DefaultErrorBoundary()
	
	html := eb.RenderFallback("test-component", errors.New("test error"))
	
	if !strings.Contains(html, "test-component") {
		t.Error("Fallback HTML should contain component ID")
	}
	
	if !strings.Contains(html, "test error") {
		t.Error("Fallback HTML should contain error message")
	}
	
	if !strings.Contains(html, "error-boundary") {
		t.Error("Fallback HTML should contain error-boundary class")
	}
}

func TestCustomFallbackRenderer(t *testing.T) {
	eb := DefaultErrorBoundary()
	
	customHTML := "<div>Custom Error</div>"
	eb.SetFallbackRenderer(func(componentID string, err error) string {
		return customHTML
	})
	
	html := eb.RenderFallback("test", errors.New("error"))
	
	if html != customHTML {
		t.Errorf("Expected custom HTML %s, got %s", customHTML, html)
	}
}

func TestCustomErrorHandler(t *testing.T) {
	eb := DefaultErrorBoundary()
	
	var handlerCalled bool
	var capturedError ComponentError
	
	eb.SetErrorHandler(func(err ComponentError) error {
		handlerCalled = true
		capturedError = err
		return nil
	})
	
	eb.CatchError("test-component", errors.New("test error"))
	
	if !handlerCalled {
		t.Error("Custom error handler was not called")
	}
	
	if capturedError.ComponentID != "test-component" {
		t.Errorf("Expected ComponentID 'test-component', got %s", capturedError.ComponentID)
	}
}

func TestSafeExecute(t *testing.T) {
	eb := DefaultErrorBoundary()
	
	t.Run("Successful execution", func(t *testing.T) {
		err := eb.SafeExecute("component1", func() error {
			return nil
		})
		
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		
		if eb.GetErrorCount() != 0 {
			t.Errorf("Expected 0 errors, got %d", eb.GetErrorCount())
		}
	})
	
	t.Run("Error execution", func(t *testing.T) {
		testErr := errors.New("execution error")
		err := eb.SafeExecute("component2", func() error {
			return testErr
		})
		
		if err == nil {
			t.Error("Expected error to be returned")
		}
		
		if eb.GetErrorCount() != 1 {
			t.Errorf("Expected 1 error to be recorded, got %d", eb.GetErrorCount())
		}
	})
	
	t.Run("Panic recovery", func(t *testing.T) {
		err := eb.SafeExecute("component3", func() error {
			panic("test panic")
		})
		
		if err == nil {
			t.Error("Expected error from panic recovery")
		}
		
		if !strings.Contains(err.Error(), "panic recovered") {
			t.Errorf("Expected panic recovery error, got %v", err)
		}
		
		if eb.GetErrorCount() != 2 { // Including previous test
			t.Errorf("Expected 2 errors to be recorded, got %d", eb.GetErrorCount())
		}
	})
	
	t.Run("Recovery disabled", func(t *testing.T) {
		ebNoRecovery := NewErrorBoundary(100, false)
		
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic to propagate when recovery is disabled")
			}
		}()
		
		ebNoRecovery.SafeExecute("component4", func() error {
			panic("should propagate")
		})
	})
}

func TestSafeRender(t *testing.T) {
	eb := DefaultErrorBoundary()
	
	t.Run("Successful render", func(t *testing.T) {
		html := eb.SafeRender("component1", func() string {
			return "<div>Success</div>"
		})
		
		if html != "<div>Success</div>" {
			t.Errorf("Expected successful render, got %s", html)
		}
		
		if eb.GetErrorCount() != 0 {
			t.Errorf("Expected 0 errors, got %d", eb.GetErrorCount())
		}
	})
	
	t.Run("Panic during render", func(t *testing.T) {
		html := eb.SafeRender("component2", func() string {
			panic("render panic")
		})
		
		if !strings.Contains(html, "error-boundary") {
			t.Error("Expected fallback HTML on panic")
		}
		
		if eb.GetErrorCount() != 1 {
			t.Errorf("Expected 1 error to be recorded, got %d", eb.GetErrorCount())
		}
	})
	
	t.Run("Recovery disabled render", func(t *testing.T) {
		ebNoRecovery := NewErrorBoundary(100, false)
		
		html := ebNoRecovery.SafeRender("component3", func() string {
			return "<div>No recovery</div>"
		})
		
		if html != "<div>No recovery</div>" {
			t.Errorf("Expected direct render when recovery disabled, got %s", html)
		}
	})
}

func TestGetErrorStats(t *testing.T) {
	eb := DefaultErrorBoundary()
	
	// Add some errors
	eb.CatchError("component1", errors.New("error1"))
	eb.CatchError("component1", errors.New("error2"))
	eb.CatchError("component2", errors.New("error3"))
	
	// Manually add an unrecovered error for testing
	eb.mu.Lock()
	eb.errors = append(eb.errors, ComponentError{
		ComponentID: "component3",
		Error:       errors.New("unrecovered"),
		Recovered:   false,
	})
	eb.mu.Unlock()
	
	stats := eb.GetErrorStats()
	
	if stats.TotalErrors != 4 {
		t.Errorf("Expected TotalErrors to be 4, got %d", stats.TotalErrors)
	}
	
	if stats.RecoveredErrors != 3 {
		t.Errorf("Expected RecoveredErrors to be 3, got %d", stats.RecoveredErrors)
	}
	
	if stats.ComponentErrors["component1"] != 2 {
		t.Errorf("Expected 2 errors for component1, got %d", stats.ComponentErrors["component1"])
	}
	
	if stats.ComponentErrors["component2"] != 1 {
		t.Errorf("Expected 1 error for component2, got %d", stats.ComponentErrors["component2"])
	}
	
	if len(stats.RecentErrors) != 4 {
		t.Errorf("Expected 4 recent errors, got %d", len(stats.RecentErrors))
	}
}

func TestConfigureErrorBoundary(t *testing.T) {
	config := &ErrorBoundaryConfig{
		MaxErrors:          50,
		RecoveryEnabled:    false,
		LogErrors:          true,
		SendErrorsToClient: true,
		CustomFallbackHTML: "<div>Custom</div>",
		RetryAttempts:      3,
		RetryDelayMs:       500,
	}
	
	eb := ConfigureErrorBoundary(config)
	
	if eb.maxErrors != 50 {
		t.Errorf("Expected maxErrors to be 50, got %d", eb.maxErrors)
	}
	
	if eb.recoveryEnabled {
		t.Error("Expected recoveryEnabled to be false")
	}
	
	html := eb.RenderFallback("test", errors.New("error"))
	if html != "<div>Custom</div>" {
		t.Errorf("Expected custom fallback HTML, got %s", html)
	}
}

func TestConcurrentErrorHandling(t *testing.T) {
	eb := DefaultErrorBoundary()
	
	var wg sync.WaitGroup
	numGoroutines := 100
	
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			eb.CatchError(fmt.Sprintf("component%d", id), fmt.Errorf("error %d", id))
		}(i)
	}
	
	wg.Wait()
	
	// With maxErrors=100, we should have exactly 100 errors
	if eb.GetErrorCount() != 100 {
		t.Errorf("Expected 100 errors after concurrent operations, got %d", eb.GetErrorCount())
	}
}

func TestWrapDriverWithErrorBoundary(t *testing.T) {
	eb := DefaultErrorBoundary()
	
	// Create a test component driver
	driver := &ComponentDriver[*TestComponent]{
		channel:     make(chan map[string]interface{}, 1),
		IdComponent: "test-component",
	}
	
	wrapped := WrapDriverWithErrorBoundary(driver, eb)
	
	if wrapped == nil {
		t.Fatal("WrapDriverWithErrorBoundary returned nil")
	}
	
	if wrapped.errorBoundary != eb {
		t.Error("Error boundary was not set on driver")
	}
}

func TestSafeCommit(t *testing.T) {
	eb := DefaultErrorBoundary()
	
	// Create a test component driver
	driver := &ComponentDriver[*TestComponent]{
		channel:     make(chan map[string]interface{}, 1),
		IdComponent: "test-component",
	}
	
	// Add error boundary
	driver.WithErrorBoundary(eb)
	
	// Test SafeCommit - this would normally call Commit()
	// Since we don't have a full component setup, we just verify the structure
	if driver.errorBoundary == nil {
		t.Error("WithErrorBoundary did not set error boundary")
	}
}

// mockLiveDriver for testing
type mockLiveDriver struct {
	onCall func()
}

func (m *mockLiveDriver) GetIDComponet() string {
	if m.onCall != nil {
		m.onCall()
	}
	return "mock-component"
}

func (m *mockLiveDriver) SetHTML(id string, html string) {
	if m.onCall != nil {
		m.onCall()
	}
}

func (m *mockLiveDriver) SetText(id string, text string) {
	if m.onCall != nil {
		m.onCall()
	}
}

func (m *mockLiveDriver) SetValue(value interface{}) {
	if m.onCall != nil {
		m.onCall()
	}
}

func (m *mockLiveDriver) SetStyle(style string) {
	if m.onCall != nil {
		m.onCall()
	}
}

func (m *mockLiveDriver) EvalScript(script string) {
	if m.onCall != nil {
		m.onCall()
	}
}

func (m *mockLiveDriver) Mount(s ...interface{}) interface{} {
	if m.onCall != nil {
		m.onCall()
	}
	return nil
}

func (m *mockLiveDriver) Commit() {
	if m.onCall != nil {
		m.onCall()
	}
}

func (m *mockLiveDriver) FillValue(s ...interface{}) interface{} {
	if m.onCall != nil {
		m.onCall()
	}
	return nil
}

func (m *mockLiveDriver) Start() {
	if m.onCall != nil {
		m.onCall()
	}
}

func TestGetCurrentTimestamp(t *testing.T) {
	// Save original timeNow
	originalTimeNow := timeNow
	defer func() {
		timeNow = originalTimeNow
	}()
	
	// Mock timeNow
	mockTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	timeNow = func() time.Time {
		return mockTime
	}
	
	timestamp := getCurrentTimestamp()
	expectedTimestamp := mockTime.UnixNano() / 1e6
	
	if timestamp != expectedTimestamp {
		t.Errorf("Expected timestamp %d, got %d", expectedTimestamp, timestamp)
	}
}