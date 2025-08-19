package components

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/arturoeanton/go-echo-live-view/liveview"
)

// KanbanBoard provides a collaborative task management board
type KanbanBoard struct {
	*liveview.ComponentDriver[*KanbanBoard]
	*liveview.CollaborativeComponent

	// Board structure
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Columns     []KanbanColumn `json:"columns"`
	Cards       []KanbanCard   `json:"cards"`
	Labels      []KanbanLabel  `json:"labels"`

	// UI state
	DraggedCard   *KanbanCard `json:"-"`
	SelectedCard  *KanbanCard `json:"-"`
	ShowCardModal bool        `json:"show_modal"`
	EditingCard   *KanbanCard `json:"editing_card"`
	Filter        string      `json:"filter"`
	SearchQuery   string      `json:"search_query"`

	// Collaboration
	ActiveUsers map[string]*UserActivity `json:"active_users"`
}

// KanbanColumn represents a column in the board
type KanbanColumn struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Color     string `json:"color"`
	Order     int    `json:"order"`
	WIPLimit  int    `json:"wip_limit"` // Work in progress limit
	Collapsed bool   `json:"collapsed"`
}

// KanbanCard represents a task card
type KanbanCard struct {
	ID           string     `json:"id"`
	Title        string     `json:"title"`
	Description  string     `json:"description"`
	ColumnID     string     `json:"column_id"`
	AssigneeID   string     `json:"assignee_id"`
	AssigneeName string     `json:"assignee_name"`
	Labels       []string   `json:"labels"`
	DueDate      *time.Time `json:"due_date,omitempty"`
	Priority     string     `json:"priority"` // low, medium, high, urgent
	Points       int        `json:"points"`   // Story points
	Order        int        `json:"order"`
	CreatedBy    string     `json:"created_by"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	Comments     []Comment  `json:"comments"`
	Attachments  []string   `json:"attachments"`
	Completed    bool       `json:"completed"`
	Blocked      bool       `json:"blocked"`
	BlockReason  string     `json:"block_reason,omitempty"`
}

// KanbanLabel for categorizing cards
type KanbanLabel struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

// Comment on a card
type Comment struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	UserName  string    `json:"user_name"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
}

// UserActivity tracks what users are doing
type UserActivity struct {
	UserID     string    `json:"user_id"`
	UserName   string    `json:"user_name"`
	CardID     string    `json:"card_id,omitempty"`
	ColumnID   string    `json:"column_id,omitempty"`
	Activity   string    `json:"activity"` // viewing, editing, dragging
	LastActive time.Time `json:"last_active"`
	Color      string    `json:"color"`
}

// Start initializes the Kanban board
func (k *KanbanBoard) Start() {
	// Initialize collaboration first - use fixed room ID so all users join same room
	if k.CollaborativeComponent == nil {
		k.CollaborativeComponent = &liveview.CollaborativeComponent{}
	}
	k.CollaborativeComponent.Driver = k.ComponentDriver
	roomID := "kanban_shared_room" // Fixed room ID for all users
	userID := fmt.Sprintf("user_%d", time.Now().UnixNano()) // Unique user ID
	k.StartCollaboration(roomID, userID, fmt.Sprintf("User-%d", time.Now().UnixNano()%10000))

	// Get shared state from room if it exists
	if k.Room != nil {
		sharedState := k.Room.GetState()
		fmt.Printf("[DEBUG] SharedState from room: %+v\n", sharedState)
		if boardData, ok := sharedState.(map[string]interface{}); ok && len(boardData) > 0 {
			// Load from shared state
			fmt.Printf("[DEBUG] Loading from shared state with %d items\n", len(boardData))
			k.loadFromSharedState(boardData)
		} else {
			// Initialize default state and save to room
			fmt.Printf("[DEBUG] No shared state found, initializing default\n")
			k.initializeDefaultState()
			k.saveToSharedState()
		}
	} else {
		// Fallback: initialize default state
		fmt.Printf("[DEBUG] No room available, using default state\n")
		k.initializeDefaultState()
	}

	k.ActiveUsers = make(map[string]*UserActivity)
	k.Commit()
}

// initializeDefaultState sets up the default board structure
func (k *KanbanBoard) initializeDefaultState() {
	k.Title = "Project Board"
	k.Description = "Collaborative task management"

	// Default columns
	k.Columns = []KanbanColumn{
		{ID: "backlog", Title: "Backlog", Color: "#95a5a6", Order: 0},
		{ID: "todo", Title: "To Do", Color: "#3498db", Order: 1, WIPLimit: 5},
		{ID: "progress", Title: "In Progress", Color: "#f39c12", Order: 2, WIPLimit: 3},
		{ID: "review", Title: "Review", Color: "#9b59b6", Order: 3, WIPLimit: 2},
		{ID: "done", Title: "Done", Color: "#27ae60", Order: 4},
	}

	// Sample cards
	k.Cards = []KanbanCard{
		{
			ID:          "card1",
			Title:       "Setup project structure",
			Description: "Initialize the project with proper folder structure",
			ColumnID:    "done",
			Priority:    "high",
			Points:      3,
			Order:       0,
			CreatedAt:   time.Now().Add(-48 * time.Hour),
			UpdatedAt:   time.Now().Add(-24 * time.Hour),
			Completed:   true,
		},
		{
			ID:          "card2",
			Title:       "Design database schema",
			Description: "Create ERD and define all tables",
			ColumnID:    "progress",
			Priority:    "high",
			Points:      5,
			Order:       0,
			CreatedAt:   time.Now().Add(-24 * time.Hour),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          "card3",
			Title:       "Implement authentication",
			Description: "Add user login and registration",
			ColumnID:    "todo",
			Priority:    "medium",
			Points:      8,
			Order:       0,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	// Default labels
	k.Labels = []KanbanLabel{
		{ID: "bug", Name: "Bug", Color: "#e74c3c"},
		{ID: "feature", Name: "Feature", Color: "#3498db"},
		{ID: "enhancement", Name: "Enhancement", Color: "#2ecc71"},
		{ID: "documentation", Name: "Documentation", Color: "#95a5a6"},
		{ID: "urgent", Name: "Urgent", Color: "#e67e22"},
	}
}

// saveToSharedState saves current board state to the collaboration room
func (k *KanbanBoard) saveToSharedState() {
	if k.Room != nil {
		boardState := map[string]interface{}{
			"title":       k.Title,
			"description": k.Description,
			"columns":     k.Columns,
			"cards":       k.Cards,
			"labels":      k.Labels,
		}
		k.SyncState("board_update", boardState)
	}
}

// loadFromSharedState loads board state from shared room state
func (k *KanbanBoard) loadFromSharedState(boardData map[string]interface{}) {
	if title, ok := boardData["title"].(string); ok {
		k.Title = title
	}
	if desc, ok := boardData["description"].(string); ok {
		k.Description = desc
	}
	
	// Load columns
	if columnsData, ok := boardData["columns"]; ok {
		if columnsJSON, err := json.Marshal(columnsData); err == nil {
			var columns []KanbanColumn
			if err := json.Unmarshal(columnsJSON, &columns); err == nil {
				k.Columns = columns
			}
		}
	}
	
	// Load cards
	if cardsData, ok := boardData["cards"]; ok {
		if cardsJSON, err := json.Marshal(cardsData); err == nil {
			var cards []KanbanCard
			if err := json.Unmarshal(cardsJSON, &cards); err == nil {
				k.Cards = cards
			}
		}
	}
	
	// Load labels
	if labelsData, ok := boardData["labels"]; ok {
		if labelsJSON, err := json.Marshal(labelsData); err == nil {
			var labels []KanbanLabel
			if err := json.Unmarshal(labelsJSON, &labels); err == nil {
				k.Labels = labels
			}
		}
	}
}

// GetTemplate returns the Kanban board HTML
func (k *KanbanBoard) GetTemplate() string {
	return `
	<div class="kanban-board" id="{{.IdComponent}}">
		<style>
			.kanban-board {
				font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
				padding: 20px;
				background: #f5f7fa;
				min-height: 100vh;
			}
			
			.board-header {
				display: flex;
				justify-content: space-between;
				align-items: center;
				margin-bottom: 20px;
				padding: 20px;
				background: white;
				border-radius: 8px;
				box-shadow: 0 2px 4px rgba(0,0,0,0.1);
			}
			
			.board-title {
				font-size: 24px;
				font-weight: 600;
				color: #2c3e50;
			}
			
			.board-controls {
				display: flex;
				gap: 10px;
				align-items: center;
			}
			
			.search-box {
				padding: 8px 12px;
				border: 1px solid #ddd;
				border-radius: 4px;
				width: 250px;
			}
			
			.btn {
				padding: 8px 16px;
				background: #3498db;
				color: white;
				border: none;
				border-radius: 4px;
				cursor: pointer;
				font-size: 14px;
				transition: background 0.2s;
			}
			
			.btn:hover {
				background: #2980b9;
			}
			
			.board-columns {
				display: flex;
				gap: 20px;
				overflow-x: auto;
				padding-bottom: 20px;
			}
			
			.kanban-column {
				min-width: 300px;
				background: #e8ecef;
				border-radius: 8px;
				padding: 15px;
			}
			
			.column-header {
				display: flex;
				justify-content: space-between;
				align-items: center;
				margin-bottom: 15px;
				padding-bottom: 10px;
				border-bottom: 2px solid;
			}
			
			.column-title {
				font-weight: 600;
				font-size: 16px;
				color: #2c3e50;
			}
			
			.column-count {
				background: rgba(0,0,0,0.1);
				padding: 2px 8px;
				border-radius: 12px;
				font-size: 12px;
			}
			
			.cards-container {
				min-height: 100px;
				transition: background 0.2s;
			}
			
			.cards-container.drag-over {
				background: rgba(52, 152, 219, 0.1);
				border: 2px dashed #3498db;
				border-radius: 4px;
			}
			
			.kanban-card {
				background: white;
				border-radius: 6px;
				padding: 12px;
				margin-bottom: 10px;
				box-shadow: 0 2px 4px rgba(0,0,0,0.1);
				cursor: move;
				transition: all 0.2s;
				position: relative;
			}
			
			.kanban-card:hover {
				box-shadow: 0 4px 8px rgba(0,0,0,0.15);
				transform: translateY(-2px);
			}
			
			.kanban-card.dragging {
				opacity: 0.5;
				transform: rotate(2deg);
			}
			
			.card-priority {
				position: absolute;
				left: 0;
				top: 0;
				bottom: 0;
				width: 4px;
				border-radius: 6px 0 0 6px;
			}
			
			.priority-urgent { background: #e74c3c; }
			.priority-high { background: #e67e22; }
			.priority-medium { background: #f39c12; }
			.priority-low { background: #95a5a6; }
			
			.card-title {
				font-weight: 500;
				color: #2c3e50;
				margin-bottom: 8px;
				padding-left: 8px;
			}
			
			.card-meta {
				display: flex;
				gap: 10px;
				align-items: center;
				font-size: 12px;
				color: #7f8c8d;
				padding-left: 8px;
			}
			
			.card-labels {
				display: flex;
				gap: 4px;
				margin-top: 8px;
				padding-left: 8px;
			}
			
			.label {
				padding: 2px 6px;
				border-radius: 3px;
				font-size: 11px;
				color: white;
			}
			
			.card-assignee {
				display: flex;
				align-items: center;
				gap: 5px;
				margin-top: 8px;
				padding-left: 8px;
			}
			
			.assignee-avatar {
				width: 24px;
				height: 24px;
				border-radius: 50%;
				background: #3498db;
				color: white;
				display: flex;
				align-items: center;
				justify-content: center;
				font-size: 11px;
				font-weight: 600;
			}
			
			.card-blocked {
				background: #ffebee;
				border: 1px solid #ef5350;
			}
			
			.blocked-badge {
				background: #ef5350;
				color: white;
				padding: 2px 6px;
				border-radius: 3px;
				font-size: 11px;
				margin-left: 8px;
			}
			
			.active-users {
				position: fixed;
				bottom: 20px;
				right: 20px;
				background: white;
				border-radius: 8px;
				padding: 15px;
				box-shadow: 0 4px 12px rgba(0,0,0,0.15);
				z-index: 1000;
			}
			
			.active-user {
				display: flex;
				align-items: center;
				gap: 10px;
				margin-bottom: 10px;
			}
			
			.user-indicator {
				width: 8px;
				height: 8px;
				border-radius: 50%;
				background: #2ecc71;
				animation: pulse 2s infinite;
			}
			
			@keyframes pulse {
				0% { opacity: 1; }
				50% { opacity: 0.5; }
				100% { opacity: 1; }
			}
			
			.add-card-btn {
				width: 100%;
				padding: 8px;
				border: 2px dashed #bdc3c7;
				background: transparent;
				border-radius: 4px;
				color: #7f8c8d;
				cursor: pointer;
				transition: all 0.2s;
			}
			
			.add-card-btn:hover {
				border-color: #3498db;
				color: #3498db;
				background: rgba(52, 152, 219, 0.05);
			}
		</style>
		
		<!-- Board Header -->
		<div class="board-header">
			<div>
				<h1 class="board-title">{{.Title}}</h1>
				<p style="color: #7f8c8d; margin-top: 5px;">{{.Description}}</p>
			</div>
			<div class="board-controls">
				<input type="text" 
				       class="search-box" 
				       placeholder="Search cards..." 
				       value="{{.SearchQuery}}"
				       onkeyup="send_event('{{.IdComponent}}', 'Search', this.value)">
				
				<select onchange="send_event('{{.IdComponent}}', 'FilterBy', this.value)">
					<option value="">All Cards</option>
					<option value="my">My Cards</option>
					<option value="urgent">Urgent</option>
					<option value="blocked">Blocked</option>
				</select>
				
				<button class="btn" onclick="send_event('{{.IdComponent}}', 'AddCard', 'backlog')">
					+ New Card
				</button>
			</div>
		</div>
		
		<!-- Kanban Columns -->
		<div class="board-columns">
			{{range .Columns}}
			<div class="kanban-column" 
			     style="border-top-color: {{.Color}};"
			     ondrop="event.preventDefault(); var cardId = event.dataTransfer.getData('cardId'); if(cardId) { send_event('{{$.IdComponent}}', 'MoveCard', JSON.stringify({cardId: cardId, columnId: '{{.ID}}'})); }"
			     ondragover="event.preventDefault(); event.currentTarget.querySelector('.cards-container').classList.add('drag-over');"
			     ondragleave="event.currentTarget.querySelector('.cards-container').classList.remove('drag-over');">
				
				<div class="column-header" style="border-bottom-color: {{.Color}};">
					<span class="column-title">{{.Title}}</span>
					<span class="column-count">
						{{$.GetCardCount .ID}}{{if .WIPLimit}}/{{.WIPLimit}}{{end}}
					</span>
				</div>
				
				<div class="cards-container" data-column-id="{{.ID}}">
					{{range $.GetCardsForColumn .ID}}
					<div class="kanban-card {{if .Blocked}}card-blocked{{end}}" 
					     draggable="true"
					     data-card-id="{{.ID}}"
					     ondragstart="event.dataTransfer.setData('cardId', '{{.ID}}'); event.target.classList.add('dragging');"
					     ondragend="event.target.classList.remove('dragging');"
					     onclick="send_event('{{$.IdComponent}}', 'SelectCard', '{{.ID}}')">
						
						<div class="card-priority priority-{{.Priority}}"></div>
						
						<div class="card-title">
							{{.Title}}
							{{if .Blocked}}<span class="blocked-badge">BLOCKED</span>{{end}}
						</div>
						
						{{if .Description}}
						<div style="color: #7f8c8d; font-size: 13px; margin: 5px 0; padding-left: 8px;">
							{{.Description}}
						</div>
						{{end}}
						
						<div class="card-meta">
							{{if .Points}}<span>{{.Points}} pts</span>{{end}}
							{{if .DueDate}}<span>📅 {{.FormatDueDate}}</span>{{end}}
							{{if .Comments}}<span>💬 {{len .Comments}}</span>{{end}}
						</div>
						
						{{if .Labels}}
						<div class="card-labels">
							{{range .Labels}}
								{{$label := $.GetLabel .}}
								{{if $label}}
								<span class="label" style="background: {{$label.Color}};">
									{{$label.Name}}
								</span>
								{{end}}
							{{end}}
						</div>
						{{end}}
						
						{{if .AssigneeName}}
						<div class="card-assignee">
							<div class="assignee-avatar">
								{{index .AssigneeName 0}}
							</div>
							<span style="font-size: 12px; color: #7f8c8d;">{{.AssigneeName}}</span>
						</div>
						{{end}}
					</div>
					{{end}}
				</div>
				
				<button class="add-card-btn" 
				        onclick="send_event('{{$.IdComponent}}', 'AddCard', '{{.ID}}')">
					+ Add Card
				</button>
			</div>
			{{end}}
		</div>
		
		<!-- Active Users -->
		{{if .ActiveUsers}}
		<div class="active-users">
			<div style="font-weight: 600; margin-bottom: 10px;">Active Now</div>
			{{range .ActiveUsers}}
			<div class="active-user">
				<div class="user-indicator"></div>
				<div class="assignee-avatar" style="background: {{.Color}};">
					{{index .UserName 0}}
				</div>
				<div>
					<div style="font-size: 13px;">{{.UserName}}</div>
					<div style="font-size: 11px; color: #7f8c8d;">{{.Activity}}</div>
				</div>
			</div>
			{{end}}
		</div>
		{{end}}
		
	</div>
	`
}

// GetDriver returns the component driver
func (k *KanbanBoard) GetDriver() liveview.LiveDriver {
	return k.ComponentDriver
}

// GetCardsForColumn returns cards in a specific column
func (k *KanbanBoard) GetCardsForColumn(columnID string) []KanbanCard {
	cards := make([]KanbanCard, 0)
	for _, card := range k.Cards {
		if card.ColumnID == columnID {
			// Apply filters
			if k.SearchQuery != "" && !k.cardMatchesSearch(card) {
				continue
			}
			if k.Filter != "" && !k.cardMatchesFilter(card) {
				continue
			}
			cards = append(cards, card)
		}
	}

	// Sort by order
	for i := 0; i < len(cards)-1; i++ {
		for j := i + 1; j < len(cards); j++ {
			if cards[i].Order > cards[j].Order {
				cards[i], cards[j] = cards[j], cards[i]
			}
		}
	}

	return cards
}

// GetCardCount returns the number of cards in a column
func (k *KanbanBoard) GetCardCount(columnID string) int {
	count := 0
	for _, card := range k.Cards {
		if card.ColumnID == columnID {
			count++
		}
	}
	return count
}

// GetLabel returns a label by ID
func (k *KanbanBoard) GetLabel(labelID string) *KanbanLabel {
	for _, label := range k.Labels {
		if label.ID == labelID {
			return &label
		}
	}
	return nil
}

// cardMatchesSearch checks if a card matches the search query
func (k *KanbanBoard) cardMatchesSearch(card KanbanCard) bool {
	query := strings.ToLower(k.SearchQuery)
	return strings.Contains(strings.ToLower(card.Title), query) ||
		strings.Contains(strings.ToLower(card.Description), query)
}

// cardMatchesFilter checks if a card matches the current filter
func (k *KanbanBoard) cardMatchesFilter(card KanbanCard) bool {
	switch k.Filter {
	case "my":
		return card.AssigneeID == k.UserID
	case "urgent":
		return card.Priority == "urgent"
	case "blocked":
		return card.Blocked
	default:
		return true
	}
}

// Search updates the search query
func (k *KanbanBoard) Search(data interface{}) {
	if query, ok := data.(string); ok {
		k.SearchQuery = query
		k.Commit()
	}
}

// FilterBy updates the filter
func (k *KanbanBoard) FilterBy(data interface{}) {
	if filter, ok := data.(string); ok {
		k.Filter = filter
		k.Commit()
	}
}

// AddCard adds a new card to a column
func (k *KanbanBoard) AddCard(data interface{}) {
	columnID := "backlog"
	if id, ok := data.(string); ok {
		columnID = id
	}

	newCard := KanbanCard{
		ID:        fmt.Sprintf("card_%d", time.Now().UnixNano()),
		Title:     "New Task",
		ColumnID:  columnID,
		Priority:  "medium",
		Order:     len(k.GetCardsForColumn(columnID)),
		CreatedBy: k.UserID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	k.Cards = append(k.Cards, newCard)
	k.EditingCard = &newCard
	k.ShowCardModal = true

	// Sync with collaboration
	k.SyncState("add_card", newCard)
	k.BroadcastAction("card_added", newCard)

	k.Commit()
}

// SelectCard selects a card for viewing/editing
func (k *KanbanBoard) SelectCard(data interface{}) {
	if cardID, ok := data.(string); ok {
		for i := range k.Cards {
			if k.Cards[i].ID == cardID {
				k.SelectedCard = &k.Cards[i]
				k.EditingCard = &k.Cards[i]
				k.ShowCardModal = true

				// Update activity
				k.updateUserActivity("viewing", cardID)

				k.Commit()
				break
			}
		}
	}
}

// MoveCard moves a card to a different column
func (k *KanbanBoard) MoveCard(data interface{}) {
	fmt.Printf("[DEBUG] MoveCard called with data: %+v\n", data)
	
	// Handle both JSON string and map
	var moveData map[string]interface{}
	
	if jsonStr, ok := data.(string); ok {
		// Parse JSON string
		if err := json.Unmarshal([]byte(jsonStr), &moveData); err != nil {
			fmt.Printf("[ERROR] Failed to parse JSON: %v\n", err)
			return
		}
	} else if md, ok := data.(map[string]interface{}); ok {
		moveData = md
	} else {
		fmt.Printf("[ERROR] Invalid data type: %T\n", data)
		return
	}
	
	cardID, _ := moveData["cardId"].(string)
	newColumnID, _ := moveData["columnId"].(string)
	
	fmt.Printf("[DEBUG] Moving card %s to column %s\n", cardID, newColumnID)
	
	if cardID != "" && newColumnID != "" {
		// Find and update card
		for i := range k.Cards {
			if k.Cards[i].ID == cardID {
				oldColumnID := k.Cards[i].ColumnID
				k.Cards[i].ColumnID = newColumnID
				k.Cards[i].UpdatedAt = time.Now()

				// Update order
				k.Cards[i].Order = len(k.GetCardsForColumn(newColumnID))

				// Sync with collaboration
				k.SyncState("move_card", map[string]interface{}{
					"card_id":     cardID,
					"from_column": oldColumnID,
					"to_column":   newColumnID,
				})

				// Broadcast to other users
				broadcastData := map[string]interface{}{
					"card_id":     cardID,
					"from_column": oldColumnID,
					"to_column":   newColumnID,
					"user_id":     k.UserID,
				}
				fmt.Printf("[DEBUG] Broadcasting card_moved: %+v\n", broadcastData)
				k.BroadcastAction("card_moved", broadcastData)

				// Update activity
				k.updateUserActivity("moved card", cardID)

				// Save updated state to room
				k.saveToSharedState()

				k.Commit()
				break
			}
		}
	}
}

// StartDrag handles drag start
func (k *KanbanBoard) StartDrag(data interface{}) {
	if cardID, ok := data.(string); ok {
		for i := range k.Cards {
			if k.Cards[i].ID == cardID {
				k.DraggedCard = &k.Cards[i]

				// Update activity
				k.updateUserActivity("dragging", cardID)

				// Broadcast to show who's dragging
				k.BroadcastAction("user_dragging", map[string]interface{}{
					"user_id": k.UserID,
					"card_id": cardID,
				})
				break
			}
		}
	}
}

// UpdateCard updates card details
func (k *KanbanBoard) UpdateCard(data interface{}) {
	if updateData, ok := data.(map[string]interface{}); ok {
		cardID := updateData["id"].(string)

		for i := range k.Cards {
			if k.Cards[i].ID == cardID {
				// Update fields
				if title, ok := updateData["title"].(string); ok {
					k.Cards[i].Title = title
				}
				if desc, ok := updateData["description"].(string); ok {
					k.Cards[i].Description = desc
				}
				if priority, ok := updateData["priority"].(string); ok {
					k.Cards[i].Priority = priority
				}

				k.Cards[i].UpdatedAt = time.Now()

				// Sync and broadcast
				k.SyncState("update_card", k.Cards[i])
				k.BroadcastAction("card_updated", k.Cards[i])

				k.Commit()
				break
			}
		}
	}
}

// AddComment adds a comment to a card
func (k *KanbanBoard) AddComment(data interface{}) {
	if commentData, ok := data.(map[string]interface{}); ok {
		cardID := commentData["card_id"].(string)
		text := commentData["text"].(string)

		for i := range k.Cards {
			if k.Cards[i].ID == cardID {
				comment := Comment{
					ID:        fmt.Sprintf("comment_%d", time.Now().UnixNano()),
					UserID:    k.UserID,
					UserName:  k.UserName,
					Text:      text,
					CreatedAt: time.Now(),
				}

				k.Cards[i].Comments = append(k.Cards[i].Comments, comment)
				k.Cards[i].UpdatedAt = time.Now()

				// Sync and broadcast
				k.SyncState("add_comment", map[string]interface{}{
					"card_id": cardID,
					"comment": comment,
				})
				k.BroadcastAction("comment_added", map[string]interface{}{
					"card_id": cardID,
					"comment": comment,
				})

				k.Commit()
				break
			}
		}
	}
}

// updateUserActivity updates the current user's activity
func (k *KanbanBoard) updateUserActivity(activity string, cardID string) {
	if k.ActiveUsers == nil {
		k.ActiveUsers = make(map[string]*UserActivity)
	}

	k.ActiveUsers[k.UserID] = &UserActivity{
		UserID:     k.UserID,
		UserName:   k.UserName,
		CardID:     cardID,
		Activity:   activity,
		LastActive: time.Now(),
		Color:      k.getUserColor(),
	}

	// Broadcast activity
	k.BroadcastAction("user_activity", k.ActiveUsers[k.UserID])
}

// getUserColor generates a consistent color for the user
func (k *KanbanBoard) getUserColor() string {
	colors := []string{"#3498db", "#e74c3c", "#2ecc71", "#f39c12", "#9b59b6"}
	hash := 0
	for _, ch := range k.UserID {
		hash = (hash + int(ch)) % len(colors)
	}
	return colors[hash]
}

// HandleCollaborationMessage handles incoming collaboration messages
func (k *KanbanBoard) HandleCollaborationMessage(data interface{}) {
	fmt.Printf("[DEBUG] HandleCollaborationMessage called with: %+v\n", data)
	if msgJSON, ok := data.(string); ok {
		var msg map[string]interface{}
		if err := json.Unmarshal([]byte(msgJSON), &msg); err == nil {
			fmt.Printf("[DEBUG] Processing collaboration message: %+v\n", msg)
			k.processCollaborativeMessage(msg)
		} else {
			fmt.Printf("[ERROR] Failed to unmarshal collaboration message: %v\n", err)
		}
	} else {
		fmt.Printf("[ERROR] Invalid collaboration message type: %T\n", data)
	}
}

// processCollaborativeMessage processes different types of collaborative messages
func (k *KanbanBoard) processCollaborativeMessage(msg map[string]interface{}) {
	msgType, _ := msg["type"].(string)
	from, _ := msg["from"].(string)
	
	// Don't process our own messages
	if from == k.UserID {
		return
	}
	
	switch msgType {
	case "card_moved":
		if moveData, ok := msg["data"].(map[string]interface{}); ok {
			cardID, _ := moveData["card_id"].(string)
			toColumn, _ := moveData["to_column"].(string)
			
			// Update card position
			for i := range k.Cards {
				if k.Cards[i].ID == cardID {
					k.Cards[i].ColumnID = toColumn
					k.Cards[i].UpdatedAt = time.Now()
					k.Commit() // Refresh UI
					break
				}
			}
		}
	case "card_added":
		if cardData, ok := msg["data"].(map[string]interface{}); ok {
			var card KanbanCard
			if cardJSON, err := json.Marshal(cardData); err == nil {
				if err := json.Unmarshal(cardJSON, &card); err == nil {
					k.updateOrAddCard(card)
					k.Commit()
				}
			}
		}
	case "user_activity":
		if activityData, ok := msg["data"].(map[string]interface{}); ok {
			var activity UserActivity
			if activityJSON, err := json.Marshal(activityData); err == nil {
				if err := json.Unmarshal(activityJSON, &activity); err == nil {
					k.ActiveUsers[activity.UserID] = &activity
					k.Commit()
				}
			}
		}
	}
}

// HandleRemoteUpdate handles updates from other users
func (k *KanbanBoard) HandleRemoteUpdate(data interface{}) {
	if updateData, ok := data.(map[string]interface{}); ok {
		updateType := updateData["type"].(string)

		switch updateType {
		case "card_added", "card_updated":
			// Update card from remote user
			if cardData, err := json.Marshal(updateData["card"]); err == nil {
				var card KanbanCard
				if err := json.Unmarshal(cardData, &card); err == nil {
					k.updateOrAddCard(card)
				}
			}

		case "card_moved":
			// Move card based on remote action
			cardID := updateData["card_id"].(string)
			toColumn := updateData["to_column"].(string)

			for i := range k.Cards {
				if k.Cards[i].ID == cardID {
					k.Cards[i].ColumnID = toColumn
					k.Cards[i].UpdatedAt = time.Now()
					break
				}
			}

		case "user_activity":
			// Update other user's activity
			if activityData, err := json.Marshal(updateData["activity"]); err == nil {
				var activity UserActivity
				if err := json.Unmarshal(activityData, &activity); err == nil {
					k.ActiveUsers[activity.UserID] = &activity
				}
			}
		}

		k.Commit()
	}
}

// updateOrAddCard updates an existing card or adds a new one
func (k *KanbanBoard) updateOrAddCard(card KanbanCard) {
	found := false
	for i := range k.Cards {
		if k.Cards[i].ID == card.ID {
			k.Cards[i] = card
			found = true
			break
		}
	}

	if !found {
		k.Cards = append(k.Cards, card)
	}
}

// ExportBoard exports the board as JSON
func (k *KanbanBoard) ExportBoard(data interface{}) {
	boardData := map[string]interface{}{
		"title":       k.Title,
		"description": k.Description,
		"columns":     k.Columns,
		"cards":       k.Cards,
		"labels":      k.Labels,
	}

	if jsonData, err := json.Marshal(boardData); err == nil {
		// Send export data via property update
		k.ComponentDriver.SetPropertie("exportData", string(jsonData))
	}
}

// ImportBoard imports board data from JSON
func (k *KanbanBoard) ImportBoard(data interface{}) {
	if jsonStr, ok := data.(string); ok {
		var boardData map[string]interface{}
		if err := json.Unmarshal([]byte(jsonStr), &boardData); err == nil {
			// Import board structure
			if title, ok := boardData["title"].(string); ok {
				k.Title = title
			}
			if desc, ok := boardData["description"].(string); ok {
				k.Description = desc
			}

			// Import columns, cards, and labels
			// (Implementation would deserialize these arrays)

			k.SyncState("import_board", boardData)
			k.BroadcastAction("board_imported", nil)
			k.Commit()
		}
	}
}
