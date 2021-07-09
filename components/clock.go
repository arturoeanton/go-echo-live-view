package components

import (
	"time"

	"github.com/arturoeanton/go-echo-live-view/liveview"
)

type Clock struct {
	Driver     *liveview.ComponentDriver
	ActualTime string
}

func (t *Clock) Start() {
	go func() {
		for {
			t.ActualTime = time.Now().Format(time.RFC3339Nano)
			t.Driver.Commit()
			time.Sleep((time.Second * 1) / 60)
		}
	}()
}

func (t *Clock) GetTemplate() string {
	return `
		<div  id="{{.Driver.IdComponent}}"  >
			<span>Time: {{ .ActualTime }}</span>
		</div>
	`
}
