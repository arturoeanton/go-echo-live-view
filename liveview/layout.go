package liveview

import (
	"context"
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
	ctx                 context.Context    // MEM-002: Context para control de goroutines
	cancel              context.CancelFunc // MEM-002: Función para cancelar context
}

func (t *Layout) GetDriver() LiveDriver {
	return t
}

var (
	MuLayout sync.Mutex         = sync.Mutex{}
	Layouts  map[string]*Layout = make(map[string]*Layout)
)

func SendToAllLayouts(msg interface{}) {
	MuLayout.Lock()
	defer MuLayout.Unlock()
	wg := sync.WaitGroup{}
	for _, v := range Layouts {
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
			v, ok := Layouts[uid]
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
	
	// MEM-002: Crear context para control de goroutines
	ctx, cancel := context.WithCancel(context.Background())
	
	// MEM-001: Channel con buffer para evitar bloqueos
	c := &Layout{
		UUID:              uid,
		Html:              paramHtml,
		ChanIn:            make(chan interface{}, 10),
		IntervalEventTime: time.Hour * 24,
		ctx:               ctx,
		cancel:            cancel,
	}
	
	// MEM-003: Protección con mutex para acceso concurrente
	MuLayout.Lock()
	Layouts[uid] = c
	MuLayout.Unlock()
	
	fmt.Println("NewLayout", uid)
	c.ComponentDriver = NewDriver(uid, c)

	// MEM-002: Goroutine con context para cancelación controlada
	go func() {
		defer func() {
			// MEM-001: Cerrar channel al terminar
			close(c.ChanIn)
			// Limpiar de Layouts
			MuLayout.Lock()
			delete(Layouts, uid)
			MuLayout.Unlock()
		}()
		
		firstTime := true
		firstTimer := time.NewTimer(250 * time.Millisecond)
		eventTimer := time.NewTimer(c.IntervalEventTime)
		
		for {
			select {
			case <-ctx.Done():
				// Context cancelado, salir limpiamente
				return
				
			case data, ok := <-c.ChanIn:
				if !ok {
					return // Channel cerrado
				}
				if c.HandlerEventIn != nil {
					(*c.HandlerEventIn)(data)
				}
				
			case <-firstTimer.C:
				if firstTime {
					firstTime = false
					if c.HandlerFirstTime != nil {
						(*c.HandlerFirstTime)()
					} else {
						SendToAllLayouts("FIRST_TIME")
					}
				}
				
			case <-eventTimer.C:
				if c.HandlerEventTime != nil {
					(*c.HandlerEventTime)()
				}
				// Reiniciar timer
				eventTimer.Reset(c.IntervalEventTime)
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

// MEM-002: Método para limpiar y cancelar el Layout
func (t *Layout) Destroy() {
	if t.cancel != nil {
		t.cancel()
	}
	if t.HandlerEventDestroy != nil {
		(*t.HandlerEventDestroy)(t.UUID)
	}
}
func (t *Layout) Start() {
	t.Commit()
}

func (t *Layout) GetTemplate() string {
	return t.Html
}
