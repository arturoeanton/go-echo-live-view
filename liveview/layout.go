package liveview

import (
	"fmt"
	"sync"
	"time"

	"github.com/arturoeanton/gocommons/utils"
	"github.com/google/uuid"
)

type Layout struct {
	*ComponentDriver[*Layout]
	UUID              string
	Html              string
	ChanIn            chan interface{}
	HandlerEventIn    *func(data interface{})
	HandlerEventTime  *func()
	IntervalEventTime time.Duration
}

func (t *Layout) GetDriver() LiveDriver {
	return t
}

var (
	MuLayout sync.Mutex         = sync.Mutex{}
	Layaouts map[string]*Layout = make(map[string]*Layout)
)

func SendToAllLayouts(msg interface{}) {
	MuLayout.Lock()
	defer MuLayout.Unlock()
	wg := sync.WaitGroup{}
	for _, v := range Layaouts {
		wg.Add(1)
		go func(v *Layout) {
			defer wg.Done()
			v.ChanIn <- msg
		}(v)
	}
	wg.Wait()
}

func SendToLayouts(msg interface{}, uuids ...string) {
	MuLayout.Lock()
	defer MuLayout.Unlock()
	wg := sync.WaitGroup{}
	for _, uid := range uuids {
		wg.Add(1)
		go func(uid string) {
			defer wg.Done()
			v, ok := Layaouts[uid]
			if ok {
				v.ChanIn <- msg
			}
		}(uid)
	}
	wg.Wait()
}

func NewLayout(html string) *ComponentDriver[*Layout] {
	if utils.Exists(html) {
		html, _ = utils.FileToString(html)
	}
	uid := uuid.NewString()
	c := &Layout{UUID: uid, Html: html, ChanIn: make(chan interface{}, 1), IntervalEventTime: time.Hour * 24}
	MuLayout.Lock()
	Layaouts[uid] = c
	MuLayout.Unlock()
	fmt.Println("NewLayout", uid)
	c.ComponentDriver = NewDriver(uid, c)

	go func() {
		for {
			select {
			case data := <-c.Component.ChanIn:
				if c.HandlerEventIn != nil {
					(*c.HandlerEventIn)(data)
				}
			case <-time.After(c.IntervalEventTime):
				if c.HandlerEventTime != nil {
					(*c.HandlerEventTime)()
				}
			}
		}
	}()

	return c.ComponentDriver
}

func (t *Layout) SetHandlerEventIn(fx func(data interface{})) {
	t.HandlerEventIn = &fx
}

func (t *Layout) SetHandlerEventTime(IntervalEventTime time.Duration, fx func()) {
	t.IntervalEventTime = IntervalEventTime
	t.HandlerEventTime = &fx
}

func (t *Layout) Start() {
	t.Commit()
}

func (t *Layout) GetTemplate() string {
	return t.Html
}
