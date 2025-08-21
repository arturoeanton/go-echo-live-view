package components

import (
	"fmt"
	"math"
	"github.com/arturoeanton/go-echo-live-view/liveview"
)

type EdgeType string

const (
	EdgeTypeStraight EdgeType = "straight"
	EdgeTypeCurved   EdgeType = "curved"
	EdgeTypeStep     EdgeType = "step"
	EdgeTypeBezier   EdgeType = "bezier"
)

type EdgeStyle string

const (
	EdgeStyleSolid  EdgeStyle = "solid"
	EdgeStyleDashed EdgeStyle = "dashed"
	EdgeStyleDotted EdgeStyle = "dotted"
)

type FlowEdge struct {
	*liveview.ComponentDriver[*FlowEdge]
	
	ID          string
	FromBox     string
	FromPort    string
	ToBox       string
	ToPort      string
	FromX       int
	FromY       int
	ToX         int
	ToY         int
	Type        EdgeType
	Style       EdgeStyle
	Color       string
	Width       int
	Label       string
	Selected    bool
	Animated    bool
	ArrowHead   bool
	Bidirectional bool
	OnClick     func(edgeID string)
}

func NewFlowEdge(id, fromBox, fromPort, toBox, toPort string) *FlowEdge {
	return &FlowEdge{
		ID:        id,
		FromBox:   fromBox,
		FromPort:  fromPort,
		ToBox:     toBox,
		ToPort:    toPort,
		Type:      EdgeTypeCurved,
		Style:     EdgeStyleSolid,
		Color:     "#64748b",
		Width:     2,
		ArrowHead: true,
		Animated:  false,
	}
}

func (e *FlowEdge) Start() {
	// Events are registered directly on the ComponentDriver
	if e.ComponentDriver != nil {
		e.ComponentDriver.Events["Click"] = func(c *FlowEdge, data interface{}) {
			c.HandleClick(data)
		}
		e.ComponentDriver.Events["Remove"] = func(c *FlowEdge, data interface{}) {
			c.HandleRemove(data)
		}
	}
}

func (e *FlowEdge) GetTemplate() string {
	return `
<svg class="flow-edge" id="edge-{{.ID}}" 
     style="position: absolute; pointer-events: none; overflow: visible; z-index: {{if .Selected}}15{{else}}5{{end}};">
	
	<defs>
		<marker id="arrowhead-{{.ID}}" 
		        markerWidth="10" 
		        markerHeight="10" 
		        refX="9" 
		        refY="3" 
		        orient="auto">
			<polygon points="0 0, 10 3, 0 6" fill="{{.Color}}" />
		</marker>
		
		{{if .Bidirectional}}
		<marker id="arrowhead-reverse-{{.ID}}" 
		        markerWidth="10" 
		        markerHeight="10" 
		        refX="1" 
		        refY="3" 
		        orient="auto">
			<polygon points="10 0, 0 3, 10 6" fill="{{.Color}}" />
		</marker>
		{{end}}
		
		{{if .Animated}}
		<style>
			@keyframes dash-{{.ID}} {
				to {
					stroke-dashoffset: -20;
				}
			}
			.animated-edge-{{.ID}} {
				animation: dash-{{.ID}} 1s linear infinite;
			}
		</style>
		{{end}}
	</defs>
	
	<style>
		.flow-edge {
			width: 100%;
			height: 100%;
			left: 0;
			top: 0;
		}
		
		.edge-path {
			fill: none;
			stroke: {{.Color}};
			stroke-width: {{.Width}}px;
			cursor: pointer;
			pointer-events: stroke;
			transition: stroke 0.2s, stroke-width 0.2s;
			{{if eq .Style "dashed"}}
			stroke-dasharray: 10, 5;
			{{else if eq .Style "dotted"}}
			stroke-dasharray: 2, 3;
			{{end}}
		}
		
		.edge-path:hover {
			stroke-width: 3px;
			stroke: {{if .Selected}}#2563eb{{else}}#475569{{end}};
		}
		
		.edge-path.selected {
			stroke: #2563eb;
			stroke-width: 3px;
		}
		
		.edge-hitbox {
			fill: none;
			stroke: transparent;
			stroke-width: 20px;
			cursor: pointer;
			pointer-events: stroke;
		}
		
		.edge-label {
			fill: #1f2937;
			font-size: 12px;
			font-weight: 500;
			text-anchor: middle;
			pointer-events: none;
			background: white;
			padding: 2px 4px;
		}
		
		.edge-label-bg {
			fill: white;
			stroke: #e5e7eb;
			stroke-width: 1px;
			rx: 3;
		}
		
		.edge-remove {
			fill: #ef4444;
			cursor: pointer;
			opacity: 0;
			transition: opacity 0.2s;
		}
		
		.flow-edge:hover .edge-remove {
			opacity: 1;
		}
	</style>
	
	{{if eq .Type "straight"}}
		<path class="edge-hitbox"
		      d="M {{.FromX}} {{.FromY}} L {{.ToX}} {{.ToY}}"
		      onclick="send_event('{{.IdComponent}}', 'Click', '{{.ID}}')" />
		      
		<path class="edge-path {{if .Selected}}selected{{end}} {{if .Animated}}animated-edge-{{.ID}}{{end}}"
		      d="M {{.FromX}} {{.FromY}} L {{.ToX}} {{.ToY}}"
		      {{if .ArrowHead}}marker-end="url(#arrowhead-{{.ID}})"{{end}}
		      {{if .Bidirectional}}marker-start="url(#arrowhead-reverse-{{.ID}})"{{end}} />
	{{else if eq .Type "curved"}}
		{{$path := .CalculateCurvedPath}}
		<path class="edge-hitbox"
		      d="{{$path}}"
		      onclick="send_event('{{.IdComponent}}', 'Click', '{{.ID}}')" />
		      
		<path class="edge-path {{if .Selected}}selected{{end}} {{if .Animated}}animated-edge-{{.ID}}{{end}}"
		      d="{{$path}}"
		      {{if .ArrowHead}}marker-end="url(#arrowhead-{{.ID}})"{{end}}
		      {{if .Bidirectional}}marker-start="url(#arrowhead-reverse-{{.ID}})"{{end}} />
	{{else if eq .Type "step"}}
		{{$path := .CalculateStepPath}}
		<path class="edge-hitbox"
		      d="{{$path}}"
		      onclick="send_event('{{.IdComponent}}', 'Click', '{{.ID}}')" />
		      
		<path class="edge-path {{if .Selected}}selected{{end}} {{if .Animated}}animated-edge-{{.ID}}{{end}}"
		      d="{{$path}}"
		      {{if .ArrowHead}}marker-end="url(#arrowhead-{{.ID}})"{{end}}
		      {{if .Bidirectional}}marker-start="url(#arrowhead-reverse-{{.ID}})"{{end}} />
	{{else}}
		{{$path := .CalculateBezierPath}}
		<path class="edge-hitbox"
		      d="{{$path}}"
		      onclick="send_event('{{.IdComponent}}', 'Click', '{{.ID}}')" />
		      
		<path class="edge-path {{if .Selected}}selected{{end}} {{if .Animated}}animated-edge-{{.ID}}{{end}}"
		      d="{{$path}}"
		      {{if .ArrowHead}}marker-end="url(#arrowhead-{{.ID}})"{{end}}
		      {{if .Bidirectional}}marker-start="url(#arrowhead-reverse-{{.ID}})"{{end}} />
	{{end}}
	
	{{if .Label}}
		{{$midX := .GetMidX}}
		{{$midY := .GetMidY}}
		<rect class="edge-label-bg"
		      x="{{.GetLabelX}}"
		      y="{{.GetLabelY}}"
		      width="60"
		      height="20" />
		<text class="edge-label"
		      x="{{$midX}}"
		      y="{{.GetLabelTextY}}">
			{{.Label}}
		</text>
	{{end}}
	
	{{if .Selected}}
		{{$midX := .GetMidX}}
		{{$midY := .GetMidY}}
		<circle class="edge-remove"
		        cx="{{$midX}}"
		        cy="{{.GetRemoveY}}"
		        r="8"
		        onclick="send_event('{{.IdComponent}}', 'Remove', '{{.ID}}')" />
		<text x="{{$midX}}"
		      y="{{.GetRemoveTextY}}"
		      text-anchor="middle"
		      fill="white"
		      font-size="12"
		      font-weight="bold"
		      pointer-events="none">×</text>
	{{end}}
</svg>
`
}

func (e *FlowEdge) GetDriver() liveview.LiveDriver {
	return e
}

func (e *FlowEdge) HandleClick(data interface{}) {
	e.Selected = !e.Selected
	
	if e.OnClick != nil {
		e.OnClick(e.ID)
	}
	
	if e.ComponentDriver != nil {
		e.Commit()
	}
}

func (e *FlowEdge) HandleRemove(data interface{}) {
	// This would typically trigger removal from parent canvas
	if e.ComponentDriver != nil {
		e.Commit()
	}
}

func (e *FlowEdge) UpdatePosition(fromX, fromY, toX, toY int) {
	e.FromX = fromX
	e.FromY = fromY
	e.ToX = toX
	e.ToY = toY
	if e.ComponentDriver != nil {
		e.Commit()
	}
}

func (e *FlowEdge) SetEdgeStyle(style EdgeStyle) {
	e.Style = style
	if e.ComponentDriver != nil {
		e.Commit()
	}
}

func (e *FlowEdge) SetType(edgeType EdgeType) {
	e.Type = edgeType
	if e.ComponentDriver != nil {
		e.Commit()
	}
}

func (e *FlowEdge) SetAnimated(animated bool) {
	e.Animated = animated
	if e.ComponentDriver != nil {
		e.Commit()
	}
}

func (e *FlowEdge) SetSelected(selected bool) {
	e.Selected = selected
	if e.ComponentDriver != nil {
		e.Commit()
	}
}

func (e *FlowEdge) CalculateCurvedPath() string {
	dx := e.ToX - e.FromX
	
	// Control point offset
	offset := int(math.Min(math.Abs(float64(dx))/2, 50))
	
	cx1 := e.FromX + offset
	cy1 := e.FromY
	cx2 := e.ToX - offset
	cy2 := e.ToY
	
	return fmt.Sprintf("M %d %d C %d %d, %d %d, %d %d",
		e.FromX, e.FromY,
		cx1, cy1,
		cx2, cy2,
		e.ToX, e.ToY)
}

func (e *FlowEdge) CalculateStepPath() string {
	midX := (e.FromX + e.ToX) / 2
	
	return fmt.Sprintf("M %d %d L %d %d L %d %d L %d %d",
		e.FromX, e.FromY,
		midX, e.FromY,
		midX, e.ToY,
		e.ToX, e.ToY)
}

func (e *FlowEdge) CalculateBezierPath() string {
	dx := e.ToX - e.FromX
	dy := e.ToY - e.FromY
	
	// More pronounced curve for bezier
	offset := int(math.Max(math.Abs(float64(dx))/3, math.Abs(float64(dy))/3))
	
	cx1 := e.FromX + offset
	cy1 := e.FromY - offset/2
	cx2 := e.ToX - offset
	cy2 := e.ToY + offset/2
	
	return fmt.Sprintf("M %d %d C %d %d, %d %d, %d %d",
		e.FromX, e.FromY,
		cx1, cy1,
		cx2, cy2,
		e.ToX, e.ToY)
}

func (e *FlowEdge) GetMidX() int {
	return (e.FromX + e.ToX) / 2
}

func (e *FlowEdge) GetMidY() int {
	return (e.FromY + e.ToY) / 2
}

func (e *FlowEdge) GetLength() float64 {
	dx := float64(e.ToX - e.FromX)
	dy := float64(e.ToY - e.FromY)
	return math.Sqrt(dx*dx + dy*dy)
}

// Helper methods for template positioning
func (e *FlowEdge) GetLabelX() int {
	return e.GetMidX() - 30
}

func (e *FlowEdge) GetLabelY() int {
	return e.GetMidY() - 10
}

func (e *FlowEdge) GetLabelTextY() int {
	return e.GetMidY() + 3
}

func (e *FlowEdge) GetRemoveY() int {
	return e.GetMidY() - 20
}

func (e *FlowEdge) GetRemoveTextY() int {
	return e.GetMidY() - 16
}