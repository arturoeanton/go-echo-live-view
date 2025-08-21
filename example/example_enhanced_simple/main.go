package main

import (
	"fmt"
	"log"
	"time"

	"github.com/arturoeanton/go-echo-live-view/liveview"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// SimpleEnhancedComponent demonstrates the new framework features
type SimpleEnhancedComponent struct {
	*liveview.ComponentDriver[*SimpleEnhancedComponent]
	
	// Stats
	Counter         int
	Errors          int
	LastAction      string
	SafeScriptUsed  bool
	
	// Framework features
	ErrorBoundary   *liveview.ErrorBoundary
	EventRegistry   *liveview.EventRegistry
	StateManager    *liveview.StateManager
	TemplateCache   *liveview.TemplateCache
	Lifecycle       *liveview.LifecycleManager
}

func (c *SimpleEnhancedComponent) Start() {
	// Initialize Error Boundary
	c.ErrorBoundary = liveview.NewErrorBoundary(100, true)
	c.ErrorBoundary.SetFallbackRenderer(func(componentID string, err error) string {
		return fmt.Sprintf(`<div style="padding: 1rem; background: #fee; border: 2px solid red;">Error in %s: %v</div>`, componentID, err)
	})
	
	// Initialize Event Registry
	c.EventRegistry = liveview.NewEventRegistry(nil)
	
	// Initialize State Manager  
	c.StateManager = liveview.NewStateManager(liveview.NewInMemoryProvider())
	
	// Initialize Template Cache
	c.TemplateCache = liveview.NewTemplateCache(nil)
	
	// Initialize Lifecycle Manager
	c.Lifecycle = liveview.NewLifecycleManager("demo")
	c.Lifecycle.TransitionTo(liveview.LifecycleStateCreated)
	
	// Register event handlers
	c.EventRegistry.RegisterHandler("test.event", &liveview.EventHandler{
		Name: "test",
		Handler: func(event *liveview.Event) error {
			log.Printf("Event received: %s", event.Name)
			return nil
		},
	})
	
	// Add lifecycle hooks
	c.Lifecycle.AddHook(liveview.LifecycleStateCreated, func(data interface{}) error {
		log.Println("Component created")
		return nil
	})
	
	c.Lifecycle.AddHook(liveview.LifecycleStageMounted, func(data interface{}) error {
		log.Println("Component mounted")
		c.StateManager.Set("mounted", true)
		return nil
	})
	
	// Transition to mounted
	c.Lifecycle.TransitionTo(liveview.LifecycleStageMounted)
	
	c.Counter = 0
	c.Errors = 0
	c.LastAction = "Component initialized with enhanced features"
	c.Commit()
}

func (c *SimpleEnhancedComponent) GetTemplate() string {
	return `
<!DOCTYPE html>
<html>
<head>
	<title>Enhanced Framework Demo</title>
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
		.features {
			display: grid;
			grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
			gap: 1rem;
			margin: 2rem 0;
		}
		.feature {
			padding: 1rem;
			background: #f7fafc;
			border-radius: 8px;
			border-left: 4px solid #667eea;
		}
		.feature h3 {
			margin-top: 0;
			color: #667eea;
		}
		.stats {
			display: flex;
			gap: 2rem;
			margin: 2rem 0;
		}
		.stat {
			text-align: center;
		}
		.stat-value {
			font-size: 2rem;
			font-weight: bold;
			color: #667eea;
		}
		.stat-label {
			color: #666;
			font-size: 0.875rem;
		}
		.buttons {
			display: flex;
			gap: 1rem;
			flex-wrap: wrap;
		}
		button {
			padding: 0.75rem 1.5rem;
			background: #667eea;
			color: white;
			border: none;
			border-radius: 6px;
			cursor: pointer;
			font-size: 1rem;
			transition: all 0.2s;
		}
		button:hover {
			background: #5a67d8;
			transform: translateY(-2px);
		}
		.error-button {
			background: #f56565;
		}
		.error-button:hover {
			background: #e53e3e;
		}
		.success-button {
			background: #48bb78;
		}
		.success-button:hover {
			background: #38a169;
		}
		.status {
			padding: 1rem;
			background: #edf2f7;
			border-radius: 6px;
			margin-top: 2rem;
		}
	</style>
</head>
<body>
	<div class="container">
		<div class="card">
			<h1>🚀 Enhanced Framework Features Demo</h1>
			<p>This example demonstrates the new framework capabilities</p>
			
			<div class="features">
				<div class="feature">
					<h3>🛡️ Error Boundaries</h3>
					<p>Automatic error recovery</p>
				</div>
				<div class="feature">
					<h3>🔄 Lifecycle Hooks</h3>
					<p>Component lifecycle management</p>
				</div>
				<div class="feature">
					<h3>💾 State Management</h3>
					<p>Persistent state storage</p>
				</div>
				<div class="feature">
					<h3>📦 Template Cache</h3>
					<p>Compiled template caching</p>
				</div>
				<div class="feature">
					<h3>🎯 Event Registry</h3>
					<p>Advanced event handling</p>
				</div>
				<div class="feature">
					<h3>🔒 SafeScript API</h3>
					<p>Secure JavaScript execution</p>
				</div>
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
				<div class="stat">
					<div class="stat-value">{{if .SafeScriptUsed}}✅{{else}}❌{{end}}</div>
					<div class="stat-label">SafeScript</div>
				</div>
			</div>
			
			<div class="buttons">
				<button onclick="send_event('{{.IdComponent}}', 'Increment', null)">
					Increment Counter
				</button>
				<button onclick="send_event('{{.IdComponent}}', 'SaveState', null)" class="success-button">
					Save State
				</button>
				<button onclick="send_event('{{.IdComponent}}', 'LoadState', null)" class="success-button">
					Load State
				</button>
				<button onclick="send_event('{{.IdComponent}}', 'TriggerError', null)" class="error-button">
					Trigger Error (Handled)
				</button>
				<button onclick="send_event('{{.IdComponent}}', 'UseSafeScript', null)">
					Use SafeScript
				</button>
				<button onclick="send_event('{{.IdComponent}}', 'EmitEvent', null)">
					Emit Event
				</button>
			</div>
			
			<div class="status">
				<strong>Last Action:</strong> {{.LastAction}}
			</div>
		</div>
		
		<div class="card">
			<h2>Try These Actions:</h2>
			<ol>
				<li>Click "Increment Counter" to see basic functionality</li>
				<li>Click "Save State" then refresh and "Load State" to see persistence</li>
				<li>Click "Trigger Error" to see error boundary in action</li>
				<li>Click "Use SafeScript" to see safe JavaScript execution</li>
				<li>Click "Emit Event" to see event registry working</li>
			</ol>
		</div>
	</div>
</body>
</html>
	`
}

func (c *SimpleEnhancedComponent) GetDriver() liveview.LiveDriver {
	return c
}

func (c *SimpleEnhancedComponent) Increment(data interface{}) {
	c.Counter++
	c.LastAction = fmt.Sprintf("Counter incremented to %d", c.Counter)
	
	// Save to state manager
	c.StateManager.Set("counter", c.Counter)
	
	// Emit event
	c.EventRegistry.Emit(&liveview.Event{
		Name: "counter.incremented",
		Data: c.Counter,
	})
	
	c.Commit()
}

func (c *SimpleEnhancedComponent) SaveState(data interface{}) {
	// Save all state
	c.StateManager.Set("counter", c.Counter)
	c.StateManager.Set("errors", c.Errors)
	c.StateManager.Set("last_save", time.Now())
	
	c.LastAction = "State saved successfully"
	c.Commit()
}

func (c *SimpleEnhancedComponent) LoadState(data interface{}) {
	// Load state
	if val, exists := c.StateManager.Get("counter"); exists {
		if counter, ok := val.(int); ok {
			c.Counter = counter
		}
	}
	
	if val, exists := c.StateManager.Get("errors"); exists {
		if errors, ok := val.(int); ok {
			c.Errors = errors
		}
	}
	
	c.LastAction = "State loaded successfully"
	c.Commit()
}

func (c *SimpleEnhancedComponent) TriggerError(data interface{}) {
	// Simulate an error that gets caught by error boundary
	defer func() {
		if r := recover(); r != nil {
			c.ErrorBoundary.CatchError("demo", fmt.Errorf("panic: %v", r))
			c.Errors++
			c.LastAction = "Error caught and handled by Error Boundary"
			c.Commit()
		}
	}()
	
	// This would normally panic
	panic("simulated error")
}

func (c *SimpleEnhancedComponent) UseSafeScript(data interface{}) {
	// Use SafeScript API
	scriptCode := "console.log('Hello from SafeScript')"
	safeScript, err := liveview.NewSafeScript(scriptCode, nil)
	if err != nil {
		c.LastAction = fmt.Sprintf("SafeScript blocked dangerous code: %v", err)
	} else {
		// Script is safe to use
		c.LastAction = "SafeScript validated successfully"
		c.SafeScriptUsed = true
	}
	
	c.Commit()
}

func (c *SimpleEnhancedComponent) EmitEvent(data interface{}) {
	// Emit a custom event
	event := &liveview.Event{
		Name:      "demo.custom_event",
		Data:      map[string]interface{}{"counter": c.Counter, "timestamp": time.Now()},
		Timestamp: time.Now().Unix(),
	}
	
	c.EventRegistry.Emit(event)
	c.LastAction = fmt.Sprintf("Event emitted: %s", event.Name)
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
	<title>Enhanced Framework Examples</title>
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
			max-width: 600px;
		}
		h1 {
			color: #333;
			margin-bottom: 2rem;
		}
		.links {
			display: flex;
			flex-direction: column;
			gap: 1rem;
		}
		a {
			display: block;
			padding: 1rem 2rem;
			background: #667eea;
			color: white;
			text-decoration: none;
			border-radius: 8px;
			transition: all 0.2s;
		}
		a:hover {
			background: #5a67d8;
			transform: translateY(-2px);
		}
		.port {
			color: #666;
			font-size: 0.875rem;
			margin-top: 0.5rem;
		}
	</style>
</head>
<body>
	<div class="container">
		<h1>🚀 Enhanced Framework Examples</h1>
		<p>Choose an example to explore the new features:</p>
		<div class="links">
			<a href="/demo">
				Simple Enhanced Demo
				<div class="port">Interactive demonstration of all framework features</div>
			</a>
		</div>
	</div>
</body>
</html>
		`
		return c.HTML(200, html)
	})

	// Demo page
	demoPage := &liveview.PageControl{
		Path:   "/demo",
		Title:  "Enhanced Framework Demo",
		Router: e,
	}
	
	demoPage.Register(func() liveview.LiveDriver {
		return liveview.NewDriver("enhanced_demo", &SimpleEnhancedComponent{})
	})

	port := ":8080"
	fmt.Printf("🚀 Enhanced Framework Examples Server\n")
	fmt.Printf("🌐 Starting on http://localhost%s\n", port)
	fmt.Println("\n✨ Framework Features Active:")
	fmt.Println("  • Error Boundaries")
	fmt.Println("  • Lifecycle Hooks")
	fmt.Println("  • State Management")
	fmt.Println("  • Template Cache")
	fmt.Println("  • Event Registry")
	fmt.Println("  • SafeScript API")
	fmt.Println("\n📋 Available demos:")
	fmt.Println("  / - Main page")
	fmt.Println("  /demo - Enhanced features demo")
	
	if err := e.Start(port); err != nil {
		log.Fatal(err)
	}
}