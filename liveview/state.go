package liveview

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"sync"
	"time"
)

// StateProvider defines the interface for state storage backends
type StateProvider interface {
	// Get retrieves state for a key
	Get(ctx context.Context, key string) (interface{}, error)
	
	// Set stores state for a key
	Set(ctx context.Context, key string, value interface{}) error
	
	// Delete removes state for a key
	Delete(ctx context.Context, key string) error
	
	// Exists checks if a key exists
	Exists(ctx context.Context, key string) (bool, error)
	
	// Clear removes all state
	Clear(ctx context.Context) error
	
	// Keys returns all keys
	Keys(ctx context.Context) ([]string, error)
}

// StateManager manages component state with pluggable backends
type StateManager struct {
	mu          sync.RWMutex
	provider    StateProvider
	cache       map[string]*CacheEntry
	subscribers map[string][]StateSubscriber
	config      *StateConfig
	ctx         context.Context
	cancel      context.CancelFunc
}

// CacheEntry represents a cached state entry
type CacheEntry struct {
	Value     interface{}
	ExpiresAt time.Time
	Version   int64
}

// StateSubscriber is called when state changes
type StateSubscriber func(key string, oldValue, newValue interface{})

// StateConfig configures the state manager
type StateConfig struct {
	Provider         StateProvider
	CacheEnabled     bool
	CacheTTL         time.Duration
	AutoPersist      bool
	PersistInterval  time.Duration
	EnableVersioning bool
}

// DefaultStateConfig returns default state configuration
func DefaultStateConfig() *StateConfig {
	return &StateConfig{
		Provider:         NewMemoryStateProvider(),
		CacheEnabled:     true,
		CacheTTL:         5 * time.Minute,
		AutoPersist:      false,
		PersistInterval:  30 * time.Second,
		EnableVersioning: true,
	}
}

// NewStateManager creates a new state manager
func NewStateManager(config *StateConfig) *StateManager {
	if config == nil {
		config = DefaultStateConfig()
	}
	
	if config.Provider == nil {
		config.Provider = NewMemoryStateProvider()
	}
	
	ctx, cancel := context.WithCancel(context.Background())
	
	sm := &StateManager{
		provider:    config.Provider,
		cache:       make(map[string]*CacheEntry),
		subscribers: make(map[string][]StateSubscriber),
		config:      config,
		ctx:         ctx,
		cancel:      cancel,
	}
	
	// Start auto-persist if enabled
	if config.AutoPersist && config.PersistInterval > 0 {
		go sm.autoPersistLoop()
	}
	
	return sm
}

// Get retrieves state for a key
func (sm *StateManager) Get(key string) (interface{}, error) {
	sm.mu.RLock()
	
	// Check cache first
	if sm.config.CacheEnabled {
		if entry, exists := sm.cache[key]; exists {
			if time.Now().Before(entry.ExpiresAt) {
				sm.mu.RUnlock()
				return entry.Value, nil
			}
		}
	}
	sm.mu.RUnlock()
	
	// Get from provider
	value, err := sm.provider.Get(sm.ctx, key)
	if err != nil {
		return nil, err
	}
	
	// Update cache
	if sm.config.CacheEnabled && value != nil {
		sm.mu.Lock()
		sm.cache[key] = &CacheEntry{
			Value:     value,
			ExpiresAt: time.Now().Add(sm.config.CacheTTL),
			Version:   time.Now().UnixNano(),
		}
		sm.mu.Unlock()
	}
	
	return value, nil
}

// Set stores state for a key
func (sm *StateManager) Set(key string, value interface{}) error {
	sm.mu.Lock()
	
	// Get old value for subscribers
	var oldValue interface{}
	if entry, exists := sm.cache[key]; exists {
		oldValue = entry.Value
	}
	
	// Update cache
	if sm.config.CacheEnabled {
		sm.cache[key] = &CacheEntry{
			Value:     value,
			ExpiresAt: time.Now().Add(sm.config.CacheTTL),
			Version:   time.Now().UnixNano(),
		}
	}
	
	// Get subscribers before unlocking
	subs := make([]StateSubscriber, len(sm.subscribers[key]))
	copy(subs, sm.subscribers[key])
	
	sm.mu.Unlock()
	
	// Set in provider
	if err := sm.provider.Set(sm.ctx, key, value); err != nil {
		return err
	}
	
	// Notify subscribers
	for _, sub := range subs {
		sub(key, oldValue, value)
	}
	
	return nil
}

// Delete removes state for a key
func (sm *StateManager) Delete(key string) error {
	sm.mu.Lock()
	
	// Get old value for subscribers
	var oldValue interface{}
	if entry, exists := sm.cache[key]; exists {
		oldValue = entry.Value
	}
	
	// Remove from cache
	delete(sm.cache, key)
	
	// Get subscribers
	subs := make([]StateSubscriber, len(sm.subscribers[key]))
	copy(subs, sm.subscribers[key])
	
	sm.mu.Unlock()
	
	// Delete from provider
	if err := sm.provider.Delete(sm.ctx, key); err != nil {
		return err
	}
	
	// Notify subscribers
	for _, sub := range subs {
		sub(key, oldValue, nil)
	}
	
	return nil
}

// Subscribe adds a subscriber for state changes
func (sm *StateManager) Subscribe(key string, subscriber StateSubscriber) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	if sm.subscribers[key] == nil {
		sm.subscribers[key] = make([]StateSubscriber, 0)
	}
	sm.subscribers[key] = append(sm.subscribers[key], subscriber)
}

// Unsubscribe removes all subscribers for a key
func (sm *StateManager) Unsubscribe(key string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	delete(sm.subscribers, key)
}

// Clear removes all state
func (sm *StateManager) Clear() error {
	sm.mu.Lock()
	sm.cache = make(map[string]*CacheEntry)
	sm.mu.Unlock()
	
	return sm.provider.Clear(sm.ctx)
}

// Keys returns all state keys
func (sm *StateManager) Keys() ([]string, error) {
	return sm.provider.Keys(sm.ctx)
}

// autoPersistLoop periodically persists cached state
func (sm *StateManager) autoPersistLoop() {
	ticker := time.NewTicker(sm.config.PersistInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-sm.ctx.Done():
			return
		case <-ticker.C:
			sm.persistCache()
		}
	}
}

// persistCache persists all cached entries
func (sm *StateManager) persistCache() {
	sm.mu.RLock()
	entries := make(map[string]interface{})
	for key, entry := range sm.cache {
		if time.Now().Before(entry.ExpiresAt) {
			entries[key] = entry.Value
		}
	}
	sm.mu.RUnlock()
	
	for key, value := range entries {
		if err := sm.provider.Set(sm.ctx, key, value); err != nil {
			Debug("Failed to persist state for key %s: %v", key, err)
		}
	}
}

// Shutdown shuts down the state manager
func (sm *StateManager) Shutdown() {
	sm.cancel()
	
	// Final persist if auto-persist is enabled
	if sm.config.AutoPersist {
		sm.persistCache()
	}
}

// MemoryStateProvider is an in-memory state provider
type MemoryStateProvider struct {
	mu   sync.RWMutex
	data map[string]interface{}
}

// NewMemoryStateProvider creates a new memory state provider
func NewMemoryStateProvider() *MemoryStateProvider {
	return &MemoryStateProvider{
		data: make(map[string]interface{}),
	}
}

// Get retrieves state from memory
func (msp *MemoryStateProvider) Get(ctx context.Context, key string) (interface{}, error) {
	msp.mu.RLock()
	defer msp.mu.RUnlock()
	
	value, exists := msp.data[key]
	if !exists {
		return nil, nil
	}
	return value, nil
}

// Set stores state in memory
func (msp *MemoryStateProvider) Set(ctx context.Context, key string, value interface{}) error {
	msp.mu.Lock()
	defer msp.mu.Unlock()
	
	msp.data[key] = value
	return nil
}

// Delete removes state from memory
func (msp *MemoryStateProvider) Delete(ctx context.Context, key string) error {
	msp.mu.Lock()
	defer msp.mu.Unlock()
	
	delete(msp.data, key)
	return nil
}

// Exists checks if a key exists in memory
func (msp *MemoryStateProvider) Exists(ctx context.Context, key string) (bool, error) {
	msp.mu.RLock()
	defer msp.mu.RUnlock()
	
	_, exists := msp.data[key]
	return exists, nil
}

// Clear removes all state from memory
func (msp *MemoryStateProvider) Clear(ctx context.Context) error {
	msp.mu.Lock()
	defer msp.mu.Unlock()
	
	msp.data = make(map[string]interface{})
	return nil
}

// Keys returns all keys from memory
func (msp *MemoryStateProvider) Keys(ctx context.Context) ([]string, error) {
	msp.mu.RLock()
	defer msp.mu.RUnlock()
	
	keys := make([]string, 0, len(msp.data))
	for key := range msp.data {
		keys = append(keys, key)
	}
	return keys, nil
}

// ReactiveState provides reactive state management
type ReactiveState struct {
	manager     *StateManager
	componentID string
	bindings    map[string]*StateBinding
	mu          sync.RWMutex
}

// StateBinding represents a two-way binding between state and component
type StateBinding struct {
	Key      string
	Value    interface{}
	OnChange func(oldValue, newValue interface{})
}

// NewReactiveState creates a new reactive state for a component
func NewReactiveState(manager *StateManager, componentID string) *ReactiveState {
	return &ReactiveState{
		manager:     manager,
		componentID: componentID,
		bindings:    make(map[string]*StateBinding),
	}
}

// Bind creates a two-way binding for a state key
func (rs *ReactiveState) Bind(key string, initialValue interface{}, onChange func(oldValue, newValue interface{})) error {
	fullKey := fmt.Sprintf("%s.%s", rs.componentID, key)
	
	// Set initial value
	if err := rs.manager.Set(fullKey, initialValue); err != nil {
		return err
	}
	
	// Create binding
	binding := &StateBinding{
		Key:      fullKey,
		Value:    initialValue,
		OnChange: onChange,
	}
	
	rs.mu.Lock()
	rs.bindings[key] = binding
	rs.mu.Unlock()
	
	// Subscribe to changes
	rs.manager.Subscribe(fullKey, func(k string, oldVal, newVal interface{}) {
		rs.mu.Lock()
		if b, exists := rs.bindings[key]; exists {
			b.Value = newVal
			if b.OnChange != nil {
				b.OnChange(oldVal, newVal)
			}
		}
		rs.mu.Unlock()
	})
	
	return nil
}

// Get retrieves a bound value
func (rs *ReactiveState) Get(key string) interface{} {
	rs.mu.RLock()
	defer rs.mu.RUnlock()
	
	if binding, exists := rs.bindings[key]; exists {
		return binding.Value
	}
	return nil
}

// Set updates a bound value
func (rs *ReactiveState) Set(key string, value interface{}) error {
	rs.mu.Lock()
	binding, exists := rs.bindings[key]
	if !exists {
		rs.mu.Unlock()
		return fmt.Errorf("no binding for key: %s", key)
	}
	
	oldValue := binding.Value
	binding.Value = value
	fullKey := binding.Key
	rs.mu.Unlock()
	
	// Update in state manager
	if err := rs.manager.Set(fullKey, value); err != nil {
		return err
	}
	
	// Trigger onChange
	if binding.OnChange != nil {
		binding.OnChange(oldValue, value)
	}
	
	return nil
}

// Unbind removes a binding
func (rs *ReactiveState) Unbind(key string) {
	rs.mu.Lock()
	if binding, exists := rs.bindings[key]; exists {
		rs.manager.Unsubscribe(binding.Key)
		delete(rs.bindings, key)
	}
	rs.mu.Unlock()
}

// UnbindAll removes all bindings
func (rs *ReactiveState) UnbindAll() {
	rs.mu.Lock()
	for key, binding := range rs.bindings {
		rs.manager.Unsubscribe(binding.Key)
		delete(rs.bindings, key)
	}
	rs.mu.Unlock()
}

// StateSnapshot represents a point-in-time snapshot of state
type StateSnapshot struct {
	Timestamp time.Time
	Data      map[string]interface{}
	Version   int64
}

// TakeSnapshot creates a snapshot of current state
func (sm *StateManager) TakeSnapshot() (*StateSnapshot, error) {
	keys, err := sm.Keys()
	if err != nil {
		return nil, err
	}
	
	snapshot := &StateSnapshot{
		Timestamp: time.Now(),
		Data:      make(map[string]interface{}),
		Version:   time.Now().UnixNano(),
	}
	
	for _, key := range keys {
		value, err := sm.Get(key)
		if err != nil {
			continue
		}
		snapshot.Data[key] = value
	}
	
	return snapshot, nil
}

// RestoreSnapshot restores state from a snapshot
func (sm *StateManager) RestoreSnapshot(snapshot *StateSnapshot) error {
	if snapshot == nil {
		return fmt.Errorf("snapshot is nil")
	}
	
	// Clear current state
	if err := sm.Clear(); err != nil {
		return err
	}
	
	// Restore from snapshot
	for key, value := range snapshot.Data {
		if err := sm.Set(key, value); err != nil {
			return fmt.Errorf("failed to restore key %s: %w", key, err)
		}
	}
	
	return nil
}

// JSONStateProvider stores state as JSON (useful for persistence)
type JSONStateProvider struct {
	backend StateProvider
}

// NewJSONStateProvider creates a JSON state provider wrapping another provider
func NewJSONStateProvider(backend StateProvider) *JSONStateProvider {
	return &JSONStateProvider{backend: backend}
}

// Get retrieves and deserializes JSON state
func (jsp *JSONStateProvider) Get(ctx context.Context, key string) (interface{}, error) {
	data, err := jsp.backend.Get(ctx, key)
	if err != nil || data == nil {
		return data, err
	}
	
	// If data is already deserialized, return it
	if _, ok := data.([]byte); !ok {
		return data, nil
	}
	
	var result interface{}
	if err := json.Unmarshal(data.([]byte), &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}
	
	return result, nil
}

// Set serializes and stores state as JSON
func (jsp *JSONStateProvider) Set(ctx context.Context, key string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	
	return jsp.backend.Set(ctx, key, data)
}

// Delete delegates to backend
func (jsp *JSONStateProvider) Delete(ctx context.Context, key string) error {
	return jsp.backend.Delete(ctx, key)
}

// Exists delegates to backend
func (jsp *JSONStateProvider) Exists(ctx context.Context, key string) (bool, error) {
	return jsp.backend.Exists(ctx, key)
}

// Clear delegates to backend
func (jsp *JSONStateProvider) Clear(ctx context.Context) error {
	return jsp.backend.Clear(ctx)
}

// Keys delegates to backend
func (jsp *JSONStateProvider) Keys(ctx context.Context) ([]string, error) {
	return jsp.backend.Keys(ctx)
}

// ComputedState represents a computed value derived from other state
type ComputedState struct {
	manager      *StateManager
	dependencies []string
	compute      func(deps map[string]interface{}) interface{}
	cache        interface{}
	mu           sync.RWMutex
}

// NewComputedState creates a new computed state
func NewComputedState(manager *StateManager, dependencies []string, compute func(deps map[string]interface{}) interface{}) *ComputedState {
	cs := &ComputedState{
		manager:      manager,
		dependencies: dependencies,
		compute:      compute,
	}
	
	// Subscribe to dependency changes
	for _, dep := range dependencies {
		manager.Subscribe(dep, func(key string, oldValue, newValue interface{}) {
			cs.invalidate()
		})
	}
	
	return cs
}

// Get returns the computed value
func (cs *ComputedState) Get() interface{} {
	cs.mu.RLock()
	if cs.cache != nil {
		defer cs.mu.RUnlock()
		return cs.cache
	}
	cs.mu.RUnlock()
	
	// Compute value
	cs.mu.Lock()
	defer cs.mu.Unlock()
	
	// Double-check after acquiring write lock
	if cs.cache != nil {
		return cs.cache
	}
	
	// Get dependency values
	deps := make(map[string]interface{})
	for _, dep := range cs.dependencies {
		value, _ := cs.manager.Get(dep)
		deps[dep] = value
	}
	
	// Compute and cache
	cs.cache = cs.compute(deps)
	return cs.cache
}

// invalidate clears the cached value
func (cs *ComputedState) invalidate() {
	cs.mu.Lock()
	cs.cache = nil
	cs.mu.Unlock()
}

// StateTransaction provides transactional state updates
type StateTransaction struct {
	manager  *StateManager
	changes  map[string]interface{}
	original map[string]interface{}
	mu       sync.Mutex
}

// BeginTransaction starts a new state transaction
func (sm *StateManager) BeginTransaction() *StateTransaction {
	return &StateTransaction{
		manager:  sm,
		changes:  make(map[string]interface{}),
		original: make(map[string]interface{}),
	}
}

// Set stages a state change in the transaction
func (st *StateTransaction) Set(key string, value interface{}) {
	st.mu.Lock()
	defer st.mu.Unlock()
	
	// Store original value if not already stored
	if _, exists := st.original[key]; !exists {
		original, _ := st.manager.Get(key)
		st.original[key] = original
	}
	
	st.changes[key] = value
}

// Commit applies all staged changes
func (st *StateTransaction) Commit() error {
	st.mu.Lock()
	defer st.mu.Unlock()
	
	// Apply all changes
	for key, value := range st.changes {
		if err := st.manager.Set(key, value); err != nil {
			// Rollback on error
			st.rollback()
			return fmt.Errorf("transaction failed on key %s: %w", key, err)
		}
	}
	
	return nil
}

// Rollback reverts all changes
func (st *StateTransaction) Rollback() {
	st.mu.Lock()
	defer st.mu.Unlock()
	st.rollback()
}

// rollback internal rollback without lock
func (st *StateTransaction) rollback() {
	for key, original := range st.original {
		if original == nil {
			st.manager.Delete(key)
		} else {
			st.manager.Set(key, original)
		}
	}
}

// StateHelper provides helper functions for state management
type StateHelper struct {
	manager *StateManager
}

// NewStateHelper creates a new state helper
func NewStateHelper(manager *StateManager) *StateHelper {
	return &StateHelper{manager: manager}
}

// GetString gets a string value
func (sh *StateHelper) GetString(key string, defaultValue string) string {
	value, err := sh.manager.Get(key)
	if err != nil || value == nil {
		return defaultValue
	}
	
	if str, ok := value.(string); ok {
		return str
	}
	
	return fmt.Sprintf("%v", value)
}

// GetInt gets an int value
func (sh *StateHelper) GetInt(key string, defaultValue int) int {
	value, err := sh.manager.Get(key)
	if err != nil || value == nil {
		return defaultValue
	}
	
	switch v := value.(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	default:
		return defaultValue
	}
}

// GetBool gets a bool value
func (sh *StateHelper) GetBool(key string, defaultValue bool) bool {
	value, err := sh.manager.Get(key)
	if err != nil || value == nil {
		return defaultValue
	}
	
	if b, ok := value.(bool); ok {
		return b
	}
	
	return defaultValue
}

// GetMap gets a map value
func (sh *StateHelper) GetMap(key string) map[string]interface{} {
	value, err := sh.manager.Get(key)
	if err != nil || value == nil {
		return make(map[string]interface{})
	}
	
	if m, ok := value.(map[string]interface{}); ok {
		return m
	}
	
	// Try to convert using reflection
	v := reflect.ValueOf(value)
	if v.Kind() == reflect.Map {
		result := make(map[string]interface{})
		for _, k := range v.MapKeys() {
			result[fmt.Sprintf("%v", k.Interface())] = v.MapIndex(k).Interface()
		}
		return result
	}
	
	return make(map[string]interface{})
}

// Increment increments an integer value
func (sh *StateHelper) Increment(key string, delta int) (int, error) {
	current := sh.GetInt(key, 0)
	newValue := current + delta
	
	if err := sh.manager.Set(key, newValue); err != nil {
		return current, err
	}
	
	return newValue, nil
}

// Toggle toggles a boolean value
func (sh *StateHelper) Toggle(key string) (bool, error) {
	current := sh.GetBool(key, false)
	newValue := !current
	
	if err := sh.manager.Set(key, newValue); err != nil {
		return current, err
	}
	
	return newValue, nil
}