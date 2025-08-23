# RFC-LIVE: LiveView Protocol Specification

**Version:** 1.0.0  
**Status:** Draft  
**Date:** 2025-08-22  
**Authors:** Go Echo LiveView Contributors

## Abstract

This document specifies the LiveView Protocol, a WebSocket-based communication protocol that enables real-time, server-driven user interface updates without requiring client-side application logic. The protocol allows servers to control DOM manipulation, handle events, and maintain application state while clients act as thin presentation layers.

## 1. Introduction

### 1.1 Purpose

The LiveView Protocol enables server-side frameworks to create interactive web applications where the server maintains complete control over the UI state and logic. This approach simplifies development by eliminating the need for client-side state management and API design.

### 1.2 Design Goals

1. **Simplicity**: Minimal client-side complexity
2. **Real-time**: Low-latency bidirectional communication
3. **Reliability**: Automatic reconnection and error recovery
4. **Compatibility**: Works with standard web technologies
5. **Flexibility**: Framework and language agnostic

### 1.3 Terminology

- **Server**: The backend application implementing business logic
- **Client**: The JavaScript library running in the browser
- **Component**: A server-side UI element with state and behavior
- **Event**: A user interaction or system occurrence
- **Message**: A JSON-encoded communication packet

## 2. Transport Layer

### 2.1 WebSocket Connection

The protocol uses WebSocket (RFC 6455) as the transport layer.

#### 2.1.1 Connection URL Pattern

```
ws://[host]:[port][path]ws_goliveview
wss://[host]:[port][path]ws_goliveview
```

- Use `ws://` for HTTP origins
- Use `wss://` for HTTPS origins (required for production)
- The suffix `ws_goliveview` identifies LiveView endpoints

#### 2.1.2 Connection Lifecycle

1. **Initial HTTP Request**: Client requests the page via HTTP
2. **HTML Response**: Server returns HTML with LiveView client script
3. **WebSocket Upgrade**: Client establishes WebSocket connection
4. **Component Initialization**: Server sends initial UI state
5. **Interactive Session**: Bidirectional event/update exchange
6. **Disconnection Handling**: Automatic reconnection on failure

### 2.2 Reconnection Strategy

Clients MUST implement automatic reconnection:

```javascript
// Reconnection interval: 1000ms (1 second)
// Maximum attempts: Unlimited
// Backoff strategy: None (constant interval)
```

## 3. Message Format

All messages MUST be valid JSON objects encoded as UTF-8 strings.

### 3.1 Client-to-Server Messages

#### 3.1.1 Event Message

Used to send user interactions to the server.

```json
{
    "type": "data",
    "id": "component-identifier",
    "event": "EventName",
    "data": "event-specific-data"
}
```

**Fields:**
- `type` (string, required): MUST be "data"
- `id` (string, required): Component identifier
- `event` (string, required): Event name (e.g., "Click", "Input")
- `data` (string, optional): Event data as string or JSON string

#### 3.1.2 Get Response Message

Response to server's data request.

```json
{
    "type": "get",
    "id_ret": "request-identifier",
    "data": "requested-value"
}
```

**Fields:**
- `type` (string, required): MUST be "get"
- `id_ret` (string, required): Request identifier from server
- `data` (any, required): Requested data value

### 3.2 Server-to-Client Messages

#### 3.2.1 Fill Message

Replace element's innerHTML.

```json
{
    "type": "fill",
    "id": "element-id",
    "value": "<html>content</html>"
}
```

#### 3.2.2 Text Message

Set element's text content.

```json
{
    "type": "text",
    "id": "element-id",
    "value": "plain text"
}
```

#### 3.2.3 Style Message

Set element's CSS styles.

```json
{
    "type": "style",
    "id": "element-id",
    "value": "color: red; font-size: 14px"
}
```

#### 3.2.4 Set Message

Set element's value property.

```json
{
    "type": "set",
    "id": "element-id",
    "value": "input value"
}
```

#### 3.2.5 Property Message

Set arbitrary element property.

```json
{
    "type": "propertie",
    "id": "element-id",
    "propertie": "disabled",
    "value": true
}
```

#### 3.2.6 Remove Message

Remove element from DOM.

```json
{
    "type": "remove",
    "id": "element-id"
}
```

#### 3.2.7 Add Node Message

Add child node to element.

```json
{
    "type": "addNode",
    "id": "parent-id",
    "value": "<div>new child</div>"
}
```

#### 3.2.8 Script Message

Execute JavaScript code.

```json
{
    "type": "script",
    "value": "console.log('executed')"
}
```

**Security Warning:** Script execution requires trusted server.

#### 3.2.9 Get Request Message

Request data from client.

```json
{
    "type": "get",
    "id": "element-id",
    "id_ret": "request-123",
    "sub_type": "value",
    "value": "propertyName"
}
```

**Sub-types:**
- `value`: Get element.value
- `html`: Get element.innerHTML
- `text`: Get element.innerText
- `style`: Get style property (property name in `value`)
- `propertie`: Get any property (property name in `value`)

## 4. DOM Interaction

### 4.1 Element Identification

Elements are identified by their HTML `id` attribute. The server MUST ensure unique IDs within the document.

### 4.2 Event Binding

Events can be bound using inline handlers:

```html
<button onclick="send_event('btn-1', 'Click', null)">Click Me</button>
<input oninput="send_event('input-1', 'Input', this.value)">
```

### 4.3 Component Association

Elements can be associated with server-side components:

```html
<div data-component-id="my-component">
    <!-- Component content -->
</div>
```

## 5. Drag and Drop Protocol

### 5.1 Draggable Elements

Elements become draggable by adding the `draggable` class:

```html
<div class="draggable" 
     id="element-1"
     data-element-id="element-1"
     data-component-id="component-1">
    Draggable Content
</div>
```

### 5.2 Drag Events

The client automatically sends three events during drag operations:

1. **DragStart**: When dragging begins
2. **DragMove**: During dragging (throttled to 60 FPS)
3. **DragEnd**: When dragging completes

Event data format:
```json
{
    "element": "element-identifier",
    "x": 100,
    "y": 200
}
```

### 5.3 Disabling Drag

Temporarily disable dragging:

```html
<div class="draggable" data-drag-disabled="true">
    Not draggable
</div>
```

## 6. Security Considerations

### 6.1 Transport Security

- **MUST** use WSS (WebSocket Secure) in production
- **MUST** validate SSL/TLS certificates
- **SHOULD** implement CORS policies

### 6.2 Input Validation

- **MUST** sanitize all user input on server
- **MUST** escape HTML content to prevent XSS
- **MUST** validate message format and types

### 6.3 Script Execution

- **SHOULD** avoid using script messages when possible
- **MUST** never execute user-provided code
- **MUST** audit all server-generated scripts

### 6.4 Authentication

- **SHOULD** authenticate WebSocket connections
- **SHOULD** implement session management
- **MUST** validate authorization for operations

## 7. Performance Guidelines

### 7.1 Message Optimization

- Batch related DOM updates into single messages
- Use `fill` for complex updates, `text` for simple text
- Minimize message size by sending only changed data

### 7.2 Throttling

- Drag events: Maximum 60 FPS (16ms intervals)
- Input events: Consider debouncing rapid inputs
- Scroll events: Throttle to prevent flooding

### 7.3 Connection Management

- Implement connection pooling for multiple components
- Use heartbeat/ping to detect stale connections
- Clean up server resources on disconnect

## 8. Implementation Requirements

### 8.1 Client Requirements

Clients MUST:
1. Establish WebSocket connection to server
2. Handle all message types defined in Section 3.2
3. Send events in format specified in Section 3.1
4. Implement automatic reconnection
5. Provide global `send_event` function
6. Support drag and drop as specified

### 8.2 Server Requirements

Servers MUST:
1. Accept WebSocket connections at LiveView endpoints
2. Send only valid JSON messages
3. Maintain component state between messages
4. Handle client events appropriately
5. Clean up resources on disconnect
6. Generate unique element IDs

### 8.3 Optional Features

Implementations MAY:
1. Support binary WebSocket frames for file transfers
2. Implement message compression
3. Add custom message types with "x-" prefix
4. Support multiple WebSocket connections per client
5. Implement presence tracking

## 9. Error Handling

### 9.1 Connection Errors

- Client MUST attempt reconnection on unexpected disconnect
- Server SHOULD log connection failures
- Both MUST handle partial message transmission

### 9.2 Message Errors

- Invalid JSON: Log error, ignore message
- Unknown message type: Log warning, ignore message
- Missing required fields: Log error, ignore message
- Element not found: Log warning, continue operation

### 9.3 Recovery Strategies

1. **Client Recovery**: Reconnect and request full state refresh
2. **Server Recovery**: Rebuild component state from persistent storage
3. **Graceful Degradation**: Show user-friendly error messages

## 10. Extensibility

### 10.1 Custom Message Types

Custom message types MUST use "x-" prefix:

```json
{
    "type": "x-custom",
    "id": "element-id",
    "customField": "value"
}
```

### 10.2 Protocol Versioning

Future versions MUST:
1. Maintain backward compatibility for major version 1
2. Use version negotiation during handshake
3. Document all breaking changes

## 11. Examples

### 11.1 Simple Button Click

**HTML:**
```html
<button id="btn-1" onclick="send_event('counter', 'Increment', null)">
    Count: <span id="count">0</span>
</button>
```

**Client sends:**
```json
{"type": "data", "id": "counter", "event": "Increment", "data": ""}
```

**Server responds:**
```json
{"type": "text", "id": "count", "value": "1"}
```

### 11.2 Form Input

**HTML:**
```html
<input id="name-input" oninput="send_event('form', 'NameChanged', this.value)">
<div id="greeting"></div>
```

**Client sends:**
```json
{"type": "data", "id": "form", "event": "NameChanged", "data": "Alice"}
```

**Server responds:**
```json
{"type": "text", "id": "greeting", "value": "Hello, Alice!"}
```

### 11.3 Dynamic List

**Server sends:**
```json
{
    "type": "fill",
    "id": "list",
    "value": "<ul><li>Item 1</li><li>Item 2</li><li>Item 3</li></ul>"
}
```

## 12. References

- [RFC 6455] The WebSocket Protocol
- [RFC 7159] JavaScript Object Notation (JSON)
- [HTML5] HTML Living Standard
- [DOM] Document Object Model Specification

## Appendix A: Client Library API

### Global Functions

- `send_event(id, event, data)`: Send event to server
- `connect()`: Manually establish connection

### Global Variables

- `window.ws`: WebSocket instance
- `window.dragState`: Current drag operation state

## Appendix B: Conformance Levels

- **MUST**: Required for compliance
- **SHOULD**: Recommended for interoperability
- **MAY**: Optional enhancement

## Appendix C: Change Log

- **v1.0.0** (2025-08-22): Initial specification

## Copyright Notice

This document is released under the MIT License. Implementations may freely use this specification.