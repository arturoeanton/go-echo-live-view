# 🚀 Collaborative Components Examples

This directory contains example applications demonstrating the collaborative features of Go Echo LiveView.

## 📋 Available Examples

### 1. **kanban_simple** - Simple Kanban Board
A basic kanban board with drag-and-drop functionality.

```bash
cd kanban_simple
go run main.go
# Open http://localhost:8080
```

**Features:**
- Drag and drop cards between columns
- Add new cards
- Filter and search
- Real-time updates

### 2. **kanban_app** - Advanced Kanban Application
Full-featured project management application with multiple boards.

```bash
cd kanban_app
go run main.go
# Open http://localhost:8080
```

**Features:**
- Multiple project boards
- User assignments
- Comments and attachments
- Priority levels
- WIP limits

### 3. **collaborative_canvas** - Drawing Canvas
Real-time collaborative drawing application.

```bash
cd collaborative_canvas
go run main.go
# Open http://localhost:8080
```

**Features:**
- Multi-user drawing
- Shape tools
- Color selection
- Export to image

### 4. **collaborative_demo** - Complete Showcase
Demonstrates all collaborative components in one application.

```bash
cd collaborative_demo
go run main.go
# Open http://localhost:8080
```

**Features:**
- Canvas drawing
- Kanban boards
- Text editor
- Chat room
- Presence indicators

## 🛠️ Building WASM Module

All examples require the WASM module to be built:

```bash
# From project root
cd cmd/wasm/
GOOS=js GOARCH=wasm go build -o ../../assets/json.wasm
```

## 🎯 Key Concepts

### Collaboration Room
Each collaborative session creates a "room" where multiple users can interact:

```go
room := liveview.GetCollaborationHub().CreateRoom("room_id", "Room Name")
```

### User Presence
Track active users and their activities:

```go
participant := room.Join(userID, userName, driver)
```

### State Synchronization
Automatically sync state changes across all participants:

```go
room.UpdateSharedState(userID, "action", data)
```

### Real-time Updates
Changes are broadcast to all connected users instantly via WebSockets.

## 📚 Architecture

```
┌─────────────┐     WebSocket      ┌─────────────┐
│   Browser   │◄──────────────────►│   Server    │
│             │                     │             │
│  - Canvas   │                     │ - Room Mgmt │
│  - Kanban   │     Events          │ - State     │
│  - Editor   │◄──────────────────►│ - Broadcast │
└─────────────┘                     └─────────────┘
       ▲                                   ▲
       │                                   │
       └──────── WASM Module ──────────────┘
```

## 🔧 Customization

### Creating Custom Collaborative Components

1. Embed `CollaborativeComponent`:
```go
type MyComponent struct {
    *liveview.ComponentDriver[*MyComponent]
    *liveview.CollaborativeComponent
    // Your fields
}
```

2. Initialize collaboration:
```go
func (c *MyComponent) Start() {
    c.CollaborativeComponent = &liveview.CollaborativeComponent{
        Driver: c.ComponentDriver,
    }
    c.StartCollaboration("room_id", "user_id", "User Name")
}
```

3. Broadcast changes:
```go
c.BroadcastAction("action_name", data)
```

4. Handle remote updates:
```go
func (c *MyComponent) HandleRemoteUpdate(data interface{}) {
    // Process updates from other users
}
```

## 🚦 Running Multiple Examples

You can run multiple examples on different ports:

```bash
# Terminal 1
cd kanban_simple
go run main.go

# Terminal 2  
cd collaborative_canvas
PORT=8081 go run main.go

# Terminal 3
cd collaborative_demo
PORT=8082 go run main.go
```

## 📖 Documentation

For more information about the LiveView framework and collaborative features, see:
- [Main README](../README.md)
- [LiveView Documentation](../liveview/README.md)
- [Components Documentation](../components/README.md)

## 🤝 Contributing

Feel free to create new examples demonstrating different use cases for collaborative components!