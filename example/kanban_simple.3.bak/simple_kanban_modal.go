package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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
	RichEditor   *components.RichEditor   `json:"-"` // Rich text editor for card descriptions
	TabsComponent *components.Tabs        `json:"-"` // Tabs component for modal
	
	// Modal state - controls the popup dialog
	ShowModal     bool   `json:"show_modal"`     // Whether modal is visible
	ModalType     string `json:"modal_type"`     // Type: "edit_card", "add_card", "edit_column", "add_column", "new_board"
	ModalTitle    string `json:"modal_title"`    // Title shown in modal header
	
	// Form fields for card editing/creation modal
	FormCardID          string          `json:"form_card_id"`          // ID of card being edited
	FormCardTitle       string          `json:"form_card_title"`       // Card title input value
	FormCardDesc        string          `json:"form_card_desc"`        // Card description input value (HTML)
	FormCardColumn      string          `json:"form_card_column"`      // Selected column for the card
	FormCardPriority    string          `json:"form_card_priority"`    // Selected priority level
	FormCardPoints      int             `json:"form_card_points"`      // Story points (0-100)
	FormCardAttachments []Attachment    `json:"form_card_attachments"` // Current attachments
	FormCardLinks       []ExternalLink  `json:"form_card_links"`       // External links
	FormCardTags        []string        `json:"form_card_tags"`        // Tags/labels
	FormCardDueDate     string          `json:"form_card_due_date"`    // Due date string (for form input)
	FormCardChecklist   []ChecklistItem `json:"form_card_checklist"`   // Checklist items
	
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
	Description string       `json:"description"` // Detailed description of the task (HTML from RichEditor)
	ColumnID    string       `json:"column_id"`   // ID of the column containing this card
	Priority    string       `json:"priority"`    // Priority level: low, medium, high, urgent
	Points      int          `json:"points"`      // Story points for effort estimation (0-100)
	Attachments []Attachment `json:"attachments"` // File attachments
	Links       []ExternalLink `json:"links"`     // External links
	Tags        []string     `json:"tags"`        // Tags/labels for categorization
	DueDate     *time.Time   `json:"due_date"`    // Due date for the task
	Checklist   []ChecklistItem `json:"checklist"` // Checklist items within the card
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

// ExternalLink represents an external URL attached to a card
type ExternalLink struct {
	ID    string `json:"id"`    // Unique identifier
	Title string `json:"title"` // Display title for the link
	URL   string `json:"url"`   // The actual URL
	Icon  string `json:"icon"`  // Optional icon or favicon URL
}

// ChecklistItem represents a single item in a card's checklist
type ChecklistItem struct {
	ID      string `json:"id"`      // Unique identifier
	Text    string `json:"text"`    // Checklist item text
	Checked bool   `json:"checked"` // Whether the item is completed
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
		// Convert FileInfo to interface{} for handleFileUpload
		var filesInterface []interface{}
		for _, file := range files {
			filesInterface = append(filesInterface, file)
		}
		return board.handleFileUpload(filesInterface)
	}
	fmt.Printf("📁 FileUpload component created: %v with driver: %v\n", board.FileUpload != nil, board.FileUpload.ComponentDriver != nil)
	
	// Create RichEditor component for card descriptions
	board.RichEditor = &components.RichEditor{
		Placeholder: "Enter card description...",
		Height:      "200px",
	}
	richEditorDriver := liveview.NewDriver[*components.RichEditor]("rich_editor", board.RichEditor)
	board.RichEditor.ComponentDriver = richEditorDriver
	board.RichEditor.OnChange = func(content string) {
		board.FormCardDesc = content
	}
	fmt.Printf("📝 RichEditor component created: %v with driver: %v\n", board.RichEditor != nil, board.RichEditor.ComponentDriver != nil)
	
	// Create Tabs component for modal
	board.TabsComponent = &components.Tabs{
		Tabs: []components.Tab{},
		ActiveTab: "general",
	}
	tabsDriver := liveview.NewDriver[*components.Tabs]("modal_tabs", board.TabsComponent)
	board.TabsComponent.ComponentDriver = tabsDriver
	
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
	
	// Mount the RichEditor component
	if k.RichEditor != nil {
		k.Mount(k.RichEditor)
		fmt.Printf("✅ RichEditor component mounted successfully\n")
	}
	
	// Mount the Tabs component
	if k.TabsComponent != nil {
		k.Mount(k.TabsComponent)
		fmt.Printf("✅ Tabs component mounted successfully\n")
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
		
		/* Modal Tabs */
		.tabs-container {
			display: flex;
			border-bottom: 2px solid #ecf0f1;
			margin-bottom: 20px;
		}
		.tab-button {
			padding: 10px 20px;
			background: none;
			border: none;
			cursor: pointer;
			color: #7f8c8d;
			font-size: 14px;
			font-weight: 500;
			transition: all 0.3s;
			border-bottom: 3px solid transparent;
			margin-bottom: -2px;
		}
		.tab-button:hover {
			color: #34495e;
		}
		.tab-button.active {
			color: #3498db;
			border-bottom-color: #3498db;
		}
		.tab-content {
			display: none;
		}
		.tab-content.active {
			display: block;
		}
		.modal-body {
			max-height: calc(90vh - 200px);
			overflow-y: auto;
			padding-right: 10px;
		}
		.modal-footer {
			background: white;
			padding-top: 20px;
			margin-top: 20px;
			border-top: 1px solid #ecf0f1;
			display: flex;
			justify-content: flex-end;
			gap: 15px;
		}
		.modal-container {
			display: flex;
			flex-direction: column;
			height: 100%;
		}
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
		.column-header:hover { opacity: 0.95; }
		.column-edit-btn { 
			background: rgba(255,255,255,0.9); 
			border: none;
			border-radius: 4px; 
			padding: 4px 8px; 
			cursor: pointer; 
			font-size: 16px; 
			opacity: 0; 
			transition: all 0.2s;
			box-shadow: 0 2px 4px rgba(0,0,0,0.1);
		}
		.column-header:hover .column-edit-btn { opacity: 1; }
		.column-edit-btn:hover { background: white; transform: scale(1.1); }
		.column-cards { padding: 15px; min-height: 300px; }
		.kanban-card { background: white; padding: 15px; margin-bottom: 15px; border-radius: 8px; box-shadow: 0 2px 8px rgba(0,0,0,0.1); cursor: pointer; transition: all 0.2s; position: relative; }
		.kanban-card:hover { box-shadow: 0 4px 15px rgba(0,0,0,0.15); transform: translateY(-2px); background: #f8f9fa; }
		.kanban-card::after { content: "Click to edit"; position: absolute; top: 5px; right: 5px; font-size: 10px; color: #95a5a6; opacity: 0; transition: opacity 0.2s; }
		.kanban-card:hover::after { opacity: 1; }
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
					 title="Drag to reorder columns">
					<span style="font-weight: 600;">{{.Title}}</span>
					<div style="display: flex; align-items: center; gap: 10px;">
						<span style="font-size: 0.9em; opacity: 0.8;">({{$.GetCardCount .ID}} cards | {{$.GetColumnPoints .ID}} pts)</span>
						<button class="column-edit-btn" 
								onclick="event.stopPropagation(); send_event('{{$.IdComponent}}', 'EditColumn', '{{.ID}}');"
								title="Edit column">
							✏️
						</button>
					</div>
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
						{{if .Tags}}
						<div style="display: flex; flex-wrap: wrap; gap: 4px; margin: 8px 0;">
							{{range .Tags}}
							<span style="background: #3498db; color: white; padding: 2px 8px; border-radius: 12px; font-size: 11px;">{{.}}</span>
							{{end}}
						</div>
						{{end}}
						{{if .DueDate}}
						<div style="color: {{if $.IsOverdue .DueDate}}#e74c3c{{else}}#7f8c8d{{end}}; font-size: 12px; margin: 4px 0;">
							📅 {{.DueDate.Format "Jan 2, 2006"}}
						</div>
						{{end}}
						<div class="card-desc" style="max-height: 60px; overflow: hidden;">{{.Description}}</div>
						<div style="display: flex; align-items: center; gap: 10px; margin-top: 8px; flex-wrap: wrap;">
							{{if .Priority}}<span class="card-priority priority-{{.Priority}}">{{.Priority}}</span>{{end}}
							{{if .Checklist}}
							<span style="display: inline-flex; align-items: center; gap: 4px; color: #7f8c8d; font-size: 12px;">
								✅ {{$.CountCheckedItems .Checklist}}/{{len .Checklist}}
							</span>
							{{end}}
							{{if .Links}}
							<span style="display: inline-flex; align-items: center; gap: 4px; color: #7f8c8d; font-size: 12px;">
								🔗 {{len .Links}}
							</span>
							{{end}}
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
	<div style="position: fixed; top: 0; left: 0; right: 0; bottom: 0; background: rgba(0,0,0,0.6); z-index: 9999; display: flex; align-items: center; justify-content: center; padding: 20px;">
		<div class="modal-container" style="background: white; border-radius: 15px; width: 90%; max-width: 800px; max-height: 90vh; display: flex; flex-direction: column; box-shadow: 0 20px 60px rgba(0,0,0,0.3);">
			<div style="padding: 20px 30px; border-bottom: 1px solid #ecf0f1; display: flex; justify-content: space-between; align-items: center;">
				<h2 style="margin: 0; color: #2c3e50;">{{.ModalTitle}}</h2>
				<button onclick="send_event('{{.IdComponent}}', 'CloseModal', '')" 
						style="background: none; border: none; font-size: 24px; cursor: pointer; color: #95a5a6; padding: 5px;">&times;</button>
			</div>
			
			<div style="flex: 1; overflow-y: auto; padding: 20px 30px;">
				{{if or (eq .ModalType "edit_card") (eq .ModalType "add_card")}}
					<!-- Complete form for cards with all features -->
					<div style="display: flex; flex-direction: column; gap: 20px;">
						<!-- Title and Description -->
						<div>
							<label style="display: block; margin-bottom: 8px; font-weight: 500; color: #34495e;">Title</label>
							<input type="text" value="{{.FormCardTitle}}" 
								   oninput="send_event('{{.IdComponent}}', 'UpdateFormField', JSON.stringify({field: 'card_title', value: this.value}))"
								   style="width: 100%; padding: 12px; border: 1px solid #bdc3c7; border-radius: 6px; font-size: 14px; box-sizing: border-box;">
						</div>
						
						<div>
							<label style="display: block; margin-bottom: 8px; font-weight: 500; color: #34495e;">Description</label>
							<textarea oninput="send_event('{{.IdComponent}}', 'UpdateFormField', JSON.stringify({field: 'card_desc', value: this.value}))"
									  style="width: 100%; min-height: 120px; padding: 12px; border: 1px solid #bdc3c7; border-radius: 6px; font-size: 14px; resize: vertical; box-sizing: border-box;">{{.FormCardDesc}}</textarea>
						</div>
						
						<!-- Priority, Points, Due Date -->
						<div style="display: grid; grid-template-columns: 1fr 1fr 1fr; gap: 15px;">
							<div>
								<label style="display: block; margin-bottom: 8px; font-weight: 500; color: #34495e;">Priority</label>
								<select onchange="send_event('{{.IdComponent}}', 'UpdateFormField', JSON.stringify({field: 'card_priority', value: this.value}))"
										style="width: 100%; padding: 12px; border: 1px solid #bdc3c7; border-radius: 6px; font-size: 14px; box-sizing: border-box;">
									<option value="low" {{if eq .FormCardPriority "low"}}selected{{end}}>🟢 Low</option>
									<option value="medium" {{if eq .FormCardPriority "medium"}}selected{{end}}>🟡 Medium</option>
									<option value="high" {{if eq .FormCardPriority "high"}}selected{{end}}>🟠 High</option>
									<option value="urgent" {{if eq .FormCardPriority "urgent"}}selected{{end}}>🔴 Urgent</option>
								</select>
							</div>
							<div>
								<label style="display: block; margin-bottom: 8px; font-weight: 500; color: #34495e;">Points</label>
								<input type="number" value="{{.FormCardPoints}}" min="0" max="100"
									   oninput="send_event('{{.IdComponent}}', 'UpdateFormField', JSON.stringify({field: 'card_points', value: parseInt(this.value) || 0}))"
									   style="width: 100%; padding: 12px; border: 1px solid #bdc3c7; border-radius: 6px; font-size: 14px; box-sizing: border-box;">
							</div>
							<div>
								<label style="display: block; margin-bottom: 8px; font-weight: 500; color: #34495e;">Due Date</label>
								<input type="date" value="{{.FormCardDueDate}}" 
									   onchange="send_event('{{.IdComponent}}', 'UpdateFormField', JSON.stringify({field: 'card_due_date', value: this.value}))"
									   style="width: 100%; padding: 12px; border: 1px solid #bdc3c7; border-radius: 6px; font-size: 14px; box-sizing: border-box;">
							</div>
						</div>
						
						<!-- Tags Section -->
						<div>
							<label style="display: block; margin-bottom: 8px; font-weight: 500; color: #34495e;">🏷️ Tags</label>
							<div style="display: flex; flex-wrap: wrap; gap: 8px; margin-bottom: 8px; min-height: 32px; padding: 8px; background: #f8f9fa; border-radius: 6px;">
								{{range .FormCardTags}}
								<span style="background: #3498db; color: white; padding: 4px 12px; border-radius: 16px; font-size: 13px; display: inline-flex; align-items: center; gap: 6px;">
									{{.}}
									<button onclick="event.stopPropagation(); send_event('{{$.IdComponent}}', 'RemoveTag', '{{.}}')" 
											style="background: none; border: none; color: white; cursor: pointer; padding: 0; font-size: 16px; line-height: 1;">×</button>
								</span>
								{{end}}
								{{if not .FormCardTags}}
								<span style="color: #95a5a6; font-size: 13px;">No tags yet</span>
								{{end}}
							</div>
							<div style="display: flex; gap: 8px;">
								<input type="text" id="tag-input-{{.IdComponent}}" placeholder="Add a tag..." 
									   style="flex: 1; padding: 8px 12px; border: 1px solid #bdc3c7; border-radius: 6px; font-size: 14px;">
								<button onclick="var input = document.getElementById('tag-input-{{.IdComponent}}'); if(input.value) { send_event('{{.IdComponent}}', 'AddTag', input.value); input.value = ''; }"
										style="background: #3498db; color: white; border: none; padding: 8px 16px; border-radius: 6px; cursor: pointer; font-size: 14px;">Add Tag</button>
							</div>
						</div>
						
						<!-- External Links Section -->
						<div>
							<label style="display: block; margin-bottom: 8px; font-weight: 500; color: #34495e;">🔗 External Links</label>
							<div style="min-height: 40px; padding: 8px; background: #f8f9fa; border-radius: 6px; margin-bottom: 8px;">
								{{if .FormCardLinks}}
									{{range .FormCardLinks}}
									<div style="display: flex; align-items: center; justify-content: space-between; padding: 8px; background: white; border-radius: 4px; margin-bottom: 6px;">
										<a href="{{.URL}}" target="_blank" style="color: #3498db; text-decoration: none; flex: 1;">
											🔗 {{if .Title}}{{.Title}}{{else}}{{.URL}}{{end}}
										</a>
										<button onclick="event.stopPropagation(); send_event('{{$.IdComponent}}', 'RemoveLink', '{{.ID}}')" 
												style="background: #e74c3c; color: white; border: none; border-radius: 4px; padding: 4px 8px; cursor: pointer; font-size: 12px;">×</button>
									</div>
									{{end}}
								{{else}}
									<span style="color: #95a5a6; font-size: 13px;">No links yet</span>
								{{end}}
							</div>
							<div style="display: flex; gap: 8px;">
								<input type="text" id="link-title-{{.IdComponent}}" placeholder="Title (optional)" 
									   style="width: 30%; padding: 8px 12px; border: 1px solid #bdc3c7; border-radius: 6px; font-size: 14px;">
								<input type="url" id="link-url-{{.IdComponent}}" placeholder="https://..." 
									   style="flex: 1; padding: 8px 12px; border: 1px solid #bdc3c7; border-radius: 6px; font-size: 14px;">
								<button onclick="var title = document.getElementById('link-title-{{.IdComponent}}').value; 
												var url = document.getElementById('link-url-{{.IdComponent}}').value;
												if(url) { 
													send_event('{{.IdComponent}}', 'AddLink', JSON.stringify({title: title || url, url: url})); 
													document.getElementById('link-title-{{.IdComponent}}').value = '';
													document.getElementById('link-url-{{.IdComponent}}').value = '';
												}"
										style="background: #3498db; color: white; border: none; padding: 8px 16px; border-radius: 6px; cursor: pointer; font-size: 14px;">Add Link</button>
							</div>
						</div>
						
						<!-- Checklist Section -->
						<div>
							<label style="display: block; margin-bottom: 8px; font-weight: 500; color: #34495e;">✅ Checklist</label>
							<div style="min-height: 40px; padding: 8px; background: #f8f9fa; border-radius: 6px; margin-bottom: 8px; max-height: 200px; overflow-y: auto;">
								{{if .FormCardChecklist}}
									{{range .FormCardChecklist}}
									<div style="display: flex; align-items: center; gap: 10px; padding: 8px; background: white; border-radius: 4px; margin-bottom: 6px;">
										<input type="checkbox" {{if .Checked}}checked{{end}}
											   onchange="send_event('{{$.IdComponent}}', 'ToggleChecklistItem', '{{.ID}}')"
											   style="width: 18px; height: 18px; cursor: pointer;">
										<span style="flex: 1; {{if .Checked}}text-decoration: line-through; color: #95a5a6;{{end}}">{{.Text}}</span>
										<button onclick="event.stopPropagation(); send_event('{{$.IdComponent}}', 'RemoveChecklistItem', '{{.ID}}')" 
												style="background: #e74c3c; color: white; border: none; border-radius: 4px; padding: 4px 8px; cursor: pointer; font-size: 12px;">×</button>
									</div>
									{{end}}
								{{else}}
									<span style="color: #95a5a6; font-size: 13px;">No checklist items yet</span>
								{{end}}
							</div>
							<div style="display: flex; gap: 8px;">
								<input type="text" id="checklist-input-{{.IdComponent}}" placeholder="Add checklist item..." 
									   style="flex: 1; padding: 8px 12px; border: 1px solid #bdc3c7; border-radius: 6px; font-size: 14px;">
								<button onclick="var input = document.getElementById('checklist-input-{{.IdComponent}}'); 
												if(input.value) { 
													send_event('{{.IdComponent}}', 'AddChecklistItem', input.value); 
													input.value = ''; 
												}"
										style="background: #3498db; color: white; border: none; padding: 8px 16px; border-radius: 6px; cursor: pointer; font-size: 14px;">Add Item</button>
							</div>
						</div>
					</div>
				{{else if or (eq .ModalType "edit_column") (eq .ModalType "add_column")}}
					<div style="display: flex; flex-direction: column; gap: 20px;">
						<div>
							<label style="display: block; margin-bottom: 8px; font-weight: 500; color: #34495e;">Column Name</label>
							<input type="text" value="{{.FormColumnTitle}}" 
								   oninput="send_event('{{.IdComponent}}', 'UpdateFormField', JSON.stringify({field: 'column_title', value: this.value}))"
								   style="width: 100%; padding: 12px; border: 1px solid #bdc3c7; border-radius: 6px; font-size: 14px; box-sizing: border-box;">
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
			</div>
			
			<div style="padding: 20px 30px; border-top: 1px solid #ecf0f1; display: flex; justify-content: space-between; align-items: center;">
				<div>
					{{if eq .ModalType "edit_column"}}
					<button onclick="send_event('{{.IdComponent}}', 'DeleteColumn', '')" 
							style="background: #e74c3c; color: white; padding: 12px 25px; border: none; border-radius: 6px; cursor: pointer; font-size: 14px; font-weight: 500;">
						🗑️ Delete Column
					</button>
					{{else if eq .ModalType "edit_card"}}
					<button onclick="send_event('{{.IdComponent}}', 'DeleteCard', '')" 
							style="background: #e74c3c; color: white; padding: 12px 25px; border: none; border-radius: 6px; cursor: pointer; font-size: 14px; font-weight: 500;">
						🗑️ Delete Card
					</button>
					{{end}}
				</div>
				
				<div style="display: flex; gap: 10px;">
					<button onclick="send_event('{{.IdComponent}}', 'CloseModal', '')" 
							style="background: #95a5a6; color: white; padding: 12px 25px; border: none; border-radius: 6px; cursor: pointer; font-size: 14px; font-weight: 500;">
						Cancel
					</button>
					<button onclick="send_event('{{.IdComponent}}', 'SaveModal', '')" 
							style="background: #3498db; color: white; padding: 12px 25px; border: none; border-radius: 6px; cursor: pointer; font-size: 14px; font-weight: 500;">
						💾 {{if or (eq .ModalType "add_card") (eq .ModalType "add_column")}}Add{{else}}Save{{end}}
					</button>
				</div>
			</div>
		</div>
	</div>
	{{end}}
	
	`
}

// Temporary placeholder for GetModalContent - will be removed later
func (k *SimpleKanbanModal) GetModalContent() string {
	return ""
}

