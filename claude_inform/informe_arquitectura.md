# Informe de Arquitectura - Go Echo LiveView

## 1. Visión General de la Arquitectura

Go Echo LiveView implementa una **arquitectura híbrida server-side rendering** con **comunicación en tiempo real** via WebSockets. El patrón principal es una adaptación del **Phoenix LiveView** de Elixir al ecosistema Go.

### 1.1 Paradigma Arquitectónico

```
┌─────────────────────────────────────────────────────────────────┐
│                    PATRÓN LIVEVIEW                              │
│                                                                 │
│  Server-Side Rendering + Real-time Bidirectional Communication │
└─────────────────────────────────────────────────────────────────┘
```

**Características principales**:
- **Stateful Server**: Estado de la aplicación mantenido en el servidor
- **Stateless Client**: Cliente mínimo, principalmente para renderizado
- **Event-Driven**: Comunicación basada en eventos DOM → Server
- **Reactive Updates**: Actualizaciones automáticas Server → DOM

## 2. Arquitectura de Alto Nivel

### 2.1 Diagrama de Arquitectura General

```
┌─────────────────┐    HTTP Request     ┌─────────────────┐
│                 │ ─────────────────→ │                 │
│     Browser     │                     │   Echo Server   │
│                 │ ←───────────────── │                 │
└─────────────────┘    HTML Response    └─────────────────┘
         │                                       │
         │               WebSocket               │
         │ ←─────────────────────────────────→ │
         │                                       │
┌─────────────────┐                     ┌─────────────────┐
│  Client-Side    │                     │  Server-Side    │
│                 │                     │                 │
│  ┌───────────┐  │                     │ ┌─────────────┐ │
│  │ DOM API   │  │                     │ │PageControl  │ │
│  └───────────┘  │                     │ └─────────────┘ │
│  ┌───────────┐  │                     │ ┌─────────────┐ │
│  │WebSocket  │  │                     │ │ComponentMgr │ │
│  │Client     │  │                     │ └─────────────┘ │
│  └───────────┘  │                     │ ┌─────────────┐ │
│  ┌───────────┐  │                     │ │   Layout    │ │
│  │WASM       │  │                     │ │  Manager    │ │
│  │(Optional) │  │                     │ └─────────────┘ │
│  └───────────┘  │                     │ ┌─────────────┐ │
└─────────────────┘                     │ │ Components  │ │
                                        │ │   (Button,  │ │
                                        │ │ Input, etc) │ │
                                        │ └─────────────┘ │
                                        └─────────────────┘
```

### 2.2 Flujo de Comunicación

```
1. Initial Load:
   Browser ──HTTP GET──→ Echo Server
   Browser ←─HTML────── Echo Server

2. WebSocket Handshake:
   Browser ──WS Upgrade──→ Echo Server
   Browser ←─WS Accept──── Echo Server

3. Event Flow:
   DOM Event → Client → WebSocket → Server → Component → Template → WebSocket → Client → DOM Update
```

## 3. Capas Arquitectónicas

### 3.1 Capa de Presentación (Client-Side)

**Responsabilidades**:
- Renderizado HTML/CSS
- Captura de eventos DOM
- Comunicación WebSocket
- Aplicación de actualizaciones DOM

**Componentes**:
- **Browser DOM**: Árbol de elementos HTML
- **WebSocket Client**: `live.js` (referenciado)
- **WASM Module**: `assets/json.wasm` (opcional)

**Ubicación**: `/assets/`, navegador del usuario

### 3.2 Capa de Controlador (Web Layer)

**Responsabilidades**:
- Enrutamiento HTTP
- Manejo de WebSocket connections
- Serialización/deserialización JSON
- Gestión de sesiones por conexión

**Componentes principales**:

#### 3.2.1 PageControl
**Ubicación**: `liveview/page_content.go:14-24`
```go
type PageControl struct {
    Path      string    // Ruta HTTP
    Title     string    // Título de página  
    HeadCode  string    // HTML adicional <head>
    Lang      string    // Idioma del documento
    Css       string    // CSS personalizado
    LiveJs    string    // JavaScript personalizado
    AfterCode string    // Scripts adicionales
    Router    *echo.Echo // Router Echo
    Debug     bool      // Modo debug
}
```

**Responsabilidades**:
- Registro de rutas HTTP y WebSocket
- Configuración de página base
- Manejo del ciclo de vida de conexiones

#### 3.2.2 WebSocket Handler
**Ubicación**: `liveview/page_content.go:78-163`

**Proceso de conexión**:
```
1. WebSocket Upgrade
2. Creación de channels de comunicación  
3. Inicialización de ComponentDriver principal
4. Loop de mensaje bidireccional
5. Cleanup al desconectar
```

### 3.3 Capa de Lógica de Negocio (Business Layer)

#### 3.3.1 Component System

**Interface Component**:
**Ubicación**: `liveview/model.go:19-26`
```go
type Component interface {
    GetTemplate() string  // Template HTML
    Start()               // Inicialización
    GetDriver() LiveDriver // Driver asociado
}
```

**Implementaciones base**:
- **Button**: `components/button.go`
- **InputText**: `components/input.go`  
- **Clock**: `components/clock.go`

#### 3.3.2 ComponentDriver System

**Ubicación**: `liveview/model.go:72-83`
```go
type ComponentDriver[T Component] struct {
    Component         T                           // Componente asociado
    id                string                      // ID único
    IdComponent       string                      // ID del componente
    channel           chan (map[string]interface{}) // Canal salida
    componentsDrivers map[string]LiveDriver        // Drivers hijos
    DriversPage       *map[string]LiveDriver       // Registry global
    channelIn         *map[string]chan interface{} // Canales entrada
    Events            map[string]func(T, interface{}) // Event handlers
    Data              interface{}                  // Datos adicionales
}
```

**Responsabilidades**:
- Proxy entre Component y comunicación web
- Manejo de eventos DOM → Go functions
- Operaciones DOM remotas (SetHTML, GetValue, etc.)
- Gestión de jerarquía de componentes

### 3.4 Capa de Datos (Data Layer)

#### 3.4.1 Estado en Memoria

**Estado Global**:
```go
// liveview/model.go:14-17
var (
    componentsDrivers map[string]LiveDriver = make(map[string]LiveDriver)
    mu                sync.Mutex
)

// liveview/layout.go:28-31  
var (
    MuLayout sync.Mutex         = sync.Mutex{}
    Layaouts map[string]*Layout = make(map[string]*Layout)
)
```

**Problemas identificados**:
- Estado global con sincronización parcial
- Potenciales race conditions
- Acoplamiento fuerte entre componentes

#### 3.4.2 Persistencia

**Implementación básica**: `liveview/utils.go:17-35`
```go
func StringToFile(path string, content string) error
func FileToString(path string) (string, error)  
func Exists(path string) bool
```

**Limitaciones**:
- Solo archivos locales
- Sin base de datos
- Sin transacciones
- Sin cache

## 4. Patrones de Diseño Implementados

### 4.1 Component Pattern

**Descripción**: Encapsulación de UI y lógica en unidades reutilizables
**Implementación**: Interface `Component` + `ComponentDriver`
**Beneficios**: Reutilización, modularidad, separación de responsabilidades

### 4.2 Driver/Proxy Pattern

**Descripción**: `ComponentDriver` actúa como proxy entre componentes Go y DOM
**Implementación**: `ComponentDriver[T]` genérico
**Beneficios**: Abstracción de comunicación WebSocket, operaciones DOM unificadas

### 4.3 Observer Pattern

**Descripción**: Sistema de eventos para notificación de cambios DOM
**Implementación**: Map de event handlers en `ComponentDriver.Events`
**Beneficios**: Desacoplamiento, extensibilidad

### 4.4 Template Method Pattern

**Descripción**: Uso de Go templates con funciones personalizadas
**Implementación**: `text/template` + funciones custom (`mount`, `eqInt`)
**Beneficios**: Flexibilidad en renderizado, reutilización de templates

### 4.5 Factory Pattern

**Descripción**: Creación de componentes y drivers
**Implementación**: Funciones `New[T]()`, `NewDriver[T]()`
**Beneficios**: Construcción consistente, configuración automática

## 5. Comunicación Inter-Capa

### 5.1 Protocolo de Mensajes WebSocket

#### 5.1.1 Cliente → Servidor (Events)
```json
{
  "type": "data",
  "id": "component_id",
  "event": "Click|KeyUp|Change|...",
  "data": <event_payload>
}
```

#### 5.1.2 Servidor → Cliente (DOM Operations)  
```json
{
  "type": "fill|text|style|set|script|remove",
  "id": "element_id", 
  "value": <content>,
  "propertie": <property_name> // Para SetPropertie
}
```

#### 5.1.3 Operaciones de Lectura (Bidireccional)
```json
// Servidor → Cliente (Request)
{
  "type": "get",
  "id": "element_id",
  "sub_type": "value|html|text|style|propertie",
  "value": <property_name>, // Para style/propertie
  "id_ret": <uuid>
}

// Cliente → Servidor (Response)  
{
  "type": "get",
  "id_ret": <uuid>,
  "data": <value>
}
```

### 5.2 Component Mounting

**Proceso de montaje**:
```
1. Component creado con NewDriver()
2. Component añadido al padre con Mount()
3. Template padre usa {{mount "component_id"}}
4. Render genera <span id='mount_span_component_id'></span>
5. Component hijo renderiza dentro del span
```

**Ejemplo**:
```go
parent := components.NewLayout("parent", `
    <div>
        {{mount "child1"}}
        {{mount "child2"}}
    </div>
`)
parent.Mount(child1).Mount(child2)
```

## 6. Escalabilidad y Rendimiento

### 6.1 Puntos Fuertes

**Comunicación eficiente**:
- WebSocket persistente (evita overhead HTTP)
- JSON compacto para mensajes
- Rendering server-side (menos carga cliente)

**Arquitectura reactiva**:
- Solo se envían cambios (diff-like)
- Estado centralizado en servidor
- Componentes reutilizables

### 6.2 Limitaciones de Escalabilidad

**Estado en memoria**:
- No distribuible entre instancias
- Sin persistencia de estado
- Memory leak potenciales

**Conexión 1:1**:
- Una conexión WebSocket por usuario
- No hay balanceador de carga built-in
- Estado perdido al reiniciar servidor

**Concurrencia**:
- Variables globales con mutex
- Potenciales bottlenecks en componentes compartidos

### 6.3 Estrategias de Mejora

#### 6.3.1 Estado Distribuido
```go
type ComponentRegistry interface {
    Get(id string) (LiveDriver, error)
    Set(id string, driver LiveDriver) error
    Delete(id string) error
}

// Implementaciones:
// - MemoryRegistry (actual)
// - RedisRegistry (propuesta)
// - DatabaseRegistry (propuesta)
```

#### 6.3.2 Load Balancing
```go
type SessionStore interface {
    SaveSession(sessionID string, state []byte) error
    LoadSession(sessionID string) ([]byte, error)
}
```

## 7. Integración de Tecnologías

### 7.1 Echo Framework Integration

**Ventajas**:
- Middleware ecosystem maduro
- Routing flexible
- Performance optimizada
- Compatible con net/http standard

**Uso actual**:
```go
e := echo.New()
e.Use(middleware.Logger())
e.Use(middleware.Recover())

// Servir archivos estáticos
e.Static("/assets", "assets")

// Ruta principal  
e.GET("/", pageHandler)

// Ruta WebSocket
e.GET("/ws_goliveview", websocketHandler)
```

### 7.2 WebAssembly Integration

**Arquitectura híbrida**:
```
Opción 1: JavaScript puro (live.js)
Opción 2: WebAssembly (json.wasm) + Go código compartido
```

**Beneficios WASM**:
- Reutilización de código Go en cliente
- Performance superior a JavaScript
- Type safety en cliente

**Compilación WASM**:
```bash
cd cmd/wasm/
GOOS=js GOARCH=wasm go build -o ../../assets/json.wasm
```

### 7.3 Template Engine Integration

**Go Templates + Custom Functions**:
```go
// liveview/fxtemplate.go
var FuncMapTemplate = template.FuncMap{
    "mount": func(id string) template.HTML {
        return template.HTML(fmt.Sprintf(`<span id='mount_span_%s'></span>`, id))
    },
    "eqInt": func(a, b int) bool {
        return a == b
    },
}
```

## 8. Seguridad Arquitectónica

### 8.1 Vulnerabilidades Arquitectónicas

**Sin autenticación/autorización**:
- Cualquier cliente puede conectarse
- Sin verificación de permisos por componente
- Estado global accesible por cualquier conexión

**Validación insuficiente**:
- Sin validación de mensajes WebSocket
- Templates no sanitizados
- Operaciones DOM sin restricciones

### 8.2 Mejoras de Seguridad Propuestas

#### 8.2.1 Authentication Layer
```go
type AuthenticatedPageControl struct {
    *PageControl
    AuthProvider AuthProvider
    RequiredRoles []string
}

type AuthProvider interface {
    Authenticate(token string) (*User, error)
    Authorize(user *User, resource string) bool
}
```

#### 8.2.2 Input Validation Layer
```go
type MessageValidator interface {
    ValidateEvent(msg EventMessage) error
    ValidateData(data interface{}) error
}
```

## 9. Extensibilidad Arquitectónica

### 9.1 Puntos de Extensión

**Custom Components**:
- Implementar interface `Component`
- Registrar con `NewDriver()`
- Añadir custom template functions

**Event Handlers**:
- Métodos Go en componentes
- Callbacks en `Events` map
- Custom event types

**Template Functions**:
- Añadir a `FuncMapTemplate`
- Lógica custom de renderizado

### 9.2 Architectural Hooks

**Lifecycle Events**:
```go
type LifecycleHandler interface {
    OnComponentStart(component Component)
    OnComponentDestroy(component Component)
    OnPageLoad(page PageControl)
    OnPageUnload(page PageControl)
}
```

**Middleware Pattern**:
```go
type ComponentMiddleware func(Component, Event) Event
type PageMiddleware func(PageControl, Request) Request
```

## 10. Comparación Arquitectónica

### 10.1 vs Phoenix LiveView (Elixir)

| Aspecto | Go Echo LiveView | Phoenix LiveView |
|---------|------------------|-------------------|
| **Runtime** | Compiled binary | BEAM VM |
| **Concurrency** | Goroutines | Actor model |
| **State** | Memory + mutex | Process isolation |
| **Fault tolerance** | Panic recovery | Supervisor trees |
| **Hot reloading** | Manual restart | Built-in |
| **Ecosystem** | Limited | Mature |

### 10.2 vs React/SPA

| Aspecto | Go Echo LiveView | React SPA |
|---------|------------------|-----------|
| **State location** | Server | Client |
| **Initial load** | Fast | Bundle size dependent |
| **Offline support** | None | Possible |
| **SEO** | Good | Requires SSR |
| **Development** | Go only | JS + Build tools |
| **Deployment** | Single binary | Static files + API |

## 11. Decisiones Arquitectónicas Clave

### 11.1 Stateful Server vs Stateless

**Decisión**: Stateful server con estado en memoria
**Pros**: Simplicidad, performance, menos latencia
**Cons**: Escalabilidad limitada, estado volátil

### 11.2 WebSocket vs SSE vs HTTP Polling

**Decisión**: WebSocket bidireccional
**Pros**: Bidireccional, baja latencia, eficiente
**Cons**: Complejidad conexión, firewall issues

### 11.3 Generic ComponentDriver vs Interface

**Decisión**: Generic `ComponentDriver[T Component]`
**Pros**: Type safety, performance, menos reflection
**Cons**: Complejidad sintáctica, learning curve

## 12. Roadmap Arquitectónico

### 12.1 Mejoras a Corto Plazo (1-3 meses)

1. **Refactoring estado global** → Dependency injection
2. **Authentication/Authorization** → Middleware layer
3. **Input validation** → Validation layer
4. **Error handling** → Structured error management

### 12.2 Mejoras a Medio Plazo (3-6 meses)

1. **Distributed state** → Redis/Database integration
2. **Load balancing** → Session persistence
3. **Component testing** → Testing framework
4. **Hot reloading** → Development tools

### 12.3 Mejoras a Largo Plazo (6-12 meses)

1. **Microservices** → Component services
2. **GraphQL integration** → Data layer
3. **Mobile support** → Progressive Web App
4. **Offline capabilities** → Service worker integration

## 13. Conclusiones Arquitectónicas

### 13.1 Fortalezas

- **Simplicidad conceptual**: Fácil de entender y usar
- **Performance inicial**: Buena para aplicaciones pequeñas-medianas
- **Developer experience**: Solo Go, sin JavaScript complejo
- **Flexibilidad**: Extensible y personalizable

### 13.2 Debilidades

- **Escalabilidad**: Limitada por estado en memoria
- **Seguridad**: Insuficiente para producción
- **Robustez**: Manejo de errores inconsistente  
- **Testing**: Sin framework de testing

### 13.3 Recomendación Final

La arquitectura es **sólida para un POC** y muestra **gran potencial** para desarrollo de aplicaciones web interactivas en Go. Sin embargo, requiere **mejoras significativas en seguridad, robustez y escalabilidad** antes de ser considerada para uso en producción.

**Siguiente paso recomendado**: Implementar mejoras arquitectónicas a corto plazo, especialmente las relacionadas con seguridad y gestión de estado.