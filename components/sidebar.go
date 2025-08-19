package components

import (
	"fmt"
	"github.com/arturoeanton/go-echo-live-view/liveview"
)

type SidebarItem struct {
	ID       string
	Label    string
	Icon     string
	Active   bool
	Children []SidebarItem
	Expanded bool
}

type Sidebar struct {
	*liveview.ComponentDriver[*Sidebar]
	Items      []SidebarItem
	Collapsed  bool
	Width      string
	Background string
}

func NewSidebar(items []SidebarItem) *Sidebar {
	return &Sidebar{
		Items:      items,
		Collapsed:  false,
		Width:      "250px",
		Background: "#2c3e50",
	}
}

func (s *Sidebar) Start() {
	s.Commit()
}

func (s *Sidebar) Toggle(data interface{}) {
	s.Collapsed = !s.Collapsed
	if s.Collapsed {
		s.Width = "60px"
	} else {
		s.Width = "250px"
	}
	s.Commit()
}

func (s *Sidebar) Select(data interface{}) {
	// Handle both string and map data types
	var itemID string
	switch v := data.(type) {
	case string:
		itemID = v
	case map[string]interface{}:
		if id, ok := v["item_id"].(string); ok {
			itemID = id
		}
	default:
		return
	}
	
	if itemID != "" {
		// Check if item has children and expand/collapse it
		if s.hasChildren(itemID, s.Items) {
			s.toggleExpanded(itemID, s.Items)
		} else {
			// Only set active for leaf items
			s.setActiveItem(itemID, s.Items)
		}
		s.Commit()
	}
}

func (s *Sidebar) Expand(data interface{}) {
	// Handle both string and map data types
	var itemID string
	switch v := data.(type) {
	case string:
		itemID = v
	case map[string]interface{}:
		if id, ok := v["item_id"].(string); ok {
			itemID = id
		}
	default:
		return
	}
	
	if itemID != "" {
		s.toggleExpanded(itemID, s.Items)
		s.Commit()
	}
}

func (s *Sidebar) GetDriver() liveview.LiveDriver {
	return s
}

func (s *Sidebar) setActiveItem(id string, items []SidebarItem) bool {
	for i := range items {
		items[i].Active = items[i].ID == id
		if items[i].Active {
			return true
		}
		if s.setActiveItem(id, items[i].Children) {
			return true
		}
	}
	return false
}

func (s *Sidebar) toggleExpanded(id string, items []SidebarItem) bool {
	for i := range items {
		if items[i].ID == id {
			items[i].Expanded = !items[i].Expanded
			return true
		}
		if s.toggleExpanded(id, items[i].Children) {
			return true
		}
	}
	return false
}

func (s *Sidebar) hasChildren(id string, items []SidebarItem) bool {
	for _, item := range items {
		if item.ID == id && len(item.Children) > 0 {
			return true
		}
		if s.hasChildren(id, item.Children) {
			return true
		}
	}
	return false
}

func (s *Sidebar) renderItems(items []SidebarItem, level int) string {
	html := ""
	for _, item := range items {
		paddingLeft := 16 + (level * 20)
		bgColor := ""
		if item.Active {
			bgColor = "background: rgba(255,255,255,0.1);"
		}
		
		icon := item.Icon
		if icon == "" {
			icon = "📄"
		}
		
		label := ""
		if !s.Collapsed {
			label = fmt.Sprintf(`<span style="margin-left: 12px;">%s</span>`, item.Label)
		}
		
		html += fmt.Sprintf(`
			<div class="sidebar-item" style="padding: 10px %dpx; color: white; cursor: pointer; transition: all 0.3s; %s display: flex; align-items: center; white-space: nowrap; overflow: hidden;"
				 onmouseover="this.style.background='rgba(255,255,255,0.05)'" 
				 onmouseout="this.style.background='%s'"
				 onclick="send_event('{{.IdComponent}}', 'Select', '%s')">
				<span style="font-size: 18px;">%s</span>
				%s
			`, paddingLeft, bgColor, bgColor, item.ID, icon, label)
		
		if len(item.Children) > 0 && !s.Collapsed {
			chevron := "▶"
			if item.Expanded {
				chevron = "▼"
			}
			html += fmt.Sprintf(`
				<span style="margin-left: auto; margin-right: 8px; cursor: pointer;" 
					  onclick="event.stopPropagation(); send_event('{{.IdComponent}}', 'Expand', '%s')">%s</span>
			`, item.ID, chevron)
		}
		
		html += `</div>`
		
		if item.Expanded && len(item.Children) > 0 && !s.Collapsed {
			html += s.renderItems(item.Children, level+1)
		}
	}
	return html
}

func (s *Sidebar) GetTemplate() string {
	toggleIcon := "☰"
	
	return fmt.Sprintf(`
		<div id="{{.IdComponent}}" class="sidebar" style="width: %s; height: 100vh; background: %s; transition: width 0.3s; overflow-y: auto; overflow-x: hidden; position: relative;">
			<div class="sidebar-header" style="padding: 16px; border-bottom: 1px solid rgba(255,255,255,0.1); display: flex; justify-content: space-between; align-items: center;">
				<h3 style="color: white; margin: 0; display: %s;">Menu</h3>
				<button style="background: none; border: none; color: white; font-size: 20px; cursor: pointer; padding: 4px 8px;"
						onclick="send_event('{{.IdComponent}}', 'Toggle')">%s</button>
			</div>
			<div class="sidebar-content">
				%s
			</div>
		</div>
	`, s.Width, s.Background, map[bool]string{true: "none", false: "block"}[s.Collapsed], toggleIcon, s.renderItems(s.Items, 0))
}