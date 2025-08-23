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

	// Main page
	e.GET("/", func(c echo.Context) error {
		html := `
<!DOCTYPE html>
<html>
<head>
	<title>LiveView Collaborative Demo</title>
	<style>
		body {
			font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
			background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
			min-height: 100vh;
			margin: 0;
			padding: 40px 20px;
		}
		.container {
			max-width: 800px;
			margin: 0 auto;
			text-align: center;
		}
		h1 {
			color: white;
			font-size: 48px;
			margin-bottom: 10px;
			text-shadow: 2px 2px 4px rgba(0,0,0,0.2);
		}
		.subtitle {
			color: rgba(255,255,255,0.9);
			font-size: 20px;
			margin-bottom: 40px;
		}
		.demo-link {
			display: inline-block;
			background: white;
			color: #667eea;
			padding: 15px 30px;
			margin: 10px;
			border-radius: 8px;
			text-decoration: none;
			font-weight: 600;
			transition: all 0.3s;
			box-shadow: 0 4px 15px rgba(0,0,0,0.1);
		}
		.demo-link:hover {
			transform: translateY(-2px);
			box-shadow: 0 6px 20px rgba(0,0,0,0.2);
		}
		.features {
			background: white;
			border-radius: 12px;
			padding: 30px;
			margin-top: 40px;
			box-shadow: 0 10px 30px rgba(0,0,0,0.1);
		}
		.feature {
			margin: 20px 0;
			text-align: left;
		}
		.feature h3 {
			color: #667eea;
			margin-bottom: 10px;
		}
		.feature p {
			color: #666;
			line-height: 1.6;
		}
	</style>
</head>
<body>
	<div class="container">
		<h1>🚀 Go Echo LiveView</h1>
		<p class="subtitle">Real-time Collaborative Components</p>
		
		<div>
			<a href="/kanban" class="demo-link">📋 Kanban Board Demo</a>
		</div>
		
		<div class="features">
			<h2 style="color: #333;">✨ Features</h2>
			
			<div class="feature">
				<h3>🔄 Real-time Sync</h3>
				<p>Changes are instantly synchronized across all connected clients using WebSockets.</p>
			</div>
			
			<div class="feature">
				<h3>🎯 Drag & Drop</h3>
				<p>Intuitive drag and drop interface for organizing tasks between columns.</p>
			</div>
			
			<div class="feature">
				<h3>👥 Multi-user Support</h3>
				<p>See who's online and what they're working on in real-time.</p>
			</div>
			
			<div class="feature">
				<h3>⚡ Server-side Rendering</h3>
				<p>All logic runs on the server - no complex JavaScript frameworks needed!</p>
			</div>
		</div>
	</div>
</body>
</html>
		`
		return c.HTML(200, html)
	})

	// Kanban Board Page
	kanbanPage := &liveview.PageControl{
		Path:   "/kanban",
		Title:  "Kanban Board - LiveView Demo",
		Router: e,
	}

	// Register kanban board factory
	kanbanPage.Register(func() liveview.LiveDriver {
		// Create board instance
		board := createKanbanBoard()
		return board.ComponentDriver
	})

	// Start server
	port := ":8089"
	fmt.Printf("🚀 LiveView Collaborative Demo Server\n")
	fmt.Printf("🌐 Starting on http://localhost%s\n", port)
	fmt.Println("\n📋 Available routes:")
	fmt.Println("  / - Main page with information")
	fmt.Println("  /kanban - Interactive Kanban board")

	if err := e.Start(port); err != nil {
		log.Fatal(err)
	}
}

// createKanbanBoard creates a properly initialized kanban board
func createKanbanBoard() *components.KanbanBoard {
	// Create board
	board := &components.KanbanBoard{}

	// IMPORTANT: Initialize embedded struct first
	board.CollaborativeComponent = &liveview.CollaborativeComponent{}

	// Then create the driver
	board.ComponentDriver = liveview.NewDriver[*components.KanbanBoard]("kanban_board", board)

	// Set the driver reference
	if board.CollaborativeComponent != nil {
		board.CollaborativeComponent.Driver = board.ComponentDriver
	}

	// Initialize board data
	board.Title = "Project Management Board"
	board.Description = "Drag cards between columns to organize your work"

	// Setup columns
	board.Columns = []components.KanbanColumn{
		{
			ID:    "backlog",
			Title: "📝 Backlog",
			Color: "#95a5a6",
			Order: 0,
		},
		{
			ID:       "todo",
			Title:    "📋 To Do",
			Color:    "#3498db",
			Order:    1,
			WIPLimit: 5,
		},
		{
			ID:       "progress",
			Title:    "🚀 In Progress",
			Color:    "#f39c12",
			Order:    2,
			WIPLimit: 3,
		},
		{
			ID:       "review",
			Title:    "👀 Review",
			Color:    "#9b59b6",
			Order:    3,
			WIPLimit: 2,
		},
		{
			ID:    "done",
			Title: "✅ Done",
			Color: "#27ae60",
			Order: 4,
		},
	}

	// Add sample cards
	board.Cards = []components.KanbanCard{
		{
			ID:          "task1",
			Title:       "Setup project repository",
			Description: "Initialize Git repo and add README",
			ColumnID:    "done",
			Priority:    "high",
			Points:      2,
			CreatedAt:   time.Now().Add(-72 * time.Hour),
			UpdatedAt:   time.Now().Add(-48 * time.Hour),
			Completed:   true,
		},
		{
			ID:           "task2",
			Title:        "Design database schema",
			Description:  "Create ERD and define all tables with relationships",
			ColumnID:     "progress",
			Priority:     "high",
			Points:       5,
			CreatedAt:    time.Now().Add(-48 * time.Hour),
			UpdatedAt:    time.Now(),
			AssigneeName: "Alice",
		},
		{
			ID:           "task3",
			Title:        "Implement user authentication",
			Description:  "Add login, registration, and password reset",
			ColumnID:     "progress",
			Priority:     "high",
			Points:       8,
			CreatedAt:    time.Now().Add(-24 * time.Hour),
			UpdatedAt:    time.Now(),
			AssigneeName: "Bob",
		},
		{
			ID:          "task4",
			Title:       "Create API documentation",
			Description: "Document all REST endpoints with examples",
			ColumnID:    "todo",
			Priority:    "medium",
			Points:      3,
			CreatedAt:   time.Now().Add(-12 * time.Hour),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          "task5",
			Title:       "Setup CI/CD pipeline",
			Description: "Configure GitHub Actions for automated testing and deployment",
			ColumnID:    "todo",
			Priority:    "medium",
			Points:      5,
			CreatedAt:   time.Now().Add(-6 * time.Hour),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          "task6",
			Title:       "Add unit tests",
			Description: "Write comprehensive test coverage for core functionality",
			ColumnID:    "backlog",
			Priority:    "low",
			Points:      8,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          "task7",
			Title:       "Performance optimization",
			Description: "Profile and optimize database queries",
			ColumnID:    "backlog",
			Priority:    "low",
			Points:      5,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:           "task8",
			Title:        "Mobile responsive design",
			Description:  "Ensure UI works well on all screen sizes",
			ColumnID:     "review",
			Priority:     "medium",
			Points:       3,
			CreatedAt:    time.Now().Add(-3 * time.Hour),
			UpdatedAt:    time.Now(),
			AssigneeName: "Charlie",
		},
	}

	// Setup labels
	board.Labels = []components.KanbanLabel{
		{ID: "feature", Name: "Feature", Color: "#3498db"},
		{ID: "bug", Name: "Bug", Color: "#e74c3c"},
		{ID: "enhancement", Name: "Enhancement", Color: "#2ecc71"},
		{ID: "documentation", Name: "Docs", Color: "#95a5a6"},
		{ID: "testing", Name: "Testing", Color: "#f39c12"},
	}

	// Initialize active users map
	board.ActiveUsers = make(map[string]*components.UserActivity)

	return board
}
