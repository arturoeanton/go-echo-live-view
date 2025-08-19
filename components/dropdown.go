package components

import (
	"github.com/arturoeanton/go-echo-live-view/liveview"
)

type DropdownOption struct {
	Value    string
	Label    string
	Disabled bool
	Icon     string
}

type Dropdown struct {
	*liveview.ComponentDriver[*Dropdown]
	Options      []DropdownOption
	Selected     string
	Placeholder  string
	IsOpen       bool
	Width        string
}

func NewDropdown(options []DropdownOption, placeholder string) *Dropdown {
	return &Dropdown{
		Options:     options,
		Placeholder: placeholder,
		IsOpen:      false,
		Width:       "200px",
	}
}

func (d *Dropdown) Start() {
	d.Commit()
}

func (d *Dropdown) Toggle(data interface{}) {
	d.IsOpen = !d.IsOpen
	d.Commit()
}

func (d *Dropdown) Select(data interface{}) {
	if value, ok := data.(string); ok {
		d.Selected = value
		d.IsOpen = false
		d.Commit()
	}
}

func (d *Dropdown) Close(data interface{}) {
	d.IsOpen = false
	d.Commit()
}

func (d *Dropdown) GetSelectedLabel() string {
	for _, opt := range d.Options {
		if opt.Value == d.Selected {
			return opt.Label
		}
	}
	return d.Placeholder
}

func (d *Dropdown) GetDriver() liveview.LiveDriver {
	return d
}

func (d *Dropdown) GetTemplate() string {
	return `
<div id="{{.IdComponent}}" class="dropdown" style="position: relative; width: {{.Width}};">
	<div class="dropdown-trigger" 
		 style="border: 1px solid #ddd; border-radius: 4px; padding: 8px 12px; 
				background: white; cursor: pointer; display: flex; justify-content: space-between; 
				align-items: center; user-select: none;"
		 onclick="send_event('{{.IdComponent}}', 'Toggle')">
		<span>{{.GetSelectedLabel}}</span>
		<span style="margin-left: 8px;">{{if .IsOpen}}▲{{else}}▼{{end}}</span>
	</div>
	{{if .IsOpen}}
	<div class="dropdown-menu" style="position: absolute; top: 100%; left: 0; right: 0; 
										background: white; border: 1px solid #ddd; border-top: none; 
										border-radius: 0 0 4px 4px; max-height: 300px; overflow-y: auto; 
										z-index: 1000; box-shadow: 0 4px 6px rgba(0,0,0,0.1);">
		{{range .Options}}
		<div class="dropdown-option" 
			 style="padding: 8px 12px; cursor: {{if .Disabled}}not-allowed{{else}}pointer{{end}}; 
					{{if .Disabled}}opacity: 0.5;{{end}}
					{{if eq .Value $.Selected}}background: #e3f2fd;{{end}}
					display: flex; align-items: center;"
			 {{if not .Disabled}}
			 onmouseover="if(!this.dataset.disabled) this.style.background='#f0f0f0'" 
			 onmouseout="this.style.background='{{if eq .Value $.Selected}}#e3f2fd{{else}}white{{end}}'"
			 onclick="send_event('{{$.IdComponent}}', 'Select', '{{.Value}}')"
			 {{end}}
			 data-disabled="{{.Disabled}}">
			{{if .Icon}}<span style="margin-right: 8px;">{{.Icon}}</span>{{end}}
			{{.Label}}
		</div>
		{{end}}
	</div>
	<div style="position: fixed; top: 0; left: 0; right: 0; bottom: 0; z-index: 999;" 
		 onclick="send_event('{{.IdComponent}}', 'Close')"></div>
	{{end}}
</div>
`
}