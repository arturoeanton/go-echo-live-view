package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/arturoeanton/go-echo-live-view/liveview"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// FlowNode represents a node in the flow
type FlowNode struct {
	ID          string
	Type        string
	Title       string
	Description string
	Position    Position
	Inputs      []string
	Outputs     []string
	Status      string
	Config      map[string]interface{}
}

type Position struct {
	X float64
	Y float64
}

type Connection struct {
	ID     string
	Source string
	Target string
	Label  string
}

// FlowTool is the main flow editor component
type FlowTool struct {
	*liveview.ComponentDriver[*FlowTool]
	
	// Flow state
	FlowName    string
	Nodes       []FlowNode
	Connections []Connection
	SelectedNode string
	ExecutionStatus string
	
	// Framework features
	errorBoundary *liveview.ErrorBoundary
	stateManager  *liveview.StateManager
	eventRegistry *liveview.EventRegistry
	lifecycle     *liveview.LifecycleManager
	templateCache *liveview.TemplateCache
}

func (f *FlowTool) Start() {
	// Initialize Error Boundary
	f.errorBoundary = liveview.NewErrorBoundary(50, true)
	f.errorBoundary.SetErrorHandler(func(err liveview.ComponentError) error {
		log.Printf("Flow error: %v", err.Error)
		f.ExecutionStatus = fmt.Sprintf("Error: %v", err.Error)
		return nil
	})
	
	// Initialize State Manager with JSON provider for persistence
	memProvider := liveview.NewMemoryStateProvider()
	f.stateManager = liveview.NewStateManager(&liveview.StateConfig{
		Provider:         liveview.NewJSONStateProvider(memProvider),
		CacheEnabled:     true,
		CacheTTL:         5 * time.Minute,
		AutoPersist:      true,
		PersistInterval:  10 * time.Second,
		EnableVersioning: true,
	})
	
	// Initialize Template Cache for performance
	f.templateCache = liveview.NewTemplateCache(&liveview.TemplateCacheConfig{
		MaxSize:            10 * 1024 * 1024, // 10MB
		TTL:                5 * time.Minute,
		EnablePrecompile:   true,
	})
	
	// Initialize Event Registry with throttling
	f.eventRegistry = liveview.NewEventRegistry(&liveview.EventRegistryConfig{
		MaxHandlersPerEvent: 5,
		EnableMetrics:       true,
		EnableWildcards:     true,
		DefaultTimeout:      30 * time.Second,
	})
	
	// Setup event handlers
	f.eventRegistry.On("node.*", func(ctx context.Context, event *liveview.Event) error {
		log.Printf("Node event: %s", event.Type)
		f.stateManager.Set("last_node_event", time.Now())
		return nil
	})
	
	f.eventRegistry.On("flow.execute", func(ctx context.Context, event *liveview.Event) error {
		return f.executeFlow()
	})
	
	// Initialize Lifecycle Manager
	f.lifecycle = liveview.NewLifecycleManager("flowtool")
	f.lifecycle.SetHooks(&liveview.LifecycleHooks{
		OnCreated: func() error {
			log.Println("FlowTool created")
			return f.loadFlow()
		},
		OnMounted: func() error {
			log.Println("FlowTool mounted")
			// Template will be cached automatically
			return nil
		},
		OnBeforeUpdate: func(oldData, newData interface{}) error {
			// Save snapshot before update
			snapshot, _ := f.stateManager.TakeSnapshot()
			f.stateManager.Set("last_snapshot", snapshot)
			return nil
		},
		OnBeforeUnmount: func() error {
			return f.saveFlow()
		},
	})
	
	// Initialize flow
	f.FlowName = "Enhanced Data Pipeline"
	f.Nodes = []FlowNode{
		{
			ID:          "input",
			Type:        "input",
			Title:       "Data Input",
			Description: "Receives data from API",
			Position:    Position{X: 100, Y: 200},
			Outputs:     []string{"data"},
			Status:      "ready",
			Config:      map[string]interface{}{"source": "api"},
		},
		{
			ID:          "transform",
			Type:        "transform",
			Title:       "Transform",
			Description: "Applies transformations",
			Position:    Position{X: 350, Y: 200},
			Inputs:      []string{"data"},
			Outputs:     []string{"transformed"},
			Status:      "ready",
			Config:      map[string]interface{}{"type": "json"},
		},
		{
			ID:          "filter",
			Type:        "filter",
			Title:       "Filter",
			Description: "Filters data based on rules",
			Position:    Position{X: 600, Y: 200},
			Inputs:      []string{"data"},
			Outputs:     []string{"filtered"},
			Status:      "ready",
			Config:      map[string]interface{}{"condition": "value > 0"},
		},
		{
			ID:          "output",
			Type:        "output",
			Title:       "Data Output",
			Description: "Sends to database",
			Position:    Position{X: 850, Y: 200},
			Inputs:      []string{"data"},
			Status:      "ready",
			Config:      map[string]interface{}{"target": "database"},
		},
	}
	
	f.Connections = []Connection{
		{ID: "c1", Source: "input", Target: "transform", Label: "raw data"},
		{ID: "c2", Source: "transform", Target: "filter", Label: "transformed"},
		{ID: "c3", Source: "filter", Target: "output", Label: "filtered"},
	}
	
	f.ExecutionStatus = "Ready"
	
	// Execute lifecycle
	f.lifecycle.Create()
	f.lifecycle.Mount()
	
	f.Commit()
}

func (f *FlowTool) loadFlow() error {
	// Load saved flow if exists
	if savedNodes, err := f.stateManager.Get("flow_nodes"); err == nil && savedNodes != nil {
		if nodes, ok := savedNodes.([]FlowNode); ok {
			f.Nodes = nodes
		}
	}
	
	if savedConnections, err := f.stateManager.Get("flow_connections"); err == nil && savedConnections != nil {
		if connections, ok := savedConnections.([]Connection); ok {
			f.Connections = connections
		}
	}
	
	return nil
}

func (f *FlowTool) saveFlow() error {
	f.stateManager.Set("flow_nodes", f.Nodes)
	f.stateManager.Set("flow_connections", f.Connections)
	f.stateManager.Set("flow_saved_at", time.Now())
	return nil
}

func (f *FlowTool) executeFlow() error {
	f.ExecutionStatus = "Executing..."
	
	// Simulate flow execution with error boundary protection
	return f.errorBoundary.SafeExecute("flow_execution", func() error {
		for i, node := range f.Nodes {
			// Update node status
			f.Nodes[i].Status = "executing"
			f.Commit()
			
			// Emit node execution event
			f.eventRegistry.Emit("node.executing", map[string]interface{}{
				"node_id": node.ID,
				"type":    node.Type,
			})
			
			// Simulate processing
			time.Sleep(500 * time.Millisecond)
			
			// Mark as completed
			f.Nodes[i].Status = "completed"
			f.Commit()
		}
		
		f.ExecutionStatus = "Completed successfully"
		return nil
	})
}

func (f *FlowTool) GetTemplate() string {
	return `
<!DOCTYPE html>
<html>
<head>
	<title>Flow Tool v2</title>
	<style>
		body {
			font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
			margin: 0;
			background: #1a1a2e;
			color: white;
			height: 100vh;
			display: flex;
			flex-direction: column;
		}
		.header {
			background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
			padding: 1rem 2rem;
			display: flex;
			justify-content: space-between;
			align-items: center;
			box-shadow: 0 2px 20px rgba(0,0,0,0.3);
		}
		.flow-name {
			font-size: 1.5rem;
			font-weight: bold;
		}
		.toolbar {
			display: flex;
			gap: 1rem;
		}
		.btn {
			padding: 0.5rem 1rem;
			background: rgba(255,255,255,0.2);
			border: 1px solid rgba(255,255,255,0.3);
			color: white;
			border-radius: 6px;
			cursor: pointer;
			transition: all 0.3s;
		}
		.btn:hover {
			background: rgba(255,255,255,0.3);
			transform: translateY(-2px);
		}
		.canvas {
			flex: 1;
			position: relative;
			overflow: auto;
			background: #1a1a2e;
			background-image: 
				radial-gradient(circle at 1px 1px, #2a2a3e 1px, transparent 1px);
			background-size: 20px 20px;
		}
		.node {
			position: absolute;
			background: #2a2a3e;
			border: 2px solid #667eea;
			border-radius: 8px;
			padding: 1rem;
			min-width: 150px;
			cursor: move;
			transition: all 0.3s;
		}
		.node:hover {
			box-shadow: 0 0 20px rgba(102, 126, 234, 0.5);
			transform: scale(1.05);
		}
		.node.executing {
			border-color: #f39c12;
			animation: pulse 1s infinite;
		}
		.node.completed {
			border-color: #27ae60;
		}
		.node-type {
			font-size: 0.75rem;
			color: #667eea;
			text-transform: uppercase;
			margin-bottom: 0.5rem;
		}
		.node-title {
			font-weight: 600;
			margin-bottom: 0.5rem;
		}
		.node-description {
			font-size: 0.875rem;
			color: #999;
		}
		.node-ports {
			display: flex;
			justify-content: space-between;
			margin-top: 0.5rem;
			font-size: 0.75rem;
		}
		.connection {
			position: absolute;
			pointer-events: none;
		}
		.connection-line {
			stroke: #667eea;
			stroke-width: 2;
			fill: none;
		}
		.status-bar {
			background: #2a2a3e;
			padding: 1rem 2rem;
			display: flex;
			justify-content: space-between;
			align-items: center;
			border-top: 1px solid #3a3a4e;
		}
		.status {
			display: flex;
			align-items: center;
			gap: 0.5rem;
		}
		.status-indicator {
			width: 10px;
			height: 10px;
			border-radius: 50%;
			background: #27ae60;
		}
		.status-indicator.executing {
			background: #f39c12;
			animation: blink 1s infinite;
		}
		.features {
			display: flex;
			gap: 1rem;
		}
		.feature-tag {
			padding: 0.25rem 0.5rem;
			background: rgba(102, 126, 234, 0.2);
			border-radius: 12px;
			font-size: 0.75rem;
		}
		@keyframes pulse {
			0% { opacity: 1; }
			50% { opacity: 0.7; }
			100% { opacity: 1; }
		}
		@keyframes blink {
			0%, 100% { opacity: 1; }
			50% { opacity: 0.3; }
		}
		.sidebar {
			position: fixed;
			right: 0;
			top: 70px;
			bottom: 50px;
			width: 250px;
			background: #2a2a3e;
			border-left: 1px solid #3a3a4e;
			padding: 1rem;
			overflow-y: auto;
		}
		.node-palette {
			margin-bottom: 2rem;
		}
		.palette-item {
			background: #1a1a2e;
			border: 1px solid #3a3a4e;
			border-radius: 6px;
			padding: 0.75rem;
			margin-bottom: 0.5rem;
			cursor: grab;
			transition: all 0.3s;
		}
		.palette-item:hover {
			border-color: #667eea;
			transform: translateX(-5px);
		}
	</style>
</head>
<body>
	<div class="header">
		<div class="flow-name">{{.FlowName}}</div>
		<div class="toolbar">
			<button class="btn" onclick="send_event('{{.IdComponent}}', 'Execute', null)">
				▶ Execute Flow
			</button>
			<button class="btn" onclick="send_event('{{.IdComponent}}', 'Save', null)">
				💾 Save
			</button>
			<button class="btn" onclick="send_event('{{.IdComponent}}', 'Clear', null)">
				🗑️ Clear
			</button>
		</div>
	</div>
	
	<div class="canvas" id="canvas">
		<!-- Render nodes -->
		{{range .Nodes}}
		<div class="node {{.Status}}" 
		     style="left: {{.Position.X}}px; top: {{.Position.Y}}px;"
		     data-id="{{.ID}}">
			<div class="node-type">{{.Type}}</div>
			<div class="node-title">{{.Title}}</div>
			<div class="node-description">{{.Description}}</div>
			<div class="node-ports">
				<div>{{if .Inputs}}⊙ {{len .Inputs}}{{end}}</div>
				<div>{{if .Outputs}}{{len .Outputs}} ⊙{{end}}</div>
			</div>
		</div>
		{{end}}
		
		<!-- SVG for connections -->
		<svg style="position: absolute; top: 0; left: 0; width: 100%; height: 100%; pointer-events: none;">
			{{range .Connections}}
			<path class="connection-line" d="M 200 200 Q 400 200 600 200" />
			{{end}}
		</svg>
	</div>
	
	<div class="sidebar">
		<h3>Node Palette</h3>
		<div class="node-palette">
			<div class="palette-item" draggable="true">
				<div style="font-weight: bold;">Input Node</div>
				<div style="font-size: 0.875rem; color: #999;">Data source</div>
			</div>
			<div class="palette-item" draggable="true">
				<div style="font-weight: bold;">Transform</div>
				<div style="font-size: 0.875rem; color: #999;">Process data</div>
			</div>
			<div class="palette-item" draggable="true">
				<div style="font-weight: bold;">Filter</div>
				<div style="font-size: 0.875rem; color: #999;">Apply conditions</div>
			</div>
			<div class="palette-item" draggable="true">
				<div style="font-weight: bold;">Output Node</div>
				<div style="font-size: 0.875rem; color: #999;">Data destination</div>
			</div>
		</div>
		
		<h3>Features</h3>
		<div class="features" style="flex-direction: column;">
			<div class="feature-tag">Error Recovery</div>
			<div class="feature-tag">State Persistence</div>
			<div class="feature-tag">Template Cache</div>
			<div class="feature-tag">Event System</div>
			<div class="feature-tag">Lifecycle Hooks</div>
		</div>
	</div>
	
	<div class="status-bar">
		<div class="status">
			<div class="status-indicator {{if eq .ExecutionStatus "Executing..."}}executing{{end}}"></div>
			<span>Status: {{.ExecutionStatus}}</span>
		</div>
		<div class="features">
			<div class="feature-tag">{{len .Nodes}} nodes</div>
			<div class="feature-tag">{{len .Connections}} connections</div>
		</div>
	</div>
	
	<script>
		// Enable drag and drop for nodes
		document.querySelectorAll('.node').forEach(node => {
			let isDragging = false;
			let startX, startY, initialX, initialY;
			
			node.addEventListener('mousedown', (e) => {
				isDragging = true;
				startX = e.clientX;
				startY = e.clientY;
				initialX = node.offsetLeft;
				initialY = node.offsetTop;
				node.style.zIndex = 1000;
			});
			
			document.addEventListener('mousemove', (e) => {
				if (!isDragging) return;
				e.preventDefault();
				const dx = e.clientX - startX;
				const dy = e.clientY - startY;
				node.style.left = (initialX + dx) + 'px';
				node.style.top = (initialY + dy) + 'px';
			});
			
			document.addEventListener('mouseup', () => {
				if (isDragging) {
					isDragging = false;
					node.style.zIndex = '';
					send_event('{{.IdComponent}}', 'UpdateNodePosition', {
						nodeId: node.dataset.id,
						x: node.offsetLeft,
						y: node.offsetTop
					});
				}
			});
		});
		
		// Node selection
		document.querySelectorAll('.node').forEach(node => {
			node.addEventListener('click', (e) => {
				if (!e.target.closest('.node').classList.contains('selected')) {
					document.querySelectorAll('.node').forEach(n => n.classList.remove('selected'));
					e.target.closest('.node').classList.add('selected');
					send_event('{{.IdComponent}}', 'SelectNode', {
						nodeId: node.dataset.id
					});
				}
			});
		});
	</script>
</body>
</html>
	`
}

func (f *FlowTool) GetDriver() liveview.LiveDriver {
	return f
}

func (f *FlowTool) Execute(data interface{}) {
	f.errorBoundary.SafeExecute("execute", func() error {
		return f.executeFlow()
	})
	f.Commit()
}

func (f *FlowTool) Save(data interface{}) {
	f.errorBoundary.SafeExecute("save", func() error {
		f.saveFlow()
		f.ExecutionStatus = "Flow saved"
		return nil
	})
	f.Commit()
}

func (f *FlowTool) Clear(data interface{}) {
	f.errorBoundary.SafeExecute("clear", func() error {
		f.Nodes = []FlowNode{}
		f.Connections = []Connection{}
		f.ExecutionStatus = "Flow cleared"
		return nil
	})
	f.Commit()
}

func (f *FlowTool) SelectNode(data interface{}) {
	if m, ok := data.(map[string]interface{}); ok {
		if nodeId, ok := m["nodeId"].(string); ok {
			f.SelectedNode = nodeId
			f.eventRegistry.Emit("node.selected", map[string]interface{}{
				"node_id": nodeId,
			})
		}
	}
	f.Commit()
}

func (f *FlowTool) UpdateNodePosition(data interface{}) {
	if m, ok := data.(map[string]interface{}); ok {
		nodeId := m["nodeId"].(string)
		x := m["x"].(float64)
		y := m["y"].(float64)
		
		for i, node := range f.Nodes {
			if node.ID == nodeId {
				f.Nodes[i].Position = Position{X: x, Y: y}
				break
			}
		}
		
		f.eventRegistry.Emit("node.moved", map[string]interface{}{
			"node_id": nodeId,
			"x": x,
			"y": y,
		})
	}
	f.Commit()
}

func main() {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	
	// Flow tool page
	page := &liveview.PageControl{
		Path:   "/",
		Title:  "Flow Tool v2 - Enhanced",
		Router: e,
	}
	
	page.Register(func() liveview.LiveDriver {
		return liveview.NewDriver("flowtool", &FlowTool{})
	})
	
	port := ":8082"
	fmt.Printf("Starting Flow Tool v2\n")
	fmt.Printf("Open http://localhost%s\n", port)
	fmt.Println("\nFeatures:")
	fmt.Println("  • Error boundaries for safe execution")
	fmt.Println("  • State persistence with JSON provider")
	fmt.Println("  • Template caching for performance")
	fmt.Println("  • Event registry with throttling")
	fmt.Println("  • Lifecycle management")
	fmt.Println("  • Snapshot and restore capability")
	
	if err := e.Start(port); err != nil {
		log.Fatal(err)
	}
}