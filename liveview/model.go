package liveview

import (
	"bytes"
	"fmt"
	"log"
	"reflect"
	"sync"
	"text/template"

	"github.com/google/uuid"
)

var (
	componentsDrivers map[string]LiveDriver = make(map[string]LiveDriver)
	mu                sync.Mutex
)

// Component it is interface for implement one component
type Component interface {
	// GetTemplate return html template for render with component in the {{.}}
	GetTemplate() string
	// Start it will invoke in the mount time
	Start()
	GetDriver() LiveDriver
}

type LiveDriver interface {
	GetID() string
	SetID(string)
	StartDriver(*map[string]LiveDriver, *map[string]chan interface{}, chan (map[string]interface{}))
	GetIDComponet() string
	ExecuteEvent(name string, data interface{})

	GetComponet() Component
	Mount(component Component) LiveDriver
	MountWithStart(id string, componentDriver LiveDriver) LiveDriver

	Commit()
	Remove(string)
	AddNode(string, string)
	FillValue(string)
	SetHTML(string)
	SetText(string)
	SetPropertie(string, interface{})
	SetValue(interface{})
	EvalScript(string)
	SetStyle(string)

	FillValueById(id string, value string)

	GetPropertie(string) string
	GetDriverById(id string) LiveDriver
	GetText() string
	GetHTML() string
	GetStyle(string) string
	GetValue() string
	GetElementById(string) string

	SetData(interface{})
}

func (cw *ComponentDriver[T]) SetData(data interface{}) {
	cw.Data = data
}

func (cw *ComponentDriver[T]) GetData() interface{} {
	return cw.Data
}

// ComponentDriver this is the driver for component, with this struct we can execute our methods in the web
type ComponentDriver[T Component] struct {
	Component         T
	id                string
	IdComponent       string
	channel           chan (map[string]interface{})
	componentsDrivers map[string]LiveDriver
	DriversPage       *map[string]LiveDriver
	channelIn         *map[string]chan interface{}
	// Events has rewrite of our implementings of  events, examples click, change, keyup, keydown, etc
	Events map[string]func(c T, data interface{})
	Data   interface{}
}

func (cw *ComponentDriver[T]) SetEvent(name string, fx func(c T, data interface{})) {
	cw.Events[name] = fx
}

func (cw *ComponentDriver[T]) GetIDComponet() string {
	return cw.IdComponent
}

// Commit render of component
func (cw *ComponentDriver[T]) Commit() {
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
	cw.FillValueById(cw.GetID(), buf.String())
}

func (cw *ComponentDriver[T]) StartDriver(drivers *map[string]LiveDriver, channelIn *map[string]chan interface{}, channel chan (map[string]interface{})) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in f", r)
		}
	}()
	cw.channel = channel
	cw.channelIn = channelIn
	cw.Component.Start()
	cw.DriversPage = drivers
	mu.Lock()
	(*drivers)[cw.GetIDComponet()] = cw
	mu.Unlock()
	var wg sync.WaitGroup
	for _, c := range cw.componentsDrivers {
		wg.Add(1)
		go func(c LiveDriver) {
			defer HandleReover()
			defer wg.Done()
			c.StartDriver(drivers, channelIn, channel)
		}(c)
	}
	wg.Wait()
}

// GetID return id of driver
func (cw *ComponentDriver[T]) GetComponet() Component {
	return cw.Component
}

// GetID return id of driver
func (cw *ComponentDriver[T]) GetDriverById(id string) LiveDriver {
	if c, ok := cw.componentsDrivers["mount_span_"+id]; ok {
		return c
	}
	if c, ok := cw.componentsDrivers[id]; ok {
		return c
	}
	c := &None{}
	New(id, c)
	return c
}

// GetID return id of driver
func (cw *ComponentDriver[T]) GetID() string {
	return cw.id
}

// SetID set id of driver
func (cw *ComponentDriver[T]) SetID(id string) {
	cw.id = id
}

// Mount mount component in other component
func (cw *ComponentDriver[T]) Mount(component Component) LiveDriver {
	componentDriver := component.GetDriver()
	id := "mount_span_" + componentDriver.GetIDComponet()
	componentDriver.SetID(id)
	cw.componentsDrivers[id] = componentDriver
	return cw
}

// Mount mount component in other component"mount_span_" +
func (cw *ComponentDriver[T]) MountWithStart(id string, componentDriver LiveDriver) LiveDriver {
	componentDriver.SetID(id)
	cw.componentsDrivers[id] = componentDriver
	componentDriver.StartDriver(cw.DriversPage, cw.channelIn, cw.channel)
	return cw
}

func Join(ids ...string) {
	for _, id := range ids {
		New(id, &None{})
	}
}

func New[T Component](id string, c T) T {
	NewDriver(id, c)
	componentDriver := c.GetDriver()
	idMount := "mount_span_" + componentDriver.GetIDComponet()
	componentDriver.SetID(idMount)
	componentsDrivers[idMount] = componentDriver
	return c
}

func NewWithTemplate(id string, template string) *None {
	return New(id, &None{Template: template})
}

// Create Driver with component
func NewDriver[T Component](id string, c T) *ComponentDriver[T] {
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
	} else {
		field = ps.Elem().FieldByName("ComponentDriver")
		if field.CanSet() {
			field.Set(reflect.ValueOf(driver))
		}
	}
	return driver
}

func newDriver[T Component](c T) *ComponentDriver[T] {
	driver := &ComponentDriver[T]{Component: c}
	driver.componentsDrivers = make(map[string]LiveDriver)
	driver.Events = make(map[string]func(T, interface{}))
	return driver
}

// ExecuteEvent execute events
func (cw *ComponentDriver[T]) ExecuteEvent(name string, data interface{}) {
	if cw == nil {
		return
	}
	go func(cw *ComponentDriver[T]) {
		defer HandleReover()
		if data == nil {
			data = make(map[string]interface{})
		}

		if cw.Events != nil {
			if fx, ok := cw.Events[name]; ok {
				go func() {
					defer HandleReover()
					fx(cw.Component, data)
				}()
				return
			}
		}
		func() {
			defer HandleReoverPass()
			in := []reflect.Value{reflect.ValueOf(data)}
			reflect.ValueOf(cw.Component).MethodByName(name).Call(in)
		}()

	}(cw)
}

// Remove
func (cw *ComponentDriver[T]) Remove(id string) {
	cw.channel <- map[string]interface{}{"type": "remove", "id": id}
}

// AddNode add node to id
func (cw *ComponentDriver[T]) AddNode(id string, value string) {
	cw.channel <- map[string]interface{}{"type": "addNode", "id": id, "value": value}
}

// FillValue is same SetHTML
func (cw *ComponentDriver[T]) FillValueById(id string, value string) {
	cw.channel <- map[string]interface{}{"type": "fill", "id": id, "value": value}
}

// FillValue is same SetHTML
func (cw *ComponentDriver[T]) FillValue(value string) {
	cw.channel <- map[string]interface{}{"type": "fill", "id": cw.GetIDComponet(), "value": value}
}

// SetHTML is same FillValue :p haha, execute  document.getElementById("$id").innerHTML = $value
func (cw *ComponentDriver[T]) SetHTML(value string) {
	cw.channel <- map[string]interface{}{"type": "fill", "id": cw.GetIDComponet(), "value": value}
}

// SetText execute document.getElementById("$id").innerText = $value
func (cw *ComponentDriver[T]) SetText(value string) {
	cw.channel <- map[string]interface{}{"type": "text", "id": cw.GetIDComponet(), "value": value}
}

// SetPropertie execute  document.getElementById("$id")[$propertie] = $value
func (cw *ComponentDriver[T]) SetPropertie(propertie string, value interface{}) {
	cw.channel <- map[string]interface{}{"type": "propertie", "id": cw.GetIDComponet(), "propertie": propertie, "value": value}
}

// SetValue execute document.getElementById("$id").value = $value|
func (cw *ComponentDriver[T]) SetValue(value interface{}) {
	cw.channel <- map[string]interface{}{"type": "set", "id": cw.GetIDComponet(), "value": value}
}

// EvalScript execute eval($code);
func (cw *ComponentDriver[T]) EvalScript(code string) {
	cw.channel <- map[string]interface{}{"type": "script", "value": code}
}

// SetStyle execute  document.getElementById("$id").style.cssText = $style
func (cw *ComponentDriver[T]) SetStyle(style string) {
	cw.channel <- map[string]interface{}{"type": "style", "id": cw.GetIDComponet(), "value": style}
}

// GetElementById same as GetValue
func (cw *ComponentDriver[T]) GetElementById(id string) string {
	return cw.get(id, "value", "")
}

// GetValue return document.getElementById("$id").value
func (cw *ComponentDriver[T]) GetValue() string {
	return cw.get(cw.GetIDComponet(), "value", "")
}

// GetStyle  return document.getElementById("$id").style["$propertie"]
func (cw *ComponentDriver[T]) GetStyle(propertie string) string {
	return cw.get(cw.GetIDComponet(), "style", propertie)
}

// GetHTML  return document.getElementById("$id").innerHTML
func (cw *ComponentDriver[T]) GetHTML() string {
	return cw.get(cw.GetIDComponet(), "html", "")
}

// GetText  return document.getElementById("$id").innerText
func (cw *ComponentDriver[T]) GetText() string {
	return cw.get(cw.GetIDComponet(), "text", "")
}

// GetPropertie return document.getElementById("$id")[$propertie]
func (cw *ComponentDriver[T]) GetPropertie(name string) string {
	return cw.get(cw.GetIDComponet(), "propertie", name)
}

func (cw *ComponentDriver[T]) get(id string, subType string, value string) string {
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
