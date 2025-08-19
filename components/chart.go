package components

import (
	"fmt"
	"github.com/arturoeanton/go-echo-live-view/liveview"
)

type ChartType string

const (
	ChartBar    ChartType = "bar"
	ChartLine   ChartType = "line"
	ChartPie    ChartType = "pie"
	ChartDonut  ChartType = "donut"
)

type ChartData struct {
	Label string
	Value float64
	Color string
}

type Chart struct {
	*liveview.ComponentDriver[*Chart]
	Type   ChartType
	Data   []ChartData
	Width  int
	Height int
	Title  string
}

func (c *Chart) Start() {
	if c.Width == 0 {
		c.Width = 400
	}
	if c.Height == 0 {
		c.Height = 300
	}
	c.Commit()
}

func (c *Chart) GetTemplate() string {
	return `
	<div id="{{.IdComponent}}" class="chart-container">
		<style>
			.chart-container { padding: 1rem; background: white; border-radius: 8px; box-shadow: 0 2px 8px rgba(0,0,0,0.1); }
			.chart-title { font-size: 1.25rem; font-weight: 600; margin-bottom: 1rem; text-align: center; }
			.chart-svg { width: 100%; height: auto; }
			.chart-bar { transition: opacity 0.3s; cursor: pointer; }
			.chart-bar:hover { opacity: 0.8; }
			.chart-legend { display: flex; flex-wrap: wrap; gap: 1rem; margin-top: 1rem; justify-content: center; }
			.legend-item { display: flex; align-items: center; gap: 0.5rem; font-size: 0.875rem; }
			.legend-color { width: 16px; height: 16px; border-radius: 2px; }
		</style>
		
		{{if .Title}}<div class="chart-title">{{.Title}}</div>{{end}}
		
		{{if eq .Type "bar"}}
		<svg class="chart-svg" viewBox="0 0 {{.Width}} {{.Height}}">
			{{$maxValue := .GetMaxValue}}
			{{$barWidth := .GetBarWidth}}
			{{range $i, $d := .Data}}
			<rect class="chart-bar"
				x="{{$.GetBarX $i}}"
				y="{{$.GetBarY $d.Value $maxValue}}"
				width="{{$barWidth}}"
				height="{{$.GetBarHeight $d.Value $maxValue}}"
				fill="{{if $d.Color}}{{$d.Color}}{{else}}#4CAF50{{end}}"
				onclick="send_event('{{$.IdComponent}}', 'BarClick', {{$i}})"
			/>
			<text x="{{$.GetBarLabelX $i}}" y="{{$.Height}}" 
				text-anchor="middle" font-size="12" dy="-5">{{$d.Label}}</text>
			{{end}}
		</svg>
		{{end}}
		
		{{if eq .Type "pie"}}
		<svg class="chart-svg" viewBox="0 0 {{.Width}} {{.Height}}">
			{{$total := .GetTotal}}
			{{$centerX := .GetCenterX}}
			{{$centerY := .GetCenterY}}
			{{$radius := .GetRadius}}
			{{range $i, $d := .Data}}
			<circle cx="{{$centerX}}" cy="{{$centerY}}" r="{{$radius}}" 
				fill="{{if $d.Color}}{{$d.Color}}{{else}}{{$.GetDefaultColor $i}}{{end}}"
				stroke="white" stroke-width="2"
				class="chart-bar"
				onclick="send_event('{{$.IdComponent}}', 'SliceClick', {{$i}})"
			/>
			{{end}}
		</svg>
		{{end}}
		
		<div class="chart-legend">
			{{range $i, $d := .Data}}
			<div class="legend-item">
				<div class="legend-color" style="background: {{if $d.Color}}{{$d.Color}}{{else}}{{$.GetDefaultColor $i}}{{end}}"></div>
				<span>{{$d.Label}}: {{$d.Value}}</span>
			</div>
			{{end}}
		</div>
	</div>
	`
}

func (c *Chart) GetDriver() liveview.LiveDriver {
	return c
}

func (c *Chart) BarClick(data interface{}) {
	index := 0
	fmt.Sscanf(fmt.Sprint(data), "%d", &index)
	if index < len(c.Data) {
		c.EvalScript(fmt.Sprintf("console.log('Clicked bar: %s, value: %f')", c.Data[index].Label, c.Data[index].Value))
	}
}

func (c *Chart) SliceClick(data interface{}) {
	index := 0
	fmt.Sscanf(fmt.Sprint(data), "%d", &index)
	if index < len(c.Data) {
		c.EvalScript(fmt.Sprintf("console.log('Clicked slice: %s, value: %f')", c.Data[index].Label, c.Data[index].Value))
	}
}

func (c *Chart) GetMaxValue() float64 {
	max := 0.0
	for _, d := range c.Data {
		if d.Value > max {
			max = d.Value
		}
	}
	return max * 1.1
}

func (c *Chart) GetBarWidth() int {
	if len(c.Data) == 0 {
		return 0
	}
	return c.Width / len(c.Data) * 60 / 100
}

func (c *Chart) GetBarX(index int) int {
	if len(c.Data) == 0 {
		return 0
	}
	spacing := c.Width / len(c.Data)
	return index*spacing + spacing/5
}

func (c *Chart) GetBarY(value, maxValue float64) int {
	if maxValue == 0 {
		return c.Height - 30
	}
	return int(float64(c.Height-30) * (1 - value/maxValue))
}

func (c *Chart) GetBarHeight(value, maxValue float64) int {
	if maxValue == 0 {
		return 0
	}
	return int(float64(c.Height-30) * value / maxValue)
}

func (c *Chart) GetBarLabelX(index int) int {
	if len(c.Data) == 0 {
		return 0
	}
	spacing := c.Width / len(c.Data)
	return index*spacing + spacing/2
}

func (c *Chart) GetTotal() float64 {
	total := 0.0
	for _, d := range c.Data {
		total += d.Value
	}
	return total
}

func (c *Chart) GetCenterX() int {
	return c.Width / 2
}

func (c *Chart) GetCenterY() int {
	return c.Height / 2
}

func (c *Chart) GetRadius() int {
	min := c.Width
	if c.Height < min {
		min = c.Height
	}
	return min/2 - 20
}

func (c *Chart) GetAngle(value, total float64) float64 {
	if total == 0 {
		return 0
	}
	return value / total * 360
}

func (c *Chart) GetPieSlicePath(cx, cy, radius int, startAngle, sweepAngle float64) string {
	return fmt.Sprintf("M %d %d L %d %d A %d %d 0 %d 1 %d %d Z",
		cx, cy, cx, cy-radius, radius, radius, 0, cx, cy-radius)
}

func (c *Chart) AddAngles(a1, a2 float64) float64 {
	return a1 + a2
}

func (c *Chart) GetDefaultColor(index int) string {
	colors := []string{"#4CAF50", "#2196F3", "#FF9800", "#F44336", "#9C27B0", "#00BCD4", "#FFC107", "#795548"}
	return colors[index%len(colors)]
}

func (c *Chart) UpdateData(data []ChartData) {
	c.Data = data
	c.Commit()
}

func (c *Chart) AddData(label string, value float64, color string) {
	c.Data = append(c.Data, ChartData{Label: label, Value: value, Color: color})
	c.Commit()
}

func (c *Chart) ClearData() {
	c.Data = []ChartData{}
	c.Commit()
}