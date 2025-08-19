package components

import (
	"fmt"
	"github.com/arturoeanton/go-echo-live-view/liveview"
	"strings"
)

type BreadcrumbItem struct {
	Label  string
	Href   string
	Active bool
	Icon   string
}

type Breadcrumb struct {
	*liveview.ComponentDriver[*Breadcrumb]
	Items     []BreadcrumbItem
	Separator string
}

func NewBreadcrumb(items []BreadcrumbItem) *Breadcrumb {
	return &Breadcrumb{
		Items:     items,
		Separator: "/",
	}
}

func (b *Breadcrumb) Start() {
	b.Commit()
}

func (b *Breadcrumb) Navigate(data interface{}) {
	// Navigation would be handled here
	// In a real implementation, you might emit a custom event
	// or handle navigation differently
}

func (b *Breadcrumb) GetDriver() liveview.LiveDriver {
	return b
}

func (b *Breadcrumb) AddItem(label, href string) {
	for i := range b.Items {
		b.Items[i].Active = false
	}
	
	b.Items = append(b.Items, BreadcrumbItem{
		Label:  label,
		Href:   href,
		Active: true,
	})
	b.Commit()
}

func (b *Breadcrumb) GetTemplate() string {
	if len(b.Items) == 0 {
		return ""
	}
	
	itemsHtml := []string{}
	
	for i, item := range b.Items {
		icon := ""
		if item.Icon != "" {
			icon = fmt.Sprintf(`<span style="margin-right: 4px;">%s</span>`, item.Icon)
		}
		
		if item.Active || i == len(b.Items)-1 {
			itemsHtml = append(itemsHtml, fmt.Sprintf(`
				<span class="breadcrumb-item active" style="color: #6c757d;">
					%s%s
				</span>
			`, icon, item.Label))
		} else {
			itemsHtml = append(itemsHtml, fmt.Sprintf(`
				<a href="#" class="breadcrumb-item" 
				   style="color: #007bff; text-decoration: none; transition: color 0.3s;"
				   onmouseover="this.style.color='#0056b3'" 
				   onmouseout="this.style.color='#007bff'"
				   onclick="event.preventDefault(); send_event('{{.IdComponent}}', 'Navigate', JSON.stringify({href:'%s'}))">
					%s%s
				</a>
			`, item.Href, icon, item.Label))
		}
	}
	
	separatorHtml := fmt.Sprintf(`<span style="margin: 0 8px; color: #6c757d;">%s</span>`, b.Separator)
	breadcrumbContent := strings.Join(itemsHtml, separatorHtml)
	
	return fmt.Sprintf(`
		<nav id="{{.IdComponent}}" aria-label="breadcrumb" style="padding: 8px 16px; background: #f8f9fa; border-radius: 4px;">
			<style>
				.breadcrumb-item:hover {
					text-decoration: underline !important;
				}
			</style>
			<div style="display: flex; align-items: center; flex-wrap: wrap; font-size: 14px;">
				%s
			</div>
		</nav>
	`, breadcrumbContent)
}