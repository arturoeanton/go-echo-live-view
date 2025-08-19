# Go Echo LiveView

[English](#english) | [Español](#español)

---

## English

# Go Echo LiveView - Real-time Web Framework for Go

[![Go Version](https://img.shields.io/badge/Go-1.19+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-blue)](LICENSE)
[![Documentation](https://img.shields.io/badge/docs-API%20Reference-green)](API_DOCUMENTATION.md)

Go Echo LiveView is a powerful real-time web framework for Go that enables server-side rendering with WebSocket-based reactivity. Build interactive web applications without writing JavaScript, inspired by Phoenix LiveView.

## ✨ Features

- 🚀 **Real-time Updates**: Automatic DOM synchronization via WebSocket
- 🎯 **Server-Side Rendering**: All logic stays on the server
- 🔧 **Component-Based**: Reusable, composable components
- 🛡️ **Built-in Security**: Input validation, sanitization, and rate limiting
- 📦 **Rich Component Library**: Forms, tables, modals, charts, and more
- 🧪 **Testing Framework**: Comprehensive testing utilities included
- 💾 **Memory Efficient**: Context-based lifecycle management
- 🎨 **No JavaScript Required**: Build interactive UIs with pure Go

## 🚀 Quick Start

### Installation

```bash
go get github.com/arturoeanton/go-echo-live-view
```

### Hello World Example

```go
package main

import (
    "github.com/arturoeanton/go-echo-live-view/liveview"
    "github.com/labstack/echo/v4"
)

type HelloWorld struct {
    *liveview.ComponentDriver[*HelloWorld]
    Message string
}

func (h *HelloWorld) Start() {
    h.Message = "Hello, LiveView!"
    h.Commit()
}

func (h *HelloWorld) GetTemplate() string {
    return `<div>
        <h1>{{.Message}}</h1>
        <button onclick="send_event('{{.IdComponent}}', 'Click')">Click Me!</button>
    </div>`
}

func (h *HelloWorld) GetDriver() liveview.LiveDriver {
    return h
}

func (h *HelloWorld) Click(data interface{}) {
    h.Message = "Button clicked!"
    h.Commit()
}

func main() {
    e := echo.New()
    
    page := liveview.PageControl{
        Title:  "Hello World",
        Path:   "/",
        Router: e,
    }
    
    page.Register(func() liveview.LiveDriver {
        hello := &HelloWorld{}
        driver := liveview.NewDriver("hello", hello)
        hello.ComponentDriver = driver
        return driver
    })
    
    e.Logger.Fatal(e.Start(":8080"))
}
```

Visit `http://localhost:8080` and see your interactive app without any JavaScript!

## 📚 Documentation

- [API Documentation](API_DOCUMENTATION.md) - Complete API reference
- [Examples](example/) - Working examples and demos
- [Testing Guide](docs/testing.md) - Testing your components
- [Security Guide](docs/security.md) - Security best practices

## 🧩 Built-in Components

### UI Components
- **Table**: Sortable, filterable data tables with pagination
- **Form**: Form builder with validation
- **Modal**: Dialog windows with callbacks
- **Chart**: Bar, line, and pie charts
- **Calendar**: Date picker with events
- **Accordion**: Collapsible content panels
- **Sidebar**: Navigation sidebar
- **Alert**: Dismissible notifications
- **Dropdown**: Select menus with icons
- **Card**: Content cards with actions

### Advanced Components
- **FileUpload**: Drag-and-drop file uploads
- **RichEditor**: WYSIWYG text editor
- **Draggable**: Drag-and-drop interfaces
- **Animation**: CSS animations framework
- **NotificationSystem**: Toast notifications

## 🧪 Testing

The framework includes a comprehensive testing suite:

```go
func TestMyComponent(t *testing.T) {
    component := &MyComponent{}
    td := liveview.NewTestDriver(t, component, "test-component")
    defer td.Cleanup()
    
    // Test initial state
    td.AssertHTML(t, "Expected content")
    
    // Simulate events
    td.SimulateEvent("Click", nil)
    
    // Verify updates
    assert.Equal(t, "Updated", component.State)
}
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run benchmarks
go test -bench=. ./...
```

## 🔒 Security Features

- **Input Validation**: All WebSocket messages are validated
- **Template Sanitization**: Automatic XSS protection
- **Path Traversal Protection**: Safe file path handling
- **Rate Limiting**: Built-in request throttling
- **Context Cancellation**: Proper resource cleanup
- **Memory Management**: No memory leaks

## 📁 Project Structure

```
├── liveview/           # Core framework
│   ├── model.go        # Component system
│   ├── page_content.go # Page and WebSocket handling
│   ├── layout.go       # Layout system
│   ├── testing.go      # Testing utilities
│   └── security.go     # Security features
├── components/         # Built-in components
│   ├── table.go
│   ├── form.go
│   ├── modal.go
│   └── ...
├── example/           # Example applications
│   ├── example1/      # Basic counter
│   ├── example_todo/  # Todo list
│   └── ...
└── assets/           # Static assets
    ├── json.wasm     # WebAssembly module
    └── wasm_exec.js  # WASM executor
```

## 🤝 Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

### Development Setup

```bash
# Clone repository
git clone https://github.com/arturoeanton/go-echo-live-view
cd go-echo-live-view

# Install dependencies
go mod tidy

# Build WASM module
cd cmd/wasm/
GOOS=js GOARCH=wasm go build -o ../../assets/json.wasm
cd ../..

# Run with auto-reload (requires gomon)
go install github.com/c9s/gomon@latest
gomon
```

## 📈 Performance

- **Low Latency**: Sub-millisecond DOM updates
- **Efficient**: Minimal bandwidth usage
- **Scalable**: Handles thousands of concurrent connections
- **Optimized**: Smart diffing and patching

## 🌟 Use Cases

- **Admin Dashboards**: Real-time metrics and controls
- **Collaborative Tools**: Multi-user applications
- **Live Forms**: Dynamic form validation
- **Data Visualization**: Real-time charts and graphs
- **Chat Applications**: Instant messaging
- **Monitoring Systems**: Live status updates

## 📄 License

MIT License - see [LICENSE](LICENSE) file

---

## Español

# Go Echo LiveView - Framework Web en Tiempo Real para Go

[![Versión Go](https://img.shields.io/badge/Go-1.19+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![Licencia](https://img.shields.io/badge/licencia-MIT-blue)](LICENSE)
[![Documentación](https://img.shields.io/badge/docs-Referencia%20API-green)](API_DOCUMENTATION.md)

Go Echo LiveView es un potente framework web en tiempo real para Go que permite renderizado del lado del servidor con reactividad basada en WebSocket. Construye aplicaciones web interactivas sin escribir JavaScript, inspirado en Phoenix LiveView.

## ✨ Características

- 🚀 **Actualizaciones en Tiempo Real**: Sincronización automática del DOM vía WebSocket
- 🎯 **Renderizado del Servidor**: Toda la lógica permanece en el servidor
- 🔧 **Basado en Componentes**: Componentes reutilizables y componibles
- 🛡️ **Seguridad Integrada**: Validación de entrada, sanitización y limitación de tasa
- 📦 **Rica Biblioteca de Componentes**: Formularios, tablas, modales, gráficos y más
- 🧪 **Framework de Testing**: Utilidades de prueba completas incluidas
- 💾 **Eficiente en Memoria**: Gestión de ciclo de vida basada en contexto
- 🎨 **Sin JavaScript Requerido**: Construye UIs interactivas con Go puro

## 🚀 Inicio Rápido

### Instalación

```bash
go get github.com/arturoeanton/go-echo-live-view
```

### Ejemplo Hola Mundo

```go
package main

import (
    "github.com/arturoeanton/go-echo-live-view/liveview"
    "github.com/labstack/echo/v4"
)

type HolaMundo struct {
    *liveview.ComponentDriver[*HolaMundo]
    Mensaje string
}

func (h *HolaMundo) Start() {
    h.Mensaje = "¡Hola, LiveView!"
    h.Commit()
}

func (h *HolaMundo) GetTemplate() string {
    return `<div>
        <h1>{{.Mensaje}}</h1>
        <button onclick="send_event('{{.IdComponent}}', 'Click')">¡Haz Click!</button>
    </div>`
}

func (h *HolaMundo) GetDriver() liveview.LiveDriver {
    return h
}

func (h *HolaMundo) Click(data interface{}) {
    h.Mensaje = "¡Botón presionado!"
    h.Commit()
}

func main() {
    e := echo.New()
    
    pagina := liveview.PageControl{
        Title:  "Hola Mundo",
        Path:   "/",
        Router: e,
    }
    
    pagina.Register(func() liveview.LiveDriver {
        hola := &HolaMundo{}
        driver := liveview.NewDriver("hola", hola)
        hola.ComponentDriver = driver
        return driver
    })
    
    e.Logger.Fatal(e.Start(":8080"))
}
```

Visita `http://localhost:8080` y ve tu aplicación interactiva ¡sin JavaScript!

## 📚 Documentación

- [Documentación API](API_DOCUMENTATION.md) - Referencia API completa
- [Ejemplos](example/) - Ejemplos funcionales y demos
- [Guía de Testing](docs/testing.md) - Prueba tus componentes
- [Guía de Seguridad](docs/security.md) - Mejores prácticas de seguridad

## 🧩 Componentes Integrados

### Componentes UI
- **Table**: Tablas de datos ordenables y filtrables con paginación
- **Form**: Constructor de formularios con validación
- **Modal**: Ventanas de diálogo con callbacks
- **Chart**: Gráficos de barras, líneas y pastel
- **Calendar**: Selector de fecha con eventos
- **Accordion**: Paneles de contenido colapsables
- **Sidebar**: Barra lateral de navegación
- **Alert**: Notificaciones descartables
- **Dropdown**: Menús de selección con iconos
- **Card**: Tarjetas de contenido con acciones

### Componentes Avanzados
- **FileUpload**: Carga de archivos arrastrar y soltar
- **RichEditor**: Editor de texto WYSIWYG
- **Draggable**: Interfaces arrastrar y soltar
- **Animation**: Framework de animaciones CSS
- **NotificationSystem**: Notificaciones toast

## 🧪 Testing

El framework incluye una suite de testing completa:

```go
func TestMiComponente(t *testing.T) {
    componente := &MiComponente{}
    td := liveview.NewTestDriver(t, componente, "test-componente")
    defer td.Cleanup()
    
    // Probar estado inicial
    td.AssertHTML(t, "Contenido esperado")
    
    // Simular eventos
    td.SimulateEvent("Click", nil)
    
    // Verificar actualizaciones
    assert.Equal(t, "Actualizado", componente.Estado)
}
```

### Ejecutar Pruebas

```bash
# Ejecutar todas las pruebas
go test ./...

# Ejecutar con cobertura
go test -cover ./...

# Ejecutar benchmarks
go test -bench=. ./...
```

## 🔒 Características de Seguridad

- **Validación de Entrada**: Todos los mensajes WebSocket son validados
- **Sanitización de Plantillas**: Protección XSS automática
- **Protección Path Traversal**: Manejo seguro de rutas de archivos
- **Limitación de Tasa**: Throttling de peticiones integrado
- **Cancelación de Contexto**: Limpieza adecuada de recursos
- **Gestión de Memoria**: Sin fugas de memoria

## 📁 Estructura del Proyecto

```
├── liveview/           # Framework principal
│   ├── model.go        # Sistema de componentes
│   ├── page_content.go # Manejo de páginas y WebSocket
│   ├── layout.go       # Sistema de layouts
│   ├── testing.go      # Utilidades de prueba
│   └── security.go     # Características de seguridad
├── components/         # Componentes integrados
│   ├── table.go
│   ├── form.go
│   ├── modal.go
│   └── ...
├── example/           # Aplicaciones de ejemplo
│   ├── example1/      # Contador básico
│   ├── example_todo/  # Lista de tareas
│   └── ...
└── assets/           # Archivos estáticos
    ├── json.wasm     # Módulo WebAssembly
    └── wasm_exec.js  # Ejecutor WASM
```

## 🤝 Contribuyendo

¡Damos la bienvenida a las contribuciones! Por favor, consulta [CONTRIBUTING.md](CONTRIBUTING.md) para las pautas.

### Configuración de Desarrollo

```bash
# Clonar repositorio
git clone https://github.com/arturoeanton/go-echo-live-view
cd go-echo-live-view

# Instalar dependencias
go mod tidy

# Compilar módulo WASM
cd cmd/wasm/
GOOS=js GOARCH=wasm go build -o ../../assets/json.wasm
cd ../..

# Ejecutar con auto-reload (requiere gomon)
go install github.com/c9s/gomon@latest
gomon
```

## 📈 Rendimiento

- **Baja Latencia**: Actualizaciones DOM en sub-milisegundos
- **Eficiente**: Uso mínimo de ancho de banda
- **Escalable**: Maneja miles de conexiones concurrentes
- **Optimizado**: Diffing y patching inteligente

## 🌟 Casos de Uso

- **Dashboards Administrativos**: Métricas y controles en tiempo real
- **Herramientas Colaborativas**: Aplicaciones multi-usuario
- **Formularios en Vivo**: Validación dinámica de formularios
- **Visualización de Datos**: Gráficos en tiempo real
- **Aplicaciones de Chat**: Mensajería instantánea
- **Sistemas de Monitoreo**: Actualizaciones de estado en vivo

## 📄 Licencia

Licencia MIT - ver archivo [LICENSE](LICENSE)

## 🚧 Roadmap

### Próximas Características
- [ ] Soporte para clustering
- [ ] Persistencia de sesión Redis
- [ ] Más componentes UI
- [ ] Extensión VS Code
- [ ] CLI para scaffolding

### Completado Recientemente
- ✅ Framework de testing completo
- ✅ Documentación bilingüe
- ✅ Gestión de memoria mejorada
- ✅ 15+ componentes UI
- ✅ Seguridad reforzada

## 📞 Contacto y Soporte

- **GitHub Issues**: [Reportar problemas](https://github.com/arturoeanton/go-echo-live-view/issues)
- **Discussions**: [Preguntas y discusiones](https://github.com/arturoeanton/go-echo-live-view/discussions)

## 🌐 Comunidad

Únete a nuestra comunidad creciente de desarrolladores construyendo aplicaciones web en tiempo real con Go.

### Proyectos Usando Go Echo LiveView
- Sistema de monitoreo en tiempo real
- Dashboard de administración
- Plataforma de chat
- Herramienta de colaboración

¿Usas Go Echo LiveView? ¡Añade tu proyecto a la lista!

---

Made with ❤️ by the Go Echo LiveView community