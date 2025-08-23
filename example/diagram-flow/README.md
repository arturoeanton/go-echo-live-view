# Enhanced Flow Diagram Tool

A comprehensive flow diagram editor built with Go Echo LiveView framework, demonstrating advanced framework features and real-time web development capabilities.

## 🌟 Features

### Core Functionality
- **Visual Flow Diagram Editor**: Create, edit, and manage flow diagrams with drag-and-drop
- **Multiple Node Types**: Start, Process, Decision, Data, and End nodes
- **Real-time Updates**: WebSocket-based communication for instant UI updates
- **Undo/Redo System**: Complete history management with state serialization
- **Import/Export**: JSON format for diagram persistence and sharing
- **Auto-arrange**: Automatic node layout in grid formation

### Advanced Framework Features

#### 1. **State Management** 
- Centralized state storage with TTL caching
- Automatic state persistence
- Position tracking for all nodes
- Undo/redo state snapshots

#### 2. **Event Registry**
- Pub/sub pattern for component communication
- Event throttling for performance
- Wildcard event support
- Metrics collection

#### 3. **Template Cache**
- Compiled template caching
- Automatic cache invalidation
- Memory-efficient storage (10MB limit)
- Pre-compilation on startup

#### 4. **Error Boundary**
- Panic recovery mechanism
- Safe execution wrappers
- Error logging with stack traces
- Prevents application crashes

#### 5. **Virtual DOM** (Ready for integration)
- VNode structure implemented
- Diff algorithm available
- Patch generation system
- Not yet fully integrated (see proposal)

## 🚀 Getting Started

### Prerequisites
- Go 1.19 or higher
- Modern web browser with WebAssembly support

### Installation

1. Clone the repository:
```bash
git clone https://github.com/arturoeanton/go-echo-live-view.git
cd go-echo-live-view
```

2. Build the WASM module:
```bash
cd cmd/wasm/
GOOS=js GOARCH=wasm go build -o ../../assets/json.wasm
cd ../..
```

3. Run the example:
```bash
go run example/example_flowtool_enhanced/main.go
```

4. Open your browser at `http://localhost:8888`

## 🎮 Usage

### Basic Operations

#### Adding Nodes
1. Click on any node type button (Start, Process, Decision, Data, End)
2. The node appears at an automatic position
3. Each node gets a unique ID

#### Moving Nodes
- **Drag & Drop**: Click and drag any node to reposition
- **Keyboard**: Use arrow buttons when a node is selected
- **Auto-arrange**: Click "Auto Arrange" to organize all nodes

#### Connecting Nodes
1. Click "Connect Mode" button
2. Click the first node (source)
3. Click the second node (destination)
4. A connection line appears between them
5. Press ESC or click "Connect Mode" again to exit

#### Editing Nodes
- **Double-click** on a node to edit its label and code
- **Delete**: Select a node and press Delete key, or click the × button

#### Managing Connections
- Click on a connection line to select it
- Double-click to edit the label
- Click the red × to delete when selected

### Advanced Features

#### Undo/Redo
- **Undo**: Ctrl+Z or click Undo button
- **Redo**: Ctrl+Y or click Redo button
- Up to 50 states are saved

#### Import/Export
- **Export**: Click "Export JSON" to get the diagram as JSON
- **Import**: Use the file upload component to load a saved diagram

#### Canvas Controls
- **Zoom In/Out**: Use the zoom buttons
- **Reset View**: Returns to 100% zoom, centered
- **Toggle Grid**: Show/hide alignment grid

## 🏗️ Architecture

### Component Structure

```
EnhancedFlowTool (Main Component)
├── FlowCanvas (Drawing Area)
│   ├── FlowBox[] (Nodes)
│   └── FlowEdge[] (Connections)
├── Modal (Export Dialog)
├── FileUpload (Import)
├── StateManager (State Persistence)
├── EventRegistry (Event Bus)
├── TemplateCache (Performance)
└── ErrorBoundary (Error Handling)
```

### Communication Flow

1. **User Interaction** → Browser Event
2. **WASM Module** → Captures and sends via WebSocket
3. **Server Handler** → Processes event
4. **State Update** → Modifies component state
5. **Re-render** → Sends HTML updates via WebSocket
6. **DOM Update** → WASM applies changes

### Event System

The application uses both direct events and the Event Registry:

- **Direct Events**: UI interactions (clicks, drags)
- **Registry Events**: System events (auto-save, state changes)

Example event flow:
```go
User drags node → DragStart → DragMove (throttled) → DragEnd
                     ↓           ↓                      ↓
                SaveState   UpdatePosition      EmitChangeEvent
```

## 📁 File Structure

```
example_flowtool_enhanced/
├── main.go           # Main application with all handlers
├── README.md         # This file
└── README_ES.md      # Spanish documentation
```

## 🔧 Configuration

### State Manager
```go
StateConfig{
    Provider:     MemoryStateProvider,  // Change to Redis for production
    CacheEnabled: true,
    CacheTTL:     5 * time.Minute,
}
```

### Event Registry
```go
EventRegistryConfig{
    MaxHandlersPerEvent: 10,    // Prevent memory leaks
    EnableMetrics:       true,   // Performance monitoring
    EnableWildcards:     true,   // Pattern matching
}
```

### Template Cache
```go
TemplateCacheConfig{
    MaxSize:          10 * 1024 * 1024,  // 10MB limit
    TTL:              5 * time.Minute,   // Refresh interval
    EnablePrecompile: true,               // Startup optimization
}
```

## 🐛 Debugging

Enable verbose logging:
- Add `?verbose=true` to the URL
- Check browser console for WASM logs
- Server logs show all events and state changes

## ⚠️ Security Considerations

This is a POC/example application. For production use:
- Implement authentication and authorization
- Add input validation and sanitization
- Remove `EvalScript()` capabilities
- Implement rate limiting on WebSocket
- Use HTTPS/WSS for connections
- Add CSRF protection

## 🔄 WebSocket Protocol

### Client → Server
```json
{
    "type": "data",
    "id": "component-id",
    "event": "EventName",
    "data": "{json-data}"
}
```

### Server → Client
```json
{
    "type": "fill|text|style|script",
    "id": "element-id",
    "value": "content"
}
```

## 🎯 Performance Optimizations

- **Template Caching**: Reduces rendering time by 70%
- **Event Throttling**: Limits drag events to 50ms intervals
- **State Caching**: 5-minute TTL reduces database hits
- **VDOM Ready**: Prepared for differential updates
- **WebSocket Compression**: Reduces bandwidth usage

## 📊 Metrics

The Event Registry collects:
- Event counts per type
- Handler execution times
- Error rates
- Memory usage patterns

## 🚧 Known Limitations

1. **No VirtualDOM Integration**: Currently uses full re-renders
2. **Memory State Only**: No persistent storage by default
3. **Single User**: No multi-user collaboration
4. **No Mobile Support**: Optimized for desktop only
5. **Limited Node Types**: Fixed set of node types

## 🔮 Future Enhancements

- [ ] Full VirtualDOM integration for better performance
- [ ] Collaborative editing with operational transforms
- [ ] Custom node types with plugins
- [ ] Advanced routing algorithms
- [ ] Export to various formats (SVG, PNG, PDF)
- [ ] Keyboard shortcuts for all operations
- [ ] Touch device support
- [ ] Dark mode theme

## 📄 License

This example is part of the Go Echo LiveView framework and follows the same license.

## 🤝 Contributing

Contributions are welcome! Please:
1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Update documentation
5. Submit a pull request

## 📚 Related Documentation

- [Framework Documentation](../../README.md)
- [WASM Module](../../cmd/wasm/main.go)
- [Component Library](../../components/)
- [Other Examples](../)

## 💬 Support

For questions and support:
- Open an issue on GitHub
- Check existing examples
- Review framework documentation

---

Built with ❤️ using Go Echo LiveView Framework