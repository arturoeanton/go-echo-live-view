package main

import (
	"fmt"
	"log"

	"github.com/arturoeanton/go-echo-live-view/liveview"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// Create Echo instance
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	
	// Serve static files (for assets like live.js)
	e.Static("/assets", "assets")

	// Create page control
	page := &liveview.PageControl{
		Path:   "/",
		Title:  "Simple Kanban Board with JSON Storage",
		Router: e,
	}

	// Register the board factory - creates new connection per client
	page.Register(func() liveview.LiveDriver {
		// Create Simple KanbanBoard with modals (no collaboration to avoid crashes)
		board := NewSimpleKanbanModal()
		fmt.Println("✅ New client connected to kanban board")
		
		return board.ComponentDriver
	})

	// Start server
	port := ":8080"
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("📋 Simple Kanban Board Server starting on http://localhost%s\n", port)
	fmt.Println("💾 Board state will be saved to kanban_board.json")
	fmt.Printf("🌐 Open http://localhost%s to view the board\n", port)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	
	if err := e.Start(port); err != nil {
		log.Fatal(err)
	}
}