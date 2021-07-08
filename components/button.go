package components

import (
	"github.com/arturoeanton/go-echo-live-view/liveview"
)

type Button struct {
	Driver  *liveview.ComponentDriver
	Id      string
	I       int
	Caption string
}

func NewButton(id string, caption string) *liveview.ComponentDriver {
	c := &Button{Id: id, Caption: caption}
	c.Driver = liveview.NewDriver(c)
	return c.Driver
}

func (t *Button) Start() {
	t.Driver.Commit()
}

func (t *Button) GetTemplate() string {

	return `<Button id="{{.Id}}" onclick="send_event(this.id,'Click')" >{{.Caption}}</button>`
}

func (t *Button) Click(data interface{}) {
}

func (t *Button) GetID() string {
	return t.Id
}
