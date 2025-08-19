package components

import (
	"fmt"
	"github.com/arturoeanton/go-echo-live-view/liveview"
)

type HeaderMenuItem struct {
	ID    string
	Label string
	Href  string
}

type Header struct {
	*liveview.ComponentDriver[*Header]
	Title      string
	Logo       string
	MenuItems  []HeaderMenuItem
	Background string
	Height     string
	ShowMenu   bool
}

func NewHeader(title string, menuItems []HeaderMenuItem) *Header {
	return &Header{
		Title:      title,
		MenuItems:  menuItems,
		Background: "#2c3e50",
		Height:     "60px",
		ShowMenu:   false,
	}
}

func (h *Header) Start() {
	h.Commit()
}

func (h *Header) ToggleMenu(data interface{}) {
	h.ShowMenu = !h.ShowMenu
	h.Commit()
}

func (h *Header) Navigate(data interface{}) {
	// Handle navigation - this will be overridden by the parent component
	if h.Events != nil && h.Events["MenuClick"] != nil {
		// Pass the item ID directly
		h.Events["MenuClick"](h, data)
	}
}

func (h *Header) GetDriver() liveview.LiveDriver {
	return h
}

func (h *Header) GetTemplate() string {
	logoHtml := ""
	if h.Logo != "" {
		logoHtml = fmt.Sprintf(`<img src="%s" style="height: 40px; margin-right: 16px;" alt="Logo">`, h.Logo)
	}
	
	menuItemsHtml := ""
	for _, item := range h.MenuItems {
		menuItemsHtml += fmt.Sprintf(`
			<a href="#" style="color: white; text-decoration: none; padding: 8px 16px; margin: 0 4px; border-radius: 4px; transition: background 0.3s;"
			   onmouseover="this.style.background='rgba(255,255,255,0.1)'" 
			   onmouseout="this.style.background='transparent'"
			   onclick="event.preventDefault(); send_event('{{.IdComponent}}', 'Navigate', '%s')">
				%s
			</a>
		`, item.ID, item.Label)
	}
	
	mobileMenuHtml := ""
	if h.ShowMenu {
		mobileMenuItemsHtml := ""
		for _, item := range h.MenuItems {
			mobileMenuItemsHtml += fmt.Sprintf(`
				<a href="#" style="display: block; color: white; text-decoration: none; padding: 12px 16px; border-bottom: 1px solid rgba(255,255,255,0.1);"
				   onclick="event.preventDefault(); send_event('{{.IdComponent}}', 'Navigate', '%s')">
					%s
				</a>
			`, item.ID, item.Label)
		}
		
		mobileMenuHtml = fmt.Sprintf(`
			<div class="mobile-menu" style="position: absolute; top: %s; left: 0; right: 0; background: %s; box-shadow: 0 2px 10px rgba(0,0,0,0.1); z-index: 1000;">
				%s
			</div>
		`, h.Height, h.Background, mobileMenuItemsHtml)
	}
	
	return fmt.Sprintf(`
		<header id="{{.IdComponent}}" style="background: %s; height: %s; display: flex; align-items: center; padding: 0 24px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); position: relative;">
			<div style="display: flex; align-items: center; flex: 1;">
				%s
				<h1 style="color: white; margin: 0; font-size: 24px; font-weight: 500;">%s</h1>
			</div>
			
			<nav class="desktop-menu" style="display: flex; align-items: center;">
				<style>
					@media (max-width: 768px) {
						.desktop-menu { display: none !important; }
						.mobile-menu-btn { display: block !important; }
					}
					@media (min-width: 769px) {
						.mobile-menu { display: none !important; }
						.mobile-menu-btn { display: none !important; }
					}
				</style>
				%s
			</nav>
			
			<button class="mobile-menu-btn" style="display: none; background: none; border: none; color: white; font-size: 24px; cursor: pointer; padding: 8px;"
					onclick="send_event('{{.IdComponent}}', 'ToggleMenu')">
				☰
			</button>
			
			%s
		</header>
	`, h.Background, h.Height, logoHtml, h.Title, menuItemsHtml, mobileMenuHtml)
}