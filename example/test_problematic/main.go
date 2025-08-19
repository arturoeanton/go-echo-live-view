package main

import (
	"fmt"
	"log"
	"time"

	"github.com/arturoeanton/go-echo-live-view/components"
	"github.com/arturoeanton/go-echo-live-view/liveview"
	"github.com/labstack/echo/v4"
)

type ProblematicTest struct {
	*liveview.ComponentDriver[*ProblematicTest]
	
	Calendar  *components.Calendar
	Draggable *components.Draggable
	Animation *components.Animation
	
	ActiveComponent string
	StatusMsg       string
}

func (p *ProblematicTest) Start() {
	liveview.Info("ProblematicTest Start")
	
	p.ActiveComponent = "calendar"
	p.StatusMsg = "Testing problematic components"
	
	// Calendar
	p.Calendar = liveview.New("test_calendar", &components.Calendar{
		SelectedDate: time.Now(),
		OnSelect: func(date time.Time) {
			p.StatusMsg = fmt.Sprintf("Calendar: Selected %s", date.Format("January 2, 2006"))
			liveview.Info("Calendar date selected: %s", date.Format("2006-01-02"))
			p.Commit()
		},
	})
	
	// Draggable
	p.Draggable = liveview.New("test_draggable", &components.Draggable{
		Containers: []string{"To Do", "In Progress", "Done"},
		Items: []components.DragItem{
			{ID: "task1", Content: "Task 1", Group: "To Do"},
			{ID: "task2", Content: "Task 2", Group: "In Progress"},
			{ID: "task3", Content: "Task 3", Group: "Done"},
		},
		OnDrop: func(itemID, from, to string) {
			p.StatusMsg = fmt.Sprintf("Draggable: Moved %s from %s to %s", itemID, from, to)
			liveview.Info("Item dropped: %s from %s to %s", itemID, from, to)
		},
	})
	
	// Animation
	p.Animation = liveview.New("test_animation", &components.Animation{
		Content:        `<div style="padding: 2rem; background: linear-gradient(45deg, #4CAF50, #2196F3); color: white; border-radius: 8px; font-size: 1.5rem; font-weight: bold;">Animated Element</div>`,
		Type:          components.AnimationBounce,
		Duration:      "1s",
		IterationCount: "infinite",
	})
	
	// Mount all
	p.Mount(p.Calendar)
	p.Mount(p.Draggable)
	p.Mount(p.Animation)
	
	p.Commit()
}

func (p *ProblematicTest) GetTemplate() string {
	return `
	<div style="font-family: Arial, sans-serif; padding: 2rem; max-width: 1200px; margin: 0 auto;">
		<h1>Problematic Components Test</h1>
		
		<div style="background: #f0f0f0; padding: 1rem; border-radius: 8px; margin: 1rem 0;">
			<p style="font-size: 1.2rem; color: #333;">Status: <span id="status-msg">{{.StatusMsg}}</span></p>
		</div>
		
		<div style="display: flex; gap: 1rem; margin: 2rem 0;">
			<button onclick="send_event('{{.IdComponent}}', 'ShowCalendar', '')" 
				style="padding: 0.75rem 1.5rem; background: {{if eq .ActiveComponent "calendar"}}#4CAF50{{else}}#e0e0e0{{end}}; 
				color: {{if eq .ActiveComponent "calendar"}}white{{else}}#333{{end}}; border: none; border-radius: 4px; cursor: pointer;">
				Calendar
			</button>
			<button onclick="send_event('{{.IdComponent}}', 'ShowDraggable', '')" 
				style="padding: 0.75rem 1.5rem; background: {{if eq .ActiveComponent "draggable"}}#4CAF50{{else}}#e0e0e0{{end}}; 
				color: {{if eq .ActiveComponent "draggable"}}white{{else}}#333{{end}}; border: none; border-radius: 4px; cursor: pointer;">
				Drag & Drop
			</button>
			<button onclick="send_event('{{.IdComponent}}', 'ShowAnimation', '')" 
				style="padding: 0.75rem 1.5rem; background: {{if eq .ActiveComponent "animation"}}#4CAF50{{else}}#e0e0e0{{end}}; 
				color: {{if eq .ActiveComponent "animation"}}white{{else}}#333{{end}}; border: none; border-radius: 4px; cursor: pointer;">
				Animation
			</button>
		</div>
		
		<div style="background: white; padding: 2rem; border-radius: 8px; box-shadow: 0 2px 8px rgba(0,0,0,0.1); min-height: 400px;">
			<!-- Calendar -->
			<div id="calendar-container" style="display: {{if eq .ActiveComponent "calendar"}}block{{else}}none{{end}};">
				<h2>Calendar Component</h2>
				<div style="display: flex; justify-content: center; margin-top: 2rem;">
					{{mount "test_calendar"}}
				</div>
			</div>
			
			<!-- Draggable -->
			<div id="draggable-container" style="display: {{if eq .ActiveComponent "draggable"}}block{{else}}none{{end}};">
				<h2>Drag & Drop Component</h2>
				<div style="margin-top: 2rem;">
					{{mount "test_draggable"}}
				</div>
			</div>
			
			<!-- Animation -->
			<div id="animation-container" style="display: {{if eq .ActiveComponent "animation"}}block{{else}}none{{end}};">
				<h2>Animation Component</h2>
				<div style="margin-top: 2rem; display: flex; justify-content: center;">
					{{mount "test_animation"}}
				</div>
			</div>
		</div>
		
		<div style="margin-top: 2rem; padding: 1rem; background: #fffbf0; border: 1px solid #ffa500; border-radius: 4px;">
			<h4>Debug Actions:</h4>
			<button onclick="send_event('{{.IdComponent}}', 'TriggerAnimation', '')" 
				style="padding: 0.5rem 1rem; background: #FF9800; color: white; border: none; border-radius: 4px; cursor: pointer; margin-right: 1rem;">
				Trigger Animation
			</button>
			<button onclick="send_event('{{.IdComponent}}', 'ResetAll', '')" 
				style="padding: 0.5rem 1rem; background: #f44336; color: white; border: none; border-radius: 4px; cursor: pointer;">
				Reset All
			</button>
		</div>
	</div>
	`
}

func (p *ProblematicTest) GetDriver() liveview.LiveDriver {
	return p
}

func (p *ProblematicTest) ShowCalendar(data interface{}) {
	p.ActiveComponent = "calendar"
	p.StatusMsg = "Showing Calendar"
	liveview.Info("Switching to Calendar")
	// Now Commit preserves mounted components
	p.Commit()
}

func (p *ProblematicTest) ShowDraggable(data interface{}) {
	p.ActiveComponent = "draggable"
	p.StatusMsg = "Showing Drag & Drop"
	liveview.Info("Switching to Draggable")
	// Now Commit preserves mounted components
	p.Commit()
}

func (p *ProblematicTest) ShowAnimation(data interface{}) {
	p.ActiveComponent = "animation"
	p.StatusMsg = "Showing Animation"
	liveview.Info("Switching to Animation")
	// Now Commit preserves mounted components
	p.Commit()
}

func (p *ProblematicTest) TriggerAnimation(data interface{}) {
	// Reset animation
	p.Animation.Type = components.AnimationFadeIn
	p.Animation.Commit()
	
	time.AfterFunc(100*time.Millisecond, func() {
		p.Animation.Type = components.AnimationBounce
		p.Animation.Commit()
		p.StatusMsg = "Animation triggered!"
		p.Commit()
	})
}

func (p *ProblematicTest) ResetAll(data interface{}) {
	p.StatusMsg = "Components reset"
	p.Calendar.SelectedDate = time.Now()
	p.Calendar.CurrentMonth = time.Now()
	p.Calendar.Commit()
	p.Commit()
}

func main() {
	liveview.InitLogger(true)
	liveview.Info("Starting Problematic Components Test Server...")
	
	e := echo.New()
	e.Static("/assets", "assets")
	
	home := liveview.PageControl{
		Title:  "Problematic Components Test",
		Lang:   "en",
		Path:   "/",
		Router: e,
		Debug:  true,
	}
	
	home.Register(func() liveview.LiveDriver {
		liveview.Info("Creating new ProblematicTest instance")
		return liveview.NewDriver("problematic_test", &ProblematicTest{})
	})
	
	log.Println("==========================================")
	log.Println("   Problematic Components Test Server")
	log.Println("==========================================")
	log.Println("Server: http://localhost:8084")
	log.Println("==========================================")
	log.Println()
	log.Println("Test each component:")
	log.Println("1. Calendar - Try navigation and date selection")
	log.Println("2. Drag & Drop - Try dragging items between columns")
	log.Println("3. Animation - Should animate continuously")
	log.Println()
	
	e.Logger.Fatal(e.Start(":8084"))
}