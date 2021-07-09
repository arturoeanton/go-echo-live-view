package components

import (
	"github.com/arturoeanton/go-echo-live-view/liveview"
	"github.com/arturoeanton/gocommons/utils"
)

type Layout struct {
	Driver *liveview.ComponentDriver
	Html   string
}

func NewLayout(id string, html string) *liveview.ComponentDriver {
	if utils.Exists(html) {
		html, _ = utils.FileToString(html)
	}
	c := &Layout{Html: html}
	return liveview.NewDriver(id, c)
}

func (t *Layout) Start() {
	t.Driver.Commit()
}

func (t *Layout) GetTemplate() string {
	return t.Html
}
