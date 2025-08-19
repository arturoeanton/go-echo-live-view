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

	// Main page with demos
	e.GET("/", func(c echo.Context) error {
		html := `
<!DOCTYPE html>
<html>
<head>
	<title>Collaborative Components Demo</title>
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
	</style>
</head>
<body>
	<div class="container">
		<h1>🚀 Collaborative Components</h1>
		<p style="color: white; font-size: 20px;">Real-time collaboration demos</p>
		
		<div class="demo-grid">
			<a href="/kanban" class="demo-card">
				<div class="demo-icon">📋</div>
				<div class="demo-title">Kanban Board</div>
				<div class="demo-desc">Drag and drop task management</div>
			</a>
			
			<a href="/canvas" class="demo-card">
				<div class="demo-icon">🎨</div>
				<div class="demo-title">Drawing Canvas</div>
				<div class="demo-desc">Collaborative drawing board</div>
			</a>
			
			<a href="/presence" class="demo-card">
				<div class="demo-icon">👥</div>
				<div class="demo-title">User Presence</div>
				<div class="demo-desc">Real-time user tracking</div>
			</a>
		</div>
	</div>
</body>
</html>
		`
		return c.HTML(200, html)
	})

	// Kanban Board Demo
	kanbanPage := &liveview.PageControl{
		Path:   "/kanban",
		Title:  "Kanban Board Demo",
		Router: e,
	}
	
	kanbanPage.Register(func() liveview.LiveDriver {
		board := &components.KanbanBoard{}
		
		// Initialize CollaborativeComponent
		board.CollaborativeComponent = &liveview.CollaborativeComponent{
			Driver: nil,
		}
		
		// Create driver
		board.ComponentDriver = liveview.NewDriver[*components.KanbanBoard]("kanban_board", board)
		board.CollaborativeComponent.Driver = board.ComponentDriver
		
		// Initialize board
		board.Title = "Project Tasks"
		board.Description = "Drag and drop to organize tasks"
		
		board.Columns = []components.KanbanColumn{
			{ID: "todo", Title: "To Do", Color: "#3498db", Order: 0},
			{ID: "progress", Title: "In Progress", Color: "#f39c12", Order: 1, WIPLimit: 3},
			{ID: "done", Title: "Done", Color: "#27ae60", Order: 2},
		}
		
		board.Cards = []components.KanbanCard{
			{
				ID:          "card1",
				Title:       "Setup project",
				Description: "Initialize the repository",
				ColumnID:    "done",
				Priority:    "high",
				CreatedAt:   time.Now().Add(-48 * time.Hour),
				UpdatedAt:   time.Now(),
			},
			{
				ID:          "card2",
				Title:       "Design UI",
				Description: "Create mockups",
				ColumnID:    "progress",
				Priority:    "medium",
				CreatedAt:   time.Now().Add(-24 * time.Hour),
				UpdatedAt:   time.Now(),
			},
			{
				ID:          "card3",
				Title:       "Write tests",
				Description: "Add unit tests",
				ColumnID:    "todo",
				Priority:    "low",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
		}
		
		board.Labels = []components.KanbanLabel{
			{ID: "feature", Name: "Feature", Color: "#3498db"},
			{ID: "bug", Name: "Bug", Color: "#e74c3c"},
		}
		
		board.ActiveUsers = make(map[string]*components.UserActivity)
		
		return board.ComponentDriver
	})

	// Canvas Demo - DISABLED for now due to WebSocket issues
	// The canvas component needs proper WebSocket endpoint configuration
	e.GET("/canvas", func(c echo.Context) error {
		return c.HTML(200, `
			<!DOCTYPE html>
			<html>
			<head>
				<title>Canvas Demo - Under Construction</title>
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
					.message {
						background: white;
						padding: 40px;
						border-radius: 12px;
						text-align: center;
						box-shadow: 0 10px 30px rgba(0,0,0,0.2);
					}
					h1 { color: #667eea; }
					a {
						display: inline-block;
						margin-top: 20px;
						padding: 10px 20px;
						background: #667eea;
						color: white;
						text-decoration: none;
						border-radius: 6px;
					}
				</style>
			</head>
			<body>
				<div class="message">
					<h1>🎨 Canvas Demo</h1>
					<p>This component is under construction.</p>
					<p>Please try the Kanban Board instead!</p>
					<a href="/kanban">Go to Kanban Board</a>
					<a href="/">Back to Home</a>
				</div>
			</body>
			</html>
		`)
	})

	// Presence Demo - DISABLED for now
	e.GET("/presence", func(c echo.Context) error {
		return c.HTML(200, `
			<!DOCTYPE html>
			<html>
			<head>
				<title>Presence Demo - Under Construction</title>
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
					.message {
						background: white;
						padding: 40px;
						border-radius: 12px;
						text-align: center;
						box-shadow: 0 10px 30px rgba(0,0,0,0.2);
					}
					h1 { color: #667eea; }
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
				<div class="message">
					<h1>👥 Presence Demo</h1>
					<p>This component is under construction.</p>
					<p>Please try the Kanban Board instead!</p>
					<a href="/kanban">Go to Kanban Board</a>
					<a href="/">Back to Home</a>
				</div>
			</body>
			</html>
		`)
	})

	// Start server
	port := ":8080"
	fmt.Printf("🚀 Collaborative Components Server\n")
	fmt.Printf("🌐 Starting on http://localhost%s\n", port)
	fmt.Println("\n📋 Available demos:")
	fmt.Println("  / - Main page")
	fmt.Println("  /kanban - Kanban board")
	fmt.Println("  /canvas - Drawing canvas")
	fmt.Println("  /presence - User presence")
	
	if err := e.Start(port); err != nil {
		log.Fatal(err)
	}
}