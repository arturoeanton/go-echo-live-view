package components

import "github.com/arturoeanton/go-echo-live-view/liveview"

type Layout struct {
	driver *liveview.ComponentDriver
	Id     string
	Html   string
}

func NewLayout(id string, html string) *liveview.ComponentDriver {
	c := &Layout{Id: id, Html: html}
	c.driver = liveview.NewDriver(c)
	return c.driver
}

func (t *Layout) Start() {
	t.driver.Commit()
}

func (t *Layout) GetTemplate() string {
	return t.Html
}

func (t *Layout) GetID() string {
	return t.Id
}
