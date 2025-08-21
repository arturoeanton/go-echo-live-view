package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/arturoeanton/go-echo-live-view/liveview"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// CollaborativeBoard represents a shared workspace
type CollaborativeBoard struct {
	*liveview.ComponentDriver[*CollaborativeBoard]
	
	// Board state
	Title       string
	Users       map[string]*User
	Cards       []Card
	LastUpdate  time.Time
	
	// Framework features
	errorBoundary  *liveview.ErrorBoundary
	stateManager   *liveview.StateManager
	eventRegistry  *liveview.EventRegistry
	lifecycle      *liveview.LifecycleManager
	mu             sync.RWMutex
}

type User struct {
	ID        string
	Name      string
	Color     string
	Cursor    Position
	LastSeen  time.Time
	IsActive  bool
}

type Position struct {
	X int
	Y int
}

type Card struct {
	ID       string
	Title    string
	Content  string
	Position Position
	Author   string
	Color    string
	Created  time.Time
	Modified time.Time
}

func (c *CollaborativeBoard) Start() {
	// Initialize Error Boundary with recovery
	c.errorBoundary = liveview.NewErrorBoundary(100, true)
	c.errorBoundary.SetFallbackRenderer(func(componentID string, err error) string {
		return fmt.Sprintf(`<div class="error-fallback">Board temporarily unavailable: %v</div>`, err)
	})
	
	// Initialize State Manager with persistence
	c.stateManager = liveview.NewStateManager(&liveview.StateConfig{
		Provider:        liveview.NewMemoryStateProvider(),
		CacheEnabled:    true,
		CacheTTL:        10 * time.Minute,
		AutoPersist:     true,
		PersistInterval: 30 * time.Second,
	})
	
	// Initialize Event Registry for real-time updates
	c.eventRegistry = liveview.NewEventRegistry(&liveview.EventRegistryConfig{
		MaxHandlersPerEvent: 10,
		EnableMetrics:       true,
		EnableWildcards:     true,
		DefaultTimeout:      30 * time.Second,
	})
	
	// Setup event handlers
	c.eventRegistry.On("user.joined", func(ctx context.Context, event *liveview.Event) error {
		log.Printf("User joined: %v", event.Data)
		return nil
	})
	
	c.eventRegistry.On("card.moved", func(ctx context.Context, event *liveview.Event) error {
		log.Printf("Card moved: %v", event.Data)
		c.stateManager.Set("last_card_move", time.Now())
		return nil
	})
	
	// Initialize Lifecycle Manager
	c.lifecycle = liveview.NewLifecycleManager("collaborative_board")
	c.lifecycle.SetHooks(&liveview.LifecycleHooks{
		OnCreated: func() error {
			log.Println("Collaborative board created")
			return c.loadInitialState()
		},
		OnMounted: func() error {
			log.Println("Collaborative board mounted")
			c.eventRegistry.Emit("board.mounted", map[string]interface{}{
				"board_id": c.IdComponent,
				"time":     time.Now(),
			})
			return nil
		},
		OnBeforeUnmount: func() error {
			log.Println("Saving board state before unmount")
			return c.saveState()
		},
	})
	
	// Initialize board state
	c.Title = "Collaborative Workspace"
	c.Users = make(map[string]*User)
	c.Cards = []Card{
		{
			ID:       "card1",
			Title:    "Error Boundaries",
			Content:  "Automatic error recovery keeps the board running",
			Position: Position{X: 100, Y: 100},
			Author:   "System",
			Color:    "#667eea",
			Created:  time.Now(),
			Modified: time.Now(),
		},
		{
			ID:       "card2",
			Title:    "State Persistence",
			Content:  "All changes are automatically saved",
			Position: Position{X: 300, Y: 100},
			Author:   "System",
			Color:    "#48bb78",
			Created:  time.Now(),
			Modified: time.Now(),
		},
		{
			ID:       "card3",
			Title:    "Real-time Events",
			Content:  "Instant updates across all connected users",
			Position: Position{X: 500, Y: 100},
			Author:   "System",
			Color:    "#ed8936",
			Created:  time.Now(),
			Modified: time.Now(),
		},
	}
	c.LastUpdate = time.Now()
	
	// Execute lifecycle
	c.lifecycle.Create()
	c.lifecycle.Mount()
	
	// Start auto-save ticker
	go c.autoSaveLoop()
	
	c.Commit()
}

func (c *CollaborativeBoard) loadInitialState() error {
	// Load cards from state if available
	if savedCards, err := c.stateManager.Get("board_cards"); err == nil && savedCards != nil {
		if cards, ok := savedCards.([]Card); ok {
			c.Cards = cards
			log.Printf("Loaded %d cards from state", len(cards))
		}
	}
	
	// Load users from state
	if savedUsers, err := c.stateManager.Get("board_users"); err == nil && savedUsers != nil {
		if users, ok := savedUsers.(map[string]*User); ok {
			c.Users = users
			log.Printf("Loaded %d users from state", len(users))
		}
	}
	
	return nil
}

func (c *CollaborativeBoard) saveState() error {
	// Save current state
	if err := c.stateManager.Set("board_cards", c.Cards); err != nil {
		return err
	}
	if err := c.stateManager.Set("board_users", c.Users); err != nil {
		return err
	}
	c.stateManager.Set("last_save", time.Now())
	return nil
}

func (c *CollaborativeBoard) autoSaveLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for range ticker.C {
		if c.lifecycle.GetStage() == liveview.StageMounted {
			c.errorBoundary.SafeExecute("autosave", func() error {
				return c.saveState()
			})
		}
	}
}

func (c *CollaborativeBoard) GetTemplate() string {
	return `
<!DOCTYPE html>
<html>
<head>
	<title>Collaborative Workspace</title>
	<style>
		body {
			font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
			margin: 0;
			background: #f0f2f5;
			height: 100vh;
			overflow: hidden;
		}
		.header {
			background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
			color: white;
			padding: 1rem 2rem;
			box-shadow: 0 2px 10px rgba(0,0,0,0.1);
			display: flex;
			justify-content: space-between;
			align-items: center;
		}
		.board {
			position: relative;
			height: calc(100vh - 70px);
			overflow: auto;
			background: #f0f2f5;
			background-image: 
				linear-gradient(0deg, transparent 24%, rgba(0,0,0,.02) 25%, rgba(0,0,0,.02) 26%, transparent 27%, transparent 74%, rgba(0,0,0,.02) 75%, rgba(0,0,0,.02) 76%, transparent 77%, transparent),
				linear-gradient(90deg, transparent 24%, rgba(0,0,0,.02) 25%, rgba(0,0,0,.02) 26%, transparent 27%, transparent 74%, rgba(0,0,0,.02) 75%, rgba(0,0,0,.02) 76%, transparent 77%, transparent);
			background-size: 50px 50px;
		}
		.card {
			position: absolute;
			background: white;
			border-radius: 8px;
			padding: 1rem;
			min-width: 200px;
			box-shadow: 0 4px 6px rgba(0,0,0,0.1);
			cursor: move;
			transition: transform 0.2s, box-shadow 0.2s;
		}
		.card:hover {
			transform: translateY(-2px);
			box-shadow: 0 8px 12px rgba(0,0,0,0.15);
		}
		.card-title {
			font-weight: 600;
			margin-bottom: 0.5rem;
			padding-bottom: 0.5rem;
			border-bottom: 2px solid;
		}
		.card-content {
			color: #666;
			font-size: 0.9rem;
		}
		.card-meta {
			margin-top: 0.5rem;
			padding-top: 0.5rem;
			border-top: 1px solid #eee;
			font-size: 0.75rem;
			color: #999;
		}
		.users {
			display: flex;
			gap: 0.5rem;
		}
		.user-avatar {
			width: 32px;
			height: 32px;
			border-radius: 50%;
			display: flex;
			align-items: center;
			justify-content: center;
			color: white;
			font-weight: bold;
			font-size: 0.875rem;
		}
		.stats {
			display: flex;
			gap: 2rem;
			background: rgba(255,255,255,0.1);
			padding: 0.5rem 1rem;
			border-radius: 20px;
		}
		.stat {
			display: flex;
			align-items: center;
			gap: 0.5rem;
		}
		.stat-value {
			font-weight: bold;
		}
		.add-card-btn {
			position: fixed;
			bottom: 2rem;
			right: 2rem;
			width: 56px;
			height: 56px;
			border-radius: 50%;
			background: #667eea;
			color: white;
			border: none;
			font-size: 24px;
			cursor: pointer;
			box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
			transition: all 0.3s;
		}
		.add-card-btn:hover {
			transform: scale(1.1);
			box-shadow: 0 6px 20px rgba(102, 126, 234, 0.6);
		}
		.error-fallback {
			padding: 2rem;
			background: #fee;
			border: 2px solid #f88;
			border-radius: 8px;
			text-align: center;
			margin: 2rem;
		}
		.feature-badge {
			display: inline-block;
			padding: 0.25rem 0.5rem;
			background: rgba(255,255,255,0.2);
			border-radius: 12px;
			font-size: 0.75rem;
			margin-left: 0.5rem;
		}
	</style>
</head>
<body>
	<div class="header">
		<div>
			<h1 style="margin: 0; font-size: 1.5rem;">{{.Title}}
				<span class="feature-badge">Error Protected</span>
				<span class="feature-badge">Auto-Save</span>
				<span class="feature-badge">Real-time</span>
			</h1>
		</div>
		<div class="stats">
			<div class="stat">
				<span>Cards:</span>
				<span class="stat-value">{{len .Cards}}</span>
			</div>
			<div class="stat">
				<span>Users:</span>
				<span class="stat-value">{{len .Users}}</span>
			</div>
			<div class="stat">
				<span>Last Update:</span>
				<span class="stat-value">{{.LastUpdate.Format "15:04:05"}}</span>
			</div>
		</div>
	</div>
	
	<div class="board" id="board">
		{{range .Cards}}
		<div class="card" style="left: {{.Position.X}}px; top: {{.Position.Y}}px; border-color: {{.Color}};">
			<div class="card-title" style="border-color: {{.Color}}; color: {{.Color}};">{{.Title}}</div>
			<div class="card-content">{{.Content}}</div>
			<div class="card-meta">
				<div>By {{.Author}}</div>
				<div>{{.Modified.Format "Jan 2, 15:04"}}</div>
			</div>
		</div>
		{{end}}
	</div>
	
	<button class="add-card-btn" onclick="send_event('{{.IdComponent}}', 'AddCard', null)">+</button>
	
	<script>
		// Enable drag and drop
		document.querySelectorAll('.card').forEach(card => {
			let isDragging = false;
			let startX, startY, initialX, initialY;
			
			card.addEventListener('mousedown', (e) => {
				isDragging = true;
				startX = e.clientX;
				startY = e.clientY;
				initialX = card.offsetLeft;
				initialY = card.offsetTop;
				card.style.zIndex = 1000;
			});
			
			document.addEventListener('mousemove', (e) => {
				if (!isDragging) return;
				e.preventDefault();
				const dx = e.clientX - startX;
				const dy = e.clientY - startY;
				card.style.left = (initialX + dx) + 'px';
				card.style.top = (initialY + dy) + 'px';
			});
			
			document.addEventListener('mouseup', (e) => {
				if (isDragging) {
					isDragging = false;
					card.style.zIndex = '';
					// Send position update
					send_event('{{.IdComponent}}', 'MoveCard', {
						cardId: card.dataset.id,
						x: card.offsetLeft,
						y: card.offsetTop
					});
				}
			});
		});
	</script>
</body>
</html>
	`
}

func (c *CollaborativeBoard) GetDriver() liveview.LiveDriver {
	return c
}

func (c *CollaborativeBoard) AddCard(data interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	// Use error boundary for safe execution
	c.errorBoundary.SafeExecute("add_card", func() error {
		newCard := Card{
			ID:       fmt.Sprintf("card_%d", time.Now().Unix()),
			Title:    "New Card",
			Content:  "Click to edit",
			Position: Position{X: 100 + len(c.Cards)*20, Y: 200},
			Author:   "User",
			Color:    "#667eea",
			Created:  time.Now(),
			Modified: time.Now(),
		}
		
		c.Cards = append(c.Cards, newCard)
		c.LastUpdate = time.Now()
		
		// Emit event
		c.eventRegistry.Emit("card.added", map[string]interface{}{
			"card_id": newCard.ID,
			"author":  newCard.Author,
		})
		
		// Save state
		c.saveState()
		
		return nil
	})
	
	c.Commit()
}

func (c *CollaborativeBoard) MoveCard(data interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	// Card movement would be implemented here
	c.LastUpdate = time.Now()
	
	// Emit event
	c.eventRegistry.Emit("card.moved", data.(map[string]interface{}))
	
	c.Commit()
}

func main() {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	
	// Main page
	page := &liveview.PageControl{
		Path:   "/",
		Title:  "Collaborative Workspace v2",
		Router: e,
	}
	
	page.Register(func() liveview.LiveDriver {
		return liveview.NewDriver("collab_board", &CollaborativeBoard{})
	})
	
	port := ":8081"
	fmt.Printf("Starting Collaborative Workspace v2\n")
	fmt.Printf("Open http://localhost%s\n", port)
	fmt.Println("\nFeatures:")
	fmt.Println("  • Error Boundaries with auto-recovery")
	fmt.Println("  • State persistence with auto-save")
	fmt.Println("  • Real-time event system")
	fmt.Println("  • Lifecycle management")
	
	if err := e.Start(port); err != nil {
		log.Fatal(err)
	}
}