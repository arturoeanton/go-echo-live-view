package components

import (
	"time"
	"github.com/arturoeanton/go-echo-live-view/liveview"
)

type Calendar struct {
	*liveview.ComponentDriver[*Calendar]
	SelectedDate time.Time
	CurrentMonth time.Time
	MinDate      time.Time
	MaxDate      time.Time
	OnSelect     func(date time.Time)
}

func (c *Calendar) Start() {
	if c.CurrentMonth.IsZero() {
		c.CurrentMonth = time.Now()
	}
	if c.SelectedDate.IsZero() {
		c.SelectedDate = time.Now()
	}
	c.Commit()
}

func (c *Calendar) GetTemplate() string {
	return `
	<div id="{{.IdComponent}}" class="calendar">
		<style>
			.calendar { background: white; border-radius: 8px; box-shadow: 0 2px 8px rgba(0,0,0,0.1); padding: 1rem; width: 320px; }
			.calendar-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 1rem; }
			.calendar-title { font-size: 1.1rem; font-weight: 600; }
			.calendar-nav { display: flex; gap: 0.5rem; }
			.calendar-btn { padding: 0.5rem; background: #f5f5f5; border: none; border-radius: 4px; cursor: pointer; }
			.calendar-btn:hover { background: #e0e0e0; }
			.calendar-grid { display: grid; grid-template-columns: repeat(7, 1fr); gap: 2px; }
			.calendar-weekday { text-align: center; font-size: 0.75rem; font-weight: 600; color: #666; padding: 0.5rem; }
			.calendar-day { aspect-ratio: 1; display: flex; align-items: center; justify-content: center; cursor: pointer; border-radius: 4px; }
			.calendar-day:hover { background: #f5f5f5; }
			.calendar-day.other-month { color: #ccc; }
			.calendar-day.selected { background: #4CAF50; color: white; }
			.calendar-day.today { border: 2px solid #4CAF50; }
			.calendar-day.disabled { color: #ccc; cursor: not-allowed; }
		</style>
		
		<div class="calendar-header">
			<button class="calendar-btn" onclick="send_event('{{.IdComponent}}', 'PrevMonth', '')">‹</button>
			<div class="calendar-title">{{.GetCurrentMonthYear}}</div>
			<button class="calendar-btn" onclick="send_event('{{.IdComponent}}', 'NextMonth', '')">›</button>
		</div>
		
		<div class="calendar-grid">
			<div class="calendar-weekday">Sun</div>
			<div class="calendar-weekday">Mon</div>
			<div class="calendar-weekday">Tue</div>
			<div class="calendar-weekday">Wed</div>
			<div class="calendar-weekday">Thu</div>
			<div class="calendar-weekday">Fri</div>
			<div class="calendar-weekday">Sat</div>
			
			{{range .GetCalendarDays}}
			<div class="calendar-day {{.Classes}}" 
				onclick="{{if not .Disabled}}send_event('{{$.IdComponent}}', 'SelectDate', '{{.Date}}'){{end}}">
				{{.Day}}
			</div>
			{{end}}
		</div>
	</div>
	`
}

func (c *Calendar) GetDriver() liveview.LiveDriver {
	return c
}

func (c *Calendar) PrevMonth(data interface{}) {
	c.CurrentMonth = c.CurrentMonth.AddDate(0, -1, 0)
	c.Commit()
}

func (c *Calendar) NextMonth(data interface{}) {
	c.CurrentMonth = c.CurrentMonth.AddDate(0, 1, 0)
	c.Commit()
}

func (c *Calendar) SelectDate(data interface{}) {
	date, _ := time.Parse("2006-01-02", data.(string))
	c.SelectedDate = date
	if c.OnSelect != nil {
		c.OnSelect(date)
	}
	c.Commit()
}

func (c *Calendar) GetCurrentMonthYear() string {
	return c.CurrentMonth.Format("January 2006")
}

type CalendarDay struct {
	Day      int
	Date     string
	Classes  string
	Disabled bool
}

func (c *Calendar) GetCalendarDays() []CalendarDay {
	days := []CalendarDay{}
	
	year, month, _ := c.CurrentMonth.Date()
	firstDay := time.Date(year, month, 1, 0, 0, 0, 0, time.Local)
	lastDay := firstDay.AddDate(0, 1, -1)
	
	startOffset := int(firstDay.Weekday())
	for i := 0; i < startOffset; i++ {
		prevDate := firstDay.AddDate(0, 0, -(startOffset - i))
		days = append(days, CalendarDay{
			Day:     prevDate.Day(),
			Date:    prevDate.Format("2006-01-02"),
			Classes: "other-month",
		})
	}
	
	today := time.Now()
	for d := 1; d <= lastDay.Day(); d++ {
		date := time.Date(year, month, d, 0, 0, 0, 0, time.Local)
		day := CalendarDay{
			Day:  d,
			Date: date.Format("2006-01-02"),
		}
		
		if date.Format("2006-01-02") == c.SelectedDate.Format("2006-01-02") {
			day.Classes = "selected"
		}
		if date.Format("2006-01-02") == today.Format("2006-01-02") {
			day.Classes += " today"
		}
		
		if (!c.MinDate.IsZero() && date.Before(c.MinDate)) || (!c.MaxDate.IsZero() && date.After(c.MaxDate)) {
			day.Classes += " disabled"
			day.Disabled = true
		}
		
		days = append(days, day)
	}
	
	endPadding := 42 - len(days)
	for i := 1; i <= endPadding; i++ {
		nextDate := lastDay.AddDate(0, 0, i)
		days = append(days, CalendarDay{
			Day:     nextDate.Day(),
			Date:    nextDate.Format("2006-01-02"),
			Classes: "other-month",
		})
	}
	
	return days
}

func (c *Calendar) SetSelectedDate(date time.Time) {
	c.SelectedDate = date
	c.CurrentMonth = date
	c.Commit()
}

func (c *Calendar) GetSelectedDate() time.Time {
	return c.SelectedDate
}