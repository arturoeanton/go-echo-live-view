# Go Echo LiveView - Step-by-Step Tutorial

[English](#english) | [Español](#español)

---

## English

# Complete Step-by-Step Tutorial

## Table of Contents
1. [Getting Started](#getting-started)
2. [Your First LiveView App](#your-first-liveview-app)
3. [Understanding Components](#understanding-components)
4. [Handling Events](#handling-events)
5. [Working with State](#working-with-state)
6. [Component Composition](#component-composition)
7. [Real-time Features](#real-time-features)
8. [Testing Your Components](#testing-your-components)
9. [Deployment](#deployment)
10. [Advanced Topics](#advanced-topics)

## Getting Started

### Prerequisites
- Go 1.19 or higher installed
- Basic knowledge of Go programming
- A text editor (VS Code, GoLand, etc.)
- Terminal/Command line access

### Installation

1. **Create a new Go project:**
```bash
mkdir my-liveview-app
cd my-liveview-app
go mod init my-liveview-app
```

2. **Install Go Echo LiveView:**
```bash
go get github.com/arturoeanton/go-echo-live-view
go get github.com/labstack/echo/v4
```

3. **Verify installation:**
```bash
go mod tidy
```

## Your First LiveView App

Let's create a simple counter application that updates in real-time.

### Step 1: Create main.go

Create a file named `main.go`:

```go
package main

import (
    "github.com/arturoeanton/go-echo-live-view/liveview"
    "github.com/labstack/echo/v4"
    "github.com/labstack/echo/v4/middleware"
)

func main() {
    // Create Echo instance
    e := echo.New()
    
    // Add middleware
    e.Use(middleware.Logger())
    e.Use(middleware.Recover())
    
    // Setup LiveView page
    home := liveview.PageControl{
        Title:  "My First LiveView App",
        Path:   "/",
        Router: e,
    }
    
    // Register component
    home.Register(func() liveview.LiveDriver {
        counter := &Counter{}
        driver := liveview.NewDriver("counter", counter)
        counter.ComponentDriver = driver
        return driver
    })
    
    // Start server
    e.Logger.Fatal(e.Start(":8080"))
}
```

### Step 2: Create the Counter Component

Add this to the same file:

```go
type Counter struct {
    *liveview.ComponentDriver[*Counter]
    Count int
}

func (c *Counter) Start() {
    c.Count = 0
    c.Commit() // Render the component
}

func (c *Counter) GetTemplate() string {
    return `
    <div style="text-align: center; padding: 50px;">
        <h1>Counter: {{.Count}}</h1>
        <button onclick="send_event('{{.IdComponent}}', 'Increment')" 
                style="padding: 10px 20px; font-size: 16px;">
            Increment
        </button>
        <button onclick="send_event('{{.IdComponent}}', 'Decrement')" 
                style="padding: 10px 20px; font-size: 16px;">
            Decrement
        </button>
        <button onclick="send_event('{{.IdComponent}}', 'Reset')" 
                style="padding: 10px 20px; font-size: 16px;">
            Reset
        </button>
    </div>
    `
}

func (c *Counter) GetDriver() liveview.LiveDriver {
    return c.ComponentDriver
}

// Event handlers
func (c *Counter) Increment(data interface{}) {
    c.Count++
    c.Commit() // Update the UI
}

func (c *Counter) Decrement(data interface{}) {
    c.Count--
    c.Commit()
}

func (c *Counter) Reset(data interface{}) {
    c.Count = 0
    c.Commit()
}
```

### Step 3: Run Your App

```bash
go run main.go
```

Open your browser and navigate to `http://localhost:8080`. You should see your counter app!

## Understanding Components

### Component Structure

Every LiveView component must implement the `Component` interface:

```go
type Component interface {
    Start()                    // Initialize component
    GetTemplate() string       // Return HTML template
    GetDriver() LiveDriver     // Return the driver
}
```

### Example: Todo List Component

Let's create a more complex component - a todo list:

```go
type TodoList struct {
    *liveview.ComponentDriver[*TodoList]
    Todos []Todo
    Input string
}

type Todo struct {
    ID        int
    Text      string
    Completed bool
}

func (t *TodoList) Start() {
    t.Todos = []Todo{
        {ID: 1, Text: "Learn LiveView", Completed: false},
        {ID: 2, Text: "Build an app", Completed: false},
    }
    t.Commit()
}

func (t *TodoList) GetTemplate() string {
    return `
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h1>Todo List</h1>
        
        <div style="margin-bottom: 20px;">
            <input type="text" 
                   id="todo-input" 
                   value="{{.Input}}"
                   onkeyup="if(event.key === 'Enter') send_event('{{.IdComponent}}', 'AddTodo', this.value)"
                   placeholder="Add a new todo..."
                   style="padding: 10px; width: 70%;">
            <button onclick="send_event('{{.IdComponent}}', 'AddTodo', document.getElementById('todo-input').value)"
                    style="padding: 10px 20px;">
                Add
            </button>
        </div>
        
        <ul style="list-style: none; padding: 0;">
            {{range .Todos}}
            <li style="padding: 10px; border-bottom: 1px solid #ccc;">
                <input type="checkbox" 
                       {{if .Completed}}checked{{end}}
                       onclick="send_event('{{$.IdComponent}}', 'ToggleTodo', {{.ID}})">
                <span style="{{if .Completed}}text-decoration: line-through;{{end}}">
                    {{.Text}}
                </span>
                <button onclick="send_event('{{$.IdComponent}}', 'DeleteTodo', {{.ID}})"
                        style="float: right; color: red;">
                    Delete
                </button>
            </li>
            {{end}}
        </ul>
    </div>
    `
}

func (t *TodoList) GetDriver() liveview.LiveDriver {
    return t.ComponentDriver
}

func (t *TodoList) AddTodo(data interface{}) {
    if text, ok := data.(string); ok && text != "" {
        newTodo := Todo{
            ID:        len(t.Todos) + 1,
            Text:      text,
            Completed: false,
        }
        t.Todos = append(t.Todos, newTodo)
        t.Input = ""
        t.Commit()
    }
}

func (t *TodoList) ToggleTodo(data interface{}) {
    if id, ok := data.(float64); ok {
        for i := range t.Todos {
            if t.Todos[i].ID == int(id) {
                t.Todos[i].Completed = !t.Todos[i].Completed
                break
            }
        }
        t.Commit()
    }
}

func (t *TodoList) DeleteTodo(data interface{}) {
    if id, ok := data.(float64); ok {
        newTodos := []Todo{}
        for _, todo := range t.Todos {
            if todo.ID != int(id) {
                newTodos = append(newTodos, todo)
            }
        }
        t.Todos = newTodos
        t.Commit()
    }
}
```

## Handling Events

### Event Flow

1. User interaction triggers JavaScript `send_event()` function
2. Event is sent via WebSocket to server
3. Server finds matching method on component
4. Method updates component state
5. `Commit()` sends updates back to browser
6. DOM is updated in real-time

### Event Handler Patterns

```go
// Simple event without data
func (c *Component) Click(data interface{}) {
    // Handle click
    c.Commit()
}

// Event with string data
func (c *Component) UpdateText(data interface{}) {
    if text, ok := data.(string); ok {
        c.Text = text
        c.Commit()
    }
}

// Event with JSON data
func (c *Component) UpdateForm(data interface{}) {
    if formData, ok := data.(map[string]interface{}); ok {
        c.Name = formData["name"].(string)
        c.Email = formData["email"].(string)
        c.Commit()
    }
}
```

## Working with State

### State Management Best Practices

1. **Initialize in Start():**
```go
func (c *Component) Start() {
    c.Users = make([]User, 0)
    c.Loading = true
    c.LoadData()
    c.Commit()
}
```

2. **Update and Commit:**
```go
func (c *Component) UpdateUser(data interface{}) {
    // Update state
    c.User.Name = "New Name"
    
    // Always call Commit() to update UI
    c.Commit()
}
```

3. **Async Operations:**
```go
func (c *Component) LoadData() {
    go func() {
        // Simulate API call
        time.Sleep(2 * time.Second)
        
        c.Users = fetchUsersFromAPI()
        c.Loading = false
        c.Commit() // Safe to call from goroutine
    }()
}
```

## Component Composition

### Parent-Child Components

```go
type Dashboard struct {
    *liveview.ComponentDriver[*Dashboard]
    Header  *Header
    Sidebar *Sidebar
    Content *Content
}

func (d *Dashboard) Start() {
    // Create child components
    d.Header = &Header{Title: "Dashboard"}
    d.Sidebar = &Sidebar{Items: getMenuItems()}
    d.Content = &Content{}
    
    // Mount child components
    d.Mount(liveview.NewDriver("header", d.Header))
    d.Mount(liveview.NewDriver("sidebar", d.Sidebar))
    d.Mount(liveview.NewDriver("content", d.Content))
    
    d.Commit()
}

func (d *Dashboard) GetTemplate() string {
    return `
    <div class="dashboard">
        {{mount "header"}}
        <div class="layout">
            {{mount "sidebar"}}
            {{mount "content"}}
        </div>
    </div>
    `
}
```

## Real-time Features

### Live Updates Example

```go
type LiveChart struct {
    *liveview.ComponentDriver[*LiveChart]
    Data   []DataPoint
    ticker *time.Ticker
}

func (l *LiveChart) Start() {
    l.Data = make([]DataPoint, 0)
    
    // Update chart every second
    l.ticker = time.NewTicker(1 * time.Second)
    go func() {
        for range l.ticker.C {
            l.Data = append(l.Data, DataPoint{
                Time:  time.Now(),
                Value: rand.Float64() * 100,
            })
            
            // Keep only last 20 points
            if len(l.Data) > 20 {
                l.Data = l.Data[1:]
            }
            
            l.Commit() // Update UI
        }
    }()
    
    l.Commit()
}

func (l *LiveChart) Stop() {
    if l.ticker != nil {
        l.ticker.Stop()
    }
}
```

### Broadcasting to Multiple Clients

```go
func (c *ChatRoom) SendMessage(data interface{}) {
    if msg, ok := data.(string); ok {
        // Add message to history
        c.Messages = append(c.Messages, Message{
            User: c.CurrentUser,
            Text: msg,
            Time: time.Now(),
        })
        
        // Broadcast to all connected clients
        c.Broadcast(func(driver LiveDriver) {
            driver.Commit()
        })
    }
}
```

## Testing Your Components

### Unit Testing

```go
func TestCounter(t *testing.T) {
    // Create component
    counter := &Counter{}
    td := liveview.NewTestDriver(t, counter, "test-counter")
    defer td.Cleanup()
    
    // Test initial state
    assert.Equal(t, 0, counter.Count)
    
    // Test increment
    td.SimulateEvent("Increment", nil)
    assert.Equal(t, 1, counter.Count)
    
    // Test HTML output
    td.AssertHTML(t, "Counter: 1")
    
    // Test reset
    td.SimulateEvent("Reset", nil)
    assert.Equal(t, 0, counter.Count)
}
```

### Integration Testing

```go
func TestTodoListIntegration(t *testing.T) {
    suite := liveview.NewIntegrationTestSuite(t)
    
    // Setup page
    suite.SetupPage("/", "Todo App", func() liveview.LiveDriver {
        todo := &TodoList{}
        driver := liveview.NewDriver("todo", todo)
        todo.ComponentDriver = driver
        return driver
    })
    
    suite.Start()
    
    // Create client
    client := suite.NewClient("test-client")
    err := client.Connect(suite.WebSocketURL)
    require.NoError(t, err)
    
    // Add todo
    err = client.SendEvent("todo", "AddTodo", "Test todo")
    require.NoError(t, err)
    
    // Wait for update
    msg, err := client.WaitForMessage("fill", 2*time.Second)
    require.NoError(t, err)
    assert.NotNil(t, msg)
}
```

## Deployment

### Building for Production

1. **Create Dockerfile:**
```dockerfile
FROM golang:1.19-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
COPY --from=builder /app/assets ./assets
EXPOSE 8080
CMD ["./main"]
```

2. **Build and run:**
```bash
docker build -t my-liveview-app .
docker run -p 8080:8080 my-liveview-app
```

### Environment Configuration

```go
func main() {
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }
    
    e := echo.New()
    
    // Production settings
    if os.Getenv("ENV") == "production" {
        e.Debug = false
        e.HideBanner = true
    }
    
    // ... rest of setup
    
    e.Logger.Fatal(e.Start(":" + port))
}
```

## Advanced Topics

### Custom Middleware

```go
func AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
    return func(c echo.Context) error {
        // Check authentication
        session := c.Get("session")
        if session == nil {
            return c.Redirect(http.StatusTemporaryRedirect, "/login")
        }
        return next(c)
    }
}

// Use in page setup
home := liveview.PageControl{
    Title:      "Protected Page",
    Path:       "/dashboard",
    Router:     e,
    Middleware: []echo.MiddlewareFunc{AuthMiddleware},
}
```

### Database Integration

```go
type UserList struct {
    *liveview.ComponentDriver[*UserList]
    Users []User
    db    *sql.DB
}

func (u *UserList) Start() {
    u.db = database.GetConnection()
    u.LoadUsers()
    u.Commit()
}

func (u *UserList) LoadUsers() {
    rows, err := u.db.Query("SELECT id, name, email FROM users")
    if err != nil {
        log.Printf("Error loading users: %v", err)
        return
    }
    defer rows.Close()
    
    u.Users = []User{}
    for rows.Next() {
        var user User
        rows.Scan(&user.ID, &user.Name, &user.Email)
        u.Users = append(u.Users, user)
    }
}

func (u *UserList) DeleteUser(data interface{}) {
    if id, ok := data.(float64); ok {
        _, err := u.db.Exec("DELETE FROM users WHERE id = ?", int(id))
        if err == nil {
            u.LoadUsers()
            u.Commit()
        }
    }
}
```

### Performance Optimization

```go
// Use debouncing for frequent updates
type SearchBox struct {
    *liveview.ComponentDriver[*SearchBox]
    Query       string
    Results     []Result
    debouncer   *time.Timer
}

func (s *SearchBox) Search(data interface{}) {
    if query, ok := data.(string); ok {
        s.Query = query
        
        // Cancel previous timer
        if s.debouncer != nil {
            s.debouncer.Stop()
        }
        
        // Debounce search
        s.debouncer = time.AfterFunc(300*time.Millisecond, func() {
            s.Results = performSearch(s.Query)
            s.Commit()
        })
    }
}
```

---

## Español

# Tutorial Completo Paso a Paso

## Tabla de Contenidos
1. [Comenzando](#comenzando)
2. [Tu Primera App LiveView](#tu-primera-app-liveview)
3. [Entendiendo Componentes](#entendiendo-componentes)
4. [Manejando Eventos](#manejando-eventos)
5. [Trabajando con Estado](#trabajando-con-estado)
6. [Composición de Componentes](#composición-de-componentes)
7. [Características en Tiempo Real](#características-en-tiempo-real)
8. [Probando tus Componentes](#probando-tus-componentes)
9. [Despliegue](#despliegue)
10. [Temas Avanzados](#temas-avanzados)

## Comenzando

### Prerequisitos
- Go 1.19 o superior instalado
- Conocimiento básico de programación en Go
- Un editor de texto (VS Code, GoLand, etc.)
- Acceso a terminal/línea de comandos

### Instalación

1. **Crear un nuevo proyecto Go:**
```bash
mkdir mi-app-liveview
cd mi-app-liveview
go mod init mi-app-liveview
```

2. **Instalar Go Echo LiveView:**
```bash
go get github.com/arturoeanton/go-echo-live-view
go get github.com/labstack/echo/v4
```

3. **Verificar instalación:**
```bash
go mod tidy
```

## Tu Primera App LiveView

Vamos a crear una aplicación contador simple que se actualiza en tiempo real.

### Paso 1: Crear main.go

Crea un archivo llamado `main.go`:

```go
package main

import (
    "github.com/arturoeanton/go-echo-live-view/liveview"
    "github.com/labstack/echo/v4"
    "github.com/labstack/echo/v4/middleware"
)

func main() {
    // Crear instancia de Echo
    e := echo.New()
    
    // Agregar middleware
    e.Use(middleware.Logger())
    e.Use(middleware.Recover())
    
    // Configurar página LiveView
    inicio := liveview.PageControl{
        Title:  "Mi Primera App LiveView",
        Path:   "/",
        Router: e,
    }
    
    // Registrar componente
    inicio.Register(func() liveview.LiveDriver {
        contador := &Contador{}
        driver := liveview.NewDriver("contador", contador)
        contador.ComponentDriver = driver
        return driver
    })
    
    // Iniciar servidor
    e.Logger.Fatal(e.Start(":8080"))
}
```

### Paso 2: Crear el Componente Contador

Agrega esto al mismo archivo:

```go
type Contador struct {
    *liveview.ComponentDriver[*Contador]
    Cuenta int
}

func (c *Contador) Start() {
    c.Cuenta = 0
    c.Commit() // Renderizar el componente
}

func (c *Contador) GetTemplate() string {
    return `
    <div style="text-align: center; padding: 50px;">
        <h1>Contador: {{.Cuenta}}</h1>
        <button onclick="send_event('{{.IdComponent}}', 'Incrementar')" 
                style="padding: 10px 20px; font-size: 16px;">
            Incrementar
        </button>
        <button onclick="send_event('{{.IdComponent}}', 'Decrementar')" 
                style="padding: 10px 20px; font-size: 16px;">
            Decrementar
        </button>
        <button onclick="send_event('{{.IdComponent}}', 'Reiniciar')" 
                style="padding: 10px 20px; font-size: 16px;">
            Reiniciar
        </button>
    </div>
    `
}

func (c *Contador) GetDriver() liveview.LiveDriver {
    return c.ComponentDriver
}

// Manejadores de eventos
func (c *Contador) Incrementar(data interface{}) {
    c.Cuenta++
    c.Commit() // Actualizar la UI
}

func (c *Contador) Decrementar(data interface{}) {
    c.Cuenta--
    c.Commit()
}

func (c *Contador) Reiniciar(data interface{}) {
    c.Cuenta = 0
    c.Commit()
}
```

### Paso 3: Ejecutar tu App

```bash
go run main.go
```

Abre tu navegador y navega a `http://localhost:8080`. ¡Deberías ver tu app contador!

## Entendiendo Componentes

### Estructura de Componentes

Todo componente LiveView debe implementar la interfaz `Component`:

```go
type Component interface {
    Start()                    // Inicializar componente
    GetTemplate() string       // Retornar plantilla HTML
    GetDriver() LiveDriver     // Retornar el driver
}
```

### Ejemplo: Componente Lista de Tareas

Creemos un componente más complejo - una lista de tareas:

```go
type ListaTareas struct {
    *liveview.ComponentDriver[*ListaTareas]
    Tareas []Tarea
    Input  string
}

type Tarea struct {
    ID         int
    Texto      string
    Completada bool
}

func (l *ListaTareas) Start() {
    l.Tareas = []Tarea{
        {ID: 1, Texto: "Aprender LiveView", Completada: false},
        {ID: 2, Texto: "Construir una app", Completada: false},
    }
    l.Commit()
}

func (l *ListaTareas) GetTemplate() string {
    return `
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h1>Lista de Tareas</h1>
        
        <div style="margin-bottom: 20px;">
            <input type="text" 
                   id="tarea-input" 
                   value="{{.Input}}"
                   onkeyup="if(event.key === 'Enter') send_event('{{.IdComponent}}', 'AgregarTarea', this.value)"
                   placeholder="Agregar nueva tarea..."
                   style="padding: 10px; width: 70%;">
            <button onclick="send_event('{{.IdComponent}}', 'AgregarTarea', document.getElementById('tarea-input').value)"
                    style="padding: 10px 20px;">
                Agregar
            </button>
        </div>
        
        <ul style="list-style: none; padding: 0;">
            {{range .Tareas}}
            <li style="padding: 10px; border-bottom: 1px solid #ccc;">
                <input type="checkbox" 
                       {{if .Completada}}checked{{end}}
                       onclick="send_event('{{$.IdComponent}}', 'AlternarTarea', {{.ID}})">
                <span style="{{if .Completada}}text-decoration: line-through;{{end}}">
                    {{.Texto}}
                </span>
                <button onclick="send_event('{{$.IdComponent}}', 'EliminarTarea', {{.ID}})"
                        style="float: right; color: red;">
                    Eliminar
                </button>
            </li>
            {{end}}
        </ul>
    </div>
    `
}

func (l *ListaTareas) GetDriver() liveview.LiveDriver {
    return l.ComponentDriver
}

func (l *ListaTareas) AgregarTarea(data interface{}) {
    if texto, ok := data.(string); ok && texto != "" {
        nuevaTarea := Tarea{
            ID:         len(l.Tareas) + 1,
            Texto:      texto,
            Completada: false,
        }
        l.Tareas = append(l.Tareas, nuevaTarea)
        l.Input = ""
        l.Commit()
    }
}

func (l *ListaTareas) AlternarTarea(data interface{}) {
    if id, ok := data.(float64); ok {
        for i := range l.Tareas {
            if l.Tareas[i].ID == int(id) {
                l.Tareas[i].Completada = !l.Tareas[i].Completada
                break
            }
        }
        l.Commit()
    }
}

func (l *ListaTareas) EliminarTarea(data interface{}) {
    if id, ok := data.(float64); ok {
        nuevasTareas := []Tarea{}
        for _, tarea := range l.Tareas {
            if tarea.ID != int(id) {
                nuevasTareas = append(nuevasTareas, tarea)
            }
        }
        l.Tareas = nuevasTareas
        l.Commit()
    }
}
```

## Manejando Eventos

### Flujo de Eventos

1. Interacción del usuario activa función JavaScript `send_event()`
2. Evento se envía vía WebSocket al servidor
3. Servidor encuentra método coincidente en componente
4. Método actualiza estado del componente
5. `Commit()` envía actualizaciones de vuelta al navegador
6. DOM se actualiza en tiempo real

### Patrones de Manejadores de Eventos

```go
// Evento simple sin datos
func (c *Componente) Click(data interface{}) {
    // Manejar click
    c.Commit()
}

// Evento con datos de cadena
func (c *Componente) ActualizarTexto(data interface{}) {
    if texto, ok := data.(string); ok {
        c.Texto = texto
        c.Commit()
    }
}

// Evento con datos JSON
func (c *Componente) ActualizarFormulario(data interface{}) {
    if datosForm, ok := data.(map[string]interface{}); ok {
        c.Nombre = datosForm["nombre"].(string)
        c.Email = datosForm["email"].(string)
        c.Commit()
    }
}
```

## Trabajando con Estado

### Mejores Prácticas de Gestión de Estado

1. **Inicializar en Start():**
```go
func (c *Componente) Start() {
    c.Usuarios = make([]Usuario, 0)
    c.Cargando = true
    c.CargarDatos()
    c.Commit()
}
```

2. **Actualizar y Confirmar:**
```go
func (c *Componente) ActualizarUsuario(data interface{}) {
    // Actualizar estado
    c.Usuario.Nombre = "Nuevo Nombre"
    
    // Siempre llamar Commit() para actualizar UI
    c.Commit()
}
```

3. **Operaciones Asíncronas:**
```go
func (c *Componente) CargarDatos() {
    go func() {
        // Simular llamada API
        time.Sleep(2 * time.Second)
        
        c.Usuarios = obtenerUsuariosDeAPI()
        c.Cargando = false
        c.Commit() // Seguro llamar desde goroutine
    }()
}
```

## Composición de Componentes

### Componentes Padre-Hijo

```go
type Tablero struct {
    *liveview.ComponentDriver[*Tablero]
    Cabecera    *Cabecera
    BarraLateral *BarraLateral
    Contenido   *Contenido
}

func (t *Tablero) Start() {
    // Crear componentes hijos
    t.Cabecera = &Cabecera{Titulo: "Tablero"}
    t.BarraLateral = &BarraLateral{Items: obtenerItemsMenu()}
    t.Contenido = &Contenido{}
    
    // Montar componentes hijos
    t.Mount(liveview.NewDriver("cabecera", t.Cabecera))
    t.Mount(liveview.NewDriver("barra-lateral", t.BarraLateral))
    t.Mount(liveview.NewDriver("contenido", t.Contenido))
    
    t.Commit()
}

func (t *Tablero) GetTemplate() string {
    return `
    <div class="tablero">
        {{mount "cabecera"}}
        <div class="diseño">
            {{mount "barra-lateral"}}
            {{mount "contenido"}}
        </div>
    </div>
    `
}
```

## Características en Tiempo Real

### Ejemplo de Actualizaciones en Vivo

```go
type GraficoEnVivo struct {
    *liveview.ComponentDriver[*GraficoEnVivo]
    Datos  []PuntoDatos
    ticker *time.Ticker
}

func (g *GraficoEnVivo) Start() {
    g.Datos = make([]PuntoDatos, 0)
    
    // Actualizar gráfico cada segundo
    g.ticker = time.NewTicker(1 * time.Second)
    go func() {
        for range g.ticker.C {
            g.Datos = append(g.Datos, PuntoDatos{
                Tiempo: time.Now(),
                Valor:  rand.Float64() * 100,
            })
            
            // Mantener solo últimos 20 puntos
            if len(g.Datos) > 20 {
                g.Datos = g.Datos[1:]
            }
            
            g.Commit() // Actualizar UI
        }
    }()
    
    g.Commit()
}

func (g *GraficoEnVivo) Stop() {
    if g.ticker != nil {
        g.ticker.Stop()
    }
}
```

## Probando tus Componentes

### Pruebas Unitarias

```go
func TestContador(t *testing.T) {
    // Crear componente
    contador := &Contador{}
    td := liveview.NewTestDriver(t, contador, "test-contador")
    defer td.Cleanup()
    
    // Probar estado inicial
    assert.Equal(t, 0, contador.Cuenta)
    
    // Probar incremento
    td.SimulateEvent("Incrementar", nil)
    assert.Equal(t, 1, contador.Cuenta)
    
    // Probar salida HTML
    td.AssertHTML(t, "Contador: 1")
    
    // Probar reinicio
    td.SimulateEvent("Reiniciar", nil)
    assert.Equal(t, 0, contador.Cuenta)
}
```

## Despliegue

### Construyendo para Producción

1. **Crear Dockerfile:**
```dockerfile
FROM golang:1.19-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
COPY --from=builder /app/assets ./assets
EXPOSE 8080
CMD ["./main"]
```

2. **Construir y ejecutar:**
```bash
docker build -t mi-app-liveview .
docker run -p 8080:8080 mi-app-liveview
```

## Temas Avanzados

### Integración con Base de Datos

```go
type ListaUsuarios struct {
    *liveview.ComponentDriver[*ListaUsuarios]
    Usuarios []Usuario
    db       *sql.DB
}

func (l *ListaUsuarios) Start() {
    l.db = database.ObtenerConexion()
    l.CargarUsuarios()
    l.Commit()
}

func (l *ListaUsuarios) CargarUsuarios() {
    rows, err := l.db.Query("SELECT id, nombre, email FROM usuarios")
    if err != nil {
        log.Printf("Error cargando usuarios: %v", err)
        return
    }
    defer rows.Close()
    
    l.Usuarios = []Usuario{}
    for rows.Next() {
        var usuario Usuario
        rows.Scan(&usuario.ID, &usuario.Nombre, &usuario.Email)
        l.Usuarios = append(l.Usuarios, usuario)
    }
}
```

### Optimización de Rendimiento

```go
// Usar debouncing para actualizaciones frecuentes
type CajaBusqueda struct {
    *liveview.ComponentDriver[*CajaBusqueda]
    Consulta    string
    Resultados  []Resultado
    debouncer   *time.Timer
}

func (c *CajaBusqueda) Buscar(data interface{}) {
    if consulta, ok := data.(string); ok {
        c.Consulta = consulta
        
        // Cancelar timer anterior
        if c.debouncer != nil {
            c.debouncer.Stop()
        }
        
        // Debounce búsqueda
        c.debouncer = time.AfterFunc(300*time.Millisecond, func() {
            c.Resultados = realizarBusqueda(c.Consulta)
            c.Commit()
        })
    }
}
```

---

## Next Steps / Próximos Pasos

### English
- Explore the [example](../example/) directory for more complex examples
- Read the [API Documentation](../API_DOCUMENTATION.md) for detailed reference
- Join the community discussions

### Español
- Explora el directorio [example](../example/) para ejemplos más complejos
- Lee la [Documentación API](../API_DOCUMENTATION.md) para referencia detallada
- Únete a las discusiones de la comunidad