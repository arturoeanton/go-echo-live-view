package main

import (
	"fmt"
	"log"
	"time"

	"github.com/arturoeanton/go-echo-live-view/components"
	"github.com/arturoeanton/go-echo-live-view/liveview"
	"github.com/labstack/echo/v4"
)

type AllComponentsTest struct {
	*liveview.ComponentDriver[*AllComponentsTest]
	
	// All components for testing
	Calendar  *components.Calendar
	Draggable *components.Draggable
	Animation *components.Animation
	Form      *components.Form
	Table     *components.Table
	Chart     *components.Chart
	
	StatusMsg string
	TestResults map[string]string
}

func (t *AllComponentsTest) Start() {
	liveview.Info("AllComponentsTest Start")
	
	t.StatusMsg = "Testing all components"
	t.TestResults = make(map[string]string)
	
	// Calendar - Test date selection
	t.Calendar = liveview.New("test_calendar", &components.Calendar{
		SelectedDate: time.Now(),
		OnSelect: func(date time.Time) {
			msg := fmt.Sprintf("Calendar: Selected %s", date.Format("January 2, 2006"))
			t.TestResults["calendar"] = "✅ Working"
			t.StatusMsg = msg
			liveview.Info(msg)
			t.UpdateStatus()
		},
	})
	
	// Draggable - Test drag and drop
	t.Draggable = liveview.New("test_draggable", &components.Draggable{
		Containers: []string{"To Do", "In Progress", "Done"},
		Items: []components.DragItem{
			{ID: "task1", Content: "Test Task 1", Group: "To Do"},
			{ID: "task2", Content: "Test Task 2", Group: "In Progress"},
			{ID: "task3", Content: "Test Task 3", Group: "Done"},
		},
		OnDrop: func(itemID, from, to string) {
			msg := fmt.Sprintf("Draggable: Moved %s from %s to %s", itemID, from, to)
			t.TestResults["draggable"] = "✅ Working"
			t.StatusMsg = msg
			liveview.Info(msg)
			t.UpdateStatus()
		},
	})
	
	// Animation - Test animation playback
	t.Animation = liveview.New("test_animation", &components.Animation{
		Content:        `<div style="padding: 1rem; background: #4CAF50; color: white; border-radius: 8px;">Test Animation</div>`,
		Type:          components.AnimationBounce,
		Duration:      "2s",
		IterationCount: "1",
		IsPlaying:     false,
	})
	t.TestResults["animation"] = "✅ Visible"
	
	// Form - Test form submission
	t.Form = liveview.New("test_form", &components.Form{
		Fields: []components.FormField{
			{
				Name:        "test_field",
				Label:       "Test Input",
				Type:        "text",
				Placeholder: "Enter test value",
				Required:    true,
			},
		},
		SubmitLabel: "Test Submit",
		OnSubmit: func(data map[string]string) error {
			msg := fmt.Sprintf("Form: Submitted with value: %s", data["test_field"])
			t.TestResults["form"] = "✅ Working"
			t.StatusMsg = msg
			liveview.Info(msg)
			t.UpdateStatus()
			return nil
		},
	})
	
	// Table - Test row interaction
	t.Table = liveview.New("test_table", &components.Table{
		Columns: []components.Column{
			{Key: "id", Title: "ID", Width: "80px"},
			{Key: "name", Title: "Name"},
			{Key: "status", Title: "Status"},
		},
		Rows: []components.Row{
			{"id": 1, "name": "Test Row 1", "status": "Active"},
			{"id": 2, "name": "Test Row 2", "status": "Active"},
			{"id": 3, "name": "Test Row 3", "status": "Inactive"},
		},
		PageSize:       5,
		ShowPagination: false,
		Selectable:     true,
		OnRowClick: func(row components.Row, index int) {
			msg := fmt.Sprintf("Table: Clicked row %v", row["name"])
			t.TestResults["table"] = "✅ Working"
			t.StatusMsg = msg
			liveview.Info(msg)
			t.UpdateStatus()
		},
	})
	
	// Chart - Test rendering
	t.Chart = liveview.New("test_chart", &components.Chart{
		Type:   components.ChartBar,
		Title:  "Test Chart",
		Width:  400,
		Height: 300,
		Data: []components.ChartData{
			{Label: "A", Value: 10, Color: "#4CAF50"},
			{Label: "B", Value: 20, Color: "#2196F3"},
			{Label: "C", Value: 15, Color: "#FF9800"},
		},
	})
	t.TestResults["chart"] = "✅ Rendered"
	
	// Mount all
	t.Mount(t.Calendar)
	t.Mount(t.Draggable)
	t.Mount(t.Animation)
	t.Mount(t.Form)
	t.Mount(t.Table)
	t.Mount(t.Chart)
	
	t.Commit()
}

func (t *AllComponentsTest) UpdateStatus() {
	// Only update status without re-rendering everything
	script := fmt.Sprintf(`
		var statusEl = document.getElementById('status-msg');
		if (statusEl) statusEl.innerText = '%s';
		
		var resultsEl = document.getElementById('test-results');
		if (resultsEl) {
			var html = '<ul>';
			%s
			html += '</ul>';
			resultsEl.innerHTML = html;
		}
	`, t.StatusMsg, t.buildResultsJS())
	
	t.EvalScript(script)
}

func (t *AllComponentsTest) buildResultsJS() string {
	js := ""
	for comp, status := range t.TestResults {
		js += fmt.Sprintf("html += '<li>%s: %s</li>';", comp, status)
	}
	return js
}

func (t *AllComponentsTest) GetTemplate() string {
	return `
	<div style="font-family: Arial, sans-serif; padding: 2rem; max-width: 1400px; margin: 0 auto;">
		<h1>All Components Test</h1>
		
		<div style="background: #f0f0f0; padding: 1rem; border-radius: 8px; margin: 1rem 0;">
			<h3>Status: <span id="status-msg">{{.StatusMsg}}</span></h3>
			<div id="test-results">
				<ul>
					{{range $key, $value := .TestResults}}
					<li>{{$key}}: {{$value}}</li>
					{{end}}
				</ul>
			</div>
		</div>
		
		<div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(400px, 1fr)); gap: 2rem; margin-top: 2rem;">
			<!-- Calendar -->
			<div style="background: white; padding: 1.5rem; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1);">
				<h3>Calendar Component</h3>
				<p style="color: #666; font-size: 0.9rem;">Click on dates to test selection</p>
				{{mount "test_calendar"}}
			</div>
			
			<!-- Draggable -->
			<div style="background: white; padding: 1.5rem; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1);">
				<h3>Drag & Drop Component</h3>
				<p style="color: #666; font-size: 0.9rem;">Drag items between columns</p>
				{{mount "test_draggable"}}
			</div>
			
			<!-- Animation -->
			<div style="background: white; padding: 1.5rem; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1);">
				<h3>Animation Component</h3>
				<p style="color: #666; font-size: 0.9rem;">Click Play to test animation</p>
				{{mount "test_animation"}}
			</div>
			
			<!-- Form -->
			<div style="background: white; padding: 1.5rem; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1);">
				<h3>Form Component</h3>
				<p style="color: #666; font-size: 0.9rem;">Submit form to test</p>
				{{mount "test_form"}}
			</div>
			
			<!-- Table -->
			<div style="background: white; padding: 1.5rem; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1);">
				<h3>Table Component</h3>
				<p style="color: #666; font-size: 0.9rem;">Click rows to test interaction</p>
				{{mount "test_table"}}
			</div>
			
			<!-- Chart -->
			<div style="background: white; padding: 1.5rem; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1);">
				<h3>Chart Component</h3>
				<p style="color: #666; font-size: 0.9rem;">Visual rendering test</p>
				{{mount "test_chart"}}
			</div>
		</div>
		
		<div style="margin-top: 2rem; padding: 1rem; background: #e3f2fd; border: 1px solid #2196F3; border-radius: 4px;">
			<h4>Test Actions:</h4>
			<button onclick="send_event('{{.IdComponent}}', 'TestAnimation', '')" 
				style="padding: 0.5rem 1rem; background: #2196F3; color: white; border: none; border-radius: 4px; cursor: pointer; margin-right: 1rem;">
				Test Animation
			</button>
			<button onclick="send_event('{{.IdComponent}}', 'ResetTests', '')" 
				style="padding: 0.5rem 1rem; background: #FF9800; color: white; border: none; border-radius: 4px; cursor: pointer;">
				Reset All Tests
			</button>
		</div>
	</div>
	`
}

func (t *AllComponentsTest) GetDriver() liveview.LiveDriver {
	return t
}

func (t *AllComponentsTest) TestAnimation(data interface{}) {
	t.Animation.Play(nil)
	t.TestResults["animation"] = "✅ Animated"
	t.StatusMsg = "Animation triggered"
	t.UpdateStatus()
}

func (t *AllComponentsTest) ResetTests(data interface{}) {
	t.TestResults = make(map[string]string)
	t.TestResults["calendar"] = "⏳ Waiting for test"
	t.TestResults["draggable"] = "⏳ Waiting for test"
	t.TestResults["animation"] = "⏳ Waiting for test"
	t.TestResults["form"] = "⏳ Waiting for test"
	t.TestResults["table"] = "⏳ Waiting for test"
	t.TestResults["chart"] = "✅ Rendered"
	t.StatusMsg = "Tests reset - interact with components to test"
	t.UpdateStatus()
}

func main() {
	liveview.InitLogger(true)
	liveview.Info("Starting All Components Test Server...")
	
	e := echo.New()
	e.Static("/assets", "assets")
	
	home := liveview.PageControl{
		Title:  "All Components Test",
		Lang:   "en",
		Path:   "/",
		Router: e,
		Debug:  true,
	}
	
	home.Register(func() liveview.LiveDriver {
		liveview.Info("Creating new AllComponentsTest instance")
		return liveview.NewDriver("all_test", &AllComponentsTest{})
	})
	
	log.Println("==========================================")
	log.Println("      All Components Test Server")
	log.Println("==========================================")
	log.Println("Server: http://localhost:8085")
	log.Println("==========================================")
	log.Println()
	log.Println("Test each component by interacting with it:")
	log.Println("- Calendar: Click on dates")
	log.Println("- Drag & Drop: Drag items between columns")
	log.Println("- Animation: Click 'Play Animation'")
	log.Println("- Form: Fill and submit")
	log.Println("- Table: Click on rows")
	log.Println("- Chart: Visual test (auto-passes)")
	log.Println()
	
	e.Logger.Fatal(e.Start(":8085"))
}