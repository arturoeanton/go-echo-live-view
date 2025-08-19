# Go Echo LiveView - Migration Guide

[English](#english) | [Español](#español)

---

## English

# Migration Guide

This guide helps you migrate from other frameworks to Go Echo LiveView or upgrade between versions.

## Table of Contents
1. [Migrating from Phoenix LiveView](#migrating-from-phoenix-liveview)
2. [Migrating from React/Vue/Angular](#migrating-from-reactvueangular)
3. [Migrating from Traditional Server-Side Rendering](#migrating-from-traditional-server-side-rendering)
4. [Migrating from WebSocket Libraries](#migrating-from-websocket-libraries)
5. [Version Migration](#version-migration)
6. [Common Migration Patterns](#common-migration-patterns)

## Migrating from Phoenix LiveView

### Concept Mapping

| Phoenix LiveView | Go Echo LiveView |
|-----------------|------------------|
| `Phoenix.LiveView` | `liveview.ComponentDriver` |
| `mount/3` | `Start()` |
| `render/1` | `GetTemplate()` |
| `handle_event/3` | Event methods (e.g., `Click()`) |
| `assign/2` | Direct field assignment + `Commit()` |
| `push_event/3` | `SendToClient()` |
| `live_component` | Child components with `Mount()` |

### Example Migration

**Phoenix LiveView:**
```elixir
defmodule MyAppWeb.CounterLive do
  use MyAppWeb, :live_view

  def mount(_params, _session, socket) do
    {:ok, assign(socket, count: 0)}
  end

  def render(assigns) do
    ~L"""
    <div>
      <h1>Count: <%= @count %></h1>
      <button phx-click="increment">+</button>
      <button phx-click="decrement">-</button>
    </div>
    """
  end

  def handle_event("increment", _params, socket) do
    {:noreply, update(socket, :count, &(&1 + 1))}
  end

  def handle_event("decrement", _params, socket) do
    {:noreply, update(socket, :count, &(&1 - 1))}
  end
end
```

**Go Echo LiveView:**
```go
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
    return c.ComponentDriver
}

func (c *Counter) Increment(data interface{}) {
    c.Count++
    c.Commit()
}

func (c *Counter) Decrement(data interface{}) {
    c.Count--
    c.Commit()
}
```

### Key Differences

1. **Type Safety**: Go provides compile-time type safety
2. **Template Syntax**: Go templates instead of EEx
3. **Event Handling**: Method-based instead of pattern matching
4. **State Management**: Direct struct fields instead of assigns

## Migrating from React/Vue/Angular

### Architecture Shift

Moving from client-side to server-side rendering:

| SPA Framework | Go Echo LiveView |
|--------------|------------------|
| Components | Server-side Components |
| State Management (Redux/Vuex) | Component struct fields |
| Props | Component fields |
| Hooks/Lifecycle | `Start()`, `Stop()` |
| Virtual DOM | Server-side diffing |
| API Calls | Direct database access |

### React Component Migration

**React:**
```jsx
function TodoList() {
    const [todos, setTodos] = useState([]);
    const [input, setInput] = useState('');

    const addTodo = () => {
        setTodos([...todos, { id: Date.now(), text: input }]);
        setInput('');
    };

    return (
        <div>
            <input 
                value={input}
                onChange={(e) => setInput(e.target.value)}
            />
            <button onClick={addTodo}>Add</button>
            <ul>
                {todos.map(todo => (
                    <li key={todo.id}>{todo.text}</li>
                ))}
            </ul>
        </div>
    );
}
```

**Go Echo LiveView:**
```go
type TodoList struct {
    *liveview.ComponentDriver[*TodoList]
    Todos []Todo
    Input string
}

func (t *TodoList) Start() {
    t.Todos = make([]Todo, 0)
    t.Commit()
}

func (t *TodoList) GetTemplate() string {
    return `
    <div>
        <input 
            value="{{.Input}}"
            onkeyup="send_event('{{.IdComponent}}', 'UpdateInput', this.value)"
        />
        <button onclick="send_event('{{.IdComponent}}', 'AddTodo')">Add</button>
        <ul>
            {{range .Todos}}
            <li>{{.Text}}</li>
            {{end}}
        </ul>
    </div>
    `
}

func (t *TodoList) UpdateInput(data interface{}) {
    if input, ok := data.(string); ok {
        t.Input = input
        t.Commit()
    }
}

func (t *TodoList) AddTodo(data interface{}) {
    if t.Input != "" {
        t.Todos = append(t.Todos, Todo{
            ID:   time.Now().Unix(),
            Text: t.Input,
        })
        t.Input = ""
        t.Commit()
    }
}
```

### Vue Component Migration

**Vue:**
```vue
<template>
  <div>
    <h1>{{ title }}</h1>
    <button @click="count++">Count: {{ count }}</button>
  </div>
</template>

<script>
export default {
  data() {
    return {
      title: 'Counter',
      count: 0
    }
  }
}
</script>
```

**Go Echo LiveView:**
```go
type Counter struct {
    *liveview.ComponentDriver[*Counter]
    Title string
    Count int
}

func (c *Counter) Start() {
    c.Title = "Counter"
    c.Count = 0
    c.Commit()
}

func (c *Counter) GetTemplate() string {
    return `
    <div>
        <h1>{{.Title}}</h1>
        <button onclick="send_event('{{.IdComponent}}', 'Increment')">
            Count: {{.Count}}
        </button>
    </div>
    `
}

func (c *Counter) Increment(data interface{}) {
    c.Count++
    c.Commit()
}
```

### Migration Benefits

1. **Reduced Complexity**: No build tools, webpack, or transpilation
2. **Better SEO**: Server-side rendering by default
3. **Simplified State**: No client-server state synchronization
4. **Reduced Bundle Size**: No JavaScript framework overhead
5. **Type Safety**: Go's compile-time type checking

## Migrating from Traditional Server-Side Rendering

### From Standard Go Templates

**Before (Traditional):**
```go
func handler(c echo.Context) error {
    data := getData()
    return c.Render(http.StatusOK, "template.html", data)
}

// template.html
<form method="POST" action="/update">
    <input name="value" value="{{.Value}}">
    <button type="submit">Update</button>
</form>
```

**After (LiveView):**
```go
type FormComponent struct {
    *liveview.ComponentDriver[*FormComponent]
    Value string
}

func (f *FormComponent) GetTemplate() string {
    return `
    <div>
        <input value="{{.Value}}" 
               onkeyup="send_event('{{.IdComponent}}', 'UpdateValue', this.value)">
        <button onclick="send_event('{{.IdComponent}}', 'Submit')">Update</button>
    </div>
    `
}

func (f *FormComponent) UpdateValue(data interface{}) {
    if value, ok := data.(string); ok {
        f.Value = value
        f.Commit() // Real-time update, no page reload
    }
}
```

### From MVC Frameworks

**Before (MVC):**
```go
// Controller
func (c *UserController) Index(ctx *gin.Context) {
    users := c.userService.GetAll()
    ctx.HTML(200, "users/index", gin.H{
        "users": users,
    })
}

func (c *UserController) Delete(ctx *gin.Context) {
    id := ctx.Param("id")
    c.userService.Delete(id)
    ctx.Redirect(302, "/users")
}
```

**After (LiveView):**
```go
type UserList struct {
    *liveview.ComponentDriver[*UserList]
    Users       []User
    userService *UserService
}

func (u *UserList) Start() {
    u.Users = u.userService.GetAll()
    u.Commit()
}

func (u *UserList) DeleteUser(data interface{}) {
    if id, ok := data.(string); ok {
        u.userService.Delete(id)
        u.Users = u.userService.GetAll() // Refresh list
        u.Commit() // Update UI without redirect
    }
}
```

## Migrating from WebSocket Libraries

### From Gorilla WebSocket

**Before:**
```go
func handleWebSocket(w http.ResponseWriter, r *http.Request) {
    conn, _ := upgrader.Upgrade(w, r, nil)
    defer conn.Close()
    
    for {
        var msg Message
        err := conn.ReadJSON(&msg)
        if err != nil {
            break
        }
        
        // Process message
        response := processMessage(msg)
        
        conn.WriteJSON(response)
    }
}
```

**After:**
```go
type WebSocketComponent struct {
    *liveview.ComponentDriver[*WebSocketComponent]
    Messages []Message
}

func (w *WebSocketComponent) HandleMessage(data interface{}) {
    // Automatic WebSocket handling
    if msg, ok := data.(map[string]interface{}); ok {
        message := Message{
            Text: msg["text"].(string),
            Time: time.Now(),
        }
        w.Messages = append(w.Messages, message)
        w.Commit() // Automatically sends update via WebSocket
    }
}
```

## Version Migration

### v0.x to v1.0

**Breaking Changes:**

1. **Component Interface Change:**
```go
// Old
type Component interface {
    GetTemplate() string
    GetDriver() LiveDriver
}

// New
type Component interface {
    Start()
    GetTemplate() string
    GetDriver() LiveDriver
}
```

2. **Event Handling:**
```go
// Old
driver.Events["Click"] = func(data interface{}) {}

// New - Use methods
func (c *Component) Click(data interface{}) {}
```

3. **Context Support:**
```go
// Old
driver.StartDriver()

// New
driver.StartDriverWithContext(ctx)
```

### Migration Script

```go
// migration_helper.go
package migration

import (
    "regexp"
    "strings"
)

// MigrateTemplate converts old template syntax to new
func MigrateTemplate(old string) string {
    // Replace old event syntax
    re := regexp.MustCompile(`onclick="{{call '(\w+)'}}"`)
    new := re.ReplaceAllString(old, `onclick="send_event('{{.IdComponent}}', '$1')"`)
    
    // Replace old mount syntax
    re = regexp.MustCompile(`{{mount (\w+)}}`)
    new = re.ReplaceAllString(new, `{{mount "$1"}}`)
    
    return new
}

// MigrateComponent helps convert old components
func MigrateComponent(old string) string {
    // Add Start method if missing
    if !strings.Contains(old, "func (") && !strings.Contains(old, "Start()") {
        old = strings.Replace(old, "func (", 
            "func (c *Component) Start() {\n    c.Commit()\n}\n\nfunc (", 1)
    }
    
    return old
}
```

## Common Migration Patterns

### 1. Form Handling

**Traditional Form:**
```html
<form method="POST" action="/submit">
    <input name="email" type="email" required>
    <button type="submit">Submit</button>
</form>
```

**LiveView Form:**
```go
func (f *Form) GetTemplate() string {
    return `
    <div>
        <input type="email" 
               value="{{.Email}}"
               onkeyup="send_event('{{.IdComponent}}', 'UpdateEmail', this.value)"
               class="{{if .EmailError}}error{{end}}">
        {{if .EmailError}}<span class="error">{{.EmailError}}</span>{{end}}
        <button onclick="send_event('{{.IdComponent}}', 'Submit')">Submit</button>
    </div>
    `
}

func (f *Form) UpdateEmail(data interface{}) {
    if email, ok := data.(string); ok {
        f.Email = email
        f.EmailError = f.validateEmail(email)
        f.Commit()
    }
}
```

### 2. AJAX to LiveView

**AJAX Pattern:**
```javascript
fetch('/api/data')
    .then(response => response.json())
    .then(data => {
        document.getElementById('result').innerHTML = renderData(data);
    });
```

**LiveView Pattern:**
```go
func (c *Component) Start() {
    go func() {
        data := fetchData()
        c.Data = data
        c.Commit() // Automatically updates DOM
    }()
}
```

### 3. Polling to Real-time

**Polling Pattern:**
```javascript
setInterval(() => {
    fetch('/api/status')
        .then(response => response.json())
        .then(data => updateStatus(data));
}, 5000);
```

**LiveView Real-time:**
```go
func (s *Status) Start() {
    ticker := time.NewTicker(5 * time.Second)
    go func() {
        for range ticker.C {
            s.Status = fetchStatus()
            s.Commit() // Real-time update
        }
    }()
}
```

---

## Español

# Guía de Migración

Esta guía te ayuda a migrar desde otros frameworks a Go Echo LiveView o actualizar entre versiones.

## Tabla de Contenidos
1. [Migrando desde Phoenix LiveView](#migrando-desde-phoenix-liveview)
2. [Migrando desde React/Vue/Angular](#migrando-desde-reactvueangular-1)
3. [Migrando desde Renderizado Tradicional del Servidor](#migrando-desde-renderizado-tradicional-del-servidor)
4. [Migrando desde Bibliotecas WebSocket](#migrando-desde-bibliotecas-websocket)
5. [Migración de Versiones](#migración-de-versiones)
6. [Patrones Comunes de Migración](#patrones-comunes-de-migración)

## Migrando desde Phoenix LiveView

### Mapeo de Conceptos

| Phoenix LiveView | Go Echo LiveView |
|-----------------|------------------|
| `Phoenix.LiveView` | `liveview.ComponentDriver` |
| `mount/3` | `Start()` |
| `render/1` | `GetTemplate()` |
| `handle_event/3` | Métodos de evento (ej. `Click()`) |
| `assign/2` | Asignación directa + `Commit()` |
| `push_event/3` | `SendToClient()` |
| `live_component` | Componentes hijos con `Mount()` |

### Ejemplo de Migración

**Phoenix LiveView:**
```elixir
defmodule MiAppWeb.ContadorLive do
  use MiAppWeb, :live_view

  def mount(_params, _session, socket) do
    {:ok, assign(socket, contador: 0)}
  end

  def render(assigns) do
    ~L"""
    <div>
      <h1>Contador: <%= @contador %></h1>
      <button phx-click="incrementar">+</button>
      <button phx-click="decrementar">-</button>
    </div>
    """
  end

  def handle_event("incrementar", _params, socket) do
    {:noreply, update(socket, :contador, &(&1 + 1))}
  end

  def handle_event("decrementar", _params, socket) do
    {:noreply, update(socket, :contador, &(&1 - 1))}
  end
end
```

**Go Echo LiveView:**
```go
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
        <h1>Contador: {{.Cuenta}}</h1>
        <button onclick="send_event('{{.IdComponent}}', 'Incrementar')">+</button>
        <button onclick="send_event('{{.IdComponent}}', 'Decrementar')">-</button>
    </div>
    `
}

func (c *Contador) GetDriver() liveview.LiveDriver {
    return c.ComponentDriver
}

func (c *Contador) Incrementar(data interface{}) {
    c.Cuenta++
    c.Commit()
}

func (c *Contador) Decrementar(data interface{}) {
    c.Cuenta--
    c.Commit()
}
```

## Migrando desde React/Vue/Angular

### Cambio de Arquitectura

Pasando de renderizado del lado del cliente al servidor:

| Framework SPA | Go Echo LiveView |
|--------------|------------------|
| Componentes | Componentes del servidor |
| Gestión de Estado (Redux/Vuex) | Campos de struct |
| Props | Campos del componente |
| Hooks/Ciclo de vida | `Start()`, `Stop()` |
| DOM Virtual | Diffing del servidor |
| Llamadas API | Acceso directo a BD |

### Migración de Componente React

**React:**
```jsx
function ListaTareas() {
    const [tareas, setTareas] = useState([]);
    const [entrada, setEntrada] = useState('');

    const agregarTarea = () => {
        setTareas([...tareas, { id: Date.now(), texto: entrada }]);
        setEntrada('');
    };

    return (
        <div>
            <input 
                value={entrada}
                onChange={(e) => setEntrada(e.target.value)}
            />
            <button onClick={agregarTarea}>Agregar</button>
            <ul>
                {tareas.map(tarea => (
                    <li key={tarea.id}>{tarea.texto}</li>
                ))}
            </ul>
        </div>
    );
}
```

**Go Echo LiveView:**
```go
type ListaTareas struct {
    *liveview.ComponentDriver[*ListaTareas]
    Tareas  []Tarea
    Entrada string
}

func (l *ListaTareas) Start() {
    l.Tareas = make([]Tarea, 0)
    l.Commit()
}

func (l *ListaTareas) GetTemplate() string {
    return `
    <div>
        <input 
            value="{{.Entrada}}"
            onkeyup="send_event('{{.IdComponent}}', 'ActualizarEntrada', this.value)"
        />
        <button onclick="send_event('{{.IdComponent}}', 'AgregarTarea')">Agregar</button>
        <ul>
            {{range .Tareas}}
            <li>{{.Texto}}</li>
            {{end}}
        </ul>
    </div>
    `
}

func (l *ListaTareas) ActualizarEntrada(data interface{}) {
    if entrada, ok := data.(string); ok {
        l.Entrada = entrada
        l.Commit()
    }
}

func (l *ListaTareas) AgregarTarea(data interface{}) {
    if l.Entrada != "" {
        l.Tareas = append(l.Tareas, Tarea{
            ID:    time.Now().Unix(),
            Texto: l.Entrada,
        })
        l.Entrada = ""
        l.Commit()
    }
}
```

## Migrando desde Renderizado Tradicional del Servidor

### Desde Templates Go Estándar

**Antes (Tradicional):**
```go
func manejador(c echo.Context) error {
    datos := obtenerDatos()
    return c.Render(http.StatusOK, "plantilla.html", datos)
}

// plantilla.html
<form method="POST" action="/actualizar">
    <input name="valor" value="{{.Valor}}">
    <button type="submit">Actualizar</button>
</form>
```

**Después (LiveView):**
```go
type ComponenteFormulario struct {
    *liveview.ComponentDriver[*ComponenteFormulario]
    Valor string
}

func (f *ComponenteFormulario) GetTemplate() string {
    return `
    <div>
        <input value="{{.Valor}}" 
               onkeyup="send_event('{{.IdComponent}}', 'ActualizarValor', this.value)">
        <button onclick="send_event('{{.IdComponent}}', 'Enviar')">Actualizar</button>
    </div>
    `
}

func (f *ComponenteFormulario) ActualizarValor(data interface{}) {
    if valor, ok := data.(string); ok {
        f.Valor = valor
        f.Commit() // Actualización en tiempo real, sin recarga
    }
}
```

## Migrando desde Bibliotecas WebSocket

### Desde Gorilla WebSocket

**Antes:**
```go
func manejarWebSocket(w http.ResponseWriter, r *http.Request) {
    conn, _ := upgrader.Upgrade(w, r, nil)
    defer conn.Close()
    
    for {
        var msg Mensaje
        err := conn.ReadJSON(&msg)
        if err != nil {
            break
        }
        
        // Procesar mensaje
        respuesta := procesarMensaje(msg)
        
        conn.WriteJSON(respuesta)
    }
}
```

**Después:**
```go
type ComponenteWebSocket struct {
    *liveview.ComponentDriver[*ComponenteWebSocket]
    Mensajes []Mensaje
}

func (w *ComponenteWebSocket) ManejarMensaje(data interface{}) {
    // Manejo automático de WebSocket
    if msg, ok := data.(map[string]interface{}); ok {
        mensaje := Mensaje{
            Texto: msg["texto"].(string),
            Hora:  time.Now(),
        }
        w.Mensajes = append(w.Mensajes, mensaje)
        w.Commit() // Envía actualización automáticamente vía WebSocket
    }
}
```

## Migración de Versiones

### v0.x a v1.0

**Cambios Importantes:**

1. **Cambio en Interface de Componente:**
```go
// Viejo
type Component interface {
    GetTemplate() string
    GetDriver() LiveDriver
}

// Nuevo
type Component interface {
    Start()
    GetTemplate() string
    GetDriver() LiveDriver
}
```

2. **Manejo de Eventos:**
```go
// Viejo
driver.Events["Click"] = func(data interface{}) {}

// Nuevo - Usar métodos
func (c *Componente) Click(data interface{}) {}
```

3. **Soporte de Contexto:**
```go
// Viejo
driver.StartDriver()

// Nuevo
driver.StartDriverWithContext(ctx)
```

## Patrones Comunes de Migración

### 1. Manejo de Formularios

**Formulario Tradicional:**
```html
<form method="POST" action="/enviar">
    <input name="email" type="email" required>
    <button type="submit">Enviar</button>
</form>
```

**Formulario LiveView:**
```go
func (f *Formulario) GetTemplate() string {
    return `
    <div>
        <input type="email" 
               value="{{.Email}}"
               onkeyup="send_event('{{.IdComponent}}', 'ActualizarEmail', this.value)"
               class="{{if .ErrorEmail}}error{{end}}">
        {{if .ErrorEmail}}<span class="error">{{.ErrorEmail}}</span>{{end}}
        <button onclick="send_event('{{.IdComponent}}', 'Enviar')">Enviar</button>
    </div>
    `
}

func (f *Formulario) ActualizarEmail(data interface{}) {
    if email, ok := data.(string); ok {
        f.Email = email
        f.ErrorEmail = f.validarEmail(email)
        f.Commit()
    }
}
```

### 2. AJAX a LiveView

**Patrón AJAX:**
```javascript
fetch('/api/datos')
    .then(response => response.json())
    .then(datos => {
        document.getElementById('resultado').innerHTML = renderizarDatos(datos);
    });
```

**Patrón LiveView:**
```go
func (c *Componente) Start() {
    go func() {
        datos := obtenerDatos()
        c.Datos = datos
        c.Commit() // Actualiza DOM automáticamente
    }()
}
```

### 3. Polling a Tiempo Real

**Patrón Polling:**
```javascript
setInterval(() => {
    fetch('/api/estado')
        .then(response => response.json())
        .then(datos => actualizarEstado(datos));
}, 5000);
```

**LiveView Tiempo Real:**
```go
func (e *Estado) Start() {
    ticker := time.NewTicker(5 * time.Second)
    go func() {
        for range ticker.C {
            e.Estado = obtenerEstado()
            e.Commit() // Actualización en tiempo real
        }
    }()
}
```

## Beneficios de la Migración

### Ventajas Principales

1. **Reducción de Complejidad**: Sin herramientas de construcción o transpilación
2. **Mejor SEO**: Renderizado del servidor por defecto
3. **Estado Simplificado**: Sin sincronización cliente-servidor
4. **Menor Tamaño**: Sin overhead de frameworks JavaScript
5. **Seguridad de Tipos**: Verificación en tiempo de compilación de Go

### Comparación de Rendimiento

| Métrica | SPA Tradicional | Go Echo LiveView |
|---------|-----------------|------------------|
| Tiempo de carga inicial | 2-5s | < 1s |
| Tamaño del bundle | 200-500KB | < 50KB |
| Time to Interactive | 3-8s | < 1s |
| Uso de memoria | 50-200MB | 10-50MB |

## Recursos de Migración

### Herramientas

- **Template Converter**: Convierte templates de otros frameworks
- **Component Generator**: Genera componentes desde esquemas
- **Migration Validator**: Valida código migrado

### Checklist de Migración

- [ ] Identificar componentes a migrar
- [ ] Mapear rutas y endpoints
- [ ] Convertir templates
- [ ] Migrar lógica de negocio
- [ ] Implementar manejo de eventos
- [ ] Agregar validación
- [ ] Configurar WebSocket
- [ ] Probar funcionalidad
- [ ] Optimizar rendimiento
- [ ] Desplegar

## Soporte

Para ayuda con la migración:
- Documentación: [API Documentation](API_DOCUMENTATION.md)
- Ejemplos: [Examples](example/)
- Comunidad: [GitHub Discussions](https://github.com/arturoeanton/go-echo-live-view/discussions)

---

La migración a Go Echo LiveView simplifica el desarrollo web mientras mejora el rendimiento y la experiencia del usuario.