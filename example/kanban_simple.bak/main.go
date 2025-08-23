// Package main implements a simple Kanban board server with real-time collaboration
// This application demonstrates the Go Echo LiveView framework capabilities for
// building interactive web applications with WebSocket-based real-time updates.
package main

import (
	"fmt"
	"log"

	"github.com/arturoeanton/go-echo-live-view/liveview"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// main initializes and starts the Kanban board server
// It sets up the Echo web framework, configures middleware, registers the LiveView component,
// and starts listening on port 8080 for incoming connections.
func main() {
	// Create Echo instance - the main web framework
	e := echo.New()
	
	// Add middleware for request logging and panic recovery
	e.Use(middleware.Logger())  // Logs all HTTP requests
	e.Use(middleware.Recover()) // Recovers from panics and returns 500 error
	
	// Serve static files (for assets like live.js WebAssembly module)
	e.Static("/assets", "assets")

	// Create page control - manages the LiveView page lifecycle
	page := &liveview.PageControl{
		Path:   "/",                                    // URL path for the Kanban board
		Title:  "Simple Kanban Board with JSON Storage", // Browser title
		Router: e,                                       // Echo router instance
	}

	// Register the board factory - creates new board instance per WebSocket connection
	// This ensures each user gets their own component instance while sharing global state
	page.Register(func() liveview.LiveDriver {
		// Create new Kanban board instance with modal support
		board := NewSimpleKanbanModal()
		fmt.Println("✅ New client connected to kanban board")
		
		// Return the component driver that manages WebSocket communication
		return board.ComponentDriver
	})

	// Configure server port
	port := ":8080"
	
	// Display startup banner with helpful information
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("📋 Simple Kanban Board Server starting on http://localhost%s\n", port)
	fmt.Println("💾 Board state will be saved to kanban_board.json")
	fmt.Printf("🌐 Open http://localhost%s to view the board\n", port)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	
	// Start the HTTP server
	if err := e.Start(port); err != nil {
		log.Fatal(err)
	}
}