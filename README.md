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
}

func (h *HelloWorld) GetTemplate() string {
    return `<div>
        <h1>{{.Message}}</h1>
        <button onclick="send_event('{{.IdComponent}}', 'Click')">Click Me!</button>
    </div>`
}

func (h *HelloWorld) Click(data interface{}) {
    h.Message = "Button clicked!"
    h.Commit()
}

func main() {
    e := echo.New()
    
    page := &liveview.PageControl{
        Title:  "Hello World",
        Path:   "/",
        Router: e,
    }
    
    page.Register(func() liveview.LiveDriver {
        hello := &HelloWorld{}
        hello.ComponentDriver = liveview.NewDriver("hello", hello)
        return hello.ComponentDriver
    })
    
    e.Logger.Fatal(e.Start(":8080"))
}
```

Visit `http://localhost:8080` and see your interactive app without any JavaScript!

## 📚 Documentation

- [API Documentation](API_DOCUMENTATION.md) - Complete API reference
- [Examples](example/) - Working examples and demos
- [Testing Guide](liveview/testing_test.go) - Testing your components
- [Component Library](components/) - Built-in components

## 🎯 Examples

The `example/` directory contains various demonstration applications:

### Basic Examples
- **example1-4**: Progressive complexity demos
- **clock_ticking**: Real-time clock display
- **collaborative_editing**: Multi-user text editor
- **counter**: Simple increment/decrement counter

### Advanced Examples
- **kanban_simple**: Full-featured Kanban board with drag-and-drop
- **todo_list**: Task management with persistence
- **chat_app**: Real-time messaging
- **dashboard**: Analytics dashboard with charts
- **form_validation**: Dynamic form with validation

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
- **Breadcrumb**: Navigation breadcrumbs
- **Tabs**: Native tab components

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
│   ├── kanban_simple/ # Kanban board demo
│   ├── todo_list/     # Todo list demo
│   └── ...
└── assets/           # Static assets
    └── live.js       # Client-side LiveView handler
```

## 🤝 Contributing

We welcome contributions! Please see the guidelines below:

### Development Setup

```bash
# Clone repository
git clone https://github.com/arturoeanton/go-echo-live-view
cd go-echo-live-view

# Install dependencies
go mod tidy

# Run tests
go test ./...

# Run with auto-reload (requires gomon)
go install github.com/c9s/gomon@latest
gomon
```

### Contribution Guidelines

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

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
- **Kanban Boards**: Project management tools
- **E-commerce**: Live inventory and pricing

## 🗺️ Roadmap

- [ ] TypeScript client library
- [ ] Component marketplace
- [ ] Visual component designer
- [ ] Performance profiling tools
- [ ] Enhanced debugging capabilities
- [ ] Mobile-optimized components
- [ ] Offline support
- [ ] GraphQL integration

## 📄 License

MIT License - see [LICENSE](LICENSE) file

## 🙏 Acknowledgments

- Inspired by [Phoenix LiveView](https://github.com/phoenixframework/phoenix_live_view)
- Built on [Echo Framework](https://echo.labstack.com/)
- Community contributors and testers

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
}

func (h *HolaMundo) GetTemplate() string {
    return `<div>
        <h1>{{.Mensaje}}</h1>
        <button onclick="send_event('{{.IdComponent}}', 'Click')">¡Haz Click!</button>
    </div>`
}

func (h *HolaMundo) Click(data interface{}) {
    h.Mensaje = "¡Botón presionado!"
    h.Commit()
}

func main() {
    e := echo.New()
    
    pagina := &liveview.PageControl{
        Title:  "Hola Mundo",
        Path:   "/",
        Router: e,
    }
    
    pagina.Register(func() liveview.LiveDriver {
        hola := &HolaMundo{}
        hola.ComponentDriver = liveview.NewDriver("hola", hola)
        return hola.ComponentDriver
    })
    
    e.Logger.Fatal(e.Start(":8080"))
}
```

¡Visita `http://localhost:8080` y ve tu aplicación interactiva sin JavaScript!

## 📚 Documentación

- [Documentación API](API_DOCUMENTATION.md) - Referencia API completa
- [Ejemplos](example/) - Ejemplos funcionales y demos
- [Guía de Testing](liveview/testing_test.go) - Prueba tus componentes
- [Biblioteca de Componentes](components/) - Componentes integrados

## 🎯 Ejemplos

El directorio `example/` contiene varias aplicaciones de demostración:

### Ejemplos Básicos
- **example1-4**: Demos de complejidad progresiva
- **clock_ticking**: Reloj en tiempo real
- **collaborative_editing**: Editor de texto multiusuario
- **counter**: Contador simple incremento/decremento

### Ejemplos Avanzados
- **kanban_simple**: Tablero Kanban completo con arrastrar y soltar
- **todo_list**: Gestión de tareas con persistencia
- **chat_app**: Mensajería en tiempo real
- **dashboard**: Panel de análisis con gráficos
- **form_validation**: Formulario dinámico con validación

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
- **Breadcrumb**: Migas de pan de navegación
- **Tabs**: Componentes de pestañas nativas

### Componentes Avanzados
- **FileUpload**: Carga de archivos arrastrar y soltar
- **RichEditor**: Editor de texto WYSIWYG
- **Draggable**: Interfaces arrastrar y soltar
- **Animation**: Framework de animaciones CSS
- **NotificationSystem**: Notificaciones toast

## 🧪 Testing

El framework incluye una suite de pruebas completa:

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
- **Limitación de Tasa**: Throttling de solicitudes integrado
- **Cancelación de Contexto**: Limpieza adecuada de recursos
- **Gestión de Memoria**: Sin fugas de memoria

## 📁 Estructura del Proyecto

```
├── liveview/           # Framework principal
│   ├── model.go        # Sistema de componentes
│   ├── page_content.go # Manejo de páginas y WebSocket
│   ├── layout.go       # Sistema de layouts
│   ├── testing.go      # Utilidades de testing
│   └── security.go     # Características de seguridad
├── components/         # Componentes integrados
│   ├── table.go
│   ├── form.go
│   ├── modal.go
│   └── ...
├── example/           # Aplicaciones de ejemplo
│   ├── kanban_simple/ # Demo de tablero Kanban
│   ├── todo_list/     # Demo de lista de tareas
│   └── ...
└── assets/           # Recursos estáticos
    └── live.js       # Manejador LiveView del cliente
```

## 🤝 Contribuir

¡Damos la bienvenida a las contribuciones! Por favor, consulta las pautas a continuación:

### Configuración de Desarrollo

```bash
# Clonar repositorio
git clone https://github.com/arturoeanton/go-echo-live-view
cd go-echo-live-view

# Instalar dependencias
go mod tidy

# Ejecutar pruebas
go test ./...

# Ejecutar con recarga automática (requiere gomon)
go install github.com/c9s/gomon@latest
gomon
```

### Pautas de Contribución

1. Haz un fork del repositorio
2. Crea una rama de característica (`git checkout -b feature/caracteristica-increible`)
3. Confirma tus cambios (`git commit -m 'Agregar característica increíble'`)
4. Empuja a la rama (`git push origin feature/caracteristica-increible`)
5. Abre un Pull Request

## 📈 Rendimiento

- **Baja Latencia**: Actualizaciones del DOM en sub-milisegundos
- **Eficiente**: Uso mínimo de ancho de banda
- **Escalable**: Maneja miles de conexiones concurrentes
- **Optimizado**: Diffing y parcheo inteligente

## 🌟 Casos de Uso

- **Paneles de Administración**: Métricas y controles en tiempo real
- **Herramientas Colaborativas**: Aplicaciones multiusuario
- **Formularios en Vivo**: Validación dinámica de formularios
- **Visualización de Datos**: Gráficos y tablas en tiempo real
- **Aplicaciones de Chat**: Mensajería instantánea
- **Sistemas de Monitoreo**: Actualizaciones de estado en vivo
- **Tableros Kanban**: Herramientas de gestión de proyectos
- **E-commerce**: Inventario y precios en vivo

## 🗺️ Hoja de Ruta

- [ ] Biblioteca cliente TypeScript
- [ ] Marketplace de componentes
- [ ] Diseñador visual de componentes
- [ ] Herramientas de perfilado de rendimiento
- [ ] Capacidades mejoradas de depuración
- [ ] Componentes optimizados para móvil
- [ ] Soporte offline
- [ ] Integración con GraphQL

## 📄 Licencia

Licencia MIT - ver archivo [LICENSE](LICENSE)

## 🙏 Agradecimientos

- Inspirado por [Phoenix LiveView](https://github.com/phoenixframework/phoenix_live_view)
- Construido sobre [Echo Framework](https://echo.labstack.com/)
- Contribuidores y testers de la comunidad