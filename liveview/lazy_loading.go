package liveview

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"time"
)

// LazyLoader manages lazy loading of components
type LazyLoader struct {
	mu            sync.RWMutex
	registry      map[string]ComponentFactory
	loaded        map[string]Component
	loading       map[string]*LoadingState
	config        *LazyLoaderConfig
	interceptors  []LoadInterceptor
	errorHandlers []ErrorHandler
}

// ComponentFactory creates component instances
type ComponentFactory func() (Component, error)

// LoadInterceptor intercepts component loading
type LoadInterceptor func(name string, next func() (Component, error)) (Component, error)

// LoadingState tracks component loading state
type LoadingState struct {
	mu         sync.Mutex
	component  Component
	error      error
	loading    bool
	loadTime   time.Duration
	retryCount int
	callbacks  []func(Component, error)
}

// LazyLoaderConfig configures lazy loading
type LazyLoaderConfig struct {
	Preload          []string      // Components to preload
	MaxRetries       int           // Max load retries
	RetryDelay       time.Duration // Delay between retries
	LoadTimeout      time.Duration // Timeout for loading
	ConcurrentLoads  int           // Max concurrent loads
	EnableCaching    bool          // Cache loaded components
	EnableMetrics    bool          // Track loading metrics
	FallbackStrategy string        // Fallback strategy
}

// LazyComponent wraps a component for lazy loading
type LazyComponent struct {
	name        string
	loader      *LazyLoader
	component   Component
	placeholder Component
	loading     bool
	error       error
	mu          sync.RWMutex
}

// DefaultLazyLoaderConfig returns default configuration
func DefaultLazyLoaderConfig() *LazyLoaderConfig {
	return &LazyLoaderConfig{
		Preload:         make([]string, 0),
		MaxRetries:      3,
		RetryDelay:      1 * time.Second,
		LoadTimeout:     10 * time.Second,
		ConcurrentLoads: 10,
		EnableCaching:   true,
		EnableMetrics:   true,
		FallbackStrategy: "placeholder",
	}
}

// NewLazyLoader creates a new lazy loader
func NewLazyLoader(config *LazyLoaderConfig) *LazyLoader {
	if config == nil {
		config = DefaultLazyLoaderConfig()
	}

	ll := &LazyLoader{
		registry:      make(map[string]ComponentFactory),
		loaded:        make(map[string]Component),
		loading:       make(map[string]*LoadingState),
		config:        config,
		interceptors:  make([]LoadInterceptor, 0),
		errorHandlers: make([]ErrorHandler, 0),
	}

	// Preload components if configured
	if len(config.Preload) > 0 {
		go ll.preloadComponents()
	}

	return ll
}

// Register registers a component factory
func (ll *LazyLoader) Register(name string, factory ComponentFactory) error {
	ll.mu.Lock()
	defer ll.mu.Unlock()

	if _, exists := ll.registry[name]; exists {
		return fmt.Errorf("component %s already registered", name)
	}

	ll.registry[name] = factory
	Debug("Lazy component registered: %s", name)
	return nil
}

// Load loads a component lazily
func (ll *LazyLoader) Load(name string) (Component, error) {
	// Check if already loaded
	ll.mu.RLock()
	if component, exists := ll.loaded[name]; exists && ll.config.EnableCaching {
		ll.mu.RUnlock()
		return component, nil
	}
	ll.mu.RUnlock()

	// Get or create loading state
	ll.mu.Lock()
	state, exists := ll.loading[name]
	if !exists {
		state = &LoadingState{}
		ll.loading[name] = state
	}
	ll.mu.Unlock()

	// Lock the loading state
	state.mu.Lock()
	defer state.mu.Unlock()

	// Check if already loading
	if state.loading {
		// Wait for loading to complete
		return state.component, state.error
	}

	// Check if already loaded during wait
	if state.component != nil {
		return state.component, nil
	}

	// Start loading
	state.loading = true
	
	// Load with timeout
	ctx, cancel := context.WithTimeout(context.Background(), ll.config.LoadTimeout)
	defer cancel()

	component, err := ll.loadWithContext(ctx, name)
	
	state.loading = false
	state.component = component
	state.error = err

	// Cache if successful and caching enabled
	if err == nil && ll.config.EnableCaching {
		ll.mu.Lock()
		ll.loaded[name] = component
		ll.mu.Unlock()
	}

	// Notify callbacks
	for _, callback := range state.callbacks {
		callback(component, err)
	}
	state.callbacks = nil

	return component, err
}

// LoadAsync loads a component asynchronously
func (ll *LazyLoader) LoadAsync(name string, callback func(Component, error)) {
	go func() {
		component, err := ll.Load(name)
		if callback != nil {
			callback(component, err)
		}
	}()
}

// LoadWithFallback loads a component with fallback
func (ll *LazyLoader) LoadWithFallback(name string, fallback Component) Component {
	component, err := ll.Load(name)
	if err != nil {
		Debug("Failed to load %s, using fallback: %v", name, err)
		return fallback
	}
	return component
}

// loadWithContext loads a component with context
func (ll *LazyLoader) loadWithContext(ctx context.Context, name string) (Component, error) {
	ll.mu.RLock()
	factory, exists := ll.registry[name]
	ll.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("component %s not registered", name)
	}

	// Apply interceptors
	loadFunc := func() (Component, error) {
		return ll.loadWithRetry(factory)
	}

	for i := len(ll.interceptors) - 1; i >= 0; i-- {
		interceptor := ll.interceptors[i]
		prevLoad := loadFunc
		loadFunc = func() (Component, error) {
			return interceptor(name, prevLoad)
		}
	}

	// Load with context
	type result struct {
		component Component
		err       error
	}

	resultCh := make(chan result, 1)

	go func() {
		component, err := loadFunc()
		resultCh <- result{component, err}
	}()

	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("loading %s timed out", name)
	case res := <-resultCh:
		return res.component, res.err
	}
}

// loadWithRetry loads with retry logic
func (ll *LazyLoader) loadWithRetry(factory ComponentFactory) (Component, error) {
	var lastErr error
	
	for i := 0; i <= ll.config.MaxRetries; i++ {
		component, err := factory()
		if err == nil {
			return component, nil
		}

		lastErr = err
		
		if i < ll.config.MaxRetries {
			Debug("Retry %d/%d loading component: %v", i+1, ll.config.MaxRetries, err)
			time.Sleep(ll.config.RetryDelay)
		}
	}

	return nil, fmt.Errorf("failed after %d retries: %w", ll.config.MaxRetries, lastErr)
}

// preloadComponents preloads configured components
func (ll *LazyLoader) preloadComponents() {
	for _, name := range ll.config.Preload {
		Debug("Preloading component: %s", name)
		_, err := ll.Load(name)
		if err != nil {
			Debug("Failed to preload %s: %v", name, err)
		}
	}
}

// AddInterceptor adds a load interceptor
func (ll *LazyLoader) AddInterceptor(interceptor LoadInterceptor) {
	ll.interceptors = append(ll.interceptors, interceptor)
}

// IsLoaded checks if a component is loaded
func (ll *LazyLoader) IsLoaded(name string) bool {
	ll.mu.RLock()
	defer ll.mu.RUnlock()
	
	_, exists := ll.loaded[name]
	return exists
}

// Unload unloads a cached component
func (ll *LazyLoader) Unload(name string) {
	ll.mu.Lock()
	defer ll.mu.Unlock()
	
	delete(ll.loaded, name)
	delete(ll.loading, name)
}

// UnloadAll unloads all cached components
func (ll *LazyLoader) UnloadAll() {
	ll.mu.Lock()
	defer ll.mu.Unlock()
	
	ll.loaded = make(map[string]Component)
	ll.loading = make(map[string]*LoadingState)
}

// GetLoadedComponents returns all loaded components
func (ll *LazyLoader) GetLoadedComponents() map[string]Component {
	ll.mu.RLock()
	defer ll.mu.RUnlock()
	
	result := make(map[string]Component)
	for k, v := range ll.loaded {
		result[k] = v
	}
	return result
}

// NewLazyComponent creates a new lazy component
func NewLazyComponent(name string, loader *LazyLoader, placeholder Component) *LazyComponent {
	return &LazyComponent{
		name:        name,
		loader:      loader,
		placeholder: placeholder,
	}
}

// GetDriver implements Component interface
func (lc *LazyComponent) GetDriver() LiveDriver {
	component := lc.getComponent()
	if component != nil {
		return component.GetDriver()
	}
	if lc.placeholder != nil {
		return lc.placeholder.GetDriver()
	}
	return nil
}

// GetTemplate implements Component interface
func (lc *LazyComponent) GetTemplate() string {
	component := lc.getComponent()
	if component != nil {
		return component.GetTemplate()
	}
	if lc.placeholder != nil {
		return lc.placeholder.GetTemplate()
	}
	return "<div>Loading...</div>"
}

// Start implements Component interface
func (lc *LazyComponent) Start() {
	// Load component asynchronously
	lc.loader.LoadAsync(lc.name, func(component Component, err error) {
		lc.mu.Lock()
		defer lc.mu.Unlock()
		
		if err != nil {
			lc.error = err
			Debug("Failed to lazy load %s: %v", lc.name, err)
			return
		}
		
		lc.component = component
		lc.loading = false
		
		// Start the loaded component
		if component != nil {
			component.Start()
		}
	})
	
	// Start placeholder if available
	if lc.placeholder != nil {
		lc.placeholder.Start()
	}
}

// getComponent gets the loaded component
func (lc *LazyComponent) getComponent() Component {
	lc.mu.RLock()
	defer lc.mu.RUnlock()
	return lc.component
}

// IsLoaded checks if component is loaded
func (lc *LazyComponent) IsLoaded() bool {
	lc.mu.RLock()
	defer lc.mu.RUnlock()
	return lc.component != nil
}

// GetError returns loading error if any
func (lc *LazyComponent) GetError() error {
	lc.mu.RLock()
	defer lc.mu.RUnlock()
	return lc.error
}

// ComponentBundle bundles multiple components for lazy loading
type ComponentBundle struct {
	Name       string
	Components map[string]ComponentFactory
	Dependencies []string
}

// BundleLoader loads component bundles
type BundleLoader struct {
	loader  *LazyLoader
	bundles map[string]*ComponentBundle
	loaded  map[string]bool
	mu      sync.RWMutex
}

// NewBundleLoader creates a new bundle loader
func NewBundleLoader(loader *LazyLoader) *BundleLoader {
	return &BundleLoader{
		loader:  loader,
		bundles: make(map[string]*ComponentBundle),
		loaded:  make(map[string]bool),
	}
}

// RegisterBundle registers a component bundle
func (bl *BundleLoader) RegisterBundle(bundle *ComponentBundle) error {
	bl.mu.Lock()
	defer bl.mu.Unlock()
	
	if _, exists := bl.bundles[bundle.Name]; exists {
		return fmt.Errorf("bundle %s already registered", bundle.Name)
	}
	
	bl.bundles[bundle.Name] = bundle
	return nil
}

// LoadBundle loads a component bundle
func (bl *BundleLoader) LoadBundle(name string) error {
	bl.mu.RLock()
	bundle, exists := bl.bundles[name]
	bl.mu.RUnlock()
	
	if !exists {
		return fmt.Errorf("bundle %s not found", name)
	}
	
	// Load dependencies first
	for _, dep := range bundle.Dependencies {
		if err := bl.LoadBundle(dep); err != nil {
			return fmt.Errorf("failed to load dependency %s: %w", dep, err)
		}
	}
	
	// Check if already loaded
	bl.mu.RLock()
	if bl.loaded[name] {
		bl.mu.RUnlock()
		return nil
	}
	bl.mu.RUnlock()
	
	// Register all components in bundle
	for componentName, factory := range bundle.Components {
		if err := bl.loader.Register(componentName, factory); err != nil {
			return fmt.Errorf("failed to register %s: %w", componentName, err)
		}
	}
	
	// Mark as loaded
	bl.mu.Lock()
	bl.loaded[name] = true
	bl.mu.Unlock()
	
	Debug("Bundle loaded: %s", name)
	return nil
}

// DynamicLoader loads components dynamically based on conditions
type DynamicLoader struct {
	loader     *LazyLoader
	conditions map[string]LoadCondition
	mu         sync.RWMutex
}

// LoadCondition determines if a component should be loaded
type LoadCondition func() bool

// NewDynamicLoader creates a new dynamic loader
func NewDynamicLoader(loader *LazyLoader) *DynamicLoader {
	return &DynamicLoader{
		loader:     loader,
		conditions: make(map[string]LoadCondition),
	}
}

// RegisterCondition registers a load condition
func (dl *DynamicLoader) RegisterCondition(name string, condition LoadCondition) {
	dl.mu.Lock()
	defer dl.mu.Unlock()
	dl.conditions[name] = condition
}

// LoadIfNeeded loads component if condition is met
func (dl *DynamicLoader) LoadIfNeeded(name string) (Component, bool, error) {
	dl.mu.RLock()
	condition, exists := dl.conditions[name]
	dl.mu.RUnlock()
	
	if !exists || condition == nil || condition() {
		component, err := dl.loader.Load(name)
		return component, true, err
	}
	
	return nil, false, nil
}

// LoadingIndicator provides loading UI
type LoadingIndicator struct {
	Template string
	Message  string
	ShowSpinner bool
}

// GetTemplate returns loading indicator template
func (li *LoadingIndicator) GetTemplate() string {
	if li.Template != "" {
		return li.Template
	}
	
	template := `<div class="loading-indicator">`
	if li.ShowSpinner {
		template += `<div class="spinner"></div>`
	}
	if li.Message != "" {
		template += fmt.Sprintf(`<p>%s</p>`, li.Message)
	}
	template += `</div>`
	
	return template
}

// GetDriver returns nil driver
func (li *LoadingIndicator) GetDriver() LiveDriver {
	return nil
}

// Start does nothing for loading indicator
func (li *LoadingIndicator) Start() {}

// CreateLoadingPlaceholder creates a loading placeholder component
func CreateLoadingPlaceholder(message string) Component {
	return &LoadingIndicator{
		Message:     message,
		ShowSpinner: true,
	}
}

// LazyRoute represents a lazily loaded route
type LazyRoute struct {
	Path      string
	Component string
	Preload   bool
	Condition LoadCondition
}

// RouteLoader loads routes lazily
type RouteLoader struct {
	loader *LazyLoader
	routes map[string]*LazyRoute
	mu     sync.RWMutex
}

// NewRouteLoader creates a new route loader
func NewRouteLoader(loader *LazyLoader) *RouteLoader {
	return &RouteLoader{
		loader: loader,
		routes: make(map[string]*LazyRoute),
	}
}

// RegisterRoute registers a lazy route
func (rl *RouteLoader) RegisterRoute(route *LazyRoute) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	
	rl.routes[route.Path] = route
	
	// Preload if configured
	if route.Preload {
		go rl.loader.Load(route.Component)
	}
}

// LoadRoute loads a route's component
func (rl *RouteLoader) LoadRoute(path string) (Component, error) {
	rl.mu.RLock()
	route, exists := rl.routes[path]
	rl.mu.RUnlock()
	
	if !exists {
		return nil, fmt.Errorf("route %s not found", path)
	}
	
	// Check condition if present
	if route.Condition != nil && !route.Condition() {
		return nil, fmt.Errorf("route %s condition not met", path)
	}
	
	return rl.loader.Load(route.Component)
}

// ComponentPool manages a pool of reusable components
type ComponentPool struct {
	factory ComponentFactory
	pool    chan Component
	maxSize int
}

// NewComponentPool creates a new component pool
func NewComponentPool(factory ComponentFactory, maxSize int) *ComponentPool {
	return &ComponentPool{
		factory: factory,
		pool:    make(chan Component, maxSize),
		maxSize: maxSize,
	}
}

// Get gets a component from pool
func (cp *ComponentPool) Get() (Component, error) {
	select {
	case component := <-cp.pool:
		return component, nil
	default:
		// Create new component if pool is empty
		return cp.factory()
	}
}

// Put returns a component to pool
func (cp *ComponentPool) Put(component Component) {
	select {
	case cp.pool <- component:
		// Component returned to pool
	default:
		// Pool is full, discard component
	}
}

// ReflectionLoader loads components using reflection
type ReflectionLoader struct {
	loader *LazyLoader
	types  map[string]reflect.Type
	mu     sync.RWMutex
}

// NewReflectionLoader creates a new reflection loader
func NewReflectionLoader(loader *LazyLoader) *ReflectionLoader {
	return &ReflectionLoader{
		loader: loader,
		types:  make(map[string]reflect.Type),
	}
}

// RegisterType registers a component type
func (rl *ReflectionLoader) RegisterType(name string, typ reflect.Type) error {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	
	if typ.Kind() != reflect.Ptr || typ.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("type must be pointer to struct")
	}
	
	rl.types[name] = typ
	
	// Register factory with lazy loader
	factory := func() (Component, error) {
		instance := reflect.New(typ.Elem()).Interface()
		component, ok := instance.(Component)
		if !ok {
			return nil, fmt.Errorf("type does not implement Component interface")
		}
		return component, nil
	}
	
	return rl.loader.Register(name, factory)
}

// LoadType loads a component by type name
func (rl *ReflectionLoader) LoadType(name string) (Component, error) {
	return rl.loader.Load(name)
}