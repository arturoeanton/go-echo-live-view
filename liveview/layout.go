package liveview

import (
	"fmt"
	"sync"

	"github.com/arturoeanton/gocommons/utils"
	"github.com/google/uuid"
)

type Layout struct {
	*ComponentDriver[*Layout]
	Html    string
	ChanBus chan interface{}
	ChanIn  chan interface{}
	ChanOut chan interface{}
}

func (t *Layout) GetDriver() LiveDriver {
	return t
}

var (
	MuLayout sync.Mutex         = sync.Mutex{}
	Layaouts map[string]*Layout = make(map[string]*Layout)
)

func NewLayout(id string, html string) *ComponentDriver[*Layout] {
	if utils.Exists(html) {
		html, _ = utils.FileToString(html)
	}
	c := &Layout{Html: html, ChanBus: make(chan interface{}, 1), ChanIn: make(chan interface{}, 1)}
	uid := uuid.NewString()
	MuLayout.Lock()
	Layaouts[uid] = c
	MuLayout.Unlock()
	go func() {
		for {
			data := <-c.ChanBus
			for _, v := range Layaouts {
				v.ChanIn <- data
			}
		}
	}()
	fmt.Println("NewLayout", uid)
	c.ComponentDriver = NewDriver(uid, c)
	return c.ComponentDriver
}

func (t *Layout) Start() {
	t.Commit()
}

func (t *Layout) GetTemplate() string {
	return t.Html
}
