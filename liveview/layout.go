package liveview

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/html"
)

type Layout struct {
	*ComponentDriver[*Layout]
	UUID                string
	Html                string
	ChanIn              chan interface{}
	HandlerEventIn      *func(data interface{})
	HandlerEventTime    *func()
	HandlerEventDestroy *func(id string)
	HandlerFirstTime    *func()
	IntervalEventTime   time.Duration
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

func NewLayout(uid string, paramHtml string) *ComponentDriver[*Layout] {
	if Exists(paramHtml) {
		paramHtml, _ = FileToString(paramHtml)
	}
	c := &Layout{UUID: uid, Html: paramHtml, ChanIn: make(chan interface{}, 1), IntervalEventTime: time.Hour * 24}
	MuLayout.Lock()
	Layaouts[uid] = c
	MuLayout.Unlock()
	fmt.Println("NewLayout", uid)
	c.ComponentDriver = NewDriver(uid, c)

	go func() {
		firstTiem := true
		for {
			select {
			case data := <-c.Component.ChanIn:
				if c.HandlerEventIn != nil {
					(*c.HandlerEventIn)(data)
				}
			case <-time.After(250 * time.Millisecond):
				if c.HandlerFirstTime != nil {
					if firstTiem {
						firstTiem = false
						(*c.HandlerFirstTime)()
					}
				} else {
					if firstTiem {
						firstTiem = false
						SendToAllLayouts("FIRST_TIME")
					}
				}
			case <-time.After(c.IntervalEventTime):
				if c.HandlerEventTime != nil {
					(*c.HandlerEventTime)()
				}
			}
		}
	}()

	doc, err := html.Parse(strings.NewReader(paramHtml))
	if err != nil {
		fmt.Println(err)
	}
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode {
			for _, a := range n.Attr {
				if a.Key == "id" {
					Join(a.Val)
					break
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	return c.ComponentDriver
}

func (t *Layout) SetHandlerFirstTime(fx func()) {
	t.HandlerFirstTime = &fx
}
func (t *Layout) SetHandlerEventIn(fx func(data interface{})) {
	t.HandlerEventIn = &fx
}

func (t *Layout) SetHandlerEventTime(IntervalEventTime time.Duration, fx func()) {
	t.IntervalEventTime = IntervalEventTime
	t.HandlerEventTime = &fx
}

func (t *Layout) SetHandlerEventDestroy(fx func(id string)) {
	t.HandlerEventDestroy = &fx
}
func (t *Layout) Start() {
	t.Commit()
}

func (t *Layout) GetTemplate() string {
	return t.Html
}
