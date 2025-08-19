package components

import (
	"github.com/arturoeanton/go-echo-live-view/liveview"
)

type Card struct {
	*liveview.ComponentDriver[*Card]
	Title       string
	Subtitle    string
	Content     string
	ImageURL    string
	Footer      string
	Actions     []CardAction
	Width       string
	Elevated    bool
	Hoverable   bool
}

type CardAction struct {
	ID    string
	Label string
	Style string
}

func NewCard(title, content string) *Card {
	return &Card{
		Title:     title,
		Content:   content,
		Width:     "100%",
		Elevated:  true,
		Hoverable: false,
	}
}

func (c *Card) Start() {
	c.Commit()
}

func (c *Card) Action(data interface{}) {
	// Action handling would be implemented here
	// You could emit custom events or handle actions as needed
}

func (c *Card) GetDriver() liveview.LiveDriver {
	return c
}

func (c *Card) GetTemplate() string {
	return `
<div id="{{.IdComponent}}" class="card" style="width: {{.Width}}; background: white; border-radius: 8px; overflow: hidden; {{if .Elevated}}box-shadow: 0 4px 6px rgba(0,0,0,0.1);{{else}}box-shadow: 0 2px 4px rgba(0,0,0,0.1);{{end}}">
	{{if .ImageURL}}
	<div class="card-image" style="width: 100%; height: 200px; overflow: hidden;">
		<img src="{{.ImageURL}}" style="width: 100%; height: 100%; object-fit: cover;" alt="Card image">
	</div>
	{{end}}
	
	<div class="card-body" style="padding: 16px;">
		<h3 style="margin: 0 0 8px 0; font-size: 20px; font-weight: 500;">{{.Title}}</h3>
		{{if .Subtitle}}
		<p style="color: #666; margin: 4px 0 16px 0; font-size: 14px;">{{.Subtitle}}</p>
		{{end}}
		<div style="color: #333;">{{.Content}}</div>
	</div>
	
	{{if .Actions}}
	<div class="card-actions" style="padding: 16px; border-top: 1px solid #e0e0e0;">
		{{range .Actions}}
		<button style="padding: 8px 16px; border: none; border-radius: 4px; cursor: pointer; margin-right: 8px; {{if .Style}}{{.Style}}{{else}}background: #007bff; color: white;{{end}}"
				onclick="send_event('{{$.IdComponent}}', 'Action', JSON.stringify({action_id:'{{.ID}}'}))">
			{{.Label}}
		</button>
		{{end}}
	</div>
	{{end}}
	
	{{if .Footer}}
	<div class="card-footer" style="padding: 12px 16px; background: #f8f9fa; border-top: 1px solid #e0e0e0; font-size: 14px; color: #666;">
		{{.Footer}}
	</div>
	{{end}}
</div>
`
}