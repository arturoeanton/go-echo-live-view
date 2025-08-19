package components

import (
	"github.com/arturoeanton/go-echo-live-view/liveview"
)

type AccordionItem struct {
	ID       string
	Title    string
	Content  string
	Expanded bool
}

type Accordion struct {
	*liveview.ComponentDriver[*Accordion]
	Items []AccordionItem
}

func NewAccordion(items []AccordionItem) *Accordion {
	return &Accordion{
		Items: items,
	}
}

func (a *Accordion) Start() {
	a.Commit()
}

func (a *Accordion) GetDriver() liveview.LiveDriver {
	return a
}

func (a *Accordion) Toggle(data interface{}) {
	if itemID, ok := data.(string); ok {
		for i := range a.Items {
			if a.Items[i].ID == itemID {
				a.Items[i].Expanded = !a.Items[i].Expanded
			}
		}
		a.Commit()
	}
}

func (a *Accordion) GetTemplate() string {
	return `
<div id="{{.IdComponent}}" class="accordion" style="width: 100%; border: 1px solid #ddd; border-radius: 8px; overflow: hidden;">
	{{range $i, $item := .Items}}
	<div class="accordion-item" style="{{if gt $i 0}}border-top: 1px solid #ddd;{{end}}">
		<div class="accordion-header" 
			 style="padding: 12px 16px; background: #f8f9fa; cursor: pointer; display: flex; justify-content: space-between; align-items: center; user-select: none; transition: background 0.2s;"
			 onmouseover="this.style.background='#e9ecef'" 
			 onmouseout="this.style.background='#f8f9fa'"
			 onclick="send_event('{{$.IdComponent}}', 'Toggle', '{{$item.ID}}')">
			<span style="font-weight: 500;">{{$item.Title}}</span>
			<span style="transition: transform 0.3s;">{{if $item.Expanded}}▼{{else}}▶{{end}}</span>
		</div>
		<div class="accordion-content" style="display: {{if $item.Expanded}}block{{else}}none{{end}}; padding: 16px; background: white;">
			{{$item.Content}}
		</div>
	</div>
	{{end}}
</div>
`
}