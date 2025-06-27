# Go Echo LiveView

**Una implementación de Phoenix LiveView en Go usando Echo Framework**

Go Echo LiveView es una biblioteca que permite crear aplicaciones web interactivas y reactivas sin escribir JavaScript del lado cliente. Inspirado en Phoenix LiveView de Elixir, este proyecto utiliza WebSockets para mantener una conexión persistente entre el servidor y el navegador, permitiendo actualizaciones del DOM en tiempo real.

## 🚀 Características Principales

- **Interactividad sin JavaScript**: Escribe toda la lógica en Go, las actualizaciones del DOM se manejan automáticamente
- **Comunicación en Tiempo Real**: WebSockets para actualizaciones bidireccionales instantáneas
- **Sistema de Componentes**: Arquitectura modular con componentes reutilizables
- **Plantillas Dinámicas**: Sistema de templates integrado con Go templates
- **Integración con WASM**: Soporte opcional para WebAssembly para funcionalidades avanzadas

## 📋 Requisitos

- **Go 1.20+**
- **Navegador web moderno** con soporte para WebSockets
- **gomon** (opcional, para desarrollo con auto-reload)

## 🛠️ Instalación

### 1. Clonar el repositorio
```bash
git clone https://github.com/arturoeanton/go-echo-live-view.git
cd go-echo-live-view
```

### 2. Instalar dependencias
```bash
go mod tidy
```

### 3. (Opcional) Instalar gomon para desarrollo
```bash
go install github.com/c9s/gomon@latest
```

## 🏃‍♂️ Ejecución Rápida

### Método 1: Script automático
```bash
./build_and_run.sh
```

### Método 2: Ejecutar ejemplos individuales
```bash
# Ejemplo básico de contador
go run example/example1/example1.go

# Ejemplo con input de texto
go run example/example2/example2.go

# Ejemplo de todo list
go run example/example_todo/example_todo.go
```

### Método 3: Desarrollo con auto-reload
```bash
gomon
```

Visita `http://localhost:1323` en tu navegador.

## 📖 Uso Básico

### Ejemplo Simple: Contador con Botón

```go
package main

import (
    "fmt"
    "github.com/arturoeanton/go-echo-live-view/components"
    "github.com/arturoeanton/go-echo-live-view/liveview"
    "github.com/labstack/echo/v4"
    "github.com/labstack/echo/v4/middleware"
)

func main() {
    e := echo.New()
    e.Use(middleware.Logger())
    e.Use(middleware.Recover())

    // Configurar página principal
    home := liveview.PageControl{
        Title:  "Mi App LiveView",
        Lang:   "es",
        Path:   "/",
        Router: e,
    }

    // Registrar lógica de la página
    home.Register(func() *liveview.ComponentDriver {
        // Crear componentes
        button1 := liveview.NewDriver("contador", &components.Button{Caption: "Incrementar"})
        contador := 0

        // Definir evento del botón
        button1.Events["Click"] = func(data interface{}) {
            contador++
            button1.FillValue("resultado", fmt.Sprintf("Contador: %d", contador))
        }

        // Crear layout con template
        return components.NewLayout("home", `
            <div>
                <h1>Contador LiveView</h1>
                {{mount "contador"}}
                <div id="resultado">Contador: 0</div>
            </div>
        `).Mount(button1)
    })

    e.Logger.Fatal(e.Start(":1323"))
}
```

## 🏗️ Arquitectura del Sistema

### Componentes Principales

1. **PageControl**: Maneja las rutas HTTP y WebSocket
2. **ComponentDriver**: Proxy entre componentes Go y el DOM del navegador
3. **Component Interface**: Interface que deben implementar todos los componentes
4. **Live.js**: Cliente JavaScript que maneja la comunicación WebSocket

### Flujo de Comunicación

```
Navegador ←→ WebSocket ←→ Echo Server ←→ ComponentDriver ←→ Component Go
    ↑                                                           ↓
JavaScript Client                                      Go Templates + Lógica
```

## 🧩 Componentes Disponibles

### Componentes Base
- **Button**: Botón interactivo con eventos click
- **InputText**: Campo de texto con eventos de teclado
- **Clock**: Reloj que se actualiza automáticamente

### Crear un Componente Personalizado

```go
type MiComponente struct {
    *liveview.ComponentDriver[*MiComponente]
    Valor string
}

func (c *MiComponente) GetTemplate() string {
    return `<div id="{{.IdComponent}}">{{.Valor}}</div>`
}

func (c *MiComponente) Start() {
    c.Commit() // Renderizar el componente
}

func (c *MiComponente) GetDriver() liveview.LiveDriver {
    return c
}

// Evento personalizado
func (c *MiComponente) Click(data interface{}) {
    c.Valor = "¡Clickeado!"
    c.Commit()
}
```

## 📁 Estructura del Proyecto

```
├── liveview/           # Core del framework
│   ├── model.go        # Sistema de componentes y drivers
│   ├── page_content.go # Manejo de páginas y WebSocket
│   ├── layout.go       # Sistema de layouts
│   └── utils.go        # Utilidades
├── components/         # Componentes reutilizables
│   ├── button.go
│   ├── input.go
│   └── clock.go
├── example/           # Ejemplos de uso
│   ├── example1/      # Contador básico
│   ├── example_todo/  # Lista de tareas
│   └── pedidos_board/ # Tablero de pedidos
├── assets/            # Archivos estáticos
│   ├── json.wasm      # Módulo WebAssembly
│   └── wasm_exec.js   # Ejecutor WASM
└── cmd/wasm/          # Código fuente WASM
```

## 🔧 Desarrollo

### Comandos Útiles

```bash
# Compilar módulo WASM
cd cmd/wasm/
GOOS=js GOARCH=wasm go build -o ../../assets/json.wasm

# Ejecutar con auto-reload (requiere gomon.yaml)
gomon

# Ejecutar ejemplo específico
go run example/[nombre_ejemplo]/[nombre_ejemplo].go
```

### Configuración de gomon

El archivo `gomon.yaml` configura el auto-reload:

```yaml
name: example
include: 
  - ./example
exclude:
  - txt
  - md
commands:
  command: sh ./build_and_run.sh
  terminate: killall example
extensions:
  - go
  - html
log: true
```

## 🤝 Contribuir al Proyecto

### Estilo de Código

1. **Seguir convenciones de Go**: `gofmt`, `golint`, `go vet`
2. **Documentar funciones públicas**: Usar comentarios Go estándar
3. **Manejo de errores**: Siempre manejar errores explícitamente
4. **Naming**: Usar nombres descriptivos en inglés para APIs públicas

### Estructura de Pull Requests

1. **Fork** del repositorio
2. **Crear rama** descriptiva: `feature/nueva-funcionalidad` o `fix/corregir-bug`
3. **Commits atómicos** con mensajes descriptivos
4. **Incluir ejemplos** si se añaden nuevas funcionalidades
5. **Tests**: Añadir tests para nuevas funcionalidades (cuando el framework de testing esté disponible)

### Áreas de Contribución Prioritarias

- **Seguridad**: Mejoras en validación y sanitización
- **Componentes**: Nuevos componentes reutilizables
- **Documentación**: Ejemplos y guías
- **Testing**: Framework de testing para componentes
- **Performance**: Optimizaciones en comunicación WebSocket

## ⚠️ Advertencias de Seguridad

**IMPORTANTE**: Este proyecto es un POC (Proof of Concept) y NO debe usarse en producción sin revisiones de seguridad significativas.

### Vulnerabilidades Conocidas
- Ejecución de JavaScript arbitrario via `EvalScript()`
- Sin validación de entrada en WebSocket
- Sin autenticación/autorización
- Posibles XSS en templates

## 📚 Ejemplos Incluidos

### example1 - Reloj Simple
Reloj que se actualiza cada segundo mostrando la hora actual.

### example2 - Input Interactivo  
Campo de texto que actualiza el contenido en tiempo real mientras escribes.

### example_todo - Lista de Tareas
CRUD completo de tareas con persistencia en archivo JSON.

### pedidos_board - Tablero de Pedidos
Sistema más complejo con múltiples estados y navegación por tabs.

## 🐛 Reportar Bugs

Crea un issue en GitHub incluyendo:
- **Descripción del problema**
- **Pasos para reproducir**
- **Comportamiento esperado vs actual**
- **Versión de Go y sistema operativo**
- **Código mínimo que reproduce el error**

## 📄 Licencia

Ver archivo `LICENSE` para detalles.

## 🙏 Créditos

Proyecto inspirado en [golive](https://github.com/brendonmatos/golive) y en Phoenix LiveView de Elixir.

---

**¿Preguntas?** Abre un issue o revisa los ejemplos en la carpeta `example/`.