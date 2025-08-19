package main

import (
	"fmt"
	"log"

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

	// Create page control
	page := &liveview.PageControl{
		Path:   "/",
		Title:  "Kanban Board Demo",
		Router: e,
	}

	// Register the board factory
	page.Register(func() liveview.LiveDriver {
		// Create a new kanban board instance for each connection
		board := &components.KanbanBoard{}
		
		// Initialize CollaborativeComponent first to avoid nil pointer
		board.CollaborativeComponent = &liveview.CollaborativeComponent{
			Driver: nil, // Will be set by ComponentDriver
		}
		
		// Now create the driver
		board.ComponentDriver = liveview.NewDriver[*components.KanbanBoard]("kanban_main", board)
		
		// Set the driver reference in CollaborativeComponent
		board.CollaborativeComponent.Driver = board.ComponentDriver
		
		return board.ComponentDriver
	})

	// Start server
	port := ":8080"
	fmt.Printf("📋 Kanban Board Server starting on http://localhost%s\n", port)
	fmt.Println("🌐 Open http://localhost%s to view the board\n", port)
	
	if err := e.Start(port); err != nil {
		log.Fatal(err)
	}
}