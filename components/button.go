package components

import (
	"github.com/arturoeanton/go-echo-live-view/liveview"
)

type Button struct {
	Driver  *liveview.ComponentDriver
	I       int
	Caption string
}

func (t *Button) Start() {
	t.Driver.Commit()
}

func (t *Button) GetTemplate() string {

	return `<Button id="{{.Driver.IdComponent}}" onclick="send_event(this.id,'Click')" >{{.Caption}}</button>`
}

func (t *Button) Click(data interface{}) {
}
