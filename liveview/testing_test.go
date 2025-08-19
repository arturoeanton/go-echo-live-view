package liveview_test

import (
	"testing"
	"time"

	"github.com/arturoeanton/go-echo-live-view/liveview"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Example component for testing
type TestCounter struct {
	*liveview.ComponentDriver[*TestCounter]
	Count      int
	LastAction string
}

func (c *TestCounter) Start() {
	c.Count = 0
	c.LastAction = "initialized"
	c.Commit()
}

func (c *TestCounter) GetTemplate() string {
	return `
	<div id="counter">
		<span id="count">{{.Count}}</span>
		<span id="action">{{.LastAction}}</span>
		<button onclick="send_event('{{.IdComponent}}', 'Increment')">+</button>
		<button onclick="send_event('{{.IdComponent}}', 'Decrement')">-</button>
		<button onclick="send_event('{{.IdComponent}}', 'Reset')">Reset</button>
	</div>
	`
}

func (c *TestCounter) GetDriver() liveview.LiveDriver {
	return c
}

func (c *TestCounter) Increment(data interface{}) {
	c.Count++
	c.LastAction = "incremented"
	c.Commit()
}

func (c *TestCounter) Decrement(data interface{}) {
	c.Count--
	c.LastAction = "decremented"
	c.Commit()
}

func (c *TestCounter) Reset(data interface{}) {
	c.Count = 0
	c.LastAction = "reset"
	c.Commit()
}

// Test basic component functionality
func TestComponentBasics(t *testing.T) {
	// Create test component
	counter := &TestCounter{}
	td := liveview.NewTestDriver(t, counter, "test-counter")
	defer td.Cleanup()

	// Test initial state
	assert.Equal(t, 0, counter.Count)
	assert.Equal(t, "initialized", counter.LastAction)

	// Test HTML rendering
	td.AssertHTML(t, `<span id="count">0</span>`)
	td.AssertHTML(t, `<span id="action">initialized</span>`)
}

// Test event handling
func TestEventHandling(t *testing.T) {
	counter := &TestCounter{}
	td := liveview.NewTestDriver(t, counter, "test-counter")
	defer td.Cleanup()

	// Test increment
	err := td.SimulateEvent("Increment", nil)
	require.NoError(t, err)
	assert.Equal(t, 1, counter.Count)
	assert.Equal(t, "incremented", counter.LastAction)

	// Test multiple increments
	for i := 0; i < 5; i++ {
		td.SimulateEvent("Increment", nil)
	}
	assert.Equal(t, 6, counter.Count)

	// Test decrement
	td.SimulateEvent("Decrement", nil)
	assert.Equal(t, 5, counter.Count)
	assert.Equal(t, "decremented", counter.LastAction)

	// Test reset
	td.SimulateEvent("Reset", nil)
	assert.Equal(t, 0, counter.Count)
	assert.Equal(t, "reset", counter.LastAction)
}

// Test HTML updates after events
func TestHTMLUpdates(t *testing.T) {
	counter := &TestCounter{}
	td := liveview.NewTestDriver(t, counter, "test-counter")
	defer td.Cleanup()

	// Initial HTML
	td.AssertHTML(t, `<span id="count">0</span>`)

	// After increment
	td.SimulateEvent("Increment", nil)
	td.AssertHTML(t, `<span id="count">1</span>`)
	td.AssertHTML(t, `<span id="action">incremented</span>`)

	// After multiple operations
	td.SimulateEvent("Increment", nil)
	td.SimulateEvent("Increment", nil)
	td.AssertHTML(t, `<span id="count">3</span>`)

	td.SimulateEvent("Reset", nil)
	td.AssertHTML(t, `<span id="count">0</span>`)
	td.AssertHTML(t, `<span id="action">reset</span>`)
}

// Test component with child components
type ParentComponent struct {
	*liveview.ComponentDriver[*ParentComponent]
	Title        string
	ChildCounter *TestCounter
}

func (p *ParentComponent) Start() {
	p.Title = "Parent Component"

	// Mount child component
	p.ChildCounter = &TestCounter{}
	p.Mount(liveview.New("child-counter", p.ChildCounter))

	p.Commit()
}

func (p *ParentComponent) GetTemplate() string {
	return `
	<div id="parent">
		<h1>{{.Title}}</h1>
		{{mount "child-counter"}}
	</div>
	`
}

func (p *ParentComponent) GetDriver() liveview.LiveDriver {
	return p
}

func TestComponentComposition(t *testing.T) {
	parent := &ParentComponent{}
	td := liveview.NewTestDriver(t, parent, "test-parent")
	defer td.Cleanup()

	// Test parent initialization
	assert.Equal(t, "Parent Component", parent.Title)
	assert.NotNil(t, parent.ChildCounter)

	// Test child component state
	assert.Equal(t, 0, parent.ChildCounter.Count)

	// Test events on child component
	td.SimulateEventWithID("child-counter", "Increment", nil)
	assert.Equal(t, 1, parent.ChildCounter.Count)
}

// Test WebSocket client mock
func TestMockWebSocketClient(t *testing.T) {
	// Create test suite
	suite := liveview.NewComponentTestSuite()

	// Register test component
	suite.RegisterComponent("/", func() liveview.LiveDriver {
		counter := &TestCounter{}
		driver := liveview.NewDriver("counter", counter)
		counter.ComponentDriver = driver
		return driver
	})

	// Start server
	suite.Start()
	defer suite.Stop()

	// Create WebSocket client
	client := liveview.NewMockWebSocketClient(suite.GetWebSocketURL())

	// Connect
	err := client.Connect()
	require.NoError(t, err)
	defer client.Close()

	// Send event
	err = client.SendEvent("counter", "Increment", nil)
	require.NoError(t, err)

	// Wait for response
	msg, err := client.WaitForMessage("fill", 2*time.Second)
	require.NoError(t, err)
	assert.NotNil(t, msg)
}

// Benchmark component rendering
func BenchmarkComponentRendering(b *testing.B) {
	counter := &TestCounter{}
	counter.Start()

	liveview.BenchmarkComponent(b, counter)
}

// Benchmark event handling
func BenchmarkEventHandling(b *testing.B) {
	counter := &TestCounter{}

	liveview.BenchmarkEventHandling(b, counter, "Increment", nil)
}

// Integration test example
func TestIntegration(t *testing.T) {
	suite, client := liveview.CreateIntegrationTest(t)

	// Register component
	suite.RegisterComponent("/test", func() liveview.LiveDriver {
		counter := &TestCounter{}
		driver := liveview.NewDriver("counter", counter)
		counter.ComponentDriver = driver
		return driver
	})

	// Connect client
	err := client.Connect()
	require.NoError(t, err)

	// Send events and verify responses
	err = client.SendEvent("counter", "Increment", nil)
	require.NoError(t, err)

	// Wait for update
	msg, err := client.WaitForMessage("fill", 2*time.Second)
	require.NoError(t, err)
	assert.NotNil(t, msg)
}

// Test helper functions
func TestHelperFunctions(t *testing.T) {
	counter := &TestCounter{}
	td := liveview.NewTestDriver(t, counter, "test-counter")
	defer td.Cleanup()

	// Test SimulateClick helper
	err := liveview.SimulateClick(td, "test-counter")
	require.NoError(t, err)

	// Test SimulateInput helper
	err = liveview.SimulateInput(td, "input-field", "test value")
	require.NoError(t, err)

	// Test WaitForCondition
	done := false
	go func() {
		time.Sleep(100 * time.Millisecond)
		done = true
	}()

	err = liveview.WaitForCondition(1*time.Second, func() bool {
		return done
	})
	require.NoError(t, err)
}
