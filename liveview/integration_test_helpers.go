package liveview

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// IntegrationTestSuite provides a complete test environment for integration tests
type IntegrationTestSuite struct {
	T            *testing.T
	Server       *httptest.Server
	Echo         *echo.Echo
	WebSocketURL string
	BaseURL      string
	Components   map[string]Component
	Clients      []*IntegrationClient
	mu           sync.RWMutex
	ctx          context.Context
	cancel       context.CancelFunc
}

// IntegrationClient represents a test client for integration tests
type IntegrationClient struct {
	ID           string
	WebSocket    *websocket.Conn
	HTTPClient   *http.Client
	ReceivedMsgs []map[string]interface{}
	SentMsgs     []map[string]interface{}
	Connected    bool
	mu           sync.RWMutex
}

// NewIntegrationTestSuite creates a new integration test suite
func NewIntegrationTestSuite(t *testing.T) *IntegrationTestSuite {
	ctx, cancel := context.WithCancel(context.Background())
	e := echo.New()

	suite := &IntegrationTestSuite{
		T:          t,
		Echo:       e,
		Components: make(map[string]Component),
		Clients:    make([]*IntegrationClient, 0),
		ctx:        ctx,
		cancel:     cancel,
	}

	// Setup cleanup
	t.Cleanup(func() {
		suite.Cleanup()
	})

	return suite
}

// SetupPage registers a page with components for testing
func (s *IntegrationTestSuite) SetupPage(path string, title string, setupFunc func() LiveDriver) {
	page := PageControl{
		Title:  title,
		Path:   path,
		Router: s.Echo,
	}
	page.Register(setupFunc)
}

// Start starts the test server
func (s *IntegrationTestSuite) Start() {
	s.Server = httptest.NewServer(s.Echo)
	s.BaseURL = s.Server.URL
	s.WebSocketURL = "ws" + s.BaseURL[4:] + "/ws_goliveview"
}

// NewClient creates a new test client
func (s *IntegrationTestSuite) NewClient(id string) *IntegrationClient {
	client := &IntegrationClient{
		ID:           id,
		HTTPClient:   s.Server.Client(),
		ReceivedMsgs: make([]map[string]interface{}, 0),
		SentMsgs:     make([]map[string]interface{}, 0),
	}

	s.mu.Lock()
	s.Clients = append(s.Clients, client)
	s.mu.Unlock()

	return client
}

// Connect establishes WebSocket connection for a client
func (c *IntegrationClient) Connect(wsURL string) error {
	dialer := websocket.Dialer{}
	conn, _, err := dialer.Dial(wsURL, nil)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	c.WebSocket = conn
	c.Connected = true

	// Start message reader
	go c.readMessages()

	return nil
}

// readMessages reads incoming WebSocket messages
func (c *IntegrationClient) readMessages() {
	for c.Connected {
		var msg map[string]interface{}
		err := c.WebSocket.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				// Connection closed unexpectedly
			}
			c.Connected = false
			return
		}

		c.mu.Lock()
		c.ReceivedMsgs = append(c.ReceivedMsgs, msg)
		c.mu.Unlock()
	}
}

// SendEvent sends an event through WebSocket
func (c *IntegrationClient) SendEvent(componentID string, eventName string, data interface{}) error {
	if !c.Connected {
		return fmt.Errorf("client not connected")
	}

	msg := map[string]interface{}{
		"type":  "data",
		"id":    componentID,
		"event": eventName,
		"data":  data,
	}

	c.mu.Lock()
	c.SentMsgs = append(c.SentMsgs, msg)
	c.mu.Unlock()

	return c.WebSocket.WriteJSON(msg)
}

// GetPage makes an HTTP GET request to a page
func (c *IntegrationClient) GetPage(path string) (*http.Response, string, error) {
	resp, err := c.HTTPClient.Get(path)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp, "", err
	}

	return resp, string(body), nil
}

// WaitForMessage waits for a specific message type
func (c *IntegrationClient) WaitForMessage(msgType string, timeout time.Duration) (map[string]interface{}, error) {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		c.mu.RLock()
		for _, msg := range c.ReceivedMsgs {
			if t, ok := msg["type"].(string); ok && t == msgType {
				c.mu.RUnlock()
				return msg, nil
			}
		}
		c.mu.RUnlock()

		time.Sleep(10 * time.Millisecond)
	}

	return nil, fmt.Errorf("timeout waiting for message type: %s", msgType)
}

// WaitForDOMUpdate waits for a DOM update with specific ID
func (c *IntegrationClient) WaitForDOMUpdate(elementID string, timeout time.Duration) (string, error) {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		c.mu.RLock()
		for _, msg := range c.ReceivedMsgs {
			if msg["type"] == "fill" && msg["id"] == elementID {
				if value, ok := msg["value"].(string); ok {
					c.mu.RUnlock()
					return value, nil
				}
			}
		}
		c.mu.RUnlock()

		time.Sleep(10 * time.Millisecond)
	}

	return "", fmt.Errorf("timeout waiting for DOM update on element: %s", elementID)
}

// AssertMessageReceived asserts that a message with specific type was received
func (c *IntegrationClient) AssertMessageReceived(t *testing.T, msgType string) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, msg := range c.ReceivedMsgs {
		if mt, ok := msg["type"].(string); ok && mt == msgType {
			return
		}
	}

	t.Errorf("Expected message type '%s' was not received", msgType)
}

// Disconnect closes the WebSocket connection
func (c *IntegrationClient) Disconnect() error {
	c.Connected = false
	if c.WebSocket != nil {
		return c.WebSocket.Close()
	}
	return nil
}

// Cleanup cleans up the test suite
func (s *IntegrationTestSuite) Cleanup() {
	// Disconnect all clients
	for _, client := range s.Clients {
		client.Disconnect()
	}

	// Cancel context
	if s.cancel != nil {
		s.cancel()
	}

	// Close server
	if s.Server != nil {
		s.Server.Close()
	}
}

// Integration Test Scenarios

// TestMultipleClients tests multiple clients interacting with the same component
func TestMultipleClients(t *testing.T, suite *IntegrationTestSuite, componentID string) {
	// Create multiple clients
	client1 := suite.NewClient("client1")
	client2 := suite.NewClient("client2")
	client3 := suite.NewClient("client3")

	// Connect all clients
	require.NoError(t, client1.Connect(suite.WebSocketURL))
	require.NoError(t, client2.Connect(suite.WebSocketURL))
	require.NoError(t, client3.Connect(suite.WebSocketURL))

	// Client 1 sends event
	err := client1.SendEvent(componentID, "Update", "data from client1")
	require.NoError(t, err)

	// All clients should receive update
	msg1, err := client1.WaitForMessage("fill", 2*time.Second)
	require.NoError(t, err)
	assert.NotNil(t, msg1)

	msg2, err := client2.WaitForMessage("fill", 2*time.Second)
	require.NoError(t, err)
	assert.NotNil(t, msg2)

	msg3, err := client3.WaitForMessage("fill", 2*time.Second)
	require.NoError(t, err)
	assert.NotNil(t, msg3)
}

// TestComponentStateSync tests state synchronization between server and client
func TestComponentStateSync(t *testing.T, suite *IntegrationTestSuite, componentID string) {
	client := suite.NewClient("test-client")

	// Get initial page
	resp, body, err := client.GetPage("/")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, body, componentID)

	// Connect WebSocket
	require.NoError(t, client.Connect(suite.WebSocketURL))

	// Send multiple events
	events := []string{"Event1", "Event2", "Event3"}
	for _, event := range events {
		err := client.SendEvent(componentID, event, nil)
		require.NoError(t, err)
		time.Sleep(100 * time.Millisecond)
	}

	// Verify all updates received
	client.mu.RLock()
	assert.GreaterOrEqual(t, len(client.ReceivedMsgs), len(events))
	client.mu.RUnlock()
}

// TestReconnection tests client reconnection scenario
func TestReconnection(t *testing.T, suite *IntegrationTestSuite, componentID string) {
	client := suite.NewClient("reconnect-client")

	// Initial connection
	require.NoError(t, client.Connect(suite.WebSocketURL))

	// Send event
	err := client.SendEvent(componentID, "InitialEvent", "data1")
	require.NoError(t, err)

	// Wait for response
	msg, err := client.WaitForMessage("fill", 2*time.Second)
	require.NoError(t, err)
	assert.NotNil(t, msg)

	// Disconnect
	require.NoError(t, client.Disconnect())

	// Reconnect
	require.NoError(t, client.Connect(suite.WebSocketURL))

	// Send event after reconnection
	err = client.SendEvent(componentID, "AfterReconnect", "data2")
	require.NoError(t, err)

	// Should receive update
	msg, err = client.WaitForMessage("fill", 2*time.Second)
	require.NoError(t, err)
	assert.NotNil(t, msg)
}

// TestConcurrentEvents tests handling of concurrent events
func TestConcurrentEvents(t *testing.T, suite *IntegrationTestSuite, componentID string) {
	client := suite.NewClient("concurrent-client")
	require.NoError(t, client.Connect(suite.WebSocketURL))

	// Send many events concurrently
	var wg sync.WaitGroup
	eventCount := 50

	for i := 0; i < eventCount; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			eventName := fmt.Sprintf("ConcurrentEvent%d", index)
			client.SendEvent(componentID, eventName, index)
		}(i)
	}

	wg.Wait()

	// Give some time for all messages to be processed
	time.Sleep(2 * time.Second)

	// Verify messages were sent
	client.mu.RLock()
	assert.Equal(t, eventCount, len(client.SentMsgs))
	client.mu.RUnlock()
}

// TestLoadScenario simulates a load test scenario
func TestLoadScenario(t *testing.T, suite *IntegrationTestSuite, componentID string, clientCount int, eventsPerClient int) {
	clients := make([]*IntegrationClient, clientCount)

	// Create and connect clients
	for i := 0; i < clientCount; i++ {
		client := suite.NewClient(fmt.Sprintf("load-client-%d", i))
		require.NoError(t, client.Connect(suite.WebSocketURL))
		clients[i] = client
	}

	// Each client sends events
	var wg sync.WaitGroup
	startTime := time.Now()

	for i, client := range clients {
		wg.Add(1)
		go func(c *IntegrationClient, clientIndex int) {
			defer wg.Done()

			for j := 0; j < eventsPerClient; j++ {
				eventName := fmt.Sprintf("LoadEvent_%d_%d", clientIndex, j)
				c.SendEvent(componentID, eventName, j)
				time.Sleep(50 * time.Millisecond)
			}
		}(client, i)
	}

	wg.Wait()
	duration := time.Since(startTime)

	// Calculate metrics
	totalEvents := clientCount * eventsPerClient
	eventsPerSecond := float64(totalEvents) / duration.Seconds()

	t.Logf("Load test completed: %d clients, %d events each", clientCount, eventsPerClient)
	t.Logf("Total events: %d, Duration: %v, Events/sec: %.2f", totalEvents, duration, eventsPerSecond)

	// Verify all clients are still connected
	for _, client := range clients {
		assert.True(t, client.Connected)
	}
}

// Helper function to run a complete integration test
func RunIntegrationTest(t *testing.T, componentSetup func() LiveDriver) {
	suite := NewIntegrationTestSuite(t)

	// Setup page with component
	suite.SetupPage("/test", "Integration Test", componentSetup)

	// Start server
	suite.Start()

	// Run test scenarios
	t.Run("MultipleClients", func(t *testing.T) {
		TestMultipleClients(t, suite, "test-component")
	})

	t.Run("StateSync", func(t *testing.T) {
		TestComponentStateSync(t, suite, "test-component")
	})

	t.Run("Reconnection", func(t *testing.T) {
		TestReconnection(t, suite, "test-component")
	})

	t.Run("ConcurrentEvents", func(t *testing.T) {
		TestConcurrentEvents(t, suite, "test-component")
	})

	t.Run("LoadTest", func(t *testing.T) {
		TestLoadScenario(t, suite, "test-component", 10, 20)
	})
}

// AssertComponentState helper to verify component state
func AssertComponentState(t *testing.T, component Component, assertions func(Component)) {
	assertions(component)
}

// WaitForComponentUpdate waits for a component to update
func WaitForComponentUpdate(component Component, timeout time.Duration, condition func(Component) bool) error {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		if condition(component) {
			return nil
		}
		time.Sleep(10 * time.Millisecond)
	}

	return fmt.Errorf("timeout waiting for component update")
}

// CreateTestComponent creates a test component with preset data
func CreateTestComponent(id string, template string, data interface{}) Component {
	// This would need to be implemented based on your component structure
	// For now, returning nil as placeholder
	return nil
}

// VerifyWebSocketMessage verifies WebSocket message structure
func VerifyWebSocketMessage(t *testing.T, msg map[string]interface{}, expectedType string, expectedID string) {
	msgType, ok := msg["type"].(string)
	assert.True(t, ok, "Message should have type field")
	assert.Equal(t, expectedType, msgType)

	if expectedID != "" {
		msgID, ok := msg["id"].(string)
		assert.True(t, ok, "Message should have id field")
		assert.Equal(t, expectedID, msgID)
	}
}

// SimulateUserJourney simulates a complete user journey
func SimulateUserJourney(t *testing.T, suite *IntegrationTestSuite, journey []struct {
	Action string
	Data   interface{}
	Wait   time.Duration
}) {
	client := suite.NewClient("journey-client")
	require.NoError(t, client.Connect(suite.WebSocketURL))

	for i, step := range journey {
		t.Logf("Step %d: %s", i+1, step.Action)

		// Parse action and execute
		parts := strings.Split(step.Action, ":")
		if len(parts) == 2 {
			componentID := parts[0]
			eventName := parts[1]

			err := client.SendEvent(componentID, eventName, step.Data)
			require.NoError(t, err)
		}

		// Wait if specified
		if step.Wait > 0 {
			time.Sleep(step.Wait)
		}
	}

	// Verify journey completed successfully
	client.mu.RLock()
	assert.Greater(t, len(client.ReceivedMsgs), 0, "Should have received responses")
	client.mu.RUnlock()
}
