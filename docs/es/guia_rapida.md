# Guía Rápida - Go Echo LiveView

## Introducción

Go Echo LiveView es un framework que implementa el patrón Phoenix LiveView en Go, permitiendo crear aplicaciones web reactivas sin escribir JavaScript. Todo el comportamiento interactivo se maneja desde el servidor a través de WebSocket.

## Instalación Rápida

### Requisitos Previos
- Go 1.20 o superior
- Navegador web moderno con soporte WebSocket

### Pasos de Instalación

```bash
# Clonar el repositorio
git clone https://github.com/arturoeanton/go-echo-live-view.git
cd go-echo-live-view

# Instalar dependencias
go mod tidy

# Compilar módulo WASM (requerido)
cd cmd/wasm/
GOOS=js GOARCH=wasm go build -o ../../assets/json.wasm
cd ../..

# Opcional: Instalar gomon para desarrollo con auto-recarga
go install github.com/c9s/gomon@latest
```

## Tu Primer Componente LiveView

### 1. Estructura Básica de un Componente

```go
package main

import (
    "github.com/arturoeanton/go-echo-live-view/liveview"
)

// Definir la estructura del componente
type MiContador struct {
    liveview.LiveViewComponentWrapper[MiContador]
    Count int
}

// Implementar la interfaz Component
func (c *MiContador) GetTemplate() string {
    return `
    <div>
        <h2>Contador: {{.Count}}</h2>
        <button @click="Incrementar">+1</button>
        <button @click="Decrementar">-1</button>
        <button @click="Reset">Resetear</button>
    </div>
    `
}

// Método de inicialización
func (c *MiContador) Start() {
    c.Count = 0
}

// Manejadores de eventos
func (c *MiContador) Incrementar() {
    c.Count++
    c.Commit() // Actualizar la vista
}

func (c *MiContador) Decrementar() {
    c.Count--
    c.Commit()
}

func (c *MiContador) Reset() {
    c.Count = 0
    c.Commit()
}
```

### 2. Crear la Aplicación Principal

```go
func main() {
    // Crear instancia del componente
    contador := &MiContador{}
    
    // Registrar el componente con LiveView
    driver := liveview.NewDriver("contador", contador)
    
    // Configurar la página
    page := liveview.PageControl{
        Title:    "Mi Contador LiveView",
        HeadCode: "<style>button { margin: 5px; }</style>",
        Lang:     "es",
    }
    
    // Configurar ruta y WebSocket
    page.Register("/", driver)
    
    // Iniciar servidor en puerto 3000
    page.Start(":3000")
}
```

### 3. Ejecutar la Aplicación

```bash
# Ejecutar directamente
go run main.go

# O con auto-recarga (si tienes gomon)
gomon
```

Abre tu navegador en `http://localhost:3000` y verás tu contador interactivo funcionando.

## Conceptos Clave

### Eventos
Los eventos se manejan con la directiva `@click` en el template:
```html
<button @click="NombreMetodo">Click me</button>
```

### Actualización de Vista
Llama a `c.Commit()` después de cambiar el estado para actualizar la vista:
```go
func (c *MiComponente) MiEvento() {
    c.Estado = "nuevo valor"
    c.Commit() // Actualiza el DOM
}
```

### Componentes Anidados
Puedes montar componentes dentro de otros:
```go
// En el método Start() del componente padre
func (c *ComponentePadre) Start() {
    hijo := &ComponenteHijo{}
    c.Mount("hijo-id", hijo)
}
```

En el template:
```html
<div>
    {{mount "hijo-id"}}
</div>
```

## Ejemplos Incluidos

El proyecto incluye varios ejemplos que puedes ejecutar:

```bash
# Reloj en tiempo real
go run example/example1/example1.go

# Input de texto interactivo
go run example/example2/example2.go

# Lista de tareas completa
go run example/example_todo/example_todo.go

# Estilos dinámicos
go run example/example_style/example_style.go

# Tablero Kanban avanzado
go run example/pedidos_board/main.go
```

## Estructura de Proyecto Recomendada

```
mi-app/
├── main.go              # Punto de entrada
├── components/          # Componentes reutilizables
│   ├── header.go
│   ├── footer.go
│   └── ...
├── pages/              # Páginas/vistas principales
│   ├── home.go
│   ├── about.go
│   └── ...
├── assets/             # Archivos estáticos
│   └── json.wasm       # WASM compilado (copiado)
└── templates/          # Templates HTML opcionales
```

## Tips para Desarrollo

### 1. Usar gomon para Auto-recarga
Crea un archivo `gomon.yaml`:
```yaml
watch:
  - "*.go"
  - "**/*.go"
  - "**/*.html"
command: go run main.go
```

### 2. Depuración
- Los errores del servidor aparecen en la consola
- Los errores del cliente aparecen en la consola del navegador
- WebSocket se reconecta automáticamente si se pierde la conexión

### 3. Estilos CSS
Puedes incluir CSS de varias formas:
```go
// En HeadCode
page.HeadCode = "<style>.mi-clase { color: blue; }</style>"

// O cargar desde archivo
page.HeadCode = liveview.GetFileContent("styles.css")
```

### 4. Manejo de Formularios
```go
func (c *MiFormulario) GetTemplate() string {
    return `
    <form>
        <input type="text" @keyup="ActualizarTexto" id="mi-input">
        <p>Texto: {{.Texto}}</p>
    </form>
    `
}

func (c *MiFormulario) ActualizarTexto(data interface{}) {
    c.Texto = c.GetDriver().GetElementValueById("mi-input")
    c.Commit()
}
```

## Solución de Problemas Comunes

### Error: "WebSocket connection failed"
- Verifica que el WASM esté compilado en `assets/json.wasm`
- Asegúrate de que el puerto no esté en uso

### Los cambios no se reflejan
- Recuerda llamar a `c.Commit()` después de modificar el estado
- Verifica que el método esté exportado (primera letra mayúscula)

### Error de compilación WASM
```bash
# Asegúrate de estar en el directorio correcto
cd cmd/wasm/
GOOS=js GOARCH=wasm go build -o ../../assets/json.wasm
```

## Próximos Pasos

1. Explora los ejemplos en el directorio `example/`
2. Lee la documentación completa en el README.md
3. Revisa el código fuente de los componentes en `components/`
4. Experimenta creando tus propios componentes

## Advertencia de Seguridad

⚠️ **IMPORTANTE**: Este es un Proof of Concept (POC) y NO está listo para producción sin una revisión de seguridad exhaustiva. Vulnerabilidades conocidas incluyen:
- Ejecución arbitraria de JavaScript vía `EvalScript()`
- Sin validación de entrada en WebSocket
- Sin sistema de autenticación/autorización
- Posibles vulnerabilidades XSS en templates

Para uso en producción, se requiere implementar medidas de seguridad apropiadas.