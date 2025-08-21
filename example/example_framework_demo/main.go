package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/arturoeanton/go-echo-live-view/liveview"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// DemoComponent demonstrates the new framework features
type DemoComponent struct {
	*liveview.ComponentDriver[*DemoComponent]
	
	// Component state
	Counter         int
	LastAction      string
	Errors          int
	Features        []string
	
	// Framework components
	errorBoundary   *liveview.ErrorBoundary
	stateManager    *liveview.StateManager
	lifecycle       *liveview.LifecycleManager
	eventRegistry   *liveview.EventRegistry
}

func (c *DemoComponent) Start() {
	// Initialize Error Boundary
	c.errorBoundary = liveview.NewErrorBoundary(10, true)
	c.errorBoundary.SetFallbackRenderer(func(componentID string, err error) string {
		return fmt.Sprintf(`<div class="error">Error: %v</div>`, err)
	})
	
	// Initialize State Manager
	c.stateManager = liveview.NewStateManager(&liveview.StateConfig{
		Provider:     liveview.NewMemoryStateProvider(),
		CacheEnabled: true,
		CacheTTL:     5 * time.Minute,
	})
	
	// Initialize Lifecycle Manager
	c.lifecycle = liveview.NewLifecycleManager("demo")
	c.lifecycle.SetHooks(&liveview.LifecycleHooks{
		OnCreated: func() error {
			log.Println("Component created")
			return nil
		},
		OnMounted: func() error {
			log.Println("Component mounted")
			return c.stateManager.Set("mounted_at", time.Now())
		},
	})
	
	// Initialize Event Registry
	c.eventRegistry = liveview.NewEventRegistry(nil)
	
	// Register a test event handler
	c.eventRegistry.On("counter.changed", func(ctx context.Context, event *liveview.Event) error {
		log.Printf("Counter changed event: %v", event.Data)
		return nil
	})
	
	// Initialize component state
	c.Counter = 0
	c.LastAction = "Component initialized"
	c.Errors = 0
	c.Features = []string{
		"Error Boundaries",
		"State Management", 
		"Lifecycle Hooks",
		"Event Registry",
		"SafeScript API",
		"Template Cache",
		"Virtual DOM",
		"Lazy Loading",
	}
	
	// Execute lifecycle
	c.lifecycle.Create()
	c.lifecycle.Mount()
	
	c.Commit()
}

func (c *DemoComponent) GetTemplate() string {
	return `
<!DOCTYPE html>
<html>
<head>
	<title>Framework Features Demo</title>
	<style>
		body {
			font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
			background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
			min-height: 100vh;
			padding: 2rem;
			margin: 0;
		}
		.container {
			max-width: 800px;
			margin: 0 auto;
		}
		.card {
			background: white;
			border-radius: 12px;
			padding: 2rem;
			margin-bottom: 2rem;
			box-shadow: 0 10px 30px rgba(0,0,0,0.1);
		}
		h1 {
			color: #333;
			margin-top: 0;
		}
		.features-grid {
			display: grid;
			grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
			gap: 1rem;
			margin: 2rem 0;
		}
		.feature-card {
			padding: 1rem;
			background: #f8f9fa;
			border-radius: 8px;
			border-left: 4px solid #667eea;
		}
		.stats {
			display: flex;
			gap: 2rem;
			margin: 2rem 0;
		}
		.stat {
			text-align: center;
			flex: 1;
		}
		.stat-value {
			font-size: 2rem;
			font-weight: bold;
			color: #667eea;
		}
		.stat-label {
			color: #666;
			font-size: 0.875rem;
			margin-top: 0.5rem;
		}
		.buttons {
			display: flex;
			gap: 1rem;
			flex-wrap: wrap;
			margin: 2rem 0;
		}
		button {
			padding: 0.75rem 1.5rem;
			background: #667eea;
			color: white;
			border: none;
			border-radius: 6px;
			cursor: pointer;
			font-size: 1rem;
		}
		button:hover {
			background: #5a67d8;
		}
		.error-button {
			background: #f56565;
		}
		.error-button:hover {
			background: #e53e3e;
		}
		.status {
			padding: 1rem;
			background: #edf2f7;
			border-radius: 6px;
			margin-top: 1rem;
		}
		.error {
			padding: 1rem;
			background: #fee;
			border: 2px solid #f88;
			border-radius: 8px;
			color: #c00;
		}
	</style>
</head>
<body>
	<div class="container">
		<div class="card">
			<h1>Framework Features Demo</h1>
			<p>Demonstrating all new framework capabilities</p>
			
			<div class="features-grid">
				{{range .Features}}
				<div class="feature-card">{{.}}</div>
				{{end}}
			</div>
			
			<div class="stats">
				<div class="stat">
					<div class="stat-value">{{.Counter}}</div>
					<div class="stat-label">Counter</div>
				</div>
				<div class="stat">
					<div class="stat-value">{{.Errors}}</div>
					<div class="stat-label">Errors Caught</div>
				</div>
			</div>
			
			<div class="buttons">
				<button onclick="send_event('{{.IdComponent}}', 'Increment', null)">
					Increment Counter
				</button>
				<button onclick="send_event('{{.IdComponent}}', 'SaveState', null)">
					Save State
				</button>
				<button onclick="send_event('{{.IdComponent}}', 'LoadState', null)">
					Load State
				</button>
				<button onclick="send_event('{{.IdComponent}}', 'TestError', null)" class="error-button">
					Test Error Boundary
				</button>
			</div>
			
			<div class="status">
				<strong>Last Action:</strong> {{.LastAction}}
			</div>
		</div>
	</div>
</body>
</html>
	`
}

func (c *DemoComponent) GetDriver() liveview.LiveDriver {
	return c
}

// Event handlers
func (c *DemoComponent) Increment(data interface{}) {
	c.Counter++
	c.LastAction = fmt.Sprintf("Counter incremented to %d", c.Counter)
	
	// Save to state
	c.stateManager.Set("counter", c.Counter)
	
	// Emit event
	c.eventRegistry.Emit("counter.changed", map[string]interface{}{"value": c.Counter})
	
	c.Commit()
}

func (c *DemoComponent) SaveState(data interface{}) {
	// Save all state
	c.stateManager.Set("counter", c.Counter)
	c.stateManager.Set("errors", c.Errors)
	c.stateManager.Set("saved_at", time.Now())
	
	c.LastAction = "State saved successfully"
	c.Commit()
}

func (c *DemoComponent) LoadState(data interface{}) {
	// Load state
	if val, err := c.stateManager.Get("counter"); err == nil && val != nil {
		if counter, ok := val.(int); ok {
			c.Counter = counter
		}
	}
	
	if val, err := c.stateManager.Get("errors"); err == nil && val != nil {
		if errors, ok := val.(int); ok {
			c.Errors = errors
		}
	}
	
	c.LastAction = "State loaded successfully"
	c.Commit()
}

func (c *DemoComponent) TestError(data interface{}) {
	// Test error boundary
	err := c.errorBoundary.SafeExecute("demo", func() error {
		// Simulate an error
		return fmt.Errorf("simulated error for testing")
	})
	
	if err != nil {
		c.Errors++
		c.LastAction = "Error caught and handled"
	}
	
	c.Commit()
}

func main() {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	
	// Main page
	e.GET("/", func(c echo.Context) error {
		html := `
<!DOCTYPE html>
<html>
<head>
	<title>Framework Demo</title>
	<style>
		body {
			font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
			background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
			min-height: 100vh;
			display: flex;
			align-items: center;
			justify-content: center;
			margin: 0;
		}
		.container {
			background: white;
			padding: 3rem;
			border-radius: 12px;
			box-shadow: 0 20px 60px rgba(0,0,0,0.3);
			text-align: center;
		}
		h1 {
			color: #667eea;
			margin-bottom: 2rem;
		}
		a {
			display: inline-block;
			padding: 1rem 2rem;
			background: #667eea;
			color: white;
			text-decoration: none;
			border-radius: 8px;
			margin: 0.5rem;
		}
		a:hover {
			background: #5a67d8;
		}
	</style>
</head>
<body>
	<div class="container">
		<h1>Framework Features Demo</h1>
		<p>Explore the new framework capabilities</p>
		<a href="/demo">Launch Demo</a>
	</div>
</body>
</html>
		`
		return c.HTML(200, html)
	})
	
	// Demo page
	demoPage := &liveview.PageControl{
		Path:   "/demo",
		Title:  "Framework Demo",
		Router: e,
	}
	
	demoPage.Register(func() liveview.LiveDriver {
		return liveview.NewDriver("framework_demo", &DemoComponent{})
	})
	
	port := ":8080"
	fmt.Printf("Starting Framework Demo Server\n")
	fmt.Printf("Open http://localhost%s\n", port)
	
	if err := e.Start(port); err != nil {
		log.Fatal(err)
	}
}