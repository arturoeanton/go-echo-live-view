# API Documentation / Documentación API

[English](#english) | [Español](#español)

---

## English

# Go Echo LiveView - API Reference

Go Echo LiveView is a real-time web framework for Go that enables server-side rendering with WebSocket-based reactivity, inspired by Phoenix LiveView.

## Table of Contents

1. [Core Concepts](#core-concepts)
2. [Getting Started](#getting-started)
3. [Components API](#components-api)
4. [LiveDriver Interface](#livedriver-interface)
5. [Page Control](#page-control)
6. [Event Handling](#event-handling)
7. [WebSocket Communication](#websocket-communication)
8. [Component Lifecycle](#component-lifecycle)
9. [Built-in Components](#built-in-components)
10. [Enhanced Flow Tool](#enhanced-flow-tool)
11. [Drag and Drop Support](#drag-and-drop-support)
12. [Security Features](#security-features)
13. [Best Practices](#best-practices)

## Core Concepts

### Component-Based Architecture

LiveView uses a component-based architecture where each UI element is a self-contained component with its own state and lifecycle.

```go
type Component interface {
    GetTemplate() string    // Returns HTML template
    Start()                // Component initialization
    GetDriver() LiveDriver // Returns the component driver
}
```

### Real-time Updates

Components automatically sync with the client via WebSocket, eliminating the need for manual AJAX calls or client-side state management.

## Getting Started

### Installation

```bash
go get github.com/arturoeanton/go-echo-live-view
```

### Building WASM Module

The framework includes a WASM module for enhanced client-side functionality:

```bash
cd cmd/wasm/
GOOS=js GOARCH=wasm go build -o ../../assets/json.wasm
```

### Basic Example

```go
package main

import (
    "github.com/arturoeanton/go-echo-live-view/liveview"
    "github.com/labstack/echo/v4"
)

type Counter struct {
    *liveview.ComponentDriver[*Counter]
    Count int
}

func (c *Counter) Start() {
    c.Count = 0
    c.Commit()
}

func (c *Counter) GetTemplate() string {
    return `
    <div>
        <h1>Count: {{.Count}}</h1>
        <button onclick="send_event('{{.IdComponent}}', 'Increment')">+</button>
        <button onclick="send_event('{{.IdComponent}}', 'Decrement')">-</button>
    </div>
    `
}

func (c *Counter) GetDriver() liveview.LiveDriver {
    return c
}

func (c *Counter) Increment(data interface{}) {
    c.Count++
    c.Commit()
}

func (c *Counter) Decrement(data interface{}) {
    c.Count--
    c.Commit()
}

func main() {
    e := echo.New()
    
    page := liveview.PageControl{
        Title:  "Counter Example",
        Path:   "/",
        Router: e,
    }
    
    page.Register(func() liveview.LiveDriver {
        counter := &Counter{}
        driver := liveview.NewDriver("counter", counter)
        counter.ComponentDriver = driver
        return driver
    })
    
    e.Logger.Fatal(e.Start(":8080"))
}
```

## Components API

### Creating a Component

Components must implement the `Component` interface:

```go
type MyComponent struct {
    *liveview.ComponentDriver[*MyComponent]
    // Component state fields
    Message string
    Count   int
}

// Initialize component
func (c *MyComponent) Start() {
    c.Message = "Hello World"
    c.Commit() // Trigger re-render
}

// Define HTML template
func (c *MyComponent) GetTemplate() string {
    return `<div>{{.Message}}</div>`
}

// Return driver reference
func (c *MyComponent) GetDriver() liveview.LiveDriver {
    return c
}
```

### Component Methods

#### `Commit()`
Triggers a re-render of the component and sends updates to the client.

```go
func (c *MyComponent) UpdateMessage(data interface{}) {
    c.Message = "Updated!"
    c.Commit() // Send update to client
}
```

#### `Mount(component)`
Mounts a child component within the current component.

```go
func (c *ParentComponent) Start() {
    child := &ChildComponent{}
    c.Mount(liveview.New("child", child))
    c.Commit()
}
```

## LiveDriver Interface

The `LiveDriver` interface manages component lifecycle and WebSocket communication:

```go
type LiveDriver interface {
    GetID() string
    SetID(string)
    StartDriver(*map[string]LiveDriver, *map[string]chan interface{}, chan (map[string]interface{}))
    StartDriverWithContext(ctx context.Context, drivers *map[string]LiveDriver, channelIn *map[string]chan interface{}, channel chan map[string]interface{})
    GetIDComponet() string
    ExecuteEvent(name string, data interface{})
    GetComponet() Component
    Mount(component Component) LiveDriver
    // DOM manipulation methods
    FillValue(id, value string)
    SetHTML(id, html string)
    SetText(id, text string)
    SetStyle(id, property, value string)
    SetClass(id, class string)
    EvalScript(js string)
}
```

### DOM Manipulation Methods

#### `FillValue(id, value string)`
Updates the value of an input element.

```go
c.FillValue("username-input", "john_doe")
```

#### `SetHTML(id, html string)`
Sets the innerHTML of an element.

```go
c.SetHTML("content", "<p>New content</p>")
```

#### `SetText(id, text string)`
Sets the text content of an element.

```go
c.SetText("status", "Connected")
```

#### `SetStyle(id, property, value string)`
Updates CSS style properties.

```go
c.SetStyle("panel", "background-color", "#f0f0f0")
```

## Page Control

The `PageControl` struct manages page routing and WebSocket setup:

```go
type PageControl struct {
    Path      string              // URL path
    Title     string              // Page title
    HeadCode  string              // Additional <head> content
    Lang      string              // Page language
    Router    *echo.Echo          // Echo router instance
    Debug     bool                // Enable debug mode
}
```

### Registering Components

```go
page := liveview.PageControl{
    Title:  "My App",
    Path:   "/dashboard",
    Router: e,
    Debug:  true,
}

page.Register(func() liveview.LiveDriver {
    component := &MyComponent{}
    driver := liveview.NewDriver("my-component", component)
    component.ComponentDriver = driver
    return driver
})
```

## Event Handling

### Client-to-Server Events

Events are sent from the client using the `send_event` function:

```html
<button onclick="send_event('{{.IdComponent}}', 'Click')">Click Me</button>
<input onchange="send_event('{{.IdComponent}}', 'Change', this.value)">
```

### Server-Side Event Handlers

Event handlers are methods on your component that match the event name:

```go
func (c *MyComponent) Click(data interface{}) {
    // Handle click event
    c.Commit()
}

func (c *MyComponent) Change(data interface{}) {
    if value, ok := data.(string); ok {
        c.Value = value
        c.Commit()
    }
}
```

### Custom Event Registration

```go
func (c *MyComponent) Start() {
    c.Events["CustomEvent"] = func(comp *MyComponent, data interface{}) {
        // Handle custom event
    }
}
```

## WebSocket Communication

### Message Format

WebSocket messages are JSON-encoded with the following structure:

```json
{
    "type": "data|get|fill|script",
    "id": "component_id",
    "event": "EventName",
    "data": {},
    "value": "content"
}
```

### Message Types

- **`data`**: Component event message
- **`get`**: Request component property
- **`fill`**: Update DOM element
- **`script`**: Execute JavaScript

### Rate Limiting

WebSocket connections include built-in rate limiting:

```go
rateLimiter := liveview.NewRateLimiter(100, 60) // 100 messages per minute
```

## Component Lifecycle

### Lifecycle Hooks

1. **Creation**: Component instance created
2. **Driver Assignment**: `ComponentDriver` assigned
3. **Start()**: Component initialization
4. **Mount**: Child components mounted
5. **Render**: Initial template render
6. **Event Loop**: Handle events and updates
7. **Destroy**: Cleanup (if implemented)

### Context Management

Components support context-based lifecycle management:

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

component.StartDriverWithContext(ctx, &drivers, &channelIn, channel)
```

## Built-in Components

### Table Component

```go
table := &components.Table{
    Columns: []components.Column{
        {Key: "id", Title: "ID", Sortable: true},
        {Key: "name", Title: "Name", Sortable: true},
    },
    Rows: []components.Row{
        {"id": 1, "name": "John"},
        {"id": 2, "name": "Jane"},
    },
    ShowPagination: true,
    PageSize: 10,
}
```

### Form Component

```go
form := &components.Form{
    Fields: []components.FormField{
        {
            Name:     "email",
            Label:    "Email",
            Type:     "email",
            Required: true,
        },
    },
    OnSubmit: func(data map[string]string) error {
        // Handle form submission
        return nil
    },
}
```

### Modal Component

```go
modal := &components.Modal{
    Title:   "Confirm Action",
    Content: "Are you sure?",
    OnOk: func() {
        // Handle OK
    },
    OnCancel: func() {
        // Handle Cancel
    },
}
```

## Enhanced Flow Tool

The framework includes an enhanced flow tool for creating interactive node-based diagrams with drag-and-drop support.

### Features

- **Interactive Canvas**: Pan, zoom, and navigate flow diagrams
- **Drag & Drop**: Move boxes around the canvas with real-time updates
- **Connection Mode**: Create edges between boxes
- **Auto-arrange**: Automatically organize diagram layout
- **Import/Export**: Save and load diagrams as JSON
- **Undo/Redo**: Full history support for all operations
- **Delete Operations**: Remove boxes and edges with visual feedback

### Usage Example

```go
import "github.com/arturoeanton/go-echo-live-view/example/example_flowtool_enhanced"

tool := NewEnhancedFlowTool()
// Component is ready to use with full drag & drop support
```

## Drag and Drop Support

The framework provides built-in drag and drop functionality through its WASM module.

### Generic Drag & Drop

Any element can be made draggable by adding the appropriate classes and data attributes:

```html
<div class="draggable" 
     data-element-id="my-element"
     data-component-id="my-component">
    Draggable content
</div>
```

### Event Handling

The WASM module sends these events during drag operations:

- `DragStart`: Fired when dragging begins
- `DragMove`: Fired during drag movement (throttled)
- `DragEnd`: Fired when dragging completes

```go
func (c *MyComponent) HandleDragStart(data interface{}) {
    // data contains: {element: "element-id", x: 100, y: 200}
}
```

### Z-Index Management

To ensure draggable elements receive mouse events properly, set appropriate z-index values:

```css
.draggable-box {
    z-index: 20; /* Above SVG elements */
}

.svg-edges {
    z-index: 5-15; /* Below draggable elements */
}
```

## Security Features

### Input Validation

All WebSocket messages are validated before processing:

```go
validatedMsg, err := liveview.ValidateWebSocketMessage(msg)
```

### Template Sanitization

HTML templates are automatically sanitized to prevent XSS:

```go
sanitized := liveview.SanitizeHTML(html)
```

### Path Traversal Protection

File paths are validated to prevent directory traversal:

```go
if err := liveview.ValidatePath(path); err != nil {
    // Handle invalid path
}
```

### Rate Limiting

Built-in rate limiting prevents abuse:

```go
if !rateLimiter.Allow(clientID, time.Now().Unix()) {
    // Rate limit exceeded
}
```

## Best Practices

### 1. State Management

Keep component state minimal and focused:

```go
type TodoList struct {
    *liveview.ComponentDriver[*TodoList]
    Items []TodoItem // Good: focused state
    // Avoid storing derived state
}
```

### 2. Event Handling

Use descriptive event names and validate input:

```go
func (c *MyComponent) UpdateEmail(data interface{}) {
    email, ok := data.(string)
    if !ok || !isValidEmail(email) {
        return
    }
    c.Email = email
    c.Commit()
}
```

### 3. Component Composition

Prefer composition over inheritance:

```go
type Dashboard struct {
    *liveview.ComponentDriver[*Dashboard]
    Header *HeaderComponent
    Sidebar *SidebarComponent
    Content *ContentComponent
}
```

### 4. Error Handling

Always handle errors gracefully:

```go
func (c *MyComponent) LoadData(data interface{}) {
    result, err := fetchData()
    if err != nil {
        c.ShowError("Failed to load data")
        return
    }
    c.Data = result
    c.Commit()
}
```

### 5. Performance

Batch updates when possible:

```go
func (c *MyComponent) BulkUpdate(items []Item) {
    c.Items = items
    c.Total = len(items)
    c.UpdatedAt = time.Now()
    c.Commit() // Single commit for multiple changes
}
```

---

## Español

# Go Echo LiveView - Referencia API

Go Echo LiveView es un framework web en tiempo real para Go que permite renderizado del lado del servidor con reactividad basada en WebSocket, inspirado en Phoenix LiveView.

## Tabla de Contenidos

1. [Conceptos Fundamentales](#conceptos-fundamentales)
2. [Comenzando](#comenzando)
3. [API de Componentes](#api-de-componentes)
4. [Interfaz LiveDriver](#interfaz-livedriver)
5. [Control de Página](#control-de-página)
6. [Manejo de Eventos](#manejo-de-eventos)
7. [Comunicación WebSocket](#comunicación-websocket)
8. [Ciclo de Vida del Componente](#ciclo-de-vida-del-componente)
9. [Componentes Integrados](#componentes-integrados)
10. [Herramienta de Flujo Mejorada](#herramienta-de-flujo-mejorada)
11. [Soporte de Arrastrar y Soltar](#soporte-de-arrastrar-y-soltar)
12. [Características de Seguridad](#características-de-seguridad)
13. [Mejores Prácticas](#mejores-prácticas)

## Conceptos Fundamentales

### Arquitectura Basada en Componentes

LiveView utiliza una arquitectura basada en componentes donde cada elemento de UI es un componente autocontenido con su propio estado y ciclo de vida.

```go
type Component interface {
    GetTemplate() string    // Retorna plantilla HTML
    Start()                // Inicialización del componente
    GetDriver() LiveDriver // Retorna el driver del componente
}
```

### Actualizaciones en Tiempo Real

Los componentes se sincronizan automáticamente con el cliente vía WebSocket, eliminando la necesidad de llamadas AJAX manuales o gestión de estado del lado del cliente.

## Comenzando

### Instalación

```bash
go get github.com/arturoeanton/go-echo-live-view
```

### Compilar Módulo WASM

El framework incluye un módulo WASM para funcionalidad mejorada del lado del cliente:

```bash
cd cmd/wasm/
GOOS=js GOARCH=wasm go build -o ../../assets/json.wasm
```

### Ejemplo Básico

```go
package main

import (
    "github.com/arturoeanton/go-echo-live-view/liveview"
    "github.com/labstack/echo/v4"
)

type Contador struct {
    *liveview.ComponentDriver[*Contador]
    Cuenta int
}

func (c *Contador) Start() {
    c.Cuenta = 0
    c.Commit()
}

func (c *Contador) GetTemplate() string {
    return `
    <div>
        <h1>Cuenta: {{.Cuenta}}</h1>
        <button onclick="send_event('{{.IdComponent}}', 'Incrementar')">+</button>
        <button onclick="send_event('{{.IdComponent}}', 'Decrementar')">-</button>
    </div>
    `
}

func (c *Contador) GetDriver() liveview.LiveDriver {
    return c
}

func (c *Contador) Incrementar(data interface{}) {
    c.Cuenta++
    c.Commit()
}

func (c *Contador) Decrementar(data interface{}) {
    c.Cuenta--
    c.Commit()
}

func main() {
    e := echo.New()
    
    pagina := liveview.PageControl{
        Title:  "Ejemplo Contador",
        Path:   "/",
        Router: e,
    }
    
    pagina.Register(func() liveview.LiveDriver {
        contador := &Contador{}
        driver := liveview.NewDriver("contador", contador)
        contador.ComponentDriver = driver
        return driver
    })
    
    e.Logger.Fatal(e.Start(":8080"))
}
```

## API de Componentes

### Creando un Componente

Los componentes deben implementar la interfaz `Component`:

```go
type MiComponente struct {
    *liveview.ComponentDriver[*MiComponente]
    // Campos de estado del componente
    Mensaje string
    Cuenta  int
}

// Inicializar componente
func (c *MiComponente) Start() {
    c.Mensaje = "Hola Mundo"
    c.Commit() // Activar re-renderizado
}

// Definir plantilla HTML
func (c *MiComponente) GetTemplate() string {
    return `<div>{{.Mensaje}}</div>`
}

// Retornar referencia del driver
func (c *MiComponente) GetDriver() liveview.LiveDriver {
    return c
}
```

### Métodos del Componente

#### `Commit()`
Activa un re-renderizado del componente y envía actualizaciones al cliente.

```go
func (c *MiComponente) ActualizarMensaje(data interface{}) {
    c.Mensaje = "¡Actualizado!"
    c.Commit() // Enviar actualización al cliente
}
```

#### `Mount(component)`
Monta un componente hijo dentro del componente actual.

```go
func (c *ComponentePadre) Start() {
    hijo := &ComponenteHijo{}
    c.Mount(liveview.New("hijo", hijo))
    c.Commit()
}
```

## Interfaz LiveDriver

La interfaz `LiveDriver` gestiona el ciclo de vida del componente y la comunicación WebSocket:

```go
type LiveDriver interface {
    GetID() string
    SetID(string)
    StartDriver(*map[string]LiveDriver, *map[string]chan interface{}, chan (map[string]interface{}))
    StartDriverWithContext(ctx context.Context, drivers *map[string]LiveDriver, channelIn *map[string]chan interface{}, channel chan map[string]interface{})
    GetIDComponet() string
    ExecuteEvent(name string, data interface{})
    GetComponet() Component
    Mount(component Component) LiveDriver
    // Métodos de manipulación DOM
    FillValue(id, value string)
    SetHTML(id, html string)
    SetText(id, text string)
    SetStyle(id, property, value string)
    SetClass(id, class string)
    EvalScript(js string)
}
```

### Métodos de Manipulación DOM

#### `FillValue(id, value string)`
Actualiza el valor de un elemento input.

```go
c.FillValue("input-usuario", "juan_perez")
```

#### `SetHTML(id, html string)`
Establece el innerHTML de un elemento.

```go
c.SetHTML("contenido", "<p>Nuevo contenido</p>")
```

#### `SetText(id, text string)`
Establece el contenido de texto de un elemento.

```go
c.SetText("estado", "Conectado")
```

#### `SetStyle(id, property, value string)`
Actualiza propiedades de estilo CSS.

```go
c.SetStyle("panel", "background-color", "#f0f0f0")
```

## Control de Página

La estructura `PageControl` gestiona el enrutamiento de páginas y la configuración de WebSocket:

```go
type PageControl struct {
    Path      string              // Ruta URL
    Title     string              // Título de página
    HeadCode  string              // Contenido adicional <head>
    Lang      string              // Idioma de página
    Router    *echo.Echo          // Instancia del router Echo
    Debug     bool                // Habilitar modo debug
}
```

### Registrando Componentes

```go
pagina := liveview.PageControl{
    Title:  "Mi App",
    Path:   "/dashboard",
    Router: e,
    Debug:  true,
}

pagina.Register(func() liveview.LiveDriver {
    componente := &MiComponente{}
    driver := liveview.NewDriver("mi-componente", componente)
    componente.ComponentDriver = driver
    return driver
})
```

## Manejo de Eventos

### Eventos Cliente-a-Servidor

Los eventos se envían desde el cliente usando la función `send_event`:

```html
<button onclick="send_event('{{.IdComponent}}', 'Click')">Haz Click</button>
<input onchange="send_event('{{.IdComponent}}', 'Cambio', this.value)">
```

### Manejadores de Eventos del Servidor

Los manejadores de eventos son métodos en tu componente que coinciden con el nombre del evento:

```go
func (c *MiComponente) Click(data interface{}) {
    // Manejar evento click
    c.Commit()
}

func (c *MiComponente) Cambio(data interface{}) {
    if valor, ok := data.(string); ok {
        c.Valor = valor
        c.Commit()
    }
}
```

### Registro de Eventos Personalizados

```go
func (c *MiComponente) Start() {
    c.Events["EventoPersonalizado"] = func(comp *MiComponente, data interface{}) {
        // Manejar evento personalizado
    }
}
```

## Comunicación WebSocket

### Formato de Mensajes

Los mensajes WebSocket están codificados en JSON con la siguiente estructura:

```json
{
    "type": "data|get|fill|script",
    "id": "component_id",
    "event": "NombreEvento",
    "data": {},
    "value": "contenido"
}
```

### Tipos de Mensajes

- **`data`**: Mensaje de evento del componente
- **`get`**: Solicitar propiedad del componente
- **`fill`**: Actualizar elemento DOM
- **`script`**: Ejecutar JavaScript

### Limitación de Tasa

Las conexiones WebSocket incluyen limitación de tasa integrada:

```go
limitador := liveview.NewRateLimiter(100, 60) // 100 mensajes por minuto
```

## Ciclo de Vida del Componente

### Hooks del Ciclo de Vida

1. **Creación**: Instancia del componente creada
2. **Asignación de Driver**: `ComponentDriver` asignado
3. **Start()**: Inicialización del componente
4. **Mount**: Componentes hijos montados
5. **Render**: Renderizado inicial de plantilla
6. **Bucle de Eventos**: Manejo de eventos y actualizaciones
7. **Destroy**: Limpieza (si está implementado)

### Gestión de Contexto

Los componentes soportan gestión de ciclo de vida basada en contexto:

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

componente.StartDriverWithContext(ctx, &drivers, &channelIn, channel)
```

## Componentes Integrados

### Componente Tabla

```go
tabla := &components.Table{
    Columns: []components.Column{
        {Key: "id", Title: "ID", Sortable: true},
        {Key: "nombre", Title: "Nombre", Sortable: true},
    },
    Rows: []components.Row{
        {"id": 1, "nombre": "Juan"},
        {"id": 2, "nombre": "María"},
    },
    ShowPagination: true,
    PageSize: 10,
}
```

### Componente Formulario

```go
formulario := &components.Form{
    Fields: []components.FormField{
        {
            Name:     "email",
            Label:    "Correo",
            Type:     "email",
            Required: true,
        },
    },
    OnSubmit: func(data map[string]string) error {
        // Manejar envío del formulario
        return nil
    },
}
```

### Componente Modal

```go
modal := &components.Modal{
    Title:   "Confirmar Acción",
    Content: "¿Estás seguro?",
    OnOk: func() {
        // Manejar OK
    },
    OnCancel: func() {
        // Manejar Cancelar
    },
}
```

## Herramienta de Flujo Mejorada

El framework incluye una herramienta de flujo mejorada para crear diagramas interactivos basados en nodos con soporte de arrastrar y soltar.

### Características

- **Canvas Interactivo**: Desplazar, hacer zoom y navegar diagramas de flujo
- **Arrastrar y Soltar**: Mover cajas por el canvas con actualizaciones en tiempo real
- **Modo de Conexión**: Crear enlaces entre cajas
- **Auto-organizar**: Organizar automáticamente el diseño del diagrama
- **Importar/Exportar**: Guardar y cargar diagramas como JSON
- **Deshacer/Rehacer**: Soporte completo de historial para todas las operaciones
- **Operaciones de Eliminación**: Eliminar cajas y enlaces con retroalimentación visual

### Ejemplo de Uso

```go
import "github.com/arturoeanton/go-echo-live-view/example/example_flowtool_enhanced"

tool := NewEnhancedFlowTool()
// El componente está listo para usar con soporte completo de arrastrar y soltar
```

## Soporte de Arrastrar y Soltar

El framework proporciona funcionalidad integrada de arrastrar y soltar a través de su módulo WASM.

### Arrastrar y Soltar Genérico

Cualquier elemento puede hacerse arrastrable agregando las clases y atributos de datos apropiados:

```html
<div class="draggable" 
     data-element-id="mi-elemento"
     data-component-id="mi-componente">
    Contenido arrastrable
</div>
```

### Manejo de Eventos

El módulo WASM envía estos eventos durante las operaciones de arrastre:

- `DragStart`: Se dispara cuando comienza el arrastre
- `DragMove`: Se dispara durante el movimiento de arrastre (limitado)
- `DragEnd`: Se dispara cuando se completa el arrastre

```go
func (c *MiComponente) HandleDragStart(data interface{}) {
    // data contiene: {element: "id-elemento", x: 100, y: 200}
}
```

### Gestión de Z-Index

Para asegurar que los elementos arrastrables reciban eventos del mouse correctamente, establezca valores apropiados de z-index:

```css
.draggable-box {
    z-index: 20; /* Por encima de elementos SVG */
}

.svg-edges {
    z-index: 5-15; /* Por debajo de elementos arrastrables */
}
```

## Características de Seguridad

### Validación de Entrada

Todos los mensajes WebSocket son validados antes de procesarse:

```go
mensajeValidado, err := liveview.ValidateWebSocketMessage(msg)
```

### Sanitización de Plantillas

Las plantillas HTML se sanitizan automáticamente para prevenir XSS:

```go
sanitizado := liveview.SanitizeHTML(html)
```

### Protección contra Path Traversal

Las rutas de archivos se validan para prevenir traversal de directorios:

```go
if err := liveview.ValidatePath(ruta); err != nil {
    // Manejar ruta inválida
}
```

### Limitación de Tasa

Limitación de tasa integrada previene abuso:

```go
if !limitador.Allow(clienteID, time.Now().Unix()) {
    // Límite de tasa excedido
}
```

## Mejores Prácticas

### 1. Gestión de Estado

Mantén el estado del componente mínimo y enfocado:

```go
type ListaTareas struct {
    *liveview.ComponentDriver[*ListaTareas]
    Items []ItemTarea // Bien: estado enfocado
    // Evitar almacenar estado derivado
}
```

### 2. Manejo de Eventos

Usa nombres de eventos descriptivos y valida entrada:

```go
func (c *MiComponente) ActualizarEmail(data interface{}) {
    email, ok := data.(string)
    if !ok || !esEmailValido(email) {
        return
    }
    c.Email = email
    c.Commit()
}
```

### 3. Composición de Componentes

Prefiere composición sobre herencia:

```go
type Dashboard struct {
    *liveview.ComponentDriver[*Dashboard]
    Cabecera *ComponenteCabecera
    Lateral  *ComponenteLateral
    Contenido *ComponenteContenido
}
```

### 4. Manejo de Errores

Siempre maneja errores graciosamente:

```go
func (c *MiComponente) CargarDatos(data interface{}) {
    resultado, err := obtenerDatos()
    if err != nil {
        c.MostrarError("Fallo al cargar datos")
        return
    }
    c.Datos = resultado
    c.Commit()
}
```

### 5. Rendimiento

Agrupa actualizaciones cuando sea posible:

```go
func (c *MiComponente) ActualizacionMasiva(items []Item) {
    c.Items = items
    c.Total = len(items)
    c.ActualizadoEn = time.Now()
    c.Commit() // Un solo commit para múltiples cambios
}
```

## API Reference / Referencia API

### Core Functions / Funciones Principales

| Function | Description | Descripción |
|----------|-------------|-------------|
| `NewDriver(id, component)` | Creates a new component driver | Crea un nuevo driver de componente |
| `New(id, component)` | Creates and registers a component | Crea y registra un componente |
| `Mount(component)` | Mounts a child component | Monta un componente hijo |
| `Commit()` | Triggers component re-render | Activa re-renderizado del componente |
| `ExecuteEvent(name, data)` | Executes a component event | Ejecuta un evento del componente |

### DOM Manipulation / Manipulación DOM

| Method | Description | Descripción |
|--------|-------------|-------------|
| `FillValue(id, value)` | Updates input value | Actualiza valor de input |
| `SetHTML(id, html)` | Sets innerHTML | Establece innerHTML |
| `SetText(id, text)` | Sets text content | Establece contenido de texto |
| `SetStyle(id, prop, val)` | Updates CSS style | Actualiza estilo CSS |
| `SetClass(id, class)` | Sets CSS class | Establece clase CSS |
| `AddClass(id, class)` | Adds CSS class | Agrega clase CSS |
| `RemoveClass(id, class)` | Removes CSS class | Remueve clase CSS |

### WebSocket Events / Eventos WebSocket

| Event Type | Description | Descripción |
|------------|-------------|-------------|
| `data` | Component event | Evento de componente |
| `get` | Property request | Solicitud de propiedad |
| `fill` | DOM update | Actualización DOM |
| `script` | JavaScript execution | Ejecución JavaScript |

## Examples / Ejemplos

### Real-time Chat / Chat en Tiempo Real

```go
type Chat struct {
    *liveview.ComponentDriver[*Chat]
    Messages []Message
    Input    string
}

func (c *Chat) SendMessage(data interface{}) {
    if msg, ok := data.(string); ok && msg != "" {
        c.Messages = append(c.Messages, Message{
            Text: msg,
            Time: time.Now(),
        })
        c.Input = ""
        c.Commit()
    }
}
```

### Live Dashboard / Dashboard en Vivo

```go
type Dashboard struct {
    *liveview.ComponentDriver[*Dashboard]
    Stats     Stats
    UpdateTicker *time.Ticker
}

func (d *Dashboard) Start() {
    d.UpdateTicker = time.NewTicker(5 * time.Second)
    go func() {
        for range d.UpdateTicker.C {
            d.Stats = fetchLatestStats()
            d.Commit()
        }
    }()
}
```

## Support / Soporte

- GitHub: [github.com/arturoeanton/go-echo-live-view](https://github.com/arturoeanton/go-echo-live-view)
- Issues: [GitHub Issues](https://github.com/arturoeanton/go-echo-live-view/issues)
- Documentation: This file / Este archivo

## License / Licencia

MIT License - See LICENSE file for details / Ver archivo LICENSE para detalles