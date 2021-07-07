package components

import (
	"time"

	"github.com/arturoeanton/go-echo-live-view/liveview"
)

type Clock struct {
	driver     *liveview.ComponentDriver
	ActualTime string
	Id         string
}

func NewClock(id string) *liveview.ComponentDriver {
	c := &Clock{Id: id}
	c.driver = liveview.NewDriver(c)
	return c.driver
}

func (t *Clock) Start() {
	go func() {
		for {
			t.ActualTime = time.Now().Format(time.RFC3339Nano)
			t.driver.Commit()
			time.Sleep((time.Second * 1) / 60)
		}
	}()
}

func (t *Clock) GetTemplate() string {
	return `
		<div id="{{.Id}}" >
			<span>Time: {{ .ActualTime }}</span>
		</div>
	`
}

func (t *Clock) GetID() string { return t.Id }
