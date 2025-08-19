package components

import (
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/arturoeanton/go-echo-live-view/liveview"
)

// CollaborativeCanvas provides real-time collaborative drawing
type CollaborativeCanvas struct {
	*liveview.ComponentDriver[*CollaborativeCanvas]
	*liveview.CollaborativeComponent

	// Canvas properties
	Width      int    `json:"width"`
	Height     int    `json:"height"`
	Background string `json:"background"`

	// Drawing state
	Shapes      []Shape `json:"shapes"`
	CurrentTool string  `json:"current_tool"`
	StrokeColor string  `json:"stroke_color"`
	FillColor   string  `json:"fill_color"`
	StrokeWidth int     `json:"stroke_width"`

	// Collaboration
	Cursors    map[string]*CursorInfo `json:"cursors"`
	Selections map[string][]string    `json:"selections"` // user -> selected shape IDs
}

// Shape represents a drawable shape
type Shape struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"` // line, rect, circle, path, text
	X           float64   `json:"x"`
	Y           float64   `json:"y"`
	Width       float64   `json:"width,omitempty"`
	Height      float64   `json:"height,omitempty"`
	Radius      float64   `json:"radius,omitempty"`
	Points      []Point   `json:"points,omitempty"`
	Text        string    `json:"text,omitempty"`
	StrokeColor string    `json:"stroke_color"`
	FillColor   string    `json:"fill_color"`
	StrokeWidth int       `json:"stroke_width"`
	CreatedBy   string    `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Locked      bool      `json:"locked"`
	Layer       int       `json:"layer"`
}

// Point represents a coordinate
type Point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// CursorInfo tracks cursor information
type CursorInfo struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Color  string  `json:"color"`
	Name   string  `json:"name"`
	Active bool    `json:"active"`
}

// Start initializes the canvas
func (c *CollaborativeCanvas) Start() {
	// Initialize canvas
	c.Width = 1200
	c.Height = 800
	c.Background = "#ffffff"
	c.Shapes = make([]Shape, 0)
	c.CurrentTool = "pen"
	c.StrokeColor = "#000000"
	c.FillColor = "transparent"
	c.StrokeWidth = 2
	c.Cursors = make(map[string]*CursorInfo)
	c.Selections = make(map[string][]string)

	// Initialize collaboration
	c.CollaborativeComponent = &liveview.CollaborativeComponent{}
	roomID := fmt.Sprintf("canvas_%s", c.GetIDComponet())
	c.StartCollaboration(roomID, c.GetIDComponet(), "User")

	c.Commit()
}

// GetTemplate returns the canvas HTML template
func (c *CollaborativeCanvas) GetTemplate() string {
	return `
	<div class="collaborative-canvas" id="{{.IdComponent}}">
		<style>
			.collaborative-canvas {
				position: relative;
				border: 1px solid #ddd;
				background: {{.Background}};
				width: {{.Width}}px;
				height: {{.Height}}px;
				margin: 20px auto;
				cursor: crosshair;
			}
			
			.canvas-toolbar {
				position: absolute;
				top: 10px;
				left: 10px;
				background: white;
				border-radius: 8px;
				padding: 10px;
				box-shadow: 0 2px 10px rgba(0,0,0,0.1);
				display: flex;
				gap: 10px;
				z-index: 100;
			}
			
			.tool-button {
				width: 40px;
				height: 40px;
				border: 2px solid #ddd;
				background: white;
				border-radius: 4px;
				cursor: pointer;
				display: flex;
				align-items: center;
				justify-content: center;
				transition: all 0.2s;
			}
			
			.tool-button:hover {
				border-color: #4CAF50;
			}
			
			.tool-button.active {
				background: #4CAF50;
				color: white;
				border-color: #4CAF50;
			}
			
			.color-picker {
				width: 40px;
				height: 40px;
				border: 2px solid #ddd;
				border-radius: 4px;
				cursor: pointer;
			}
			
			.canvas-svg {
				position: absolute;
				top: 0;
				left: 0;
				width: 100%;
				height: 100%;
			}
			
			.remote-cursor {
				position: absolute;
				width: 20px;
				height: 20px;
				pointer-events: none;
				z-index: 1000;
				transition: all 0.1s ease-out;
			}
			
			.cursor-label {
				position: absolute;
				top: 20px;
				left: 5px;
				background: rgba(0,0,0,0.8);
				color: white;
				padding: 2px 6px;
				border-radius: 4px;
				font-size: 12px;
				white-space: nowrap;
			}
			
			.selection-box {
				stroke: #4CAF50;
				stroke-width: 2;
				stroke-dasharray: 5,5;
				fill: none;
				animation: dash 0.5s linear infinite;
			}
			
			@keyframes dash {
				to {
					stroke-dashoffset: -10;
				}
			}
		</style>
		
		<!-- Toolbar -->
		<div class="canvas-toolbar">
			<button class="tool-button {{if eq .CurrentTool "select"}}active{{end}}" 
			        onclick="send_event('{{.IdComponent}}', 'SelectTool', 'select')" 
			        title="Select">
				<svg width="20" height="20"><path d="M4 4 L16 12 L10 14 L8 20 Z" fill="currentColor"/></svg>
			</button>
			
			<button class="tool-button {{if eq .CurrentTool "pen"}}active{{end}}" 
			        onclick="send_event('{{.IdComponent}}', 'SelectTool', 'pen')" 
			        title="Pen">
				<svg width="20" height="20"><path d="M3 17 L15 5 L17 7 L5 19 L3 17 Z" fill="currentColor"/></svg>
			</button>
			
			<button class="tool-button {{if eq .CurrentTool "rect"}}active{{end}}" 
			        onclick="send_event('{{.IdComponent}}', 'SelectTool', 'rect')" 
			        title="Rectangle">
				<svg width="20" height="20"><rect x="4" y="4" width="12" height="12" fill="none" stroke="currentColor" stroke-width="2"/></svg>
			</button>
			
			<button class="tool-button {{if eq .CurrentTool "circle"}}active{{end}}" 
			        onclick="send_event('{{.IdComponent}}', 'SelectTool', 'circle')" 
			        title="Circle">
				<svg width="20" height="20"><circle cx="10" cy="10" r="7" fill="none" stroke="currentColor" stroke-width="2"/></svg>
			</button>
			
			<button class="tool-button {{if eq .CurrentTool "text"}}active{{end}}" 
			        onclick="send_event('{{.IdComponent}}', 'SelectTool', 'text')" 
			        title="Text">
				<svg width="20" height="20"><text x="10" y="15" text-anchor="middle" font-size="16">T</text></svg>
			</button>
			
			<div style="border-left: 1px solid #ddd; margin: 0 5px;"></div>
			
			<input type="color" class="color-picker" value="{{.StrokeColor}}" 
			       onchange="send_event('{{.IdComponent}}', 'ChangeStrokeColor', this.value)" 
			       title="Stroke Color">
			
			<input type="color" class="color-picker" value="{{.FillColor}}" 
			       onchange="send_event('{{.IdComponent}}', 'ChangeFillColor', this.value)" 
			       title="Fill Color">
			
			<input type="range" min="1" max="10" value="{{.StrokeWidth}}" 
			       onchange="send_event('{{.IdComponent}}', 'ChangeStrokeWidth', this.value)" 
			       title="Stroke Width">
			
			<div style="border-left: 1px solid #ddd; margin: 0 5px;"></div>
			
			<button class="tool-button" onclick="send_event('{{.IdComponent}}', 'Undo')" title="Undo">
				↶
			</button>
			
			<button class="tool-button" onclick="send_event('{{.IdComponent}}', 'Redo')" title="Redo">
				↷
			</button>
			
			<button class="tool-button" onclick="send_event('{{.IdComponent}}', 'Clear')" title="Clear">
				🗑
			</button>
		</div>
		
		<!-- Canvas SVG -->
		<svg class="canvas-svg" 
		     onmousedown="handleCanvasMouseDown(event, '{{.IdComponent}}')"
		     onmousemove="handleCanvasMouseMove(event, '{{.IdComponent}}')"
		     onmouseup="handleCanvasMouseUp(event, '{{.IdComponent}}')">
			
			<!-- Render shapes -->
			{{range .Shapes}}
				{{if eq .Type "rect"}}
					<rect id="shape-{{.ID}}" 
					      x="{{.X}}" y="{{.Y}}" 
					      width="{{.Width}}" height="{{.Height}}"
					      fill="{{.FillColor}}" 
					      stroke="{{.StrokeColor}}" 
					      stroke-width="{{.StrokeWidth}}"
					      data-shape-id="{{.ID}}"/>
				{{else if eq .Type "circle"}}
					<circle id="shape-{{.ID}}" 
					        cx="{{.X}}" cy="{{.Y}}" r="{{.Radius}}"
					        fill="{{.FillColor}}" 
					        stroke="{{.StrokeColor}}" 
					        stroke-width="{{.StrokeWidth}}"
					        data-shape-id="{{.ID}}"/>
				{{else if eq .Type "path"}}
					<path id="shape-{{.ID}}" 
					      d="M {{range $i, $p := .Points}}{{if $i}}L {{end}}{{$p.X}} {{$p.Y}} {{end}}"
					      fill="none" 
					      stroke="{{.StrokeColor}}" 
					      stroke-width="{{.StrokeWidth}}"
					      stroke-linecap="round" 
					      stroke-linejoin="round"
					      data-shape-id="{{.ID}}"/>
				{{else if eq .Type "text"}}
					<text id="shape-{{.ID}}" 
					      x="{{.X}}" y="{{.Y}}"
					      fill="{{.StrokeColor}}" 
					      font-size="16"
					      data-shape-id="{{.ID}}">{{.Text}}</text>
				{{end}}
			{{end}}
			
			<!-- Selection boxes -->
			{{range $userID, $selections := .Selections}}
				{{range $selections}}
					<rect class="selection-box" x="0" y="0" width="50" height="50"/>
				{{end}}
			{{end}}
		</svg>
		
		<!-- Remote cursors -->
		{{range $id, $cursor := .Cursors}}
			{{if $cursor.Active}}
			<div class="remote-cursor" style="left: {{$cursor.X}}px; top: {{$cursor.Y}}px;">
				<svg width="20" height="20">
					<path d="M0 0 L15 15 L7 15 L5 20 Z" fill="{{$cursor.Color}}"/>
				</svg>
				<div class="cursor-label" style="background: {{$cursor.Color}}">{{$cursor.Name}}</div>
			</div>
			{{end}}
		{{end}}
		
		<script>
			// Canvas interaction handlers
			let isDrawing = false;
			let startX = 0;
			let startY = 0;
			let currentPath = [];
			
			function handleCanvasMouseDown(event, componentId) {
				const rect = event.currentTarget.getBoundingClientRect();
				startX = event.clientX - rect.left;
				startY = event.clientY - rect.top;
				isDrawing = true;
				currentPath = [{x: startX, y: startY}];
				
				send_event(componentId, 'StartDrawing', {
					x: startX,
					y: startY,
					tool: getCurrentTool()
				});
			}
			
			function handleCanvasMouseMove(event, componentId) {
				const rect = event.currentTarget.getBoundingClientRect();
				const x = event.clientX - rect.left;
				const y = event.clientY - rect.top;
				
				// Update cursor position
				send_event(componentId, 'UpdateCursor', {x: x, y: y});
				
				if (isDrawing) {
					currentPath.push({x: x, y: y});
					
					send_event(componentId, 'Drawing', {
						x: x,
						y: y,
						path: currentPath
					});
				}
			}
			
			function handleCanvasMouseUp(event, componentId) {
				if (isDrawing) {
					const rect = event.currentTarget.getBoundingClientRect();
					const endX = event.clientX - rect.left;
					const endY = event.clientY - rect.top;
					
					send_event(componentId, 'EndDrawing', {
						startX: startX,
						startY: startY,
						endX: endX,
						endY: endY,
						path: currentPath
					});
					
					isDrawing = false;
					currentPath = [];
				}
			}
			
			function getCurrentTool() {
				return document.querySelector('.tool-button.active')?.dataset?.tool || 'pen';
			}
		</script>
	</div>
	`
}

// GetDriver returns the component driver
func (c *CollaborativeCanvas) GetDriver() liveview.LiveDriver {
	return c.ComponentDriver
}

// SelectTool changes the current drawing tool
func (c *CollaborativeCanvas) SelectTool(data interface{}) {
	if tool, ok := data.(string); ok {
		c.CurrentTool = tool
		c.Commit()
	}
}

// ChangeStrokeColor updates stroke color
func (c *CollaborativeCanvas) ChangeStrokeColor(data interface{}) {
	if color, ok := data.(string); ok {
		c.StrokeColor = color
		c.Commit()
	}
}

// ChangeFillColor updates fill color
func (c *CollaborativeCanvas) ChangeFillColor(data interface{}) {
	if color, ok := data.(string); ok {
		c.FillColor = color
		c.Commit()
	}
}

// ChangeStrokeWidth updates stroke width
func (c *CollaborativeCanvas) ChangeStrokeWidth(data interface{}) {
	if width, ok := data.(float64); ok {
		c.StrokeWidth = int(width)
		c.Commit()
	}
}

// StartDrawing begins a new shape
func (c *CollaborativeCanvas) StartDrawing(data interface{}) {
	if drawData, ok := data.(map[string]interface{}); ok {
		// Broadcast drawing start to other users
		c.BroadcastAction("drawing_start", drawData)
	}
}

// Drawing handles ongoing drawing
func (c *CollaborativeCanvas) Drawing(data interface{}) {
	if drawData, ok := data.(map[string]interface{}); ok {
		// Update temporary drawing state
		c.BroadcastAction("drawing_update", drawData)
	}
}

// EndDrawing completes a shape
func (c *CollaborativeCanvas) EndDrawing(data interface{}) {
	if drawData, ok := data.(map[string]interface{}); ok {
		shape := c.createShapeFromData(drawData)
		if shape != nil {
			c.Shapes = append(c.Shapes, *shape)

			// Sync with collaboration room
			c.SyncState("add_shape", shape)

			// Broadcast to other users
			c.BroadcastAction("shape_added", shape)

			c.Commit()
		}
	}
}

// createShapeFromData creates a shape from drawing data
func (c *CollaborativeCanvas) createShapeFromData(data map[string]interface{}) *Shape {
	shape := &Shape{
		ID:          fmt.Sprintf("shape_%d", time.Now().UnixNano()),
		StrokeColor: c.StrokeColor,
		FillColor:   c.FillColor,
		StrokeWidth: c.StrokeWidth,
		CreatedBy:   c.UserID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Layer:       len(c.Shapes),
	}

	switch c.CurrentTool {
	case "rect":
		shape.Type = "rect"
		shape.X = data["startX"].(float64)
		shape.Y = data["startY"].(float64)
		shape.Width = data["endX"].(float64) - shape.X
		shape.Height = data["endY"].(float64) - shape.Y

	case "circle":
		shape.Type = "circle"
		shape.X = data["startX"].(float64)
		shape.Y = data["startY"].(float64)
		dx := data["endX"].(float64) - shape.X
		dy := data["endY"].(float64) - shape.Y
		shape.Radius = math.Sqrt(dx*dx + dy*dy)

	case "pen", "path":
		shape.Type = "path"
		if pathData, ok := data["path"].([]interface{}); ok {
			shape.Points = make([]Point, 0, len(pathData))
			for _, p := range pathData {
				if point, ok := p.(map[string]interface{}); ok {
					shape.Points = append(shape.Points, Point{
						X: point["x"].(float64),
						Y: point["y"].(float64),
					})
				}
			}
		}

	case "text":
		shape.Type = "text"
		shape.X = data["x"].(float64)
		shape.Y = data["y"].(float64)
		shape.Text = "New Text"
	}

	return shape
}

// UpdateCursor updates cursor position for collaboration
func (c *CollaborativeCanvas) UpdateCursor(data interface{}) {
	if cursorData, ok := data.(map[string]interface{}); ok {
		x := cursorData["x"].(float64)
		y := cursorData["y"].(float64)

		// Update local cursor display for remote users
		c.UpdateCursorPosition(x, y, c.GetIDComponet())

		// Broadcast cursor position
		c.BroadcastAction("cursor_move", map[string]interface{}{
			"user_id": c.UserID,
			"x":       x,
			"y":       y,
		})
	}
}

// Clear removes all shapes
func (c *CollaborativeCanvas) Clear(data interface{}) {
	c.Shapes = []Shape{}
	c.SyncState("clear", nil)
	c.BroadcastAction("canvas_cleared", nil)
	c.Commit()
}

// Undo removes the last shape
func (c *CollaborativeCanvas) Undo(data interface{}) {
	if len(c.Shapes) > 0 {
		lastShape := c.Shapes[len(c.Shapes)-1]
		c.Shapes = c.Shapes[:len(c.Shapes)-1]

		c.SyncState("remove_shape", lastShape.ID)
		c.BroadcastAction("shape_removed", lastShape.ID)
		c.Commit()
	}
}

// Redo re-adds a removed shape
func (c *CollaborativeCanvas) Redo(data interface{}) {
	// Implementation would require maintaining a redo stack
	c.Commit()
}

// HandleRemoteUpdate processes updates from other users
func (c *CollaborativeCanvas) HandleRemoteUpdate(data interface{}) {
	if updateData, ok := data.(map[string]interface{}); ok {
		updateType := updateData["type"].(string)

		switch updateType {
		case "shape_added":
			// Add shape from remote user
			if shapeData, ok := updateData["shape"].(map[string]interface{}); ok {
				var shape Shape
				if jsonData, err := json.Marshal(shapeData); err == nil {
					if err := json.Unmarshal(jsonData, &shape); err == nil {
						c.Shapes = append(c.Shapes, shape)
						c.Commit()
					}
				}
			}

		case "shape_removed":
			// Remove shape
			if shapeID, ok := updateData["shape_id"].(string); ok {
				newShapes := make([]Shape, 0)
				for _, s := range c.Shapes {
					if s.ID != shapeID {
						newShapes = append(newShapes, s)
					}
				}
				c.Shapes = newShapes
				c.Commit()
			}

		case "cursor_move":
			// Update remote cursor position
			if userID, ok := updateData["user_id"].(string); ok {
				if userID != c.UserID {
					x := updateData["x"].(float64)
					y := updateData["y"].(float64)

					if cursor, exists := c.Cursors[userID]; exists {
						cursor.X = x
						cursor.Y = y
						cursor.Active = true
					} else {
						c.Cursors[userID] = &CursorInfo{
							X:      x,
							Y:      y,
							Color:  c.getColorForUser(userID),
							Name:   userID,
							Active: true,
						}
					}
					c.Commit()
				}
			}
		}
	}
}

// getColorForUser generates a consistent color for a user
func (c *CollaborativeCanvas) getColorForUser(userID string) string {
	colors := []string{"#FF6B6B", "#4ECDC4", "#45B7D1", "#96CEB4", "#FECA57"}
	hash := 0
	for _, ch := range userID {
		hash = (hash + int(ch)) % len(colors)
	}
	return colors[hash]
}

// Export exports canvas as JSON
func (c *CollaborativeCanvas) Export(data interface{}) {
	exportData := map[string]interface{}{
		"width":      c.Width,
		"height":     c.Height,
		"background": c.Background,
		"shapes":     c.Shapes,
	}

	if jsonData, err := json.Marshal(exportData); err == nil {
		// Send export data via property update
		c.ComponentDriver.SetPropertie("exportData", string(jsonData))
	}
}

// Import imports canvas from JSON
func (c *CollaborativeCanvas) Import(data interface{}) {
	if jsonStr, ok := data.(string); ok {
		var importData map[string]interface{}
		if err := json.Unmarshal([]byte(jsonStr), &importData); err == nil {
			// Update canvas properties
			if width, ok := importData["width"].(float64); ok {
				c.Width = int(width)
			}
			if height, ok := importData["height"].(float64); ok {
				c.Height = int(height)
			}
			if bg, ok := importData["background"].(string); ok {
				c.Background = bg
			}

			// Import shapes
			if shapes, ok := importData["shapes"].([]interface{}); ok {
				c.Shapes = make([]Shape, 0, len(shapes))
				for _, s := range shapes {
					if shapeData, err := json.Marshal(s); err == nil {
						var shape Shape
						if err := json.Unmarshal(shapeData, &shape); err == nil {
							c.Shapes = append(c.Shapes, shape)
						}
					}
				}
			}

			c.SyncState("import", importData)
			c.BroadcastAction("canvas_imported", nil)
			c.Commit()
		}
	}
}
