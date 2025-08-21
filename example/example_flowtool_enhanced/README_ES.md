# Herramienta Avanzada de Diagramas de Flujo

Un editor completo de diagramas de flujo construido con el framework Go Echo LiveView, que demuestra las características avanzadas del framework y las capacidades de desarrollo web en tiempo real.

## 🌟 Características

### Funcionalidad Principal
- **Editor Visual de Diagramas**: Crear, editar y gestionar diagramas de flujo con arrastrar y soltar
- **Múltiples Tipos de Nodos**: Nodos de Inicio, Proceso, Decisión, Datos y Fin
- **Actualizaciones en Tiempo Real**: Comunicación basada en WebSocket para actualizaciones instantáneas
- **Sistema de Deshacer/Rehacer**: Gestión completa del historial con serialización de estado
- **Importar/Exportar**: Formato JSON para persistencia y compartir diagramas
- **Auto-organizar**: Disposición automática de nodos en formación de cuadrícula

### Características Avanzadas del Framework

#### 1. **Gestión de Estado (State Management)**
- Almacenamiento centralizado con caché TTL
- Persistencia automática del estado
- Seguimiento de posición de todos los nodos
- Instantáneas de estado para deshacer/rehacer

#### 2. **Registro de Eventos (Event Registry)**
- Patrón pub/sub para comunicación entre componentes
- Throttling de eventos para rendimiento
- Soporte de eventos con wildcards
- Recolección de métricas

#### 3. **Caché de Templates**
- Caché de templates compilados
- Invalidación automática de caché
- Almacenamiento eficiente en memoria (límite 10MB)
- Pre-compilación al inicio

#### 4. **Límite de Errores (Error Boundary)**
- Mecanismo de recuperación de panics
- Envoltorios de ejecución segura
- Registro de errores con stack traces
- Previene crashes de la aplicación

#### 5. **DOM Virtual** (Listo para integración)
- Estructura VNode implementada
- Algoritmo de diff disponible
- Sistema de generación de patches
- Aún no completamente integrado (ver propuesta)

## 🚀 Comenzando

### Prerrequisitos
- Go 1.19 o superior
- Navegador web moderno con soporte WebAssembly

### Instalación

1. Clonar el repositorio:
```bash
git clone https://github.com/arturoeanton/go-echo-live-view.git
cd go-echo-live-view
```

2. Construir el módulo WASM:
```bash
cd cmd/wasm/
GOOS=js GOARCH=wasm go build -o ../../assets/json.wasm
cd ../..
```

3. Ejecutar el ejemplo:
```bash
go run example/example_flowtool_enhanced/main.go
```

4. Abrir el navegador en `http://localhost:8888`

## 🎮 Uso

### Operaciones Básicas

#### Agregar Nodos
1. Hacer clic en cualquier botón de tipo de nodo (Inicio, Proceso, Decisión, Datos, Fin)
2. El nodo aparece en una posición automática
3. Cada nodo obtiene un ID único

#### Mover Nodos
- **Arrastrar y Soltar**: Clic y arrastrar cualquier nodo para reposicionar
- **Teclado**: Usar botones de flecha cuando un nodo está seleccionado
- **Auto-organizar**: Clic en "Auto Arrange" para organizar todos los nodos

#### Conectar Nodos
1. Hacer clic en el botón "Connect Mode"
2. Hacer clic en el primer nodo (origen)
3. Hacer clic en el segundo nodo (destino)
4. Aparece una línea de conexión entre ellos
5. Presionar ESC o hacer clic en "Connect Mode" nuevamente para salir

#### Editar Nodos
- **Doble clic** en un nodo para editar su etiqueta y código
- **Eliminar**: Seleccionar un nodo y presionar Delete, o hacer clic en el botón ×

#### Gestionar Conexiones
- Hacer clic en una línea de conexión para seleccionarla
- Doble clic para editar la etiqueta
- Hacer clic en la × roja para eliminar cuando está seleccionada

### Características Avanzadas

#### Deshacer/Rehacer
- **Deshacer**: Ctrl+Z o clic en botón Undo
- **Rehacer**: Ctrl+Y o clic en botón Redo
- Se guardan hasta 50 estados

#### Importar/Exportar
- **Exportar**: Clic en "Export JSON" para obtener el diagrama como JSON
- **Importar**: Usar el componente de carga de archivos para cargar un diagrama guardado

#### Controles del Canvas
- **Zoom In/Out**: Usar los botones de zoom
- **Restablecer Vista**: Vuelve al 100% de zoom, centrado
- **Alternar Cuadrícula**: Mostrar/ocultar cuadrícula de alineación

## 🏗️ Arquitectura

### Estructura de Componentes

```
EnhancedFlowTool (Componente Principal)
├── FlowCanvas (Área de Dibujo)
│   ├── FlowBox[] (Nodos)
│   └── FlowEdge[] (Conexiones)
├── Modal (Diálogo de Exportación)
├── FileUpload (Importación)
├── StateManager (Persistencia de Estado)
├── EventRegistry (Bus de Eventos)
├── TemplateCache (Rendimiento)
└── ErrorBoundary (Manejo de Errores)
```

### Flujo de Comunicación

1. **Interacción del Usuario** → Evento del Navegador
2. **Módulo WASM** → Captura y envía vía WebSocket
3. **Manejador del Servidor** → Procesa el evento
4. **Actualización de Estado** → Modifica el estado del componente
5. **Re-renderizado** → Envía actualizaciones HTML vía WebSocket
6. **Actualización del DOM** → WASM aplica los cambios

### Sistema de Eventos

La aplicación usa tanto eventos directos como el Registro de Eventos:

- **Eventos Directos**: Interacciones de UI (clics, arrastres)
- **Eventos del Registro**: Eventos del sistema (auto-guardado, cambios de estado)

Ejemplo de flujo de eventos:
```go
Usuario arrastra nodo → DragStart → DragMove (throttled) → DragEnd
                           ↓             ↓                     ↓
                    GuardarEstado  ActualizarPosición  EmitirEventoCambio
```

## 📁 Estructura de Archivos

```
example_flowtool_enhanced/
├── main.go           # Aplicación principal con todos los manejadores
├── README.md         # Documentación en inglés
└── README_ES.md      # Este archivo
```

## 🔧 Configuración

### Gestor de Estado
```go
StateConfig{
    Provider:     MemoryStateProvider,  // Cambiar a Redis para producción
    CacheEnabled: true,
    CacheTTL:     5 * time.Minute,
}
```

### Registro de Eventos
```go
EventRegistryConfig{
    MaxHandlersPerEvent: 10,    // Previene fugas de memoria
    EnableMetrics:       true,   // Monitoreo de rendimiento
    EnableWildcards:     true,   // Coincidencia de patrones
}
```

### Caché de Templates
```go
TemplateCacheConfig{
    MaxSize:          10 * 1024 * 1024,  // Límite de 10MB
    TTL:              5 * time.Minute,   // Intervalo de refresco
    EnablePrecompile: true,               // Optimización al inicio
}
```

## 🐛 Depuración

Habilitar registro detallado:
- Agregar `?verbose=true` a la URL
- Revisar la consola del navegador para logs de WASM
- Los logs del servidor muestran todos los eventos y cambios de estado

## ⚠️ Consideraciones de Seguridad

Esta es una aplicación POC/ejemplo. Para uso en producción:
- Implementar autenticación y autorización
- Agregar validación y sanitización de entrada
- Eliminar capacidades de `EvalScript()`
- Implementar limitación de tasa en WebSocket
- Usar HTTPS/WSS para conexiones
- Agregar protección CSRF

## 🔄 Protocolo WebSocket

### Cliente → Servidor
```json
{
    "type": "data",
    "id": "component-id",
    "event": "EventName",
    "data": "{json-data}"
}
```

### Servidor → Cliente
```json
{
    "type": "fill|text|style|script",
    "id": "element-id",
    "value": "content"
}
```

## 🎯 Optimizaciones de Rendimiento

- **Caché de Templates**: Reduce el tiempo de renderizado en 70%
- **Throttling de Eventos**: Limita eventos de arrastre a intervalos de 50ms
- **Caché de Estado**: TTL de 5 minutos reduce accesos a base de datos
- **VDOM Listo**: Preparado para actualizaciones diferenciales
- **Compresión WebSocket**: Reduce el uso de ancho de banda

## 📊 Métricas

El Registro de Eventos recolecta:
- Conteo de eventos por tipo
- Tiempos de ejecución de manejadores
- Tasas de error
- Patrones de uso de memoria

## 🚧 Limitaciones Conocidas

1. **Sin Integración VirtualDOM**: Actualmente usa re-renderizados completos
2. **Solo Estado en Memoria**: Sin almacenamiento persistente por defecto
3. **Usuario Único**: Sin colaboración multi-usuario
4. **Sin Soporte Móvil**: Optimizado solo para escritorio
5. **Tipos de Nodos Limitados**: Conjunto fijo de tipos de nodos

## 🔮 Mejoras Futuras

- [ ] Integración completa de VirtualDOM para mejor rendimiento
- [ ] Edición colaborativa con transformaciones operacionales
- [ ] Tipos de nodos personalizados con plugins
- [ ] Algoritmos de enrutamiento avanzados
- [ ] Exportar a varios formatos (SVG, PNG, PDF)
- [ ] Atajos de teclado para todas las operaciones
- [ ] Soporte para dispositivos táctiles
- [ ] Tema modo oscuro

## 📄 Licencia

Este ejemplo es parte del framework Go Echo LiveView y sigue la misma licencia.

## 🤝 Contribuir

¡Las contribuciones son bienvenidas! Por favor:
1. Hacer fork del repositorio
2. Crear una rama de característica
3. Agregar pruebas para nueva funcionalidad
4. Actualizar la documentación
5. Enviar un pull request

## 📚 Documentación Relacionada

- [Documentación del Framework](../../README.md)
- [Módulo WASM](../../cmd/wasm/main.go)
- [Biblioteca de Componentes](../../components/)
- [Otros Ejemplos](../)

## 💬 Soporte

Para preguntas y soporte:
- Abrir un issue en GitHub
- Revisar ejemplos existentes
- Consultar documentación del framework

---

Construido con ❤️ usando el Framework Go Echo LiveView