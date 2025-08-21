# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go implementation of Phoenix LiveView-style reactive web development using the Echo framework. It enables server-side rendered components that can be dynamically updated via WebSocket connections without full page reloads.

**IMPORTANT**: This is a Proof of Concept (POC) - not production-ready without significant security hardening.

## Key Commands

### Building and Running
- **Build WASM module**: `cd cmd/wasm/ && GOOS=js GOARCH=wasm go build -o ../../assets/json.wasm`
- **Run with auto-reload**: `gomon` (uses gomon.yaml config, install with: `go install github.com/c9s/gomon@latest`)
- **Build and run script**: `./build_and_run.sh` (builds WASM then runs example2)
- **Run specific example**: `go run example/<example_name>/<example_name>.go`

### Code Quality
- **Format code**: `gofmt -w .`
- **Lint code**: `golint ./...`
- **Vet code**: `go vet ./...`

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
- Includes auto-reconnect functionality on disconnect

### Testing Examples

Examples serve as both documentation and integration tests:
- `example1/`: Basic clock display with auto-update
- `example2/`: Text input with real-time updates  
- `example3/`: Basic example with custom layout
- `example_todo/`: Complex todo list with JSON persistence
- `example_style/`: Dynamic CSS styling changes
- `pedidos_board/`: Advanced board interface with drag-and-drop

When developing new features, create examples in the `example/` directory following the established naming pattern.

## Security Considerations

**WARNING**: This is a POC with known security vulnerabilities:
- `EvalScript()` allows arbitrary JavaScript execution
- No input validation on WebSocket messages
- No authentication/authorization system
- Potential XSS vulnerabilities in templates
- Not suitable for production without comprehensive security review

## Development Guidelines

- Always use `gofmt` for consistent formatting
- Run `golint` and `go vet` before committing
- Document public functions with standard Go comments
- Handle errors explicitly - no silent failures
- Use descriptive English names for public APIs
- Follow existing patterns when adding new components


# CRITICAL RULES (MUST)

- ⚙️ **Keep @cmd/wasm/main.go generic**:  
  The file `@cmd/wasm/main.go` is part of the framework and **must not include example-specific logic**.  
  It must always remain generic, reusable, and decoupled from particular use cases.

- 🧩 **This is a framework, not an application**:  
  The project must always be treated as a **framework/library**, never as a finished app, website, or product.  
  Code generation should focus on reusable, generic, and extensible components — not on example-specific or application-specific logic.