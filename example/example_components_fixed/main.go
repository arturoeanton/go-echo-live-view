package main

import (
	"github.com/arturoeanton/go-echo-live-view/components"
	"github.com/arturoeanton/go-echo-live-view/liveview"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type ComponentShowcase struct {
	*liveview.ComponentDriver[*ComponentShowcase]
	Alert      *components.Alert
	Accordion  *components.Accordion
	Dropdown   *components.Dropdown
	Card1      *components.Card
	Card2      *components.Card
	Breadcrumb *components.Breadcrumb
	Header     *components.Header
	Sidebar    *components.Sidebar
	
	// Buttons for alerts
	BtnInfo    *components.Button
	BtnSuccess *components.Button
	BtnWarning *components.Button
	BtnError   *components.Button
	
	// Additional components for different views
	Table      *components.Table
	Chart      *components.Chart
	Clock      *components.Clock
	
	// Current view
	CurrentView string
}

func NewComponentShowcase() *ComponentShowcase {
	return &ComponentShowcase{
		CurrentView: "dashboard",
	}
}

func (c *ComponentShowcase) Start() {
	// Initialize Alert
	alert := components.NewAlert("This is a success message! All components are working correctly.", "success")
	alert.Title = "Success!"
	c.Alert = liveview.New("alert", alert)
	
	// Initialize Alert Buttons with proper LiveView components
	c.BtnInfo = liveview.New("btn_info", &components.Button{Caption: "Show Info"}).
		SetClick(func(btn *components.Button, data interface{}) {
			c.Alert.Show("This is an informational message.", "info")
		})
		
	c.BtnSuccess = liveview.New("btn_success", &components.Button{Caption: "Show Success"}).
		SetClick(func(btn *components.Button, data interface{}) {
			c.Alert.Show("Operation completed successfully!", "success")
		})
		
	c.BtnWarning = liveview.New("btn_warning", &components.Button{Caption: "Show Warning"}).
		SetClick(func(btn *components.Button, data interface{}) {
			c.Alert.Show("Please review this warning message.", "warning")
		})
		
	c.BtnError = liveview.New("btn_error", &components.Button{Caption: "Show Error"}).
		SetClick(func(btn *components.Button, data interface{}) {
			c.Alert.Show("An error occurred. Please try again.", "error")
		})
	
	// Initialize Accordion
	accordion := components.NewAccordion([]components.AccordionItem{
		{ID: "1", Title: "What is Go Echo LiveView?", Content: "Go Echo LiveView is a framework for building reactive web applications using server-side rendering with Go and the Echo web framework.", Expanded: true},
		{ID: "2", Title: "How does it work?", Content: "It uses WebSockets to maintain a persistent connection between the server and client, allowing for real-time updates without full page reloads."},
		{ID: "3", Title: "What are the benefits?", Content: "You can build interactive web applications using only Go, without writing JavaScript. The framework handles all the client-server communication for you."},
	})
	c.Accordion = liveview.New("accordion", accordion)
	
	// Initialize Dropdown
	dropdown := components.NewDropdown([]components.DropdownOption{
		{Value: "go", Label: "Go", Icon: "🐹"},
		{Value: "python", Label: "Python", Icon: "🐍"},
		{Value: "javascript", Label: "JavaScript", Icon: "🟨"},
		{Value: "rust", Label: "Rust", Icon: "🦀"},
		{Value: "ruby", Label: "Ruby", Icon: "💎", Disabled: true},
	}, "Select a language")
	dropdown.Width = "250px"
	c.Dropdown = liveview.New("dropdown", dropdown)
	
	// Initialize Cards
	card1 := components.NewCard("Interactive Components", "This showcase demonstrates various UI components built with Go Echo LiveView. Each component is reactive and updates in real-time.")
	// card1.ImageURL = "" // Remove external image to avoid DNS errors
	card1.Width = "100%"
	card1.Hoverable = true
	card1.Actions = []components.CardAction{
		{ID: "learn", Label: "Learn More", Style: "background: #007bff; color: white;"},
		{ID: "docs", Label: "Documentation", Style: "background: transparent; color: #007bff; border: 1px solid #007bff;"},
	}
	c.Card1 = liveview.New("card1", card1)
	
	card2 := components.NewCard("Server-Side Rendering", "All components are rendered on the server and updated via WebSocket connections. No JavaScript framework required!")
	card2.Subtitle = "Powered by Go and Echo"
	card2.Width = "100%"
	card2.Footer = "Updated in real-time"
	card2.Hoverable = true
	c.Card2 = liveview.New("card2", card2)
	
	// Initialize Table
	c.Table = liveview.New("data_table", &components.Table{
		Columns: []components.Column{
			{Key: "id", Title: "ID", Width: "80px"},
			{Key: "name", Title: "Name", Sortable: true},
			{Key: "language", Title: "Language", Sortable: true},
			{Key: "status", Title: "Status"},
		},
		Rows: []components.Row{
			{"id": "1", "name": "Echo Framework", "language": "Go", "status": "Active"},
			{"id": "2", "name": "LiveView", "language": "Go", "status": "Active"},
			{"id": "3", "name": "Phoenix", "language": "Elixir", "status": "Active"},
			{"id": "4", "name": "Rails", "language": "Ruby", "status": "Active"},
			{"id": "5", "name": "Django", "language": "Python", "status": "Active"},
		},
		ShowPagination: true,
		PageSize: 10,
	})
	
	// Initialize Chart
	c.Chart = liveview.New("demo_chart", &components.Chart{
		Type:   components.ChartBar,
		Title:  "Programming Languages Popularity",
		Width:  600,
		Height: 400,
		Data: []components.ChartData{
			{Label: "Go", Value: 85, Color: "#00ADD8"},
			{Label: "Python", Value: 90, Color: "#3776AB"},
			{Label: "JavaScript", Value: 95, Color: "#F7DF1E"},
			{Label: "Rust", Value: 70, Color: "#DEA584"},
			{Label: "Ruby", Value: 65, Color: "#CC342D"},
		},
	})
	
	// Initialize Clock
	c.Clock = liveview.New("live_clock", &components.Clock{})
	
	// Initialize Breadcrumb
	breadcrumb := components.NewBreadcrumb([]components.BreadcrumbItem{
		{Label: "Home", Href: "/", Icon: "🏠"},
		{Label: "Examples", Href: "/examples"},
		{Label: "Component Showcase", Active: true},
	})
	c.Breadcrumb = liveview.New("breadcrumb", breadcrumb)
	
	// Initialize Header
	header := components.NewHeader("Go Echo LiveView", []components.HeaderMenuItem{
		{ID: "home", Label: "Home", Href: "#"},
		{ID: "components", Label: "Components", Href: "#"},
		{ID: "examples", Label: "Examples", Href: "#"},
		{ID: "docs", Label: "Documentation", Href: "#"},
	})
	c.Header = liveview.New("header", header)
	
	// Initialize Sidebar 
	sidebar := components.NewSidebar([]components.SidebarItem{
		{ID: "dashboard", Label: "Dashboard", Icon: "📊", Active: true},
		{ID: "charts", Label: "Charts", Icon: "📈"},
		{ID: "tables", Label: "Tables", Icon: "📋"},
		{ID: "settings", Label: "Settings", Icon: "⚙️"},
	})
	c.Sidebar = liveview.New("sidebar", sidebar)
	
	// Mount all components BEFORE setting events
	c.Mount(c.Alert)
	c.Mount(c.Accordion)
	c.Mount(c.Dropdown)
	c.Mount(c.Card1)
	c.Mount(c.Card2)
	c.Mount(c.Breadcrumb)
	c.Mount(c.Header)
	c.Mount(c.Sidebar)
	c.Mount(c.BtnInfo)
	c.Mount(c.BtnSuccess)
	c.Mount(c.BtnWarning)
	c.Mount(c.BtnError)
	c.Mount(c.Table)
	c.Mount(c.Chart)
	c.Mount(c.Clock)
	
	// Set custom sidebar event after mounting
	c.Sidebar.Events["Select"] = func(s *components.Sidebar, data interface{}) {
		if itemID, ok := data.(string); ok {
			s.Select(data)
			c.CurrentView = itemID
			// Use JavaScript to hide/show content without re-rendering
			c.EvalScript(`
				// Hide all content sections
				document.querySelectorAll('[data-view]').forEach(function(el) {
					el.style.display = 'none';
				});
				// Show selected section
				var selectedView = document.querySelector('[data-view="` + itemID + `"]');
				if (selectedView) {
					selectedView.style.display = 'block';
				}
			`)
		}
	}
	
	// Set custom header event after mounting
	c.Header.Events["MenuClick"] = func(h *components.Header, data interface{}) {
		if itemID, ok := data.(string); ok {
			switch itemID {
			case "home":
				c.CurrentView = "dashboard"
				c.EvalScript(`
					document.querySelectorAll('[data-view]').forEach(function(el) {
						el.style.display = 'none';
					});
					var dashboard = document.querySelector('[data-view="dashboard"]');
					if (dashboard) dashboard.style.display = 'block';
				`)
				// Also update sidebar selection
				c.Sidebar.Select("dashboard")
			}
		}
	}
	
	c.Commit()
}

func (c *ComponentShowcase) GetDriver() liveview.LiveDriver {
	return c
}

func (c *ComponentShowcase) GetTemplate() string {
	return `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Component Showcase - Go Echo LiveView</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
            background: #f5f5f5;
        }
        
        .app-container {
            display: flex;
            height: 100vh;
        }
        
        .main-content {
            flex: 1;
            display: flex;
            flex-direction: column;
            overflow: hidden;
        }
        
        .content-area {
            flex: 1;
            overflow-y: auto;
            padding: 24px;
        }
        
        .section {
            background: white;
            border-radius: 8px;
            padding: 24px;
            margin-bottom: 24px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        
        .section-title {
            font-size: 24px;
            font-weight: 600;
            margin-bottom: 20px;
            color: #333;
        }
        
        .cards-container {
            display: flex;
            gap: 20px;
            flex-wrap: wrap;
            justify-content: space-between;
        }
        
        .cards-container > * {
            flex: 0 1 calc(50% - 10px);
        }
        
        .alert-buttons {
            display: flex;
            gap: 12px;
            margin-top: 16px;
        }
        
        /* Style the LiveView buttons */
        .alert-buttons button {
            padding: 8px 16px;
            border: none;
            border-radius: 4px;
            cursor: pointer;
            font-size: 14px;
            transition: opacity 0.3s;
        }
        
        .alert-buttons button:hover {
            opacity: 0.8;
        }
        
        #btn_info { background: #17a2b8; color: white; }
        #btn_success { background: #28a745; color: white; }
        #btn_warning { background: #ffc107; color: #333; }
        #btn_error { background: #dc3545; color: white; }
        
        .placeholder-section {
            min-height: 300px;
            display: flex;
            align-items: center;
            justify-content: center;
            color: #666;
            font-size: 18px;
        }
    </style>
</head>
<body>
    <div class="app-container">
        {{mount "sidebar"}}
        
        <div class="main-content">
            {{mount "header"}}
            
            <div class="content-area">
                {{mount "breadcrumb"}}
                
                <div data-view="dashboard" style="display: block;">
                <div class="section">
                    <h2 class="section-title">Alert Component</h2>
                    {{mount "alert"}}
                    <div class="alert-buttons">
                        {{mount "btn_info"}}
                        {{mount "btn_success"}}
                        {{mount "btn_warning"}}
                        {{mount "btn_error"}}
                    </div>
                </div>
                
                <div class="section">
                    <h2 class="section-title">Live Clock</h2>
                    {{mount "live_clock"}}
                </div>
                
                <div class="section">
                    <h2 class="section-title">Dropdown Component</h2>
                    {{mount "dropdown"}}
                </div>
                
                <div class="section">
                    <h2 class="section-title">Accordion Component</h2>
                    {{mount "accordion"}}
                </div>
                
                <div class="section">
                    <h2 class="section-title">Card Components</h2>
                    <div class="cards-container">
                        {{mount "card1"}}
                        {{mount "card2"}}
                    </div>
                </div>
                </div>
                
                <div data-view="charts" style="display: none;">
                <div class="section">
                    <h2 class="section-title">📈 Chart Component</h2>
                    <p style="margin-bottom: 20px;">Interactive chart component with real-time updates</p>
                    {{mount "demo_chart"}}
                </div>
                </div>
                
                <div data-view="tables" style="display: none;">
                <div class="section">
                    <h2 class="section-title">📋 Table Component</h2>
                    <p style="margin-bottom: 20px;">Data table with sorting, filtering and pagination</p>
                    {{mount "data_table"}}
                </div>
                </div>
                
                <div data-view="settings" style="display: none;">
                <div class="section">
                    <h2 class="section-title">⚙️ Settings</h2>
                    <div class="section">
                        <h3>Application Settings</h3>
                        {{mount "dropdown"}}
                    </div>
                    <div class="section">
                        <h3>Live Time</h3>
                        {{mount "live_clock"}}
                    </div>
                </div>
                </div>
            </div>
        </div>
    </div>
</body>
</html>
`
}

func main() {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Static("/assets", "assets")

	home := liveview.PageControl{
		Title:  "Component Showcase",
		Lang:   "en",
		Path:   "/",
		Router: e,
	}

	home.Register(func() liveview.LiveDriver {
		showcase := NewComponentShowcase()
		cmp := liveview.NewDriver("showcase", showcase)
		showcase.ComponentDriver = cmp
		return cmp
	})

	e.Logger.Fatal(e.Start(":8080"))
}