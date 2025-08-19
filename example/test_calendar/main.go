package main

import (
	"log"
	"time"

	"github.com/arturoeanton/go-echo-live-view/components"
	"github.com/arturoeanton/go-echo-live-view/liveview"
	"github.com/labstack/echo/v4"
)

type CalendarTest struct {
	*liveview.ComponentDriver[*CalendarTest]
	Calendar *components.Calendar
	SelectedInfo string
}

func (c *CalendarTest) Start() {
	liveview.Info("CalendarTest Start")
	
	c.SelectedInfo = "No date selected yet"
	
	c.Calendar = liveview.New("test_calendar", &components.Calendar{
		SelectedDate: time.Now(),
		OnSelect: func(date time.Time) {
			c.SelectedInfo = "Selected: " + date.Format("January 2, 2006")
			liveview.Info("Date selected: %s", date.Format("2006-01-02"))
			c.Commit()
		},
	})
	
	c.Mount(c.Calendar)
	c.Commit()
}

func (c *CalendarTest) GetTemplate() string {
	return `
	<div style="font-family: Arial, sans-serif; padding: 2rem; max-width: 800px; margin: 0 auto;">
		<h1>Calendar Component Test</h1>
		
		<div style="background: #f0f0f0; padding: 1rem; border-radius: 8px; margin: 1rem 0;">
			<p style="font-size: 1.2rem; color: #333;">{{.SelectedInfo}}</p>
		</div>
		
		<div style="margin: 2rem 0; display: flex; justify-content: center;">
			{{mount "test_calendar"}}
		</div>
		
		<div style="margin-top: 2rem; padding: 1rem; background: #fffbf0; border: 1px solid #ffa500; border-radius: 4px;">
			<h4>Instructions:</h4>
			<ul>
				<li>Click &lt; or &gt; to navigate months</li>
				<li>Click on any day to select it</li>
				<li>Today's date should have a green border</li>
				<li>Selected date should have green background</li>
			</ul>
		</div>
	</div>
	`
}

func (c *CalendarTest) GetDriver() liveview.LiveDriver {
	return c
}

func main() {
	liveview.InitLogger(true)
	liveview.Info("Starting Calendar Test Server...")
	
	e := echo.New()
	e.Static("/assets", "assets")
	
	home := liveview.PageControl{
		Title:  "Calendar Test",
		Lang:   "en",
		Path:   "/",
		Router: e,
		Debug:  true,
	}
	
	home.Register(func() liveview.LiveDriver {
		liveview.Info("Creating new CalendarTest instance")
		return liveview.NewDriver("calendar_test", &CalendarTest{})
	})
	
	log.Println("==========================================")
	log.Println("   Calendar Component Test Server")
	log.Println("==========================================")
	log.Println("Server: http://localhost:8083")
	log.Println("==========================================")
	
	e.Logger.Fatal(e.Start(":8083"))
}