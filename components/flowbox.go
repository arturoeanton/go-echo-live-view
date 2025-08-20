package components

import (
	"github.com/arturoeanton/go-echo-live-view/liveview"
)

type BoxType string

const (
	BoxTypeStart    BoxType = "start"
	BoxTypeProcess  BoxType = "process"
	BoxTypeDecision BoxType = "decision"
	BoxTypeData     BoxType = "data"
	BoxTypeEnd      BoxType = "end"
	BoxTypeCustom   BoxType = "custom"
)

type Port struct {
	ID        string
	Type      string // "input" or "output"
	Position  string // "top", "right", "bottom", "left"
	Label     string
	Connected bool
}

type FlowBox struct {
	*liveview.ComponentDriver[*FlowBox]

	ID          string
	Label       string
	Description string
	Type        BoxType
	X           int
	Y           int
	Width       int
	Height      int
	Selected    bool
	Dragging    bool
	Color       string
	Icon        string
	Ports       []Port
	Data        map[string]interface{}
	OnClick     func(boxID string)
	OnConnect   func(fromBox, fromPort, toBox, toPort string)
}

func NewFlowBox(id, label string, boxType BoxType, x, y int) *FlowBox {
	box := &FlowBox{
		ID:     id,
		Label:  label,
		Type:   boxType,
		X:      x,
		Y:      y,
		Width:  150,
		Height: 80,
		Color:  getBoxColor(boxType),
		Ports:  generatePorts(boxType),
		Data:   make(map[string]interface{}),
	}

	if boxType == BoxTypeDecision {
		box.Height = 100
	}

	return box
}

func (b *FlowBox) Start() {
	// Events are registered directly on the ComponentDriver
	if b.ComponentDriver != nil {
		b.ComponentDriver.Events["Click"] = func(c *FlowBox, data interface{}) {
			c.HandleClick(data)
		}
		b.ComponentDriver.Events["StartDrag"] = func(c *FlowBox, data interface{}) {
			c.StartDrag(data)
		}
		b.ComponentDriver.Events["Drag"] = func(c *FlowBox, data interface{}) {
			c.HandleDrag(data)
		}
		b.ComponentDriver.Events["EndDrag"] = func(c *FlowBox, data interface{}) {
			c.EndDrag(data)
		}
		b.ComponentDriver.Events["ConnectPort"] = func(c *FlowBox, data interface{}) {
			c.HandlePortConnect(data)
		}
		b.ComponentDriver.Events["DisconnectPort"] = func(c *FlowBox, data interface{}) {
			c.HandlePortDisconnect(data)
		}
	}
}

func (b *FlowBox) GetTemplate() string {
	return `
<div class="flow-box {{.Type}}" 
     id="box-{{.ID}}"
     style="left: {{.X}}px; top: {{.Y}}px; width: {{.Width}}px; height: {{.Height}}px;"
     onclick="send_event('{{.IdComponent}}', 'Click', '{{.ID}}')">
	
	<style>
		.flow-box {
			position: absolute;
			background: {{.Color}};
			border: 2px solid {{if .Selected}}#2563eb{{else}}#cbd5e1{{end}};
			border-radius: {{if eq .Type "decision"}}0{{else}}8{{end}}px;
			padding: 0.75rem;
			cursor: move;
			transition: all 0.2s;
			box-shadow: 0 2px 4px rgba(0,0,0,0.1);
			display: flex;
			flex-direction: column;
			user-select: none;
		}
		
		.flow-box.decision {
			transform: rotate(45deg);
		}
		
		.flow-box.decision .box-content {
			transform: rotate(-45deg);
		}
		
		.flow-box:hover {
			box-shadow: 0 4px 8px rgba(0,0,0,0.15);
			z-index: 10;
		}
		
		.flow-box.selected {
			box-shadow: 0 0 0 3px rgba(37, 99, 235, 0.2);
			z-index: 20;
		}
		
		.flow-box.dragging {
			opacity: 0.8;
			z-index: 100;
		}
		
		.box-header {
			display: flex;
			align-items: center;
			gap: 0.5rem;
			margin-bottom: 0.25rem;
		}
		
		.box-icon {
			width: 20px;
			height: 20px;
			opacity: 0.7;
		}
		
		.box-label {
			font-weight: 600;
			color: #1f2937;
			font-size: 0.875rem;
			white-space: nowrap;
			overflow: hidden;
			text-overflow: ellipsis;
		}
		
		.box-description {
			font-size: 0.75rem;
			color: #6b7280;
			overflow: hidden;
			text-overflow: ellipsis;
			display: -webkit-box;
			-webkit-line-clamp: 2;
			-webkit-box-orient: vertical;
		}
		
		.box-ports {
			position: absolute;
			width: 100%;
			height: 100%;
			top: 0;
			left: 0;
			pointer-events: none;
		}
		
		.port {
			position: absolute;
			width: 12px;
			height: 12px;
			background: white;
			border: 2px solid #64748b;
			border-radius: 50%;
			cursor: crosshair;
			pointer-events: all;
			transition: all 0.2s;
		}
		
		.port:hover {
			transform: scale(1.3);
			border-color: #2563eb;
			background: #dbeafe;
		}
		
		.port.connected {
			background: #10b981;
			border-color: #059669;
		}
		
		.port.input {
			left: -6px;
		}
		
		.port.output {
			right: -6px;
		}
		
		.port.top {
			top: -6px;
			left: 50%;
			transform: translateX(-50%);
		}
		
		.port.bottom {
			bottom: -6px;
			left: 50%;
			transform: translateX(-50%);
		}
		
		.port.left {
			top: 50%;
			transform: translateY(-50%);
		}
		
		.port.right {
			top: 50%;
			transform: translateY(-50%);
		}
		
		.drag-handle {
			position: absolute;
			top: 0;
			left: 0;
			right: 0;
			height: 30px;
			cursor: move;
		}
	</style>
	
	<div class="drag-handle"
	     onmousedown="send_event('{{.IdComponent}}', 'StartDrag', {id: '{{.ID}}', x: event.clientX, y: event.clientY})"
	     onmousemove="if({{.Dragging}}) send_event('{{.IdComponent}}', 'Drag', {id: '{{.ID}}', x: event.clientX, y: event.clientY})"
	     onmouseup="send_event('{{.IdComponent}}', 'EndDrag', '{{.ID}}')">
	</div>
	
	<div class="box-content">
		<div class="box-header">
			{{if .Icon}}
				<div class="box-icon">{{.Icon}}</div>
			{{end}}
			<div class="box-label">{{.Label}}</div>
		</div>
		{{if .Description}}
			<div class="box-description">{{.Description}}</div>
		{{end}}
	</div>
	
	<div class="box-ports">
		{{range .Ports}}
			<div class="port {{.Type}} {{.Position}} {{if .Connected}}connected{{end}}"
			     title="{{.Label}}"
			     onclick="event.stopPropagation(); send_event('{{$.IdComponent}}', 'ConnectPort', {boxId: '{{$.ID}}', portId: '{{.ID}}'})">
			</div>
		{{end}}
	</div>
</div>
`
}

func (b *FlowBox) GetDriver() liveview.LiveDriver {
	return b
}

func (b *FlowBox) HandleClick(data interface{}) {
	b.Selected = !b.Selected

	if b.OnClick != nil {
		b.OnClick(b.ID)
	}

	if b.ComponentDriver != nil {
		b.Commit()
	}
}

func (b *FlowBox) StartDrag(data interface{}) {
	b.Dragging = true

	if b.ComponentDriver != nil {
		b.Commit()
	}
}

func (b *FlowBox) HandleDrag(data interface{}) {
	if !b.Dragging {
		return
	}

	dragData, ok := data.(map[string]interface{})
	if !ok {
		return
	}
	
	// Get new position from the data - these should be relative positions
	if newX, ok := dragData["x"].(float64); ok {
		b.X = int(newX)
	}
	if newY, ok := dragData["y"].(float64); ok {
		b.Y = int(newY)
	}

	// Constrain to canvas bounds
	if b.X < 0 {
		b.X = 0
	}
	if b.Y < 0 {
		b.Y = 0
	}
	// Constrain to right and bottom edges (assuming max canvas width/height)
	maxX := 1200 - b.Width
	maxY := 600 - b.Height
	if b.X > maxX {
		b.X = maxX
	}
	if b.Y > maxY {
		b.Y = maxY
	}

	if b.ComponentDriver != nil {
		b.Commit()
	}
}

func (b *FlowBox) EndDrag(data interface{}) {
	b.Dragging = false

	if b.ComponentDriver != nil {
		b.Commit()
	}
}

func (b *FlowBox) HandlePortConnect(data interface{}) {
	portData := data.(map[string]interface{})
	portID := portData["portId"].(string)

	// Find and toggle port connection
	for i := range b.Ports {
		if b.Ports[i].ID == portID {
			b.Ports[i].Connected = !b.Ports[i].Connected
			break
		}
	}

	if b.OnConnect != nil {
		b.OnConnect(b.ID, portID, "", "")
	}

	if b.ComponentDriver != nil {
		b.Commit()
	}
}

func (b *FlowBox) HandlePortDisconnect(data interface{}) {
	portID := data.(string)

	for i := range b.Ports {
		if b.Ports[i].ID == portID {
			b.Ports[i].Connected = false
			break
		}
	}

	if b.ComponentDriver != nil {
		b.Commit()
	}
}

func (b *FlowBox) MoveTo(x, y int) {
	b.X = x
	b.Y = y
	if b.ComponentDriver != nil {
		b.Commit()
	}
}

func (b *FlowBox) Resize(width, height int) {
	b.Width = width
	b.Height = height
	if b.ComponentDriver != nil {
		b.Commit()
	}
}

func (b *FlowBox) SetSelected(selected bool) {
	b.Selected = selected
	if b.ComponentDriver != nil {
		b.Commit()
	}
}

func (b *FlowBox) UpdateLabel(label string) {
	b.Label = label
	if b.ComponentDriver != nil {
		b.Commit()
	}
}

func (b *FlowBox) UpdateDescription(description string) {
	b.Description = description
	if b.ComponentDriver != nil {
		b.Commit()
	}
}

func (b *FlowBox) AddPort(port Port) {
	b.Ports = append(b.Ports, port)
	if b.ComponentDriver != nil {
		b.Commit()
	}
}

func (b *FlowBox) RemovePort(portID string) {
	newPorts := []Port{}
	for _, p := range b.Ports {
		if p.ID != portID {
			newPorts = append(newPorts, p)
		}
	}
	b.Ports = newPorts
	if b.ComponentDriver != nil {
		b.Commit()
	}
}

func getBoxColor(boxType BoxType) string {
	switch boxType {
	case BoxTypeStart:
		return "#dcfce7" // green-100
	case BoxTypeEnd:
		return "#fee2e2" // red-100
	case BoxTypeProcess:
		return "#dbeafe" // blue-100
	case BoxTypeDecision:
		return "#fef3c7" // yellow-100
	case BoxTypeData:
		return "#e9d5ff" // purple-100
	default:
		return "#f3f4f6" // gray-100
	}
}

func generatePorts(boxType BoxType) []Port {
	switch boxType {
	case BoxTypeStart:
		return []Port{
			{ID: "out1", Type: "output", Position: "right", Label: "Start"},
		}
	case BoxTypeEnd:
		return []Port{
			{ID: "in1", Type: "input", Position: "left", Label: "End"},
		}
	case BoxTypeDecision:
		return []Port{
			{ID: "in1", Type: "input", Position: "left", Label: "Input"},
			{ID: "out1", Type: "output", Position: "top", Label: "True"},
			{ID: "out2", Type: "output", Position: "bottom", Label: "False"},
		}
	case BoxTypeProcess:
		return []Port{
			{ID: "in1", Type: "input", Position: "left", Label: "Input"},
			{ID: "out1", Type: "output", Position: "right", Label: "Output"},
		}
	case BoxTypeData:
		return []Port{
			{ID: "in1", Type: "input", Position: "top", Label: "Write"},
			{ID: "out1", Type: "output", Position: "bottom", Label: "Read"},
		}
	default:
		return []Port{
			{ID: "in1", Type: "input", Position: "left", Label: "In"},
			{ID: "out1", Type: "output", Position: "right", Label: "Out"},
		}
	}
}
