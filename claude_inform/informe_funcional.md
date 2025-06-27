# Informe Funcional - Go Echo LiveView

## 1. Resumen de Funcionalidades

Go Echo LiveView es un framework que permite crear aplicaciones web interactivas y reactivas sin necesidad de escribir JavaScript del lado cliente. Implementa el patrón LiveView de Phoenix (Elixir) en Go, proporcionando comunicación en tiempo real mediante WebSockets.

## 2. Funcionalidades Principales

### 2.1 Sistema de Componentes Reactivos

**Descripción**: Permite crear componentes UI que se actualizan automáticamente cuando cambia su estado en el servidor.

**Implementación**: Interface `Component` en `liveview/model.go:19-26`
```go
type Component interface {
    GetTemplate() string  // Template HTML del componente
    Start()               // Inicialización del componente  
    GetDriver() LiveDriver // Acceso al driver de comunicación
}
```

**Funcionalidades específicas**:
- **Renderizado automático**: Los componentes se re-renderizan cuando cambia su estado
- **Templates dinámicos**: Uso de Go templates con datos del componente
- **Ciclo de vida**: Método `Start()` para inicialización personalizada

**Ejemplo de uso**: `components/button.go:5-27`
```go
type Button struct {
    *liveview.ComponentDriver[*Button]
    I       int    // Estado interno
    Caption string // Texto del botón
}
```

### 2.2 Comunicación WebSocket Bidireccional

**Descripción**: Mantiene una conexión persistente entre servidor y cliente para actualizaciones en tiempo real.

**Implementación**: `liveview/page_content.go:78-163`

**Funcionalidades específicas**:
- **Conexión automática**: Se establece automáticamente al cargar la página
- **Protocolo de mensajes**: JSON estructurado para eventos y actualizaciones
- **Reconexión**: Manejo básico de desconexiones

**Tipos de mensajes soportados**:

#### 2.2.1 Cliente → Servidor
```json
{
  "type": "data",
  "id": "component_id",
  "event": "Click", 
  "data": {...}
}
```

#### 2.2.2 Servidor → Cliente  
```json
{
  "type": "fill|text|style|set|script",
  "id": "element_id",
  "value": "content"
}
```

### 2.3 Operaciones DOM Remotas

**Descripción**: Permite manipular el DOM del navegador desde código Go en el servidor.

**Implementación**: Métodos en `ComponentDriver` (`liveview/model.go:256-346`)

#### 2.3.1 Operaciones de Escritura

| Método | Descripción | Implementación |
|--------|-------------|----------------|
| `FillValue(value)` | Actualiza innerHTML | `model.go:272-274` |
| `SetHTML(value)` | Igual que FillValue | `model.go:277-279` |
| `SetText(value)` | Actualiza innerText | `model.go:282-284` |
| `SetValue(value)` | Actualiza value de inputs | `model.go:292-294` |
| `SetStyle(style)` | Modifica CSS del elemento | `model.go:302-304` |
| `SetPropertie(prop, val)` | Modifica propiedades DOM | `model.go:287-289` |
| `EvalScript(code)` | Ejecuta JavaScript | `model.go:296-299` ⚠️ |
| `Remove(id)` | Elimina elemento | `model.go:257-259` |
| `AddNode(id, html)` | Añade nodo HTML | `model.go:262-264` |

#### 2.3.2 Operaciones de Lectura

| Método | Descripción | Implementación |
|--------|-------------|----------------|
| `GetValue()` | Lee value del elemento | `model.go:312-314` |
| `GetHTML()` | Lee innerHTML | `model.go:322-324` |
| `GetText()` | Lee innerText | `model.go:327-329` |
| `GetStyle(prop)` | Lee propiedad CSS | `model.go:317-319` |
| `GetPropertie(prop)` | Lee propiedad DOM | `model.go:332-334` |
| `GetElementById(id)` | Lee value de elemento por ID | `model.go:307-309` |

### 2.4 Sistema de Eventos

**Descripción**: Manejo de eventos DOM que se procesan en el servidor Go.

**Implementación**: Map de eventos en `ComponentDriver.Events` (`model.go:81`)

**Eventos soportados**:
- **Click**: Clics en elementos
- **KeyUp/KeyDown**: Eventos de teclado
- **Change**: Cambios en inputs
- **Focus/Blur**: Enfoque de elementos
- **Submit**: Envío de formularios
- **Eventos personalizados**: Definidos por el desarrollador

**Ejemplo de manejo**: `example/example2/example2.go:62-65`
```go
text1.Events["KeyUp"] = func(data interface{}) {
    text1.FillValue("div_text_result", data.(string))
}
```

### 2.5 Sistema de Montaje de Componentes

**Descripción**: Permite crear jerarquías de componentes donde unos se montan dentro de otros.

**Implementación**: Método `Mount()` en `ComponentDriver` (`model.go:163-169`)

**Funcionalidades específicas**:
- **Montaje automático**: `{{mount "component_id"}}` en templates
- **Jerarquía de componentes**: Componentes padre e hijos
- **Comunicación entre componentes**: Acceso a componentes montados

**Ejemplo**: `example/example2/example2.go:83-84`
```go
return components.NewLayout("home", `
    {{ mount "text1"}}
    <div id="div_text_result"></div>
    {{mount "button1"}}
`).Mount(text1).Mount(button1)
```

### 2.6 Sistema de Layouts

**Descripción**: Componente especial que actúa como contenedor principal de otros componentes.

**Implementación**: `liveview/layout.go:14-23`

**Funcionalidades específicas**:
- **Template base**: Define la estructura HTML principal
- **Gestión de componentes**: Maneja el ciclo de vida de componentes hijos
- **Eventos de ciclo de vida**: `HandlerFirstTime`, `HandlerEventDestroy`

### 2.7 Persistencia de Datos

**Descripción**: Funcionalidades básicas para guardar y cargar datos desde archivos.

**Implementación**: `liveview/utils.go:17-35`

**Funciones disponibles**:
- `StringToFile(path, content)`: Guarda string en archivo
- `FileToString(path)`: Lee archivo como string
- `Exists(path)`: Verifica existencia de archivo

**Ejemplo de uso**: `example/example_todo/example_todo.go:55-77`
```go
func (t *Todo) Save(data interface{}) {
    // Persistir lista de tareas en JSON
    liveview.StringToFile("tasks.json", string(content))
}
```

## 3. Casos de Uso Implementados

### 3.1 Reloj en Tiempo Real (`example1`)

**Archivo**: `example/example1/example1.go`

**Funcionalidades**:
- **Actualización automática**: Cada segundo
- **Formato de fecha/hora**: Personalizable
- **Sin JavaScript cliente**: Solo código Go

**Componentes utilizados**:
- `Clock` component (`components/clock.go`)
- Timer automático con `time.Ticker`

### 3.2 Input Interactivo (`example2`) 

**Archivo**: `example/example2/example2.go`

**Funcionalidades**:
- **Escritura en tiempo real**: Actualiza mientras se escribe
- **Múltiples componentes**: Button + InputText
- **Contador de clics**: Estado persistente del botón

**Flujo funcional**:
1. Usuario escribe en input → `KeyUp` event → Actualiza div resultado
2. Usuario hace clic → `Click` event → Incrementa contador + Muestra texto input

### 3.3 Lista de Tareas CRUD (`example_todo`)

**Archivo**: `example/example_todo/example_todo.go`

**Funcionalidades completas**:
- **Crear tareas**: Formulario de nueva tarea
- **Listar tareas**: Visualización de todas las tareas
- **Marcar completadas**: Toggle de estado
- **Eliminar tareas**: Borrado individual
- **Persistencia**: Guardado automático en `tasks.json`
- **Filtros**: Mostrar todas/pendientes/completadas

**Estructura de datos**:
```go
type Task struct {
    ID       string `json:"id"`
    Text     string `json:"text"`
    Complete bool   `json:"complete"`
}
```

### 3.4 Tablero de Pedidos (`pedidos_board`)

**Archivo**: `example/pedidos_board/pedidos_board.go`

**Funcionalidades avanzadas**:
- **Múltiples estados**: Pendiente, En Proceso, Completado
- **Navegación por tabs**: Interfaz con pestañas
- **Vista dinámica**: Cambia contenido según tab activo
- **Estado complejo**: Manejo de múltiples entidades

## 4. Integración con WebAssembly

### 4.1 Módulo WASM Opcional

**Archivo**: `cmd/wasm/main.go`

**Funcionalidades**:
- **Procesamiento JSON**: Alternativa al cliente JavaScript
- **Operaciones DOM**: Desde WebAssembly
- **Comunicación**: Con servidor Go via WebSocket

**Compilación**: 
```bash
cd cmd/wasm/
GOOS=js GOARCH=wasm go build -o ../../assets/json.wasm
```

### 4.2 Cliente JavaScript

**Archivos**: 
- `assets/wasm_exec.js`: Loader de WebAssembly
- `live.js` (referenciado pero no presente): Cliente WebSocket

**Funcionalidades del cliente**:
- **Conexión WebSocket**: Automática al cargar página
- **Manejo de eventos**: Captura eventos DOM y los envía al servidor
- **Aplicación de cambios**: Recibe comandos del servidor y modifica DOM

## 5. Herramientas de Desarrollo

### 5.1 Auto-reload con gomon

**Configuración**: `gomon.yaml`

**Funcionalidades**:
- **Vigilancia de archivos**: `.go` y `.html`
- **Compilación automática**: WASM + aplicación
- **Reinicio automático**: Cuando cambian los archivos
- **Logging**: Salida de debug opcional

### 5.2 Script de Build

**Archivo**: `build_and_run.sh`

**Proceso automatizado**:
1. Compila módulo WebAssembly
2. Ejecuta ejemplo de demostración (`example2`)

## 6. API y Extensibilidad

### 6.1 Crear Componentes Personalizados

**Pasos requeridos**:
1. **Implementar interface Component**
2. **Definir template HTML**
3. **Manejar eventos si es necesario**
4. **Registrar con NewDriver()**

**Ejemplo**:
```go
type MiComponente struct {
    *liveview.ComponentDriver[*MiComponente]
    Estado string
}

func (c *MiComponente) GetTemplate() string {
    return `<div id="{{.IdComponent}}">{{.Estado}}</div>`
}

func (c *MiComponente) Start() {
    c.Commit() // Renderizar inicial
}

func (c *MiComponente) GetDriver() liveview.LiveDriver {
    return c
}
```

### 6.2 Funciones de Template Personalizadas

**Implementación**: `liveview/fxtemplate.go`

**Funciones disponibles**:
- `mount "id"`: Monta componente hijo
- `eqInt a b`: Compara dos enteros
- **Extensible**: Fácil añadir nuevas funciones

### 6.3 Configuración de Páginas

**Opciones disponibles** (`PageControl`):
- `Title`: Título de la página HTML
- `Lang`: Idioma del documento
- `Path`: Ruta URL de la página
- `HeadCode`: HTML adicional en `<head>` (desde archivo)
- `Css`: CSS personalizado
- `AfterCode`: JavaScript adicional (desde archivo)
- `Debug`: Modo debug con logging de mensajes WebSocket

## 7. Limitaciones Funcionales Identificadas

### 7.1 Limitaciones Actuales

1. **Sin autenticación/autorización**: Cualquier cliente puede acceder
2. **Sin paginación**: Para listas grandes de datos
3. **Sin validación de formularios**: Validación manual requerida
4. **Sin manejo de archivos**: Upload/download no implementado
5. **Sin i18n**: Internacionalización no soportada
6. **Sin testing framework**: Para testing de componentes

### 7.2 Dependencias Externas

**Limitado a**:
- Echo v4 para HTTP/routing
- Gorilla WebSocket para comunicación
- Go templates para renderizado
- Dependencias mínimas (ventaja para deployment)

## 8. Performance y Escalabilidad

### 8.1 Puntos Fuertes

- **Comunicación eficiente**: WebSocket persistente
- **Renderizado servidor**: Menos carga en cliente
- **Estado centralizado**: En servidor Go
- **Componentes reutilizables**: Arquitectura modular

### 8.2 Limitaciones de Escalabilidad

- **Estado en memoria**: Sin persistencia distribuida
- **Una conexión por usuario**: Sin balanceador de carga built-in
- **Sin cache**: De templates o componentes
- **Coupling fuerte**: Cliente-servidor vía WebSocket

## 9. Casos de Uso Recomendados

### 9.1 Ideal Para:

- **Dashboards interactivos**: Con datos en tiempo real
- **Formularios complejos**: Con validación dinámica
- **Aplicaciones CRUD**: Simples a moderadas
- **Prototipos rápidos**: Sin necesidad de JavaScript
- **Apps internas**: Donde la seguridad es controlable

### 9.2 No Recomendado Para:

- **Aplicaciones públicas**: Sin medidas de seguridad adicionales
- **Apps offline**: Requieren conexión constante
- **Juegos**: Latencia crítica
- **Apps móviles**: Mejor usar REST API
- **SEO crítico**: Renderizado dinámico puede afectar indexación

## 10. Conclusiones Funcionales

Go Echo LiveView ofrece un **conjunto sólido de funcionalidades** para crear aplicaciones web interactivas con un enfoque **server-first**. La arquitectura es **intuitiva para desarrolladores Go** y permite **desarrollo rápido** sin conocimientos profundos de JavaScript.

**Fortalezas principales**:
- API simple y consistente
- Ejemplos funcionales completos
- Arquitectura extensible
- Integración natural con el ecosistema Go

**Áreas de mejora funcional**:
- Sistema de validación built-in
- Manejo de archivos y uploads
- Framework de testing para componentes
- Optimizaciones de performance
- Funcionalidades de escalabilidad

El framework muestra **gran potencial** para casos de uso específicos, pero requiere **desarrollo adicional** para ser considerado una solución completa para aplicaciones web modernas.