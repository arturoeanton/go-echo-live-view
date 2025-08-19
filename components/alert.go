package components

import (
	"fmt"
	"github.com/arturoeanton/go-echo-live-view/liveview"
)

type Alert struct {
	*liveview.ComponentDriver[*Alert]
	Message    string
	Type       string // info, success, warning, error
	Title      string
	Dismissible bool
	Visible    bool
	Icon       bool
}

func NewAlert(message string, alertType string) *Alert {
	return &Alert{
		Message:     message,
		Type:        alertType,
		Visible:     true,
		Dismissible: true,
		Icon:        true,
	}
}

func (a *Alert) Start() {
	a.Commit()
}

func (a *Alert) Dismiss(data interface{}) {
	a.Visible = false
	a.Commit()
}

func (a *Alert) GetDriver() liveview.LiveDriver {
	return a
}

func (a *Alert) Show(message string, alertType string) {
	a.Message = message
	a.Type = alertType
	a.Visible = true
	a.Commit()
}

func (a *Alert) getColors() (bg, border, text, icon string) {
	switch a.Type {
	case "success":
		return "#d4edda", "#c3e6cb", "#155724", "✓"
	case "warning":
		return "#fff3cd", "#ffeeba", "#856404", "⚠"
	case "error", "danger":
		return "#f8d7da", "#f5c6cb", "#721c24", "✕"
	default: // info
		return "#d1ecf1", "#bee5eb", "#0c5460", "ℹ"
	}
}

func (a *Alert) GetTemplate() string {
	if !a.Visible {
		return ""
	}
	
	bg, border, text, iconChar := a.getColors()
	
	titleHtml := ""
	if a.Title != "" {
		titleHtml = fmt.Sprintf(`<strong style="display: block; margin-bottom: 4px;">%s</strong>`, a.Title)
	}
	
	iconHtml := ""
	if a.Icon {
		iconHtml = fmt.Sprintf(`
			<span style="display: inline-flex; align-items: center; justify-content: center; 
						 width: 24px; height: 24px; border-radius: 50%%; background: %s; 
						 color: white; margin-right: 12px; flex-shrink: 0; font-weight: bold;">
				%s
			</span>
		`, text, iconChar)
	}
	
	dismissButton := ""
	if a.Dismissible {
		dismissButton = fmt.Sprintf(`
			<button style="background: none; border: none; color: %s; font-size: 20px; 
						   cursor: pointer; padding: 0; margin-left: auto; opacity: 0.5; 
						   transition: opacity 0.3s;"
					onmouseover="this.style.opacity='1'" 
					onmouseout="this.style.opacity='0.5'"
					onclick="send_event('{{.IdComponent}}', 'Dismiss')">
				×
			</button>
		`, text)
	}
	
	return fmt.Sprintf(`
		<div id="{{.IdComponent}}" class="alert alert-%s" role="alert" 
			 style="background: %s; border: 1px solid %s; color: %s; 
					padding: 12px 16px; border-radius: 4px; display: flex; 
					align-items: center; margin-bottom: 16px; 
					animation: slideDown 0.3s ease-out;">
			<style>
				@keyframes slideDown {
					from {
						opacity: 0;
						transform: translateY(-10px);
					}
					to {
						opacity: 1;
						transform: translateY(0);
					}
				}
			</style>
			%s
			<div style="flex: 1;">
				%s
				%s
			</div>
			%s
		</div>
	`, a.Type, bg, border, text, iconHtml, titleHtml, a.Message, dismissButton)
}