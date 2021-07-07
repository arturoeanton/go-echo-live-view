package liveview

import (
	"bytes"
	"reflect"
	"text/template"

	"github.com/google/uuid"
)

type Component interface {
	GetTemplate() string
	Start()
	GetID() string
}

type ComponentDriver struct {
	id                string
	Component         Component
	channel           chan (map[string]string)
	componentsDrivers map[string]*ComponentDriver
	DriversPage       *map[string]*ComponentDriver
	channelIn         *map[string]chan interface{}
	Events            map[string]func(data interface{})
}

func (cw *ComponentDriver) Commit() {
	t := template.Must(template.New("component").Parse(cw.Component.GetTemplate()))
	buf := new(bytes.Buffer)
	_ = t.Execute(buf, cw.Component)
	cw.FillValue(cw.id, buf.String())
}

func (cw *ComponentDriver) Start(drivers *map[string]*ComponentDriver, channelIn *map[string]chan interface{}, channel chan (map[string]string)) {
	cw.channel = channel
	cw.channelIn = channelIn
	cw.Component.Start()
	cw.DriversPage = drivers
	(*drivers)[cw.Component.GetID()] = cw
	for _, c := range cw.componentsDrivers {
		c.Start(drivers, channelIn, channel)
	}
}

func (cw *ComponentDriver) FillValue(id string, data string) {
	cw.channel <- map[string]string{"type": "fill", "id": id, "value": data}
}

func (cw *ComponentDriver) SetValue(id string, data string) {
	cw.channel <- map[string]string{"type": "set", "id": id, "value": data}
}
func (cw *ComponentDriver) EvalScript(data string) {
	cw.channel <- map[string]string{"type": "script", "value": data}
}

func (cw *ComponentDriver) GetID() string {
	return cw.id
}

func (cw *ComponentDriver) SetID(id string) {
	cw.id = id
}

func (cw *ComponentDriver) Mount(id string, componentDriver *ComponentDriver) {
	componentDriver.SetID(id)
	cw.componentsDrivers[id] = componentDriver
}

func NewDriver(c Component) *ComponentDriver {
	driver := &ComponentDriver{Component: c}
	driver.componentsDrivers = make(map[string]*ComponentDriver)
	driver.Events = make(map[string]func(interface{}))
	return driver
}

func (cw *ComponentDriver) ExecuteEvent(name string, data interface{}) {
	go func() {
		if data == nil {
			data = make(map[string]interface{})
		}

		if fx, ok := cw.Events[name]; ok {
			go fx(data)
			return
		}

		in := []reflect.Value{reflect.ValueOf(data)}

		reflect.ValueOf(cw.Component).MethodByName(name).Call(in)

	}()
}

func (cw *ComponentDriver) GetElementById(name string) string {
	uid := uuid.NewString()
	(*cw.channelIn)[uid] = make(chan interface{})
	defer delete((*cw.channelIn), uid)
	cw.channel <- map[string]string{"type": "get", "id": name, "id_ret": uid}
	data := <-(*cw.channelIn)[uid]
	return data.(string)
}
