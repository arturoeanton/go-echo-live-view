# LiveView JavaScript Client Library

[Español](#cliente-javascript-liveview) | English

## Overview

The LiveView JavaScript Client (`live.js`) is a lightweight, framework-agnostic library that enables real-time, server-driven UI updates through WebSocket connections. It replaces the previous WebAssembly implementation with pure JavaScript, providing better compatibility, easier debugging, and simpler deployment.

## Features

- 🔄 **Automatic WebSocket Management**: Handles connection, reconnection, and error recovery
- 📡 **Bidirectional Communication**: Seamless client-server event exchange
- 🎯 **DOM Manipulation**: Server-controlled UI updates without page reloads
- 🖱️ **Drag & Drop Support**: Built-in draggable elements with server synchronization
- 📦 **Zero Dependencies**: Pure JavaScript, no external libraries required
- 🔌 **Framework Agnostic**: Works with any server implementing the LiveView protocol
- 🐛 **Debug Mode**: Verbose logging with `?verbose=true` URL parameter

## Installation

### Option 1: Include from CDN
```html
<script src="https://your-cdn.com/live.js"></script>
```

### Option 2: Local Installation
```html
<script src="/assets/live.js"></script>
```

### Option 3: Module Import
```javascript
import '/assets/live.js';
```

## Basic Usage

### HTML Setup
```html
<!DOCTYPE html>
<html>
<head>
    <title>LiveView App</title>
    <script src="/assets/live.js"></script>
</head>
<body>
    <div id="content">
        <!-- Server-rendered content appears here -->
    </div>
</body>
</html>
```

### Sending Events to Server
```javascript
// Send a button click event
send_event('button-1', 'Click', null);

// Send form data
send_event('form-1', 'Submit', {
    name: 'John Doe',
    email: 'john@example.com'
});

// Send custom event with data
send_event('component-1', 'CustomEvent', {
    action: 'update',
    value: 42
});
```

### Making Elements Draggable
```html
<!-- Basic draggable element -->
<div class="draggable" 
     id="box-1"
     style="position: absolute; left: 100px; top: 100px;">
    Drag me!
</div>

<!-- Draggable with component association -->
<div class="draggable"
     id="card-1"
     data-element-id="card-1"
     data-component-id="kanban-board"
     style="position: absolute;">
    Draggable Card
</div>

<!-- Temporarily disable dragging -->
<div class="draggable"
     id="locked-box"
     data-drag-disabled="true">
    Currently not draggable
</div>
```

## Server Message Types

The client handles the following message types from the server:

### Content Updates
```json
// Replace HTML content
{"type": "fill", "id": "element-id", "value": "<p>New content</p>"}

// Set text content
{"type": "text", "id": "element-id", "value": "Plain text"}

// Set input value
{"type": "set", "id": "input-id", "value": "New value"}
```

### DOM Manipulation
```json
// Remove element
{"type": "remove", "id": "element-id"}

// Add child node
{"type": "addNode", "id": "parent-id", "value": "<div>New child</div>"}

// Set CSS styles
{"type": "style", "id": "element-id", "value": "color: red; font-size: 16px"}
```

### Property Updates
```json
// Set any property
{"type": "propertie", "id": "element-id", "propertie": "disabled", "value": true}
```

### Script Execution
```json
// Execute JavaScript
{"type": "script", "value": "console.log('Hello from server')"}
```

### Data Retrieval
```json
// Request element value
{"type": "get", "id": "input-id", "sub_type": "value", "id_ret": "req-123"}

// Request HTML content
{"type": "get", "id": "div-id", "sub_type": "html", "id_ret": "req-124"}
```

## Client-to-Server Events

Events sent from client to server follow this format:

```json
{
    "type": "data",
    "id": "component-id",
    "event": "EventName",
    "data": "event data (string or JSON)"
}
```

### Drag Events
When dragging elements, the client automatically sends:

1. **DragStart**: When dragging begins
```json
{
    "type": "data",
    "id": "component-id",
    "event": "DragStart",
    "data": "{\"element\":\"box-1\",\"x\":150,\"y\":200}"
}
```

2. **DragMove**: During dragging (throttled to 60 FPS)
```json
{
    "type": "data",
    "id": "component-id",
    "event": "DragMove",
    "data": "{\"element\":\"box-1\",\"x\":250,\"y\":300}"
}
```

3. **DragEnd**: When dragging completes
```json
{
    "type": "data",
    "id": "component-id",
    "event": "DragEnd",
    "data": "{\"element\":\"box-1\",\"x\":280,\"y\":320}"
}
```

## Debug Mode

Enable verbose logging by adding `?verbose=true` or `?debug=true` to your URL:
```
http://localhost:8080/app?verbose=true
```

This will log:
- Connection status changes
- All incoming/outgoing messages
- Drag operations
- Error details

## API Reference

### Global Functions

#### `send_event(id, event, data)`
Send an event to the server.

- **id** (string): Component identifier
- **event** (string): Event name
- **data** (any): Event data (will be JSON stringified if object)

#### `connect()`
Manually establish or re-establish WebSocket connection.

### Global Variables

#### `window.ws`
Direct access to the WebSocket instance for debugging.

```javascript
// Check connection status
console.log(window.ws.readyState);
// 0 = CONNECTING, 1 = OPEN, 2 = CLOSING, 3 = CLOSED

// Close connection manually
window.ws.close();
```

#### `window.dragState`
Current drag operation state (for debugging).

```javascript
console.log(window.dragState);
// {
//   isDragging: false,
//   draggedElement: '',
//   componentId: '',
//   startX: 0,
//   startY: 0,
//   initX: 0,
//   initY: 0,
//   lastUpdate: 0
// }
```

## Browser Compatibility

- Chrome 50+
- Firefox 45+
- Safari 10+
- Edge 14+
- Opera 37+
- Mobile browsers (iOS Safari 10+, Chrome Mobile)

## Performance Considerations

1. **Message Throttling**: Drag events are throttled to 60 FPS
2. **Reconnection**: Automatic reconnection every 1 second when disconnected
3. **Large Content**: Fill operations with large HTML are logged by size, not content
4. **Event Queuing**: Events before connection are queued and sent when ready

## Security Notes

- Always use WSS (WebSocket Secure) in production
- The server should sanitize all user input before sending commands
- Script execution uses `eval()` - ensure server is trusted
- Validate all incoming messages on the server side

## License

MIT License - See LICENSE file for details

---

# Cliente JavaScript LiveView

Español | [English](#liveview-javascript-client-library)

## Descripción General

El Cliente JavaScript de LiveView (`live.js`) es una biblioteca ligera e independiente del framework que permite actualizaciones de UI en tiempo real controladas por el servidor a través de conexiones WebSocket. Reemplaza la implementación anterior de WebAssembly con JavaScript puro, proporcionando mejor compatibilidad, depuración más fácil y despliegue más simple.

## Características

- 🔄 **Gestión Automática de WebSocket**: Maneja conexión, reconexión y recuperación de errores
- 📡 **Comunicación Bidireccional**: Intercambio fluido de eventos cliente-servidor
- 🎯 **Manipulación del DOM**: Actualizaciones de UI controladas por el servidor sin recargar la página
- 🖱️ **Soporte de Arrastrar y Soltar**: Elementos arrastrables integrados con sincronización del servidor
- 📦 **Sin Dependencias**: JavaScript puro, no requiere bibliotecas externas
- 🔌 **Agnóstico al Framework**: Funciona con cualquier servidor que implemente el protocolo LiveView
- 🐛 **Modo de Depuración**: Registro detallado con el parámetro URL `?verbose=true`

## Instalación

### Opción 1: Incluir desde CDN
```html
<script src="https://tu-cdn.com/live.js"></script>
```

### Opción 2: Instalación Local
```html
<script src="/assets/live.js"></script>
```

### Opción 3: Importación de Módulo
```javascript
import '/assets/live.js';
```

## Uso Básico

### Configuración HTML
```html
<!DOCTYPE html>
<html>
<head>
    <title>Aplicación LiveView</title>
    <script src="/assets/live.js"></script>
</head>
<body>
    <div id="content">
        <!-- El contenido renderizado por el servidor aparece aquí -->
    </div>
</body>
</html>
```

### Envío de Eventos al Servidor
```javascript
// Enviar evento de clic de botón
send_event('button-1', 'Click', null);

// Enviar datos de formulario
send_event('form-1', 'Submit', {
    nombre: 'Juan Pérez',
    email: 'juan@ejemplo.com'
});

// Enviar evento personalizado con datos
send_event('component-1', 'EventoPersonalizado', {
    accion: 'actualizar',
    valor: 42
});
```

### Hacer Elementos Arrastrables
```html
<!-- Elemento arrastrable básico -->
<div class="draggable" 
     id="caja-1"
     style="position: absolute; left: 100px; top: 100px;">
    ¡Arrástrame!
</div>

<!-- Arrastrable con asociación de componente -->
<div class="draggable"
     id="tarjeta-1"
     data-element-id="tarjeta-1"
     data-component-id="tablero-kanban"
     style="position: absolute;">
    Tarjeta Arrastrable
</div>

<!-- Deshabilitar arrastre temporalmente -->
<div class="draggable"
     id="caja-bloqueada"
     data-drag-disabled="true">
    Actualmente no arrastrable
</div>
```

## Tipos de Mensajes del Servidor

El cliente maneja los siguientes tipos de mensajes del servidor:

### Actualizaciones de Contenido
```json
// Reemplazar contenido HTML
{"type": "fill", "id": "id-elemento", "value": "<p>Nuevo contenido</p>"}

// Establecer contenido de texto
{"type": "text", "id": "id-elemento", "value": "Texto plano"}

// Establecer valor de input
{"type": "set", "id": "id-input", "value": "Nuevo valor"}
```

### Manipulación del DOM
```json
// Eliminar elemento
{"type": "remove", "id": "id-elemento"}

// Agregar nodo hijo
{"type": "addNode", "id": "id-padre", "value": "<div>Nuevo hijo</div>"}

// Establecer estilos CSS
{"type": "style", "id": "id-elemento", "value": "color: red; font-size: 16px"}
```

### Actualizaciones de Propiedades
```json
// Establecer cualquier propiedad
{"type": "propertie", "id": "id-elemento", "propertie": "disabled", "value": true}
```

### Ejecución de Script
```json
// Ejecutar JavaScript
{"type": "script", "value": "console.log('Hola desde el servidor')"}
```

### Recuperación de Datos
```json
// Solicitar valor del elemento
{"type": "get", "id": "id-input", "sub_type": "value", "id_ret": "req-123"}

// Solicitar contenido HTML
{"type": "get", "id": "id-div", "sub_type": "html", "id_ret": "req-124"}
```

## Eventos Cliente-Servidor

Los eventos enviados del cliente al servidor siguen este formato:

```json
{
    "type": "data",
    "id": "id-componente",
    "event": "NombreEvento",
    "data": "datos del evento (cadena o JSON)"
}
```

### Eventos de Arrastre
Al arrastrar elementos, el cliente envía automáticamente:

1. **DragStart**: Cuando comienza el arrastre
```json
{
    "type": "data",
    "id": "id-componente",
    "event": "DragStart",
    "data": "{\"element\":\"caja-1\",\"x\":150,\"y\":200}"
}
```

2. **DragMove**: Durante el arrastre (limitado a 60 FPS)
```json
{
    "type": "data",
    "id": "id-componente",
    "event": "DragMove",
    "data": "{\"element\":\"caja-1\",\"x\":250,\"y\":300}"
}
```

3. **DragEnd**: Cuando termina el arrastre
```json
{
    "type": "data",
    "id": "id-componente",
    "event": "DragEnd",
    "data": "{\"element\":\"caja-1\",\"x\":280,\"y\":320}"
}
```

## Modo de Depuración

Habilita el registro detallado agregando `?verbose=true` o `?debug=true` a tu URL:
```
http://localhost:8080/app?verbose=true
```

Esto registrará:
- Cambios de estado de conexión
- Todos los mensajes entrantes/salientes
- Operaciones de arrastre
- Detalles de errores

## Referencia de API

### Funciones Globales

#### `send_event(id, evento, datos)`
Envía un evento al servidor.

- **id** (string): Identificador del componente
- **evento** (string): Nombre del evento
- **datos** (any): Datos del evento (se convertirán a JSON si es objeto)

#### `connect()`
Establecer o reestablecer manualmente la conexión WebSocket.

### Variables Globales

#### `window.ws`
Acceso directo a la instancia WebSocket para depuración.

```javascript
// Verificar estado de conexión
console.log(window.ws.readyState);
// 0 = CONECTANDO, 1 = ABIERTO, 2 = CERRANDO, 3 = CERRADO

// Cerrar conexión manualmente
window.ws.close();
```

#### `window.dragState`
Estado actual de la operación de arrastre (para depuración).

```javascript
console.log(window.dragState);
// {
//   isDragging: false,
//   draggedElement: '',
//   componentId: '',
//   startX: 0,
//   startY: 0,
//   initX: 0,
//   initY: 0,
//   lastUpdate: 0
// }
```

## Compatibilidad con Navegadores

- Chrome 50+
- Firefox 45+
- Safari 10+
- Edge 14+
- Opera 37+
- Navegadores móviles (iOS Safari 10+, Chrome Mobile)

## Consideraciones de Rendimiento

1. **Limitación de Mensajes**: Los eventos de arrastre están limitados a 60 FPS
2. **Reconexión**: Reconexión automática cada 1 segundo cuando está desconectado
3. **Contenido Grande**: Las operaciones de relleno con HTML grande se registran por tamaño, no contenido
4. **Cola de Eventos**: Los eventos antes de la conexión se ponen en cola y se envían cuando está listo

## Notas de Seguridad

- Siempre use WSS (WebSocket Seguro) en producción
- El servidor debe sanear toda entrada del usuario antes de enviar comandos
- La ejecución de scripts usa `eval()` - asegúrese de que el servidor sea confiable
- Valide todos los mensajes entrantes en el lado del servidor

## Licencia

Licencia MIT - Ver archivo LICENSE para más detalles