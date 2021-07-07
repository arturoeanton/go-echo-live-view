package components

import "github.com/arturoeanton/go-echo-live-view/liveview"

type None struct {
	driver *liveview.ComponentDriver
	Id     string
}

func NewNone(id string) *liveview.ComponentDriver {
	c := &None{Id: id}
	c.driver = liveview.NewDriver(c)
	return c.driver
}

func (t *None) Start() {
	t.driver.Commit()
}

func (t *None) GetTemplate() string {
	return ``
}

func (t *None) GetID() string {
	return t.Id
}
