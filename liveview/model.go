package liveview

import (
	"bytes"
	"fmt"
	"log"
	"reflect"
	"text/template"

	"github.com/google/uuid"
)

// Component it is interface for implement one component
type Component interface {
	// GetTemplate return html template for render with component in the {{.}}
	GetTemplate() string
	// Start it will invoke in the mount time
	Start()
}

// ComponentDriver this is the driver for component, with this struct we can execute our methods in the web
type ComponentDriver struct {
	id                string
	IdComponent       string
	Component         Component
	channel           chan (map[string]interface{})
	componentsDrivers map[string]*ComponentDriver
	DriversPage       *map[string]*ComponentDriver
	channelIn         *map[string]chan interface{}
	// Events has rewrite of our implementings of  events, examples click, change, keyup, keydown, etc
	Events map[string]func(data interface{})
}

// Commit render of component
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

// GetID return id of driver
func (cw *ComponentDriver) GetID() string {
	return cw.id
}

// SetID set id of driver
func (cw *ComponentDriver) SetID(id string) {
	cw.id = id
}

// Mount mount component in other component
func (cw *ComponentDriver) Mount(componentDriver *ComponentDriver) *ComponentDriver {
	id := "mount_span_" + componentDriver.IdComponent
	componentDriver.SetID(id)
	cw.componentsDrivers[id] = componentDriver
	return cw
}

// Mount mount component in other component
func (cw *ComponentDriver) MountWithStart(id string, componentDriver *ComponentDriver) *ComponentDriver {
	componentDriver.SetID(id)
	cw.componentsDrivers[id] = componentDriver
	componentDriver.Start(cw.DriversPage, cw.channelIn, cw.channel)
	return cw
}

// Create Driver with component
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

// ExecuteEvent execute events
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

//Remove
func (cw *ComponentDriver) Remove(id string) {
	cw.channel <- map[string]interface{}{"type": "remove", "id": id}
}

//AddNode add node to id
func (cw *ComponentDriver) AddValue(id string, value string) {
	cw.channel <- map[string]interface{}{"type": "addNode", "id": id, "value": value}
}

//FillValue is same SetHTML
func (cw *ComponentDriver) FillValue(id string, value string) {
	cw.channel <- map[string]interface{}{"type": "fill", "id": id, "value": value}
}

//SetHTML is same FillValue :p haha, execute  document.getElementById("$id").innerHTML = $value
func (cw *ComponentDriver) SetHTML(id string, value string) {
	cw.channel <- map[string]interface{}{"type": "fill", "id": id, "value": value}
}

//SetText execute document.getElementById("$id").innerText = $value
func (cw *ComponentDriver) SetText(id string, value string) {
	cw.channel <- map[string]interface{}{"type": "text", "id": id, "value": value}
}

//SetPropertie execute  document.getElementById("$id")[$propertie] = $value
func (cw *ComponentDriver) SetPropertie(id string, propertie string, value interface{}) {
	cw.channel <- map[string]interface{}{"type": "propertie", "id": id, "propertie": propertie, "value": value}
}

//SetValue execute document.getElementById("$id").value = $value|
func (cw *ComponentDriver) SetValue(id string, value interface{}) {
	cw.channel <- map[string]interface{}{"type": "set", "id": id, "value": value}
}

//EvalScript execute eval($code);
func (cw *ComponentDriver) EvalScript(code string) {
	cw.channel <- map[string]interface{}{"type": "script", "value": code}
}

//SetStyle execute  document.getElementById("$id").style.cssText = $style
func (cw *ComponentDriver) SetStyle(id string, style string) {
	cw.channel <- map[string]interface{}{"type": "style", "id": id, "value": style}
}

//GetElementById same as GetValue
func (cw *ComponentDriver) GetElementById(id string) string {
	return cw.get(id, "value", "")
}

//GetValue return document.getElementById("$id").value
func (cw *ComponentDriver) GetValue(id string) string {
	return cw.get(id, "value", "")
}

//GetStyle  return document.getElementById("$id").style["$propertie"]
func (cw *ComponentDriver) GetStyle(id string, propertie string) string {
	return cw.get(id, "style", propertie)
}

//GetHTML  return document.getElementById("$id").innerHTML
func (cw *ComponentDriver) GetHTML(id string) string {
	return cw.get(id, "html", "")
}

//GetText  return document.getElementById("$id").innerText
func (cw *ComponentDriver) GetText(id string) string {
	return cw.get(id, "text", "")
}

//GetPropertie return document.getElementById("$id")[$propertie]
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
