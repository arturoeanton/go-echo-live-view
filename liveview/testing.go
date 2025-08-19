package liveview

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDriver provides a test harness for LiveView components.
// It simulates WebSocket communication and DOM updates for testing.
type TestDriver struct {
	Component  Component
	Driver     LiveDriver
	Messages   []map[string]interface{}
	DOMUpdates map[string]string
	EventQueue []Event
	Context    context.Context
	CancelFunc context.CancelFunc
	wsConn     *websocket.Conn
	server     *httptest.Server
	echoServer *echo.Echo
}

// Event represents a test event to be sent to a component
type Event struct {
	ComponentID string
	EventName   string
	Data        interface{}
}

// NewTestDriver creates a new test driver for a component
func NewTestDriver(t *testing.T, component Component, componentID string) *TestDriver {
	driver := NewDriver(componentID, component)

	// Set the driver on the component if it has a ComponentDriver field
	if setter, ok := component.(interface{ SetDriver(LiveDriver) }); ok {
		setter.SetDriver(driver)
	}

	ctx, cancel := context.WithCancel(context.Background())

	td := &TestDriver{
		Component:  component,
		Driver:     driver,
		Messages:   make([]map[string]interface{}, 0),
		DOMUpdates: make(map[string]string),
		EventQueue: make([]Event, 0),
		Context:    ctx,
		CancelFunc: cancel,
	}

	// Initialize component
	component.Start()

	return td
}

// SimulateEvent simulates sending an event from the client
func (td *TestDriver) SimulateEvent(eventName string, data interface{}) error {
	td.Driver.ExecuteEvent(eventName, data)
	return nil
}

// SimulateEventWithID simulates an event for a specific component ID
func (td *TestDriver) SimulateEventWithID(componentID, eventName string, data interface{}) error {
	event := Event{
		ComponentID: componentID,
		EventName:   eventName,
		Data:        data,
	}
	td.EventQueue = append(td.EventQueue, event)

	// Find and execute on the specific driver
	if driver := td.Driver.GetDriverById(componentID); driver != nil {
		driver.ExecuteEvent(eventName, data)
	}

	return nil
}

// GetHTML returns the rendered HTML of the component
func (td *TestDriver) GetHTML() (string, error) {
	tmpl := td.Component.GetTemplate()

	// Parse and execute template
	t, err := ParseTemplate(tmpl)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, td.Component); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// AssertHTML checks if the rendered HTML contains the expected string
func (td *TestDriver) AssertHTML(t *testing.T, expected string) {
	html, err := td.GetHTML()
	require.NoError(t, err)
	assert.Contains(t, html, expected)
}

// AssertNotHTML checks if the rendered HTML does not contain the string
func (td *TestDriver) AssertNotHTML(t *testing.T, notExpected string) {
	html, err := td.GetHTML()
	require.NoError(t, err)
	assert.NotContains(t, html, notExpected)
}

// GetComponentState returns the current state of the component
func (td *TestDriver) GetComponentState() interface{} {
	return td.Component
}

// Cleanup cleans up test resources
func (td *TestDriver) Cleanup() {
	if td.CancelFunc != nil {
		td.CancelFunc()
	}
	if td.wsConn != nil {
		td.wsConn.Close()
	}
	if td.server != nil {
		td.server.Close()
	}
}

// MockWebSocketClient provides a mock WebSocket client for testing
type MockWebSocketClient struct {
	URL              string
	Messages         []interface{}
	ReceivedMessages []map[string]interface{}
	conn             *websocket.Conn
	connected        bool
}

// NewMockWebSocketClient creates a new mock WebSocket client
func NewMockWebSocketClient(url string) *MockWebSocketClient {
	return &MockWebSocketClient{
		URL:              url,
		Messages:         make([]interface{}, 0),
		ReceivedMessages: make([]map[string]interface{}, 0),
		connected:        false,
	}
}

// Connect establishes a WebSocket connection
func (c *MockWebSocketClient) Connect() error {
	dialer := websocket.Dialer{
		HandshakeTimeout: 5 * time.Second,
	}

	conn, _, err := dialer.Dial(c.URL, nil)
	if err != nil {
		return err
	}

	c.conn = conn
	c.connected = true

	// Start reading messages
	go c.readMessages()

	return nil
}

// readMessages reads messages from the WebSocket connection
func (c *MockWebSocketClient) readMessages() {
	for c.connected {
		var msg map[string]interface{}
		err := c.conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				fmt.Printf("WebSocket error: %v\n", err)
			}
			break
		}
		c.ReceivedMessages = append(c.ReceivedMessages, msg)
	}
}

// SendEvent sends an event message to the server
func (c *MockWebSocketClient) SendEvent(componentID, eventName string, data interface{}) error {
	if !c.connected {
		return fmt.Errorf("not connected")
	}

	msg := map[string]interface{}{
		"type":  "data",
		"id":    componentID,
		"event": eventName,
		"data":  data,
	}

	return c.conn.WriteJSON(msg)
}

// SendMessage sends a raw message to the server
func (c *MockWebSocketClient) SendMessage(msg interface{}) error {
	if !c.connected {
		return fmt.Errorf("not connected")
	}

	return c.conn.WriteJSON(msg)
}

// Close closes the WebSocket connection
func (c *MockWebSocketClient) Close() error {
	c.connected = false
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// WaitForMessage waits for a specific message type with timeout
func (c *MockWebSocketClient) WaitForMessage(msgType string, timeout time.Duration) (map[string]interface{}, error) {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		for _, msg := range c.ReceivedMessages {
			if t, ok := msg["type"].(string); ok && t == msgType {
				return msg, nil
			}
		}
		time.Sleep(10 * time.Millisecond)
	}

	return nil, fmt.Errorf("timeout waiting for message type: %s", msgType)
}

// ComponentTestSuite provides utilities for component testing
type ComponentTestSuite struct {
	Echo   *echo.Echo
	Server *httptest.Server
}

// NewComponentTestSuite creates a new test suite
func NewComponentTestSuite() *ComponentTestSuite {
	e := echo.New()
	return &ComponentTestSuite{
		Echo: e,
	}
}

// RegisterComponent registers a component for testing
func (s *ComponentTestSuite) RegisterComponent(path string, component func() LiveDriver) {
	page := PageControl{
		Title:  "Test Page",
		Path:   path,
		Router: s.Echo,
	}
	page.Register(component)
}

// Start starts the test server
func (s *ComponentTestSuite) Start() {
	s.Server = httptest.NewServer(s.Echo)
}

// Stop stops the test server
func (s *ComponentTestSuite) Stop() {
	if s.Server != nil {
		s.Server.Close()
	}
}

// GetURL returns the test server URL
func (s *ComponentTestSuite) GetURL() string {
	if s.Server != nil {
		return s.Server.URL
	}
	return ""
}

// GetWebSocketURL returns the WebSocket URL for testing
func (s *ComponentTestSuite) GetWebSocketURL() string {
	if s.Server != nil {
		url := strings.Replace(s.Server.URL, "http://", "ws://", 1)
		return url + "/ws_goliveview"
	}
	return ""
}

// TestHelpers provides helper functions for testing

// AssertEventCalled checks if an event was called
func AssertEventCalled(t *testing.T, td *TestDriver, eventName string) {
	found := false
	for _, event := range td.EventQueue {
		if event.EventName == eventName {
			found = true
			break
		}
	}
	assert.True(t, found, "Event %s was not called", eventName)
}

// AssertEventData checks if an event was called with specific data
func AssertEventData(t *testing.T, td *TestDriver, eventName string, expectedData interface{}) {
	for _, event := range td.EventQueue {
		if event.EventName == eventName {
			assert.Equal(t, expectedData, event.Data)
			return
		}
	}
	t.Errorf("Event %s was not found", eventName)
}

// SimulateFormSubmit simulates a form submission
func SimulateFormSubmit(td *TestDriver, formData map[string]string) error {
	data, _ := json.Marshal(formData)
	return td.SimulateEvent("Submit", string(data))
}

// SimulateClick simulates a button click
func SimulateClick(td *TestDriver, buttonID string) error {
	return td.SimulateEventWithID(buttonID, "Click", nil)
}

// SimulateInput simulates text input
func SimulateInput(td *TestDriver, inputID string, value string) error {
	return td.SimulateEventWithID(inputID, "Change", value)
}

// BenchmarkHelpers provides benchmark utilities

// BenchmarkComponent benchmarks component rendering
func BenchmarkComponent(b *testing.B, component Component) {
	td := &TestDriver{
		Component: component,
		Driver:    NewDriver("bench", component),
	}

	component.Start()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = td.GetHTML()
	}
}

// BenchmarkEventHandling benchmarks event handling
func BenchmarkEventHandling(b *testing.B, component Component, eventName string, data interface{}) {
	driver := NewDriver("bench", component)
	component.Start()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		driver.ExecuteEvent(eventName, data)
	}
}

// IntegrationTestHelpers provides integration test utilities

// CreateIntegrationTest creates a full integration test environment
func CreateIntegrationTest(t *testing.T) (*ComponentTestSuite, *MockWebSocketClient) {
	suite := NewComponentTestSuite()
	suite.Start()

	client := NewMockWebSocketClient(suite.GetWebSocketURL())

	t.Cleanup(func() {
		client.Close()
		suite.Stop()
	})

	return suite, client
}

// WaitForCondition waits for a condition to be true with timeout
func WaitForCondition(timeout time.Duration, condition func() bool) error {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		if condition() {
			return nil
		}
		time.Sleep(10 * time.Millisecond)
	}

	return fmt.Errorf("condition not met within timeout")
}

// ParseTemplate is a helper to parse component templates for testing
func ParseTemplate(tmpl string) (*template.Template, error) {
	return template.New("test").Parse(tmpl)
}
