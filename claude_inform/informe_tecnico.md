# Informe Técnico - Go Echo LiveView

## 1. Resumen Ejecutivo

Go Echo LiveView es una implementación del patrón Phoenix LiveView en Go, que permite crear aplicaciones web interactivas sin JavaScript del lado cliente. El sistema utiliza WebSockets para mantener una comunicación bidireccional entre el servidor Go y el navegador, permitiendo actualizaciones del DOM en tiempo real.

## 2. Arquitectura General

### 2.1 Patrón Arquitectónico Principal

El proyecto implementa el **patrón LiveView** con las siguientes características:

- **Server-Side Rendering (SSR)** con interactividad
- **Comunicación WebSocket** para actualizaciones en tiempo real
- **Component-Driver Pattern** para abstracción de componentes
- **Template-based Rendering** usando Go templates

### 2.2 Diagrama de Arquitectura

```
┌─────────────────┐    WebSocket    ┌─────────────────┐
│     Browser     │ ←──────────────→ │   Echo Server   │
│                 │                  │                 │
│  ┌───────────┐  │                  │  ┌───────────┐  │
│  │ live.js   │  │                  │  │PageControl│  │
│  │ (Client)  │  │                  │  │           │  │
│  └───────────┘  │                  │  └───────────┘  │
│                 │                  │         │       │
│  ┌───────────┐  │                  │  ┌───────────┐  │
│  │   DOM     │  │                  │  │Component  │  │
│  │ Elements  │  │                  │  │ Driver    │  │
│  └───────────┘  │                  │  └───────────┘  │
└─────────────────┘                  │         │       │
                                     │  ┌───────────┐  │
                                     │  │Component  │  │
                                     │  │Interface  │  │
                                     │  └───────────┘  │
                                     └─────────────────┘
```

## 3. Módulos Principales

### 3.1 Módulo `liveview` (Core)

**Ubicación**: `/liveview/`

#### 3.1.1 `model.go` - Sistema de Componentes
- **ComponentDriver[T]**: Driver genérico para componentes
  - **Líneas clave**: 72-83 (definición struct)
  - **Funcionalidad**: Manejo de eventos, comunicación WebSocket, mounting
- **Component Interface**: Interface que deben implementar los componentes
  - **Líneas clave**: 19-26
  - **Métodos**: `GetTemplate()`, `Start()`, `GetDriver()`

#### 3.1.2 `page_content.go` - Controlador de Páginas
- **PageControl struct**: Configuración de páginas y rutas
  - **Líneas clave**: 14-24
  - **Responsabilidades**: Registro de rutas HTTP/WebSocket, configuración
- **Register method**: Registra lógica de página y WebSocket handler
  - **Líneas clave**: 54-163
  - **Proceso**: Setup de WebSocket, manejo de mensajes, routing de eventos

#### 3.1.3 `layout.go` - Sistema de Layouts
- **Layout struct**: Componente contenedor principal
  - **Líneas clave**: 14-23
  - **Funcionalidad**: Manejo de templates, mounting de componentes
- **Global Layout Management**: Estado global compartido
  - **Líneas clave**: 28-31
  - **Problema identificado**: Variables globales con mutex

#### 3.1.4 Archivos Auxiliares
- **`utils.go`**: Utilidades de archivos (línea 4: import deprecated `io/ioutil`)
- **`recover.go`**: Manejo de panics y recovery
- **`fxtemplate.go`**: Funciones de template personalizadas
- **`bimap.go`**: Mapeo bidireccional para IDs

### 3.2 Módulo `components` (Componentes Base)

#### 3.2.1 `button.go`
- **Button struct**: Componente botón interactivo
  - **Líneas clave**: 5-9
  - **Bug crítico línea 16**: HTML malformado (`<Button>` con `</button>`)

#### 3.2.2 `input.go`
- **InputText struct**: Campo de entrada de texto
  - **Eventos soportados**: KeyUp, Change, Focus, Blur

#### 3.2.3 `clock.go`
- **Clock struct**: Reloj con auto-actualización
  - **Funcionamiento**: Timer que actualiza cada segundo

### 3.3 Sistema de Ejemplos (`/example/`)

#### 3.3.1 Ejemplos Disponibles
- **example1**: Reloj simple (`/example/example1/example1.go`)
- **example2**: Input interactivo (`/example/example2/example2.go`)
- **example_todo**: CRUD de tareas (`/example/example_todo/example_todo.go`)
- **pedidos_board**: Tablero de pedidos (`/example/pedidos_board/pedidos_board.go`)

## 4. Dependencias Clave

### 4.1 Dependencias Externas

```go
// go.mod líneas 5-10
github.com/google/uuid v1.3.0        // Generación de UUIDs únicos
github.com/gorilla/websocket v1.5.0  // Comunicación WebSocket
github.com/labstack/echo/v4 v4.10.2  // Framework web HTTP
golang.org/x/net v0.9.0              // Parsing HTML y utilidades de red
```

### 4.2 Análisis de Dependencias

- **Echo v4**: Framework maduro y estable para HTTP
- **Gorilla WebSocket**: Implementación WebSocket estándar en Go
- **Google UUID**: Generación confiable de identificadores únicos
- **golang.org/x/net**: Librería oficial de Go para operaciones de red

## 5. Patrones de Diseño Implementados

### 5.1 Component Pattern
**Implementación**: Interface `Component` en `liveview/model.go:19-26`
```go
type Component interface {
    GetTemplate() string  // Template HTML del componente
    Start()               // Inicialización del componente
    GetDriver() LiveDriver // Acceso al driver asociado
}
```

### 5.2 Driver Pattern
**Implementación**: `ComponentDriver[T]` en `liveview/model.go:72-83`
- **Propósito**: Abstrae la comunicación entre componentes Go y DOM del navegador
- **Responsabilidades**: Manejo de eventos, operaciones DOM, comunicación WebSocket

### 5.3 Observer Pattern
**Implementación**: Sistema de eventos en `ComponentDriver.Events`
- **Ubicación**: `liveview/model.go:81`
- **Funcionamiento**: Map de eventos con callbacks

### 5.4 Template Method Pattern
**Implementación**: Uso de Go templates con funciones personalizadas
- **Templates personalizados**: `fxtemplate.go`
- **Funciones**: `mount`, `eqInt`, etc.

## 6. Comunicación WebSocket

### 6.1 Protocol de Mensajes

#### 6.1.1 Mensajes Cliente → Servidor
```json
{
  "type": "data",
  "id": "component_id", 
  "event": "Click",
  "data": {...}
}
```

#### 6.1.2 Mensajes Servidor → Cliente
```json
{
  "type": "fill|text|style|set|script",
  "id": "element_id",
  "value": "content"
}
```

### 6.2 Tipos de Operaciones DOM

**Implementación**: `liveview/model.go:256-304`

- **fill/SetHTML**: Actualizar innerHTML
- **text/SetText**: Actualizar innerText  
- **style/SetStyle**: Modificar CSS
- **set/SetValue**: Cambiar value de inputs
- **script/EvalScript**: Ejecutar JavaScript (⚠️ Riesgo de seguridad)

## 7. Sistema de Templates

### 7.1 Funciones Personalizadas

**Ubicación**: `liveview/fxtemplate.go`

- **mount**: Monta componentes hijos
- **eqInt**: Comparación de enteros
- **Extensibilidad**: Fácil agregar nuevas funciones

### 7.2 Template Base

**Ubicación**: `liveview/page_content.go:27-50`

Estructura HTML base con:
- Carga de WebAssembly 
- Conexión WebSocket automática
- Contenedor principal `#content`

## 8. Integración WebAssembly

### 8.1 Módulo WASM

**Ubicación**: `/cmd/wasm/main.go`
- **Propósito**: Alternativa a cliente JavaScript puro
- **Compilación**: `GOOS=js GOARCH=wasm go build`
- **Assets**: `/assets/json.wasm`, `/assets/wasm_exec.js`

## 9. Herramientas de Desarrollo

### 9.1 Auto-reload con gomon

**Configuración**: `gomon.yaml`
```yaml
name: example
include: ["./example"]
extensions: ["go", "html"]
commands:
  command: sh ./build_and_run.sh
```

### 9.2 Scripts de Build

**build_and_run.sh**:
1. Compila módulo WASM
2. Ejecuta ejemplo de demostración

## 10. Análisis de Performance

### 10.1 Puntos Fuertes
- **Comunicación eficiente**: WebSocket mantiene conexión persistente
- **Renderizado servidor**: Menos carga en cliente
- **Componentes reutilizables**: Arquitectura modular

### 10.2 Puntos de Mejora
- **Estado global**: Variables globales con mutex (`liveview/layout.go:28-31`)
- **Reflection overhead**: Uso intensivo de reflection (`liveview/model.go:249-251`)
- **Memory leaks potenciales**: Channels no cerrados explícitamente

## 11. Extensibilidad

### 11.1 Crear Nuevos Componentes
Implementar interface `Component` con:
1. `GetTemplate()`: HTML template
2. `Start()`: Lógica de inicialización  
3. `GetDriver()`: Retorno del driver asociado

### 11.2 Eventos Personalizados
Agregar métodos al componente o usar `Events` map:
```go
button.Events["CustomEvent"] = func(data interface{}) {
    // Lógica del evento
}
```

## 12. Consideraciones de Mantenimiento

### 12.1 Áreas Críticas para Refactoring
1. **Estado global**: `componentsDrivers`, `Layouts` maps globales
2. **Error handling**: Múltiples estrategias de recovery inconsistentes
3. **Validation**: Falta validación de entrada en WebSocket

### 12.2 Mejoras Arquitectónicas Sugeridas
1. **Dependency Injection**: Eliminar estado global
2. **Context propagation**: Para cancelación y timeouts
3. **Structured logging**: Reemplazar fmt.Println por logger apropiado
4. **Testing framework**: Para testing de componentes

## 13. Conclusiones Técnicas

Go Echo LiveView demuestra una implementación sólida del patrón LiveView en Go. La arquitectura es limpia y extensible, pero requiere mejoras significativas en:

1. **Seguridad**: Validación de entrada y eliminación de `EvalScript`
2. **Robustez**: Manejo consistente de errores y eliminación de estado global
3. **Testing**: Framework de testing para componentes
4. **Documentation**: Documentación técnica de APIs

El proyecto tiene potencial como framework de desarrollo rápido de aplicaciones web interactivas en Go, pero necesita maduración antes de considerarse para producción.