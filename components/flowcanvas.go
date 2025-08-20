package components

import (
	"fmt"

	"github.com/arturoeanton/go-echo-live-view/liveview"
)

type GridSize int

const (
	GridNone   GridSize = 0
	GridSmall  GridSize = 10
	GridMedium GridSize = 20
	GridLarge  GridSize = 40
)

type FlowCanvas struct {
	*liveview.ComponentDriver[*FlowCanvas]

	ID             string
	Width          int
	Height         int
	Boxes          map[string]*FlowBox
	Edges          map[string]*FlowEdge
	GridSize       GridSize
	ShowGrid       bool
	Zoom           float64
	PanX           int
	PanY           int
	SelectedBox    string
	SelectedEdge   string
	ConnectingFrom string
	ConnectingPort string
	IsConnecting   bool
	ReadOnly       bool
	OnBoxClick     func(boxID string)
	OnEdgeClick    func(edgeID string)
	OnConnection   func(fromBox, fromPort, toBox, toPort string)
	OnBoxMove      func(boxID string, x, y int)
	OnCanvasClick  func(x, y int)
}

func NewFlowCanvas(id string, width, height int) *FlowCanvas {
	return &FlowCanvas{
		ID:       id,
		Width:    width,
		Height:   height,
		Boxes:    make(map[string]*FlowBox),
		Edges:    make(map[string]*FlowEdge),
		GridSize: GridMedium,
		ShowGrid: true,
		Zoom:     1.0,
		PanX:     0,
		PanY:     0,
	}
}

func (c *FlowCanvas) Start() {
	// Events are registered directly on the ComponentDriver
	if c.ComponentDriver != nil {
		c.ComponentDriver.Events["CanvasClick"] = func(comp *FlowCanvas, data interface{}) {
			comp.HandleCanvasClick(data)
		}
		c.ComponentDriver.Events["AddBox"] = func(comp *FlowCanvas, data interface{}) {
			comp.HandleAddBox(data)
		}
		c.ComponentDriver.Events["RemoveBox"] = func(comp *FlowCanvas, data interface{}) {
			comp.HandleRemoveBox(data)
		}
		c.ComponentDriver.Events["AddEdge"] = func(comp *FlowCanvas, data interface{}) {
			comp.HandleAddEdge(data)
		}
		c.ComponentDriver.Events["RemoveEdge"] = func(comp *FlowCanvas, data interface{}) {
			comp.HandleRemoveEdge(data)
		}
		c.ComponentDriver.Events["SelectBox"] = func(comp *FlowCanvas, data interface{}) {
			comp.HandleSelectBox(data)
		}
		c.ComponentDriver.Events["SelectEdge"] = func(comp *FlowCanvas, data interface{}) {
			comp.HandleSelectEdge(data)
		}
		c.ComponentDriver.Events["StartConnection"] = func(comp *FlowCanvas, data interface{}) {
			comp.HandleStartConnection(data)
		}
		c.ComponentDriver.Events["CompleteConnection"] = func(comp *FlowCanvas, data interface{}) {
			comp.HandleCompleteConnection(data)
		}
		c.ComponentDriver.Events["CancelConnection"] = func(comp *FlowCanvas, data interface{}) {
			comp.HandleCancelConnection(data)
		}
		c.ComponentDriver.Events["MoveBox"] = func(comp *FlowCanvas, data interface{}) {
			comp.HandleMoveBox(data)
		}
		c.ComponentDriver.Events["ZoomIn"] = func(comp *FlowCanvas, data interface{}) {
			comp.HandleZoomIn(data)
		}
		c.ComponentDriver.Events["ZoomOut"] = func(comp *FlowCanvas, data interface{}) {
			comp.HandleZoomOut(data)
		}
		c.ComponentDriver.Events["ResetView"] = func(comp *FlowCanvas, data interface{}) {
			comp.HandleResetView(data)
		}
		c.ComponentDriver.Events["ToggleGrid"] = func(comp *FlowCanvas, data interface{}) {
			comp.HandleToggleGrid(data)
		}
		c.ComponentDriver.Events["Clear"] = func(comp *FlowCanvas, data interface{}) {
			comp.HandleClear(data)
		}
	}
}

func (c *FlowCanvas) GetTemplate() string {
	return `
<div class="flow-canvas-container" id="{{.ID}}">
	<style>
		.flow-canvas-container {
			position: relative;
			width: {{.Width}}px;
			height: {{.Height}}px;
			border: 2px solid #e5e7eb;
			border-radius: 8px;
			overflow: hidden;
			background: #fafafa;
		}
		
		.flow-canvas {
			position: relative;
			width: 100%;
			height: 100%;
			transform: scale({{.Zoom}}) translate({{.PanX}}px, {{.PanY}}px);
			transform-origin: 0 0;
			transition: transform 0.2s;
		}
		
		{{if .ShowGrid}}
		.flow-canvas {
			{{if ne .GridSize 0}}
			background-image: 
				linear-gradient(#e5e7eb 1px, transparent 1px),
				linear-gradient(90deg, #e5e7eb 1px, transparent 1px);
			background-size: {{.GridSize}}px {{.GridSize}}px;
			{{end}}
		}
		{{end}}
		
		.canvas-toolbar {
			position: absolute;
			top: 10px;
			right: 10px;
			display: flex;
			gap: 0.5rem;
			background: white;
			padding: 0.5rem;
			border-radius: 6px;
			box-shadow: 0 2px 8px rgba(0,0,0,0.1);
			z-index: 100;
		}
		
		.toolbar-btn {
			padding: 0.5rem;
			background: white;
			border: 1px solid #d1d5db;
			border-radius: 4px;
			cursor: pointer;
			font-size: 0.875rem;
			transition: all 0.2s;
			display: flex;
			align-items: center;
			gap: 0.25rem;
		}
		
		.toolbar-btn:hover {
			background: #f3f4f6;
			border-color: #9ca3af;
		}
		
		.toolbar-btn.active {
			background: #3b82f6;
			color: white;
			border-color: #3b82f6;
		}
		
		.canvas-status {
			position: absolute;
			bottom: 10px;
			left: 10px;
			background: white;
			padding: 0.5rem 1rem;
			border-radius: 6px;
			box-shadow: 0 2px 8px rgba(0,0,0,0.1);
			font-size: 0.75rem;
			color: #6b7280;
			z-index: 100;
		}
		
		.connection-line {
			position: absolute;
			pointer-events: none;
			z-index: 1000;
		}
		
		.add-box-menu {
			position: absolute;
			background: white;
			border: 1px solid #d1d5db;
			border-radius: 6px;
			padding: 0.5rem;
			box-shadow: 0 4px 12px rgba(0,0,0,0.15);
			z-index: 200;
			display: none;
		}
		
		.add-box-menu.show {
			display: block;
		}
		
		.menu-item {
			padding: 0.5rem 1rem;
			cursor: pointer;
			border-radius: 4px;
			font-size: 0.875rem;
			transition: background 0.2s;
		}
		
		.menu-item:hover {
			background: #f3f4f6;
		}
		
		.minimap {
			position: absolute;
			bottom: 10px;
			right: 10px;
			width: 150px;
			height: 100px;
			background: white;
			border: 1px solid #d1d5db;
			border-radius: 4px;
			z-index: 90;
			overflow: hidden;
		}
		
		.minimap-viewport {
			position: absolute;
			background: rgba(59, 130, 246, 0.2);
			border: 1px solid #3b82f6;
		}
	</style>
	
	<div class="canvas-toolbar">
		<button class="toolbar-btn" 
		        onclick="send_event('{{.ID}}', 'ZoomIn', null)"
		        title="Zoom In">
			<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
				<circle cx="11" cy="11" r="8"/>
				<path d="m21 21-4.35-4.35"/>
				<line x1="11" y1="8" x2="11" y2="14"/>
				<line x1="8" y1="11" x2="14" y2="11"/>
			</svg>
		</button>
		
		<button class="toolbar-btn" 
		        onclick="send_event('{{.ID}}', 'ZoomOut', null)"
		        title="Zoom Out">
			<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
				<circle cx="11" cy="11" r="8"/>
				<path d="m21 21-4.35-4.35"/>
				<line x1="8" y1="11" x2="14" y2="11"/>
			</svg>
		</button>
		
		<button class="toolbar-btn" 
		        onclick="send_event('{{.ID}}', 'ResetView', null)"
		        title="Reset View">
			<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
				<path d="M3 12a9 9 0 1 0 9-9 9.75 9.75 0 0 0-6.74 2.74L3 8"/>
				<path d="M3 3v5h5"/>
			</svg>
		</button>
		
		<button class="toolbar-btn {{if .ShowGrid}}active{{end}}" 
		        onclick="send_event('{{.ID}}', 'ToggleGrid', null)"
		        title="Toggle Grid">
			<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
				<rect x="3" y="3" width="7" height="7"/>
				<rect x="14" y="3" width="7" height="7"/>
				<rect x="14" y="14" width="7" height="7"/>
				<rect x="3" y="14" width="7" height="7"/>
			</svg>
		</button>
		
		{{if not .ReadOnly}}
		<button class="toolbar-btn" 
		        onclick="send_event('{{.ID}}', 'Clear', null)"
		        title="Clear Canvas">
			<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
				<polyline points="3 6 5 6 21 6"/>
				<path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/>
			</svg>
		</button>
		{{end}}
	</div>
	
	<div class="flow-canvas" 
	     onclick="send_event('{{.ID}}', 'CanvasClick', {x: event.offsetX, y: event.offsetY})">
		
		<!-- Render all edges first (below boxes) -->
		{{range .Edges}}
			{{.GetTemplate}}
		{{end}}
		
		<!-- Render all boxes -->
		{{range .Boxes}}
			{{.GetTemplate}}
		{{end}}
		
		<!-- Connection line while connecting -->
		{{if .IsConnecting}}
		<svg class="connection-line">
			<line x1="0" y1="0" x2="100" y2="100" 
			      stroke="#3b82f6" 
			      stroke-width="2" 
			      stroke-dasharray="5,5"/>
		</svg>
		{{end}}
	</div>
	
	<!-- Add box menu -->
	<div class="add-box-menu" id="add-menu-{{.ID}}">
		<div class="menu-item" onclick="send_event('{{.ID}}', 'AddBox', {type: 'start', x: 0, y: 0})">
			Add Start Node
		</div>
		<div class="menu-item" onclick="send_event('{{.ID}}', 'AddBox', {type: 'process', x: 0, y: 0})">
			Add Process Node
		</div>
		<div class="menu-item" onclick="send_event('{{.ID}}', 'AddBox', {type: 'decision', x: 0, y: 0})">
			Add Decision Node
		</div>
		<div class="menu-item" onclick="send_event('{{.ID}}', 'AddBox', {type: 'data', x: 0, y: 0})">
			Add Data Node
		</div>
		<div class="menu-item" onclick="send_event('{{.ID}}', 'AddBox', {type: 'end', x: 0, y: 0})">
			Add End Node
		</div>
	</div>
	
	<div class="canvas-status">
		Boxes: {{len .Boxes}} | Edges: {{len .Edges}} | Zoom: {{.ZoomPercent}}%
		{{if .IsConnecting}} | Connecting from {{.ConnectingFrom}}{{end}}
	</div>
</div>
`
}

func (c *FlowCanvas) GetDriver() liveview.LiveDriver {
	return c
}

func (c *FlowCanvas) HandleCanvasClick(data interface{}) {
	clickData := data.(map[string]interface{})
	x := int(clickData["x"].(float64))
	y := int(clickData["y"].(float64))

	// Adjust for zoom and pan
	x = int(float64(x)/c.Zoom) - c.PanX
	y = int(float64(y)/c.Zoom) - c.PanY

	// Snap to grid if enabled
	if c.ShowGrid && c.GridSize > 0 {
		gridSize := int(c.GridSize)
		x = (x / gridSize) * gridSize
		y = (y / gridSize) * gridSize
	}

	if c.OnCanvasClick != nil {
		c.OnCanvasClick(x, y)
	}

	// Clear selections
	c.SelectedBox = ""
	c.SelectedEdge = ""
	for _, box := range c.Boxes {
		box.Selected = false
	}
	for _, edge := range c.Edges {
		edge.Selected = false
	}

	if c.ComponentDriver != nil {
		c.Commit()
	}
}

func (c *FlowCanvas) HandleAddBox(data interface{}) {
	addData := data.(map[string]interface{})
	boxType := BoxType(addData["type"].(string))
	x := int(addData["x"].(float64))
	y := int(addData["y"].(float64))

	boxID := fmt.Sprintf("box_%d", len(c.Boxes)+1)
	label := fmt.Sprintf("%s Node", boxType)

	box := NewFlowBox(boxID, label, boxType, x, y)
	box.OnClick = func(id string) {
		c.HandleSelectBox(id)
	}

	c.Boxes[boxID] = box
	if c.ComponentDriver != nil {
		c.Commit()
	}
}

func (c *FlowCanvas) HandleRemoveBox(data interface{}) {
	boxID := data.(string)

	// Remove associated edges
	for edgeID, edge := range c.Edges {
		if edge.FromBox == boxID || edge.ToBox == boxID {
			delete(c.Edges, edgeID)
		}
	}

	delete(c.Boxes, boxID)
	if c.ComponentDriver != nil {
		c.Commit()
	}
}

func (c *FlowCanvas) HandleAddEdge(data interface{}) {
	edgeData := data.(map[string]interface{})
	fromBox := edgeData["fromBox"].(string)
	fromPort := edgeData["fromPort"].(string)
	toBox := edgeData["toBox"].(string)
	toPort := edgeData["toPort"].(string)

	edgeID := fmt.Sprintf("edge_%s_%s", fromBox, toBox)

	edge := NewFlowEdge(edgeID, fromBox, fromPort, toBox, toPort)
	edge.OnClick = func(id string) {
		c.HandleSelectEdge(id)
	}

	// Calculate positions
	if from, ok := c.Boxes[fromBox]; ok {
		if to, ok := c.Boxes[toBox]; ok {
			edge.FromX = from.X + from.Width
			edge.FromY = from.Y + from.Height/2
			edge.ToX = to.X
			edge.ToY = to.Y + to.Height/2
		}
	}

	c.Edges[edgeID] = edge
	if c.ComponentDriver != nil {
		c.Commit()
	}
}

func (c *FlowCanvas) HandleRemoveEdge(data interface{}) {
	edgeID := data.(string)
	delete(c.Edges, edgeID)
	if c.ComponentDriver != nil {
		c.Commit()
	}
}

func (c *FlowCanvas) HandleSelectBox(data interface{}) {
	boxID := data.(string)

	// Clear previous selections
	for id, box := range c.Boxes {
		box.Selected = (id == boxID)
	}
	for _, edge := range c.Edges {
		edge.Selected = false
	}

	c.SelectedBox = boxID
	c.SelectedEdge = ""

	if c.OnBoxClick != nil {
		c.OnBoxClick(boxID)
	}

	if c.ComponentDriver != nil {
		c.Commit()
	}
}

func (c *FlowCanvas) HandleSelectEdge(data interface{}) {
	edgeID := data.(string)

	// Clear previous selections
	for _, box := range c.Boxes {
		box.Selected = false
	}
	for id, edge := range c.Edges {
		edge.Selected = (id == edgeID)
	}

	c.SelectedBox = ""
	c.SelectedEdge = edgeID

	if c.OnEdgeClick != nil {
		c.OnEdgeClick(edgeID)
	}

	if c.ComponentDriver != nil {
		c.Commit()
	}
}

func (c *FlowCanvas) HandleStartConnection(data interface{}) {
	connData := data.(map[string]interface{})
	boxID := connData["boxId"].(string)
	portID := connData["portId"].(string)

	c.IsConnecting = true
	c.ConnectingFrom = boxID
	c.ConnectingPort = portID

	if c.ComponentDriver != nil {
		c.Commit()
	}
}

func (c *FlowCanvas) HandleCompleteConnection(data interface{}) {
	connData := data.(map[string]interface{})
	toBox := connData["boxId"].(string)
	toPort := connData["portId"].(string)

	if c.IsConnecting && c.ConnectingFrom != "" && c.ConnectingFrom != toBox {
		// Create edge
		c.HandleAddEdge(map[string]interface{}{
			"fromBox":  c.ConnectingFrom,
			"fromPort": c.ConnectingPort,
			"toBox":    toBox,
			"toPort":   toPort,
		})

		if c.OnConnection != nil {
			c.OnConnection(c.ConnectingFrom, c.ConnectingPort, toBox, toPort)
		}
	}

	c.IsConnecting = false
	c.ConnectingFrom = ""
	c.ConnectingPort = ""

	if c.ComponentDriver != nil {
		c.Commit()
	}
}

func (c *FlowCanvas) HandleCancelConnection(data interface{}) {
	c.IsConnecting = false
	c.ConnectingFrom = ""
	c.ConnectingPort = ""
	if c.ComponentDriver != nil {
		c.Commit()
	}
}

func (c *FlowCanvas) HandleMoveBox(data interface{}) {
	moveData := data.(map[string]interface{})
	boxID := moveData["boxId"].(string)
	x := int(moveData["x"].(float64))
	y := int(moveData["y"].(float64))

	if box, ok := c.Boxes[boxID]; ok {
		// Snap to grid
		if c.ShowGrid && c.GridSize > 0 {
			gridSize := int(c.GridSize)
			x = (x / gridSize) * gridSize
			y = (y / gridSize) * gridSize
		}

		box.MoveTo(x, y)

		// Update connected edges
		c.updateEdgePositions()

		if c.OnBoxMove != nil {
			c.OnBoxMove(boxID, x, y)
		}
	}

	if c.ComponentDriver != nil {
		c.Commit()
	}
}

func (c *FlowCanvas) HandleZoomIn(data interface{}) {
	c.Zoom = min(c.Zoom*1.2, 3.0)
	if c.ComponentDriver != nil {
		c.Commit()
	}
}

func (c *FlowCanvas) HandleZoomOut(data interface{}) {
	c.Zoom = max(c.Zoom/1.2, 0.3)
	if c.ComponentDriver != nil {
		c.Commit()
	}
}

func (c *FlowCanvas) HandleResetView(data interface{}) {
	c.Zoom = 1.0
	c.PanX = 0
	c.PanY = 0
	if c.ComponentDriver != nil {
		c.Commit()
	}
}

func (c *FlowCanvas) HandleToggleGrid(data interface{}) {
	c.ShowGrid = !c.ShowGrid
	if c.ComponentDriver != nil {
		c.Commit()
	}
}

func (c *FlowCanvas) HandleClear(data interface{}) {
	c.Boxes = make(map[string]*FlowBox)
	c.Edges = make(map[string]*FlowEdge)
	c.SelectedBox = ""
	c.SelectedEdge = ""
	if c.ComponentDriver != nil {
		c.Commit()
	}
}

func (c *FlowCanvas) updateEdgePositions() {
	for _, edge := range c.Edges {
		if fromBox, ok := c.Boxes[edge.FromBox]; ok {
			if toBox, ok := c.Boxes[edge.ToBox]; ok {
				// Simple positioning - can be improved based on port positions
				edge.FromX = fromBox.X + fromBox.Width
				edge.FromY = fromBox.Y + fromBox.Height/2
				edge.ToX = toBox.X
				edge.ToY = toBox.Y + toBox.Height/2
			}
		}
	}
}

func (c *FlowCanvas) ZoomPercent() int {
	return int(c.Zoom * 100)
}

func (c *FlowCanvas) AddBox(box *FlowBox) {
	c.Boxes[box.ID] = box
	if c.ComponentDriver != nil {
		c.Commit()
	}
}

func (c *FlowCanvas) AddEdge(edge *FlowEdge) {
	c.Edges[edge.ID] = edge
	c.updateEdgePositions()
	if c.ComponentDriver != nil {
		c.Commit()
	}
}

func (c *FlowCanvas) RemoveBox(boxID string) {
	c.HandleRemoveBox(boxID)
}

func (c *FlowCanvas) RemoveEdge(edgeID string) {
	c.HandleRemoveEdge(edgeID)
}

func (c *FlowCanvas) Clear() {
	c.HandleClear(nil)
}

func (c *FlowCanvas) ExportJSON() map[string]interface{} {
	boxes := []map[string]interface{}{}
	for _, box := range c.Boxes {
		boxes = append(boxes, map[string]interface{}{
			"id":    box.ID,
			"label": box.Label,
			"type":  box.Type,
			"x":     box.X,
			"y":     box.Y,
		})
	}

	edges := []map[string]interface{}{}
	for _, edge := range c.Edges {
		edges = append(edges, map[string]interface{}{
			"id":       edge.ID,
			"fromBox":  edge.FromBox,
			"fromPort": edge.FromPort,
			"toBox":    edge.ToBox,
			"toPort":   edge.ToPort,
			"type":     edge.Type,
		})
	}

	return map[string]interface{}{
		"boxes": boxes,
		"edges": edges,
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
