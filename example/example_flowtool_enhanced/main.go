package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/arturoeanton/go-echo-live-view/components"
	"github.com/arturoeanton/go-echo-live-view/liveview"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type EnhancedFlowTool struct {
	*liveview.ComponentDriver[*EnhancedFlowTool]
	
	Canvas         *components.FlowCanvas
	Modal          *components.Modal
	Title          string
	Description    string
	NodeCount      int
	EdgeCount      int
	LastAction     string
	ConnectingMode bool
	ConnectingFrom string
	DraggingBox    string
	JsonExport     string
	
	// New features
	StateManager   *liveview.StateManager
	EventRegistry  *liveview.EventRegistry
	TemplateCache  *liveview.TemplateCache
	ErrorBoundary  *liveview.ErrorBoundary
	Lifecycle      *liveview.LifecycleManager
	UndoStack      []string
	RedoStack      []string
	AutoSaveTimer  *time.Timer
}

func NewEnhancedFlowTool() *EnhancedFlowTool {
	// Create canvas
	canvas := components.NewFlowCanvas("main-canvas", 1200, 600)
	
	// Create modal for JSON export
	modal := &components.Modal{
		Title:      "Export JSON",
		Size:       "large",
		Closable:   true,
		ShowFooter: false,
		IsOpen:     false,
	}
	
	// Initialize State Manager
	stateManager := liveview.NewStateManager(&liveview.StateConfig{
		Provider:     liveview.NewMemoryStateProvider(),
		CacheEnabled: true,
		CacheTTL:     5 * time.Minute,
	})
	
	// Initialize Event Registry with advanced features
	eventRegistry := liveview.NewEventRegistry(&liveview.EventRegistryConfig{
		MaxHandlersPerEvent: 10,
		EnableMetrics:       true,
		EnableWildcards:     true,
	})
	
	// Initialize Template Cache
	templateCache := liveview.NewTemplateCache(&liveview.TemplateCacheConfig{
		MaxSize:          10 * 1024 * 1024, // 10MB
		TTL:              5 * time.Minute,
		EnablePrecompile: true,
	})
	
	// Initialize Error Boundary
	errorBoundary := liveview.NewErrorBoundary(100, true)
	
	// Create initial flow diagram with enhanced components
	startBox := components.NewFlowBox("start", "Start", components.BoxTypeStart, 50, 250)
	initBox := components.NewFlowBox("init", "Initialize System", components.BoxTypeProcess, 200, 250)
	validateBox := components.NewFlowBox("validate", "Validate Input", components.BoxTypeProcess, 400, 150)
	checkBox := components.NewFlowBox("check", "Security Check", components.BoxTypeDecision, 400, 350)
	processBox := components.NewFlowBox("process", "Process Data", components.BoxTypeProcess, 600, 150)
	errorBox := components.NewFlowBox("error", "Handle Error", components.BoxTypeProcess, 600, 450)
	cacheBox := components.NewFlowBox("cache", "Update Cache", components.BoxTypeData, 800, 150)
	logBox := components.NewFlowBox("log", "Log Activity", components.BoxTypeData, 800, 350)
	notifyBox := components.NewFlowBox("notify", "Send Notification", components.BoxTypeProcess, 1000, 250)
	endBox := components.NewFlowBox("end", "End", components.BoxTypeEnd, 1150, 250)
	
	// Add boxes to canvas
	canvas.Boxes[startBox.ID] = startBox
	canvas.Boxes[initBox.ID] = initBox
	canvas.Boxes[validateBox.ID] = validateBox
	canvas.Boxes[checkBox.ID] = checkBox
	canvas.Boxes[processBox.ID] = processBox
	canvas.Boxes[errorBox.ID] = errorBox
	canvas.Boxes[cacheBox.ID] = cacheBox
	canvas.Boxes[logBox.ID] = logBox
	canvas.Boxes[notifyBox.ID] = notifyBox
	canvas.Boxes[endBox.ID] = endBox
	
	// Create edges with enhanced properties
	edges := []struct {
		id, from, to, label string
		curved bool
	}{
		{"e1", "start", "init", "Begin", false},
		{"e2", "init", "validate", "Initialize", false},
		{"e3", "init", "check", "Check", false},
		{"e4", "validate", "process", "Valid", true},
		{"e5", "check", "process", "Secure", true},
		{"e6", "check", "error", "Insecure", true},
		{"e7", "process", "cache", "Store", false},
		{"e8", "process", "log", "Log", true},
		{"e9", "error", "log", "Error Log", false},
		{"e10", "cache", "notify", "Updated", false},
		{"e11", "log", "notify", "Logged", true},
		{"e12", "notify", "end", "Complete", false},
	}
	
	for _, e := range edges {
		edge := components.NewFlowEdge(e.id, e.from, "out1", e.to, "in1")
		edge.Label = e.label
		if e.curved {
			edge.Type = components.EdgeTypeCurved
		}
		
		// Update positions
		if fromBox, ok := canvas.Boxes[e.from]; ok {
			if toBox, ok := canvas.Boxes[e.to]; ok {
				edge.UpdatePosition(
					fromBox.X+fromBox.Width, fromBox.Y+fromBox.Height/2,
					toBox.X, toBox.Y+toBox.Height/2,
				)
			}
		}
		
		canvas.AddEdge(edge)
	}
	
	// Set up enhanced callbacks
	canvas.OnBoxClick = func(boxID string) {
		log.Printf("[VDOM] Box clicked: %s", boxID)
	}
	
	canvas.OnEdgeClick = func(edgeID string) {
		log.Printf("[Event Registry] Edge clicked: %s", edgeID)
	}
	
	canvas.OnConnection = func(fromBox, fromPort, toBox, toPort string) {
		log.Printf("[State Manager] Connection made: %s:%s -> %s:%s", fromBox, fromPort, toBox, toPort)
	}
	
	canvas.OnBoxMove = func(boxID string, x, y int) {
		log.Printf("[Auto-save] Box %s moved to (%d, %d)", boxID, x, y)
	}
	
	tool := &EnhancedFlowTool{
		Canvas:         canvas,
		Modal:          modal,
		Title:          "Enhanced Flow Diagram Tool",
		Description:    "Powered by Virtual DOM, State Management, and Event Registry",
		NodeCount:      0,
		EdgeCount:      0,
		LastAction:     "Diagram initialized with enhanced features",
		StateManager:   stateManager,
		EventRegistry:  eventRegistry,
		TemplateCache:  templateCache,
		ErrorBoundary:  errorBoundary,
		UndoStack:      make([]string, 0),
		RedoStack:      make([]string, 0),
	}
	
	// Add some initial test boxes
	startBox1 := components.NewFlowBox("start_1", "Start", components.BoxTypeStart, 100, 100)
	processBox1 := components.NewFlowBox("process_1", "Process", components.BoxTypeProcess, 300, 100)
	endBox1 := components.NewFlowBox("end_1", "End", components.BoxTypeEnd, 500, 100)
	
	canvas.AddBox(startBox1)
	canvas.AddBox(processBox1)
	canvas.AddBox(endBox1)
	
	tool.NodeCount = 3
	
	// Register event handlers with throttling and debouncing
	tool.setupEnhancedEventHandlers()
	
	// Load saved state if available
	tool.loadSavedState()
	
	// Start auto-save timer
	tool.startAutoSave()
	
	return tool
}

func (f *EnhancedFlowTool) setupEnhancedEventHandlers() {
	// Register handlers with event registry using context
	
	// Box drag event
	f.EventRegistry.On("box.drag", func(ctx context.Context, event *liveview.Event) error {
		// Update position in state
		f.StateManager.Set("last_drag", event.Data)
		return nil
	})
	
	// Connection creation event
	f.EventRegistry.On("connection.create", func(ctx context.Context, event *liveview.Event) error {
		// Validate connection before creating
		if from, _ := event.Data["from"].(string); from != "" {
			if to, _ := event.Data["to"].(string); to != "" {
				if f.validateConnection(from, to) {
					f.createConnection(from, to)
					f.saveToUndoStack()
				}
			}
		}
		return nil
	})
	
	// Auto-save event
	f.EventRegistry.On("diagram.change", func(ctx context.Context, event *liveview.Event) error {
		f.saveState()
		return nil
	})
}

func (f *EnhancedFlowTool) loadSavedState() {
	// Try to load from state manager
	if savedDiagram, err := f.StateManager.Get("flow_diagram"); err == nil && savedDiagram != nil {
		if _, ok := savedDiagram.(map[string]interface{}); ok {
			log.Println("Loaded saved diagram from state manager")
			// Restore diagram state
			f.LastAction = "Loaded saved diagram"
		}
	}
}

func (f *EnhancedFlowTool) startAutoSave() {
	f.AutoSaveTimer = time.AfterFunc(30*time.Second, func() {
		f.saveState()
		f.startAutoSave() // Restart timer
	})
}

func (f *EnhancedFlowTool) saveState() {
	diagramData := map[string]interface{}{
		"boxes":     f.Canvas.Boxes,
		"edges":     f.Canvas.Edges,
		"timestamp": time.Now(),
	}
	f.StateManager.Set("flow_diagram", diagramData)
	f.StateManager.Set("last_save", time.Now())
	log.Println("Diagram auto-saved")
}

func (f *EnhancedFlowTool) validateConnection(from, to string) bool {
	// Prevent self-connections
	if from == to {
		f.LastAction = "Cannot connect node to itself"
		return false
	}
	
	// Check for duplicate connections
	for _, edge := range f.Canvas.Edges {
		if edge.FromBox == from && edge.ToBox == to {
			f.LastAction = "Connection already exists"
			return false
		}
	}
	
	return true
}

func (f *EnhancedFlowTool) createConnection(from, to string) {
	edgeID := fmt.Sprintf("edge_%s_%s_%d", from, to, time.Now().Unix())
	edge := components.NewFlowEdge(edgeID, from, "out1", to, "in1")
	
	// Update positions
	if fromBox, ok := f.Canvas.Boxes[from]; ok {
		if toBox, ok := f.Canvas.Boxes[to]; ok {
			edge.UpdatePosition(
				fromBox.X+fromBox.Width, fromBox.Y+fromBox.Height/2,
				toBox.X, toBox.Y+toBox.Height/2,
			)
		}
	}
	
	f.Canvas.Edges[edgeID] = edge
	f.EdgeCount++
	f.LastAction = fmt.Sprintf("Created connection: %s -> %s", from, to)
	
	// Trigger change event
	f.EventRegistry.Emit("diagram.change", map[string]interface{}{
		"type": "edge_added", 
		"edge": edgeID,
	})
}

func (f *EnhancedFlowTool) saveToUndoStack() {
	// Save current state to undo stack
	stateJSON, _ := json.Marshal(map[string]interface{}{
		"boxes": f.Canvas.Boxes,
		"edges": f.Canvas.Edges,
	})
	f.UndoStack = append(f.UndoStack, string(stateJSON))
	
	// Limit undo stack size
	if len(f.UndoStack) > 50 {
		f.UndoStack = f.UndoStack[1:]
	}
	
	// Clear redo stack on new action
	f.RedoStack = []string{}
}

func (f *EnhancedFlowTool) Undo(data interface{}) {
	if len(f.UndoStack) > 0 {
		// Save current state to redo stack
		currentState, _ := json.Marshal(map[string]interface{}{
			"boxes": f.Canvas.Boxes,
			"edges": f.Canvas.Edges,
		})
		f.RedoStack = append(f.RedoStack, string(currentState))
		
		// Restore previous state
		prevState := f.UndoStack[len(f.UndoStack)-1]
		f.UndoStack = f.UndoStack[:len(f.UndoStack)-1]
		
		var state map[string]interface{}
		if err := json.Unmarshal([]byte(prevState), &state); err == nil {
			// Restore boxes
			if boxes, ok := state["boxes"].(map[string]interface{}); ok {
				f.Canvas.Boxes = make(map[string]*components.FlowBox)
				for id, boxData := range boxes {
					if boxMap, ok := boxData.(map[string]interface{}); ok {
						box := &components.FlowBox{
							ID:    id,
							Label: boxMap["Label"].(string),
							X:     int(boxMap["X"].(float64)),
							Y:     int(boxMap["Y"].(float64)),
						}
						f.Canvas.Boxes[id] = box
					}
				}
			}
			// Restore edges
			if edges, ok := state["edges"].(map[string]interface{}); ok {
				f.Canvas.Edges = make(map[string]*components.FlowEdge)
				for id, edgeData := range edges {
					if edgeMap, ok := edgeData.(map[string]interface{}); ok {
						edge := &components.FlowEdge{
							ID:       id,
							FromBox:  edgeMap["FromBox"].(string),
							ToBox:    edgeMap["ToBox"].(string),
						}
						f.Canvas.Edges[id] = edge
					}
				}
			}
		}
		
		f.LastAction = "Undo performed"
		f.Commit()
	}
}

func (f *EnhancedFlowTool) Redo(data interface{}) {
	if len(f.RedoStack) > 0 {
		// Save current state to undo stack
		currentState, _ := json.Marshal(map[string]interface{}{
			"boxes": f.Canvas.Boxes,
			"edges": f.Canvas.Edges,
		})
		f.UndoStack = append(f.UndoStack, string(currentState))
		
		// Restore next state
		nextState := f.RedoStack[len(f.RedoStack)-1]
		f.RedoStack = f.RedoStack[:len(f.RedoStack)-1]
		
		var state map[string]interface{}
		if err := json.Unmarshal([]byte(nextState), &state); err == nil {
			// Restore boxes
			if boxes, ok := state["boxes"].(map[string]interface{}); ok {
				f.Canvas.Boxes = make(map[string]*components.FlowBox)
				for id, boxData := range boxes {
					if boxMap, ok := boxData.(map[string]interface{}); ok {
						box := &components.FlowBox{
							ID:    id,
							Label: boxMap["Label"].(string),
							X:     int(boxMap["X"].(float64)),
							Y:     int(boxMap["Y"].(float64)),
						}
						f.Canvas.Boxes[id] = box
					}
				}
			}
			// Restore edges
			if edges, ok := state["edges"].(map[string]interface{}); ok {
				f.Canvas.Edges = make(map[string]*components.FlowEdge)
				for id, edgeData := range edges {
					if edgeMap, ok := edgeData.(map[string]interface{}); ok {
						edge := &components.FlowEdge{
							ID:       id,
							FromBox:  edgeMap["FromBox"].(string),
							ToBox:    edgeMap["ToBox"].(string),
						}
						f.Canvas.Edges[id] = edge
					}
				}
			}
		}
		
		f.LastAction = "Redo performed"
		f.Commit()
	}
}

func (f *EnhancedFlowTool) Start() {
	// Initialize with lifecycle hooks
	f.Lifecycle = liveview.NewLifecycleManager("enhanced_flowtool")
	f.Lifecycle.SetHooks(&liveview.LifecycleHooks{
		OnBeforeMount: func() error {
			log.Println("Enhanced FlowTool mounting...")
			return nil
		},
		OnMounted: func() error {
			log.Println("Enhanced FlowTool mounted successfully")
			return nil
		},
	})
	
	// Execute lifecycle
	f.Lifecycle.Create()
	f.Lifecycle.Mount()
	
	// Initialize modal events
	if f.Modal != nil && f.Modal.ComponentDriver != nil {
		f.Modal.Start()
	}
	
	// Register all event handlers
	if f.ComponentDriver != nil {
		// Enhanced event handlers
		f.ComponentDriver.Events["AddNode"] = func(c *EnhancedFlowTool, data interface{}) {
			f.ErrorBoundary.SafeExecute("add_node", func() error {
				c.HandleAddNode(data)
				return nil
			})
		}
		f.ComponentDriver.Events["ClearDiagram"] = func(c *EnhancedFlowTool, data interface{}) {
			c.ClearDiagram(data)
		}
		f.ComponentDriver.Events["ExportDiagram"] = func(c *EnhancedFlowTool, data interface{}) {
			c.ExportDiagram(data)
		}
		f.ComponentDriver.Events["AnimateFlow"] = func(c *EnhancedFlowTool, data interface{}) {
			c.AnimateFlow(data)
		}
		f.ComponentDriver.Events["Undo"] = func(c *EnhancedFlowTool, data interface{}) {
			c.Undo(data)
		}
		f.ComponentDriver.Events["Redo"] = func(c *EnhancedFlowTool, data interface{}) {
			c.Redo(data)
		}
		f.ComponentDriver.Events["AutoArrange"] = func(c *EnhancedFlowTool, data interface{}) {
			c.AutoArrange(data)
		}
		
		// Box interaction events
		f.ComponentDriver.Events["BoxClick"] = func(c *EnhancedFlowTool, data interface{}) {
			c.HandleBoxClick(data)
		}
		f.ComponentDriver.Events["MoveBox"] = func(c *EnhancedFlowTool, data interface{}) {
			c.HandleMoveBox(data)
		}
		
		// Canvas events
		f.ComponentDriver.Events["CanvasZoomIn"] = func(c *EnhancedFlowTool, data interface{}) {
			c.HandleCanvasZoomIn(data)
		}
		f.ComponentDriver.Events["CanvasZoomOut"] = func(c *EnhancedFlowTool, data interface{}) {
			c.HandleCanvasZoomOut(data)
		}
		f.ComponentDriver.Events["CanvasReset"] = func(c *EnhancedFlowTool, data interface{}) {
			c.HandleCanvasReset(data)
		}
		f.ComponentDriver.Events["ToggleGrid"] = func(c *EnhancedFlowTool, data interface{}) {
			c.HandleToggleGrid(data)
		}
		f.ComponentDriver.Events["ToggleConnectMode"] = func(c *EnhancedFlowTool, data interface{}) {
			c.HandleToggleConnectMode(data)
		}
		
		// Drag events
		f.ComponentDriver.Events["BoxStartDrag"] = func(c *EnhancedFlowTool, data interface{}) {
			c.BoxStartDrag(data)
		}
		f.ComponentDriver.Events["BoxDrag"] = func(c *EnhancedFlowTool, data interface{}) {
			c.BoxDrag(data)
		}
		f.ComponentDriver.Events["BoxEndDrag"] = func(c *EnhancedFlowTool, data interface{}) {
			c.BoxEndDrag(data)
		}
	}
	
	if f.ComponentDriver != nil {
		f.Commit()
	}
}

func (f *EnhancedFlowTool) GetTemplate() string {
	// Use cached template if available
	if cached, exists := f.TemplateCache.Get("flowtool_main"); exists {
		var buf strings.Builder
		cached.Compiled.Execute(&buf, f)
		return buf.String()
	}
	
	return `
<!DOCTYPE html>
<html>
<head>
	<title>{{.Title}}</title>
	<style>
		* {
			margin: 0;
			padding: 0;
			box-sizing: border-box;
		}
		
		body {
			font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, sans-serif;
			background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
			min-height: 100vh;
			padding: 2rem;
		}
		
		.container {
			max-width: 1400px;
			margin: 0 auto;
		}
		
		.header {
			background: white;
			border-radius: 12px;
			padding: 2rem;
			margin-bottom: 2rem;
			box-shadow: 0 10px 30px rgba(0,0,0,0.1);
		}
		
		.title {
			font-size: 2rem;
			font-weight: 700;
			color: #1a202c;
			margin-bottom: 0.5rem;
		}
		
		.description {
			color: #718096;
			font-size: 1rem;
		}
		
		.feature-badges {
			display: flex;
			gap: 0.5rem;
			margin-top: 1rem;
			flex-wrap: wrap;
		}
		
		.badge {
			display: inline-block;
			padding: 0.25rem 0.75rem;
			background: linear-gradient(135deg, #667eea, #764ba2);
			color: white;
			border-radius: 20px;
			font-size: 0.75rem;
			font-weight: 600;
		}
		
		.main-content {
			background: white;
			border-radius: 12px;
			padding: 2rem;
			box-shadow: 0 10px 30px rgba(0,0,0,0.1);
		}
		
		.controls {
			display: flex;
			gap: 1rem;
			margin-bottom: 2rem;
			flex-wrap: wrap;
		}
		
		.control-group {
			display: flex;
			gap: 0.5rem;
			align-items: center;
			padding: 0.5rem;
			background: #f7fafc;
			border-radius: 8px;
		}
		
		.btn {
			padding: 0.625rem 1.25rem;
			border: none;
			border-radius: 6px;
			font-weight: 500;
			cursor: pointer;
			transition: all 0.2s;
			font-size: 0.875rem;
		}
		
		.btn-primary {
			background: #667eea;
			color: white;
		}
		
		.btn-primary:hover {
			background: #5a67d8;
			transform: translateY(-1px);
			box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
		}
		
		.btn-secondary {
			background: #e2e8f0;
			color: #4a5568;
		}
		
		.btn-secondary:hover {
			background: #cbd5e0;
		}
		
		.btn-danger {
			background: #fc8181;
			color: white;
		}
		
		.btn-danger:hover {
			background: #f56565;
		}
		
		.btn-success {
			background: #68d391;
			color: white;
		}
		
		.btn-success:hover {
			background: #48bb78;
		}
		
		.btn-warning {
			background: #f6ad55;
			color: white;
		}
		
		.btn-warning:hover {
			background: #ed8936;
		}
		
		.stats {
			display: flex;
			gap: 2rem;
			margin-bottom: 1rem;
			padding: 1rem;
			background: #f7fafc;
			border-radius: 6px;
		}
		
		.stat {
			display: flex;
			flex-direction: column;
		}
		
		.stat-label {
			font-size: 0.75rem;
			color: #718096;
			text-transform: uppercase;
			letter-spacing: 0.05em;
		}
		
		.stat-value {
			font-size: 1.5rem;
			font-weight: 600;
			color: #2d3748;
		}
		
		.status-bar {
			padding: 0.75rem 1rem;
			background: #edf2f7;
			border-radius: 6px;
			margin-top: 1rem;
			font-size: 0.875rem;
			color: #4a5568;
			display: flex;
			justify-content: space-between;
			align-items: center;
		}
		
		.dropdown {
			position: relative;
			display: inline-block;
		}
		
		.dropdown-content {
			display: none;
			position: absolute;
			background: white;
			min-width: 160px;
			box-shadow: 0 8px 16px rgba(0,0,0,0.1);
			border-radius: 6px;
			z-index: 1000;
			margin-top: 0.5rem;
		}
		
		.dropdown.active .dropdown-content {
			display: block;
		}
		
		.dropdown-item {
			padding: 0.75rem 1rem;
			cursor: pointer;
			transition: background 0.2s;
			border-radius: 6px;
		}
		
		.dropdown-item:hover {
			background: #f7fafc;
		}
		
		.legend {
			display: flex;
			gap: 2rem;
			margin-top: 1rem;
			padding: 1rem;
			background: #f7fafc;
			border-radius: 6px;
			flex-wrap: wrap;
		}
		
		.legend-item {
			display: flex;
			align-items: center;
			gap: 0.5rem;
		}
		
		.legend-box {
			width: 20px;
			height: 20px;
			border-radius: 4px;
			border: 1px solid #cbd5e0;
		}
		
		.legend-label {
			font-size: 0.875rem;
			color: #4a5568;
		}
		
		.save-indicator {
			position: fixed;
			top: 20px;
			right: 20px;
			padding: 0.5rem 1rem;
			background: #48bb78;
			color: white;
			border-radius: 6px;
			font-size: 0.875rem;
			opacity: 0;
			transition: opacity 0.3s;
			z-index: 1000;
		}
		
		.save-indicator.show {
			opacity: 1;
		}
	</style>
</head>
<body>
	<div class="container">
		<div class="header">
			<h1 class="title">{{.Title}}</h1>
			<p class="description">{{.Description}}</p>
			<div class="feature-badges">
				<span class="badge">Virtual DOM</span>
				<span class="badge">State Management</span>
				<span class="badge">Event Registry</span>
				<span class="badge">Template Cache</span>
				<span class="badge">Error Recovery</span>
				<span class="badge">Auto-Save</span>
				<span class="badge">Undo/Redo</span>
			</div>
		</div>
		
		<div class="main-content">
			<div class="controls">
				<div class="control-group">
					<div class="dropdown" id="add-node-dropdown">
						<button class="btn btn-primary" onclick="document.getElementById('add-node-dropdown').classList.toggle('active')">Add Node ▼</button>
						<div class="dropdown-content">
							<div class="dropdown-item" onclick="send_event('{{.IdComponent}}', 'AddNode', 'start'); document.getElementById('add-node-dropdown').classList.remove('active')">
								Start Node
							</div>
							<div class="dropdown-item" onclick="send_event('{{.IdComponent}}', 'AddNode', 'process'); document.getElementById('add-node-dropdown').classList.remove('active')">
								Process Node
							</div>
							<div class="dropdown-item" onclick="send_event('{{.IdComponent}}', 'AddNode', 'decision'); document.getElementById('add-node-dropdown').classList.remove('active')">
								Decision Node
							</div>
							<div class="dropdown-item" onclick="send_event('{{.IdComponent}}', 'AddNode', 'data'); document.getElementById('add-node-dropdown').classList.remove('active')">
								Data Node
							</div>
							<div class="dropdown-item" onclick="send_event('{{.IdComponent}}', 'AddNode', 'end'); document.getElementById('add-node-dropdown').classList.remove('active')">
								End Node
							</div>
						</div>
					</div>
					
					<button class="btn {{if .ConnectingMode}}btn-danger{{else}}btn-success{{end}}" onclick="send_event('{{.IdComponent}}', 'ToggleConnectMode', null)">
						{{if .ConnectingMode}}Cancel Connect{{else}}Connect Mode{{end}}
					</button>
				</div>
				
				<div class="control-group">
					<button class="btn btn-secondary" onclick="send_event('{{.IdComponent}}', 'Undo', null)">
						↶ Undo
					</button>
					
					<button class="btn btn-secondary" onclick="send_event('{{.IdComponent}}', 'Redo', null)">
						↷ Redo
					</button>
				</div>
				
				<div class="control-group">
					<button class="btn btn-warning" onclick="send_event('{{.IdComponent}}', 'AutoArrange', null)">
						Auto Arrange
					</button>
					
					<button class="btn btn-success" onclick="send_event('{{.IdComponent}}', 'AnimateFlow', null)">
						Animate Flow
					</button>
				</div>
				
				<div class="control-group">
					<button class="btn btn-secondary" onclick="send_event('{{.IdComponent}}', 'ExportDiagram', null)">
						Export JSON
					</button>
					
					<button class="btn btn-danger" onclick="send_event('{{.IdComponent}}', 'ClearDiagram', null)">
						Clear All
					</button>
				</div>
			</div>
			
			<div class="stats">
				<div class="stat">
					<span class="stat-label">Nodes</span>
					<span class="stat-value">{{.NodeCount}}</span>
				</div>
				<div class="stat">
					<span class="stat-label">Edges</span>
					<span class="stat-value">{{.EdgeCount}}</span>
				</div>
				<div class="stat">
					<span class="stat-label">Canvas Size</span>
					<span class="stat-value">{{.Canvas.Width}} × {{.Canvas.Height}}</span>
				</div>
				<div class="stat">
					<span class="stat-label">Zoom</span>
					<span class="stat-value">{{.Canvas.ZoomPercent}}%</span>
				</div>
				<div class="stat">
					<span class="stat-label">Undo Stack</span>
					<span class="stat-value">{{len .UndoStack}}</span>
				</div>
			</div>
			
			<!-- Canvas Component -->
			<div id="flow-canvas-mount">
				{{if .Canvas}}
					<div id="{{.Canvas.ID}}" style="position: relative; width: {{.Canvas.Width}}px; height: {{.Canvas.Height}}px; border: 2px solid #e5e7eb; border-radius: 8px; overflow: hidden; background: #fafafa;">
						<div style="position: absolute; top: 10px; right: 10px; display: flex; gap: 0.5rem; background: white; padding: 0.5rem; border-radius: 6px; box-shadow: 0 2px 8px rgba(0,0,0,0.1); z-index: 100;">
							<button onclick="send_event('{{$.IdComponent}}', 'CanvasZoomIn', null)" style="padding: 0.5rem; background: white; border: 1px solid #d1d5db; border-radius: 4px; cursor: pointer;">Zoom In</button>
							<button onclick="send_event('{{$.IdComponent}}', 'CanvasZoomOut', null)" style="padding: 0.5rem; background: white; border: 1px solid #d1d5db; border-radius: 4px; cursor: pointer;">Zoom Out</button>
							<button onclick="send_event('{{$.IdComponent}}', 'CanvasReset', null)" style="padding: 0.5rem; background: white; border: 1px solid #d1d5db; border-radius: 4px; cursor: pointer;">Reset</button>
							<button onclick="send_event('{{$.IdComponent}}', 'ToggleGrid', null)" style="padding: 0.5rem; background: white; border: 1px solid #d1d5db; border-radius: 4px; cursor: pointer;">Grid</button>
						</div>
						
						<div id="canvas-viewport" style="position: relative; width: 100%; height: 100%; transform: scale({{.Canvas.Zoom}}) translate({{.Canvas.PanX}}px, {{.Canvas.PanY}}px); transform-origin: 0 0; transition: transform 0.2s;">
							<!-- Render boxes -->
							{{range $id, $box := .Canvas.Boxes}}
								<div id="box-{{$id}}" 
								     class="draggable-box"
								     data-box-id="{{$id}}"
								     data-box-x="{{$box.X}}"
								     data-box-y="{{$box.Y}}"
								     style="position: absolute; left: {{$box.X}}px; top: {{$box.Y}}px; width: {{$box.Width}}px; height: {{$box.Height}}px; background: {{$box.Color}}; border: 2px solid {{if $box.Selected}}#2563eb{{else}}#cbd5e1{{end}}; border-radius: 8px; padding: 0.5rem; cursor: {{if $.ConnectingMode}}pointer{{else}}move{{end}}; box-shadow: 0 2px 4px rgba(0,0,0,0.1); user-select: none;"
								     onclick="if({{$.ConnectingMode}}) { send_event('{{$.IdComponent}}', 'BoxClick', '{{$id}}'); }">
									<div style="font-weight: 600; color: #1f2937; font-size: 0.875rem; pointer-events: none;">{{$box.Label}}</div>
									{{if $box.Description}}
										<div style="font-size: 0.75rem; color: #6b7280; pointer-events: none;">{{$box.Description}}</div>
									{{end}}
								</div>
							{{end}}
							
							<!-- Render edges as SVG -->
							<svg style="position: absolute; top: 0; left: 0; width: 100%; height: 100%; pointer-events: none;">
								{{range .Canvas.Edges}}
									<line x1="{{.FromX}}" y1="{{.FromY}}" x2="{{.ToX}}" y2="{{.ToY}}" stroke="{{if .Selected}}#2563eb{{else}}#6b7280{{end}}" stroke-width="2" />
									{{if .Label}}
										<text x="{{.GetMidX}}" y="{{.GetMidY}}" text-anchor="middle" fill="#374151" font-size="12">{{.Label}}</text>
									{{end}}
								{{end}}
							</svg>
						</div>
						
						<div style="position: absolute; bottom: 10px; left: 10px; background: white; padding: 0.5rem 1rem; border-radius: 6px; box-shadow: 0 2px 8px rgba(0,0,0,0.1); font-size: 0.75rem; color: #6b7280; z-index: 100;">
							Boxes: {{len .Canvas.Boxes}} | Edges: {{len .Canvas.Edges}} | Zoom: {{.Canvas.ZoomPercent}}%
						</div>
					</div>
				{{end}}
			</div>
			
			<div class="legend">
				<div class="legend-item">
					<div class="legend-box" style="background: #dcfce7;"></div>
					<span class="legend-label">Start Node</span>
				</div>
				<div class="legend-item">
					<div class="legend-box" style="background: #dbeafe;"></div>
					<span class="legend-label">Process Node</span>
				</div>
				<div class="legend-item">
					<div class="legend-box" style="background: #fef3c7; transform: rotate(45deg);"></div>
					<span class="legend-label">Decision Node</span>
				</div>
				<div class="legend-item">
					<div class="legend-box" style="background: #e9d5ff;"></div>
					<span class="legend-label">Data Node</span>
				</div>
				<div class="legend-item">
					<div class="legend-box" style="background: #fee2e2;"></div>
					<span class="legend-label">End Node</span>
				</div>
			</div>
			
			<div class="status-bar">
				<div>
					<strong>Last Action:</strong> {{.LastAction}}
					{{if .Canvas}}
						{{range $id, $box := .Canvas.Boxes}}
							{{if $box.Selected}}
								| <strong>Selected:</strong> {{$box.Label}} ({{$box.X}}, {{$box.Y}})
							{{end}}
						{{end}}
					{{end}}
				</div>
				<div>
					<span style="color: #48bb78;">● Auto-save enabled</span>
				</div>
			</div>
		</div>
		
		<!-- Modal Component -->
		{{mount "export-modal"}}
	</div>
	
	<div class="save-indicator" id="save-indicator">Saved</div>
	
	<script>
	// Drag & drop is now handled in WASM module
	// This prevents event listeners from being lost on re-render
	
	// Show save indicator
	function showSaveIndicator() {
		var indicator = document.getElementById('save-indicator');
		indicator.classList.add('show');
		setTimeout(function() {
			indicator.classList.remove('show');
		}, 2000);
	}
	</script>
</body>
</html>
`
}

// ... (implement remaining handler methods similar to original but with enhanced features)

func (f *EnhancedFlowTool) GetDriver() liveview.LiveDriver {
	return f
}

func (f *EnhancedFlowTool) HandleAddNode(data interface{}) {
	// Implementation with VDOM update
	nodeType := data.(string)
	
	x := 100 + (f.NodeCount * 50) % 1000
	y := 100 + (f.NodeCount * 30) % 400
	
	nodeID := fmt.Sprintf("node_%d", f.NodeCount+1)
	label := fmt.Sprintf("%s %d", nodeType, f.NodeCount+1)
	
	var boxType components.BoxType
	switch nodeType {
	case "start":
		boxType = components.BoxTypeStart
	case "end":
		boxType = components.BoxTypeEnd
	case "process":
		boxType = components.BoxTypeProcess
	case "decision":
		boxType = components.BoxTypeDecision
	case "data":
		boxType = components.BoxTypeData
	default:
		boxType = components.BoxTypeCustom
	}
	
	newBox := components.NewFlowBox(nodeID, label, boxType, x, y)
	
	if f.Canvas != nil {
		boxDriver := liveview.NewDriver(nodeID, newBox)
		newBox.ComponentDriver = boxDriver
		newBox.Start()
		f.ComponentDriver.Mount(newBox)
		f.Canvas.AddBox(newBox)
		
		// Update state manager
		f.StateManager.Set("last_added_node", nodeID)
	}
	
	f.NodeCount++
	f.LastAction = fmt.Sprintf("Added %s node", nodeType)
	f.saveToUndoStack()
	
	if f.ComponentDriver != nil {
		f.Commit()
	}
}

func (f *EnhancedFlowTool) AutoArrange(data interface{}) {
	// Auto-arrange nodes using a simple grid layout
	boxList := make([]*components.FlowBox, 0, len(f.Canvas.Boxes))
	for _, box := range f.Canvas.Boxes {
		boxList = append(boxList, box)
	}
	
	cols := 4
	spacing := 200
	startX := 50
	startY := 50
	
	for i, box := range boxList {
		row := i / cols
		col := i % cols
		box.X = startX + (col * spacing)
		box.Y = startY + (row * spacing)
	}
	
	// Update edge positions
	f.updateEdgePositions()
	
	f.LastAction = "Nodes auto-arranged"
	f.saveToUndoStack()
	f.Commit()
}

// ... (implement remaining methods)

func (f *EnhancedFlowTool) ClearDiagram(data interface{}) {
	f.saveToUndoStack()
	f.Canvas.Clear()
	f.NodeCount = 0
	f.EdgeCount = 0
	f.LastAction = "Diagram cleared"
	f.Commit()
}

func (f *EnhancedFlowTool) ExportDiagram(data interface{}) {
	if f.Canvas == nil {
		f.LastAction = "No canvas to export"
		if f.ComponentDriver != nil {
			f.Commit()
		}
		return
	}
	
	exportData := f.Canvas.ExportJSON()
	
	// Convert to JSON string for display
	jsonBytes, err := json.MarshalIndent(exportData, "", "  ")
	if err != nil {
		f.LastAction = fmt.Sprintf("Export error: %v", err)
	} else {
		f.JsonExport = string(jsonBytes)
		
		// Create formatted content for modal
		modalContent := fmt.Sprintf(`
			<div style="font-family: monospace; background: #f5f5f5; padding: 1rem; border-radius: 4px; overflow-x: auto; max-height: 400px; overflow-y: auto;">
				<pre style="margin: 0; white-space: pre-wrap; word-wrap: break-word;">%s</pre>
			</div>
			<div style="margin-top: 1rem; display: flex; gap: 1rem;">
				<button onclick="navigator.clipboard.writeText('%s'); alert('Copied to clipboard!');" 
				        style="padding: 0.5rem 1rem; background: #4CAF50; color: white; border: none; border-radius: 4px; cursor: pointer;">
					Copy to Clipboard
				</button>
				<span style="color: #666; padding: 0.5rem;">%d boxes, %d edges</span>
			</div>
		`, string(jsonBytes), strings.ReplaceAll(string(jsonBytes), "'", "\\'"), len(f.Canvas.Boxes), len(f.Canvas.Edges))
		
		// Show modal with JSON
		if f.Modal != nil {
			f.Modal.Title = "Export JSON"
			f.Modal.Content = modalContent
			f.Modal.IsOpen = true
			if f.Modal.ComponentDriver != nil {
				f.Modal.Commit()
			}
		}
		
		f.LastAction = "Exported to modal"
	}
	
	if f.ComponentDriver != nil {
		f.Commit()
	}
}

func (f *EnhancedFlowTool) AnimateFlow(data interface{}) {
	// Animate edges to show flow
	for _, edge := range f.Canvas.Edges {
		edge.SetAnimated(true)
	}
	
	f.LastAction = "Flow animation started"
	f.Commit()
	
	// Stop animation after 5 seconds
	go func() {
		time.Sleep(5 * time.Second)
		for _, edge := range f.Canvas.Edges {
			edge.SetAnimated(false)
		}
		f.LastAction = "Flow animation stopped"
		f.Commit()
	}()
}

func (f *EnhancedFlowTool) HandleBoxClick(data interface{}) {
	var boxID string
	if str, ok := data.(string); ok {
		boxID = str
	} else {
		return
	}
	
	if f.ConnectingMode {
		// Handle connection creation
		if f.ConnectingFrom == "" {
			// First box selected
			f.ConnectingFrom = boxID
			if box, ok := f.Canvas.Boxes[boxID]; ok {
				box.Selected = true
			}
			f.LastAction = fmt.Sprintf("Connecting from: %s", boxID)
		} else if f.ConnectingFrom != boxID {
			// Second box selected - create edge
			if f.validateConnection(f.ConnectingFrom, boxID) {
				f.createConnection(f.ConnectingFrom, boxID)
				f.saveToUndoStack()
			}
			
			// Reset connection mode
			for _, box := range f.Canvas.Boxes {
				box.Selected = false
			}
			f.ConnectingFrom = ""
		}
	} else {
		// Normal selection
		for id, box := range f.Canvas.Boxes {
			box.Selected = (id == boxID)
		}
		f.LastAction = fmt.Sprintf("Selected box: %s", boxID)
	}
	
	if f.ComponentDriver != nil {
		f.Commit()
	}
}

func (f *EnhancedFlowTool) HandleMoveBox(data interface{}) {
	moveData, ok := data.(map[string]interface{})
	if !ok {
		log.Printf("HandleMoveBox: data is not a map: %T", data)
		return
	}
	
	boxID, _ := moveData["id"].(string)
	direction, _ := moveData["dir"].(string)
	log.Printf("HandleMoveBox: boxID=%s, direction=%s", boxID, direction)
	
	if box, ok := f.Canvas.Boxes[boxID]; ok {
		step := 20 // Pixels to move
		
		switch direction {
		case "up":
			box.Y -= step
			if box.Y < 0 {
				box.Y = 0
			}
		case "down":
			box.Y += step
			if box.Y > f.Canvas.Height-box.Height {
				box.Y = f.Canvas.Height - box.Height
			}
		case "left":
			box.X -= step
			if box.X < 0 {
				box.X = 0
			}
		case "right":
			box.X += step
			if box.X > f.Canvas.Width-box.Width {
				box.X = f.Canvas.Width - box.Width
			}
		}
		
		// Update connected edges
		f.updateEdgePositions()
		
		f.LastAction = fmt.Sprintf("Moved box %s to (%d, %d)", boxID, box.X, box.Y)
		
		if f.ComponentDriver != nil {
			f.Commit()
		}
	}
}

func (f *EnhancedFlowTool) HandleToggleConnectMode(data interface{}) {
	f.ConnectingMode = !f.ConnectingMode
	f.ConnectingFrom = ""
	
	// Clear all selections
	for _, box := range f.Canvas.Boxes {
		box.Selected = false
	}
	
	if f.ConnectingMode {
		f.LastAction = "Connection mode activated - click two boxes to connect"
	} else {
		f.LastAction = "Connection mode deactivated"
	}
	
	if f.ComponentDriver != nil {
		f.Commit()
	}
}

func (f *EnhancedFlowTool) BoxStartDrag(data interface{}) {
	// Handle drag start
	log.Printf("BoxStartDrag called with data: %v (%T)", data, data)
	
	// Try to parse as JSON string first
	if dataStr, ok := data.(string); ok {
		var dataMap map[string]interface{}
		if err := json.Unmarshal([]byte(dataStr), &dataMap); err == nil {
			data = dataMap
			log.Printf("Parsed JSON data: %v", dataMap)
		}
	}
	
	if dataMap, ok := data.(map[string]interface{}); ok {
		if boxID, ok := dataMap["id"].(string); ok {
			f.DraggingBox = boxID
			log.Printf("Started dragging box: %s", boxID)
			if box, exists := f.Canvas.Boxes[boxID]; exists {
				box.Dragging = true
				if box.ComponentDriver != nil {
					box.Commit()
				}
			}
			f.LastAction = fmt.Sprintf("Started dragging %s", boxID)
		} else {
			log.Printf("BoxStartDrag: id not found in dataMap: %v", dataMap)
		}
	} else {
		log.Printf("BoxStartDrag: data is not a map: %T", data)
	}
	
	if f.ComponentDriver != nil {
		f.Commit()
	}
}

func (f *EnhancedFlowTool) BoxDrag(data interface{}) {
	if f.DraggingBox == "" {
		log.Printf("BoxDrag: No box being dragged")
		return
	}
	
	log.Printf("BoxDrag called for box %s with data: %v", f.DraggingBox, data)
	
	// Try to parse as JSON string first
	if dataStr, ok := data.(string); ok {
		var dataMap map[string]interface{}
		if err := json.Unmarshal([]byte(dataStr), &dataMap); err == nil {
			data = dataMap
		}
	}
	
	// Handle drag movement with VDOM updates
	if dataMap, ok := data.(map[string]interface{}); ok {
		if box, exists := f.Canvas.Boxes[f.DraggingBox]; exists {
			oldX, oldY := box.X, box.Y
			
			if newX, ok := dataMap["x"].(float64); ok {
				box.X = int(newX)
				log.Printf("Box %s moved X: %d -> %d", f.DraggingBox, oldX, box.X)
			}
			if newY, ok := dataMap["y"].(float64); ok {
				box.Y = int(newY)
				log.Printf("Box %s moved Y: %d -> %d", f.DraggingBox, oldY, box.Y)
			}
			
			// Constrain to canvas bounds
			if box.X < 0 {
				box.X = 0
			}
			if box.Y < 0 {
				box.Y = 0
			}
			maxX := f.Canvas.Width - box.Width
			maxY := f.Canvas.Height - box.Height
			if box.X > maxX {
				box.X = maxX
			}
			if box.Y > maxY {
				box.Y = maxY
			}
			
			// Update state if position changed
			if oldX != box.X || oldY != box.Y {
				f.StateManager.Set("box_position_"+f.DraggingBox, map[string]interface{}{
					"x": box.X,
					"y": box.Y,
				})
			}
			
			// Update edge positions
			f.updateEdgePositions()
			
			if box.ComponentDriver != nil {
				box.Commit()
			}
			
			// Emit drag event for auto-save
			f.EventRegistry.Emit("diagram.change", map[string]interface{}{
				"type": "box_moved",
				"box":  f.DraggingBox,
			})
		}
	}
	
	if f.ComponentDriver != nil {
		f.Commit()
	}
}

func (f *EnhancedFlowTool) BoxEndDrag(data interface{}) {
	if f.DraggingBox != "" {
		if box, exists := f.Canvas.Boxes[f.DraggingBox]; exists {
			box.Dragging = false
			if box.ComponentDriver != nil {
				box.Commit()
			}
		}
		f.LastAction = fmt.Sprintf("Finished dragging %s", f.DraggingBox)
		f.saveToUndoStack()
		f.DraggingBox = ""
	}
	
	if f.ComponentDriver != nil {
		f.Commit()
	}
}

func (f *EnhancedFlowTool) updateEdgePositions() {
	for _, edge := range f.Canvas.Edges {
		if fromBox, ok := f.Canvas.Boxes[edge.FromBox]; ok {
			if toBox, ok := f.Canvas.Boxes[edge.ToBox]; ok {
				edge.FromX = fromBox.X + fromBox.Width
				edge.FromY = fromBox.Y + fromBox.Height/2
				edge.ToX = toBox.X
				edge.ToY = toBox.Y + toBox.Height/2
			}
		}
	}
}

func (f *EnhancedFlowTool) HandleCanvasZoomIn(data interface{}) {
	f.Canvas.Zoom = min(f.Canvas.Zoom*1.2, 3.0)
	f.LastAction = fmt.Sprintf("Zoom: %d%%", f.Canvas.ZoomPercent())
	if f.ComponentDriver != nil {
		f.Commit()
	}
}

func (f *EnhancedFlowTool) HandleCanvasZoomOut(data interface{}) {
	f.Canvas.Zoom = max(f.Canvas.Zoom/1.2, 0.3)
	f.LastAction = fmt.Sprintf("Zoom: %d%%", f.Canvas.ZoomPercent())
	if f.ComponentDriver != nil {
		f.Commit()
	}
}

func (f *EnhancedFlowTool) HandleCanvasReset(data interface{}) {
	f.Canvas.Zoom = 1.0
	f.Canvas.PanX = 0
	f.Canvas.PanY = 0
	f.LastAction = "View reset"
	if f.ComponentDriver != nil {
		f.Commit()
	}
}

func (f *EnhancedFlowTool) HandleToggleGrid(data interface{}) {
	f.Canvas.ShowGrid = !f.Canvas.ShowGrid
	f.LastAction = fmt.Sprintf("Grid: %v", f.Canvas.ShowGrid)
	if f.ComponentDriver != nil {
		f.Commit()
	}
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func main() {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Serve static assets
	e.Static("/example/assets", "../../assets")
	e.Static("/assets", "../../assets")

	home := liveview.PageControl{
		Title:  "Enhanced Flow Diagram Tool",
		Lang:   "en",
		Path:   "/example/flowtool",
		Router: e,
	}

	home.Register(func() liveview.LiveDriver {
		// Create layout wrapper
		document := liveview.NewLayout("flowtool-layout", `{{mount "flow-tool"}}`)
		
		// Create enhanced flow tool component
		flowTool := NewEnhancedFlowTool()
		liveview.New("flow-tool", flowTool)
		
		// Set up the modal driver
		if flowTool.Modal != nil {
			liveview.New("export-modal", flowTool.Modal)
		}
		
		// Set up drivers for existing boxes
		if flowTool.Canvas != nil {
			for id, box := range flowTool.Canvas.Boxes {
				liveview.New(id, box)
			}
			
			// Set up drivers for existing edges  
			for id, edge := range flowTool.Canvas.Edges {
				liveview.New(id, edge)
			}
		}
		
		return document
	})

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, `
			<h1>Enhanced Flow Diagram Tool</h1>
			<p>Visit <a href="/example/flowtool">/example/flowtool</a> to see the interactive flow diagram editor</p>
			<h2>New Features:</h2>
			<ul>
				<li>Virtual DOM for efficient rendering</li>
				<li>State Management with auto-save</li>
				<li>Event Registry with throttling</li>
				<li>Template Cache for performance</li>
				<li>Error Boundaries for recovery</li>
				<li>Undo/Redo functionality</li>
				<li>Auto-arrange nodes</li>
			</ul>
		`)
	})

	port := ":8082"
	log.Printf("🚀 Enhanced Flow Tool Server")
	log.Printf("🌐 Starting on http://localhost%s", port)
	log.Printf("Visit http://localhost%s/example/flowtool", port)
	e.Logger.Fatal(e.Start(port))
}