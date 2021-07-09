package components

import "github.com/arturoeanton/go-echo-live-view/liveview"

type None struct {
	Driver *liveview.ComponentDriver
}

func (t *None) Start() {
	t.Driver.Commit()
}

func (t *None) GetTemplate() string {
	return ``
}
