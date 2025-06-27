# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go implementation of Phoenix LiveView-style reactive web development using the Echo framework. It enables server-side rendered components that can be dynamically updated via WebSocket connections without full page reloads.

## Key Commands

### Building and Running
- **Build WASM module**: `cd cmd/wasm/ && GOOS=js GOARCH=wasm go build -o ../../assets/json.wasm`
- **Run with auto-reload**: `gomon` (uses gomon.yaml config)
- **Build and run script**: `./build_and_run.sh` (builds WASM then runs example2)
- **Run specific example**: `go run example/example1/example1.go`

### Development
- **Run any example**: `go run example/<example_name>/<example_name>.go`
- **Auto-reload during development**: Use `gomon` which watches .go and .html files

## Architecture Overview

### Core Components

1. **liveview.ComponentDriver[T]**: Generic driver that wraps components and handles WebSocket communication
   - Manages component lifecycle, event handling, and DOM updates
   - Uses channels for bidirectional communication with browser
   - Supports mounting child components

2. **liveview.PageControl**: Main page controller that sets up Echo routes
   - Registers both HTTP route (serves initial HTML) and WebSocket route
   - Handles WebSocket upgrade and message routing

3. **Component Interface**: All interactive components must implement:
   ```go
   type Component interface {
       GetTemplate() string  // HTML template
       Start()              // Initialization logic
       GetDriver() LiveDriver
   }
   ```

### Communication Flow

1. **Initial Load**: Browser requests page, receives HTML with embedded WASM loader
2. **WebSocket Connection**: WASM connects to `/ws_goliveview` endpoint  
3. **Bidirectional Updates**: 
   - Browser events sent as JSON: `{"type": "data", "id": "component_id", "event": "Click", "data": {...}}`
   - Server updates sent as JSON: `{"type": "fill", "id": "element_id", "value": "new_html"}`

### Key Patterns

- **Component Registration**: Use `liveview.NewDriver(id, component)` to create component drivers
- **Event Handling**: Components can define event handlers via `Events` map or method names
- **DOM Updates**: Use driver methods like `FillValue()`, `SetHTML()`, `SetText()`, `SetStyle()`
- **Component Mounting**: Use `Mount()` to embed child components, referenced via `{{mount "component_id"}}`

### Directory Structure

- `liveview/`: Core framework code (drivers, page control, templates)
- `components/`: Reusable UI components (Button, Input, Clock)  
- `example/`: Complete working examples showing different features
- `assets/`: Static files including compiled WASM module
- `cmd/wasm/`: WASM build target for browser-side JSON handling

### WASM Integration

The framework uses WebAssembly for browser-side JSON processing:
- Built from `cmd/wasm/main.go` into `assets/json.wasm`
- Loaded automatically in the base HTML template
- Handles WebSocket message parsing and DOM manipulation

### Testing Examples

Examples serve as both documentation and integration tests:
- `example1/`: Basic counter with button clicks
- `example2/`: Text input with real-time updates  
- `example_todo/`: Complex todo list with JSON persistence
- `example_style/`: Dynamic CSS styling changes
- `pedidos_board/`: Advanced board interface

When developing new features, create examples in the `example/` directory following the established naming pattern.