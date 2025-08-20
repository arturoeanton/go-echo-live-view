package liveview_test

import (
	"testing"

	"github.com/arturoeanton/go-echo-live-view/liveview"
)

func TestBiMapCreation(t *testing.T) {
	bimap := liveview.NewBiMap[string, int]()

	if bimap == nil {
		t.Fatal("NewBiMap should not return nil")
	}

	// Test empty bimap
	_, exists := bimap.Get("nonexistent")
	if exists {
		t.Error("Empty BiMap should return false for Get")
	}
}

func TestBiMapSet(t *testing.T) {
	bimap := liveview.NewBiMap[string, int]()

	// Test basic set
	bimap.Set("one", 1)
	bimap.Set("two", 2)
	bimap.Set("three", 3)

	// Test duplicate key (should not update due to Set logic)
	bimap.Set("one", 10)
	val, _ := bimap.Get("one")
	if val != 1 {
		t.Errorf("Set should not update existing key, expected 1, got %d", val)
	}

	// Test duplicate value with different key (should not set)
	bimap.Set("ten", 2) // 2 already exists with key "two"
	_, exists := bimap.Get("ten")
	if exists {
		t.Error("Set should not add key with existing value")
	}
}

func TestBiMapGet(t *testing.T) {
	bimap := liveview.NewBiMap[string, int]()

	bimap.Set("alpha", 1)
	bimap.Set("beta", 2)
	bimap.Set("gamma", 3)

	// Test existing keys
	val, exists := bimap.Get("alpha")
	if !exists || val != 1 {
		t.Errorf("Get('alpha') = %d, %v; want 1, true", val, exists)
	}

	val, exists = bimap.Get("beta")
	if !exists || val != 2 {
		t.Errorf("Get('beta') = %d, %v; want 2, true", val, exists)
	}

	// Test non-existing key
	val, exists = bimap.Get("delta")
	if exists {
		t.Errorf("Get('delta') should return false for non-existing key")
	}
}

func TestBiMapGetByValue(t *testing.T) {
	bimap := liveview.NewBiMap[string, int]()

	bimap.Set("one", 1)
	bimap.Set("two", 2)
	bimap.Set("three", 3)

	// Test existing values
	key, exists := bimap.GetByValue(1)
	if !exists || key != "one" {
		t.Errorf("GetByValue(1) = %s, %v; want 'one', true", key, exists)
	}

	key, exists = bimap.GetByValue(2)
	if !exists || key != "two" {
		t.Errorf("GetByValue(2) = %s, %v; want 'two', true", key, exists)
	}

	// Test non-existing value
	key, exists = bimap.GetByValue(99)
	if exists {
		t.Errorf("GetByValue(99) should return false for non-existing value")
	}
}

func TestBiMapDelete(t *testing.T) {
	bimap := liveview.NewBiMap[string, int]()

	bimap.Set("a", 1)
	bimap.Set("b", 2)
	bimap.Set("c", 3)

	// Delete existing key
	bimap.Delete("b")
	
	// Verify deletion
	_, exists := bimap.Get("b")
	if exists {
		t.Error("Key 'b' should not exist after deletion")
	}

	_, exists = bimap.GetByValue(2)
	if exists {
		t.Error("Value 2 should not exist after deletion")
	}

	// Delete non-existing key (should not panic)
	bimap.Delete("z")
}

func TestBiMapDeleteByValue(t *testing.T) {
	bimap := liveview.NewBiMap[string, int]()

	bimap.Set("x", 10)
	bimap.Set("y", 20)
	bimap.Set("z", 30)

	// Delete existing value
	bimap.DeleteByValue(20)
	
	// Verify deletion
	_, exists := bimap.Get("y")
	if exists {
		t.Error("Key 'y' should not exist after deletion")
	}

	_, exists = bimap.GetByValue(20)
	if exists {
		t.Error("Value 20 should not exist after deletion")
	}

	// Delete non-existing value (should not panic)
	bimap.DeleteByValue(99)
}

func TestBiMapGetAll(t *testing.T) {
	bimap := liveview.NewBiMap[string, int]()

	bimap.Set("a", 1)
	bimap.Set("b", 2)
	bimap.Set("c", 3)

	all := bimap.GetAll()
	if len(all) != 3 {
		t.Errorf("Expected 3 entries in GetAll, got %d", len(all))
	}

	// Verify all entries
	if all["a"] != 1 || all["b"] != 2 || all["c"] != 3 {
		t.Error("GetAll returned incorrect map")
	}
}

func TestBiMapGetAllValues(t *testing.T) {
	bimap := liveview.NewBiMap[string, int]()

	bimap.Set("a", 1)
	bimap.Set("b", 2)
	bimap.Set("c", 3)

	allValues := bimap.GetAllValues()
	if len(allValues) != 3 {
		t.Errorf("Expected 3 entries in GetAllValues, got %d", len(allValues))
	}

	// Verify all entries
	if allValues[1] != "a" || allValues[2] != "b" || allValues[3] != "c" {
		t.Error("GetAllValues returned incorrect map")
	}
}

func TestBiMapComplexScenarios(t *testing.T) {
	bimap := liveview.NewBiMap[string, int]()

	// Test that Set doesn't update existing keys
	bimap.Set("key1", 100)
	bimap.Set("key2", 200)
	bimap.Set("key1", 150) // Should not update

	val, _ := bimap.Get("key1")
	if val != 100 {
		t.Errorf("Expected value 100 for key1 (no update), got %d", val)
	}

	// Test with empty BiMap operations
	emptyBimap := liveview.NewBiMap[string, string]()
	emptyBimap.Delete("nonexistent")
	emptyBimap.DeleteByValue("nonexistent")
	
	all := emptyBimap.GetAll()
	if len(all) != 0 {
		t.Error("Empty BiMap GetAll should return empty map")
	}
}

func TestBiMapWithDifferentTypes(t *testing.T) {
	// Test with int keys and string values
	intStringMap := liveview.NewBiMap[int, string]()
	
	intStringMap.Set(1, "one")
	intStringMap.Set(2, "two")
	intStringMap.Set(3, "three")

	val, exists := intStringMap.Get(2)
	if !exists || val != "two" {
		t.Errorf("Get(2) = %s, %v; want 'two', true", val, exists)
	}

	key, exists := intStringMap.GetByValue("three")
	if !exists || key != 3 {
		t.Errorf("GetByValue('three') = %d, %v; want 3, true", key, exists)
	}

	// Test with custom struct types
	type Person struct {
		ID   int
		Name string
	}

	personMap := liveview.NewBiMap[string, Person]()
	
	p1 := Person{ID: 1, Name: "Alice"}
	p2 := Person{ID: 2, Name: "Bob"}
	
	personMap.Set("user1", p1)
	personMap.Set("user2", p2)

	person, exists := personMap.Get("user1")
	if !exists || person.Name != "Alice" {
		t.Error("Failed to retrieve person by key")
	}
}

func TestBiMapThreadSafety(t *testing.T) {
	// This test verifies that BiMap operations don't panic under concurrent access
	bimap := liveview.NewBiMap[int, int]()

	// Pre-populate some data
	for i := 0; i < 10; i++ {
		bimap.Set(i, i*10)
	}

	done := make(chan bool)

	// Concurrent reads
	go func() {
		for i := 0; i < 100; i++ {
			bimap.Get(i % 10)
			bimap.GetByValue(i % 100)
		}
		done <- true
	}()

	// Concurrent writes
	go func() {
		for i := 10; i < 20; i++ {
			bimap.Set(i, i*10)
		}
		done <- true
	}()

	// Concurrent deletes
	go func() {
		for i := 0; i < 5; i++ {
			bimap.Delete(i)
		}
		done <- true
	}()

	// Wait for all goroutines
	for i := 0; i < 3; i++ {
		<-done
	}

	// If we get here without panic, thread safety is working
	t.Log("BiMap thread safety test passed")
}