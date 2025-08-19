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
	e.Use(middleware.CORS())

	// Serve static files
	e.Static("/assets", "assets")

	// Store for different project boards
	boards := make(map[string]*components.KanbanBoard)

	// Main page - Project selection
	e.GET("/", func(c echo.Context) error {
		html := `
<!DOCTYPE html>
<html>
<head>
	<title>Collaborative Kanban Board</title>
	<style>
		body {
			font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
			background: linear-gradient(135deg, #3498db 0%, #2ecc71 100%);
			min-height: 100vh;
			margin: 0;
			padding: 20px;
		}
		.container {
			max-width: 1200px;
			margin: 0 auto;
		}
		.header {
			text-align: center;
			color: white;
			margin-bottom: 40px;
		}
		h1 {
			font-size: 48px;
			margin-bottom: 10px;
			text-shadow: 2px 2px 4px rgba(0,0,0,0.2);
		}
		.subtitle {
			font-size: 20px;
			opacity: 0.9;
		}
		.projects-grid {
			display: grid;
			grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
			gap: 20px;
			margin-bottom: 40px;
		}
		.project-card {
			background: white;
			border-radius: 12px;
			padding: 25px;
			box-shadow: 0 10px 30px rgba(0,0,0,0.1);
			text-decoration: none;
			color: #333;
			transition: all 0.3s;
			position: relative;
			overflow: hidden;
		}
		.project-card:hover {
			transform: translateY(-5px);
			box-shadow: 0 15px 40px rgba(0,0,0,0.2);
		}
		.project-card::before {
			content: '';
			position: absolute;
			top: 0;
			left: 0;
			right: 0;
			height: 4px;
			background: linear-gradient(90deg, #3498db, #2ecc71);
		}
		.project-title {
			font-size: 24px;
			font-weight: 600;
			margin-bottom: 10px;
		}
		.project-description {
			color: #666;
			margin-bottom: 20px;
			line-height: 1.5;
		}
		.project-stats {
			display: flex;
			gap: 20px;
			font-size: 14px;
			color: #999;
		}
		.stat {
			display: flex;
			align-items: center;
			gap: 5px;
		}
		.create-button {
			background: white;
			color: #3498db;
			border: 2px dashed #3498db;
			border-radius: 12px;
			padding: 40px;
			text-align: center;
			cursor: pointer;
			transition: all 0.3s;
			display: flex;
			flex-direction: column;
			align-items: center;
			justify-content: center;
			min-height: 200px;
		}
		.create-button:hover {
			background: rgba(255,255,255,0.95);
			border-color: #2ecc71;
			color: #2ecc71;
		}
		.create-icon {
			font-size: 48px;
			margin-bottom: 10px;
		}
		.templates {
			background: white;
			border-radius: 12px;
			padding: 30px;
			margin-top: 40px;
		}
		.templates h2 {
			color: #333;
			margin-bottom: 20px;
		}
		.template-list {
			display: flex;
			gap: 15px;
			flex-wrap: wrap;
		}
		.template-btn {
			padding: 10px 20px;
			background: #f0f0f0;
			border: none;
			border-radius: 6px;
			cursor: pointer;
			transition: all 0.2s;
		}
		.template-btn:hover {
			background: #3498db;
			color: white;
		}
	</style>
</head>
<body>
	<div class="container">
		<div class="header">
			<h1>📋 Collaborative Kanban</h1>
			<p class="subtitle">Real-time project management for teams</p>
		</div>
		
		<div class="projects-grid">
			<a href="/board/product-dev" class="project-card">
				<div class="project-title">🚀 Product Development</div>
				<div class="project-description">
					Main product roadmap and feature development tracking
				</div>
				<div class="project-stats">
					<span class="stat">📝 24 tasks</span>
					<span class="stat">👥 5 members</span>
					<span class="stat">🔥 3 urgent</span>
				</div>
			</a>
			
			<a href="/board/marketing" class="project-card">
				<div class="project-title">📢 Marketing Campaign</div>
				<div class="project-description">
					Q4 marketing initiatives and content calendar
				</div>
				<div class="project-stats">
					<span class="stat">📝 18 tasks</span>
					<span class="stat">👥 3 members</span>
					<span class="stat">✅ 60% done</span>
				</div>
			</a>
			
			<a href="/board/bugs" class="project-card">
				<div class="project-title">🐛 Bug Tracking</div>
				<div class="project-description">
					Critical bugs and issues reported by users
				</div>
				<div class="project-stats">
					<span class="stat">📝 7 bugs</span>
					<span class="stat">👥 2 members</span>
					<span class="stat">🚨 2 critical</span>
				</div>
			</a>
			
			<div class="create-button" onclick="createNewBoard()">
				<div class="create-icon">➕</div>
				<div style="font-size: 18px; font-weight: 600;">Create New Board</div>
				<div style="font-size: 14px; margin-top: 10px;">Start from scratch or use a template</div>
			</div>
		</div>
		
		<div class="templates">
			<h2>🎯 Quick Start Templates</h2>
			<div class="template-list">
				<button class="template-btn" onclick="createFromTemplate('agile')">
					Agile Sprint
				</button>
				<button class="template-btn" onclick="createFromTemplate('personal')">
					Personal Tasks
				</button>
				<button class="template-btn" onclick="createFromTemplate('support')">
					Customer Support
				</button>
				<button class="template-btn" onclick="createFromTemplate('content')">
					Content Pipeline
				</button>
				<button class="template-btn" onclick="createFromTemplate('hiring')">
					Hiring Pipeline
				</button>
			</div>
		</div>
	</div>
	
	<script>
		function createNewBoard() {
			const name = prompt('Enter board name:');
			if (name) {
				window.location.href = '/board/' + encodeURIComponent(name);
			}
		}
		
		function createFromTemplate(template) {
			window.location.href = '/board/template-' + template;
		}
	</script>
</body>
</html>
		`
		return c.HTML(200, html)
	})

	// Board handler
	boardHandler := func(boardID string) *components.KanbanBoard {
		board, exists := boards[boardID]
		if !exists {
			// Create new board
			board = &components.KanbanBoard{}
			board.ComponentDriver = liveview.NewDriver[*components.KanbanBoard](fmt.Sprintf("board_%s", boardID), board)

			// Initialize based on board type
			if boardID == "product-dev" {
				initializeProductBoard(board)
			} else if boardID == "marketing" {
				initializeMarketingBoard(board)
			} else if boardID == "bugs" {
				initializeBugBoard(board)
			} else {
				// Default initialization
				board.Start()
			}

			boards[boardID] = board
		}
		return board
	}

	// Create PageControl for boards
	boardPage := &liveview.PageControl{
		Path:   "/board",
		Title:  "Kanban Board",
		Router: e,
	}

	// Register board handler
	boardPage.Register(func() liveview.LiveDriver {
		// Get board ID from the request context
		// For now, use a default board
		board := boardHandler("default")
		return board.ComponentDriver
	})

	// Alternative routes for different boards
	e.GET("/board/:id", func(c echo.Context) error {
		boardID := c.Param("id")
		_ = boardHandler(boardID)

		// Return the board page HTML
		return c.HTML(200, fmt.Sprintf(`
			<!DOCTYPE html>
			<html>
			<head>
				<title>Kanban Board - %s</title>
				<meta charset="utf-8"/>
				<script src="/assets/wasm_exec.js"></script>
			</head>
			<body>
				<div id="content"></div>
				<script>
					const go = new Go();
					WebAssembly.instantiateStreaming(fetch("/assets/json.wasm"), go.importObject).then((result) => {
						go.run(result.instance);
					});
				</script>
			</body>
			</html>
		`, boardID))
	})

	// Start server
	port := ":8080"
	fmt.Printf("📋 Kanban Board Server starting on http://localhost%s\n", port)
	fmt.Println("📝 Available routes:")
	fmt.Println("  - / : Project selection")
	fmt.Println("  - /board/:id : Open a kanban board")

	if err := e.Start(port); err != nil {
		log.Fatal(err)
	}
}

// Initialize product development board
func initializeProductBoard(board *components.KanbanBoard) {
	board.Title = "Product Development"
	board.Description = "Track features and improvements"

	// Custom columns for product dev
	board.Columns = []components.KanbanColumn{
		{ID: "ideas", Title: "Ideas", Color: "#95a5a6", Order: 0},
		{ID: "approved", Title: "Approved", Color: "#3498db", Order: 1},
		{ID: "development", Title: "Development", Color: "#f39c12", Order: 2, WIPLimit: 3},
		{ID: "testing", Title: "Testing", Color: "#e67e22", Order: 3, WIPLimit: 2},
		{ID: "staging", Title: "Staging", Color: "#9b59b6", Order: 4},
		{ID: "production", Title: "Production", Color: "#27ae60", Order: 5},
	}

	// Add sample cards
	board.Cards = []components.KanbanCard{
		{
			ID:          "feat1",
			Title:       "User Dashboard Redesign",
			Description: "Modernize the user dashboard with new metrics and visualizations",
			ColumnID:    "development",
			Priority:    "high",
			Points:      8,
			Labels:      []string{"feature", "ui"},
			CreatedAt:   time.Now().Add(-72 * time.Hour),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          "feat2",
			Title:       "API Rate Limiting",
			Description: "Implement rate limiting to prevent API abuse",
			ColumnID:    "testing",
			Priority:    "medium",
			Points:      5,
			Labels:      []string{"feature", "backend"},
			CreatedAt:   time.Now().Add(-48 * time.Hour),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          "feat3",
			Title:       "Mobile App Push Notifications",
			Description: "Add push notification support for mobile apps",
			ColumnID:    "approved",
			Priority:    "medium",
			Points:      13,
			Labels:      []string{"feature", "mobile"},
			CreatedAt:   time.Now().Add(-24 * time.Hour),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          "feat4",
			Title:       "Dark Mode Support",
			Description: "Implement dark mode across all UI components",
			ColumnID:    "ideas",
			Priority:    "low",
			Points:      5,
			Labels:      []string{"enhancement", "ui"},
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	board.Labels = []components.KanbanLabel{
		{ID: "feature", Name: "Feature", Color: "#3498db"},
		{ID: "enhancement", Name: "Enhancement", Color: "#2ecc71"},
		{ID: "ui", Name: "UI/UX", Color: "#e74c3c"},
		{ID: "backend", Name: "Backend", Color: "#f39c12"},
		{ID: "mobile", Name: "Mobile", Color: "#9b59b6"},
	}

	board.ActiveUsers = make(map[string]*components.UserActivity)
	board.Commit()
}

// Initialize marketing board
func initializeMarketingBoard(board *components.KanbanBoard) {
	board.Title = "Marketing Campaign"
	board.Description = "Q4 marketing initiatives"

	board.Columns = []components.KanbanColumn{
		{ID: "planning", Title: "Planning", Color: "#95a5a6", Order: 0},
		{ID: "content", Title: "Content Creation", Color: "#3498db", Order: 1, WIPLimit: 4},
		{ID: "review", Title: "Review", Color: "#f39c12", Order: 2, WIPLimit: 2},
		{ID: "scheduled", Title: "Scheduled", Color: "#9b59b6", Order: 3},
		{ID: "published", Title: "Published", Color: "#27ae60", Order: 4},
	}

	board.Cards = []components.KanbanCard{
		{
			ID:          "mkt1",
			Title:       "Black Friday Campaign",
			Description: "Email and social media campaign for Black Friday",
			ColumnID:    "content",
			Priority:    "urgent",
			DueDate:     &[]time.Time{time.Now().Add(7 * 24 * time.Hour)}[0],
			CreatedAt:   time.Now().Add(-48 * time.Hour),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          "mkt2",
			Title:       "Product Launch Blog Post",
			Description: "Write blog post announcing new features",
			ColumnID:    "review",
			Priority:    "high",
			CreatedAt:   time.Now().Add(-24 * time.Hour),
			UpdatedAt:   time.Now(),
		},
	}

	board.Labels = []components.KanbanLabel{
		{ID: "blog", Name: "Blog", Color: "#3498db"},
		{ID: "social", Name: "Social Media", Color: "#e74c3c"},
		{ID: "email", Name: "Email", Color: "#2ecc71"},
		{ID: "video", Name: "Video", Color: "#9b59b6"},
	}

	board.ActiveUsers = make(map[string]*components.UserActivity)
	board.Commit()
}

// Initialize bug tracking board
func initializeBugBoard(board *components.KanbanBoard) {
	board.Title = "Bug Tracking"
	board.Description = "Track and fix reported issues"

	board.Columns = []components.KanbanColumn{
		{ID: "reported", Title: "Reported", Color: "#e74c3c", Order: 0},
		{ID: "triaged", Title: "Triaged", Color: "#f39c12", Order: 1},
		{ID: "fixing", Title: "Fixing", Color: "#3498db", Order: 2, WIPLimit: 2},
		{ID: "testing", Title: "Testing", Color: "#9b59b6", Order: 3, WIPLimit: 1},
		{ID: "resolved", Title: "Resolved", Color: "#27ae60", Order: 4},
	}

	board.Cards = []components.KanbanCard{
		{
			ID:          "bug1",
			Title:       "Login fails with special characters",
			Description: "Users cannot login if password contains #, @, or &",
			ColumnID:    "fixing",
			Priority:    "urgent",
			Labels:      []string{"bug", "authentication"},
			Blocked:     false,
			CreatedAt:   time.Now().Add(-72 * time.Hour),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          "bug2",
			Title:       "Export CSV has wrong encoding",
			Description: "Non-ASCII characters appear corrupted in exported CSV files",
			ColumnID:    "triaged",
			Priority:    "medium",
			Labels:      []string{"bug", "export"},
			CreatedAt:   time.Now().Add(-48 * time.Hour),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          "bug3",
			Title:       "Mobile menu doesn't close",
			Description: "Menu stays open after selecting an item on mobile devices",
			ColumnID:    "reported",
			Priority:    "low",
			Labels:      []string{"bug", "mobile", "ui"},
			CreatedAt:   time.Now().Add(-24 * time.Hour),
			UpdatedAt:   time.Now(),
		},
	}

	board.Labels = []components.KanbanLabel{
		{ID: "bug", Name: "Bug", Color: "#e74c3c"},
		{ID: "critical", Name: "Critical", Color: "#c0392b"},
		{ID: "authentication", Name: "Auth", Color: "#3498db"},
		{ID: "export", Name: "Export", Color: "#95a5a6"},
		{ID: "mobile", Name: "Mobile", Color: "#9b59b6"},
		{ID: "ui", Name: "UI", Color: "#2ecc71"},
	}

	board.ActiveUsers = make(map[string]*components.UserActivity)
	board.Commit()
}
