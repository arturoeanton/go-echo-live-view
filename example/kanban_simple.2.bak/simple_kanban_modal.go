package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/arturoeanton/go-echo-live-view/components"
	"github.com/arturoeanton/go-echo-live-view/liveview"
)

// BOARDS_DIR is the directory where kanban board JSON files are stored
const BOARDS_DIR = "kanban_data"

// Global state for synchronization across multiple connected clients
// These variables ensure all users see the same board state in real-time
var (
	globalMutex     sync.RWMutex                    // Protects read/write access to globalBoards
	globalBoards    map[string]*KanbanBoardData     // Map of board name to board data
	activeBoards    []*SimpleKanbanModal            // List of all active board instances
	activeMutex     sync.Mutex                      // Protects access to activeBoards slice
	currentBoardName string = "default"             // Currently active board name
)

// KanbanBoardData represents the persistent board state that is saved to JSON
// This structure contains all columns and cards in the Kanban board
type KanbanBoardData struct {
	Columns []KanbanColumn `json:"columns"` // List of board columns
	Cards   []KanbanCard   `json:"cards"`   // List of all cards across columns
}

// SimpleKanbanModal is the main component for the Kanban board application
// It manages board state, handles user interactions, and coordinates real-time updates
type SimpleKanbanModal struct {
	*liveview.ComponentDriver[*SimpleKanbanModal] `json:"-"` // Embedded LiveView driver for WebSocket communication
	
	// Board data - synchronized across all users
	Columns      []KanbanColumn `json:"columns"`       // Current columns in the board
	Cards        []KanbanCard   `json:"cards"`         // Current cards across all columns
	CurrentBoard string         `json:"current_board"` // Name of the currently active board
	BoardsList   []string       `json:"boards_list"`   // List of available board names
	
	// UI state
	ShowAlert    bool   `json:"show_alert"`    // Whether to show instructions alert
	
	// Components
	Dropdown     *components.Dropdown     `json:"-"` // Dropdown component instance
	FileUpload   *components.FileUpload   `json:"-"` // File upload component for attachments
	
	// Modal state - controls the popup dialog
	ShowModal     bool   `json:"show_modal"`     // Whether modal is visible
	ModalType     string `json:"modal_type"`     // Type: "edit_card", "add_card", "edit_column", "add_column", "new_board"
	ModalTitle    string `json:"modal_title"`    // Title shown in modal header
	
	// Form fields for card editing/creation modal
	FormCardID          string       `json:"form_card_id"`          // ID of card being edited
	FormCardTitle       string       `json:"form_card_title"`       // Card title input value
	FormCardDesc        string       `json:"form_card_desc"`        // Card description input value
	FormCardColumn      string       `json:"form_card_column"`      // Selected column for the card
	FormCardPriority    string       `json:"form_card_priority"`    // Selected priority level
	FormCardPoints      int          `json:"form_card_points"`      // Story points (0-100)
	FormCardAttachments []Attachment `json:"form_card_attachments"` // Current attachments
	
	// Form fields for column editing/creation modal
	FormColumnID    string `json:"form_column_id"`    // ID of column being edited
	FormColumnTitle string `json:"form_column_title"` // Column title input value
	FormColumnColor string `json:"form_column_color"` // Column header color
	
	// Form field for new board creation
	FormBoardName string `json:"form_board_name"` // Name for the new board
}

// KanbanColumn represents a single column in the Kanban board
// Columns can be reordered by dragging their headers
type KanbanColumn struct {
	ID    string `json:"id"`    // Unique identifier for the column
	Title string `json:"title"` // Display name of the column
	Color string `json:"color"` // Background color for the column header (hex format)
	Order int    `json:"order"` // Display order (lower numbers appear first)
}

// KanbanCard represents a single task/card in the Kanban board
// Cards can be moved between columns via drag and drop
type KanbanCard struct {
	ID          string       `json:"id"`          // Unique identifier for the card
	Title       string       `json:"title"`       // Card title (required)
	Description string       `json:"description"` // Detailed description of the task
	ColumnID    string       `json:"column_id"`   // ID of the column containing this card
	Priority    string       `json:"priority"`    // Priority level: low, medium, high, urgent
	Points      int          `json:"points"`      // Story points for effort estimation (0-100)
	Attachments []Attachment `json:"attachments"` // File attachments
	CreatedAt   time.Time    `json:"created_at"`  // Timestamp when card was created
	UpdatedAt   time.Time    `json:"updated_at"`  // Timestamp of last modification
}

// Attachment represents a file attached to a card
type Attachment struct {
	ID       string    `json:"id"`       // Unique identifier
	Name     string    `json:"name"`     // Original file name
	Size     int64     `json:"size"`     // File size in bytes
	Type     string    `json:"type"`     // MIME type
	Path     string    `json:"path"`     // Server file path
	UploadAt time.Time `json:"upload_at"` // Upload timestamp
}

// loadBoardData loads the board state from the specified JSON file
// If the file doesn't exist, it returns a default board configuration
func loadBoardData(boardName string) *KanbanBoardData {
	filePath := filepath.Join(BOARDS_DIR, boardName+".json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		// Return default board if file doesn't exist
		fmt.Printf("Board file %s not found, creating default board: %v\n", filePath, err)
		return &KanbanBoardData{
			Columns: []KanbanColumn{
				{ID: "todo", Title: "To Do", Color: "#e3e8ef", Order: 0},
				{ID: "doing", Title: "In Progress", Color: "#ffd4a3", Order: 1},
				{ID: "done", Title: "Done", Color: "#a3e4d7", Order: 2},
			},
			Cards: []KanbanCard{
				{
					ID:          "welcome",
					Title:       "Welcome to Simple Kanban!",
					Description: "Click me to edit • Double-click columns to edit them • Use + buttons to add new items",
					ColumnID:    "todo",
					Priority:    "medium",
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				},
			},
		}
	}
	
	var boardData KanbanBoardData
	if err := json.Unmarshal(data, &boardData); err != nil {
		fmt.Printf("Error parsing JSON: %v\n", err)
		return nil
	}
	
	fmt.Println("✅ Loaded board data from JSON file")
	return &boardData
}

// saveBoardData persists the current board state to the JSON file
// It uses mutex locking to ensure thread-safe file writes
func saveBoardData(boardName string, board *KanbanBoardData) error {
	globalMutex.Lock()
	defer globalMutex.Unlock()
	
	// Ensure boards directory exists
	if err := os.MkdirAll(BOARDS_DIR, 0755); err != nil {
		return fmt.Errorf("error creating boards directory: %v", err)
	}
	
	filePath := filepath.Join(BOARDS_DIR, boardName+".json")
	data, err := json.MarshalIndent(board, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling data: %v", err)
	}
	
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("error writing file: %v", err)
	}
	
	fmt.Printf("💾 Board data saved to %s\n", filePath)
	return nil
}

// getAvailableBoards returns a list of all available board JSON files
func getAvailableBoards() []string {
	// Ensure boards directory exists
	if err := os.MkdirAll(BOARDS_DIR, 0755); err != nil {
		fmt.Printf("Error creating boards directory: %v\n", err)
		return []string{"default"}
	}
	
	files, err := os.ReadDir(BOARDS_DIR)
	if err != nil {
		fmt.Printf("Error reading boards directory: %v\n", err)
		return []string{"default"}
	}
	
	var boards []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".json") {
			boardName := strings.TrimSuffix(file.Name(), ".json")
			boards = append(boards, boardName)
		}
	}
	
	if len(boards) == 0 {
		boards = append(boards, "default")
	}
	
	return boards
}

// createNewBoard creates a new empty board with the given name
func createNewBoard(boardName string) (*KanbanBoardData, error) {
	if boardName == "" {
		return nil, fmt.Errorf("board name cannot be empty")
	}
	
	// Create a new board with default columns
	newBoard := &KanbanBoardData{
		Columns: []KanbanColumn{
			{ID: "todo", Title: "To Do", Color: "#e3e8ef", Order: 0},
			{ID: "doing", Title: "In Progress", Color: "#ffd4a3", Order: 1},
			{ID: "done", Title: "Done", Color: "#a3e4d7", Order: 2},
		},
		Cards: []KanbanCard{},
	}
	
	// Save the new board
	if err := saveBoardData(boardName, newBoard); err != nil {
		return nil, err
	}
	
	return newBoard, nil
}

// registerBoard adds a new board instance to the active boards list
// This is called when a new WebSocket connection is established
func registerBoard(board *SimpleKanbanModal) {
	activeMutex.Lock()
	defer activeMutex.Unlock()
	activeBoards = append(activeBoards, board)
	fmt.Printf("📝 Registered board, total active: %d\n", len(activeBoards))
}

// unregisterBoard removes a board instance from the active boards list
// This is called when a WebSocket connection is closed
func unregisterBoard(board *SimpleKanbanModal) {
	activeMutex.Lock()
	defer activeMutex.Unlock()
	for i, b := range activeBoards {
		if b == board {
			activeBoards = append(activeBoards[:i], activeBoards[i+1:]...)
			break
		}
	}
	fmt.Printf("📝 Unregistered board, total active: %d\n", len(activeBoards))
}

// broadcastUpdate sends the current global board state to all connected clients
// This ensures real-time synchronization across all users viewing the board
// It includes panic recovery to handle closed WebSocket connections gracefully
func broadcastUpdate() {
	activeMutex.Lock()
	defer activeMutex.Unlock()
	
	globalMutex.RLock()
	defer globalMutex.RUnlock()
	
	if globalBoards == nil {
		return
	}
	
	fmt.Printf("📡 Broadcasting update to %d active boards\n", len(activeBoards))
	
	for _, board := range activeBoards {
		if board != nil && board.ComponentDriver != nil {
			currentBoard := globalBoards[board.CurrentBoard]
			if currentBoard == nil {
				continue
			}
			
			// Update board data from global state
			board.Columns = make([]KanbanColumn, len(currentBoard.Columns))
			copy(board.Columns, currentBoard.Columns)
			board.Cards = make([]KanbanCard, len(currentBoard.Cards))
			copy(board.Cards, currentBoard.Cards)
			
			// Trigger UI update with panic recovery
			func() {
				defer func() {
					if r := recover(); r != nil {
						fmt.Printf("Recovering from panic during broadcast: %v\n", r)
					}
				}()
				board.Commit()
				
				// Re-initialize drag & drop after DOM update
				board.initializeColumnDragDrop()
			}()
		}
	}
}

// updateGlobalState updates the global board state and triggers synchronization
// This function is called whenever any user makes changes to the board
// It saves to JSON asynchronously and broadcasts updates to all connected clients
func (k *SimpleKanbanModal) updateGlobalState(columns []KanbanColumn, cards []KanbanCard) {
	globalMutex.Lock()
	if globalBoards == nil {
		globalBoards = make(map[string]*KanbanBoardData)
	}
	if globalBoards[k.CurrentBoard] == nil {
		globalBoards[k.CurrentBoard] = &KanbanBoardData{}
	}
	globalBoards[k.CurrentBoard].Columns = make([]KanbanColumn, len(columns))
	copy(globalBoards[k.CurrentBoard].Columns, columns)
	globalBoards[k.CurrentBoard].Cards = make([]KanbanCard, len(cards))
	copy(globalBoards[k.CurrentBoard].Cards, cards)
	globalMutex.Unlock()
	
	// Save to JSON
	go func() {
		if err := saveBoardData(k.CurrentBoard, globalBoards[k.CurrentBoard]); err != nil {
			fmt.Printf("Error saving board data: %v\n", err)
		}
	}()
	
	// Broadcast to all active boards
	broadcastUpdate()
}

// NewSimpleKanbanModal creates and initializes a new Kanban board instance
// It loads the board state from the global shared state or creates a default board
// Each instance represents a single user's connection to the board
func NewSimpleKanbanModal() *SimpleKanbanModal {
	board := &SimpleKanbanModal{
		ShowModal:    false,
		ShowAlert:    true,  // Show instructions alert initially
		CurrentBoard: "default",
	}
	
	board.ComponentDriver = liveview.NewDriver[*SimpleKanbanModal]("kanban_board", board)
	
	// Get list of available boards
	board.BoardsList = getAvailableBoards()
	
	// Initialize global state if this is the first board
	globalMutex.Lock()
	if globalBoards == nil {
		globalBoards = make(map[string]*KanbanBoardData)
	}
	
	// Load default board if not loaded
	if globalBoards[board.CurrentBoard] == nil {
		boardData := loadBoardData(board.CurrentBoard)
		if boardData != nil {
			globalBoards[board.CurrentBoard] = boardData
		}
	}
	
	// Load data from global state
	if globalBoards[board.CurrentBoard] != nil {
		currentBoard := globalBoards[board.CurrentBoard]
		board.Columns = make([]KanbanColumn, len(currentBoard.Columns))
		copy(board.Columns, currentBoard.Columns)
		board.Cards = make([]KanbanCard, len(currentBoard.Cards))
		copy(board.Cards, currentBoard.Cards)
	}
	globalMutex.Unlock()
	
	// Create dropdown component for board selection
	var dropdownOptions []components.DropdownOption
	for _, boardName := range board.BoardsList {
		dropdownOptions = append(dropdownOptions, components.DropdownOption{
			Value: boardName,
			Label: strings.Title(strings.ReplaceAll(boardName, "_", " ")),
		})
	}
	board.Dropdown = components.NewDropdown(dropdownOptions, "Select Board")
	board.Dropdown.Selected = board.CurrentBoard
	board.Dropdown.ComponentDriver = liveview.NewDriver("board_dropdown", board.Dropdown)
	
	// Create file upload component for attachments
	board.FileUpload = &components.FileUpload{
		Multiple: true,
		Accept:   "image/*,.pdf,.doc,.docx,.txt,.json,.csv",
		MaxSize:  5 * 1024 * 1024, // 5MB
		MaxFiles: 5,
		Label:    "Drop files here or click to upload",
	}
	// Create the driver with proper type
	fileUploadDriver := liveview.NewDriver[*components.FileUpload]("file_upload", board.FileUpload)
	board.FileUpload.ComponentDriver = fileUploadDriver
	board.FileUpload.OnUpload = func(files []components.FileInfo) error {
		return board.handleFileUpload(files)
	}
	fmt.Printf("📁 FileUpload component created: %v with driver: %v\n", board.FileUpload != nil, board.FileUpload.ComponentDriver != nil)
	
	// Register this board for broadcasting
	registerBoard(board)
	
	return board
}

// GetDriver returns the LiveView component driver for WebSocket communication
// This is required by the LiveView framework for managing the component lifecycle
func (k *SimpleKanbanModal) GetDriver() liveview.LiveDriver {
	return k.ComponentDriver
}

// Start initializes the component and registers all event handlers
// This method is called when a WebSocket connection is established
// It sets up all the event handlers for user interactions
func (k *SimpleKanbanModal) Start() {
	k.SetID("content")
	
	// Register all event handlers explicitly
	if k.ComponentDriver != nil {
		k.ComponentDriver.Events["ReorderColumns"] = func(c *SimpleKanbanModal, data interface{}) {
			fmt.Println("📋 ReorderColumns event received via Events map")
			c.ReorderColumns(data)
		}
		k.ComponentDriver.Events["EditCard"] = func(c *SimpleKanbanModal, data interface{}) {
			c.EditCard(data)
		}
		k.ComponentDriver.Events["AddCard"] = func(c *SimpleKanbanModal, data interface{}) {
			c.AddCard(data)
		}
		k.ComponentDriver.Events["EditColumn"] = func(c *SimpleKanbanModal, data interface{}) {
			c.EditColumn(data)
		}
		k.ComponentDriver.Events["AddColumn"] = func(c *SimpleKanbanModal, data interface{}) {
			c.AddColumn(data)
		}
		k.ComponentDriver.Events["CloseModal"] = func(c *SimpleKanbanModal, data interface{}) {
			c.CloseModal(data)
		}
		k.ComponentDriver.Events["SaveModal"] = func(c *SimpleKanbanModal, data interface{}) {
			c.SaveModal(data)
		}
		k.ComponentDriver.Events["DeleteCard"] = func(c *SimpleKanbanModal, data interface{}) {
			c.DeleteCard(data)
		}
		k.ComponentDriver.Events["DeleteColumn"] = func(c *SimpleKanbanModal, data interface{}) {
			c.DeleteColumn(data)
		}
		k.ComponentDriver.Events["UpdateFormField"] = func(c *SimpleKanbanModal, data interface{}) {
			c.UpdateFormField(data)
		}
		k.ComponentDriver.Events["MoveCard"] = func(c *SimpleKanbanModal, data interface{}) {
			c.MoveCard(data)
		}
		k.ComponentDriver.Events["SwitchBoard"] = func(c *SimpleKanbanModal, data interface{}) {
			c.SwitchBoard(data)
		}
		k.ComponentDriver.Events["NewBoard"] = func(c *SimpleKanbanModal, data interface{}) {
			c.NewBoard(data)
		}
		k.ComponentDriver.Events["CreateBoard"] = func(c *SimpleKanbanModal, data interface{}) {
			c.CreateBoard(data)
		}
		k.ComponentDriver.Events["DismissAlert"] = func(c *SimpleKanbanModal, data interface{}) {
			c.DismissAlert(data)
		}
		k.ComponentDriver.Events["RefreshBoards"] = func(c *SimpleKanbanModal, data interface{}) {
			c.RefreshBoards(data)
		}
		k.ComponentDriver.Events["ArchiveBoard"] = func(c *SimpleKanbanModal, data interface{}) {
			c.ArchiveBoard(data)
		}
		k.ComponentDriver.Events["UploadFiles"] = func(c *SimpleKanbanModal, data interface{}) {
			c.UploadFiles(data)
		}
		k.ComponentDriver.Events["RemoveAttachment"] = func(c *SimpleKanbanModal, data interface{}) {
			c.RemoveAttachment(data)
		}
		k.ComponentDriver.Events["RefreshAttachments"] = func(c *SimpleKanbanModal, data interface{}) {
			c.RefreshAttachments(data)
		}
	}
	
	// Mount the dropdown component
	if k.Dropdown != nil {
		k.Mount(k.Dropdown)
	}
	
	// Mount the file upload component
	if k.FileUpload != nil {
		k.Mount(k.FileUpload)
		fmt.Printf("✅ FileUpload component mounted successfully - Driver: %v\n", k.FileUpload.ComponentDriver != nil)
	} else {
		fmt.Println("⚠️ FileUpload component is nil")
	}
	
	k.Commit()
	
	// Inject the upload and download functions globally
	uploadScript := `
	if (typeof window.downloadFile === 'undefined') {
		window.downloadFile = function(boardID, cardID, filename) {
			var downloadUrl = '/api/download/' + boardID + '/' + cardID + '/' + filename;
			
			// Create a temporary link element and trigger download
			var link = document.createElement('a');
			link.href = downloadUrl;
			link.download = filename;
			document.body.appendChild(link);
			link.click();
			document.body.removeChild(link);
		};
		console.log('[Kanban] Download function injected');
	}
	
	if (typeof window.uploadFiles === 'undefined') {
		window.uploadFiles = function(files, boardID, cardID) {
			if (!cardID || cardID === '') {
				alert('Please save the card first before adding attachments');
				return;
			}
			
			var formData = new FormData();
			var validFiles = 0;
			
			for (var i = 0; i < files.length; i++) {
				if (files[i].size > 5 * 1024 * 1024) {
					alert('File ' + files[i].name + ' is too large (max 5MB)');
					continue;
				}
				formData.append('files', files[i]);
				validFiles++;
			}
			
			if (validFiles === 0) {
				return;
			}
			
			// Show progress if element exists
			var progressEl = document.getElementById('upload-progress');
			var statusEl = document.getElementById('upload-status');
			if (progressEl) {
				progressEl.style.display = 'block';
				if (statusEl) {
					statusEl.textContent = 'Uploading ' + validFiles + ' file(s)...';
				}
			}
			
			// Upload via AJAX
			fetch('/api/upload/' + boardID + '/' + cardID, {
				method: 'POST',
				body: formData
			})
			.then(function(response) { 
				return response.json(); 
			})
			.then(function(data) {
				if (progressEl) {
					progressEl.style.display = 'none';
				}
				if (data.success) {
					// Notify via WebSocket to refresh attachments
					if (typeof send_event === 'function') {
						send_event('kanban_board', 'RefreshAttachments', JSON.stringify({
							cardID: cardID,
							files: data.files
						}));
					}
					alert(data.message);
				} else {
					alert('Upload failed: ' + (data.error || 'Unknown error'));
				}
			})
			.catch(function(error) {
				if (progressEl) {
					progressEl.style.display = 'none';
				}
				alert('Upload error: ' + error.message);
			});
		};
		console.log('[Kanban] Upload function injected');
	}
	`
	k.ComponentDriver.EvalScript(uploadScript)
	
	fmt.Println("Simple Kanban Modal Board initialized with explicit event registration")
	
	// Initialize column drag & drop via EvalScript
	k.initializeColumnDragDrop()
	
	// Auto-dismiss alert after 10 seconds
	if k.ShowAlert {
		k.EvalScript(`
			setTimeout(function() {
				const alert = document.getElementById('instructions-alert');
				if (alert) {
					send_event('` + k.IdComponent + `', 'DismissAlert', '');
				}
			}, 10000);
		`)
	}
}

// initializeColumnDragDrop sets up drag & drop for columns via JavaScript
func (k *SimpleKanbanModal) initializeColumnDragDrop() {
	script := `
	(function() {
		const container = document.getElementById('columns-container');
		if (!container) {
			console.log('[DRAG] No columns-container found');
			return;
		}
		
		// Get component ID from a data attribute
		const componentId = container.getAttribute('data-component-id') || 'kanban_board';
		
		let draggedColumn = null;
		let draggedIndex = -1;
		
		console.log('[DRAG] Initializing column drag & drop with component ID:', componentId);
		
		// Add event listeners to column headers
		function initColumnDragDrop() {
			const headers = container.querySelectorAll('.column-header[draggable="true"]');
			console.log('[DRAG] Found', headers.length, 'draggable column headers');
			
			// Remove existing listeners first to avoid duplicates
			headers.forEach((header) => {
				const newHeader = header.cloneNode(true);
				header.parentNode.replaceChild(newHeader, header);
			});
			
			// Now get the fresh headers and add listeners
			const freshHeaders = container.querySelectorAll('.column-header[draggable="true"]');
			
			freshHeaders.forEach((header, index) => {
				// Get the parent column element
				const column = header.parentElement;
				const columnIndex = parseInt(column.dataset.columnIndex);
				
				header.addEventListener('dragstart', function(e) {
					console.log('[DRAG] Drag started on column', columnIndex);
					draggedColumn = column;
					draggedIndex = columnIndex;
					column.classList.add('dragging');
					e.dataTransfer.effectAllowed = 'move';
					e.dataTransfer.setData('text/html', column.outerHTML);
					
					// Stop propagation to prevent card dragging
					e.stopPropagation();
				});
				
				header.addEventListener('dragend', function(e) {
					console.log('[DRAG] Drag ended');
					column.classList.remove('dragging');
					// Remove drag over effects from all columns
					container.querySelectorAll('.column').forEach(col => col.classList.remove('drag-over-column'));
				});
			});
			
			// Add dragover and drop events to all columns (not just headers)
			const columns = container.querySelectorAll('.column');
			columns.forEach((column) => {
				column.addEventListener('dragover', function(e) {
					e.preventDefault();
					if (this !== draggedColumn) {
						this.classList.add('drag-over-column');
					}
				});
				
				column.addEventListener('dragleave', function(e) {
					// Only remove if we're actually leaving the column
					if (!this.contains(e.relatedTarget)) {
						this.classList.remove('drag-over-column');
					}
				});
				
				column.addEventListener('drop', function(e) {
					e.preventDefault();
					this.classList.remove('drag-over-column');
					
					if (this !== draggedColumn && draggedColumn) {
						const targetIndex = parseInt(this.dataset.columnIndex);
						console.log('[DRAG] Dropping column', draggedIndex, 'on', targetIndex);
						
						// Send reorder event
						console.log('[DRAG] Sending reorder event to component:', componentId);
						if (typeof send_event === 'function') {
							send_event(componentId, 'ReorderColumns', JSON.stringify({
								sourceIndex: draggedIndex,
								targetIndex: targetIndex
							}));
						} else {
							console.error('[DRAG] send_event function not found!');
						}
					}
				});
			});
		}
		
		// Initialize immediately
		initColumnDragDrop();
		
		console.log('[DRAG] Column drag & drop initialized successfully');
	})();
	`
	
	// Execute the script to initialize drag & drop
	k.ComponentDriver.EvalScript(script)
}

// GetTemplate returns the template with modals
func (k *SimpleKanbanModal) GetTemplate() string {
	return `
	<style>
		body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; margin: 0; padding: 20px; background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); min-height: 100vh; }
		.kanban-board { max-width: 1200px; margin: 0 auto; }
		.board-header { text-align: center; color: white; margin-bottom: 20px; }
		.alert { background: rgba(255,255,255,0.95); color: #2c3e50; padding: 15px 20px; border-radius: 8px; margin-bottom: 20px; display: flex; justify-content: space-between; align-items: center; box-shadow: 0 3px 10px rgba(0,0,0,0.1); animation: slideDown 0.3s ease-out; }
		@keyframes slideDown { from { transform: translateY(-20px); opacity: 0; } to { transform: translateY(0); opacity: 1; } }
		.alert-close { background: none; border: none; font-size: 24px; cursor: pointer; color: #7f8c8d; padding: 0 5px; transition: color 0.2s; }
		.alert-close:hover { color: #34495e; }
		.columns-container { display: flex; gap: 20px; overflow-x: auto; padding: 20px 0; }
		.column { background: rgba(255,255,255,0.95); border-radius: 10px; min-width: 280px; box-shadow: 0 5px 15px rgba(0,0,0,0.1); transition: all 0.3s; }
		.column.dragging { opacity: 0.5; transform: rotate(5deg); }
		.column.drag-over-column { transform: scale(1.05); box-shadow: 0 8px 25px rgba(52, 152, 219, 0.3); }
		.column-header { padding: 15px 20px; border-radius: 10px 10px 0 0; cursor: grab; display: flex; justify-content: space-between; align-items: center; }
		.column.dragging .column-header { cursor: grabbing; }
		.column-header:hover { opacity: 0.9; }
		.column-cards { padding: 15px; min-height: 300px; }
		.kanban-card { background: white; padding: 15px; margin-bottom: 15px; border-radius: 8px; box-shadow: 0 2px 8px rgba(0,0,0,0.1); cursor: pointer; transition: all 0.2s; }
		.kanban-card:hover { box-shadow: 0 4px 15px rgba(0,0,0,0.15); transform: translateY(-2px); }
		.kanban-card[draggable="true"] { cursor: move; }
		.card-title { font-weight: 600; margin-bottom: 8px; color: #2c3e50; }
		.card-desc { color: #7f8c8d; font-size: 0.9em; line-height: 1.4; }
		.card-priority { display: inline-block; padding: 2px 8px; border-radius: 12px; font-size: 0.8em; margin-top: 5px; }
		.card-points-badge { position: absolute; bottom: 12px; right: 12px; background: #3498db; color: white; padding: 5px 12px; border-radius: 16px; font-size: 0.9em; font-weight: bold; box-shadow: 0 2px 6px rgba(0,0,0,0.15); z-index: 1; }
		.priority-low { background: #d5dbdb; color: #2c3e50; }
		.priority-medium { background: #f39c12; color: white; }
		.priority-high { background: #e74c3c; color: white; }
		.priority-urgent { background: #8e44ad; color: white; }
		.add-btn { width: 100%; padding: 12px; border: 2px dashed #bdc3c7; background: transparent; border-radius: 5px; cursor: pointer; color: #7f8c8d; transition: all 0.2s; }
		.add-btn:hover { border-color: #3498db; color: #3498db; background: rgba(52, 152, 219, 0.05); }
		.add-column-btn { min-width: 250px; display: flex; align-items: center; justify-content: center; }
		.drag-over { background: rgba(52, 152, 219, 0.1); border-radius: 5px; }
	</style>
	
	<div id="kanban_board" class="kanban-board">
		{{if .ShowAlert}}
		<div class="alert" id="instructions-alert">
			<span style="display: flex; align-items: center; gap: 10px;">
				<span style="font-size: 1.5em;">👉</span>
				<strong>Tips:</strong> Drag cards between columns to organize tasks • Click cards or columns to edit them • Drag column headers to reorder
			</span>
			<button class="alert-close" onclick="send_event('{{.IdComponent}}', 'DismissAlert', '')" title="Dismiss">×</button>
		</div>
		{{end}}
		
		<div class="board-header">
			<div style="display: flex; justify-content: center; align-items: center; gap: 15px;">
				<div style="display: flex; align-items: center; gap: 10px; background: rgba(255,255,255,0.15); padding: 10px 20px; border-radius: 8px; backdrop-filter: blur(10px);">
					<label style="color: white; font-weight: 600;">Board:</label>
					<select id="board-selector" 
							style="padding: 8px 15px; border-radius: 5px; border: none; background: white; cursor: pointer; min-width: 200px;"
							onchange="send_event('{{.IdComponent}}', 'SwitchBoard', this.value)">
						{{range .BoardsList}}
						<option value="{{.}}" {{if eq . $.CurrentBoard}}selected{{end}}>{{.}}</option>
						{{end}}
					</select>
					<button onclick="send_event('{{.IdComponent}}', 'RefreshBoards', '')" 
							style="padding: 8px 12px; background: #3498db; color: white; border: none; border-radius: 5px; cursor: pointer; transition: all 0.2s;"
							title="Refresh board list"
							onmouseover="this.style.background='#5dade2'"
							onmouseout="this.style.background='#3498db'">
						🔄
					</button>
					<button 
						{{if eq $.CurrentBoard "default"}}
							disabled
							style="padding: 8px 12px; background: #95a5a6; color: white; border: none; border-radius: 5px; cursor: not-allowed; opacity: 0.5; transition: all 0.2s;"
							title="Cannot archive default board"
						{{else}}
							onclick="if(confirm('Are you sure you want to archive this board? It will be moved to the archive folder.')) { send_event('{{$.IdComponent}}', 'ArchiveBoard', ''); }" 
							style="padding: 8px 12px; background: #e74c3c; color: white; border: none; border-radius: 5px; cursor: pointer; transition: all 0.2s;"
							title="Archive this board"
							onmouseover="this.style.background='#c0392b'"
							onmouseout="this.style.background='#e74c3c'"
						{{end}}>
						🗑️
					</button>
				</div>
				<button onclick="send_event('{{.IdComponent}}', 'NewBoard', '')" 
						style="padding: 10px 20px; background: #27ae60; color: white; border: none; border-radius: 8px; cursor: pointer; font-weight: 600; transition: all 0.2s; box-shadow: 0 2px 5px rgba(0,0,0,0.2);"
						onmouseover="this.style.background='#2ecc71'; this.style.transform='translateY(-2px)'; this.style.boxShadow='0 4px 8px rgba(0,0,0,0.3)'"
						onmouseout="this.style.background='#27ae60'; this.style.transform='translateY(0)'; this.style.boxShadow='0 2px 5px rgba(0,0,0,0.2)'">
					➕ New Board
				</button>
			</div>
		</div>
		
		<div class="columns-container" id="columns-container" data-component-id="{{.IdComponent}}">
			{{range $index, $col := .GetOrderedColumns}}
			<div class="column" 
				 data-column-index="{{$index}}"
				 data-column-id="{{.ID}}">
				<div class="column-header" 
					 draggable="true"
					 style="background: {{.Color}};"
					 ondblclick="send_event('{{$.IdComponent}}', 'EditColumn', '{{.ID}}')"
					 title="Double-click to edit • Drag to reorder">
					<span>{{.Title}}</span>
					<span style="font-size: 0.9em;">({{$.GetCardCount .ID}} cards | {{$.GetColumnPoints .ID}} pts)</span>
				</div>
				<div class="column-cards" 
					 data-column-id="{{.ID}}"
					 ondrop="event.preventDefault(); var cardId = event.dataTransfer.getData('cardId'); if(cardId) { send_event('{{$.IdComponent}}', 'MoveCard', JSON.stringify({cardId: cardId, columnId: '{{.ID}}'})); }"
					 ondragover="event.preventDefault(); this.classList.add('drag-over');"
					 ondragleave="this.classList.remove('drag-over');">
					
					{{range $.GetCardsForColumn .ID}}
					<div class="kanban-card" style="position: relative;"
						 draggable="true"
						 data-card-id="{{.ID}}"
						 ondragstart="event.dataTransfer.setData('cardId', '{{.ID}}'); this.classList.add('dragging');"
						 ondragend="this.classList.remove('dragging');"
						 onclick="send_event('{{$.IdComponent}}', 'EditCard', '{{.ID}}')">
						{{if gt .Points 0}}
						<div class="card-points-badge">{{.Points}} pts</div>
						{{end}}
						<div class="card-title">{{.Title}}</div>
						<div class="card-desc">{{.Description}}</div>
						<div style="display: flex; align-items: center; gap: 10px; margin-top: 8px;">
							{{if .Priority}}<span class="card-priority priority-{{.Priority}}">{{.Priority}}</span>{{end}}
							{{if .Attachments}}
							<span style="display: inline-flex; align-items: center; gap: 4px; color: #7f8c8d; font-size: 12px;">
								📎 {{len .Attachments}}
							</span>
							{{end}}
						</div>
					</div>
					{{end}}
					
					<button class="add-btn" onclick="send_event('{{$.IdComponent}}', 'AddCard', '{{.ID}}')">
						+ Add Card
					</button>
				</div>
			</div>
			{{end}}
			
			<div class="add-column-btn">
				<button class="add-btn" onclick="send_event('{{$.IdComponent}}', 'AddColumn', '')" style="padding: 15px 30px; font-size: 16px;">
					+ Add Column
				</button>
			</div>
		</div>
	</div>
	
	{{if .ShowModal}}
	<div style="position: fixed; top: 0; left: 0; right: 0; bottom: 0; background: rgba(0,0,0,0.6); z-index: 9999; display: flex; align-items: center; justify-content: center;">
		<div style="background: white; border-radius: 15px; padding: 30px; min-width: 450px; max-width: 600px; box-shadow: 0 20px 60px rgba(0,0,0,0.3);">
			<div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 25px; padding-bottom: 15px; border-bottom: 1px solid #ecf0f1;">
				<h2 style="margin: 0; color: #2c3e50;">{{.ModalTitle}}</h2>
				<button onclick="send_event('{{.IdComponent}}', 'CloseModal', '')" 
						style="background: none; border: none; font-size: 24px; cursor: pointer; color: #95a5a6; padding: 5px;">&times;</button>
			</div>
			
			{{if or (eq .ModalType "edit_card") (eq .ModalType "add_card")}}
			<div style="display: flex; flex-direction: column; gap: 20px;">
				<div>
					<label style="display: block; margin-bottom: 8px; font-weight: 500; color: #34495e;">Title</label>
					<input type="text" value="{{.FormCardTitle}}" 
						   oninput="send_event('{{.IdComponent}}', 'UpdateFormField', JSON.stringify({field: 'card_title', value: this.value}))"
						   style="width: 100%; padding: 12px; border: 1px solid #bdc3c7; border-radius: 6px; font-size: 14px;">
				</div>
				
				<div>
					<label style="display: block; margin-bottom: 8px; font-weight: 500; color: #34495e;">Description</label>
					<textarea oninput="send_event('{{.IdComponent}}', 'UpdateFormField', JSON.stringify({field: 'card_desc', value: this.value}))"
							  style="width: 100%; padding: 12px; border: 1px solid #bdc3c7; border-radius: 6px; font-size: 14px; min-height: 100px; resize: vertical;">{{.FormCardDesc}}</textarea>
				</div>
				
				<div>
					<label style="display: block; margin-bottom: 8px; font-weight: 500; color: #34495e;">Priority</label>
					<select onchange="send_event('{{.IdComponent}}', 'UpdateFormField', JSON.stringify({field: 'card_priority', value: this.value}))"
							style="width: 100%; padding: 12px; border: 1px solid #bdc3c7; border-radius: 6px; font-size: 14px;">
						<option value="low" {{if eq .FormCardPriority "low"}}selected{{end}}>Low</option>
						<option value="medium" {{if eq .FormCardPriority "medium"}}selected{{end}}>Medium</option>
						<option value="high" {{if eq .FormCardPriority "high"}}selected{{end}}>High</option>
						<option value="urgent" {{if eq .FormCardPriority "urgent"}}selected{{end}}>Urgent</option>
					</select>
				</div>
				
				<div>
					<label style="display: block; margin-bottom: 8px; font-weight: 500; color: #34495e;">Points</label>
					<input type="number" value="{{.FormCardPoints}}" min="0" max="100"
						   oninput="send_event('{{.IdComponent}}', 'UpdateFormField', JSON.stringify({field: 'card_points', value: parseInt(this.value) || 0}))"
						   style="width: 100%; padding: 12px; border: 1px solid #bdc3c7; border-radius: 6px; font-size: 14px;">
				</div>
				
				{{if eq .ModalType "edit_card"}}
				<div>
					<label style="display: block; margin-bottom: 8px; font-weight: 500; color: #34495e;">Column</label>
					<select onchange="send_event('{{.IdComponent}}', 'UpdateFormField', JSON.stringify({field: 'card_column', value: this.value}))"
							style="width: 100%; padding: 12px; border: 1px solid #bdc3c7; border-radius: 6px; font-size: 14px;">
						{{range .GetOrderedColumns}}
						<option value="{{.ID}}" {{if eq $.FormCardColumn .ID}}selected{{end}}>{{.Title}}</option>
						{{end}}
					</select>
				</div>
				{{end}}
				
				<!-- Attachments Section -->
				<div>
					<label style="display: block; margin-bottom: 8px; font-weight: 500; color: #34495e;">Attachments</label>
					
					<!-- Display existing attachments -->
					{{if .FormCardAttachments}}
					<div style="background: #f8f9fa; border-radius: 6px; padding: 10px; margin-bottom: 10px;">
						{{range $i, $attachment := .FormCardAttachments}}
						<div style="display: flex; align-items: center; justify-content: space-between; padding: 8px; background: white; border-radius: 4px; margin-bottom: 6px;">
							<div style="display: flex; align-items: center; gap: 10px; flex: 1;">
								<span style="font-size: 18px;">📎</span>
								<div style="flex: 1;">
									<div style="font-weight: 500; color: #2c3e50;">{{$attachment.Name}}</div>
									<div style="font-size: 12px; color: #95a5a6;">{{$.FormatFileSize $attachment.Size}}</div>
								</div>
							</div>
							<div style="display: flex; gap: 5px;">
								<button onclick="downloadFile('{{$.CurrentBoard}}', '{{$.FormCardID}}', '{{$attachment.ID}}_{{$attachment.Name}}')" 
										style="background: #3498db; color: white; border: none; border-radius: 4px; padding: 4px 8px; cursor: pointer; font-size: 12px;">
									📥 Download
								</button>
								<button onclick="send_event('{{$.IdComponent}}', 'RemoveAttachment', '{{$attachment.ID}}')" 
										style="background: #e74c3c; color: white; border: none; border-radius: 4px; padding: 4px 8px; cursor: pointer; font-size: 12px;">
									🗑️ Remove
								</button>
							</div>
						</div>
						{{end}}
					</div>
					{{end}}
					
					<!-- File Upload Area -->
					<div style="margin-top: 10px;">
						<div id="upload-progress" style="display: none; margin-bottom: 10px;">
							<div style="background: #e3f2fd; border-radius: 4px; padding: 10px; color: #1976d2;">
								<span id="upload-status">Uploading files...</span>
							</div>
						</div>
						<div class="file-drop-zone" style="border: 2px dashed #bdc3c7; border-radius: 8px; padding: 20px; text-align: center; background: #fafafa; cursor: pointer; transition: all 0.3s;"
							 ondragover="event.preventDefault(); this.style.borderColor='#3498db'; this.style.background='#ecf7ff';"
							 ondragleave="event.preventDefault(); this.style.borderColor='#bdc3c7'; this.style.background='#fafafa';"
							 ondrop="event.preventDefault(); this.style.borderColor='#bdc3c7'; this.style.background='#fafafa';
								var files = event.dataTransfer.files;
								uploadFiles(files, '{{.CurrentBoard}}', '{{.FormCardID}}');"
							 onclick="document.getElementById('file-input-{{.IdComponent}}').click()">
							<div style="font-size: 32px; color: #95a5a6; margin-bottom: 10px;">📁</div>
							<div style="color: #7f8c8d; font-weight: 500;">Drop files here or click to browse</div>
							<div style="color: #95a5a6; font-size: 12px; margin-top: 5px;">Max 5 files, 5MB each • Images, PDFs, Documents</div>
						</div>
						<input type="file" id="file-input-{{.IdComponent}}" multiple 
							   accept="image/*,.pdf,.doc,.docx,.txt,.json,.csv,.xml,.zip"
							   style="display: none;"
							   onchange="uploadFiles(this.files, '{{.CurrentBoard}}', '{{.FormCardID}}')">
						
					</div>
				</div>
			</div>
			{{end}}
			
			{{if or (eq .ModalType "edit_column") (eq .ModalType "add_column")}}
			<div style="display: flex; flex-direction: column; gap: 20px;">
				<div>
					<label style="display: block; margin-bottom: 8px; font-weight: 500; color: #34495e;">Column Name</label>
					<input type="text" value="{{.FormColumnTitle}}" 
						   oninput="send_event('{{.IdComponent}}', 'UpdateFormField', JSON.stringify({field: 'column_title', value: this.value}))"
						   style="width: 100%; padding: 12px; border: 1px solid #bdc3c7; border-radius: 6px; font-size: 14px;">
				</div>
				
				<div>
					<label style="display: block; margin-bottom: 8px; font-weight: 500; color: #34495e;">Color</label>
					<div style="display: flex; gap: 10px; flex-wrap: wrap;">
						<div onclick="send_event('{{.IdComponent}}', 'UpdateFormField', JSON.stringify({field: 'column_color', value: '#e3e8ef'}))"
							 style="width: 50px; height: 40px; background: #e3e8ef; border-radius: 6px; cursor: pointer; {{if eq .FormColumnColor "#e3e8ef"}}box-shadow: 0 0 0 3px #3498db;{{end}}"></div>
						<div onclick="send_event('{{.IdComponent}}', 'UpdateFormField', JSON.stringify({field: 'column_color', value: '#ffd4a3'}))"
							 style="width: 50px; height: 40px; background: #ffd4a3; border-radius: 6px; cursor: pointer; {{if eq .FormColumnColor "#ffd4a3"}}box-shadow: 0 0 0 3px #3498db;{{end}}"></div>
						<div onclick="send_event('{{.IdComponent}}', 'UpdateFormField', JSON.stringify({field: 'column_color', value: '#a3e4d7'}))"
							 style="width: 50px; height: 40px; background: #a3e4d7; border-radius: 6px; cursor: pointer; {{if eq .FormColumnColor "#a3e4d7"}}box-shadow: 0 0 0 3px #3498db;{{end}}"></div>
						<div onclick="send_event('{{.IdComponent}}', 'UpdateFormField', JSON.stringify({field: 'column_color', value: '#f8b3d0'}))"
							 style="width: 50px; height: 40px; background: #f8b3d0; border-radius: 6px; cursor: pointer; {{if eq .FormColumnColor "#f8b3d0"}}box-shadow: 0 0 0 3px #3498db;{{end}}"></div>
						<div onclick="send_event('{{.IdComponent}}', 'UpdateFormField', JSON.stringify({field: 'column_color', value: '#b3d4f8'}))"
							 style="width: 50px; height: 40px; background: #b3d4f8; border-radius: 6px; cursor: pointer; {{if eq .FormColumnColor "#b3d4f8"}}box-shadow: 0 0 0 3px #3498db;{{end}}"></div>
						<div onclick="send_event('{{.IdComponent}}', 'UpdateFormField', JSON.stringify({field: 'column_color', value: '#d4b3f8'}))"
							 style="width: 50px; height: 40px; background: #d4b3f8; border-radius: 6px; cursor: pointer; {{if eq .FormColumnColor "#d4b3f8"}}box-shadow: 0 0 0 3px #3498db;{{end}}"></div>
					</div>
				</div>
			</div>
			{{end}}
			
			{{if eq .ModalType "new_board"}}
			<div style="display: flex; flex-direction: column; gap: 20px;">
				<div>
					<label style="display: block; margin-bottom: 8px; font-weight: 500; color: #34495e;">Board Name</label>
					<input type="text" value="{{.FormBoardName}}" 
						   placeholder="Enter board name"
						   oninput="send_event('{{.IdComponent}}', 'UpdateFormField', JSON.stringify({field: 'board_name', value: this.value}))"
						   style="width: 100%; padding: 12px; border: 1px solid #bdc3c7; border-radius: 6px; font-size: 14px;">
				</div>
			</div>
			{{end}}
			
			<div style="display: flex; justify-content: flex-end; gap: 15px; margin-top: 30px; padding-top: 20px; border-top: 1px solid #ecf0f1;">
				{{if eq .ModalType "edit_column"}}
				<button onclick="send_event('{{.IdComponent}}', 'DeleteColumn', '')" 
						style="background: #e74c3c; color: white; padding: 12px 25px; border: none; border-radius: 6px; cursor: pointer; margin-right: auto;">
					Delete Column
				</button>
				{{end}}
				
				{{if eq .ModalType "edit_card"}}
				<button onclick="send_event('{{.IdComponent}}', 'DeleteCard', '')" 
						style="background: #e74c3c; color: white; padding: 12px 25px; border: none; border-radius: 6px; cursor: pointer; margin-right: auto;">
					Delete Card
				</button>
				{{end}}
				
				<button onclick="send_event('{{.IdComponent}}', 'CloseModal', '')" 
						style="background: #95a5a6; color: white; padding: 12px 25px; border: none; border-radius: 6px; cursor: pointer;">
					Cancel
				</button>
				{{if eq .ModalType "new_board"}}
				<button onclick="send_event('{{.IdComponent}}', 'CreateBoard', '')" 
						style="background: #27ae60; color: white; padding: 12px 25px; border: none; border-radius: 6px; cursor: pointer;">
					Create Board
				</button>
				{{else}}
				<button onclick="send_event('{{.IdComponent}}', 'SaveModal', '')" 
						style="background: #3498db; color: white; padding: 12px 25px; border: none; border-radius: 6px; cursor: pointer;">
					{{if or (eq .ModalType "add_card") (eq .ModalType "add_column")}}Add{{else}}Save{{end}}
				</button>
				{{end}}
			</div>
		</div>
	</div>
	{{end}}
	`
}

// Helper functions for template rendering and data queries

// GetCardsForColumn returns all cards that belong to a specific column
// Used in the template to display cards within each column
func (k *SimpleKanbanModal) GetCardsForColumn(columnID string) []KanbanCard {
	var cards []KanbanCard
	for _, card := range k.Cards {
		if card.ColumnID == columnID {
			cards = append(cards, card)
		}
	}
	return cards
}

// GetCardCount returns the number of cards in a specific column
// Used in the template to display card count in column headers
func (k *SimpleKanbanModal) GetCardCount(columnID string) int {
	count := 0
	for _, card := range k.Cards {
		if card.ColumnID == columnID {
			count++
		}
	}
	return count
}

// GetColumnPoints calculates and returns the total story points for all cards in a column
// Used in the template to display total points in column headers
func (k *SimpleKanbanModal) GetColumnPoints(columnID string) int {
	total := 0
	for _, card := range k.Cards {
		if card.ColumnID == columnID {
			total += card.Points
		}
	}
	return total
}

// GetOrderedColumns returns a sorted copy of columns based on their Order field
// This ensures columns are displayed in the correct sequence in the UI
func (k *SimpleKanbanModal) GetOrderedColumns() []KanbanColumn {
	// Create a copy to avoid modifying original slice
	columns := make([]KanbanColumn, len(k.Columns))
	copy(columns, k.Columns)
	
	// Sort by Order field
	for i := 0; i < len(columns)-1; i++ {
		for j := 0; j < len(columns)-i-1; j++ {
			if columns[j].Order > columns[j+1].Order {
				columns[j], columns[j+1] = columns[j+1], columns[j]
			}
		}
	}
	return columns
}

// Event Handlers - These methods handle user interactions from the browser

// MoveCard handles dragging and dropping cards between columns
// It updates the card's column assignment and saves the changes
func (k *SimpleKanbanModal) MoveCard(data interface{}) {
	var event map[string]interface{}
	if jsonData, ok := data.(string); ok {
		json.Unmarshal([]byte(jsonData), &event)
		
		cardID := event["cardId"].(string)
		columnID := event["columnId"].(string)
		
		// Create updated cards slice
		updatedCards := make([]KanbanCard, len(k.Cards))
		copy(updatedCards, k.Cards)
		
		for i := range updatedCards {
			if updatedCards[i].ID == cardID {
				if updatedCards[i].ColumnID != columnID {
					fmt.Printf("Moving card %s to column %s\n", cardID, columnID)
					updatedCards[i].ColumnID = columnID
					updatedCards[i].UpdatedAt = time.Now()
				}
				break
			}
		}
		
		// Update global state and broadcast
		k.updateGlobalState(k.Columns, updatedCards)
	}
}

// EditCard opens the modal dialog for editing an existing card
// It populates the form fields with the card's current data
func (k *SimpleKanbanModal) EditCard(data interface{}) {
	cardID := ""
	if id, ok := data.(string); ok {
		cardID = id
	}
	
	fmt.Printf("🔧 EditCard called - FileUpload exists: %v\n", k.FileUpload != nil)
	
	for _, card := range k.Cards {
		if card.ID == cardID {
			k.ShowModal = true
			k.ModalType = "edit_card"
			k.ModalTitle = "Edit Card"
			k.FormCardID = card.ID
			k.FormCardTitle = card.Title
			k.FormCardDesc = card.Description
			k.FormCardColumn = card.ColumnID
			k.FormCardPriority = card.Priority
			k.FormCardPoints = card.Points
			k.FormCardAttachments = card.Attachments
			if k.FormCardPriority == "" {
				k.FormCardPriority = "medium"
			}
			// Reset file upload component
			if k.FileUpload != nil {
				k.FileUpload.Files = []components.FileInfo{}
				k.FileUpload.Commit()
				fmt.Println("📁 FileUpload component reset for edit modal")
			}
			break
		}
	}
	k.Commit()
}

// AddCard opens the modal dialog for creating a new card
// It initializes the form with default values and the target column
func (k *SimpleKanbanModal) AddCard(data interface{}) {
	columnID := ""
	if id, ok := data.(string); ok {
		columnID = id
	}
	
	k.ShowModal = true
	k.ModalType = "add_card"
	k.ModalTitle = "Add New Card"
	k.FormCardID = ""
	k.FormCardTitle = ""
	k.FormCardDesc = ""
	k.FormCardColumn = columnID
	k.FormCardPriority = "medium"
	k.FormCardPoints = 0
	k.FormCardAttachments = []Attachment{}
	// Reset file upload component
	if k.FileUpload != nil {
		k.FileUpload.Files = []components.FileInfo{}
		k.FileUpload.Commit()
	}
	k.Commit()
}

// EditColumn opens the modal dialog for editing a column's properties
// Users can change the column title and color
func (k *SimpleKanbanModal) EditColumn(data interface{}) {
	columnID := ""
	if id, ok := data.(string); ok {
		columnID = id
	}
	
	for _, col := range k.Columns {
		if col.ID == columnID {
			k.ShowModal = true
			k.ModalType = "edit_column"
			k.ModalTitle = "Edit Column"
			k.FormColumnID = col.ID
			k.FormColumnTitle = col.Title
			k.FormColumnColor = col.Color
			break
		}
	}
	k.Commit()
}

// AddColumn opens modal to add a new column
func (k *SimpleKanbanModal) AddColumn(data interface{}) {
	k.ShowModal = true
	k.ModalType = "add_column"
	k.ModalTitle = "Add New Column"
	k.FormColumnID = ""
	k.FormColumnTitle = ""
	k.FormColumnColor = "#e3e8ef"
	k.Commit()
}

// UpdateFormField updates a form field
func (k *SimpleKanbanModal) UpdateFormField(data interface{}) {
	var event map[string]interface{}
	if jsonData, ok := data.(string); ok {
		json.Unmarshal([]byte(jsonData), &event)
		
		fieldInterface, ok := event["field"]
		if !ok {
			return
		}
		
		field, ok := fieldInterface.(string)
		if !ok {
			return
		}
		
		switch field {
		case "card_title":
			if value, ok := event["value"].(string); ok {
				k.FormCardTitle = value
			}
		case "card_desc":
			if value, ok := event["value"].(string); ok {
				k.FormCardDesc = value
			}
		case "card_column":
			if value, ok := event["value"].(string); ok {
				k.FormCardColumn = value
			}
		case "card_priority":
			if value, ok := event["value"].(string); ok {
				k.FormCardPriority = value
			}
		case "card_points":
			// Handle points which comes as a number from JavaScript
			switch v := event["value"].(type) {
			case float64:
				k.FormCardPoints = int(v)
			case string:
				// Try to parse string as integer
				if points, err := strconv.Atoi(v); err == nil {
					k.FormCardPoints = points
				}
			}
		case "column_title":
			if value, ok := event["value"].(string); ok {
				k.FormColumnTitle = value
			}
		case "column_color":
			if value, ok := event["value"].(string); ok {
				k.FormColumnColor = value
			}
		case "board_name":
			if value, ok := event["value"].(string); ok {
				k.FormBoardName = value
			}
		}
	}
}

// CloseModal closes the modal dialog without saving changes
func (k *SimpleKanbanModal) CloseModal(data interface{}) {
	k.ShowModal = false
	k.Commit()
}

// SaveModal processes and saves the modal form data
// It handles creating new cards/columns or updating existing ones
// Changes are saved to the global state and broadcast to all users
func (k *SimpleKanbanModal) SaveModal(data interface{}) {
	// Create copies for modification
	updatedCards := make([]KanbanCard, len(k.Cards))
	copy(updatedCards, k.Cards)
	updatedColumns := make([]KanbanColumn, len(k.Columns))
	copy(updatedColumns, k.Columns)
	
	switch k.ModalType {
	case "edit_card":
		for i := range updatedCards {
			if updatedCards[i].ID == k.FormCardID {
				updatedCards[i].Title = k.FormCardTitle
				updatedCards[i].Description = k.FormCardDesc
				updatedCards[i].ColumnID = k.FormCardColumn
				updatedCards[i].Priority = k.FormCardPriority
				updatedCards[i].Points = k.FormCardPoints
				updatedCards[i].Attachments = k.FormCardAttachments
				updatedCards[i].UpdatedAt = time.Now()
				break
			}
		}
		
	case "add_card":
		newCard := KanbanCard{
			ID:          fmt.Sprintf("card_%d", time.Now().UnixNano()),
			Title:       k.FormCardTitle,
			Description: k.FormCardDesc,
			ColumnID:    k.FormCardColumn,
			Priority:    k.FormCardPriority,
			Points:      k.FormCardPoints,
			Attachments: k.FormCardAttachments,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		updatedCards = append(updatedCards, newCard)
		
	case "edit_column":
		for i := range updatedColumns {
			if updatedColumns[i].ID == k.FormColumnID {
				updatedColumns[i].Title = k.FormColumnTitle
				updatedColumns[i].Color = k.FormColumnColor
				break
			}
		}
		
	case "add_column":
		maxOrder := 0
		for _, col := range updatedColumns {
			if col.Order > maxOrder {
				maxOrder = col.Order
			}
		}
		
		newColumn := KanbanColumn{
			ID:    fmt.Sprintf("col_%d", time.Now().UnixNano()),
			Title: k.FormColumnTitle,
			Color: k.FormColumnColor,
			Order: maxOrder + 1,
		}
		updatedColumns = append(updatedColumns, newColumn)
	}
	
	k.ShowModal = false
	
	// Update global state and broadcast
	k.updateGlobalState(updatedColumns, updatedCards)
}

// DeleteCard deletes the current card
func (k *SimpleKanbanModal) DeleteCard(data interface{}) {
	newCards := []KanbanCard{}
	for _, card := range k.Cards {
		if card.ID != k.FormCardID {
			newCards = append(newCards, card)
		}
	}
	k.ShowModal = false
	
	// Update global state and broadcast
	k.updateGlobalState(k.Columns, newCards)
}

// DeleteColumn deletes the current column and its cards
func (k *SimpleKanbanModal) DeleteColumn(data interface{}) {
	// Delete all cards in this column
	newCards := []KanbanCard{}
	for _, card := range k.Cards {
		if card.ColumnID != k.FormColumnID {
			newCards = append(newCards, card)
		}
	}
	
	// Delete the column
	newColumns := []KanbanColumn{}
	for _, col := range k.Columns {
		if col.ID != k.FormColumnID {
			newColumns = append(newColumns, col)
		}
	}
	
	k.ShowModal = false
	
	// Update global state and broadcast
	k.updateGlobalState(newColumns, newCards)
}

// ReorderColumns handles column reordering via drag & drop
// It swaps the position of two columns when one is dropped on another
// The swap is done by exchanging the Order field values
func (k *SimpleKanbanModal) ReorderColumns(data interface{}) {
	var event map[string]interface{}
	if jsonData, ok := data.(string); ok {
		json.Unmarshal([]byte(jsonData), &event)
		
		sourceIndexFloat, ok1 := event["sourceIndex"].(float64)
		targetIndexFloat, ok2 := event["targetIndex"].(float64)
		
		if !ok1 || !ok2 {
			fmt.Printf("Invalid reorder data: sourceIndex=%v, targetIndex=%v\n", event["sourceIndex"], event["targetIndex"])
			return
		}
		
		sourceIndex := int(sourceIndexFloat)
		targetIndex := int(targetIndexFloat)
		
		// Get ordered columns first
		orderedColumns := k.GetOrderedColumns()
		
		if sourceIndex < 0 || sourceIndex >= len(orderedColumns) || 
		   targetIndex < 0 || targetIndex >= len(orderedColumns) ||
		   sourceIndex == targetIndex {
			fmt.Printf("Invalid column reorder indices: source=%d, target=%d, len=%d\n", sourceIndex, targetIndex, len(orderedColumns))
			return
		}
		
		fmt.Printf("Swapping columns: %s (index %d, order %d) with %s (index %d, order %d)\n", 
			orderedColumns[sourceIndex].Title, sourceIndex, orderedColumns[sourceIndex].Order,
			orderedColumns[targetIndex].Title, targetIndex, orderedColumns[targetIndex].Order)
		
		// Find the actual columns to swap in the original k.Columns slice
		sourceColID := orderedColumns[sourceIndex].ID
		targetColID := orderedColumns[targetIndex].ID
		
		// Create updated columns slice
		updatedColumns := make([]KanbanColumn, len(k.Columns))
		copy(updatedColumns, k.Columns)
		
		// Find and swap the Order values ONLY for these two specific columns
		var sourceIdx, targetIdx int = -1, -1
		for i := range updatedColumns {
			if updatedColumns[i].ID == sourceColID {
				sourceIdx = i
			}
			if updatedColumns[i].ID == targetColID {
				targetIdx = i
			}
		}
		
		if sourceIdx != -1 && targetIdx != -1 {
			// Swap ONLY the Order values of these two columns
			tempOrder := updatedColumns[sourceIdx].Order
			updatedColumns[sourceIdx].Order = updatedColumns[targetIdx].Order
			updatedColumns[targetIdx].Order = tempOrder
			
			fmt.Printf("✅ Swapped orders: %s (now order %d) <-> %s (now order %d)\n", 
				updatedColumns[sourceIdx].Title, updatedColumns[sourceIdx].Order,
				updatedColumns[targetIdx].Title, updatedColumns[targetIdx].Order)
		} else {
			fmt.Printf("❌ Could not find columns to swap\n")
			return
		}
		
		// Update global state and broadcast
		k.updateGlobalState(updatedColumns, k.Cards)
	}
}

// SwitchBoard switches the current board to a different one
func (k *SimpleKanbanModal) SwitchBoard(data interface{}) {
	if boardName, ok := data.(string); ok && boardName != "" {
		fmt.Printf("Switching to board: %s\n", boardName)
		
		// Update current board name
		k.CurrentBoard = boardName
		
		// Load the selected board
		globalMutex.Lock()
		if globalBoards[boardName] == nil {
			// Load board from file if not in memory
			boardData := loadBoardData(boardName)
			if boardData != nil {
				globalBoards[boardName] = boardData
			}
		}
		
		// Update local state with new board data
		if globalBoards[boardName] != nil {
			currentBoard := globalBoards[boardName]
			k.Columns = make([]KanbanColumn, len(currentBoard.Columns))
			copy(k.Columns, currentBoard.Columns)
			k.Cards = make([]KanbanCard, len(currentBoard.Cards))
			copy(k.Cards, currentBoard.Cards)
		}
		globalMutex.Unlock()
		
		// Update dropdown selection
		if k.Dropdown != nil {
			k.Dropdown.Selected = boardName
			k.Dropdown.Commit()
		}
		
		k.Commit()
		k.initializeColumnDragDrop()
	}
}

// NewBoard opens the modal to create a new board
func (k *SimpleKanbanModal) NewBoard(data interface{}) {
	k.ShowModal = true
	k.ModalType = "new_board"
	k.ModalTitle = "Create New Board"
	k.FormBoardName = ""
	k.Commit()
}

// CreateBoard creates a new board with the given name
func (k *SimpleKanbanModal) CreateBoard(data interface{}) {
	boardName := strings.TrimSpace(k.FormBoardName)
	if boardName == "" {
		fmt.Println("Board name cannot be empty")
		return
	}
	
	// Replace spaces with underscores for file naming
	boardName = strings.ReplaceAll(boardName, " ", "_")
	
	// Create new board
	newBoard, err := createNewBoard(boardName)
	if err != nil {
		fmt.Printf("Error creating new board: %v\n", err)
		return
	}
	
	// Add to global boards
	globalMutex.Lock()
	if globalBoards == nil {
		globalBoards = make(map[string]*KanbanBoardData)
	}
	globalBoards[boardName] = newBoard
	globalMutex.Unlock()
	
	// Update boards list
	k.BoardsList = getAvailableBoards()
	
	// Update dropdown options
	if k.Dropdown != nil {
		var dropdownOptions []components.DropdownOption
		for _, bn := range k.BoardsList {
			dropdownOptions = append(dropdownOptions, components.DropdownOption{
				Value: bn,
				Label: strings.Title(strings.ReplaceAll(bn, "_", " ")),
			})
		}
		k.Dropdown.Options = dropdownOptions
		k.Dropdown.Commit()
	}
	
	// Broadcast board list update to all connected clients
	broadcastBoardListUpdate()
	
	// Switch to the new board
	k.SwitchBoard(boardName)
	
	// Close modal
	k.ShowModal = false
	k.Commit()
}

// DismissAlert dismisses the instructions alert
func (k *SimpleKanbanModal) DismissAlert(data interface{}) {
	k.ShowAlert = false
	k.Commit()
}

// RefreshBoards refreshes the board list from disk
func (k *SimpleKanbanModal) RefreshBoards(data interface{}) {
	k.BoardsList = getAvailableBoards()
	k.Commit()
}

// ArchiveBoard moves the current board to the archived folder
func (k *SimpleKanbanModal) ArchiveBoard(data interface{}) {
	fmt.Printf("🗑️ ArchiveBoard called for board: %s\n", k.CurrentBoard)
	
	// Don't allow archiving if it's the last board
	if len(k.BoardsList) <= 1 {
		fmt.Println("Cannot archive the last board")
		return
	}
	
	// Don't allow archiving the default board
	if k.CurrentBoard == "default" {
		fmt.Println("Cannot archive the default board")
		return
	}
	
	// Archive the current board
	if err := archiveBoardFile(k.CurrentBoard); err != nil {
		fmt.Printf("Error archiving board: %v\n", err)
		return
	}
	
	// Remove from global boards
	globalMutex.Lock()
	delete(globalBoards, k.CurrentBoard)
	globalMutex.Unlock()
	
	// Get updated board list
	k.BoardsList = getAvailableBoards()
	
	// Switch to first available board
	if len(k.BoardsList) > 0 {
		k.SwitchBoard(k.BoardsList[0])
	}
	
	// Broadcast board list update to all clients
	broadcastBoardListUpdate()
}

// archiveBoardFile moves a board file to the archived folder
func archiveBoardFile(boardName string) error {
	// Ensure archive directory exists
	archiveDir := filepath.Join(BOARDS_DIR, "archived")
	if err := os.MkdirAll(archiveDir, 0755); err != nil {
		return fmt.Errorf("error creating archive directory: %v", err)
	}
	
	// Move the file
	sourcePath := filepath.Join(BOARDS_DIR, boardName+".json")
	destPath := filepath.Join(archiveDir, boardName+"_"+time.Now().Format("20060102_150405")+".json")
	
	if err := os.Rename(sourcePath, destPath); err != nil {
		return fmt.Errorf("error archiving board file: %v", err)
	}
	
	fmt.Printf("📦 Board '%s' archived to %s\n", boardName, destPath)
	return nil
}

// broadcastBoardListUpdate sends the updated board list to all connected clients
func broadcastBoardListUpdate() {
	activeMutex.Lock()
	defer activeMutex.Unlock()
	
	newBoardsList := getAvailableBoards()
	fmt.Printf("📡 Broadcasting board list update to %d active boards\n", len(activeBoards))
	
	for _, board := range activeBoards {
		if board != nil && board.ComponentDriver != nil {
			// Update board list
			board.BoardsList = newBoardsList
			
			// Trigger UI update with panic recovery
			func() {
				defer func() {
					if r := recover(); r != nil {
						fmt.Printf("Recovering from panic during board list broadcast: %v\n", r)
					}
				}()
				board.Commit()
			}()
		}
	}
}

// handleFileUpload processes uploaded files and saves them to disk
func (k *SimpleKanbanModal) handleFileUpload(files []components.FileInfo) error {
	for _, file := range files {
		// Generate unique file ID
		attachmentID := fmt.Sprintf("attach_%d", time.Now().UnixNano())
		
		// Create file path
		filePath := filepath.Join("attachments", k.CurrentBoard, k.FormCardID, attachmentID+"_"+file.Name)
		
		// Ensure directory exists
		dir := filepath.Dir(filePath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %v", err)
		}
		
		// Decode and save file
		var fileData []byte
		if strings.HasPrefix(file.Data, "data:") {
			// Handle data URL format
			parts := strings.Split(file.Data, ",")
			if len(parts) == 2 {
				decoded, err := base64.StdEncoding.DecodeString(parts[1])
				if err != nil {
					return fmt.Errorf("failed to decode file data: %v", err)
				}
				fileData = decoded
			}
		} else {
			// Handle plain text (JSON files)
			fileData = []byte(file.Data)
		}
		
		// Write file to disk
		if err := os.WriteFile(filePath, fileData, 0644); err != nil {
			return fmt.Errorf("failed to save file: %v", err)
		}
		
		// Add to attachments list
		attachment := Attachment{
			ID:       attachmentID,
			Name:     file.Name,
			Size:     file.Size,
			Type:     file.Type,
			Path:     filePath,
			UploadAt: time.Now(),
		}
		k.FormCardAttachments = append(k.FormCardAttachments, attachment)
	}
	
	k.Commit()
	return nil
}

// RefreshAttachments updates the attachments list after successful upload via REST API
func (k *SimpleKanbanModal) RefreshAttachments(data interface{}) {
	dataStr := ""
	if str, ok := data.(string); ok {
		dataStr = str
	}
	
	// Parse the response data
	var response struct {
		CardID string `json:"cardID"`
		Files  []struct {
			ID       string `json:"id"`
			Name     string `json:"name"`
			Size     int64  `json:"size"`
			Path     string `json:"path"`
			Uploaded string `json:"uploaded"`
		} `json:"files"`
	}
	
	if err := json.Unmarshal([]byte(dataStr), &response); err != nil {
		fmt.Printf("Error parsing refresh data: %v\n", err)
		return
	}
	
	// Add new attachments to the current card
	for _, file := range response.Files {
		uploadTime, _ := time.Parse(time.RFC3339, file.Uploaded)
		attachment := Attachment{
			ID:       file.ID,
			Name:     file.Name,
			Size:     file.Size,
			Type:     "", // Could be determined from extension
			Path:     file.Path,
			UploadAt: uploadTime,
		}
		k.FormCardAttachments = append(k.FormCardAttachments, attachment)
	}
	
	// Update the card in the cards list if we're editing
	if k.FormCardID == response.CardID {
		for i, card := range k.Cards {
			if card.ID == k.FormCardID {
				k.Cards[i].Attachments = k.FormCardAttachments
				break
			}
		}
		
		// Save to JSON
		saveBoardData(k.CurrentBoard, &KanbanBoardData{
			Columns: k.Columns,
			Cards:   k.Cards,
		})
	}
	
	fmt.Printf("📎 Refreshed attachments for card %s: %d new files\n", response.CardID, len(response.Files))
	k.Commit()
}

// UploadFiles handles file upload events from the UI
func (k *SimpleKanbanModal) UploadFiles(data interface{}) {
	if k.FileUpload != nil && len(k.FileUpload.Files) > 0 {
		if err := k.handleFileUpload(k.FileUpload.Files); err != nil {
			fmt.Printf("Error uploading files: %v\n", err)
			k.EvalScript(fmt.Sprintf("alert('Error uploading files: %v')", err))
		} else {
			// Clear the upload component after successful upload
			k.FileUpload.Clear()
		}
		k.Commit()
	}
}

// FormatFileSize formats file size in a human-readable format (public for template access)
func (k *SimpleKanbanModal) FormatFileSize(size int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)
	
	switch {
	case size >= GB:
		return fmt.Sprintf("%.2f GB", float64(size)/float64(GB))
	case size >= MB:
		return fmt.Sprintf("%.2f MB", float64(size)/float64(MB))
	case size >= KB:
		return fmt.Sprintf("%.2f KB", float64(size)/float64(KB))
	default:
		return fmt.Sprintf("%d bytes", size)
	}
}

// RemoveAttachment removes an attachment from the current card
func (k *SimpleKanbanModal) RemoveAttachment(data interface{}) {
	attachmentID := ""
	if id, ok := data.(string); ok {
		attachmentID = id
	}
	
	// Remove from form attachments
	newAttachments := []Attachment{}
	for _, att := range k.FormCardAttachments {
		if att.ID != attachmentID {
			newAttachments = append(newAttachments, att)
		} else {
			// Delete file from disk
			os.Remove(att.Path)
		}
	}
	k.FormCardAttachments = newAttachments
	k.Commit()
}