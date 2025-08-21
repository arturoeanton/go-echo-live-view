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
	FileUpload     *components.FileUpload
	Title          string
	Description    string
	NodeCount      int
	EdgeCount      int
	LastAction     string
	ConnectingMode bool
	ConnectingFrom string
	DraggingBox    string
	JsonExport     string
	
	// Edit mode fields
	EditingMode    bool
	EditingType    string // "box" or "edge"
	EditingID      string
	EditingValue   string
	EditingCode    string  // Code metadata for boxes
	
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
	
	// Create file upload component for JSON import
	fileUpload := &components.FileUpload{
		Label:    "Import JSON Diagram",
		Accept:   ".json",
		Multiple: false,
		MaxSize:  5 * 1024 * 1024, // 5MB max
		MaxFiles: 1,
	}
	
	tool := &EnhancedFlowTool{
		Canvas:         canvas,
		Modal:          modal,
		FileUpload:     fileUpload,
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
	
	// Configure file upload callback
	fileUpload.OnUpload = func(files []components.FileInfo) error {
		if len(files) > 0 {
			// Get the raw file data (base64 decoded)
			fileData, err := fileUpload.GetFileData(0)
			if err != nil {
				return err
			}
			// Import the diagram using the decoded JSON content
			tool.ImportDiagram(string(fileData))
			// Clear the file upload after successful import
			fileUpload.Clear()
		}
		return nil
	}
	
	// Register event handlers with throttling and debouncing
	tool.setupEnhancedEventHandlers()
	
	// Load saved state if available
	tool.loadSavedState()
	
	// Disable auto-save for debugging
	// tool.startAutoSave()
	
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
	// Save current state and clear redo stack (for normal operations)
	f.saveCurrentStateToUndoStack()
	// Clear redo stack on new action
	f.RedoStack = []string{}
}

func (f *EnhancedFlowTool) saveCurrentStateToUndoStack() {
	// Save current state WITHOUT clearing redo stack (for undo/redo operations)
	// Create a serializable state
	boxesData := make(map[string]map[string]interface{})
	for id, box := range f.Canvas.Boxes {
		boxData := map[string]interface{}{
			"id":          box.ID,
			"label":       box.Label,
			"description": box.Description,
			"type":        string(box.Type),
			"x":           box.X,
			"y":           box.Y,
			"width":       box.Width,
			"height":      box.Height,
			"color":       box.Color,
			"selected":    box.Selected,
		}
		// Include Data field with code
		if box.Data != nil {
			boxData["data"] = box.Data
		}
		boxesData[id] = boxData
	}
	
	edgesData := make(map[string]map[string]interface{})
	for id, edge := range f.Canvas.Edges {
		edgesData[id] = map[string]interface{}{
			"id":       edge.ID,
			"fromBox":  edge.FromBox,
			"fromPort": edge.FromPort,
			"toBox":    edge.ToBox,
			"toPort":   edge.ToPort,
			"label":    edge.Label,
			"type":     string(edge.Type),
			"fromX":    edge.FromX,
			"fromY":    edge.FromY,
			"toX":      edge.ToX,
			"toY":      edge.ToY,
		}
	}
	
	state := map[string]interface{}{
		"boxes":     boxesData,
		"edges":     edgesData,
		"nodeCount": f.NodeCount,
		"edgeCount": f.EdgeCount,
	}
	
	stateJSON, _ := json.Marshal(state)
	f.UndoStack = append(f.UndoStack, string(stateJSON))
	
	// Limit undo stack size
	if len(f.UndoStack) > 50 {
		f.UndoStack = f.UndoStack[1:]
	}
}

func (f *EnhancedFlowTool) Undo(data interface{}) {
	if len(f.UndoStack) == 0 {
		f.LastAction = "Nothing to undo"
		f.Commit()
		return
	}
	
	// Save current state to redo stack before undoing
	f.saveToRedoStack()
	
	// Get and remove the last state from undo stack
	prevState := f.UndoStack[len(f.UndoStack)-1]
	f.UndoStack = f.UndoStack[:len(f.UndoStack)-1]
	
	// Restore the state
	f.restoreState(prevState)
	
	f.LastAction = "Undo performed"
	f.Commit()
}

func (f *EnhancedFlowTool) saveToRedoStack() {
	// Save current state to redo stack (same as saveToUndoStack but for redo)
	boxesData := make(map[string]map[string]interface{})
	for id, box := range f.Canvas.Boxes {
		boxData := map[string]interface{}{
			"id":          box.ID,
			"label":       box.Label,
			"description": box.Description,
			"type":        string(box.Type),
			"x":           box.X,
			"y":           box.Y,
			"width":       box.Width,
			"height":      box.Height,
			"color":       box.Color,
			"selected":    box.Selected,
		}
		if box.Data != nil {
			boxData["data"] = box.Data
		}
		boxesData[id] = boxData
	}
	
	edgesData := make(map[string]map[string]interface{})
	for id, edge := range f.Canvas.Edges {
		edgesData[id] = map[string]interface{}{
			"id":       edge.ID,
			"fromBox":  edge.FromBox,
			"fromPort": edge.FromPort,
			"toBox":    edge.ToBox,
			"toPort":   edge.ToPort,
			"label":    edge.Label,
			"type":     string(edge.Type),
			"fromX":    edge.FromX,
			"fromY":    edge.FromY,
			"toX":      edge.ToX,
			"toY":      edge.ToY,
		}
	}
	
	state := map[string]interface{}{
		"boxes":     boxesData,
		"edges":     edgesData,
		"nodeCount": f.NodeCount,
		"edgeCount": f.EdgeCount,
	}
	
	stateJSON, _ := json.Marshal(state)
	f.RedoStack = append(f.RedoStack, string(stateJSON))
}

func (f *EnhancedFlowTool) restoreState(stateJSON string) {
	var state map[string]interface{}
	if err := json.Unmarshal([]byte(stateJSON), &state); err != nil {
		log.Printf("Error unmarshaling state: %v", err)
		return
	}
	
	// Clear current canvas
	f.Canvas.Boxes = make(map[string]*components.FlowBox)
	f.Canvas.Edges = make(map[string]*components.FlowEdge)
	
	// Restore boxes
	if boxesData, ok := state["boxes"].(map[string]interface{}); ok {
		for id, boxData := range boxesData {
			if boxMap, ok := boxData.(map[string]interface{}); ok {
				// Get box type
				boxType := components.BoxTypeProcess
				if typeStr, ok := boxMap["type"].(string); ok {
					boxType = components.BoxType(typeStr)
				}
				
				// Get dimensions with defaults
				width := 120
				height := 60
				if w, ok := boxMap["width"].(float64); ok && w > 0 {
					width = int(w)
				}
				if h, ok := boxMap["height"].(float64); ok && h > 0 {
					height = int(h)
				}
				
				// Create new box
				box := &components.FlowBox{
					ID:     id,
					Label:  boxMap["label"].(string),
					Type:   boxType,
					X:      int(boxMap["x"].(float64)),
					Y:      int(boxMap["y"].(float64)),
					Width:  width,
					Height: height,
				}
				
				// Restore optional fields
				if desc, ok := boxMap["description"].(string); ok {
					box.Description = desc
				}
				if color, ok := boxMap["color"].(string); ok {
					box.Color = color
				} else {
					// Set default color based on type
					switch boxType {
					case components.BoxTypeStart:
						box.Color = "#dcfce7"
					case components.BoxTypeEnd:
						box.Color = "#fee2e2"
					case components.BoxTypeDecision:
						box.Color = "#fef3c7"
					case components.BoxTypeData:
						box.Color = "#e9d5ff"
					default:
						box.Color = "#dbeafe"
					}
				}
				
				// Restore Data field (including code)
				if data, ok := boxMap["data"].(map[string]interface{}); ok {
					box.Data = data
				}
				
				// Register driver
				liveview.New(id, box)
				f.Canvas.Boxes[id] = box
			}
		}
	}
	
	// Restore edges
	if edgesData, ok := state["edges"].(map[string]interface{}); ok {
		for id, edgeData := range edgesData {
			if edgeMap, ok := edgeData.(map[string]interface{}); ok {
				edge := &components.FlowEdge{
					ID:       id,
					FromBox:  edgeMap["fromBox"].(string),
					FromPort: edgeMap["fromPort"].(string),
					ToBox:    edgeMap["toBox"].(string),
					ToPort:   edgeMap["toPort"].(string),
				}
				
				// Restore optional fields
				if label, ok := edgeMap["label"].(string); ok {
					edge.Label = label
				}
				if typeStr, ok := edgeMap["type"].(string); ok {
					edge.Type = components.EdgeType(typeStr)
				}
				
				// Restore positions
				if x, ok := edgeMap["fromX"].(float64); ok {
					edge.FromX = int(x)
				}
				if y, ok := edgeMap["fromY"].(float64); ok {
					edge.FromY = int(y)
				}
				if x, ok := edgeMap["toX"].(float64); ok {
					edge.ToX = int(x)
				}
				if y, ok := edgeMap["toY"].(float64); ok {
					edge.ToY = int(y)
				}
				
				f.Canvas.Edges[id] = edge
			}
		}
	}
	
	// Restore counts
	if nodeCount, ok := state["nodeCount"].(float64); ok {
		f.NodeCount = int(nodeCount)
	}
	if edgeCount, ok := state["edgeCount"].(float64); ok {
		f.EdgeCount = int(edgeCount)
	}
}

func (f *EnhancedFlowTool) Redo(data interface{}) {
	if len(f.RedoStack) == 0 {
		f.LastAction = "Nothing to redo"
		f.Commit()
		return
	}
	
	// Save current state to undo stack WITHOUT clearing redo stack
	f.saveCurrentStateToUndoStack()
	
	// Get and remove the last state from redo stack
	nextState := f.RedoStack[len(f.RedoStack)-1]
	f.RedoStack = f.RedoStack[:len(f.RedoStack)-1]
	
	// Restore the state
	f.restoreState(nextState)
	
	f.LastAction = "Redo performed"
	f.Commit()
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
	
	// Initialize file upload component
	if f.FileUpload != nil && f.FileUpload.ComponentDriver != nil {
		f.FileUpload.Start()
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
		f.ComponentDriver.Events["ImportDiagram"] = func(c *EnhancedFlowTool, data interface{}) {
			c.ImportDiagram(data)
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
			log.Printf("📨 BoxClick event received: %v", data)
			c.HandleBoxClick(data)
		}
		f.ComponentDriver.Events["DeleteBox"] = func(c *EnhancedFlowTool, data interface{}) {
			c.DeleteBox(data)
		}
		f.ComponentDriver.Events["DeleteEdge"] = func(c *EnhancedFlowTool, data interface{}) {
			c.DeleteEdge(data)
		}
		f.ComponentDriver.Events["SelectEdge"] = func(c *EnhancedFlowTool, data interface{}) {
			c.SelectEdge(data)
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
			log.Printf("📨 ToggleConnectMode event received: %v", data)
			c.HandleToggleConnectMode(data)
		}
		
		// Generic drag events from WASM
		f.ComponentDriver.Events["DragStart"] = func(c *EnhancedFlowTool, data interface{}) {
			log.Printf("📨 DragStart event received: %v", data)
			c.HandleDragStart(data)
		}
		f.ComponentDriver.Events["DragMove"] = func(c *EnhancedFlowTool, data interface{}) {
			c.HandleDragMove(data)
		}
		f.ComponentDriver.Events["DragEnd"] = func(c *EnhancedFlowTool, data interface{}) {
			c.HandleDragEnd(data)
		}
		
		// Backward compatibility: specific box drag events
		f.ComponentDriver.Events["BoxStartDrag"] = func(c *EnhancedFlowTool, data interface{}) {
			log.Printf("📨 BoxStartDrag event received: %v", data)
			c.HandleBoxStartDrag(data)
		}
		f.ComponentDriver.Events["BoxDrag"] = func(c *EnhancedFlowTool, data interface{}) {
			c.HandleBoxDrag(data)
		}
		f.ComponentDriver.Events["BoxEndDrag"] = func(c *EnhancedFlowTool, data interface{}) {
			c.HandleBoxEndDrag(data)
		}
		
		// Edit events
		f.ComponentDriver.Events["EditBox"] = func(c *EnhancedFlowTool, data interface{}) {
			c.HandleEditBox(data)
		}
		f.ComponentDriver.Events["EditEdge"] = func(c *EnhancedFlowTool, data interface{}) {
			c.HandleEditEdge(data)
		}
		f.ComponentDriver.Events["SaveEdit"] = func(c *EnhancedFlowTool, data interface{}) {
			c.HandleSaveEdit(data)
		}
		f.ComponentDriver.Events["CancelEdit"] = func(c *EnhancedFlowTool, data interface{}) {
			c.HandleCancelEdit(data)
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
		
		.draggable-box:hover .delete-box-btn {
			display: block !important;
		}
		
		.delete-box-btn:hover {
			background: #dc2626 !important;
			transform: scale(1.1);
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
				</div>
				
				<div class="control-group">
					<button class="btn btn-secondary" onclick="send_event('{{.IdComponent}}', 'ExportDiagram', null)">
						📥 Export JSON
					</button>
					
					<button class="btn btn-secondary" onclick="document.getElementById('import-file-input').click()">
						📤 Import JSON
					</button>
					<input id="import-file-input" type="file" accept=".json" style="display: none;" 
						onchange="if(this.files[0]) { 
							var reader = new FileReader(); 
							reader.onload = function(e) { 
								send_event('{{.IdComponent}}', 'ImportDiagram', e.target.result); 
							}; 
							reader.readAsText(this.files[0]); 
						}">
					
					<button class="btn btn-danger" onclick="send_event('{{.IdComponent}}', 'ClearDiagram', null)">
						🗑️ Clear All
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
							{{$component := .}}
							{{range $id, $box := .Canvas.Boxes}}
								<div id="box-{{$id}}" 
								     class="draggable draggable-box flow-box"
								     data-element-id="box-{{$id}}"
								     data-component-id="{{$component.IdComponent}}"
								     data-box-id="{{$id}}"
								     data-box-x="{{$box.X}}"
								     data-box-y="{{$box.Y}}"
								     {{if $component.ConnectingMode}}data-drag-disabled="true"{{end}}
								     style="position: absolute; left: {{$box.X}}px; top: {{$box.Y}}px; width: {{$box.Width}}px; height: {{$box.Height}}px; background: {{$box.Color}}; border: 2px solid {{if $box.Selected}}#2563eb{{else}}#cbd5e1{{end}}; border-radius: 8px; padding: 0.5rem; cursor: {{if $component.ConnectingMode}}pointer{{else}}move{{end}}; box-shadow: 0 2px 4px rgba(0,0,0,0.1); user-select: none; z-index: 20;"
								     onclick="if({{$component.ConnectingMode}}) { send_event('{{$component.IdComponent}}', 'BoxClick', '{{$id}}'); }"
								     ondblclick="send_event('{{$component.IdComponent}}', 'EditBox', '{{$id}}'); event.stopPropagation();">
									<button class="delete-box-btn" 
									        onclick="event.stopPropagation(); if(confirm('Delete this box?')) { send_event('{{$component.IdComponent}}', 'DeleteBox', '{{$id}}'); }"
									        style="position: absolute; top: -8px; right: -8px; width: 20px; height: 20px; border-radius: 50%; background: #ef4444; color: white; border: 2px solid white; cursor: pointer; font-size: 12px; line-height: 1; padding: 0; display: {{if $box.Selected}}block{{else}}none{{end}}; z-index: 10;">
									        ×
									</button>
									<div style="font-weight: 600; color: #1f2937; font-size: 0.875rem; pointer-events: none;">{{$box.Label}}</div>
									{{if $box.Description}}
										<div style="font-size: 0.75rem; color: #6b7280; pointer-events: none;">{{$box.Description}}</div>
									{{end}}
								</div>
							{{end}}
							
							<!-- Render edges as SVG -->
							<svg style="position: absolute; top: 0; left: 0; width: 100%; height: 100%; pointer-events: auto;">
								{{range $id, $edge := .Canvas.Edges}}
									<g class="edge-group">
										<line x1="{{$edge.FromX}}" y1="{{$edge.FromY}}" x2="{{$edge.ToX}}" y2="{{$edge.ToY}}" 
										      stroke="{{if $edge.Selected}}#2563eb{{else}}#6b7280{{end}}" 
										      stroke-width="{{if $edge.Selected}}3{{else}}2{{end}}"
										      style="cursor: pointer;"
										      onclick="send_event('{{$.IdComponent}}', 'SelectEdge', '{{$id}}')"
										      ondblclick="send_event('{{$.IdComponent}}', 'EditEdge', '{{$id}}'); event.stopPropagation();" />
										{{if $edge.Label}}
											<text x="{{$edge.GetMidX}}" y="{{$edge.GetMidY}}" text-anchor="middle" fill="#374151" font-size="12">{{$edge.Label}}</text>
										{{end}}
										{{if $edge.Selected}}
											<circle cx="{{$edge.GetMidX}}" cy="{{$edge.GetMidY}}" r="10" 
											        fill="#ef4444" 
											        style="cursor: pointer;"
											        onclick="event.stopPropagation(); if(confirm('Delete this connection?')) { send_event('{{$.IdComponent}}', 'DeleteEdge', '{{$id}}'); }">
											</circle>
											<text x="{{$edge.GetMidX}}" y="{{$edge.GetMidY}}" 
											      text-anchor="middle" 
											      fill="white" 
											      font-size="14" 
											      font-weight="bold"
											      pointer-events="none">×</text>
										{{end}}
									</g>
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
		
		<!-- Edit Modal -->
		{{if .EditingMode}}
		<div style="position: fixed; top: 0; left: 0; right: 0; bottom: 0; background: rgba(0,0,0,0.5); display: flex; align-items: center; justify-content: center; z-index: 2000; overflow-y: auto;">
			<div style="background: white; padding: 2rem; border-radius: 8px; min-width: 600px; max-width: 800px; max-height: 90vh; overflow-y: auto; box-shadow: 0 10px 30px rgba(0,0,0,0.2); margin: 2rem;">
				<h3 style="margin-top: 0; color: #333;">
					{{if eq .EditingType "box"}}Edit Node{{else}}Edit Edge Label{{end}}
				</h3>
				
				<label style="display: block; margin-bottom: 0.5rem; color: #666; font-weight: 500;">
					{{if eq .EditingType "box"}}Node Name:{{else}}Edge Label:{{end}}
				</label>
				<input id="edit-input" type="text" value="{{.EditingValue}}" 
					style="width: 100%; padding: 0.75rem; border: 2px solid #ddd; border-radius: 4px; font-size: 1rem; margin-bottom: 1rem;"
					{{if eq .EditingType "edge"}}onkeypress="if(event.key === 'Enter') { event.preventDefault(); document.getElementById('save-edit-btn').click(); }"{{end}}>
				
				{{if eq .EditingType "box"}}
				<label style="display: block; margin-bottom: 0.5rem; color: #666; font-weight: 500;">
					Code/Script:
				</label>
				<textarea id="edit-code" 
					style="width: 100%; min-height: 200px; padding: 0.75rem; border: 2px solid #ddd; border-radius: 4px; font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace; font-size: 0.9rem; margin-bottom: 1rem; resize: vertical; background: #f8f9fa;"
					placeholder="// Enter code, script, or notes here..."
					>{{.EditingCode}}</textarea>
				
				<div style="display: flex; gap: 1rem; margin-bottom: 1rem;">
					<button class="btn btn-secondary" onclick="document.getElementById('edit-code').style.height = '400px';" style="padding: 0.5rem 1rem; font-size: 0.875rem;">
						↕ Expand
					</button>
					<button class="btn btn-secondary" onclick="
						var code = document.getElementById('edit-code');
						code.value = '// Function template\nfunction process() {\n    // Your code here\n    \n    return result;\n}';
						" style="padding: 0.5rem 1rem; font-size: 0.875rem;">
						📝 Template
					</button>
					<button class="btn btn-secondary" onclick="
						var code = document.getElementById('edit-code');
						code.select();
						document.execCommand('copy');
						alert('Code copied to clipboard!');
						" style="padding: 0.5rem 1rem; font-size: 0.875rem;">
						📋 Copy
					</button>
				</div>
				{{end}}
				
				<div style="display: flex; gap: 1rem; justify-content: flex-end;">
					<button class="btn btn-secondary" onclick="send_event('{{.IdComponent}}', 'CancelEdit', null)">
						Cancel
					</button>
					<button id="save-edit-btn" class="btn btn-primary" onclick="
						var data = {
							value: document.getElementById('edit-input').value
							{{if eq .EditingType "box"}},
							code: document.getElementById('edit-code').value
							{{end}}
						};
						send_event('{{.IdComponent}}', 'SaveEdit', JSON.stringify(data));
					">
						Save
					</button>
				</div>
			</div>
		</div>
		<script>
			// Focus the input when modal opens
			setTimeout(function() {
				var input = document.getElementById('edit-input');
				if(input) {
					input.focus();
					input.select();
				}
			}, 100);
		</script>
		{{end}}
	</div>
	
	<div class="save-indicator" id="save-indicator">Saved</div>
	
	<script>
	// Drag & drop is now handled in WASM module
	// This prevents event listeners from being lost on re-render
	
	// Handle keyboard shortcuts
	document.addEventListener('keydown', function(e) {
		// Delete key to delete selected box
		if (e.key === 'Delete' || e.key === 'Backspace') {
			e.preventDefault();
			// Find selected box
			{{range $id, $box := .Canvas.Boxes}}
				{{if $box.Selected}}
					if (confirm('Delete selected box?')) {
						send_event('{{.IdComponent}}', 'DeleteBox', '{{$id}}');
					}
				{{end}}
			{{end}}
		}
		
		// Escape key to exit connect mode
		if (e.key === 'Escape') {
			{{if .ConnectingMode}}
				send_event('{{.IdComponent}}', 'ToggleConnectMode', null);
			{{end}}
		}
	});
	
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
	// Save state BEFORE making changes
	f.saveToUndoStack()
	
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
		// Create and register the driver properly
		liveview.New(nodeID, newBox)
		f.Canvas.AddBox(newBox)
		
		// Update state manager
		f.StateManager.Set("last_added_node", nodeID)
	}
	
	f.NodeCount++
	f.LastAction = fmt.Sprintf("Added %s node", nodeType)
	
	if f.ComponentDriver != nil {
		f.Commit()
	}
}

func (f *EnhancedFlowTool) AutoArrange(data interface{}) {
	// Save state BEFORE making changes
	f.saveToUndoStack()
	
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
	
	// Create custom export data that includes code metadata
	boxes := []map[string]interface{}{}
	for _, box := range f.Canvas.Boxes {
		boxData := map[string]interface{}{
			"id":    box.ID,
			"label": box.Label,
			"type":  box.Type,
			"x":     box.X,
			"y":     box.Y,
		}
		
		// Include code if present
		if box.Data != nil {
			if code, ok := box.Data["code"].(string); ok && code != "" {
				boxData["code"] = code
			}
			// Include any other metadata
			for key, value := range box.Data {
				if key != "code" {
					boxData[key] = value
				}
			}
		}
		
		boxes = append(boxes, boxData)
	}
	
	edges := []map[string]interface{}{}
	for _, edge := range f.Canvas.Edges {
		edges = append(edges, map[string]interface{}{
			"id":       edge.ID,
			"fromBox":  edge.FromBox,
			"fromPort": edge.FromPort,
			"toBox":    edge.ToBox,
			"toPort":   edge.ToPort,
			"type":     edge.Type,
			"label":    edge.Label,
		})
	}
	
	exportData := map[string]interface{}{
		"boxes": boxes,
		"edges": edges,
		"metadata": map[string]interface{}{
			"exportTime": time.Now().Format(time.RFC3339),
			"version":    "1.0",
			"tool":       "Enhanced Flow Diagram Tool",
		},
	}
	
	// Convert to JSON string for display
	jsonBytes, err := json.MarshalIndent(exportData, "", "  ")
	if err != nil {
		f.LastAction = fmt.Sprintf("Export error: %v", err)
	} else {
		f.JsonExport = string(jsonBytes)
		
		// Create JavaScript to handle the download
		downloadScript := fmt.Sprintf(`
			var jsonData = %s;
			var dataStr = JSON.stringify(jsonData, null, 2);
			var dataUri = 'data:application/json;charset=utf-8,'+ encodeURIComponent(dataStr);
			var exportFileDefaultName = 'flow_diagram_%d.json';
			var linkElement = document.createElement('a');
			linkElement.setAttribute('href', dataUri);
			linkElement.setAttribute('download', exportFileDefaultName);
			linkElement.click();
		`, string(jsonBytes), time.Now().Unix())
		
		// Execute the download script
		if f.ComponentDriver != nil {
			f.ComponentDriver.EvalScript(downloadScript)
		}
		
		f.LastAction = fmt.Sprintf("Exported diagram with %d boxes and %d edges", len(f.Canvas.Boxes), len(f.Canvas.Edges))
	}
	
	if f.ComponentDriver != nil {
		f.Commit()
	}
}

func (f *EnhancedFlowTool) ImportDiagram(data interface{}) {
	// Parse JSON data
	var jsonStr string
	if str, ok := data.(string); ok {
		jsonStr = str
	} else {
		f.LastAction = "Invalid import data"
		f.Commit()
		return
	}
	
	var importData map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &importData); err != nil {
		f.LastAction = fmt.Sprintf("Import error: %v", err)
		f.Commit()
		return
	}
	
	// Save state BEFORE making changes
	f.saveToUndoStack()
	
	// Clear current diagram
	f.Canvas.Clear()
	f.NodeCount = 0
	f.EdgeCount = 0
	
	// Import boxes
	if boxes, ok := importData["boxes"].([]interface{}); ok {
		for _, boxData := range boxes {
			if box, ok := boxData.(map[string]interface{}); ok {
				id := box["id"].(string)
				label := box["label"].(string)
				boxType := components.BoxType(box["type"].(string))
				x := int(box["x"].(float64))
				y := int(box["y"].(float64))
				
				newBox := components.NewFlowBox(id, label, boxType, x, y)
				
				// Import code and other metadata
				if code, ok := box["code"].(string); ok {
					if newBox.Data == nil {
						newBox.Data = make(map[string]interface{})
					}
					newBox.Data["code"] = code
				}
				
				// Import any other metadata
				for key, value := range box {
					if key != "id" && key != "label" && key != "type" && key != "x" && key != "y" && key != "code" {
						if newBox.Data == nil {
							newBox.Data = make(map[string]interface{})
						}
						newBox.Data[key] = value
					}
				}
				
				// Register the driver properly for imported boxes
				liveview.New(id, newBox)
				f.Canvas.AddBox(newBox)
				f.NodeCount++
			}
		}
	}
	
	// Import edges
	if edges, ok := importData["edges"].([]interface{}); ok {
		for _, edgeData := range edges {
			if edge, ok := edgeData.(map[string]interface{}); ok {
				id := edge["id"].(string)
				fromBox := edge["fromBox"].(string)
				fromPort := edge["fromPort"].(string)
				toBox := edge["toBox"].(string)
				toPort := edge["toPort"].(string)
				
				newEdge := components.NewFlowEdge(id, fromBox, fromPort, toBox, toPort)
				if edgeType, ok := edge["type"].(string); ok {
					newEdge.Type = components.EdgeType(edgeType)
				}
				if label, ok := edge["label"].(string); ok {
					newEdge.Label = label
				}
				
				f.Canvas.AddEdge(newEdge)
				f.EdgeCount++
			}
		}
	}
	
	// Update edge positions
	f.updateEdgePositions()
	
	f.LastAction = fmt.Sprintf("Imported %d boxes and %d edges", f.NodeCount, f.EdgeCount)
	f.Commit()
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
				// Save state BEFORE creating connection
				f.saveToUndoStack()
				f.createConnection(f.ConnectingFrom, boxID)
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

func (f *EnhancedFlowTool) DeleteBox(data interface{}) {
	var boxID string
	
	// Handle different data types
	if str, ok := data.(string); ok {
		boxID = str
	} else if dataMap, ok := data.(map[string]interface{}); ok {
		if id, ok := dataMap["id"].(string); ok {
			boxID = id
		}
	}
	
	if boxID == "" {
		f.LastAction = "No box selected to delete"
		f.Commit()
		return
	}
	
	// Save state for undo
	f.saveToUndoStack()
	
	// Remove all edges connected to this box
	for edgeID, edge := range f.Canvas.Edges {
		if edge.FromBox == boxID || edge.ToBox == boxID {
			delete(f.Canvas.Edges, edgeID)
			f.EdgeCount--
		}
	}
	
	// Remove the box
	if _, exists := f.Canvas.Boxes[boxID]; exists {
		delete(f.Canvas.Boxes, boxID)
		f.NodeCount--
		f.LastAction = fmt.Sprintf("Deleted box: %s", boxID)
	} else {
		f.LastAction = fmt.Sprintf("Box not found: %s", boxID)
	}
	
	if f.ComponentDriver != nil {
		f.Commit()
	}
}

func (f *EnhancedFlowTool) DeleteEdge(data interface{}) {
	var edgeID string
	
	// Handle different data types
	if str, ok := data.(string); ok {
		edgeID = str
	} else if dataMap, ok := data.(map[string]interface{}); ok {
		if id, ok := dataMap["id"].(string); ok {
			edgeID = id
		}
	}
	
	if edgeID == "" {
		f.LastAction = "No edge selected to delete"
		f.Commit()
		return
	}
	
	// Save state for undo
	f.saveToUndoStack()
	
	// Remove the edge
	if _, exists := f.Canvas.Edges[edgeID]; exists {
		delete(f.Canvas.Edges, edgeID)
		f.EdgeCount--
		f.LastAction = fmt.Sprintf("Deleted edge: %s", edgeID)
	} else {
		f.LastAction = fmt.Sprintf("Edge not found: %s", edgeID)
	}
	
	if f.ComponentDriver != nil {
		f.Commit()
	}
}

func (f *EnhancedFlowTool) SelectEdge(data interface{}) {
	edgeID, ok := data.(string)
	if !ok {
		f.LastAction = "Invalid edge ID for selection"
		f.Commit()
		return
	}
	
	edge, exists := f.Canvas.Edges[edgeID]
	if !exists {
		f.LastAction = fmt.Sprintf("Edge not found: %s", edgeID)
		f.Commit()
		return
	}
	
	// Deselect other edges
	for _, e := range f.Canvas.Edges {
		e.Selected = false
	}
	
	// Select this edge
	edge.Selected = !edge.Selected
	f.LastAction = fmt.Sprintf("Selected edge: %s", edgeID)
	
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
	log.Printf("🔗 HandleToggleConnectMode called with data: %v", data)
	f.ConnectingMode = !f.ConnectingMode
	f.ConnectingFrom = ""
	log.Printf("🔗 ConnectingMode toggled to: %v", f.ConnectingMode)
	
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
		log.Printf("🔄 Calling Commit() to update UI...")
		f.Commit()
		log.Printf("✅ Commit() completed")
	}
}

// Backward compatibility methods for old BoxDrag events
func (f *EnhancedFlowTool) HandleBoxStartDrag(data interface{}) {
	log.Printf("🚀 HandleBoxStartDrag called with data: %v (%T)", data, data)
	
	// Convert old format to new format and delegate to HandleDragStart
	if dataStr, ok := data.(string); ok {
		var oldData map[string]interface{}
		if err := json.Unmarshal([]byte(dataStr), &oldData); err == nil {
			// Convert old format {id, x, y} to new format {element, x, y}
			if id, hasId := oldData["id"].(string); hasId {
				newData := map[string]interface{}{
					"element": "box-" + id,
					"x":       oldData["x"],
					"y":       oldData["y"],
				}
				newDataJSON, _ := json.Marshal(newData)
				f.HandleDragStart(string(newDataJSON))
				return
			}
		}
	}
	
	// Fallback to direct call
	f.HandleDragStart(data)
}

func (f *EnhancedFlowTool) HandleBoxDrag(data interface{}) {
	// Convert old format to new format and delegate to HandleDragMove
	if dataStr, ok := data.(string); ok {
		var oldData map[string]interface{}
		if err := json.Unmarshal([]byte(dataStr), &oldData); err == nil {
			// Convert old format {id, x, y} to new format {element, x, y}
			if id, hasId := oldData["id"].(string); hasId {
				newData := map[string]interface{}{
					"element": "box-" + id,
					"x":       oldData["x"],
					"y":       oldData["y"],
				}
				newDataJSON, _ := json.Marshal(newData)
				f.HandleDragMove(string(newDataJSON))
				return
			}
		}
	}
	
	// Fallback to direct call
	f.HandleDragMove(data)
}

func (f *EnhancedFlowTool) HandleBoxEndDrag(data interface{}) {
	// Convert old format and delegate to HandleDragEnd
	if dataStr, ok := data.(string); ok {
		// For BoxEndDrag, the data might just be the box ID
		newData := map[string]interface{}{
			"element": "box-" + dataStr,
			"x":       0, // Will be updated by the actual handler
			"y":       0,
		}
		newDataJSON, _ := json.Marshal(newData)
		f.HandleDragEnd(string(newDataJSON))
		return
	}
	
	// Fallback to direct call
	f.HandleDragEnd(data)
}

func (f *EnhancedFlowTool) HandleDragStart(data interface{}) {
	// Handle generic drag start
	log.Printf("🚀 DragStart called with data: %v (%T)", data, data)
	
	// Save state BEFORE starting drag
	f.saveToUndoStack()
	
	// Try to parse as JSON string first
	if dataStr, ok := data.(string); ok {
		var dataMap map[string]interface{}
		if err := json.Unmarshal([]byte(dataStr), &dataMap); err == nil {
			data = dataMap
			log.Printf("Parsed JSON data: %v", dataMap)
		}
	}
	
	if dataMap, ok := data.(map[string]interface{}); ok {
		if element, ok := dataMap["element"].(string); ok {
			// Extract box ID from element ID (format: "box-{id}")
			if strings.HasPrefix(element, "box-") {
				boxID := strings.TrimPrefix(element, "box-")
				f.DraggingBox = boxID
				log.Printf("Started dragging box: %s", boxID)
				if box, exists := f.Canvas.Boxes[boxID]; exists {
					box.Dragging = true
					if box.ComponentDriver != nil {
						box.Commit()
					}
				}
				f.LastAction = fmt.Sprintf("Started dragging %s", boxID)
			}
		}
	}
	
	if f.ComponentDriver != nil {
		f.Commit()
	}
}

func (f *EnhancedFlowTool) HandleDragMove(data interface{}) {
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
			
			// Emit drag event for auto-save
			f.EventRegistry.Emit("diagram.change", map[string]interface{}{
				"type": "box_moved",
				"box":  f.DraggingBox,
			})
		}
	}
	
	// Call Commit to update the edges
	if f.ComponentDriver != nil {
		f.Commit()
	}
}

func (f *EnhancedFlowTool) HandleDragEnd(data interface{}) {
	if f.DraggingBox != "" {
		if box, exists := f.Canvas.Boxes[f.DraggingBox]; exists {
			box.Dragging = false
			if box.ComponentDriver != nil {
				box.Commit()
			}
		}
		f.LastAction = fmt.Sprintf("Finished dragging %s", f.DraggingBox)
		// Don't save state here - it was already saved in HandleDragStart
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

// Edit handlers
func (f *EnhancedFlowTool) HandleEditBox(data interface{}) {
	boxID := ""
	if str, ok := data.(string); ok {
		boxID = str
	}
	
	if box, ok := f.Canvas.Boxes[boxID]; ok {
		f.EditingMode = true
		f.EditingType = "box"
		f.EditingID = boxID
		f.EditingValue = box.Label
		
		// Load existing code from box Data
		if box.Data == nil {
			box.Data = make(map[string]interface{})
		}
		if code, ok := box.Data["code"].(string); ok {
			f.EditingCode = code
		} else {
			f.EditingCode = ""
		}
		
		f.LastAction = fmt.Sprintf("Editing box: %s", box.Label)
		f.Commit()
	}
}

func (f *EnhancedFlowTool) HandleEditEdge(data interface{}) {
	edgeID := ""
	if str, ok := data.(string); ok {
		edgeID = str
	}
	
	if edge, ok := f.Canvas.Edges[edgeID]; ok {
		f.EditingMode = true
		f.EditingType = "edge"
		f.EditingID = edgeID
		f.EditingValue = edge.Label
		f.LastAction = fmt.Sprintf("Editing edge label")
		f.Commit()
	}
}

func (f *EnhancedFlowTool) HandleSaveEdit(data interface{}) {
	// Parse the data which could be a string or JSON object
	var editData map[string]interface{}
	
	if str, ok := data.(string); ok {
		// Try to parse as JSON first
		if err := json.Unmarshal([]byte(str), &editData); err != nil {
			// If not JSON, treat as simple string (for backward compatibility)
			editData = map[string]interface{}{"value": str}
		}
	} else if m, ok := data.(map[string]interface{}); ok {
		editData = m
	}
	
	newValue := ""
	if val, ok := editData["value"].(string); ok {
		newValue = val
	}
	
	if f.EditingType == "box" {
		if box, ok := f.Canvas.Boxes[f.EditingID]; ok {
			// Save state BEFORE making changes
			f.saveToUndoStack()
			
			box.Label = newValue
			
			// Save code to box Data
			if box.Data == nil {
				box.Data = make(map[string]interface{})
			}
			if code, ok := editData["code"].(string); ok {
				box.Data["code"] = code
				f.LastAction = fmt.Sprintf("Updated box '%s' with code", newValue)
			} else {
				f.LastAction = fmt.Sprintf("Renamed box to: %s", newValue)
			}
		}
	} else if f.EditingType == "edge" {
		if edge, ok := f.Canvas.Edges[f.EditingID]; ok {
			// Save state BEFORE making changes
			f.saveToUndoStack()
			edge.Label = newValue
			f.LastAction = fmt.Sprintf("Updated edge label to: %s", newValue)
		}
	}
	
	f.EditingMode = false
	f.EditingID = ""
	f.EditingValue = ""
	f.EditingCode = ""
	f.Commit()
}

func (f *EnhancedFlowTool) HandleCancelEdit(data interface{}) {
	f.EditingMode = false
	f.EditingID = ""
	f.EditingValue = ""
	f.EditingCode = ""
	f.LastAction = "Edit cancelled"
	f.Commit()
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
		// Use simple layout that doesn't cause remounting
		document := liveview.NewLayout("flowtool-layout", `{{mount "flow-tool"}}`)
		
		// Create enhanced flow tool component
		flowTool := NewEnhancedFlowTool()
		liveview.New("flow-tool", flowTool)
		
		// Set up the modal driver
		if flowTool.Modal != nil {
			liveview.New("export-modal", flowTool.Modal)
		}
		
		// Set up the file upload driver
		if flowTool.FileUpload != nil {
			liveview.New("file-upload", flowTool.FileUpload)
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
		return c.HTML(http.StatusOK, `<!DOCTYPE html>
<html>
<head>
	<title>Enhanced Flow Diagram Tool</title>
	<style>
		body {
			font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, sans-serif;
			max-width: 800px;
			margin: 50px auto;
			padding: 20px;
			background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
			min-height: 100vh;
		}
		.container {
			background: white;
			border-radius: 12px;
			padding: 2rem;
			box-shadow: 0 10px 30px rgba(0,0,0,0.1);
		}
		h1 {
			color: #1a202c;
			margin-bottom: 1rem;
		}
		h2 {
			color: #4a5568;
			margin-top: 2rem;
			margin-bottom: 1rem;
		}
		a {
			color: #667eea;
			text-decoration: none;
			font-weight: 500;
		}
		a:hover {
			text-decoration: underline;
		}
		ul {
			list-style: none;
			padding: 0;
		}
		li {
			padding: 0.5rem 0;
			padding-left: 1.5rem;
			position: relative;
		}
		li:before {
			content: "✓";
			position: absolute;
			left: 0;
			color: #48bb78;
			font-weight: bold;
		}
		.button {
			display: inline-block;
			padding: 0.75rem 1.5rem;
			background: #667eea;
			color: white;
			border-radius: 6px;
			margin-top: 1rem;
			transition: all 0.2s;
		}
		.button:hover {
			background: #5a67d8;
			transform: translateY(-2px);
			box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
			text-decoration: none;
		}
	</style>
</head>
<body>
	<div class="container">
		<h1>Enhanced Flow Diagram Tool</h1>
		<p>An advanced visual flow diagram editor with drag-and-drop functionality and real-time collaboration features.</p>
		<a href="/example/flowtool" class="button">Open Flow Diagram Editor</a>
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
	</div>
</body>
</html>`)
	})

	port := ":8082"
	log.Printf("🚀 Enhanced Flow Tool Server")
	log.Printf("🌐 Starting on http://localhost%s", port)
	log.Printf("Visit http://localhost%s/example/flowtool", port)
	e.Logger.Fatal(e.Start(port))
}