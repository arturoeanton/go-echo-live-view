package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/arturoeanton/go-echo-live-view/components"
	"github.com/arturoeanton/go-echo-live-view/liveview"
)

// EnhancedKanban extends KanbanBoard with modal functionality
type EnhancedKanban struct {
	*components.KanbanBoard
	
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
	
	FormColumnID    string `json:"form_column_id"`
	FormColumnTitle string `json:"form_column_title"`
	FormColumnColor string `json:"form_column_color"`
}

// NewEnhancedKanban creates a new enhanced kanban board
func NewEnhancedKanban() *EnhancedKanban {
	// Create base board
	baseBoard := &components.KanbanBoard{}
	
	// Initialize embedded struct
	baseBoard.CollaborativeComponent = &liveview.CollaborativeComponent{}
	
	// Create enhanced board
	board := &EnhancedKanban{
		KanbanBoard: baseBoard,
		ShowModal:   false,
	}
	
	// Create the driver
	board.ComponentDriver = liveview.NewDriver[*components.KanbanBoard]("kanban_board", board.KanbanBoard)
	
	// Set the driver reference
	if board.CollaborativeComponent != nil {
		board.CollaborativeComponent.Driver = board.ComponentDriver
	}
	
	// Initialize board data
	board.Title = "📋 Simple Kanban Board"
	board.Description = "Drag cards between columns • Click cards/columns to edit"
	
	// Setup default columns
	board.Columns = []components.KanbanColumn{
		{
			ID:    "todo",
			Title: "To Do",
			Color: "#e3e8ef",
			Order: 0,
		},
		{
			ID:    "doing",
			Title: "In Progress",
			Color: "#ffd4a3",
			Order: 1,
		},
		{
			ID:    "done",
			Title: "Done",
			Color: "#a3e4d7",
			Order: 2,
		},
	}
	
	// Add sample card
	board.Cards = []components.KanbanCard{
		{
			ID:          "welcome",
			Title:       "Welcome to Enhanced Kanban!",
			Description: "Click me to edit • Double-click columns to edit them • Use + buttons to add new items",
			ColumnID:    "todo",
			Priority:    "medium",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}
	
	// Initialize active users map
	board.ActiveUsers = make(map[string]*components.UserActivity)
	
	return board
}

// GetTemplate returns the enhanced template with modal
func (k *EnhancedKanban) GetTemplate() string {
	// Get base template from parent
	baseTemplate := k.KanbanBoard.GetTemplate()
	
	// Add our modal template
	modalTemplate := `
	{{if .ShowModal}}
	<div id="modal-overlay" style="position: fixed; top: 0; left: 0; right: 0; bottom: 0; background: rgba(0,0,0,0.5); z-index: 9999; display: flex; align-items: center; justify-content: center;">
		<div style="background: white; border-radius: 10px; padding: 25px; min-width: 400px; max-width: 600px; box-shadow: 0 10px 40px rgba(0,0,0,0.3);">
			<div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px;">
				<h2 style="margin: 0; color: #2c3e50;">{{.ModalTitle}}</h2>
				<button onclick="send_event('{{.IdComponent}}', 'CloseModal', '')" style="background: none; border: none; font-size: 24px; cursor: pointer; color: #95a5a6;">&times;</button>
			</div>
			
			{{if or (eq .ModalType "edit_card") (eq .ModalType "add_card")}}
			<div style="display: flex; flex-direction: column; gap: 15px;">
				<div>
					<label style="display: block; margin-bottom: 5px; font-weight: 500; color: #34495e;">Title</label>
					<input type="text" value="{{.FormCardTitle}}" 
						   oninput="send_event('{{.IdComponent}}', 'UpdateFormField', JSON.stringify({field: 'card_title', value: this.value}))"
						   style="width: 100%; padding: 10px; border: 1px solid #bdc3c7; border-radius: 5px; font-size: 14px;">
				</div>
				
				<div>
					<label style="display: block; margin-bottom: 5px; font-weight: 500; color: #34495e;">Description</label>
					<textarea oninput="send_event('{{.IdComponent}}', 'UpdateFormField', JSON.stringify({field: 'card_desc', value: this.value}))"
							  style="width: 100%; padding: 10px; border: 1px solid #bdc3c7; border-radius: 5px; font-size: 14px; min-height: 100px; resize: vertical;">{{.FormCardDesc}}</textarea>
				</div>
				
				<div>
					<label style="display: block; margin-bottom: 5px; font-weight: 500; color: #34495e;">Priority</label>
					<select onchange="send_event('{{.IdComponent}}', 'UpdateFormField', JSON.stringify({field: 'card_priority', value: this.value}))"
							style="width: 100%; padding: 10px; border: 1px solid #bdc3c7; border-radius: 5px; font-size: 14px;">
						<option value="low" {{if eq .FormCardPriority "low"}}selected{{end}}>Low</option>
						<option value="medium" {{if eq .FormCardPriority "medium"}}selected{{end}}>Medium</option>
						<option value="high" {{if eq .FormCardPriority "high"}}selected{{end}}>High</option>
						<option value="urgent" {{if eq .FormCardPriority "urgent"}}selected{{end}}>Urgent</option>
					</select>
				</div>
				
				{{if eq .ModalType "edit_card"}}
				<div>
					<label style="display: block; margin-bottom: 5px; font-weight: 500; color: #34495e;">Column</label>
					<select onchange="send_event('{{.IdComponent}}', 'UpdateFormField', JSON.stringify({field: 'card_column', value: this.value}))"
							style="width: 100%; padding: 10px; border: 1px solid #bdc3c7; border-radius: 5px; font-size: 14px;">
						{{range .Columns}}
						<option value="{{.ID}}" {{if eq $.FormCardColumn .ID}}selected{{end}}>{{.Title}}</option>
						{{end}}
					</select>
				</div>
				{{end}}
			</div>
			{{end}}
			
			{{if or (eq .ModalType "edit_column") (eq .ModalType "add_column")}}
			<div style="display: flex; flex-direction: column; gap: 15px;">
				<div>
					<label style="display: block; margin-bottom: 5px; font-weight: 500; color: #34495e;">Column Name</label>
					<input type="text" value="{{.FormColumnTitle}}" 
						   oninput="send_event('{{.IdComponent}}', 'UpdateFormField', JSON.stringify({field: 'column_title', value: this.value}))"
						   style="width: 100%; padding: 10px; border: 1px solid #bdc3c7; border-radius: 5px; font-size: 14px;">
				</div>
				
				<div>
					<label style="display: block; margin-bottom: 5px; font-weight: 500; color: #34495e;">Color</label>
					<div style="display: flex; gap: 10px; flex-wrap: wrap;">
						<div onclick="send_event('{{.IdComponent}}', 'UpdateFormField', JSON.stringify({field: 'column_color', value: '#e3e8ef'}))"
							 style="width: 40px; height: 40px; background: #e3e8ef; border-radius: 5px; cursor: pointer; {{if eq .FormColumnColor "#e3e8ef"}}box-shadow: 0 0 0 3px #3498db;{{end}}"></div>
						<div onclick="send_event('{{.IdComponent}}', 'UpdateFormField', JSON.stringify({field: 'column_color', value: '#ffd4a3'}))"
							 style="width: 40px; height: 40px; background: #ffd4a3; border-radius: 5px; cursor: pointer; {{if eq .FormColumnColor "#ffd4a3"}}box-shadow: 0 0 0 3px #3498db;{{end}}"></div>
						<div onclick="send_event('{{.IdComponent}}', 'UpdateFormField', JSON.stringify({field: 'column_color', value: '#a3e4d7'}))"
							 style="width: 40px; height: 40px; background: #a3e4d7; border-radius: 5px; cursor: pointer; {{if eq .FormColumnColor "#a3e4d7"}}box-shadow: 0 0 0 3px #3498db;{{end}}"></div>
						<div onclick="send_event('{{.IdComponent}}', 'UpdateFormField', JSON.stringify({field: 'column_color', value: '#f8b3d0'}))"
							 style="width: 40px; height: 40px; background: #f8b3d0; border-radius: 5px; cursor: pointer; {{if eq .FormColumnColor "#f8b3d0"}}box-shadow: 0 0 0 3px #3498db;{{end}}"></div>
						<div onclick="send_event('{{.IdComponent}}', 'UpdateFormField', JSON.stringify({field: 'column_color', value: '#b3d4f8'}))"
							 style="width: 40px; height: 40px; background: #b3d4f8; border-radius: 5px; cursor: pointer; {{if eq .FormColumnColor "#b3d4f8"}}box-shadow: 0 0 0 3px #3498db;{{end}}"></div>
						<div onclick="send_event('{{.IdComponent}}', 'UpdateFormField', JSON.stringify({field: 'column_color', value: '#d4b3f8'}))"
							 style="width: 40px; height: 40px; background: #d4b3f8; border-radius: 5px; cursor: pointer; {{if eq .FormColumnColor "#d4b3f8"}}box-shadow: 0 0 0 3px #3498db;{{end}}"></div>
					</div>
				</div>
			</div>
			{{end}}
			
			<div style="display: flex; justify-content: flex-end; gap: 10px; margin-top: 25px;">
				{{if eq .ModalType "edit_column"}}
				<button onclick="send_event('{{.IdComponent}}', 'DeleteColumn', '')" 
						style="background: #e74c3c; color: white; padding: 10px 20px; border: none; border-radius: 5px; cursor: pointer; margin-right: auto;">
					Delete Column
				</button>
				{{end}}
				
				{{if eq .ModalType "edit_card"}}
				<button onclick="send_event('{{.IdComponent}}', 'DeleteCard', '')" 
						style="background: #e74c3c; color: white; padding: 10px 20px; border: none; border-radius: 5px; cursor: pointer; margin-right: auto;">
					Delete Card
				</button>
				{{end}}
				
				<button onclick="send_event('{{.IdComponent}}', 'CloseModal', '')" 
						style="background: #95a5a6; color: white; padding: 10px 20px; border: none; border-radius: 5px; cursor: pointer;">
					Cancel
				</button>
				<button onclick="send_event('{{.IdComponent}}', 'SaveModal', '')" 
						style="background: #3498db; color: white; padding: 10px 20px; border: none; border-radius: 5px; cursor: pointer;">
					{{if or (eq .ModalType "add_card") (eq .ModalType "add_column")}}Add{{else}}Save{{end}}
				</button>
			</div>
		</div>
	</div>
	{{end}}
	
	<script>
	// Override click handlers for cards and columns
	document.addEventListener('DOMContentLoaded', function() {
		// Add click handlers for cards
		setTimeout(function() {
			document.querySelectorAll('.kanban-card').forEach(function(card) {
				card.style.cursor = 'pointer';
				card.onclick = function(e) {
					e.stopPropagation();
					var cardId = this.getAttribute('data-card-id');
					send_event('{{.IdComponent}}', 'EditCard', cardId);
				};
			});
			
			// Add double-click handlers for column headers
			document.querySelectorAll('.column-header').forEach(function(header) {
				header.style.cursor = 'pointer';
				header.ondblclick = function(e) {
					e.stopPropagation();
					var columnId = this.parentElement.querySelector('.cards-container').getAttribute('data-column-id');
					send_event('{{.IdComponent}}', 'EditColumn', columnId);
				};
			});
			
			// Add "Add Card" buttons to each column
			document.querySelectorAll('.cards-container').forEach(function(container) {
				var columnId = container.getAttribute('data-column-id');
				var addBtn = document.createElement('button');
				addBtn.innerHTML = '+ Add Card';
				addBtn.style.cssText = 'width: 100%; padding: 10px; margin-top: 10px; border: 2px dashed #bdc3c7; background: transparent; border-radius: 5px; cursor: pointer; color: #7f8c8d;';
				addBtn.onclick = function() {
					send_event('{{.IdComponent}}', 'AddCard', columnId);
				};
				container.appendChild(addBtn);
			});
			
			// Add "Add Column" button
			var columnsContainer = document.querySelector('.kanban-columns');
			if (columnsContainer) {
				var addColBtn = document.createElement('div');
				addColBtn.className = 'kanban-column';
				addColBtn.style.cssText = 'min-width: 250px; display: flex; align-items: center; justify-content: center;';
				addColBtn.innerHTML = '<button onclick="send_event(\'{{.IdComponent}}\', \'AddColumn\', \'\')" style="padding: 15px 30px; border: 2px dashed #bdc3c7; background: transparent; border-radius: 5px; cursor: pointer; color: #7f8c8d; font-size: 16px;">+ Add Column</button>';
				columnsContainer.appendChild(addColBtn);
			}
		}, 500);
	});
	</script>
	`
	
	// Combine templates
	return baseTemplate + modalTemplate
}

// Event Handlers for Modal

// EditCard opens modal to edit a card
func (k *EnhancedKanban) EditCard(data interface{}) {
	cardID := ""
	if id, ok := data.(string); ok {
		cardID = id
	}
	
	// Find the card
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
			if k.FormCardPriority == "" {
				k.FormCardPriority = "medium"
			}
			break
		}
	}
	
	k.Commit()
}

// AddCard opens modal to add a new card
func (k *EnhancedKanban) AddCard(data interface{}) {
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
	
	k.Commit()
}

// EditColumn opens modal to edit a column
func (k *EnhancedKanban) EditColumn(data interface{}) {
	columnID := ""
	if id, ok := data.(string); ok {
		columnID = id
	}
	
	// Find the column
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
func (k *EnhancedKanban) AddColumn(data interface{}) {
	k.ShowModal = true
	k.ModalType = "add_column"
	k.ModalTitle = "Add New Column"
	k.FormColumnID = ""
	k.FormColumnTitle = ""
	k.FormColumnColor = "#e3e8ef"
	
	k.Commit()
}

// UpdateFormField updates a form field in the modal
func (k *EnhancedKanban) UpdateFormField(data interface{}) {
	var event map[string]interface{}
	if jsonData, ok := data.(string); ok {
		json.Unmarshal([]byte(jsonData), &event)
		
		field := event["field"].(string)
		value := event["value"].(string)
		
		switch field {
		case "card_title":
			k.FormCardTitle = value
		case "card_desc":
			k.FormCardDesc = value
		case "card_column":
			k.FormCardColumn = value
		case "card_priority":
			k.FormCardPriority = value
		case "column_title":
			k.FormColumnTitle = value
		case "column_color":
			k.FormColumnColor = value
		}
	}
}

// CloseModal closes the modal
func (k *EnhancedKanban) CloseModal(data interface{}) {
	k.ShowModal = false
	k.Commit()
}

// SaveModal saves the modal form
func (k *EnhancedKanban) SaveModal(data interface{}) {
	switch k.ModalType {
	case "edit_card":
		// Update existing card
		for i := range k.Cards {
			if k.Cards[i].ID == k.FormCardID {
				k.Cards[i].Title = k.FormCardTitle
				k.Cards[i].Description = k.FormCardDesc
				k.Cards[i].ColumnID = k.FormCardColumn
				k.Cards[i].Priority = k.FormCardPriority
				k.Cards[i].UpdatedAt = time.Now()
				break
			}
		}
		
	case "add_card":
		// Create new card
		newCard := components.KanbanCard{
			ID:          fmt.Sprintf("card_%d", time.Now().UnixNano()),
			Title:       k.FormCardTitle,
			Description: k.FormCardDesc,
			ColumnID:    k.FormCardColumn,
			Priority:    k.FormCardPriority,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		k.Cards = append(k.Cards, newCard)
		
	case "edit_column":
		// Update existing column
		for i := range k.Columns {
			if k.Columns[i].ID == k.FormColumnID {
				k.Columns[i].Title = k.FormColumnTitle
				k.Columns[i].Color = k.FormColumnColor
				break
			}
		}
		
	case "add_column":
		// Create new column
		maxOrder := 0
		for _, col := range k.Columns {
			if col.Order > maxOrder {
				maxOrder = col.Order
			}
		}
		
		newColumn := components.KanbanColumn{
			ID:    fmt.Sprintf("col_%d", time.Now().UnixNano()),
			Title: k.FormColumnTitle,
			Color: k.FormColumnColor,
			Order: maxOrder + 1,
		}
		k.Columns = append(k.Columns, newColumn)
	}
	
	k.ShowModal = false
	k.Commit()
}

// DeleteCard deletes the current card being edited
func (k *EnhancedKanban) DeleteCard(data interface{}) {
	newCards := []components.KanbanCard{}
	for _, card := range k.Cards {
		if card.ID != k.FormCardID {
			newCards = append(newCards, card)
		}
	}
	k.Cards = newCards
	
	k.ShowModal = false
	k.Commit()
}

// DeleteColumn deletes the current column being edited
func (k *EnhancedKanban) DeleteColumn(data interface{}) {
	// First, delete all cards in this column
	newCards := []components.KanbanCard{}
	for _, card := range k.Cards {
		if card.ColumnID != k.FormColumnID {
			newCards = append(newCards, card)
		}
	}
	k.Cards = newCards
	
	// Then delete the column
	newColumns := []components.KanbanColumn{}
	for _, col := range k.Columns {
		if col.ID != k.FormColumnID {
			newColumns = append(newColumns, col)
		}
	}
	k.Columns = newColumns
	
	k.ShowModal = false
	k.Commit()
}