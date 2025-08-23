package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/arturoeanton/go-echo-live-view/liveview"
)

const JSON_FILE = "kanban_board.json"

// Global state for synchronization
var (
	globalMutex     sync.RWMutex
	globalBoard     *KanbanBoardData
	activeBoards    []*SimpleKanbanModal
	activeMutex     sync.Mutex
)

// KanbanBoardData represents the persistent board state
type KanbanBoardData struct {
	Columns []KanbanColumn `json:"columns"`
	Cards   []KanbanCard   `json:"cards"`
}

// SimpleKanbanModal - Kanban board without collaboration to avoid channel panic
type SimpleKanbanModal struct {
	*liveview.ComponentDriver[*SimpleKanbanModal] `json:"-"`
	
	// Board data
	Title   string         `json:"title"`
	Columns []KanbanColumn `json:"columns"`
	Cards   []KanbanCard   `json:"cards"`
	
	// Modal state
	ShowModal     bool   `json:"show_modal"`
	ModalType     string `json:"modal_type"` // "edit_card", "add_card", "edit_column", "add_column"
	ModalTitle    string `json:"modal_title"`
	
	// Form fields for modal
	FormCardID      string `json:"form_card_id"`
	FormCardTitle   string `json:"form_card_title"`
	FormCardDesc    string `json:"form_card_desc"`
	FormCardColumn  string `json:"form_card_column"`
	FormCardPriority string `json:"form_card_priority"`
	FormCardPoints  int    `json:"form_card_points"`
	
	FormColumnID    string `json:"form_column_id"`
	FormColumnTitle string `json:"form_column_title"`
	FormColumnColor string `json:"form_column_color"`
}

// KanbanColumn represents a column
type KanbanColumn struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Color string `json:"color"`
	Order int    `json:"order"`
}

// KanbanCard represents a card
type KanbanCard struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	ColumnID    string    `json:"column_id"`
	Priority    string    `json:"priority"`
	Points      int       `json:"points"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// loadBoardData loads board data from JSON file
func loadBoardData() *KanbanBoardData {
	data, err := os.ReadFile(JSON_FILE)
	if err != nil {
		// Return default board if file doesn't exist
		fmt.Printf("JSON file not found, creating default board: %v\n", err)
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

// saveBoardData saves board data to JSON file
func saveBoardData(board *KanbanBoardData) error {
	globalMutex.Lock()
	defer globalMutex.Unlock()
	
	data, err := json.MarshalIndent(board, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling data: %v", err)
	}
	
	if err := os.WriteFile(JSON_FILE, data, 0644); err != nil {
		return fmt.Errorf("error writing file: %v", err)
	}
	
	fmt.Println("💾 Board data saved to JSON file")
	return nil
}

// registerBoard adds a board instance to the active boards list
func registerBoard(board *SimpleKanbanModal) {
	activeMutex.Lock()
	defer activeMutex.Unlock()
	activeBoards = append(activeBoards, board)
	fmt.Printf("📝 Registered board, total active: %d\n", len(activeBoards))
}

// unregisterBoard removes a board instance from the active boards list
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

// broadcastUpdate sends updates to all active board instances
func broadcastUpdate() {
	activeMutex.Lock()
	defer activeMutex.Unlock()
	
	globalMutex.RLock()
	defer globalMutex.RUnlock()
	
	if globalBoard == nil {
		return
	}
	
	fmt.Printf("📡 Broadcasting update to %d active boards\n", len(activeBoards))
	
	for _, board := range activeBoards {
		if board != nil && board.ComponentDriver != nil {
			// Update board data from global state
			board.Columns = make([]KanbanColumn, len(globalBoard.Columns))
			copy(board.Columns, globalBoard.Columns)
			board.Cards = make([]KanbanCard, len(globalBoard.Cards))
			copy(board.Cards, globalBoard.Cards)
			
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

// updateGlobalState updates global state and broadcasts to all boards
func updateGlobalState(columns []KanbanColumn, cards []KanbanCard) {
	globalMutex.Lock()
	if globalBoard == nil {
		globalBoard = &KanbanBoardData{}
	}
	globalBoard.Columns = make([]KanbanColumn, len(columns))
	copy(globalBoard.Columns, columns)
	globalBoard.Cards = make([]KanbanCard, len(cards))
	copy(globalBoard.Cards, cards)
	globalMutex.Unlock()
	
	// Save to JSON
	go func() {
		if err := saveBoardData(globalBoard); err != nil {
			fmt.Printf("Error saving board data: %v\n", err)
		}
	}()
	
	// Broadcast to all active boards
	broadcastUpdate()
}

// NewSimpleKanbanModal creates a new simple kanban board with modals
func NewSimpleKanbanModal() *SimpleKanbanModal {
	board := &SimpleKanbanModal{
		Title:     "📋 Simple Kanban Board",
		ShowModal: false,
	}
	
	board.ComponentDriver = liveview.NewDriver[*SimpleKanbanModal]("kanban_board", board)
	
	// Initialize global state if this is the first board
	globalMutex.Lock()
	if globalBoard == nil {
		boardData := loadBoardData()
		if boardData != nil {
			globalBoard = boardData
		}
	}
	
	// Load data from global state
	if globalBoard != nil {
		board.Columns = make([]KanbanColumn, len(globalBoard.Columns))
		copy(board.Columns, globalBoard.Columns)
		board.Cards = make([]KanbanCard, len(globalBoard.Cards))
		copy(board.Cards, globalBoard.Cards)
	}
	globalMutex.Unlock()
	
	// Register this board for broadcasting
	registerBoard(board)
	
	return board
}

// GetDriver returns the component driver
func (k *SimpleKanbanModal) GetDriver() liveview.LiveDriver {
	return k.ComponentDriver
}

// Start initializes the board
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
	}
	
	k.Commit()
	fmt.Println("Simple Kanban Modal Board initialized with explicit event registration")
	
	// Initialize column drag & drop via EvalScript
	k.initializeColumnDragDrop()
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
		.board-header { text-align: center; color: white; margin-bottom: 30px; }
		.board-header h1 { font-size: 2.5em; margin-bottom: 10px; text-shadow: 2px 2px 4px rgba(0,0,0,0.3); }
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
	
	<div class="kanban-board">
		<div class="board-header">
			<h1>{{.Title}}</h1>
			<p>Drag cards between columns • Click cards/columns to edit</p>
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
						<div>
							{{if .Priority}}<span class="card-priority priority-{{.Priority}}">{{.Priority}}</span>{{end}}
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
				<button onclick="send_event('{{.IdComponent}}', 'SaveModal', '')" 
						style="background: #3498db; color: white; padding: 12px 25px; border: none; border-radius: 6px; cursor: pointer;">
					{{if or (eq .ModalType "add_card") (eq .ModalType "add_column")}}Add{{else}}Save{{end}}
				</button>
			</div>
		</div>
	</div>
	{{end}}
	`
}

// Helper functions
func (k *SimpleKanbanModal) GetCardsForColumn(columnID string) []KanbanCard {
	var cards []KanbanCard
	for _, card := range k.Cards {
		if card.ColumnID == columnID {
			cards = append(cards, card)
		}
	}
	return cards
}

func (k *SimpleKanbanModal) GetCardCount(columnID string) int {
	count := 0
	for _, card := range k.Cards {
		if card.ColumnID == columnID {
			count++
		}
	}
	return count
}

// GetColumnPoints returns total points for a column
func (k *SimpleKanbanModal) GetColumnPoints(columnID string) int {
	total := 0
	for _, card := range k.Cards {
		if card.ColumnID == columnID {
			total += card.Points
		}
	}
	return total
}

// GetOrderedColumns returns columns sorted by their Order field
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

// Event Handlers

// MoveCard handles HTML5 drag and drop
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
		updateGlobalState(k.Columns, updatedCards)
	}
}

// EditCard opens modal to edit a card
func (k *SimpleKanbanModal) EditCard(data interface{}) {
	cardID := ""
	if id, ok := data.(string); ok {
		cardID = id
	}
	
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
			if k.FormCardPriority == "" {
				k.FormCardPriority = "medium"
			}
			break
		}
	}
	k.Commit()
}

// AddCard opens modal to add a new card
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
	k.Commit()
}

// EditColumn opens modal to edit a column
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
		}
	}
}

// CloseModal closes the modal
func (k *SimpleKanbanModal) CloseModal(data interface{}) {
	k.ShowModal = false
	k.Commit()
}

// SaveModal saves the modal form
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
	updateGlobalState(updatedColumns, updatedCards)
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
	updateGlobalState(k.Columns, newCards)
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
	updateGlobalState(newColumns, newCards)
}

// ReorderColumns handles column reordering via drag & drop
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
		updateGlobalState(updatedColumns, k.Cards)
	}
}