package components

import (
	"github.com/arturoeanton/go-echo-live-view/liveview"
	"github.com/arturoeanton/gocommons/utils"
)

type Layout struct {
	*liveview.ComponentDriver[*Layout]
	Html string
}

func (t *Layout) GetDriver() liveview.LiveDriver {
	return t
}
func NewLayout(id string, html string) *liveview.ComponentDriver[*Layout] {
	if utils.Exists(html) {
		html, _ = utils.FileToString(html)
	}
	c := &Layout{Html: html}
	c.ComponentDriver = liveview.NewDriver(id, c)
	return c.ComponentDriver
}

func (t *Layout) Start() {
	t.Commit()
}

func (t *Layout) GetTemplate() string {
	return t.Html
}
