package main

import (
	"fmt"
	"time"

	"github.com/arturoeanton/go-echo-live-view/components"
	"github.com/arturoeanton/go-echo-live-view/liveview"
	"github.com/labstack/echo/v4"
)

type ComponentShowcaseV2 struct {
	*liveview.ComponentDriver[*ComponentShowcaseV2]
	
	// All components are preserved as they're mounted separately
	Form          *components.Form
	FileUpload    *components.FileUpload
	Table         *components.Table
	Modal         *components.Modal
	Notifications *components.NotificationSystem
	Chart         *components.Chart
	RichEditor    *components.RichEditor
	Calendar      *components.Calendar
	Draggable     *components.Draggable
	Animation     *components.Animation
}

func (c *ComponentShowcaseV2) Start() {
	// Initialize Form
	c.Form = liveview.New("showcase_form", &components.Form{
		Fields: []components.FormField{
			{
				Name:        "email",
				Label:       "Email Address",
				Type:        "email",
				Placeholder: "user@example.com",
				Required:    true,
				Rules: []components.ValidationRule{
					{Type: "email", Message: "Please enter a valid email"},
				},
			},
			{
				Name:        "password",
				Label:       "Password",
				Type:        "password",
				Placeholder: "Enter password",
				Required:    true,
				MinLength:   8,
			},
			{
				Name:        "age",
				Label:       "Age",
				Type:        "number",
				Min:         "18",
				Max:         "100",
				Required:    true,
			},
		},
		SubmitLabel: "Submit Form",
		OnSubmit: func(data map[string]string) error {
			c.Notifications.Success("Form Submitted", fmt.Sprintf("Email: %s", data["email"]))
			return nil
		},
	})
	
	// Initialize File Upload
	c.FileUpload = liveview.New("showcase_upload", &components.FileUpload{
		Multiple: true,
		Accept:   "image/*,.pdf",
		MaxSize:  5 * 1024 * 1024,
		Label:    "Drop files here or click to browse",
		OnUpload: func(files []components.FileInfo) error {
			c.Notifications.Success("Files Uploaded", fmt.Sprintf("%d files uploaded successfully", len(files)))
			return nil
		},
	})
	
	// Initialize Table
	c.Table = liveview.New("showcase_table", &components.Table{
		Columns: []components.Column{
			{Key: "id", Title: "ID", Width: "80px", Sortable: true},
			{Key: "name", Title: "Name", Sortable: true},
			{Key: "email", Title: "Email", Sortable: true},
			{Key: "role", Title: "Role", Sortable: true},
			{Key: "status", Title: "Status", Width: "100px"},
		},
		Rows: []components.Row{
			{"id": 1, "name": "John Doe", "email": "john@example.com", "role": "Admin", "status": "Active"},
			{"id": 2, "name": "Jane Smith", "email": "jane@example.com", "role": "User", "status": "Active"},
			{"id": 3, "name": "Bob Johnson", "email": "bob@example.com", "role": "User", "status": "Inactive"},
			{"id": 4, "name": "Alice Brown", "email": "alice@example.com", "role": "Manager", "status": "Active"},
			{"id": 5, "name": "Charlie Wilson", "email": "charlie@example.com", "role": "User", "status": "Active"},
		},
		PageSize:       3,
		ShowPagination: true,
		Selectable:     true,
		OnRowClick: func(row components.Row, index int) {
			c.Notifications.Info("Row Clicked", fmt.Sprintf("Selected: %v", row["name"]))
		},
	})
	
	// Initialize Modal
	c.Modal = liveview.New("showcase_modal", &components.Modal{
		Title:      "Example Modal",
		Content:    "This is a demonstration of the modal component. You can add any content here!",
		Size:       "medium",
		Closable:   true,
		ShowFooter: true,
		OnOk: func() {
			c.Notifications.Success("Modal", "OK button clicked")
		},
		OnCancel: func() {
			c.Notifications.Info("Modal", "Cancel button clicked")
		},
	})
	
	// Initialize Notifications
	c.Notifications = liveview.New("showcase_notifications", &components.NotificationSystem{
		Position:   "top-right",
		MaxVisible: 5,
	})
	
	// Initialize Chart
	c.Chart = liveview.New("showcase_chart", &components.Chart{
		Type:   components.ChartBar,
		Title:  "Sales by Month",
		Width:  600,
		Height: 400,
		Data: []components.ChartData{
			{Label: "Jan", Value: 150, Color: "#4CAF50"},
			{Label: "Feb", Value: 200, Color: "#2196F3"},
			{Label: "Mar", Value: 180, Color: "#FF9800"},
			{Label: "Apr", Value: 220, Color: "#F44336"},
			{Label: "May", Value: 250, Color: "#9C27B0"},
		},
	})
	
	// Initialize Rich Editor
	c.RichEditor = liveview.New("showcase_editor", &components.RichEditor{
		Content:     "<h2>Welcome to the Rich Editor!</h2><p>You can format text, create lists, and more...</p>",
		Placeholder: "Start typing your content...",
		Height:      "250px",
		OnChange: func(content string) {
			fmt.Println("Editor content changed")
		},
	})
	
	// Initialize Calendar
	c.Calendar = liveview.New("showcase_calendar", &components.Calendar{
		SelectedDate: time.Now(),
		OnSelect: func(date time.Time) {
			c.Notifications.Info("Date Selected", date.Format("January 2, 2006"))
		},
	})
	
	// Initialize Draggable
	c.Draggable = liveview.New("showcase_draggable", &components.Draggable{
		Containers: []string{"To Do", "In Progress", "Done"},
		Items: []components.DragItem{
			{ID: "task1", Content: "Design homepage", Group: "To Do"},
			{ID: "task2", Content: "Implement login", Group: "In Progress"},
			{ID: "task3", Content: "Write tests", Group: "To Do"},
			{ID: "task4", Content: "Deploy to staging", Group: "Done"},
		},
		OnDrop: func(itemID, from, to string) {
			c.Notifications.Success("Item Moved", fmt.Sprintf("Moved from %s to %s", from, to))
		},
	})
	
	// Initialize Animation
	c.Animation = liveview.New("showcase_animation", &components.Animation{
		Content:        "<div style='padding: 2rem; background: linear-gradient(45deg, #4CAF50, #2196F3); color: white; border-radius: 8px; font-size: 1.5rem; font-weight: bold;'>Animated Element</div>",
		Type:          components.AnimationBounce,
		Duration:      "2s",
		IterationCount: "1",
		IsPlaying:     false,
	})
	
	// Mount all components
	c.Mount(c.Form)
	c.Mount(c.FileUpload)
	c.Mount(c.Table)
	c.Mount(c.Modal)
	c.Mount(c.Notifications)
	c.Mount(c.Chart)
	c.Mount(c.RichEditor)
	c.Mount(c.Calendar)
	c.Mount(c.Draggable)
	c.Mount(c.Animation)
	
	c.Commit()
}

func (c *ComponentShowcaseV2) GetTemplate() string {
	return `
	<div style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; padding: 2rem; max-width: 1400px; margin: 0 auto;">
		<h1 style="color: #333; margin-bottom: 2rem;">Go Echo LiveView Components Showcase V2</h1>
		
		{{mount "showcase_notifications"}}
		{{mount "showcase_modal"}}
		
		<div class="tab-container" id="main-tabs">
			<style>
				.tab-buttons {
					display: flex;
					gap: 1rem;
					margin-bottom: 2rem;
					flex-wrap: wrap;
					border-bottom: 2px solid #eee;
					padding-bottom: 1rem;
				}
				.tab-button {
					padding: 0.75rem 1.5rem;
					background: white;
					color: #333;
					border: 1px solid #ddd;
					border-radius: 4px;
					cursor: pointer;
					font-weight: 500;
				}
				.tab-button:hover {
					background: #f5f5f5;
				}
				.tab-button.active {
					background: #4CAF50;
					color: white;
				}
				.tab-content {
					background: white;
					padding: 2rem;
					border-radius: 8px;
					box-shadow: 0 2px 8px rgba(0,0,0,0.1);
				}
				.tab-pane {
					display: none;
				}
				.tab-pane.active {
					display: block;
				}
			</style>
			
			<div class="tab-buttons">
				<button class="tab-button active" id="btn-form" onclick="send_event('{{.IdComponent}}', 'SwitchTab', 'form')">Form Validation</button>
				<button class="tab-button" id="btn-upload" onclick="send_event('{{.IdComponent}}', 'SwitchTab', 'upload')">File Upload</button>
				<button class="tab-button" id="btn-table" onclick="send_event('{{.IdComponent}}', 'SwitchTab', 'table')">Data Table</button>
				<button class="tab-button" id="btn-chart" onclick="send_event('{{.IdComponent}}', 'SwitchTab', 'chart')">Charts</button>
				<button class="tab-button" id="btn-editor" onclick="send_event('{{.IdComponent}}', 'SwitchTab', 'editor')">Rich Editor</button>
				<button class="tab-button" id="btn-calendar" onclick="send_event('{{.IdComponent}}', 'SwitchTab', 'calendar')">Calendar</button>
				<button class="tab-button" id="btn-drag" onclick="send_event('{{.IdComponent}}', 'SwitchTab', 'drag')">Drag & Drop</button>
				<button class="tab-button" id="btn-animation" onclick="send_event('{{.IdComponent}}', 'SwitchTab', 'animation')">Animations</button>
			</div>
			
			<div class="tab-content">
				<div id="tab-form" class="tab-pane active">
					<h2>Form Validation Component</h2>
					<p style="color: #666; margin-bottom: 2rem;">Comprehensive form with built-in validation rules</p>
					{{mount "showcase_form"}}
				</div>
				
				<div id="tab-upload" class="tab-pane">
					<h2>File Upload Component</h2>
					<p style="color: #666; margin-bottom: 2rem;">Drag & drop file upload with preview</p>
					{{mount "showcase_upload"}}
				</div>
				
				<div id="tab-table" class="tab-pane">
					<h2>Data Table Component</h2>
					<p style="color: #666; margin-bottom: 2rem;">Feature-rich table with sorting, filtering, and pagination</p>
					{{mount "showcase_table"}}
				</div>
				
				<div id="tab-chart" class="tab-pane">
					<h2>Chart Component</h2>
					<p style="color: #666; margin-bottom: 2rem;">Interactive charts and visualizations</p>
					{{mount "showcase_chart"}}
					<div style="margin-top: 2rem;">
						<button onclick="send_event('{{.IdComponent}}', 'ChangeChartType', '')" 
							style="padding: 0.5rem 1rem; background: #2196F3; color: white; border: none; border-radius: 4px; cursor: pointer;">
							Toggle Chart Type
						</button>
					</div>
				</div>
				
				<div id="tab-editor" class="tab-pane">
					<h2>Rich Text Editor</h2>
					<p style="color: #666; margin-bottom: 2rem;">WYSIWYG editor with formatting tools</p>
					{{mount "showcase_editor"}}
				</div>
				
				<div id="tab-calendar" class="tab-pane">
					<h2>Calendar Component</h2>
					<p style="color: #666; margin-bottom: 2rem;">Interactive date picker</p>
					<div style="display: flex; justify-content: center;">
						{{mount "showcase_calendar"}}
					</div>
				</div>
				
				<div id="tab-drag" class="tab-pane">
					<h2>Drag & Drop Component</h2>
					<p style="color: #666; margin-bottom: 2rem;">Kanban-style drag and drop interface</p>
					{{mount "showcase_draggable"}}
				</div>
				
				<div id="tab-animation" class="tab-pane">
					<h2>Animation Framework</h2>
					<p style="color: #666; margin-bottom: 2rem;">Built-in animation effects</p>
					{{mount "showcase_animation"}}
				</div>
			</div>
		</div>
		
		<div style="margin-top: 2rem; padding: 1rem; background: #f5f5f5; border-radius: 8px;">
			<h3>Demo Actions</h3>
			<div style="display: flex; gap: 1rem; flex-wrap: wrap; margin-top: 1rem;">
				<button onclick="send_event('{{.IdComponent}}', 'ShowModal', '')" 
					style="padding: 0.5rem 1rem; background: #9C27B0; color: white; border: none; border-radius: 4px; cursor: pointer;">
					Show Modal
				</button>
				<button onclick="send_event('{{.IdComponent}}', 'ShowSuccess', '')" 
					style="padding: 0.5rem 1rem; background: #4CAF50; color: white; border: none; border-radius: 4px; cursor: pointer;">
					Success Notification
				</button>
				<button onclick="send_event('{{.IdComponent}}', 'ShowError', '')" 
					style="padding: 0.5rem 1rem; background: #F44336; color: white; border: none; border-radius: 4px; cursor: pointer;">
					Error Notification
				</button>
				<button onclick="send_event('{{.IdComponent}}', 'ShowWarning', '')" 
					style="padding: 0.5rem 1rem; background: #FF9800; color: white; border: none; border-radius: 4px; cursor: pointer;">
					Warning Notification
				</button>
				<button onclick="send_event('{{.IdComponent}}', 'ShowInfo', '')" 
					style="padding: 0.5rem 1rem; background: #2196F3; color: white; border: none; border-radius: 4px; cursor: pointer;">
					Info Notification
				</button>
			</div>
		</div>
	</div>
	`
}

func (c *ComponentShowcaseV2) GetDriver() liveview.LiveDriver {
	return c
}

func (c *ComponentShowcaseV2) SwitchTab(data interface{}) {
	tabName := data.(string)
	
	liveview.Info("Switching to tab: %s", tabName)
	
	// Use JavaScript to switch tabs without destroying mounted components
	script := fmt.Sprintf(`
		(function() {
			// Hide all tabs
			document.querySelectorAll('.tab-pane').forEach(function(pane) {
				pane.classList.remove('active');
			});
			
			// Show selected tab
			var selectedTab = document.getElementById('tab-%s');
			if (selectedTab) {
				selectedTab.classList.add('active');
			}
			
			// Update button states
			document.querySelectorAll('.tab-button').forEach(function(btn) {
				btn.classList.remove('active');
			});
			var activeBtn = document.getElementById('btn-%s');
			if (activeBtn) {
				activeBtn.classList.add('active');
			}
			
			console.log('Switched to tab: %s');
		})();
	`, tabName, tabName, tabName)
	
	c.EvalScript(script)
}

func (c *ComponentShowcaseV2) ShowModal(data interface{}) {
	c.Modal.Open()
}

func (c *ComponentShowcaseV2) ShowSuccess(data interface{}) {
	c.Notifications.Success("Success!", "This is a success notification")
}

func (c *ComponentShowcaseV2) ShowError(data interface{}) {
	c.Notifications.Error("Error!", "This is an error notification")
}

func (c *ComponentShowcaseV2) ShowWarning(data interface{}) {
	c.Notifications.Warning("Warning!", "This is a warning notification")
}

func (c *ComponentShowcaseV2) ShowInfo(data interface{}) {
	c.Notifications.Info("Info!", "This is an info notification")
}

func (c *ComponentShowcaseV2) ChangeChartType(data interface{}) {
	if c.Chart.Type == components.ChartBar {
		c.Chart.Type = components.ChartPie
	} else {
		c.Chart.Type = components.ChartBar
	}
	c.Chart.Commit()
}

func main() {
	// Inicializar logger con modo verbose
	liveview.InitLogger(true)
	liveview.Info("Starting Component Showcase V2 Server...")
	
	e := echo.New()
	e.Static("/assets", "assets")
	
	home := liveview.PageControl{
		Title:  "Component Showcase V2",
		Lang:   "en",
		Path:   "/",
		Router: e,
		Debug:  true,
	}
	
	home.Register(func() liveview.LiveDriver {
		return liveview.NewDriver("showcase_v2", &ComponentShowcaseV2{})
	})
	
	fmt.Println("Server starting at http://localhost:8086")
	fmt.Println("Visit http://localhost:8086 to see all components with proper tab switching!")
	e.Logger.Fatal(e.Start(":8086"))
}