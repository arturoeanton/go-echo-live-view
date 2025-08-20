package liveview

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"
)

// TestPlugin is a simple test plugin
type TestPlugin struct {
	name         string
	version      string
	initialized  bool
	shutdown     bool
	initError    error
	shutdownError error
}

func (tp *TestPlugin) Name() string {
	return tp.name
}

func (tp *TestPlugin) Version() string {
	return tp.version
}

func (tp *TestPlugin) Initialize(ctx context.Context) error {
	tp.initialized = true
	return tp.initError
}

func (tp *TestPlugin) Shutdown() error {
	tp.shutdown = true
	return tp.shutdownError
}

func TestNewPluginManager(t *testing.T) {
	pm := NewPluginManager(nil)
	
	if pm == nil {
		t.Fatal("NewPluginManager returned nil")
	}
	
	if pm.config == nil {
		t.Error("Expected default config to be set")
	}
	
	if len(pm.plugins) != 0 {
		t.Error("Expected empty plugins map")
	}
	
	if len(pm.middleware) != 0 {
		t.Error("Expected empty middleware slice")
	}
}

func TestPluginManagerRegister(t *testing.T) {
	pm := NewPluginManager(nil)
	
	plugin := &TestPlugin{
		name:    "test-plugin",
		version: "1.0.0",
	}
	
	err := pm.Register(plugin)
	if err != nil {
		t.Errorf("Register() error = %v", err)
	}
	
	if !plugin.initialized {
		t.Error("Plugin was not initialized")
	}
	
	// Try to register same plugin again
	err = pm.Register(plugin)
	if err == nil {
		t.Error("Expected error when registering duplicate plugin")
	}
	
	// Test with nil plugin
	err = pm.Register(nil)
	if err == nil {
		t.Error("Expected error when registering nil plugin")
	}
	
	// Test with empty name
	emptyNamePlugin := &TestPlugin{name: "", version: "1.0.0"}
	err = pm.Register(emptyNamePlugin)
	if err == nil {
		t.Error("Expected error when registering plugin with empty name")
	}
}

func TestPluginManagerRegisterWithError(t *testing.T) {
	pm := NewPluginManager(nil)
	
	plugin := &TestPlugin{
		name:      "error-plugin",
		version:   "1.0.0",
		initError: errors.New("init failed"),
	}
	
	err := pm.Register(plugin)
	if err == nil {
		t.Error("Expected error when plugin initialization fails")
	}
}

func TestPluginManagerUnregister(t *testing.T) {
	pm := NewPluginManager(nil)
	
	plugin := &TestPlugin{
		name:    "test-plugin",
		version: "1.0.0",
	}
	
	pm.Register(plugin)
	
	err := pm.Unregister("test-plugin")
	if err != nil {
		t.Errorf("Unregister() error = %v", err)
	}
	
	if !plugin.shutdown {
		t.Error("Plugin was not shut down")
	}
	
	// Try to unregister non-existent plugin
	err = pm.Unregister("non-existent")
	if err == nil {
		t.Error("Expected error when unregistering non-existent plugin")
	}
}

func TestPluginManagerGet(t *testing.T) {
	pm := NewPluginManager(nil)
	
	plugin := &TestPlugin{
		name:    "test-plugin",
		version: "1.0.0",
	}
	
	pm.Register(plugin)
	
	retrieved, exists := pm.Get("test-plugin")
	if !exists {
		t.Error("Plugin should exist")
	}
	
	if retrieved != plugin {
		t.Error("Retrieved plugin doesn't match original")
	}
	
	_, exists = pm.Get("non-existent")
	if exists {
		t.Error("Non-existent plugin should not exist")
	}
}

func TestPluginManagerList(t *testing.T) {
	pm := NewPluginManager(nil)
	
	plugin1 := &TestPlugin{name: "plugin1", version: "1.0.0"}
	plugin2 := &TestPlugin{name: "plugin2", version: "1.0.0"}
	
	pm.Register(plugin1)
	pm.Register(plugin2)
	
	plugins := pm.List()
	if len(plugins) != 2 {
		t.Errorf("Expected 2 plugins, got %d", len(plugins))
	}
}

func TestPluginManagerMaxPlugins(t *testing.T) {
	config := &PluginConfig{
		MaxPlugins:     2,
		AutoInitialize: false,
	}
	pm := NewPluginManager(config)
	
	pm.Register(&TestPlugin{name: "plugin1", version: "1.0.0"})
	pm.Register(&TestPlugin{name: "plugin2", version: "1.0.0"})
	
	err := pm.Register(&TestPlugin{name: "plugin3", version: "1.0.0"})
	if err == nil {
		t.Error("Expected error when exceeding max plugins limit")
	}
}

func TestPluginManagerAllowOverwrite(t *testing.T) {
	config := &PluginConfig{
		MaxPlugins:     10,
		AllowOverwrite: true,
		AutoInitialize: false,
	}
	pm := NewPluginManager(config)
	
	plugin1 := &TestPlugin{name: "plugin", version: "1.0.0"}
	plugin2 := &TestPlugin{name: "plugin", version: "2.0.0"}
	
	err := pm.Register(plugin1)
	if err != nil {
		t.Errorf("Failed to register first plugin: %v", err)
	}
	
	err = pm.Register(plugin2)
	if err != nil {
		t.Errorf("Should allow overwrite when configured: %v", err)
	}
	
	retrieved, exists := pm.Get("plugin")
	if !exists {
		t.Fatal("Plugin should exist after overwrite")
	}
	
	if retrieved.Version() != "2.0.0" {
		t.Error("Plugin was not overwritten")
	}
}

func TestMiddleware(t *testing.T) {
	pm := NewPluginManager(nil)
	
	var executionOrder []string
	
	middleware1 := func(next HandlerFunc) HandlerFunc {
		return func(ctx context.Context, data interface{}) error {
			executionOrder = append(executionOrder, "middleware1-before")
			err := next(ctx, data)
			executionOrder = append(executionOrder, "middleware1-after")
			return err
		}
	}
	
	middleware2 := func(next HandlerFunc) HandlerFunc {
		return func(ctx context.Context, data interface{}) error {
			executionOrder = append(executionOrder, "middleware2-before")
			err := next(ctx, data)
			executionOrder = append(executionOrder, "middleware2-after")
			return err
		}
	}
	
	pm.Use(middleware1)
	pm.Use(middleware2)
	
	handler := func(ctx context.Context, data interface{}) error {
		executionOrder = append(executionOrder, "handler")
		return nil
	}
	
	err := pm.Execute(context.Background(), handler, nil)
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}
	
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
			t.Errorf("Execution order[%d]: expected %s, got %s", i, exp, executionOrder[i])
		}
	}
}

func TestHooks(t *testing.T) {
	pm := NewPluginManager(nil)
	
	var hookCalled bool
	var hookArgs []interface{}
	
	hook := func(ctx context.Context, args ...interface{}) error {
		hookCalled = true
		hookArgs = args
		return nil
	}
	
	pm.RegisterHook("test.hook", hook)
	
	err := pm.TriggerHook("test.hook", "arg1", "arg2")
	if err != nil {
		t.Errorf("TriggerHook() error = %v", err)
	}
	
	if !hookCalled {
		t.Error("Hook was not called")
	}
	
	if len(hookArgs) != 2 {
		t.Errorf("Expected 2 hook args, got %d", len(hookArgs))
	}
}

func TestHooksWithError(t *testing.T) {
	pm := NewPluginManager(nil)
	
	hookError := errors.New("hook failed")
	hook := func(ctx context.Context, args ...interface{}) error {
		return hookError
	}
	
	pm.RegisterHook("error.hook", hook)
	
	err := pm.TriggerHook("error.hook")
	if err == nil {
		t.Error("Expected error from hook")
	}
}

func TestHooksDisabled(t *testing.T) {
	config := &PluginConfig{
		EnableHooks: false,
	}
	pm := NewPluginManager(config)
	
	var hookCalled bool
	hook := func(ctx context.Context, args ...interface{}) error {
		hookCalled = true
		return nil
	}
	
	pm.RegisterHook("test.hook", hook)
	pm.TriggerHook("test.hook")
	
	if hookCalled {
		t.Error("Hook should not be called when hooks are disabled")
	}
}

func TestPluginManagerShutdown(t *testing.T) {
	pm := NewPluginManager(nil)
	
	plugin1 := &TestPlugin{name: "plugin1", version: "1.0.0"}
	plugin2 := &TestPlugin{name: "plugin2", version: "1.0.0"}
	
	pm.Register(plugin1)
	pm.Register(plugin2)
	
	err := pm.Shutdown()
	if err != nil {
		t.Errorf("Shutdown() error = %v", err)
	}
	
	if !plugin1.shutdown {
		t.Error("Plugin1 was not shut down")
	}
	
	if !plugin2.shutdown {
		t.Error("Plugin2 was not shut down")
	}
	
	if len(pm.plugins) != 0 {
		t.Error("Plugins were not cleared")
	}
}

func TestCompositePlugin(t *testing.T) {
	cp := NewCompositePlugin("composite", "1.0.0")
	
	if cp.Name() != "composite" {
		t.Errorf("Expected name 'composite', got %s", cp.Name())
	}
	
	if cp.Version() != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got %s", cp.Version())
	}
	
	// Test with component
	component := &TestComponent{}
	cp.WithComponent(component)
	
	if cp.GetComponent() != component {
		t.Error("Component not set correctly")
	}
	
	// Test with middleware
	middleware := func(next HandlerFunc) HandlerFunc {
		return next
	}
	cp.WithMiddleware(middleware)
	
	if len(cp.GetMiddleware()) != 1 {
		t.Error("Middleware not added")
	}
	
	// Test with hook
	hook := func(ctx context.Context, args ...interface{}) error {
		return nil
	}
	cp.WithHook("test.hook", hook)
	
	if len(cp.GetHooks()) != 1 {
		t.Error("Hook not added")
	}
	
	// Test initializer
	var initCalled bool
	cp.WithInitializer(func(ctx context.Context) error {
		initCalled = true
		return nil
	})
	
	cp.Initialize(context.Background())
	if !initCalled {
		t.Error("Initializer not called")
	}
	
	// Test shutdown
	var shutdownCalled bool
	cp.WithShutdown(func() error {
		shutdownCalled = true
		return nil
	})
	
	cp.Shutdown()
	if !shutdownCalled {
		t.Error("Shutdown not called")
	}
}

func TestPluginRegistry(t *testing.T) {
	registry := &PluginRegistry{
		managers: make(map[string]*PluginManager),
	}
	
	// Create manager
	manager, err := registry.CreateManager("test-manager", nil)
	if err != nil {
		t.Errorf("CreateManager() error = %v", err)
	}
	
	if manager == nil {
		t.Fatal("Manager is nil")
	}
	
	// Try to create duplicate
	_, err = registry.CreateManager("test-manager", nil)
	if err == nil {
		t.Error("Expected error when creating duplicate manager")
	}
	
	// Get manager
	retrieved, exists := registry.GetManager("test-manager")
	if !exists {
		t.Error("Manager should exist")
	}
	
	if retrieved != manager {
		t.Error("Retrieved manager doesn't match original")
	}
	
	// Remove manager
	err = registry.RemoveManager("test-manager")
	if err != nil {
		t.Errorf("RemoveManager() error = %v", err)
	}
	
	_, exists = registry.GetManager("test-manager")
	if exists {
		t.Error("Manager should not exist after removal")
	}
}

func TestPluginInjector(t *testing.T) {
	injector := NewPluginInjector()
	
	// Provide a service
	service := &TestService{Value: "test-value"}
	injector.Provide(service)
	
	// Target struct with injection
	target := &TestTarget{}
	
	err := injector.Inject(target)
	if err != nil {
		t.Errorf("Inject() error = %v", err)
	}
	
	if target.Service != service {
		t.Error("Service was not injected")
	}
	
	// Test with non-pointer
	err = injector.Inject(TestTarget{})
	if err == nil {
		t.Error("Expected error when injecting into non-pointer")
	}
	
	// Test with non-struct pointer
	var str string
	err = injector.Inject(&str)
	if err == nil {
		t.Error("Expected error when injecting into non-struct")
	}
}

// Test types for injection
type TestService struct {
	Value string
}

type TestTarget struct {
	Service *TestService `inject:"true"`
	Other   string       `inject:"false"`
}

func TestConcurrentPluginOperations(t *testing.T) {
	pm := NewPluginManager(nil)
	
	var wg sync.WaitGroup
	numGoroutines := 100
	
	// Concurrent registrations
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			plugin := &TestPlugin{
				name:    fmt.Sprintf("plugin%d", id),
				version: "1.0.0",
			}
			pm.Register(plugin)
		}(i)
	}
	
	wg.Wait()
	
	plugins := pm.List()
	// Due to max plugins limit, we might not have all 100
	if len(plugins) == 0 {
		t.Error("No plugins registered")
	}
	
	// Concurrent middleware execution
	var executionCount int
	var mu sync.Mutex
	
	middleware := func(next HandlerFunc) HandlerFunc {
		return func(ctx context.Context, data interface{}) error {
			mu.Lock()
			executionCount++
			mu.Unlock()
			return next(ctx, data)
		}
	}
	
	pm.Use(middleware)
	
	handler := func(ctx context.Context, data interface{}) error {
		return nil
	}
	
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			pm.Execute(context.Background(), handler, nil)
		}()
	}
	
	wg.Wait()
	
	if executionCount != numGoroutines {
		t.Errorf("Expected %d executions, got %d", numGoroutines, executionCount)
	}
}

func TestMiddlewareErrorPropagation(t *testing.T) {
	pm := NewPluginManager(nil)
	
	expectedError := errors.New("handler error")
	
	middleware := func(next HandlerFunc) HandlerFunc {
		return func(ctx context.Context, data interface{}) error {
			err := next(ctx, data)
			if err != nil {
				// Middleware can handle or propagate error
				return fmt.Errorf("wrapped: %w", err)
			}
			return nil
		}
	}
	
	pm.Use(middleware)
	
	handler := func(ctx context.Context, data interface{}) error {
		return expectedError
	}
	
	err := pm.Execute(context.Background(), handler, nil)
	if err == nil {
		t.Error("Expected error to propagate")
	}
	
	if !errors.Is(err, expectedError) {
		t.Error("Error chain broken")
	}
}

func TestPluginManagerContext(t *testing.T) {
	pm := NewPluginManager(nil)
	
	// Register a plugin that uses context
	plugin := &ContextAwarePlugin{
		TestPlugin: TestPlugin{
			name:    "context-plugin",
			version: "1.0.0",
		},
	}
	
	err := pm.Register(plugin)
	if err != nil {
		t.Errorf("Register() error = %v", err)
	}
	
	// Shutdown should cancel context
	pm.Shutdown()
	
	// Give time for context cancellation to propagate
	time.Sleep(10 * time.Millisecond)
	
	if !plugin.contextCancelled {
		t.Error("Context was not cancelled on shutdown")
	}
}

// ContextAwarePlugin tests context cancellation
type ContextAwarePlugin struct {
	TestPlugin
	contextCancelled bool
}

func (cap *ContextAwarePlugin) Initialize(ctx context.Context) error {
	go func() {
		<-ctx.Done()
		cap.contextCancelled = true
	}()
	return cap.TestPlugin.Initialize(ctx)
}

func TestPluginDependencies(t *testing.T) {
	pm := NewPluginManager(nil)
	
	// Create plugins with dependencies
	plugin1 := &TestPlugin{name: "plugin1", version: "1.0.0"}
	plugin2 := &DependentPlugin{
		TestPlugin: TestPlugin{name: "plugin2", version: "1.0.0"},
		dependency: "plugin1",
		manager:    pm,
	}
	
	// Register in correct order
	pm.Register(plugin1)
	err := pm.Register(plugin2)
	
	if err != nil {
		t.Errorf("Failed to register dependent plugin: %v", err)
	}
	
	if !plugin2.dependencyResolved {
		t.Error("Dependency was not resolved")
	}
}

// DependentPlugin has a dependency on another plugin
type DependentPlugin struct {
	TestPlugin
	dependency         string
	dependencyResolved bool
	manager            *PluginManager
}

func (dp *DependentPlugin) Initialize(ctx context.Context) error {
	// Check if dependency exists
	if dp.manager != nil {
		if _, exists := dp.manager.Get(dp.dependency); exists {
			dp.dependencyResolved = true
		}
	}
	return dp.TestPlugin.Initialize(ctx)
}

func TestPluginChaining(t *testing.T) {
	pm := NewPluginManager(nil)
	
	// Create a chain of middleware from plugins
	plugin1 := NewCompositePlugin("plugin1", "1.0.0").
		WithMiddleware(func(next HandlerFunc) HandlerFunc {
			return func(ctx context.Context, data interface{}) error {
				if d, ok := data.(*[]string); ok {
					*d = append(*d, "plugin1")
				}
				return next(ctx, data)
			}
		})
	
	plugin2 := NewCompositePlugin("plugin2", "1.0.0").
		WithMiddleware(func(next HandlerFunc) HandlerFunc {
			return func(ctx context.Context, data interface{}) error {
				if d, ok := data.(*[]string); ok {
					*d = append(*d, "plugin2")
				}
				return next(ctx, data)
			}
		})
	
	pm.Register(plugin1)
	pm.Register(plugin2)
	
	// Apply middleware from plugins
	for _, middleware := range plugin1.GetMiddleware() {
		pm.Use(middleware)
	}
	for _, middleware := range plugin2.GetMiddleware() {
		pm.Use(middleware)
	}
	
	// Execute with data collection
	var data []string
	handler := func(ctx context.Context, d interface{}) error {
		if data, ok := d.(*[]string); ok {
			*data = append(*data, "handler")
		}
		return nil
	}
	
	pm.Execute(context.Background(), handler, &data)
	
	expected := []string{"plugin1", "plugin2", "handler"}
	if len(data) != len(expected) {
		t.Fatalf("Expected %d items, got %d", len(expected), len(data))
	}
	
	for i, exp := range expected {
		if data[i] != exp {
			t.Errorf("Data[%d]: expected %s, got %s", i, exp, data[i])
		}
	}
}

func TestProviderTypeMatching(t *testing.T) {
	injector := NewPluginInjector()
	
	// Test exact type matching
	service := &TestService{Value: "exact"}
	injector.Provide(service)
	
	// Should match exact type
	target := &TestTarget{}
	injector.Inject(target)
	
	if target.Service == nil {
		t.Error("Exact type match failed")
	}
	
	if target.Service.Value != "exact" {
		t.Errorf("Expected value 'exact', got %s", target.Service.Value)
	}
}

func TestPluginPriority(t *testing.T) {
	pm := NewPluginManager(nil)
	
	// Middleware should execute in order of registration
	var order []int
	
	for i := 0; i < 3; i++ {
		idx := i // Capture loop variable
		middleware := func(next HandlerFunc) HandlerFunc {
			return func(ctx context.Context, data interface{}) error {
				order = append(order, idx)
				return next(ctx, data)
			}
		}
		pm.Use(middleware)
	}
	
	handler := func(ctx context.Context, data interface{}) error {
		order = append(order, -1) // Handler marker
		return nil
	}
	
	pm.Execute(context.Background(), handler, nil)
	
	expected := []int{0, 1, 2, -1}
	if !reflect.DeepEqual(order, expected) {
		t.Errorf("Expected order %v, got %v", expected, order)
	}
}