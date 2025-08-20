package liveview

import (
	"bytes"
	"context"
	"fmt"
	"reflect"
	"sync"
	"text/template"
	"time"

	"github.com/google/uuid"
)

var (
	// componentsDrivers stores all registered component drivers globally
	// Used for component lookup and management across the application
	componentsDrivers map[string]LiveDriver = make(map[string]LiveDriver)
	
	// mu protects concurrent access to componentsDrivers map
	mu sync.Mutex
)

// Component is the interface that all LiveView components must implement.
// It defines the core contract for creating interactive, server-rendered components.
type Component interface {
	// GetTemplate returns the HTML template string for rendering the component.
	// The template uses Go's text/template syntax with the component as context {{.}}
	GetTemplate() string
	
	// Start is called when the component is mounted and initialized.
	// Use this method to set initial state and perform setup logic.
	Start()
	
	// GetDriver returns the LiveDriver instance associated with this component.
	// The driver handles WebSocket communication and DOM updates.
	GetDriver() LiveDriver
}

// LiveDriver is the interface that manages component lifecycle, WebSocket communication,
// and DOM manipulation. It acts as the bridge between server-side components and client-side UI.
type LiveDriver interface {
	// GetID returns the DOM element ID where this component is rendered
	GetID() string
	
	// SetID sets the DOM element ID for this component
	SetID(string)
	
	// StartDriver initializes the driver with WebSocket channels and driver registry
	// Deprecated: Use StartDriverWithContext for better resource management
	StartDriver(*map[string]LiveDriver, *map[string]chan interface{}, chan (map[string]interface{}))
	
	// StartDriverWithContext initializes the driver with context support for cancellation
	// This is the preferred method for starting drivers with proper lifecycle management
	StartDriverWithContext(ctx context.Context, drivers *map[string]LiveDriver, channelIn *map[string]chan interface{}, channel chan map[string]interface{})
	
	// GetIDComponet returns the component's unique identifier
	GetIDComponet() string
	
	// ExecuteEvent triggers a named event on the component with optional data
	ExecuteEvent(name string, data interface{})

	// GetComponet returns the Component instance managed by this driver
	GetComponet() Component
	
	// Mount attaches a child component to this component
	Mount(component Component) LiveDriver
	
	// MountWithStart mounts and immediately starts a child component
	MountWithStart(id string, componentDriver LiveDriver) LiveDriver

	// Commit triggers a re-render of the component and sends updates to the client
	Commit()
	
	// Remove removes a DOM element by ID
	Remove(string)
	
	// AddNode adds a new DOM node with specified ID and HTML content
	AddNode(string, string)
	
	// FillValue updates the value of an input element
	FillValue(string)
	
	// SetHTML sets the innerHTML of an element
	SetHTML(string)
	
	// SetText sets the text content of an element
	SetText(string)
	
	// SetPropertie sets a property on a DOM element
	SetPropertie(string, interface{})
	
	// SetValue sets the component's value
	SetValue(interface{})
	
	// EvalScript executes JavaScript code on the client
	// Warning: Use with caution, prefer other DOM manipulation methods
	EvalScript(string)
	
	// SetStyle updates CSS styles on an element
	SetStyle(string)

	// FillValueById updates the value of a specific input element by ID
	FillValueById(id string, value string)

	// GetPropertie retrieves a property value from a DOM element
	GetPropertie(string) string
	
	// GetDriverById returns the driver instance for a specific component ID
	GetDriverById(id string) LiveDriver
	
	// GetText retrieves the text content of the component's root element
	GetText() string
	
	// GetHTML retrieves the HTML content of the component's root element
	GetHTML() string
	
	// GetStyle retrieves a CSS style value from the component's root element
	GetStyle(string) string
	
	// GetValue retrieves the value of the component's root element (for inputs)
	GetValue() string
	
	// GetElementById retrieves the HTML content of a specific element by ID
	GetElementById(string) string

	// SetData stores arbitrary data associated with the component
	SetData(interface{})
}

// SetData stores arbitrary data in the component driver.
// This data persists across renders and can be used for component state.
func (cw *ComponentDriver[T]) SetData(data interface{}) {
	cw.Data = data
}

// GetData retrieves the arbitrary data stored in the component driver
func (cw *ComponentDriver[T]) GetData() interface{} {
	return cw.Data
}

// ComponentDriver is the core driver implementation for LiveView components.
// It manages the component lifecycle, WebSocket communication, and DOM updates.
// The generic type T must implement the Component interface.
type ComponentDriver[T Component] struct {
	// Component is the actual component instance being managed
	Component T
	
	// id is the DOM element ID where this component is rendered
	id string
	
	// IdComponent is the unique identifier for this component instance
	IdComponent string
	
	// channel is used to send WebSocket messages to the client
	channel chan (map[string]interface{})
	
	// componentsDrivers stores child component drivers
	componentsDrivers map[string]LiveDriver
	
	// DriversPage is a reference to all drivers on the current page
	DriversPage *map[string]LiveDriver
	
	// channelIn handles incoming WebSocket messages
	channelIn *map[string]chan interface{}
	
	// Events maps event names to handler functions.
	// Allows dynamic registration of event handlers like click, change, keyup, etc.
	Events map[string]func(c T, data interface{})
	
	// Data stores arbitrary component data that persists across renders
	Data interface{}
	
	// errorBoundary provides error recovery for this component
	errorBoundary *ErrorBoundary
	
	// lifecycleCommit is an optional commit function with lifecycle support
	lifecycleCommit func()
}

// SetEvent registers a custom event handler for the component.
// The handler will be called when the client sends an event with the specified name.
// Example: SetEvent("CustomClick", func(c *MyComponent, data interface{}) { ... })
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
			Error("Recovered in Commit: %v", r)
		}
	}()
	
	LogComponent(cw.IdComponent, "Commit", "Starting")
	
	// SEC-003: Sanitizar template antes de parsear (deshabilitado temporalmente para desarrollo)
	rawTemplate := cw.Component.GetTemplate()
	// TODO: Habilitar sanitización en producción
	// sanitizedTemplate, err := SanitizeTemplate(rawTemplate)
	// if err != nil {
	// 	log.Printf("Template sanitization error: %v", err)
	// 	return
	// }
	
	t := template.Must(template.New("component").Funcs(FuncMapTemplate).Parse(rawTemplate))
	buf := new(bytes.Buffer)
	err := t.Execute(buf, cw.Component)
	if err != nil {
		Error("Template execution error: %v", err)
	}
	
	html := buf.String()
	LogTemplate(cw.IdComponent, "Rendered", fmt.Sprintf("%d bytes", len(html)))
	
	// Always use FillValueById for now
	// TODO: Implement proper mount preservation without JavaScript
	cw.FillValueById(cw.GetID(), html)
}

func (cw *ComponentDriver[T]) StartDriver(drivers *map[string]LiveDriver, channelIn *map[string]chan interface{}, channel chan (map[string]interface{})) {
	// MEM-002: Delegar a la versión con context
	cw.StartDriverWithContext(context.Background(), drivers, channelIn, channel)
}

func (cw *ComponentDriver[T]) StartDriverWithContext(ctx context.Context, drivers *map[string]LiveDriver, channelIn *map[string]chan interface{}, channel chan map[string]interface{}) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in StartDriverWithContext", r)
		}
	}()
	
	// MEM-003: Proteger acceso concurrente con mutex
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
			// MEM-002: Propagar context a componentes hijos
			c.StartDriverWithContext(ctx, drivers, channelIn, channel)
		}(c)
	}
	
	// MEM-002: Esperar con timeout usando context
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	
	select {
	case <-done:
		// Todos los drivers iniciados correctamente
	case <-ctx.Done():
		// Context cancelado, detener espera
		Warn("Context cancelled while starting drivers for %s", cw.GetIDComponet())
	case <-time.After(30 * time.Second):
		// Timeout de seguridad
		Error("Timeout starting drivers for %s", cw.GetIDComponet())
	}
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

// Mount mount component in other component
func (cw *ComponentDriver[T]) MountWithStart(id string, componentDriver LiveDriver) LiveDriver {
	componentDriver.SetID(id)
	cw.componentsDrivers[id] = componentDriver
	// MEM-002: Usar context al montar componente
	ctx := context.Background()
	componentDriver.StartDriverWithContext(ctx, cw.DriversPage, cw.channelIn, cw.channel)
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
			method := reflect.ValueOf(cw.Component).MethodByName(name)
			if method.IsValid() {
				in := []reflect.Value{reflect.ValueOf(data)}
				method.Call(in)
			}
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
// DEPRECATED: This method is deprecated and will be removed in the next major version.
// Use ExecuteSafeScript or ExecutePredefinedAction instead for secure script execution.
func (cw *ComponentDriver[T]) EvalScript(code string) {
	cw.DeprecatedEvalScript(code)
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
	// MEM-001: Crear channel con buffer para evitar bloqueos
	ch := make(chan interface{}, 1)
	(*cw.channelIn)[uid] = ch
	// MEM-001: Asegurar limpieza del channel
	defer func() {
		delete((*cw.channelIn), uid)
		close(ch)
	}()
	
	cw.channel <- map[string]interface{}{"type": "get", "id": id, "value": value, "id_ret": uid, "sub_type": subType}
	
	// MEM-004: Agregar timeout para evitar bloqueo indefinido
	select {
	case data := <-ch:
		if data != nil {
			return fmt.Sprint(data)
		}
	case <-time.After(5 * time.Second):
		Warn("Timeout waiting for response in get() for id: %s", id)
	}
	return ""
}
