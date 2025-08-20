package main

import (
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

type FlowToolExample struct {
	*liveview.ComponentDriver[*FlowToolExample]
	
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
}

func NewFlowToolExample() *FlowToolExample {
	// Create canvas without driver for now
	canvas := components.NewFlowCanvas("main-canvas", 1200, 600)
	
	// Create modal for JSON export
	modal := &components.Modal{
		Title:      "Export JSON",
		Size:       "large",
		Closable:   true,
		ShowFooter: false,
		IsOpen:     false,
	}
	
	// Create initial flow diagram - just add to maps directly
	startBox := components.NewFlowBox("start", "Start", components.BoxTypeStart, 50, 250)
	processBox1 := components.NewFlowBox("process1", "Initialize", components.BoxTypeProcess, 200, 250)
	decisionBox := components.NewFlowBox("decision1", "Check Condition", components.BoxTypeDecision, 400, 250)
	processBox2 := components.NewFlowBox("process2", "Process A", components.BoxTypeProcess, 600, 150)
	processBox3 := components.NewFlowBox("process3", "Process B", components.BoxTypeProcess, 600, 350)
	dataBox := components.NewFlowBox("data1", "Store Result", components.BoxTypeData, 800, 250)
	endBox := components.NewFlowBox("end", "End", components.BoxTypeEnd, 1000, 250)
	
	// Add boxes directly to map
	canvas.Boxes[startBox.ID] = startBox
	canvas.Boxes[processBox1.ID] = processBox1
	canvas.Boxes[decisionBox.ID] = decisionBox
	canvas.Boxes[processBox2.ID] = processBox2
	canvas.Boxes[processBox3.ID] = processBox3
	canvas.Boxes[dataBox.ID] = dataBox
	canvas.Boxes[endBox.ID] = endBox
	
	// Create edges
	edge1 := components.NewFlowEdge("edge1", "start", "out1", "process1", "in1")
	edge1.Label = "Begin"
	edge1.UpdatePosition(startBox.X+startBox.Width, startBox.Y+startBox.Height/2,
		processBox1.X, processBox1.Y+processBox1.Height/2)
	
	edge2 := components.NewFlowEdge("edge2", "process1", "out1", "decision1", "in1")
	edge2.UpdatePosition(processBox1.X+processBox1.Width, processBox1.Y+processBox1.Height/2,
		decisionBox.X, decisionBox.Y+decisionBox.Height/2)
	
	edge3 := components.NewFlowEdge("edge3", "decision1", "out1", "process2", "in1")
	edge3.Label = "True"
	edge3.Type = components.EdgeTypeCurved
	edge3.UpdatePosition(decisionBox.X+decisionBox.Width, decisionBox.Y,
		processBox2.X, processBox2.Y+processBox2.Height/2)
	
	edge4 := components.NewFlowEdge("edge4", "decision1", "out2", "process3", "in1")
	edge4.Label = "False"
	edge4.Type = components.EdgeTypeCurved
	edge4.UpdatePosition(decisionBox.X+decisionBox.Width, decisionBox.Y+decisionBox.Height,
		processBox3.X, processBox3.Y+processBox3.Height/2)
	
	edge5 := components.NewFlowEdge("edge5", "process2", "out1", "data1", "in1")
	edge5.UpdatePosition(processBox2.X+processBox2.Width, processBox2.Y+processBox2.Height/2,
		dataBox.X, dataBox.Y)
	
	edge6 := components.NewFlowEdge("edge6", "process3", "out1", "data1", "in1")
	edge6.UpdatePosition(processBox3.X+processBox3.Width, processBox3.Y+processBox3.Height/2,
		dataBox.X, dataBox.Y+dataBox.Height)
	
	edge7 := components.NewFlowEdge("edge7", "data1", "out1", "end", "in1")
	edge7.Label = "Complete"
	edge7.UpdatePosition(dataBox.X+dataBox.Width, dataBox.Y+dataBox.Height/2,
		endBox.X, endBox.Y+endBox.Height/2)
	
	// Add edges to canvas
	canvas.AddEdge(edge1)
	canvas.AddEdge(edge2)
	canvas.AddEdge(edge3)
	canvas.AddEdge(edge4)
	canvas.AddEdge(edge5)
	canvas.AddEdge(edge6)
	canvas.AddEdge(edge7)
	
	// Set up callbacks
	canvas.OnBoxClick = func(boxID string) {
		log.Printf("Box clicked: %s", boxID)
	}
	
	canvas.OnEdgeClick = func(edgeID string) {
		log.Printf("Edge clicked: %s", edgeID)
	}
	
	canvas.OnConnection = func(fromBox, fromPort, toBox, toPort string) {
		log.Printf("Connection made: %s:%s -> %s:%s", fromBox, fromPort, toBox, toPort)
	}
	
	canvas.OnBoxMove = func(boxID string, x, y int) {
		log.Printf("Box %s moved to (%d, %d)", boxID, x, y)
	}
	
	return &FlowToolExample{
		Canvas:      canvas,
		Modal:       modal,
		Title:       "Flow Diagram Tool",
		Description: "Interactive flow diagram editor without JavaScript",
		NodeCount:   7,
		EdgeCount:   7,
		LastAction:  "Diagram initialized",
	}
}

func (f *FlowToolExample) Start() {
	// Initialize modal events
	if f.Modal != nil && f.Modal.ComponentDriver != nil {
		f.Modal.Start()
	}
	
	// Initialize event handlers directly on the ComponentDriver
	if f.ComponentDriver != nil {
		f.ComponentDriver.Events["AddNode"] = func(c *FlowToolExample, data interface{}) {
			c.HandleAddNode(data)
		}
		f.ComponentDriver.Events["ClearDiagram"] = func(c *FlowToolExample, data interface{}) {
			c.ClearDiagram(data)
		}
		f.ComponentDriver.Events["ExportDiagram"] = func(c *FlowToolExample, data interface{}) {
			c.ExportDiagram(data)
		}
		f.ComponentDriver.Events["AnimateFlow"] = func(c *FlowToolExample, data interface{}) {
			c.AnimateFlow(data)
		}
		// Add box interaction events
		f.ComponentDriver.Events["BoxClick"] = func(c *FlowToolExample, data interface{}) {
			c.HandleBoxClick(data)
		}
		f.ComponentDriver.Events["MoveBox"] = func(c *FlowToolExample, data interface{}) {
			c.HandleMoveBox(data)
		}
		// Add canvas events
		f.ComponentDriver.Events["CanvasZoomIn"] = func(c *FlowToolExample, data interface{}) {
			c.HandleCanvasZoomIn(data)
		}
		f.ComponentDriver.Events["CanvasZoomOut"] = func(c *FlowToolExample, data interface{}) {
			c.HandleCanvasZoomOut(data)
		}
		f.ComponentDriver.Events["CanvasReset"] = func(c *FlowToolExample, data interface{}) {
			c.HandleCanvasReset(data)
		}
		f.ComponentDriver.Events["ToggleGrid"] = func(c *FlowToolExample, data interface{}) {
			c.HandleToggleGrid(data)
		}
		f.ComponentDriver.Events["ToggleConnectMode"] = func(c *FlowToolExample, data interface{}) {
			c.HandleToggleConnectMode(data)
		}
		// Add drag events
		f.ComponentDriver.Events["BoxStartDrag"] = func(c *FlowToolExample, data interface{}) {
			c.BoxStartDrag(data)
		}
		f.ComponentDriver.Events["BoxDrag"] = func(c *FlowToolExample, data interface{}) {
			c.BoxDrag(data)
		}
		f.ComponentDriver.Events["BoxEndDrag"] = func(c *FlowToolExample, data interface{}) {
			c.BoxEndDrag(data)
		}
	}
	
	if f.ComponentDriver != nil {
		f.Commit()
	}
}

func (f *FlowToolExample) GetTemplate() string {
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
	</style>
</head>
<body>
	<div class="container">
		<div class="header">
			<h1 class="title">{{.Title}}</h1>
			<p class="description">{{.Description}}</p>
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
					
					<button class="btn btn-success" onclick="send_event('{{.IdComponent}}', 'AnimateFlow', null)">
						Animate Flow
					</button>
					
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
							<button onclick="send_event('{{$.IdComponent}}', 'ClearDiagram', null)" style="padding: 0.5rem; background: white; border: 1px solid #d1d5db; border-radius: 4px; cursor: pointer;">Clear</button>
						</div>
						
						<div id="canvas-viewport" style="position: relative; width: 100%; height: 100%; transform: scale({{.Canvas.Zoom}}) translate({{.Canvas.PanX}}px, {{.Canvas.PanY}}px); transform-origin: 0 0; transition: transform 0.2s;">
							<!-- Render boxes -->
							{{range $id, $box := .Canvas.Boxes}}
								<div id="box-{{$id}}" 
								     style="position: absolute; left: {{$box.X}}px; top: {{$box.Y}}px; width: {{$box.Width}}px; height: {{$box.Height}}px; background: {{$box.Color}}; border: 2px solid {{if $box.Selected}}#2563eb{{else}}#cbd5e1{{end}}; border-radius: 8px; padding: 0.5rem; cursor: {{if $.ConnectingMode}}pointer{{else}}move{{end}}; box-shadow: 0 2px 4px rgba(0,0,0,0.1); user-select: none;"
								     onclick="if({{$.ConnectingMode}}) { send_event('{{$.IdComponent}}', 'BoxClick', '{{$id}}'); } else { send_event('{{$.IdComponent}}', 'BoxClick', '{{$id}}'); }"
								     onmousedown="if(!{{$.ConnectingMode}}) { event.preventDefault(); window.draggedBox = '{{$id}}'; window.dragStartX = event.clientX; window.dragStartY = event.clientY; window.dragInitX = {{$box.X}}; window.dragInitY = {{$box.Y}}; send_event('{{$.IdComponent}}', 'BoxStartDrag', {id: '{{$id}}', x: event.clientX, y: event.clientY}); }">
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
				<strong>Last Action:</strong> {{.LastAction}}
				{{if .Canvas}}
					{{range $id, $box := .Canvas.Boxes}}
						{{if $box.Selected}}
							| <strong>Selected:</strong> {{$box.Label}} ({{$box.X}}, {{$box.Y}})
						{{end}}
					{{end}}
				{{end}}
			</div>
		</div>
		
		<!-- Modal Component -->
		{{mount "export-modal"}}
	</div>
	
	<script>
	// Global drag handling for boxes
	document.addEventListener('mousemove', function(e) {
		if (window.draggedBox) {
			var deltaX = e.clientX - window.dragStartX;
			var deltaY = e.clientY - window.dragStartY;
			var newX = window.dragInitX + deltaX;
			var newY = window.dragInitY + deltaY;
			
			// Send to the main component to handle drag
			send_event('{{.IdComponent}}', 'BoxDrag', {
				id: window.draggedBox,
				x: newX,
				y: newY
			});
		}
	});
	
	document.addEventListener('mouseup', function(e) {
		if (window.draggedBox) {
			send_event('{{.IdComponent}}', 'BoxEndDrag', window.draggedBox);
			window.draggedBox = null;
		}
	});
	</script>
</body>
</html>
`
}

func (f *FlowToolExample) GetDriver() liveview.LiveDriver {
	return f
}

func (f *FlowToolExample) HandleAddNode(data interface{}) {
	nodeType := data.(string)
	
	// Calculate position for new node
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
	
	// Create driver for the new box
	if f.Canvas != nil {
		boxDriver := liveview.NewDriver(nodeID, newBox)
		newBox.ComponentDriver = boxDriver
		newBox.Start()
		f.ComponentDriver.Mount(newBox) // Mount the new box
		f.Canvas.AddBox(newBox)
	}
	
	f.NodeCount++
	f.LastAction = fmt.Sprintf("Added %s node", nodeType)
	if f.ComponentDriver != nil {
		f.Commit()
	}
}

func (f *FlowToolExample) ClearDiagram(data interface{}) {
	f.Canvas.Clear()
	f.NodeCount = 0
	f.EdgeCount = 0
	f.LastAction = "Diagram cleared"
	f.Commit()
}

func (f *FlowToolExample) ExportDiagram(data interface{}) {
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
		// Store the JSON
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

func (f *FlowToolExample) AnimateFlow(data interface{}) {
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

func (f *FlowToolExample) HandleBoxClick(data interface{}) {
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
			edgeID := fmt.Sprintf("edge_%s_%s", f.ConnectingFrom, boxID)
			
			// Check if edge already exists
			if _, exists := f.Canvas.Edges[edgeID]; !exists {
				edge := components.NewFlowEdge(edgeID, f.ConnectingFrom, "out1", boxID, "in1")
				
				// Set positions
				if fromBox, ok := f.Canvas.Boxes[f.ConnectingFrom]; ok {
					if toBox, ok := f.Canvas.Boxes[boxID]; ok {
						edge.FromX = fromBox.X + fromBox.Width
						edge.FromY = fromBox.Y + fromBox.Height/2
						edge.ToX = toBox.X
						edge.ToY = toBox.Y + toBox.Height/2
					}
				}
				
				f.Canvas.Edges[edgeID] = edge
				f.EdgeCount++
				f.LastAction = fmt.Sprintf("Created edge: %s -> %s", f.ConnectingFrom, boxID)
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

func (f *FlowToolExample) HandleMoveBox(data interface{}) {
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

func (f *FlowToolExample) HandleToggleConnectMode(data interface{}) {
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

func (f *FlowToolExample) BoxStartDrag(data interface{}) {
	// Handle drag start
	if dataMap, ok := data.(map[string]interface{}); ok {
		if boxID, ok := dataMap["id"].(string); ok {
			f.DraggingBox = boxID
			if box, exists := f.Canvas.Boxes[boxID]; exists {
				box.Dragging = true
				if box.ComponentDriver != nil {
					box.Commit()
				}
			}
			f.LastAction = fmt.Sprintf("Started dragging %s", boxID)
		}
	}
	
	if f.ComponentDriver != nil {
		f.Commit()
	}
}

func (f *FlowToolExample) BoxDrag(data interface{}) {
	if f.DraggingBox == "" {
		return
	}
	
	// Handle drag movement
	if dataMap, ok := data.(map[string]interface{}); ok {
		if box, exists := f.Canvas.Boxes[f.DraggingBox]; exists {
			if newX, ok := dataMap["x"].(float64); ok {
				box.X = int(newX)
			}
			if newY, ok := dataMap["y"].(float64); ok {
				box.Y = int(newY)
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
			
			// Update edge positions
			f.updateEdgePositions()
			
			if box.ComponentDriver != nil {
				box.Commit()
			}
		}
	}
	
	if f.ComponentDriver != nil {
		f.Commit()
	}
}

func (f *FlowToolExample) BoxEndDrag(data interface{}) {
	if f.DraggingBox != "" {
		if box, exists := f.Canvas.Boxes[f.DraggingBox]; exists {
			box.Dragging = false
			if box.ComponentDriver != nil {
				box.Commit()
			}
		}
		f.LastAction = fmt.Sprintf("Finished dragging %s", f.DraggingBox)
		f.DraggingBox = ""
	}
	
	if f.ComponentDriver != nil {
		f.Commit()
	}
}

func (f *FlowToolExample) updateEdgePositions() {
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

func (f *FlowToolExample) HandleCanvasZoomIn(data interface{}) {
	f.Canvas.Zoom = min(f.Canvas.Zoom*1.2, 3.0)
	f.LastAction = fmt.Sprintf("Zoom: %d%%", f.Canvas.ZoomPercent())
	if f.ComponentDriver != nil {
		f.Commit()
	}
}

func (f *FlowToolExample) HandleCanvasZoomOut(data interface{}) {
	f.Canvas.Zoom = max(f.Canvas.Zoom/1.2, 0.3)
	f.LastAction = fmt.Sprintf("Zoom: %d%%", f.Canvas.ZoomPercent())
	if f.ComponentDriver != nil {
		f.Commit()
	}
}

func (f *FlowToolExample) HandleCanvasReset(data interface{}) {
	f.Canvas.Zoom = 1.0
	f.Canvas.PanX = 0
	f.Canvas.PanY = 0
	f.LastAction = "View reset"
	if f.ComponentDriver != nil {
		f.Commit()
	}
}

func (f *FlowToolExample) HandleToggleGrid(data interface{}) {
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
		Title:  "Flow Diagram Tool Example",
		Lang:   "en",
		Path:   "/example/flowtool",
		Router: e,
	}

	home.Register(func() liveview.LiveDriver {
		// Create layout wrapper
		document := liveview.NewLayout("flowtool-layout", `{{mount "flow-tool"}}`)
		
		// Create flow tool component
		flowTool := NewFlowToolExample()
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
			<h1>Flow Diagram Tool Example</h1>
			<p>Visit <a href="/example/flowtool">/example/flowtool</a> to see the interactive flow diagram editor</p>
		`)
	})

	port := ":8080"
	log.Printf("Starting server on %s", port)
	log.Printf("Visit http://localhost%s/example/flowtool", port)
	e.Logger.Fatal(e.Start(port))
}