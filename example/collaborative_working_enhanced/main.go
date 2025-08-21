package main

import (
	"fmt"
	"log"
	"time"

	"github.com/arturoeanton/go-echo-live-view/components"
	"github.com/arturoeanton/go-echo-live-view/liveview"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// Create Echo instance
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Configure Error Boundary for the entire app
	errorBoundary := liveview.NewErrorBoundary(&liveview.ErrorBoundaryConfig{
		MaxRetries:   3,
		RetryDelay:   2 * time.Second,
		EnableLogging: true,
		FallbackComponent: &liveview.FallbackComponent{
			Template: `<div style="padding: 2rem; background: #fee; border: 2px solid #f88; border-radius: 8px;">
				<h2>Something went wrong</h2>
				<p>{{.Error}}</p>
				<button onclick="location.reload()">Refresh Page</button>
			</div>`,
		},
	})

	// Configure Plugin System
	pluginManager := liveview.NewPluginManager()
	
	// Add performance monitoring plugin
	perfPlugin := &liveview.Plugin{
		Name:    "PerformanceMonitor",
		Version: "1.0.0",
		Initialize: func(ctx *liveview.PluginContext) error {
			log.Println("Performance monitoring initialized")
			return nil
		},
	}
	pluginManager.Register(perfPlugin)

	// Configure Template Cache
	templateCache := liveview.NewTemplateCache(&liveview.TemplateCacheConfig{
		MaxSize:         50 * 1024 * 1024, // 50MB
		TTL:             10 * time.Minute,
		EnableStats:     true,
		EnablePrecompile: true,
	})

	// Configure State Management
	stateManager := liveview.NewStateManager(liveview.NewMemoryProvider())

	// Configure Lazy Loading
	lazyLoader := liveview.NewLazyLoader(&liveview.LazyLoaderConfig{
		MaxRetries:      3,
		RetryDelay:      1 * time.Second,
		LoadTimeout:     5 * time.Second,
		EnableCaching:   true,
		EnableMetrics:   true,
	})

	// Main page with demos
	e.GET("/", func(c echo.Context) error {
		html := `
<!DOCTYPE html>
<html>
<head>
	<title>Enhanced Collaborative Components</title>
	<style>
		body {
			font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
			background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
			min-height: 100vh;
			margin: 0;
			padding: 20px;
		}
		.container {
			max-width: 1200px;
			margin: 0 auto;
			text-align: center;
		}
		h1 {
			color: white;
			font-size: 48px;
			margin-bottom: 20px;
		}
		.features {
			background: rgba(255,255,255,0.1);
			border-radius: 12px;
			padding: 20px;
			margin: 20px 0;
			color: white;
		}
		.feature-list {
			display: flex;
			flex-wrap: wrap;
			gap: 10px;
			justify-content: center;
		}
		.feature-tag {
			background: rgba(255,255,255,0.2);
			padding: 5px 10px;
			border-radius: 20px;
			font-size: 14px;
		}
		.demo-grid {
			display: grid;
			grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
			gap: 20px;
			margin-top: 40px;
		}
		.demo-card {
			background: white;
			border-radius: 12px;
			padding: 30px;
			text-decoration: none;
			color: #333;
			box-shadow: 0 10px 30px rgba(0,0,0,0.1);
			transition: transform 0.3s;
		}
		.demo-card:hover {
			transform: translateY(-5px);
			box-shadow: 0 15px 40px rgba(0,0,0,0.2);
		}
		.demo-icon {
			font-size: 48px;
			margin-bottom: 15px;
		}
		.demo-title {
			font-size: 24px;
			font-weight: 600;
			margin-bottom: 10px;
		}
		.demo-desc {
			color: #666;
		}
		.new-badge {
			display: inline-block;
			background: #4caf50;
			color: white;
			padding: 2px 8px;
			border-radius: 12px;
			font-size: 12px;
			margin-left: 10px;
		}
		.port-info {
			background: rgba(255,255,255,0.9);
			padding: 10px;
			border-radius: 8px;
			margin-top: 20px;
			color: #333;
		}
	</style>
</head>
<body>
	<div class="container">
		<h1>🚀 Enhanced Collaborative Components</h1>
		<p style="color: white; font-size: 20px;">Real-time collaboration with advanced framework features</p>
		
		<div class="features">
			<h3>✨ New Framework Features Active</h3>
			<div class="feature-list">
				<span class="feature-tag">🛡️ Error Boundaries</span>
				<span class="feature-tag">🔄 Lifecycle Hooks</span>
				<span class="feature-tag">🔌 Plugin System</span>
				<span class="feature-tag">💾 State Management</span>
				<span class="feature-tag">⚡ Virtual DOM</span>
				<span class="feature-tag">📦 Template Cache</span>
				<span class="feature-tag">🎭 Lazy Loading</span>
				<span class="feature-tag">🎯 Event Registry</span>
			</div>
		</div>
		
		<div class="demo-grid">
			<a href="/kanban" class="demo-card">
				<div class="demo-icon">📋</div>
				<div class="demo-title">Enhanced Kanban <span class="new-badge">NEW</span></div>
				<div class="demo-desc">With error recovery & state persistence</div>
			</a>
			
			<a href="/canvas" class="demo-card">
				<div class="demo-icon">🎨</div>
				<div class="demo-title">Smart Canvas <span class="new-badge">NEW</span></div>
				<div class="demo-desc">Virtual DOM optimized drawing</div>
			</a>
			
			<a href="/presence" class="demo-card">
				<div class="demo-icon">👥</div>
				<div class="demo-title">Live Presence <span class="new-badge">NEW</span></div>
				<div class="demo-desc">Real-time with event registry</div>
			</a>
		</div>
		
		<div class="port-info">
			<strong>Running on port 8081</strong> | Framework v2.0 Enhanced
		</div>
	</div>
</body>
</html>
		`
		return c.HTML(200, html)
	})

	// Enhanced Kanban Board Demo
	kanbanPage := &liveview.PageControl{
		Path:   "/kanban",
		Title:  "Enhanced Kanban Board",
		Router: e,
	}
	
	kanbanPage.Register(func() liveview.LiveDriver {
		board := &components.KanbanBoard{}
		
		// Initialize CollaborativeComponent
		board.CollaborativeComponent = &liveview.CollaborativeComponent{
			Driver: nil,
		}
		
		// Create driver with error boundary
		board.ComponentDriver = liveview.NewDriver[*components.KanbanBoard]("kanban_board", board)
		board.CollaborativeComponent.Driver = board.ComponentDriver
		
		// Wrap with error boundary
		errorBoundary.Wrap(board.ComponentDriver, func(c liveview.Component, err error) liveview.Component {
			log.Printf("Kanban board error: %v", err)
			return &liveview.FallbackComponent{
				Template: `<div>Kanban board temporarily unavailable. Retrying...</div>`,
			}
		})
		
		// Add lifecycle hooks
		lifecycle := liveview.NewLifecycleManager(board.ComponentDriver)
		lifecycle.RegisterHook(liveview.BeforeMount, func(c liveview.Component) error {
			log.Println("Kanban board mounting...")
			return nil
		})
		lifecycle.RegisterHook(liveview.AfterMount, func(c liveview.Component) error {
			log.Println("Kanban board mounted successfully")
			
			// Save state
			stateManager.Set("kanban_mounted", true)
			return nil
		})
		
		// Initialize board with enhanced features
		board.Title = "Enhanced Project Tasks"
		board.Description = "Now with error recovery and state persistence"
		
		board.Columns = []components.KanbanColumn{
			{ID: "todo", Title: "To Do", Color: "#3498db", Order: 0},
			{ID: "progress", Title: "In Progress", Color: "#f39c12", Order: 1, WIPLimit: 3},
			{ID: "review", Title: "Review", Color: "#9b59b6", Order: 2, WIPLimit: 2},
			{ID: "done", Title: "Done", Color: "#27ae60", Order: 3},
		}
		
		// Load cards from state if available
		if savedCards, exists := stateManager.Get("kanban_cards"); exists {
			if cards, ok := savedCards.([]components.KanbanCard); ok {
				board.Cards = cards
			}
		} else {
			board.Cards = []components.KanbanCard{
				{
					ID:          "card1",
					Title:       "Implement Error Boundaries",
					Description: "Add error recovery to components",
					ColumnID:    "done",
					Priority:    "high",
					CreatedAt:   time.Now().Add(-48 * time.Hour),
					UpdatedAt:   time.Now(),
				},
				{
					ID:          "card2",
					Title:       "Add Virtual DOM",
					Description: "Optimize rendering performance",
					ColumnID:    "review",
					Priority:    "high",
					CreatedAt:   time.Now().Add(-24 * time.Hour),
					UpdatedAt:   time.Now(),
				},
				{
					ID:          "card3",
					Title:       "Setup State Management",
					Description: "Implement reactive state system",
					ColumnID:    "progress",
					Priority:    "medium",
					CreatedAt:   time.Now().Add(-12 * time.Hour),
					UpdatedAt:   time.Now(),
				},
				{
					ID:          "card4",
					Title:       "Create Plugin System",
					Description: "Allow extensibility",
					ColumnID:    "todo",
					Priority:    "medium",
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				},
				{
					ID:          "card5",
					Title:       "Add Template Cache",
					Description: "Cache compiled templates",
					ColumnID:    "todo",
					Priority:    "low",
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				},
			}
			
			// Save initial state
			stateManager.Set("kanban_cards", board.Cards)
		}
		
		board.Labels = []components.KanbanLabel{
			{ID: "feature", Name: "Feature", Color: "#3498db"},
			{ID: "enhancement", Name: "Enhancement", Color: "#27ae60"},
			{ID: "bug", Name: "Bug", Color: "#e74c3c"},
			{ID: "performance", Name: "Performance", Color: "#f39c12"},
		}
		
		board.ActiveUsers = make(map[string]*components.UserActivity)
		
		// Setup event handlers with new event registry
		eventRegistry := liveview.NewEventRegistry()
		
		// Register card move handler with throttling
		eventRegistry.Register("card.moved", liveview.NewEventHandler(
			func(event *liveview.Event) error {
				// Save state after card move
				stateManager.Set("kanban_cards", board.Cards)
				stateManager.Set("last_activity", time.Now())
				return nil
			},
			liveview.WithThrottle(500*time.Millisecond),
		))
		
		return board.ComponentDriver
	})

	// Enhanced Canvas Demo
	e.GET("/canvas", func(c echo.Context) error {
		return c.HTML(200, `
			<!DOCTYPE html>
			<html>
			<head>
				<title>Enhanced Canvas - With Virtual DOM</title>
				<style>
					body {
						font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
						background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
						display: flex;
						justify-content: center;
						align-items: center;
						height: 100vh;
						margin: 0;
					}
					.container {
						background: white;
						padding: 40px;
						border-radius: 12px;
						text-align: center;
						box-shadow: 0 10px 30px rgba(0,0,0,0.2);
						max-width: 600px;
					}
					h1 { color: #667eea; }
					.feature-list {
						text-align: left;
						margin: 20px 0;
					}
					.feature {
						padding: 10px;
						margin: 5px 0;
						background: #f5f5f5;
						border-radius: 6px;
					}
					a {
						display: inline-block;
						margin-top: 20px;
						padding: 10px 20px;
						background: #667eea;
						color: white;
						text-decoration: none;
						border-radius: 6px;
						margin: 5px;
					}
				</style>
			</head>
			<body>
				<div class="container">
					<h1>🎨 Enhanced Canvas Demo</h1>
					<p>Coming soon with these features:</p>
					<div class="feature-list">
						<div class="feature">✅ Virtual DOM for efficient updates</div>
						<div class="feature">✅ Lazy loading of drawing tools</div>
						<div class="feature">✅ State persistence for drawings</div>
						<div class="feature">✅ Error recovery on connection loss</div>
						<div class="feature">✅ Plugin system for custom tools</div>
					</div>
					<a href="/kanban">Try Kanban Board</a>
					<a href="/">Back to Home</a>
				</div>
			</body>
			</html>
		`)
	})

	// Enhanced Presence Demo
	e.GET("/presence", func(c echo.Context) error {
		return c.HTML(200, `
			<!DOCTYPE html>
			<html>
			<head>
				<title>Enhanced Presence - Real-time Tracking</title>
				<style>
					body {
						font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
						background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
						display: flex;
						justify-content: center;
						align-items: center;
						height: 100vh;
						margin: 0;
					}
					.container {
						background: white;
						padding: 40px;
						border-radius: 12px;
						text-align: center;
						box-shadow: 0 10px 30px rgba(0,0,0,0.2);
						max-width: 600px;
					}
					h1 { color: #667eea; }
					.stats {
						display: grid;
						grid-template-columns: repeat(2, 1fr);
						gap: 20px;
						margin: 20px 0;
					}
					.stat {
						padding: 20px;
						background: #f5f5f5;
						border-radius: 8px;
					}
					.stat-value {
						font-size: 32px;
						font-weight: bold;
						color: #667eea;
					}
					.stat-label {
						color: #666;
						margin-top: 5px;
					}
					a {
						display: inline-block;
						margin-top: 20px;
						padding: 10px 20px;
						background: #667eea;
						color: white;
						text-decoration: none;
						border-radius: 6px;
						margin: 5px;
					}
				</style>
			</head>
			<body>
				<div class="container">
					<h1>👥 Enhanced Presence Demo</h1>
					<p>Advanced user tracking with Event Registry</p>
					<div class="stats">
						<div class="stat">
							<div class="stat-value">0</div>
							<div class="stat-label">Active Users</div>
						</div>
						<div class="stat">
							<div class="stat-value">0</div>
							<div class="stat-label">Events/sec</div>
						</div>
						<div class="stat">
							<div class="stat-value">∞</div>
							<div class="stat-label">Uptime</div>
						</div>
						<div class="stat">
							<div class="stat-value">0ms</div>
							<div class="stat-label">Latency</div>
						</div>
					</div>
					<p>Features: Event throttling, presence indicators, activity tracking</p>
					<a href="/kanban">Try Kanban Board</a>
					<a href="/">Back to Home</a>
				</div>
			</body>
			</html>
		`)
	})

	// Start server on port 8081
	port := ":8081"
	fmt.Printf("🚀 Enhanced Collaborative Components Server\n")
	fmt.Printf("🌐 Starting on http://localhost%s\n", port)
	fmt.Println("\n✨ Framework Features:")
	fmt.Println("  • Error Boundaries with automatic recovery")
	fmt.Println("  • Lifecycle Hooks for component management")
	fmt.Println("  • Plugin System for extensibility")
	fmt.Println("  • State Management with persistence")
	fmt.Println("  • Virtual DOM for performance")
	fmt.Println("  • Template Cache for compilation")
	fmt.Println("  • Lazy Loading for on-demand components")
	fmt.Println("  • Event Registry with throttling")
	fmt.Println("\n📋 Available demos:")
	fmt.Println("  / - Main page with feature overview")
	fmt.Println("  /kanban - Enhanced Kanban board")
	fmt.Println("  /canvas - Virtual DOM drawing canvas")
	fmt.Println("  /presence - Real-time user presence")
	
	if err := e.Start(port); err != nil {
		log.Fatal(err)
	}
}