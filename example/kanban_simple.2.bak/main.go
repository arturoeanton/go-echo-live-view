// Package main implements a simple Kanban board server with real-time collaboration
// This application demonstrates the Go Echo LiveView framework capabilities for
// building interactive web applications with WebSocket-based real-time updates.
package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

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
	e.Static("/assets", "../../assets")
	// Serve local assets for this example
	e.Static("/kanban-assets", "assets")

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

	// Add API endpoints for file management
	e.POST("/api/upload/:board/:card", handleFileUpload)
	e.GET("/api/download/:board/:card/:filename", handleFileDownload)
	
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

// handleFileUpload handles file upload via REST API
func handleFileUpload(c echo.Context) error {
	// Get board and card IDs from URL parameters
	boardID := c.Param("board")
	cardID := c.Param("card")
	
	// Parse multipart form
	form, err := c.MultipartForm()
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Failed to parse form",
		})
	}
	
	files := form.File["files"]
	if len(files) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "No files uploaded",
		})
	}
	
	// Create directory for attachments
	dirPath := filepath.Join("attachments", boardID, cardID)
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to create directory",
		})
	}
	
	uploadedFiles := []map[string]interface{}{}
	
	// Process each file
	for _, file := range files {
		// Check file size (5MB max)
		if file.Size > 5*1024*1024 {
			continue // Skip files larger than 5MB
		}
		
		// Open uploaded file
		src, err := file.Open()
		if err != nil {
			continue
		}
		defer src.Close()
		
		// Generate unique filename
		attachmentID := fmt.Sprintf("attach_%d", time.Now().UnixNano())
		filename := attachmentID + "_" + file.Filename
		filePath := filepath.Join(dirPath, filename)
		
		// Create destination file
		dst, err := os.Create(filePath)
		if err != nil {
			continue
		}
		defer dst.Close()
		
		// Copy file content
		if _, err = io.Copy(dst, src); err != nil {
			continue
		}
		
		// Add to uploaded files list
		uploadedFiles = append(uploadedFiles, map[string]interface{}{
			"id":       attachmentID,
			"name":     file.Filename,
			"size":     file.Size,
			"path":     filePath,
			"uploaded": time.Now().Format(time.RFC3339),
		})
		
		fmt.Printf("✅ File uploaded: %s to %s/%s\n", file.Filename, boardID, cardID)
	}
	
	// Update the board state (this would need to be synchronized with the kanban board)
	// For now, we'll return success and let the client update via WebSocket
	
	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"files":   uploadedFiles,
		"message": fmt.Sprintf("Uploaded %d file(s)", len(uploadedFiles)),
	})
}

// handleFileDownload handles file download via REST API
func handleFileDownload(c echo.Context) error {
	// Get parameters from URL
	boardID := c.Param("board")
	cardID := c.Param("card")
	filename := c.Param("filename")
	
	// Construct file path
	filePath := filepath.Join("attachments", boardID, cardID, filename)
	
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "File not found",
		})
	}
	
	// Get the original filename (remove the attachment ID prefix)
	originalName := filename
	if idx := len("attach_") + 19; idx < len(filename) && filename[idx] == '_' {
		// Remove "attach_XXXXXXXXXXXXX_" prefix
		originalName = filename[idx+1:]
	}
	
	// Set headers for download
	c.Response().Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", originalName))
	
	// Serve the file
	return c.File(filePath)
}