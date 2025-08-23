# RFC-LIVE: Especificación del Protocolo LiveView

**Versión:** 1.0.0  
**Estado:** Borrador  
**Fecha:** 2025-08-22  
**Autores:** Contribuidores de Go Echo LiveView

## Resumen

Este documento especifica el Protocolo LiveView, un protocolo de comunicación basado en WebSocket que permite actualizaciones de interfaz de usuario en tiempo real controladas por el servidor sin requerir lógica de aplicación del lado del cliente. El protocolo permite a los servidores controlar la manipulación del DOM, manejar eventos y mantener el estado de la aplicación mientras que los clientes actúan como capas de presentación ligeras.

## 1. Introducción

### 1.1 Propósito

El Protocolo LiveView permite a los frameworks del lado del servidor crear aplicaciones web interactivas donde el servidor mantiene control completo sobre el estado y la lógica de la UI. Este enfoque simplifica el desarrollo eliminando la necesidad de gestión de estado del lado del cliente y diseño de API.

### 1.2 Objetivos de Diseño

1. **Simplicidad**: Complejidad mínima del lado del cliente
2. **Tiempo real**: Comunicación bidireccional de baja latencia
3. **Confiabilidad**: Reconexión automática y recuperación de errores
4. **Compatibilidad**: Funciona con tecnologías web estándar
5. **Flexibilidad**: Agnóstico al framework y lenguaje

### 1.3 Terminología

- **Servidor**: La aplicación backend que implementa la lógica de negocio
- **Cliente**: La biblioteca JavaScript ejecutándose en el navegador
- **Componente**: Un elemento de UI del lado del servidor con estado y comportamiento
- **Evento**: Una interacción del usuario o ocurrencia del sistema
- **Mensaje**: Un paquete de comunicación codificado en JSON

## 2. Capa de Transporte

### 2.1 Conexión WebSocket

El protocolo usa WebSocket (RFC 6455) como capa de transporte.

#### 2.1.1 Patrón de URL de Conexión

```
ws://[host]:[puerto][ruta]ws_goliveview
wss://[host]:[puerto][ruta]ws_goliveview
```

- Usar `ws://` para orígenes HTTP
- Usar `wss://` para orígenes HTTPS (requerido en producción)
- El sufijo `ws_goliveview` identifica los endpoints de LiveView

#### 2.1.2 Ciclo de Vida de la Conexión

1. **Solicitud HTTP Inicial**: El cliente solicita la página vía HTTP
2. **Respuesta HTML**: El servidor devuelve HTML con el script cliente de LiveView
3. **Actualización a WebSocket**: El cliente establece conexión WebSocket
4. **Inicialización del Componente**: El servidor envía el estado inicial de la UI
5. **Sesión Interactiva**: Intercambio bidireccional de eventos/actualizaciones
6. **Manejo de Desconexión**: Reconexión automática en caso de falla

### 2.2 Estrategia de Reconexión

Los clientes DEBEN implementar reconexión automática:

```javascript
// Intervalo de reconexión: 1000ms (1 segundo)
// Intentos máximos: Ilimitado
// Estrategia de retroceso: Ninguna (intervalo constante)
```

## 3. Formato de Mensajes

Todos los mensajes DEBEN ser objetos JSON válidos codificados como cadenas UTF-8.

### 3.1 Mensajes Cliente-a-Servidor

#### 3.1.1 Mensaje de Evento

Usado para enviar interacciones del usuario al servidor.

```json
{
    "type": "data",
    "id": "identificador-componente",
    "event": "NombreEvento",
    "data": "datos-específicos-evento"
}
```

**Campos:**
- `type` (string, requerido): DEBE ser "data"
- `id` (string, requerido): Identificador del componente
- `event` (string, requerido): Nombre del evento (ej., "Click", "Input")
- `data` (string, opcional): Datos del evento como cadena o cadena JSON

#### 3.1.2 Mensaje de Respuesta Get

Respuesta a la solicitud de datos del servidor.

```json
{
    "type": "get",
    "id_ret": "identificador-solicitud",
    "data": "valor-solicitado"
}
```

**Campos:**
- `type` (string, requerido): DEBE ser "get"
- `id_ret` (string, requerido): Identificador de solicitud del servidor
- `data` (any, requerido): Valor de datos solicitado

### 3.2 Mensajes Servidor-a-Cliente

#### 3.2.1 Mensaje Fill

Reemplazar innerHTML del elemento.

```json
{
    "type": "fill",
    "id": "id-elemento",
    "value": "<html>contenido</html>"
}
```

#### 3.2.2 Mensaje Text

Establecer contenido de texto del elemento.

```json
{
    "type": "text",
    "id": "id-elemento",
    "value": "texto plano"
}
```

#### 3.2.3 Mensaje Style

Establecer estilos CSS del elemento.

```json
{
    "type": "style",
    "id": "id-elemento",
    "value": "color: red; font-size: 14px"
}
```

#### 3.2.4 Mensaje Set

Establecer propiedad value del elemento.

```json
{
    "type": "set",
    "id": "id-elemento",
    "value": "valor del input"
}
```

#### 3.2.5 Mensaje Property

Establecer propiedad arbitraria del elemento.

```json
{
    "type": "propertie",
    "id": "id-elemento",
    "propertie": "disabled",
    "value": true
}
```

#### 3.2.6 Mensaje Remove

Eliminar elemento del DOM.

```json
{
    "type": "remove",
    "id": "id-elemento"
}
```

#### 3.2.7 Mensaje Add Node

Agregar nodo hijo al elemento.

```json
{
    "type": "addNode",
    "id": "id-padre",
    "value": "<div>nuevo hijo</div>"
}
```

#### 3.2.8 Mensaje Script

Ejecutar código JavaScript.

```json
{
    "type": "script",
    "value": "console.log('ejecutado')"
}
```

**Advertencia de Seguridad:** La ejecución de scripts requiere servidor confiable.

#### 3.2.9 Mensaje Get Request

Solicitar datos del cliente.

```json
{
    "type": "get",
    "id": "id-elemento",
    "id_ret": "solicitud-123",
    "sub_type": "value",
    "value": "nombrePropiedad"
}
```

**Sub-tipos:**
- `value`: Obtener element.value
- `html`: Obtener element.innerHTML
- `text`: Obtener element.innerText
- `style`: Obtener propiedad de estilo (nombre de propiedad en `value`)
- `propertie`: Obtener cualquier propiedad (nombre de propiedad en `value`)

## 4. Interacción con el DOM

### 4.1 Identificación de Elementos

Los elementos se identifican por su atributo HTML `id`. El servidor DEBE asegurar IDs únicos dentro del documento.

### 4.2 Vinculación de Eventos

Los eventos pueden vincularse usando manejadores en línea:

```html
<button onclick="send_event('btn-1', 'Click', null)">Haz clic</button>
<input oninput="send_event('input-1', 'Input', this.value)">
```

### 4.3 Asociación de Componentes

Los elementos pueden asociarse con componentes del lado del servidor:

```html
<div data-component-id="mi-componente">
    <!-- Contenido del componente -->
</div>
```

## 5. Protocolo de Arrastrar y Soltar

### 5.1 Elementos Arrastrables

Los elementos se vuelven arrastrables agregando la clase `draggable`:

```html
<div class="draggable" 
     id="elemento-1"
     data-element-id="elemento-1"
     data-component-id="componente-1">
    Contenido Arrastrable
</div>
```

### 5.2 Eventos de Arrastre

El cliente envía automáticamente tres eventos durante las operaciones de arrastre:

1. **DragStart**: Cuando comienza el arrastre
2. **DragMove**: Durante el arrastre (limitado a 60 FPS)
3. **DragEnd**: Cuando termina el arrastre

Formato de datos del evento:
```json
{
    "element": "identificador-elemento",
    "x": 100,
    "y": 200
}
```

### 5.3 Deshabilitar Arrastre

Deshabilitar arrastre temporalmente:

```html
<div class="draggable" data-drag-disabled="true">
    No arrastrable
</div>
```

## 6. Consideraciones de Seguridad

### 6.1 Seguridad del Transporte

- **DEBE** usar WSS (WebSocket Seguro) en producción
- **DEBE** validar certificados SSL/TLS
- **DEBERÍA** implementar políticas CORS

### 6.2 Validación de Entrada

- **DEBE** sanear toda entrada del usuario en el servidor
- **DEBE** escapar contenido HTML para prevenir XSS
- **DEBE** validar formato y tipos de mensajes

### 6.3 Ejecución de Scripts

- **DEBERÍA** evitar usar mensajes de script cuando sea posible
- **NUNCA DEBE** ejecutar código proporcionado por el usuario
- **DEBE** auditar todos los scripts generados por el servidor

### 6.4 Autenticación

- **DEBERÍA** autenticar conexiones WebSocket
- **DEBERÍA** implementar gestión de sesiones
- **DEBE** validar autorización para operaciones

## 7. Guías de Rendimiento

### 7.1 Optimización de Mensajes

- Agrupar actualizaciones DOM relacionadas en mensajes únicos
- Usar `fill` para actualizaciones complejas, `text` para texto simple
- Minimizar tamaño de mensaje enviando solo datos cambiados

### 7.2 Limitación de Frecuencia

- Eventos de arrastre: Máximo 60 FPS (intervalos de 16ms)
- Eventos de entrada: Considerar debouncing para entradas rápidas
- Eventos de scroll: Limitar para prevenir inundación

### 7.3 Gestión de Conexiones

- Implementar pooling de conexiones para múltiples componentes
- Usar heartbeat/ping para detectar conexiones obsoletas
- Limpiar recursos del servidor al desconectar

## 8. Requisitos de Implementación

### 8.1 Requisitos del Cliente

Los clientes DEBEN:
1. Establecer conexión WebSocket al servidor
2. Manejar todos los tipos de mensaje definidos en Sección 3.2
3. Enviar eventos en formato especificado en Sección 3.1
4. Implementar reconexión automática
5. Proporcionar función global `send_event`
6. Soportar arrastrar y soltar según lo especificado

### 8.2 Requisitos del Servidor

Los servidores DEBEN:
1. Aceptar conexiones WebSocket en endpoints LiveView
2. Enviar solo mensajes JSON válidos
3. Mantener estado del componente entre mensajes
4. Manejar eventos del cliente apropiadamente
5. Limpiar recursos al desconectar
6. Generar IDs de elementos únicos

### 8.3 Características Opcionales

Las implementaciones PUEDEN:
1. Soportar frames WebSocket binarios para transferencias de archivos
2. Implementar compresión de mensajes
3. Agregar tipos de mensaje personalizados con prefijo "x-"
4. Soportar múltiples conexiones WebSocket por cliente
5. Implementar seguimiento de presencia

## 9. Manejo de Errores

### 9.1 Errores de Conexión

- Cliente DEBE intentar reconexión en desconexión inesperada
- Servidor DEBERÍA registrar fallas de conexión
- Ambos DEBEN manejar transmisión parcial de mensajes

### 9.2 Errores de Mensaje

- JSON inválido: Registrar error, ignorar mensaje
- Tipo de mensaje desconocido: Registrar advertencia, ignorar mensaje
- Campos requeridos faltantes: Registrar error, ignorar mensaje
- Elemento no encontrado: Registrar advertencia, continuar operación

### 9.3 Estrategias de Recuperación

1. **Recuperación del Cliente**: Reconectar y solicitar actualización completa del estado
2. **Recuperación del Servidor**: Reconstruir estado del componente desde almacenamiento persistente
3. **Degradación Elegante**: Mostrar mensajes de error amigables al usuario

## 10. Extensibilidad

### 10.1 Tipos de Mensaje Personalizados

Los tipos de mensaje personalizados DEBEN usar prefijo "x-":

```json
{
    "type": "x-personalizado",
    "id": "id-elemento",
    "campoPersonalizado": "valor"
}
```

### 10.2 Versionado del Protocolo

Las versiones futuras DEBEN:
1. Mantener compatibilidad hacia atrás para versión mayor 1
2. Usar negociación de versión durante handshake
3. Documentar todos los cambios incompatibles

## 11. Ejemplos

### 11.1 Clic de Botón Simple

**HTML:**
```html
<button id="btn-1" onclick="send_event('contador', 'Incrementar', null)">
    Cuenta: <span id="cuenta">0</span>
</button>
```

**Cliente envía:**
```json
{"type": "data", "id": "contador", "event": "Incrementar", "data": ""}
```

**Servidor responde:**
```json
{"type": "text", "id": "cuenta", "value": "1"}
```

### 11.2 Input de Formulario

**HTML:**
```html
<input id="input-nombre" oninput="send_event('formulario', 'NombreCambiado', this.value)">
<div id="saludo"></div>
```

**Cliente envía:**
```json
{"type": "data", "id": "formulario", "event": "NombreCambiado", "data": "Alicia"}
```

**Servidor responde:**
```json
{"type": "text", "id": "saludo", "value": "¡Hola, Alicia!"}
```

### 11.3 Lista Dinámica

**Servidor envía:**
```json
{
    "type": "fill",
    "id": "lista",
    "value": "<ul><li>Elemento 1</li><li>Elemento 2</li><li>Elemento 3</li></ul>"
}
```

## 12. Referencias

- [RFC 6455] El Protocolo WebSocket
- [RFC 7159] Notación de Objetos JavaScript (JSON)
- [HTML5] Estándar HTML Viviente
- [DOM] Especificación del Modelo de Objetos del Documento

## Apéndice A: API de la Biblioteca Cliente

### Funciones Globales

- `send_event(id, evento, datos)`: Enviar evento al servidor
- `connect()`: Establecer conexión manualmente

### Variables Globales

- `window.ws`: Instancia WebSocket
- `window.dragState`: Estado actual de operación de arrastre

## Apéndice B: Niveles de Conformidad

- **DEBE**: Requerido para cumplimiento
- **DEBERÍA**: Recomendado para interoperabilidad
- **PUEDE**: Mejora opcional

## Apéndice C: Registro de Cambios

- **v1.0.0** (2025-08-22): Especificación inicial

## Aviso de Copyright

Este documento se publica bajo la Licencia MIT. Las implementaciones pueden usar libremente esta especificación.