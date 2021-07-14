package liveview

import (
	"bytes"
	"fmt"
	"log"
	"reflect"
	"text/template"

	"github.com/google/uuid"
)

type Component interface {
	GetTemplate() string
	Start()
}

type ComponentDriver struct {
	id                string
	IdComponent       string
	Component         Component
	channel           chan (map[string]interface{})
	componentsDrivers map[string]*ComponentDriver
	DriversPage       *map[string]*ComponentDriver
	channelIn         *map[string]chan interface{}
	Events            map[string]func(data interface{})
}

func (cw *ComponentDriver) Commit() {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered in Commit:", r)
		}
	}()
	t := template.Must(template.New("component").Funcs(FuncMapTemplate).Parse(cw.Component.GetTemplate()))
	buf := new(bytes.Buffer)
	err := t.Execute(buf, cw.Component)
	if err != nil {
		log.Println(err)
	}
	cw.FillValue(cw.id, buf.String())
}

func (cw *ComponentDriver) Start(drivers *map[string]*ComponentDriver, channelIn *map[string]chan interface{}, channel chan (map[string]interface{})) {
	cw.channel = channel
	cw.channelIn = channelIn
	cw.Component.Start()
	cw.DriversPage = drivers
	(*drivers)[cw.IdComponent] = cw
	for _, c := range cw.componentsDrivers {
		c.Start(drivers, channelIn, channel)
	}
}

func (cw *ComponentDriver) GetID() string {
	return cw.id
}

func (cw *ComponentDriver) SetID(id string) {
	cw.id = id
}

func (cw *ComponentDriver) Mount(componentDriver *ComponentDriver) *ComponentDriver {
	id := "mount_span_" + componentDriver.IdComponent
	componentDriver.SetID(id)
	cw.componentsDrivers[id] = componentDriver
	return cw
}

func NewDriver(id string, c Component) *ComponentDriver {
	driver := newDriver(c)
	driver.IdComponent = id
	ps := reflect.ValueOf(c)
	field := ps.Elem().FieldByName("Id")
	if field.CanSet() {
		field.SetString(id)
	}
	field = ps.Elem().FieldByName("Driver")
	if field.CanSet() {
		field.Set(reflect.ValueOf(driver))
	}

	return driver
}

func newDriver(c Component) *ComponentDriver {
	driver := &ComponentDriver{Component: c}
	driver.componentsDrivers = make(map[string]*ComponentDriver)
	driver.Events = make(map[string]func(interface{}))
	return driver
}

func (cw *ComponentDriver) ExecuteEvent(name string, data interface{}) {
	if cw == nil {
		return
	}
	go func(cw *ComponentDriver) {
		defer func() {
			if r := recover(); r != nil {
				log.Println("Recovered in ExecuteEvent:", r)
			}
		}()
		if data == nil {
			data = make(map[string]interface{})
		}

		if cw.Events != nil {
			if fx, ok := cw.Events[name]; ok {
				go fx(data)
				return
			}
		}
		in := []reflect.Value{reflect.ValueOf(data)}
		reflect.ValueOf(cw.Component).MethodByName(name).Call(in)

	}(cw)
}

func (cw *ComponentDriver) FillValue(id string, value string) {
	cw.channel <- map[string]interface{}{"type": "fill", "id": id, "value": value}
}

func (cw *ComponentDriver) SetHTML(id string, value string) {
	cw.channel <- map[string]interface{}{"type": "fill", "id": id, "value": value}
}

func (cw *ComponentDriver) SetText(id string, value string) {
	cw.channel <- map[string]interface{}{"type": "text", "id": id, "value": value}
}

func (cw *ComponentDriver) SetPropertie(id string, propertie string, value interface{}) {
	cw.channel <- map[string]interface{}{"type": "propertie", "id": id, "propertie": propertie, "value": value}
}

func (cw *ComponentDriver) SetValue(id string, value interface{}) {
	cw.channel <- map[string]interface{}{"type": "set", "id": id, "value": value}
}
func (cw *ComponentDriver) EvalScript(code string) {
	cw.channel <- map[string]interface{}{"type": "script", "value": code}
}

func (cw *ComponentDriver) SetStyle(id string, style string) {
	cw.channel <- map[string]interface{}{"type": "style", "id": id, "value": style}
}

func (cw *ComponentDriver) GetElementById(id string) string {
	return cw.get(id, "value", "")
}

func (cw *ComponentDriver) GetValue(id string) string {
	return cw.get(id, "value", "")
}

func (cw *ComponentDriver) GetStyle(id string, propertie string) string {
	return cw.get(id, "style", propertie)
}

func (cw *ComponentDriver) GetHTML(id string) string {
	return cw.get(id, "html", "")
}

func (cw *ComponentDriver) GetText(id string) string {
	return cw.get(id, "text", "")
}

func (cw *ComponentDriver) GetPropertie(id string, name string) string {
	return cw.get(id, "propertie", name)
}

func (cw *ComponentDriver) get(id string, subType string, value string) string {
	uid := uuid.NewString()
	(*cw.channelIn)[uid] = make(chan interface{})
	defer delete((*cw.channelIn), uid)
	cw.channel <- map[string]interface{}{"type": "get", "id": id, "value": value, "id_ret": uid, "sub_type": subType}
	data := <-(*cw.channelIn)[uid]
	if data != nil {
		return fmt.Sprint(data)
	}
	return ""
}
