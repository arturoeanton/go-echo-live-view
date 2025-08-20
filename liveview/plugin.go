package liveview

import (
	"context"
	"fmt"
	"reflect"
	"sync"
)

// Plugin represents a plugin that can extend the framework
type Plugin interface {
	// Name returns the unique name of the plugin
	Name() string
	
	// Version returns the version of the plugin
	Version() string
	
	// Initialize is called when the plugin is registered
	Initialize(ctx context.Context) error
	
	// Shutdown is called when the plugin is being unregistered
	Shutdown() error
}

// Middleware represents a function that can intercept and modify behavior
type Middleware func(next HandlerFunc) HandlerFunc

// HandlerFunc represents a generic handler function
type HandlerFunc func(ctx context.Context, data interface{}) error

// PluginManager manages plugins and middleware
type PluginManager struct {
	mu         sync.RWMutex
	plugins    map[string]Plugin
	middleware []Middleware
	hooks      map[string][]HookFunc
	config     *PluginConfig
	ctx        context.Context
	cancel     context.CancelFunc
}

// HookFunc represents a hook function
type HookFunc func(ctx context.Context, args ...interface{}) error

// PluginConfig configures the plugin manager
type PluginConfig struct {
	MaxPlugins     int
	AllowOverwrite bool
	AutoInitialize bool
	EnableHooks    bool
}

// DefaultPluginConfig returns default plugin configuration
func DefaultPluginConfig() *PluginConfig {
	return &PluginConfig{
		MaxPlugins:     100,
		AllowOverwrite: false,
		AutoInitialize: true,
		EnableHooks:    true,
	}
}

// NewPluginManager creates a new plugin manager
func NewPluginManager(config *PluginConfig) *PluginManager {
	if config == nil {
		config = DefaultPluginConfig()
	}
	
	ctx, cancel := context.WithCancel(context.Background())
	
	return &PluginManager{
		plugins:    make(map[string]Plugin),
		middleware: make([]Middleware, 0),
		hooks:      make(map[string][]HookFunc),
		config:     config,
		ctx:        ctx,
		cancel:     cancel,
	}
}

// Register registers a new plugin
func (pm *PluginManager) Register(plugin Plugin) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	
	if plugin == nil {
		return fmt.Errorf("plugin cannot be nil")
	}
	
	name := plugin.Name()
	if name == "" {
		return fmt.Errorf("plugin name cannot be empty")
	}
	
	// Check if plugin already exists
	if _, exists := pm.plugins[name]; exists && !pm.config.AllowOverwrite {
		return fmt.Errorf("plugin %s already registered", name)
	}
	
	// Check max plugins limit
	if len(pm.plugins) >= pm.config.MaxPlugins {
		return fmt.Errorf("maximum number of plugins (%d) reached", pm.config.MaxPlugins)
	}
	
	// Initialize plugin if auto-initialize is enabled
	if pm.config.AutoInitialize {
		if err := plugin.Initialize(pm.ctx); err != nil {
			return fmt.Errorf("failed to initialize plugin %s: %w", name, err)
		}
	}
	
	pm.plugins[name] = plugin
	
	// Trigger registered hook
	pm.triggerHook("plugin.registered", plugin)
	
	Debug("Plugin %s v%s registered", name, plugin.Version())
	return nil
}

// Unregister removes a plugin
func (pm *PluginManager) Unregister(name string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	
	plugin, exists := pm.plugins[name]
	if !exists {
		return fmt.Errorf("plugin %s not found", name)
	}
	
	// Shutdown plugin
	if err := plugin.Shutdown(); err != nil {
		Debug("Error shutting down plugin %s: %v", name, err)
	}
	
	delete(pm.plugins, name)
	
	// Trigger unregistered hook
	pm.triggerHook("plugin.unregistered", name)
	
	Debug("Plugin %s unregistered", name)
	return nil
}

// Get retrieves a plugin by name
func (pm *PluginManager) Get(name string) (Plugin, bool) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	
	plugin, exists := pm.plugins[name]
	return plugin, exists
}

// List returns all registered plugins
func (pm *PluginManager) List() []Plugin {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	
	plugins := make([]Plugin, 0, len(pm.plugins))
	for _, plugin := range pm.plugins {
		plugins = append(plugins, plugin)
	}
	return plugins
}

// Use adds a middleware to the chain
func (pm *PluginManager) Use(middleware Middleware) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	
	pm.middleware = append(pm.middleware, middleware)
	
	// Trigger middleware added hook
	pm.triggerHook("middleware.added", middleware)
}

// Execute runs a handler through the middleware chain
func (pm *PluginManager) Execute(ctx context.Context, handler HandlerFunc, data interface{}) error {
	pm.mu.RLock()
	middlewares := make([]Middleware, len(pm.middleware))
	copy(middlewares, pm.middleware)
	pm.mu.RUnlock()
	
	// Build the handler chain
	finalHandler := handler
	for i := len(middlewares) - 1; i >= 0; i-- {
		finalHandler = middlewares[i](finalHandler)
	}
	
	// Execute the final handler
	return finalHandler(ctx, data)
}

// RegisterHook registers a hook function
func (pm *PluginManager) RegisterHook(name string, hook HookFunc) {
	if !pm.config.EnableHooks {
		return
	}
	
	pm.mu.Lock()
	defer pm.mu.Unlock()
	
	if pm.hooks[name] == nil {
		pm.hooks[name] = make([]HookFunc, 0)
	}
	pm.hooks[name] = append(pm.hooks[name], hook)
}

// TriggerHook triggers all hooks for a given name
func (pm *PluginManager) TriggerHook(name string, args ...interface{}) error {
	if !pm.config.EnableHooks {
		return nil
	}
	
	pm.mu.RLock()
	hooks := pm.hooks[name]
	pm.mu.RUnlock()
	
	for _, hook := range hooks {
		if err := hook(pm.ctx, args...); err != nil {
			return fmt.Errorf("hook %s failed: %w", name, err)
		}
	}
	
	return nil
}

// triggerHook internal hook trigger without lock
func (pm *PluginManager) triggerHook(name string, args ...interface{}) {
	if !pm.config.EnableHooks {
		return
	}
	
	hooks := pm.hooks[name]
	for _, hook := range hooks {
		if err := hook(pm.ctx, args...); err != nil {
			Debug("Hook %s error: %v", name, err)
		}
	}
}

// Shutdown shuts down all plugins
func (pm *PluginManager) Shutdown() error {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	
	// Cancel context
	pm.cancel()
	
	// Shutdown all plugins
	var errors []error
	for name, plugin := range pm.plugins {
		if err := plugin.Shutdown(); err != nil {
			errors = append(errors, fmt.Errorf("failed to shutdown plugin %s: %w", name, err))
		}
	}
	
	// Clear all data
	pm.plugins = make(map[string]Plugin)
	pm.middleware = make([]Middleware, 0)
	pm.hooks = make(map[string][]HookFunc)
	
	if len(errors) > 0 {
		return fmt.Errorf("shutdown errors: %v", errors)
	}
	
	return nil
}

// ComponentPlugin is a plugin that provides a component
type ComponentPlugin interface {
	Plugin
	// GetComponent returns the component provided by this plugin
	GetComponent() Component
}

// MiddlewarePlugin is a plugin that provides middleware
type MiddlewarePlugin interface {
	Plugin
	// GetMiddleware returns the middleware provided by this plugin
	GetMiddleware() []Middleware
}

// HookPlugin is a plugin that registers hooks
type HookPlugin interface {
	Plugin
	// GetHooks returns the hooks provided by this plugin
	GetHooks() map[string]HookFunc
}

// CompositePlugin combines multiple plugin types
type CompositePlugin struct {
	name       string
	version    string
	component  Component
	middleware []Middleware
	hooks      map[string]HookFunc
	onInit     func(context.Context) error
	onShutdown func() error
}

// NewCompositePlugin creates a new composite plugin
func NewCompositePlugin(name, version string) *CompositePlugin {
	return &CompositePlugin{
		name:       name,
		version:    version,
		middleware: make([]Middleware, 0),
		hooks:      make(map[string]HookFunc),
	}
}

// Name returns the plugin name
func (cp *CompositePlugin) Name() string {
	return cp.name
}

// Version returns the plugin version
func (cp *CompositePlugin) Version() string {
	return cp.version
}

// Initialize initializes the plugin
func (cp *CompositePlugin) Initialize(ctx context.Context) error {
	if cp.onInit != nil {
		return cp.onInit(ctx)
	}
	return nil
}

// Shutdown shuts down the plugin
func (cp *CompositePlugin) Shutdown() error {
	if cp.onShutdown != nil {
		return cp.onShutdown()
	}
	return nil
}

// WithComponent adds a component to the plugin
func (cp *CompositePlugin) WithComponent(component Component) *CompositePlugin {
	cp.component = component
	return cp
}

// WithMiddleware adds middleware to the plugin
func (cp *CompositePlugin) WithMiddleware(middleware ...Middleware) *CompositePlugin {
	cp.middleware = append(cp.middleware, middleware...)
	return cp
}

// WithHook adds a hook to the plugin
func (cp *CompositePlugin) WithHook(name string, hook HookFunc) *CompositePlugin {
	cp.hooks[name] = hook
	return cp
}

// WithInitializer sets the initialization function
func (cp *CompositePlugin) WithInitializer(init func(context.Context) error) *CompositePlugin {
	cp.onInit = init
	return cp
}

// WithShutdown sets the shutdown function
func (cp *CompositePlugin) WithShutdown(shutdown func() error) *CompositePlugin {
	cp.onShutdown = shutdown
	return cp
}

// GetComponent returns the component if this is a ComponentPlugin
func (cp *CompositePlugin) GetComponent() Component {
	return cp.component
}

// GetMiddleware returns the middleware if this is a MiddlewarePlugin
func (cp *CompositePlugin) GetMiddleware() []Middleware {
	return cp.middleware
}

// GetHooks returns the hooks if this is a HookPlugin
func (cp *CompositePlugin) GetHooks() map[string]HookFunc {
	return cp.hooks
}

// PluginRegistry is a global registry for plugins
type PluginRegistry struct {
	mu       sync.RWMutex
	managers map[string]*PluginManager
}

// GlobalPluginRegistry is the global plugin registry
var GlobalPluginRegistry = &PluginRegistry{
	managers: make(map[string]*PluginManager),
}

// CreateManager creates a new plugin manager with a name
func (pr *PluginRegistry) CreateManager(name string, config *PluginConfig) (*PluginManager, error) {
	pr.mu.Lock()
	defer pr.mu.Unlock()
	
	if _, exists := pr.managers[name]; exists {
		return nil, fmt.Errorf("manager %s already exists", name)
	}
	
	manager := NewPluginManager(config)
	pr.managers[name] = manager
	return manager, nil
}

// GetManager retrieves a plugin manager by name
func (pr *PluginRegistry) GetManager(name string) (*PluginManager, bool) {
	pr.mu.RLock()
	defer pr.mu.RUnlock()
	
	manager, exists := pr.managers[name]
	return manager, exists
}

// RemoveManager removes a plugin manager
func (pr *PluginRegistry) RemoveManager(name string) error {
	pr.mu.Lock()
	defer pr.mu.Unlock()
	
	manager, exists := pr.managers[name]
	if !exists {
		return fmt.Errorf("manager %s not found", name)
	}
	
	// Shutdown the manager
	if err := manager.Shutdown(); err != nil {
		return err
	}
	
	delete(pr.managers, name)
	return nil
}

// PluginInjector provides dependency injection for plugins
type PluginInjector struct {
	providers map[reflect.Type]interface{}
	mu        sync.RWMutex
}

// NewPluginInjector creates a new plugin injector
func NewPluginInjector() *PluginInjector {
	return &PluginInjector{
		providers: make(map[reflect.Type]interface{}),
	}
}

// Provide registers a provider for a type
func (pi *PluginInjector) Provide(instance interface{}) {
	pi.mu.Lock()
	defer pi.mu.Unlock()
	
	typ := reflect.TypeOf(instance)
	pi.providers[typ] = instance
}

// Inject injects dependencies into a struct
func (pi *PluginInjector) Inject(target interface{}) error {
	pi.mu.RLock()
	defer pi.mu.RUnlock()
	
	targetValue := reflect.ValueOf(target)
	if targetValue.Kind() != reflect.Ptr {
		return fmt.Errorf("target must be a pointer")
	}
	
	targetValue = targetValue.Elem()
	if targetValue.Kind() != reflect.Struct {
		return fmt.Errorf("target must be a pointer to struct")
	}
	
	targetType := targetValue.Type()
	
	for i := 0; i < targetValue.NumField(); i++ {
		field := targetValue.Field(i)
		fieldType := targetType.Field(i)
		
		// Check for inject tag
		if tag := fieldType.Tag.Get("inject"); tag == "true" {
			// Find provider for this type
			if provider, exists := pi.providers[field.Type()]; exists {
				field.Set(reflect.ValueOf(provider))
			}
		}
	}
	
	return nil
}