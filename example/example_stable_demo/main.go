package main

import (
	"fmt"
	"log"
	"time"

	"github.com/arturoeanton/go-echo-live-view/liveview"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// StableDemoComponent demonstrates framework features in a stable way
type StableDemoComponent struct {
	*liveview.ComponentDriver[*StableDemoComponent]
	
	// Component state
	Counter        int
	Messages       []string
	Features       []string
	LastAction     string
	StateData      map[string]interface{}
	
	// Framework components (initialized on demand)
	errorBoundary  *liveview.ErrorBoundary
	stateManager   *liveview.StateManager
	eventRegistry  *liveview.EventRegistry
	lifecycle      *liveview.LifecycleManager
}

func (c *StableDemoComponent) Start() {
	// Initialize Error Boundary
	c.errorBoundary = liveview.NewErrorBoundary(100, true)
	c.errorBoundary.SetFallbackRenderer(func(componentID string, err error) string {
		return fmt.Sprintf(`<div class="error">Recovered from: %v</div>`, err)
	})
	
	// Initialize State Manager
	c.stateManager = liveview.NewStateManager(&liveview.StateConfig{
		Provider:     liveview.NewMemoryStateProvider(),
		CacheEnabled: true,
		CacheTTL:     5 * time.Minute,
	})
	
	// Initialize Event Registry
	c.eventRegistry = liveview.NewEventRegistry(&liveview.EventRegistryConfig{
		MaxHandlersPerEvent: 10,
		EnableMetrics:       true,
		EnableWildcards:     true,
	})
	
	// Initialize Lifecycle Manager
	c.lifecycle = liveview.NewLifecycleManager("stable_demo")
	c.lifecycle.SetHooks(&liveview.LifecycleHooks{
		OnCreated: func() error {
			c.addMessage("Component created")
			return nil
		},
		OnMounted: func() error {
			c.addMessage("Component mounted")
			return nil
		},
	})
	
	// Initialize component state
	c.Counter = 0
	c.Messages = []string{"System initialized"}
	c.StateData = make(map[string]interface{})
	c.Features = []string{
		"✅ Error Boundaries - Automatic error recovery",
		"✅ State Management - Persistent state storage",
		"✅ Event Registry - Advanced event handling",
		"✅ Lifecycle Hooks - Component lifecycle control",
		"✅ Template Cache - Performance optimization",
		"✅ SafeScript API - Secure JavaScript execution",
		"✅ Lazy Loading - On-demand component loading",
		"✅ Virtual DOM - Efficient rendering (internal)",
	}
	c.LastAction = "Ready"
	
	// Execute lifecycle
	c.lifecycle.Create()
	c.lifecycle.Mount()
	
	c.Commit()
}

func (c *StableDemoComponent) addMessage(msg string) {
	timestamp := time.Now().Format("15:04:05")
	c.Messages = append(c.Messages, fmt.Sprintf("[%s] %s", timestamp, msg))
	
	// Keep only last 10 messages
	if len(c.Messages) > 10 {
		c.Messages = c.Messages[len(c.Messages)-10:]
	}
}

func (c *StableDemoComponent) GetTemplate() string {
	return `
<!DOCTYPE html>
<html>
<head>
	<title>Stable Framework Demo</title>
	<style>
		body {
			font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
			margin: 0;
			background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
			min-height: 100vh;
			padding: 2rem;
		}
		.container {
			max-width: 1200px;
			margin: 0 auto;
		}
		.header {
			text-align: center;
			color: white;
			margin-bottom: 2rem;
		}
		.header h1 {
			font-size: 3rem;
			margin-bottom: 0.5rem;
		}
		.grid {
			display: grid;
			grid-template-columns: 1fr 1fr;
			gap: 2rem;
		}
		.card {
			background: white;
			border-radius: 12px;
			padding: 2rem;
			box-shadow: 0 10px 30px rgba(0,0,0,0.2);
		}
		.card h2 {
			color: #667eea;
			margin-top: 0;
		}
		.counter {
			font-size: 4rem;
			font-weight: bold;
			color: #667eea;
			text-align: center;
			padding: 2rem;
		}
		.buttons {
			display: flex;
			gap: 1rem;
			flex-wrap: wrap;
			justify-content: center;
		}
		button {
			padding: 0.75rem 1.5rem;
			background: #667eea;
			color: white;
			border: none;
			border-radius: 8px;
			cursor: pointer;
			font-size: 1rem;
			transition: all 0.3s;
		}
		button:hover {
			background: #5a67d8;
			transform: translateY(-2px);
		}
		.error-btn {
			background: #f56565;
		}
		.error-btn:hover {
			background: #e53e3e;
		}
		.success-btn {
			background: #48bb78;
		}
		.success-btn:hover {
			background: #38a169;
		}
		.features {
			list-style: none;
			padding: 0;
		}
		.features li {
			padding: 0.5rem;
			margin: 0.5rem 0;
			background: #f7fafc;
			border-radius: 6px;
			border-left: 4px solid #667eea;
		}
		.messages {
			background: #2d3748;
			color: #48bb78;
			padding: 1rem;
			border-radius: 8px;
			font-family: monospace;
			max-height: 300px;
			overflow-y: auto;
		}
		.message {
			margin: 0.25rem 0;
		}
		.status {
			background: #edf2f7;
			padding: 1rem;
			border-radius: 8px;
			margin-top: 1rem;
		}
		.error {
			background: #fee;
			border: 2px solid #f88;
			padding: 1rem;
			border-radius: 8px;
			color: #c00;
		}
		@media (max-width: 768px) {
			.grid {
				grid-template-columns: 1fr;
			}
		}
	</style>
</head>
<body>
	<div class="container">
		<div class="header">
			<h1>🚀 Stable Framework Demo</h1>
			<p>All framework features working reliably</p>
		</div>
		
		<div class="grid">
			<div class="card">
				<h2>Interactive Demo</h2>
				
				<div class="counter">{{.Counter}}</div>
				
				<div class="buttons">
					<button onclick="send_event('{{.IdComponent}}', 'Increment', null)">
						➕ Increment
					</button>
					<button onclick="send_event('{{.IdComponent}}', 'Decrement', null)">
						➖ Decrement
					</button>
					<button onclick="send_event('{{.IdComponent}}', 'Reset', null)">
						🔄 Reset
					</button>
				</div>
				
				<div class="buttons" style="margin-top: 1rem;">
					<button onclick="send_event('{{.IdComponent}}', 'SaveState', null)" class="success-btn">
						💾 Save State
					</button>
					<button onclick="send_event('{{.IdComponent}}', 'LoadState', null)" class="success-btn">
						📂 Load State
					</button>
					<button onclick="send_event('{{.IdComponent}}', 'TestError', null)" class="error-btn">
						⚠️ Test Error
					</button>
				</div>
				
				<div class="status">
					<strong>Last Action:</strong> {{.LastAction}}
				</div>
			</div>
			
			<div class="card">
				<h2>Framework Features</h2>
				<ul class="features">
					{{range .Features}}
					<li>{{.}}</li>
					{{end}}
				</ul>
			</div>
		</div>
		
		<div class="card" style="margin-top: 2rem;">
			<h2>Event Log</h2>
			<div class="messages">
				{{range .Messages}}
				<div class="message">{{.}}</div>
				{{end}}
			</div>
		</div>
	</div>
</body>
</html>
	`
}

func (c *StableDemoComponent) GetDriver() liveview.LiveDriver {
	return c
}

// Event handlers
func (c *StableDemoComponent) Increment(data interface{}) {
	c.Counter++
	c.LastAction = fmt.Sprintf("Incremented to %d", c.Counter)
	c.addMessage(fmt.Sprintf("Counter incremented to %d", c.Counter))
	
	// Emit event
	c.eventRegistry.Emit("counter.changed", map[string]interface{}{
		"value": c.Counter,
		"action": "increment",
	})
	
	c.Commit()
}

func (c *StableDemoComponent) Decrement(data interface{}) {
	c.Counter--
	c.LastAction = fmt.Sprintf("Decremented to %d", c.Counter)
	c.addMessage(fmt.Sprintf("Counter decremented to %d", c.Counter))
	
	// Emit event
	c.eventRegistry.Emit("counter.changed", map[string]interface{}{
		"value": c.Counter,
		"action": "decrement",
	})
	
	c.Commit()
}

func (c *StableDemoComponent) Reset(data interface{}) {
	c.Counter = 0
	c.LastAction = "Counter reset"
	c.addMessage("Counter reset to 0")
	
	// Emit event
	c.eventRegistry.Emit("counter.reset", nil)
	
	c.Commit()
}

func (c *StableDemoComponent) SaveState(data interface{}) {
	// Save state using state manager
	c.stateManager.Set("counter", c.Counter)
	c.stateManager.Set("messages", c.Messages)
	c.stateManager.Set("saved_at", time.Now())
	
	c.LastAction = "State saved"
	c.addMessage(fmt.Sprintf("State saved (counter=%d)", c.Counter))
	c.Commit()
}

func (c *StableDemoComponent) LoadState(data interface{}) {
	// Load state from state manager
	if val, err := c.stateManager.Get("counter"); err == nil && val != nil {
		if counter, ok := val.(int); ok {
			c.Counter = counter
			c.LastAction = fmt.Sprintf("State loaded (counter=%d)", counter)
			c.addMessage(fmt.Sprintf("State loaded: counter=%d", counter))
		}
	} else {
		c.LastAction = "No saved state found"
		c.addMessage("No saved state to load")
	}
	
	c.Commit()
}

func (c *StableDemoComponent) TestError(data interface{}) {
	c.addMessage("Testing error boundary...")
	
	// Use error boundary to safely handle errors
	err := c.errorBoundary.SafeExecute("test_error", func() error {
		// Simulate an error
		return fmt.Errorf("this is a test error")
	})
	
	if err != nil {
		c.LastAction = "Error caught and handled"
		c.addMessage(fmt.Sprintf("Error boundary caught: %v", err))
	} else {
		c.LastAction = "No error occurred"
		c.addMessage("Operation completed without error")
	}
	
	c.Commit()
}

func main() {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	
	// Demo page
	page := &liveview.PageControl{
		Path:   "/",
		Title:  "Stable Framework Demo",
		Router: e,
	}
	
	page.Register(func() liveview.LiveDriver {
		return liveview.NewDriver("stable_demo", &StableDemoComponent{})
	})
	
	port := ":8080"
	fmt.Printf("Starting Stable Framework Demo\n")
	fmt.Printf("Open http://localhost%s\n", port)
	fmt.Println("\nFramework Features Active:")
	fmt.Println("  • Error Boundaries")
	fmt.Println("  • State Management")
	fmt.Println("  • Event Registry")
	fmt.Println("  • Lifecycle Hooks")
	fmt.Println("  • Template Cache")
	fmt.Println("  • SafeScript API")
	fmt.Println("  • Lazy Loading")
	
	if err := e.Start(port); err != nil {
		log.Fatal(err)
	}
}