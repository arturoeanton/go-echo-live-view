package liveview

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestNewStateManager(t *testing.T) {
	sm := NewStateManager(nil)
	
	if sm == nil {
		t.Fatal("NewStateManager returned nil")
	}
	
	if sm.provider == nil {
		t.Error("Expected default provider to be set")
	}
	
	if sm.config == nil {
		t.Error("Expected default config to be set")
	}
	
	if len(sm.cache) != 0 {
		t.Error("Expected empty cache")
	}
	
	if len(sm.subscribers) != 0 {
		t.Error("Expected empty subscribers")
	}
}

func TestStateManagerGetSet(t *testing.T) {
	sm := NewStateManager(nil)
	
	// Set a value
	err := sm.Set("key1", "value1")
	if err != nil {
		t.Errorf("Set() error = %v", err)
	}
	
	// Get the value
	value, err := sm.Get("key1")
	if err != nil {
		t.Errorf("Get() error = %v", err)
	}
	
	if value != "value1" {
		t.Errorf("Expected 'value1', got %v", value)
	}
	
	// Get non-existent key
	value, err = sm.Get("nonexistent")
	if err != nil {
		t.Errorf("Get() should not error for non-existent key: %v", err)
	}
	
	if value != nil {
		t.Errorf("Expected nil for non-existent key, got %v", value)
	}
}

func TestStateManagerDelete(t *testing.T) {
	sm := NewStateManager(nil)
	
	// Set and delete a value
	sm.Set("key1", "value1")
	
	err := sm.Delete("key1")
	if err != nil {
		t.Errorf("Delete() error = %v", err)
	}
	
	// Try to get deleted value
	value, err := sm.Get("key1")
	if err != nil {
		t.Errorf("Get() error = %v", err)
	}
	
	if value != nil {
		t.Errorf("Expected nil after delete, got %v", value)
	}
}

func TestStateManagerSubscribe(t *testing.T) {
	sm := NewStateManager(nil)
	
	var oldVal, newVal interface{}
	var notified bool
	
	// Subscribe to changes
	sm.Subscribe("key1", func(key string, old, new interface{}) {
		notified = true
		oldVal = old
		newVal = new
	})
	
	// Set a value (should trigger subscriber)
	sm.Set("key1", "value1")
	
	if !notified {
		t.Error("Subscriber was not notified")
	}
	
	if oldVal != nil {
		t.Errorf("Expected nil old value, got %v", oldVal)
	}
	
	if newVal != "value1" {
		t.Errorf("Expected 'value1' new value, got %v", newVal)
	}
	
	// Update value
	notified = false
	sm.Set("key1", "value2")
	
	if !notified {
		t.Error("Subscriber was not notified on update")
	}
	
	if oldVal != "value1" {
		t.Errorf("Expected 'value1' old value, got %v", oldVal)
	}
	
	if newVal != "value2" {
		t.Errorf("Expected 'value2' new value, got %v", newVal)
	}
	
	// Delete value
	notified = false
	sm.Delete("key1")
	
	if !notified {
		t.Error("Subscriber was not notified on delete")
	}
	
	if newVal != nil {
		t.Errorf("Expected nil new value on delete, got %v", newVal)
	}
}

func TestStateManagerUnsubscribe(t *testing.T) {
	sm := NewStateManager(nil)
	
	var notified bool
	
	sm.Subscribe("key1", func(key string, old, new interface{}) {
		notified = true
	})
	
	// Unsubscribe
	sm.Unsubscribe("key1")
	
	// Set value (should not trigger subscriber)
	sm.Set("key1", "value1")
	
	if notified {
		t.Error("Subscriber was notified after unsubscribe")
	}
}

func TestStateManagerClear(t *testing.T) {
	sm := NewStateManager(nil)
	
	// Set multiple values
	sm.Set("key1", "value1")
	sm.Set("key2", "value2")
	sm.Set("key3", "value3")
	
	// Clear all
	err := sm.Clear()
	if err != nil {
		t.Errorf("Clear() error = %v", err)
	}
	
	// Check all values are gone
	keys, err := sm.Keys()
	if err != nil {
		t.Errorf("Keys() error = %v", err)
	}
	
	if len(keys) != 0 {
		t.Errorf("Expected 0 keys after clear, got %d", len(keys))
	}
}

func TestStateManagerKeys(t *testing.T) {
	sm := NewStateManager(nil)
	
	// Set multiple values
	sm.Set("key1", "value1")
	sm.Set("key2", "value2")
	sm.Set("key3", "value3")
	
	keys, err := sm.Keys()
	if err != nil {
		t.Errorf("Keys() error = %v", err)
	}
	
	if len(keys) != 3 {
		t.Errorf("Expected 3 keys, got %d", len(keys))
	}
	
	// Check all keys are present
	keyMap := make(map[string]bool)
	for _, key := range keys {
		keyMap[key] = true
	}
	
	for _, expected := range []string{"key1", "key2", "key3"} {
		if !keyMap[expected] {
			t.Errorf("Missing key: %s", expected)
		}
	}
}

func TestStateManagerCache(t *testing.T) {
	config := &StateConfig{
		Provider:     NewMemoryStateProvider(),
		CacheEnabled: true,
		CacheTTL:     100 * time.Millisecond,
	}
	sm := NewStateManager(config)
	
	// Set a value
	sm.Set("key1", "value1")
	
	// First get should cache
	value1, _ := sm.Get("key1")
	
	// Second get should come from cache
	value2, _ := sm.Get("key1")
	
	if value1 != value2 {
		t.Error("Cached values don't match")
	}
	
	// Wait for cache to expire
	time.Sleep(150 * time.Millisecond)
	
	// This should fetch from provider again
	value3, _ := sm.Get("key1")
	
	if value3 != "value1" {
		t.Errorf("Expected 'value1' after cache expiry, got %v", value3)
	}
}

func TestStateManagerAutoPersist(t *testing.T) {
	config := &StateConfig{
		Provider:        NewMemoryStateProvider(),
		CacheEnabled:    true,
		CacheTTL:        5 * time.Second,
		AutoPersist:     true,
		PersistInterval: 50 * time.Millisecond,
	}
	sm := NewStateManager(config)
	defer sm.Shutdown()
	
	// Set a value in cache
	sm.Set("key1", "value1")
	
	// Wait for auto-persist
	time.Sleep(100 * time.Millisecond)
	
	// Value should be persisted
	provider := sm.provider
	value, err := provider.Get(context.Background(), "key1")
	if err != nil {
		t.Errorf("Provider Get() error = %v", err)
	}
	
	if value != "value1" {
		t.Errorf("Expected persisted value 'value1', got %v", value)
	}
}

func TestMemoryStateProvider(t *testing.T) {
	msp := NewMemoryStateProvider()
	ctx := context.Background()
	
	// Test Set and Get
	err := msp.Set(ctx, "key1", "value1")
	if err != nil {
		t.Errorf("Set() error = %v", err)
	}
	
	value, err := msp.Get(ctx, "key1")
	if err != nil {
		t.Errorf("Get() error = %v", err)
	}
	
	if value != "value1" {
		t.Errorf("Expected 'value1', got %v", value)
	}
	
	// Test Exists
	exists, err := msp.Exists(ctx, "key1")
	if err != nil {
		t.Errorf("Exists() error = %v", err)
	}
	
	if !exists {
		t.Error("Key should exist")
	}
	
	// Test Delete
	err = msp.Delete(ctx, "key1")
	if err != nil {
		t.Errorf("Delete() error = %v", err)
	}
	
	exists, err = msp.Exists(ctx, "key1")
	if err != nil {
		t.Errorf("Exists() error = %v", err)
	}
	
	if exists {
		t.Error("Key should not exist after delete")
	}
	
	// Test Keys
	msp.Set(ctx, "a", 1)
	msp.Set(ctx, "b", 2)
	msp.Set(ctx, "c", 3)
	
	keys, err := msp.Keys(ctx)
	if err != nil {
		t.Errorf("Keys() error = %v", err)
	}
	
	if len(keys) != 3 {
		t.Errorf("Expected 3 keys, got %d", len(keys))
	}
	
	// Test Clear
	err = msp.Clear(ctx)
	if err != nil {
		t.Errorf("Clear() error = %v", err)
	}
	
	keys, err = msp.Keys(ctx)
	if err != nil {
		t.Errorf("Keys() error = %v", err)
	}
	
	if len(keys) != 0 {
		t.Errorf("Expected 0 keys after clear, got %d", len(keys))
	}
}

func TestReactiveState(t *testing.T) {
	sm := NewStateManager(nil)
	rs := NewReactiveState(sm, "component1")
	
	var changeNotified bool
	var oldValue, newValue interface{}
	
	// Bind a value with change handler
	err := rs.Bind("prop1", "initial", func(old, new interface{}) {
		changeNotified = true
		oldValue = old
		newValue = new
	})
	
	if err != nil {
		t.Errorf("Bind() error = %v", err)
	}
	
	// Get bound value
	value := rs.Get("prop1")
	if value != "initial" {
		t.Errorf("Expected 'initial', got %v", value)
	}
	
	// Set bound value
	err = rs.Set("prop1", "updated")
	if err != nil {
		t.Errorf("Set() error = %v", err)
	}
	
	if !changeNotified {
		t.Error("Change handler was not notified")
	}
	
	if oldValue != "initial" {
		t.Errorf("Expected old value 'initial', got %v", oldValue)
	}
	
	if newValue != "updated" {
		t.Errorf("Expected new value 'updated', got %v", newValue)
	}
	
	// Get updated value
	value = rs.Get("prop1")
	if value != "updated" {
		t.Errorf("Expected 'updated', got %v", value)
	}
}

func TestReactiveStateUnbind(t *testing.T) {
	sm := NewStateManager(nil)
	rs := NewReactiveState(sm, "component1")
	
	// Bind and unbind
	rs.Bind("prop1", "value1", nil)
	rs.Unbind("prop1")
	
	// Should return nil for unbound property
	value := rs.Get("prop1")
	if value != nil {
		t.Errorf("Expected nil for unbound property, got %v", value)
	}
	
	// Set should error for unbound property
	err := rs.Set("prop1", "newvalue")
	if err == nil {
		t.Error("Expected error when setting unbound property")
	}
}

func TestReactiveStateUnbindAll(t *testing.T) {
	sm := NewStateManager(nil)
	rs := NewReactiveState(sm, "component1")
	
	// Bind multiple properties
	rs.Bind("prop1", "value1", nil)
	rs.Bind("prop2", "value2", nil)
	rs.Bind("prop3", "value3", nil)
	
	// Unbind all
	rs.UnbindAll()
	
	// All should return nil
	if rs.Get("prop1") != nil {
		t.Error("prop1 should be unbound")
	}
	if rs.Get("prop2") != nil {
		t.Error("prop2 should be unbound")
	}
	if rs.Get("prop3") != nil {
		t.Error("prop3 should be unbound")
	}
}

func TestStateSnapshot(t *testing.T) {
	sm := NewStateManager(nil)
	
	// Set up initial state
	sm.Set("key1", "value1")
	sm.Set("key2", 42)
	sm.Set("key3", true)
	
	// Take snapshot
	snapshot, err := sm.TakeSnapshot()
	if err != nil {
		t.Errorf("TakeSnapshot() error = %v", err)
	}
	
	if snapshot == nil {
		t.Fatal("Snapshot is nil")
	}
	
	if len(snapshot.Data) != 3 {
		t.Errorf("Expected 3 items in snapshot, got %d", len(snapshot.Data))
	}
	
	// Verify snapshot data
	if snapshot.Data["key1"] != "value1" {
		t.Errorf("Snapshot key1: expected 'value1', got %v", snapshot.Data["key1"])
	}
	
	if snapshot.Data["key2"] != 42 {
		t.Errorf("Snapshot key2: expected 42, got %v", snapshot.Data["key2"])
	}
	
	if snapshot.Data["key3"] != true {
		t.Errorf("Snapshot key3: expected true, got %v", snapshot.Data["key3"])
	}
	
	// Modify state
	sm.Set("key1", "modified")
	sm.Set("key4", "new")
	
	// Restore snapshot
	err = sm.RestoreSnapshot(snapshot)
	if err != nil {
		t.Errorf("RestoreSnapshot() error = %v", err)
	}
	
	// Verify restored state
	value1, _ := sm.Get("key1")
	if value1 != "value1" {
		t.Errorf("Restored key1: expected 'value1', got %v", value1)
	}
	
	value4, _ := sm.Get("key4")
	if value4 != nil {
		t.Errorf("key4 should not exist after restore, got %v", value4)
	}
}

func TestJSONStateProvider(t *testing.T) {
	backend := NewMemoryStateProvider()
	jsp := NewJSONStateProvider(backend)
	ctx := context.Background()
	
	// Test with complex data structure
	data := map[string]interface{}{
		"name":  "test",
		"count": 42,
		"items": []string{"a", "b", "c"},
	}
	
	// Set JSON data
	err := jsp.Set(ctx, "complex", data)
	if err != nil {
		t.Errorf("Set() error = %v", err)
	}
	
	// Get JSON data
	retrieved, err := jsp.Get(ctx, "complex")
	if err != nil {
		t.Errorf("Get() error = %v", err)
	}
	
	// Verify structure
	if retrievedMap, ok := retrieved.(map[string]interface{}); ok {
		if retrievedMap["name"] != "test" {
			t.Errorf("Expected name 'test', got %v", retrievedMap["name"])
		}
		
		// JSON numbers are float64
		if count, ok := retrievedMap["count"].(float64); !ok || count != 42 {
			t.Errorf("Expected count 42, got %v", retrievedMap["count"])
		}
	} else {
		t.Errorf("Retrieved data is not a map: %T", retrieved)
	}
}

func TestComputedState(t *testing.T) {
	sm := NewStateManager(nil)
	
	// Set up dependencies
	sm.Set("price", 100.0)
	sm.Set("quantity", 5)
	sm.Set("discount", 0.1)
	
	// Create computed state for total
	cs := NewComputedState(sm, []string{"price", "quantity", "discount"}, func(deps map[string]interface{}) interface{} {
		price, _ := deps["price"].(float64)
		quantity, _ := deps["quantity"].(int)
		discount, _ := deps["discount"].(float64)
		
		subtotal := price * float64(quantity)
		total := subtotal * (1 - discount)
		return total
	})
	
	// Get computed value
	total := cs.Get()
	if total != 450.0 {
		t.Errorf("Expected total 450.0, got %v", total)
	}
	
	// Update dependency
	sm.Set("quantity", 10)
	
	// Computed value should update
	total = cs.Get()
	if total != 900.0 {
		t.Errorf("Expected total 900.0 after update, got %v", total)
	}
}

func TestStateTransaction(t *testing.T) {
	sm := NewStateManager(nil)
	
	// Set initial state
	sm.Set("balance", 1000)
	sm.Set("pending", 0)
	
	// Begin transaction
	tx := sm.BeginTransaction()
	
	// Stage changes
	tx.Set("balance", 800)
	tx.Set("pending", 200)
	
	// Values should not be changed yet
	balance, _ := sm.Get("balance")
	if balance != 1000 {
		t.Errorf("Balance changed before commit: %v", balance)
	}
	
	// Commit transaction
	err := tx.Commit()
	if err != nil {
		t.Errorf("Commit() error = %v", err)
	}
	
	// Values should be updated
	balance, _ = sm.Get("balance")
	if balance != 800 {
		t.Errorf("Expected balance 800 after commit, got %v", balance)
	}
	
	pending, _ := sm.Get("pending")
	if pending != 200 {
		t.Errorf("Expected pending 200 after commit, got %v", pending)
	}
}

func TestStateTransactionRollback(t *testing.T) {
	sm := NewStateManager(nil)
	
	// Set initial state
	sm.Set("value1", "initial1")
	sm.Set("value2", "initial2")
	
	// Begin transaction
	tx := sm.BeginTransaction()
	
	// Stage changes
	tx.Set("value1", "changed1")
	tx.Set("value2", "changed2")
	
	// Rollback transaction
	tx.Rollback()
	
	// Values should be unchanged
	value1, _ := sm.Get("value1")
	if value1 != "initial1" {
		t.Errorf("value1 changed after rollback: %v", value1)
	}
	
	value2, _ := sm.Get("value2")
	if value2 != "initial2" {
		t.Errorf("value2 changed after rollback: %v", value2)
	}
}

func TestStateHelper(t *testing.T) {
	sm := NewStateManager(nil)
	sh := NewStateHelper(sm)
	
	// Test GetString
	sm.Set("str", "hello")
	if sh.GetString("str", "default") != "hello" {
		t.Error("GetString failed")
	}
	
	if sh.GetString("nonexistent", "default") != "default" {
		t.Error("GetString default failed")
	}
	
	// Test GetInt
	sm.Set("int", 42)
	if sh.GetInt("int", 0) != 42 {
		t.Error("GetInt failed")
	}
	
	sm.Set("float", 3.14)
	if sh.GetInt("float", 0) != 3 {
		t.Error("GetInt float conversion failed")
	}
	
	// Test GetBool
	sm.Set("bool", true)
	if !sh.GetBool("bool", false) {
		t.Error("GetBool failed")
	}
	
	if sh.GetBool("nonexistent", true) != true {
		t.Error("GetBool default failed")
	}
	
	// Test GetMap
	mapData := map[string]interface{}{"key": "value"}
	sm.Set("map", mapData)
	
	retrieved := sh.GetMap("map")
	if retrieved["key"] != "value" {
		t.Error("GetMap failed")
	}
	
	// Test Increment
	sm.Set("counter", 10)
	newVal, err := sh.Increment("counter", 5)
	if err != nil {
		t.Errorf("Increment() error = %v", err)
	}
	
	if newVal != 15 {
		t.Errorf("Expected 15 after increment, got %d", newVal)
	}
	
	// Test Toggle
	sm.Set("flag", false)
	toggled, err := sh.Toggle("flag")
	if err != nil {
		t.Errorf("Toggle() error = %v", err)
	}
	
	if !toggled {
		t.Error("Expected true after toggle")
	}
}

func TestConcurrentStateOperations(t *testing.T) {
	sm := NewStateManager(nil)
	
	var wg sync.WaitGroup
	numGoroutines := 100
	
	// Concurrent sets
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			key := fmt.Sprintf("key%d", id)
			value := fmt.Sprintf("value%d", id)
			sm.Set(key, value)
		}(i)
	}
	
	wg.Wait()
	
	// Verify all values were set
	for i := 0; i < numGoroutines; i++ {
		key := fmt.Sprintf("key%d", i)
		expectedValue := fmt.Sprintf("value%d", i)
		
		value, err := sm.Get(key)
		if err != nil {
			t.Errorf("Get(%s) error = %v", key, err)
		}
		
		if value != expectedValue {
			t.Errorf("Expected %s for %s, got %v", expectedValue, key, value)
		}
	}
	
	// Concurrent updates with subscribers
	var updateCount int
	var mu sync.Mutex
	
	sm.Subscribe("shared", func(key string, old, new interface{}) {
		mu.Lock()
		updateCount++
		mu.Unlock()
	})
	
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			sm.Set("shared", id)
		}(i)
	}
	
	wg.Wait()
	
	if updateCount == 0 {
		t.Error("No updates recorded")
	}
}

func TestStateManagerShutdown(t *testing.T) {
	config := &StateConfig{
		Provider:        NewMemoryStateProvider(),
		AutoPersist:     true,
		PersistInterval: 50 * time.Millisecond,
	}
	sm := NewStateManager(config)
	
	// Set some values
	sm.Set("key1", "value1")
	sm.Set("key2", "value2")
	
	// Shutdown
	sm.Shutdown()
	
	// Auto-persist loop should stop
	// Give it time to stop
	time.Sleep(100 * time.Millisecond)
	
	// Context should be cancelled
	select {
	case <-sm.ctx.Done():
		// Expected
	default:
		t.Error("Context not cancelled after shutdown")
	}
}

func TestStateManagerVersioning(t *testing.T) {
	config := &StateConfig{
		Provider:         NewMemoryStateProvider(),
		CacheEnabled:     true,
		EnableVersioning: true,
	}
	sm := NewStateManager(config)
	
	// Set initial value
	sm.Set("key1", "version1")
	
	// Get cache entry
	sm.mu.RLock()
	entry1 := sm.cache["key1"]
	version1 := entry1.Version
	sm.mu.RUnlock()
	
	// Update value
	time.Sleep(1 * time.Millisecond) // Ensure different timestamp
	sm.Set("key1", "version2")
	
	// Get new cache entry
	sm.mu.RLock()
	entry2 := sm.cache["key1"]
	version2 := entry2.Version
	sm.mu.RUnlock()
	
	if version2 <= version1 {
		t.Error("Version should increase on update")
	}
}

func TestStateManagerNilProvider(t *testing.T) {
	config := &StateConfig{
		Provider: nil, // Will use default
	}
	sm := NewStateManager(config)
	
	if sm.provider == nil {
		t.Error("Provider should not be nil")
	}
	
	// Should work with default provider
	err := sm.Set("key", "value")
	if err != nil {
		t.Errorf("Set() with default provider error = %v", err)
	}
	
	value, err := sm.Get("key")
	if err != nil {
		t.Errorf("Get() with default provider error = %v", err)
	}
	
	if value != "value" {
		t.Errorf("Expected 'value', got %v", value)
	}
}