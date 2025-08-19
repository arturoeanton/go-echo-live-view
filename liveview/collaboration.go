package liveview

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// CollaborationHub manages real-time collaboration features
type CollaborationHub struct {
	// Rooms store collaborative sessions
	rooms map[string]*CollaborationRoom
	mu    sync.RWMutex
}

// CollaborationRoom represents a collaborative space
type CollaborationRoom struct {
	ID           string
	Name         string
	Participants map[string]*Participant
	SharedState  interface{}
	History      []StateChange
	mu           sync.RWMutex
	broadcast    chan BroadcastMessage
	join         chan *Participant
	leave        chan string
	ctx          context.Context
	cancel       context.CancelFunc
}

// Participant represents a user in a collaborative session
type Participant struct {
	ID       string                 `json:"id"`
	Name     string                 `json:"name"`
	Color    string                 `json:"color"`
	Cursor   *CursorPosition        `json:"cursor,omitempty"`
	Metadata map[string]interface{} `json:"metadata"`
	LastSeen time.Time              `json:"last_seen"`
	Driver   LiveDriver             `json:"-"`
}

// CursorPosition tracks cursor location
type CursorPosition struct {
	X         float64 `json:"x"`
	Y         float64 `json:"y"`
	ElementID string  `json:"element_id,omitempty"`
}

// StateChange represents a change in shared state
type StateChange struct {
	ID        string      `json:"id"`
	Timestamp time.Time   `json:"timestamp"`
	UserID    string      `json:"user_id"`
	Action    string      `json:"action"`
	Data      interface{} `json:"data"`
	OldValue  interface{} `json:"old_value,omitempty"`
	NewValue  interface{} `json:"new_value,omitempty"`
}

// BroadcastMessage for room-wide communication
type BroadcastMessage struct {
	Type      string      `json:"type"`
	From      string      `json:"from"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
}

// Global collaboration hub
var globalHub = &CollaborationHub{
	rooms: make(map[string]*CollaborationRoom),
}

// GetCollaborationHub returns the global collaboration hub
func GetCollaborationHub() *CollaborationHub {
	return globalHub
}

// CreateRoom creates a new collaboration room
func (h *CollaborationHub) CreateRoom(id, name string) *CollaborationRoom {
	h.mu.Lock()
	defer h.mu.Unlock()

	ctx, cancel := context.WithCancel(context.Background())

	room := &CollaborationRoom{
		ID:           id,
		Name:         name,
		Participants: make(map[string]*Participant),
		History:      make([]StateChange, 0),
		broadcast:    make(chan BroadcastMessage, 100),
		join:         make(chan *Participant, 10),
		leave:        make(chan string, 10),
		ctx:          ctx,
		cancel:       cancel,
	}

	h.rooms[id] = room
	go room.run()

	return room
}

// GetRoom retrieves a room by ID
func (h *CollaborationHub) GetRoom(id string) (*CollaborationRoom, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	room, exists := h.rooms[id]
	return room, exists
}

// RemoveRoom removes a collaboration room
func (h *CollaborationHub) RemoveRoom(id string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if room, exists := h.rooms[id]; exists {
		room.Close()
		delete(h.rooms, id)
	}
}

// run manages the room's event loop
func (r *CollaborationRoom) run() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-r.ctx.Done():
			return

		case participant := <-r.join:
			r.handleJoin(participant)

		case userID := <-r.leave:
			r.handleLeave(userID)

		case msg := <-r.broadcast:
			r.handleBroadcast(msg)

		case <-ticker.C:
			r.checkInactiveParticipants()
		}
	}
}

// handleJoin processes a participant joining
func (r *CollaborationRoom) handleJoin(p *Participant) {
	r.mu.Lock()
	r.Participants[p.ID] = p
	r.mu.Unlock()

	// Notify all participants
	r.BroadcastPresenceUpdate()

	// Send current state to new participant
	if p.Driver != nil {
		state := r.GetState()
		// Send state update to client
		stateJSON, _ := json.Marshal(state)
		p.Driver.SetPropertie("roomState", string(stateJSON))

		// Send participant list
		participantsJSON, _ := json.Marshal(r.GetParticipants())
		p.Driver.SetPropertie("participants", string(participantsJSON))
	}
}

// handleLeave processes a participant leaving
func (r *CollaborationRoom) handleLeave(userID string) {
	r.mu.Lock()
	delete(r.Participants, userID)
	r.mu.Unlock()

	// Notify remaining participants
	r.BroadcastPresenceUpdate()
}

// handleBroadcast sends message to all participants
func (r *CollaborationRoom) handleBroadcast(msg BroadcastMessage) {
	r.mu.RLock()
	participants := make([]*Participant, 0, len(r.Participants))
	for _, p := range r.Participants {
		participants = append(participants, p)
	}
	r.mu.RUnlock()

	// Send to all participants
	for _, p := range participants {
		if p.Driver != nil && p.ID != msg.From {
			// Send message to client via properties
			msgJSON, _ := json.Marshal(msg)
			p.Driver.SetPropertie("collaborationMessage", string(msgJSON))
		}
	}
}

// checkInactiveParticipants removes inactive users
func (r *CollaborationRoom) checkInactiveParticipants() {
	r.mu.Lock()
	defer r.mu.Unlock()

	timeout := 5 * time.Minute
	now := time.Now()

	for id, p := range r.Participants {
		if now.Sub(p.LastSeen) > timeout {
			delete(r.Participants, id)

			// Schedule presence update
			go r.BroadcastPresenceUpdate()
		}
	}
}

// Join adds a participant to the room
func (r *CollaborationRoom) Join(userID, userName string, driver LiveDriver) *Participant {
	colors := []string{"#FF6B6B", "#4ECDC4", "#45B7D1", "#96CEB4", "#FECA57", "#48C9B0", "#9B59B6", "#E74C3C"}

	participant := &Participant{
		ID:       userID,
		Name:     userName,
		Color:    colors[len(r.Participants)%len(colors)],
		Metadata: make(map[string]interface{}),
		LastSeen: time.Now(),
		Driver:   driver,
	}

	select {
	case r.join <- participant:
	case <-time.After(time.Second):
		// Timeout
	}

	return participant
}

// Leave removes a participant from the room
func (r *CollaborationRoom) Leave(userID string) {
	select {
	case r.leave <- userID:
	case <-time.After(time.Second):
		// Timeout
	}
}

// Broadcast sends a message to all participants
func (r *CollaborationRoom) Broadcast(msgType string, from string, data interface{}) {
	msg := BroadcastMessage{
		Type:      msgType,
		From:      from,
		Data:      data,
		Timestamp: time.Now(),
	}

	select {
	case r.broadcast <- msg:
	case <-time.After(time.Second):
		// Timeout
	}
}

// UpdateCursor updates a participant's cursor position
func (r *CollaborationRoom) UpdateCursor(userID string, x, y float64, elementID string) {
	r.mu.Lock()
	if p, exists := r.Participants[userID]; exists {
		p.Cursor = &CursorPosition{
			X:         x,
			Y:         y,
			ElementID: elementID,
		}
		p.LastSeen = time.Now()
	}
	r.mu.Unlock()

	// Broadcast cursor update
	r.Broadcast("cursor_update", userID, map[string]interface{}{
		"user_id": userID,
		"x":       x,
		"y":       y,
		"element": elementID,
	})
}

// UpdateSharedState updates the room's shared state
func (r *CollaborationRoom) UpdateSharedState(userID, action string, data interface{}) {
	change := StateChange{
		ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
		Timestamp: time.Now(),
		UserID:    userID,
		Action:    action,
		Data:      data,
		OldValue:  r.SharedState,
		NewValue:  data,
	}

	r.mu.Lock()
	r.SharedState = data
	r.History = append(r.History, change)

	// Keep only last 1000 history items
	if len(r.History) > 1000 {
		r.History = r.History[len(r.History)-1000:]
	}
	r.mu.Unlock()

	// Broadcast state change
	r.Broadcast("state_change", userID, change)
}

// GetState returns the current shared state
func (r *CollaborationRoom) GetState() interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.SharedState
}

// GetParticipants returns list of participants
func (r *CollaborationRoom) GetParticipants() []Participant {
	r.mu.RLock()
	defer r.mu.RUnlock()

	participants := make([]Participant, 0, len(r.Participants))
	for _, p := range r.Participants {
		participants = append(participants, *p)
	}
	return participants
}

// GetHistory returns state change history
func (r *CollaborationRoom) GetHistory(limit int) []StateChange {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if limit <= 0 || limit > len(r.History) {
		limit = len(r.History)
	}

	start := len(r.History) - limit
	if start < 0 {
		start = 0
	}

	return r.History[start:]
}

// BroadcastPresenceUpdate notifies all participants of presence changes
func (r *CollaborationRoom) BroadcastPresenceUpdate() {
	participants := r.GetParticipants()
	r.Broadcast("presence_update", "", participants)
}

// Close shuts down the room
func (r *CollaborationRoom) Close() {
	if r.cancel != nil {
		r.cancel()
	}
}

// CollaborativeComponent is a base component with collaboration features
type CollaborativeComponent struct {
	Driver       LiveDriver
	RoomID       string
	Room         *CollaborationRoom
	UserID       string
	UserName     string
	Participants []Participant
}

// StartCollaboration initializes collaboration features
func (c *CollaborativeComponent) StartCollaboration(roomID, userID, userName string) {
	hub := GetCollaborationHub()

	// Get or create room
	room, exists := hub.GetRoom(roomID)
	if !exists {
		room = hub.CreateRoom(roomID, fmt.Sprintf("Room %s", roomID))
	}

	c.RoomID = roomID
	c.Room = room
	c.UserID = userID
	c.UserName = userName

	// Join room
	if c.Driver != nil {
		c.Room.Join(userID, userName, c.Driver)
	}

	// Get initial participants
	c.Participants = c.Room.GetParticipants()
}

// StopCollaboration leaves the collaboration room
func (c *CollaborativeComponent) StopCollaboration() {
	if c.Room != nil {
		c.Room.Leave(c.UserID)
	}
}

// GetDriver returns the component's driver
func (c *CollaborativeComponent) GetDriver() LiveDriver {
	return c.Driver
}

// GetTemplate returns empty template as this is a mixin
func (c *CollaborativeComponent) GetTemplate() string {
	return ""
}

// Start initializes the component
func (c *CollaborativeComponent) Start() {
	// This is a mixin, so Start is typically called by the embedding component
}

// BroadcastAction broadcasts an action to all participants
func (c *CollaborativeComponent) BroadcastAction(action string, data interface{}) {
	if c.Room != nil {
		c.Room.Broadcast(action, c.UserID, data)
	}
}

// UpdateCursorPosition updates cursor position
func (c *CollaborativeComponent) UpdateCursorPosition(x, y float64, elementID string) {
	if c.Room != nil {
		c.Room.UpdateCursor(c.UserID, x, y, elementID)
	}
}

// SyncState synchronizes shared state
func (c *CollaborativeComponent) SyncState(action string, data interface{}) {
	if c.Room != nil {
		c.Room.UpdateSharedState(c.UserID, action, data)
	}
}

// PresenceIndicator component shows online users
type PresenceIndicator struct {
	*ComponentDriver[*PresenceIndicator]
	Participants []Participant
	RoomID       string
}

func (p *PresenceIndicator) Start() {
	p.Participants = make([]Participant, 0)
	p.Commit()
}

func (p *PresenceIndicator) GetTemplate() string {
	return `
	<div class="presence-indicator">
		<style>
			.presence-indicator {
				position: fixed;
				top: 10px;
				right: 10px;
				background: white;
				border-radius: 8px;
				padding: 10px;
				box-shadow: 0 2px 10px rgba(0,0,0,0.1);
				z-index: 1000;
			}
			.participant-list {
				display: flex;
				gap: 10px;
				align-items: center;
			}
			.participant-avatar {
				width: 32px;
				height: 32px;
				border-radius: 50%;
				display: flex;
				align-items: center;
				justify-content: center;
				color: white;
				font-weight: bold;
				font-size: 14px;
				position: relative;
			}
			.participant-avatar.online::after {
				content: '';
				position: absolute;
				bottom: 0;
				right: 0;
				width: 10px;
				height: 10px;
				background: #4CAF50;
				border-radius: 50%;
				border: 2px solid white;
			}
			.participant-count {
				color: #666;
				font-size: 14px;
			}
		</style>
		<div class="participant-list">
			<span class="participant-count">{{len .Participants}} online</span>
			{{range $i, $p := .Participants}}
				{{if lt $i 5}}
				<div class="participant-avatar online" 
				     style="background-color: {{$p.Color}}" 
				     title="{{$p.Name}}">
					{{index $p.Name 0}}
				</div>
				{{end}}
			{{end}}
			{{if gt (len .Participants) 5}}
				<div class="participant-avatar" style="background-color: #999">
					+{{len .Participants}}
				</div>
			{{end}}
		</div>
	</div>
	`
}

func (p *PresenceIndicator) GetDriver() LiveDriver {
	return p.ComponentDriver
}

func (p *PresenceIndicator) UpdatePresence(data interface{}) {
	if participants, ok := data.([]Participant); ok {
		p.Participants = participants
		p.Commit()
	}
}

// Helper function for template
func minus(a, b int) int {
	return a - b
}

// CollaborationConfig provides configuration for collaborative features
type CollaborationConfig struct {
	EnablePresence      bool
	EnableCursors       bool
	EnableHistory       bool
	MaxHistorySize      int
	InactivityTimeout   time.Duration
	BroadcastBufferSize int
}

// DefaultCollaborationConfig returns default configuration
func DefaultCollaborationConfig() *CollaborationConfig {
	return &CollaborationConfig{
		EnablePresence:      true,
		EnableCursors:       true,
		EnableHistory:       true,
		MaxHistorySize:      1000,
		InactivityTimeout:   5 * time.Minute,
		BroadcastBufferSize: 100,
	}
}

// ConflictResolution strategies for concurrent edits
type ConflictResolution int

const (
	LastWriteWins ConflictResolution = iota
	FirstWriteWins
	MergeChanges
	RequireConsensus
)

// OperationalTransform for conflict-free collaborative editing
type OperationalTransform struct {
	ID        string
	Operation string
	Position  int
	Content   string
	Version   int
	UserID    string
}

// ApplyTransform applies an operational transform
func ApplyTransform(current string, transform OperationalTransform) (string, error) {
	switch transform.Operation {
	case "insert":
		if transform.Position > len(current) {
			return current, fmt.Errorf("position out of bounds")
		}
		return current[:transform.Position] + transform.Content + current[transform.Position:], nil

	case "delete":
		if transform.Position > len(current) || transform.Position+len(transform.Content) > len(current) {
			return current, fmt.Errorf("position out of bounds")
		}
		return current[:transform.Position] + current[transform.Position+len(transform.Content):], nil

	case "replace":
		// Implementation for replace operation
		return transform.Content, nil

	default:
		return current, fmt.Errorf("unknown operation: %s", transform.Operation)
	}
}

// CollaborativeTextEditor for real-time text editing
type CollaborativeTextEditor struct {
	Driver                 LiveDriver
	*CollaborativeComponent
	Content    string
	Version    int
	Transforms []OperationalTransform
}

func (e *CollaborativeTextEditor) Start() {
	e.CollaborativeComponent = &CollaborativeComponent{
		Driver: e.Driver,
	}
	// Use a default room ID for now
	roomID := "text_editor_room"
	userID := "user_" + fmt.Sprintf("%d", time.Now().UnixNano())
	e.StartCollaboration(roomID, userID, "User")
	e.Content = ""
	e.Version = 0
	e.Transforms = make([]OperationalTransform, 0)
}

func (e *CollaborativeTextEditor) ApplyEdit(data interface{}) {
	if editData, ok := data.(map[string]interface{}); ok {
		transform := OperationalTransform{
			ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
			Operation: editData["operation"].(string),
			Position:  int(editData["position"].(float64)),
			Content:   editData["content"].(string),
			Version:   e.Version + 1,
			UserID:    e.UserID,
		}

		// Apply transform locally
		newContent, err := ApplyTransform(e.Content, transform)
		if err == nil {
			e.Content = newContent
			e.Version++
			e.Transforms = append(e.Transforms, transform)

			// Broadcast to other participants
			e.BroadcastAction("text_transform", transform)
		}
	}
}

// GetDriver returns the text editor's driver
func (e *CollaborativeTextEditor) GetDriver() LiveDriver {
	return e.Driver
}

// GetTemplate returns the text editor template
func (e *CollaborativeTextEditor) GetTemplate() string {
	return `<div class="collaborative-text-editor">{{.Content}}</div>`
}

// Serialization helpers for persistence
func (r *CollaborationRoom) ToJSON() ([]byte, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	data := map[string]interface{}{
		"id":           r.ID,
		"name":         r.Name,
		"participants": r.GetParticipants(),
		"state":        r.SharedState,
		"history":      r.History,
	}

	return json.Marshal(data)
}

func (r *CollaborationRoom) FromJSON(data []byte) error {
	var roomData map[string]interface{}
	if err := json.Unmarshal(data, &roomData); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.ID = roomData["id"].(string)
	r.Name = roomData["name"].(string)
	r.SharedState = roomData["state"]

	return nil
}
