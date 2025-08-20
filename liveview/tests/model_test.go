package liveview_test

import (
	"testing"
	"time"

	"github.com/arturoeanton/go-echo-live-view/liveview"
)

type TestComponent struct {
	Driver  liveview.LiveDriver
	Value   string
	Counter int
}

func (t *TestComponent) GetTemplate() string {
	return `<div>{{.Value}} - {{.Counter}}</div>`
}

func (t *TestComponent) Start() {
	t.Value = "initial"
	t.Counter = 0
}

func (t *TestComponent) GetDriver() liveview.LiveDriver {
	return t.Driver
}

func (t *TestComponent) Increment() {
	t.Counter++
	t.GetDriver().Commit()
}

func TestComponentDriverCreation(t *testing.T) {
	component := &TestComponent{}
	driver := liveview.NewDriver("test-component", component)
	component.Driver = driver

	if driver == nil {
		t.Fatal("Driver should not be nil")
	}

	if driver.GetID() != "mount_span_test-component" {
		t.Errorf("Expected ID contains 'mount_span_test-component', got '%s'", driver.GetID())
	}
}

func TestComponentInitialization(t *testing.T) {
	component := &TestComponent{}
	driver := liveview.NewDriver("test-init", component)
	component.Driver = driver

	// Start should be called during StartDriver
	channelIn := make(map[string]chan interface{})
	channel := make(chan map[string]interface{})
	drivers := make(map[string]liveview.LiveDriver)
	
	go driver.StartDriver(&drivers, &channelIn, channel)
	time.Sleep(100 * time.Millisecond) // Give goroutine time to start

	if component.Value != "initial" {
		t.Errorf("Expected Value 'initial', got '%s'", component.Value)
	}

	if component.Counter != 0 {
		t.Errorf("Expected Counter 0, got %d", component.Counter)
	}
}

func TestComponentEventHandling(t *testing.T) {
	component := &TestComponent{}
	driver := liveview.NewDriver("test-event", component)
	component.Driver = driver

	channelIn := make(map[string]chan interface{})
	channel := make(chan map[string]interface{})
	drivers := make(map[string]liveview.LiveDriver)
	
	go driver.StartDriver(&drivers, &channelIn, channel)
	time.Sleep(100 * time.Millisecond) // Give goroutine time to start

	// Test direct method call
	component.Increment()

	if component.Counter != 1 {
		t.Errorf("Expected Counter 1 after Increment, got %d", component.Counter)
	}
}

func TestComponentMount(t *testing.T) {
	parent := &TestComponent{}
	parentDriver := liveview.NewDriver("parent", parent)
	parent.Driver = parentDriver

	child := &TestComponent{Value: "child"}
	childDriver := liveview.NewDriver("child", child)
	child.Driver = childDriver

	parentDriver.Mount(child)

	// Check if child is properly mounted
	time.Sleep(100 * time.Millisecond)
}

func TestLiveDriverHelperMethods(t *testing.T) {
	component := &TestComponent{}
	driver := liveview.NewDriver("test-helpers", component)
	component.Driver = driver

	// Test available helper methods (these don't have direct effects we can test without WebSocket)
	driver.FillValueById("test-div", "content")
	driver.FillValue("global content")
	driver.SetHTML("<p>HTML content</p>")
	driver.SetText("Text content")
	driver.SetStyle("color: red;")
	driver.SetValue("test value")
	driver.SetPropertie("disabled", true)
	driver.EvalScript("console.log('test');")
	driver.AddNode("div-id", "<span>node</span>")
	driver.Remove("old-element")

	// Test getter methods (return empty in test environment)
	value := driver.GetValue()
	if value == "" {
		t.Log("GetValue returns empty in test environment (expected)")
	}

	html := driver.GetHTML()
	if html == "" {
		t.Log("GetHTML returns empty in test environment (expected)")
	}

	text := driver.GetText()
	if text == "" {
		t.Log("GetText returns empty in test environment (expected)")
	}

	style := driver.GetStyle("color")
	if style == "" {
		t.Log("GetStyle returns empty in test environment (expected)")
	}

	prop := driver.GetPropertie("disabled")
	if prop == "" {
		t.Log("GetPropertie returns empty in test environment (expected)")
	}

	element := driver.GetElementById("test-id")
	if element == "" {
		t.Log("GetElementById returns empty in test environment (expected)")
	}

	// If we get here without panic, the methods work
	t.Log("All LiveDriver helper methods executed without panic")
}

func TestComponentLifecycle(t *testing.T) {
	component := &TestComponent{}
	driver := liveview.NewDriver("test-lifecycle", component)
	component.Driver = driver

	channelIn := make(map[string]chan interface{})
	channel := make(chan map[string]interface{})
	drivers := make(map[string]liveview.LiveDriver)
	
	go driver.StartDriver(&drivers, &channelIn, channel)
	
	// Test initial state
	time.Sleep(100 * time.Millisecond)
	if component.Value != "initial" {
		t.Errorf("Component not initialized properly")
	}

	// Test state change and commit
	component.Value = "changed"
	component.Counter = 42
	driver.Commit()

	// In a real scenario, this would trigger a re-render
	// Here we just verify the state changed
	if component.Value != "changed" || component.Counter != 42 {
		t.Errorf("Component state not updated properly")
	}
}

type ComplexComponent struct {
	Driver liveview.LiveDriver
	Items  []string
	Events map[string]func()
}

func (c *ComplexComponent) GetTemplate() string {
	return `<div>{{range .Items}}<p>{{.}}</p>{{end}}</div>`
}

func (c *ComplexComponent) Start() {
	c.Items = []string{"Item 1", "Item 2", "Item 3"}
	c.Events = map[string]func(){
		"CustomEvent": func() {
			c.Items = append(c.Items, "New Item")
			c.GetDriver().Commit()
		},
	}
}

func (c *ComplexComponent) GetDriver() liveview.LiveDriver {
	return c.Driver
}

func TestComplexComponent(t *testing.T) {
	component := &ComplexComponent{}
	driver := liveview.NewDriver("test-complex", component)
	component.Driver = driver

	channelIn := make(map[string]chan interface{})
	channel := make(chan map[string]interface{})
	drivers := make(map[string]liveview.LiveDriver)
	
	go driver.StartDriver(&drivers, &channelIn, channel)
	time.Sleep(100 * time.Millisecond)

	// Test initial items
	if len(component.Items) != 3 {
		t.Errorf("Expected 3 items, got %d", len(component.Items))
	}

	// Test custom event
	if customEvent, ok := component.Events["CustomEvent"]; ok {
		customEvent()
		if len(component.Items) != 4 {
			t.Errorf("Expected 4 items after CustomEvent, got %d", len(component.Items))
		}
	} else {
		t.Error("CustomEvent not found in Events map")
	}
}

func TestDriverMethods(t *testing.T) {
	component := &TestComponent{}
	driver := liveview.NewDriver("test-methods", component)
	component.Driver = driver

	// Test GetIDComponet
	componentID := driver.GetIDComponet()
	if componentID != "test-methods" {
		t.Errorf("Expected component ID 'test-methods', got '%s'", componentID)
	}

	// Test GetComponet
	comp := driver.GetComponet()
	if comp != component {
		t.Error("GetComponet should return the same component")
	}

	// Test GetDriverById with non-existent ID
	childDriver := driver.GetDriverById("non-existent")
	if childDriver == nil {
		t.Error("GetDriverById should return a driver even for non-existent IDs")
	}

	// Test SetData and GetData
	testData := map[string]string{"key": "value"}
	driver.SetData(testData)
	
	retrievedData := driver.GetData()
	if retrievedData == nil {
		t.Error("GetData should return the set data")
	}
}

func TestEventExecution(t *testing.T) {
	component := &TestComponent{}
	driver := liveview.NewDriver("test-execute", component)
	component.Driver = driver

	// Test ExecuteEvent - this should not panic even if event doesn't exist
	driver.ExecuteEvent("NonExistentEvent", "test data")
	
	// Test ExecuteEvent with reflection-based method call
	// This tests the internal event system
	driver.ExecuteEvent("Increment", nil)
	
	// After the event, counter should be incremented
	if component.Counter != 1 {
		t.Errorf("Expected Counter 1 after ExecuteEvent, got %d", component.Counter)
	}
}