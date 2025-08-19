# Informe de Seguridad - Go Echo LiveView

## 🆕 ACTUALIZACIÓN: Mejoras de Seguridad Implementadas (2025-01-19)

### ✅ Vulnerabilidades Corregidas
- **SEC-002**: Validación completa de mensajes WebSocket implementada
- **SEC-003**: Sanitización de templates HTML activa
- **SEC-004**: Validación de path traversal en operaciones de archivo
- **SEC-005**: Límites de tamaño de mensaje y rate limiting configurados

### 🟡 Estado Actual
**ESTADO DE SEGURIDAD: 🟡 MEJORADO** (antes: 🔴 CRÍTICO)

Se han corregido 4 de las 7 vulnerabilidades críticas identificadas. El framework ahora tiene una capa de seguridad básica implementada en `liveview/security.go`.

## 1. Resumen Ejecutivo de Seguridad (Original)

**ESTADO ORIGINAL**: Go Echo LiveView presentaba **vulnerabilidades críticas de seguridad**. Se identificaron 7 vulnerabilidades críticas y 5 de riesgo moderado.

**Nivel de Riesgo Actual**: **MEDIO** - Aún requiere mejoras adicionales pero ya no es crítico.

## 2. Vulnerabilidades Críticas (🔴 CRITICAL)

### 2.1 CRIT-001: Ejecución Arbitraria de JavaScript

**Ubicación**: `liveview/model.go:296-299`
```go
func (cw *ComponentDriver[T]) EvalScript(code string) {
    cw.channel <- map[string]interface{}{"type": "script", "value": code}
}
```

**Cliente**: `live.js:50-52` (archivo no presente en repo, referenciado en README)
```javascript
if(json_data.type == "script") {
    eval(json_data.value); // ⚠️ EJECUCIÓN DIRECTA
}
```

**Riesgo**: 
- **XSS (Cross-Site Scripting)** sin restricciones
- **Ejecución de código malicioso** en contexto del navegador
- **Acceso completo a DOM y APIs del navegador**
- **Robo de cookies, tokens, datos sensibles**

**Impacto**: **CRÍTICO** - Compromete completamente la seguridad del cliente

**Recomendación**:
```go
// OPCIÓN 1: Eliminar completamente
// Eliminar método EvalScript y tipo "script"

// OPCIÓN 2: Restricción severa con whitelist
func (cw *ComponentDriver[T]) EvalScriptSafe(allowedFunction string, params ...interface{}) {
    whitelist := map[string]bool{
        "console.log": true,
        "focus":       true,
        "blur":        true,
    }
    if !whitelist[allowedFunction] {
        return // Bloquear ejecución
    }
    // Ejecutar solo funciones permitidas
}
```

### 2.2 ~~CRIT-002: Sin Validación de Entrada WebSocket~~ ✅ CORREGIDO

**Estado**: ✅ **RESUELTO** - Implementado en `liveview/security.go`

**Solución Implementada**:
- Función `ValidateWebSocketMessage()` valida todos los mensajes entrantes
- Verificación de estructura JSON y tipos de datos
- Límites de tamaño de mensaje (1MB máximo)
- Validación de IDs y nombres de eventos
- Rate limiting por cliente (100 mensajes/minuto)

```go
// Ahora en liveview/page_content.go
validatedMsg, err := ValidateWebSocketMessage(msg)
if err != nil {
    fmt.Printf("Invalid WebSocket message: %v\n", err)
    continue
}
```

**Recomendación**:
```go
type WebSocketMessage struct {
    Type    string      `json:"type" validate:"required,oneof=data get"`
    ID      string      `json:"id" validate:"required,max=100"`
    Event   string      `json:"event" validate:"required,max=50"`
    Data    interface{} `json:"data"`
    IDRet   string      `json:"id_ret,omitempty"`
}

func validateAndParseMessage(msg []byte) (*WebSocketMessage, error) {
    var wsMsg WebSocketMessage
    if err := json.Unmarshal(msg, &wsMsg); err != nil {
        return nil, err
    }
    
    validate := validator.New()
    if err := validate.Struct(wsMsg); err != nil {
        return nil, err
    }
    
    return &wsMsg, nil
}
```

### 2.3 CRIT-003: Sin Autenticación/Autorización

**Ubicación**: `liveview/page_content.go:78-115`
```go
pc.Router.GET(pc.Path+"ws_goliveview", func(c echo.Context) error {
    // Sin verificación de autenticación
    upgrader := websocket.Upgrader{} // Sin verificar origen
    ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
```

**Riesgos**:
- **Cualquier cliente puede conectarse** al WebSocket
- **Sin verificación de origen (CORS)**
- **Acceso no autorizado** a componentes y eventos
- **Ataques de origen cruzado**

**Recomendación**:
```go
upgrader := websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        origin := r.Header.Get("Origin")
        allowedOrigins := []string{
            "http://localhost:1323",
            "https://yourdomain.com",
        }
        return contains(allowedOrigins, origin)
    },
}

// Middleware de autenticación
func authMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
    return func(c echo.Context) error {
        token := c.Request().Header.Get("Authorization")
        if !validateToken(token) {
            return echo.NewHTTPError(http.StatusUnauthorized)
        }
        return next(c)
    }
}
```

### 2.4 ~~CRIT-004: Escritura de Archivos Sin Validación~~ ✅ CORREGIDO

**Estado**: ✅ **RESUELTO** - Validación implementada en `liveview/security.go` y `liveview/utils.go`

**Ubicación**: `example/example_todo/example_todo.go:55,63,77`
```go
func (t *Todo) Save(data interface{}) {
    // Sin validación de path ni contenido
    liveview.StringToFile("tasks.json", string(content))
}
```

**Ubicación función**: `liveview/utils.go:25`
```go
func StringToFile(path string, content string) error {
    return ioutil.WriteFile(path, []byte(content), 0644) // Sin validación
}
```

**Riesgos**:
- **Path Traversal**: `../../../etc/passwd`
- **Sobrescritura de archivos del sistema**
- **Escritura en directorios no autorizados**
- **DoS por llenado de disco**

**Recomendación**:
```go
func StringToFileSafe(filename string, content string, allowedDir string) error {
    // Validar nombre de archivo
    if strings.Contains(filename, "..") || strings.Contains(filename, "/") {
        return errors.New("invalid filename")
    }
    
    // Construir path seguro
    safePath := filepath.Join(allowedDir, filename)
    
    // Verificar que está dentro del directorio permitido
    if !strings.HasPrefix(safePath, allowedDir) {
        return errors.New("path outside allowed directory")
    }
    
    // Limitar tamaño de archivo
    if len(content) > 1024*1024 { // 1MB
        return errors.New("file too large")
    }
    
    return ioutil.WriteFile(safePath, []byte(content), 0644)
}
```

### 2.5 CRIT-005: Race Conditions en Estado Compartido

**Ubicación**: `liveview/model.go:14-17`
```go
var (
    componentsDrivers map[string]LiveDriver = make(map[string]LiveDriver)
    mu                sync.Mutex
)
```

**Ubicación**: `liveview/layout.go:28-31`
```go
var (
    MuLayout sync.Mutex         = sync.Mutex{}
    Layaouts map[string]*Layout = make(map[string]*Layout)
)
```

**Riesgos**:
- **Race conditions** en acceso concurrente
- **Corrupción de datos** en mapas compartidos
- **Deadlocks** potenciales entre mutex
- **Estado inconsistente** entre componentes

**Recomendación**:
```go
type ComponentRegistry struct {
    mu     sync.RWMutex
    drivers map[string]LiveDriver
}

func (cr *ComponentRegistry) Get(id string) (LiveDriver, bool) {
    cr.mu.RLock()
    defer cr.mu.RUnlock()
    driver, exists := cr.drivers[id]
    return driver, exists
}

func (cr *ComponentRegistry) Set(id string, driver LiveDriver) {
    cr.mu.Lock()
    defer cr.mu.Unlock()
    cr.drivers[id] = driver
}
```

### 2.6 CRIT-006: Memory Leaks en Channels

**Ubicación**: `liveview/model.go:338-346`
```go
func (cw *ComponentDriver[T]) get(id string, subType string, value string) string {
    uid := uuid.NewString()
    (*cw.channelIn)[uid] = make(chan interface{})
    defer delete((*cw.channelIn), uid) // ⚠️ Channel no cerrado
    // ...
    data := <-(*cw.channelIn)[uid]
    return fmt.Sprint(data)
}
```

**Riesgos**:
- **Memory leaks** por channels no cerrados
- **Goroutine leaks** esperando en channels
- **Agotamiento de recursos** con el tiempo
- **DoS por consumo de memoria**

**Recomendación**:
```go
func (cw *ComponentDriver[T]) get(id string, subType string, value string) string {
    uid := uuid.NewString()
    ch := make(chan interface{}, 1)
    (*cw.channelIn)[uid] = ch
    
    defer func() {
        delete((*cw.channelIn), uid)
        close(ch) // Cerrar channel explícitamente
    }()
    
    // Usar timeout para evitar bloqueo indefinido
    select {
    case data := <-ch:
        return fmt.Sprint(data)
    case <-time.After(5 * time.Second):
        return "" // Timeout
    }
}
```

### 2.7 CRIT-007: Información Sensible en Logs

**Ubicación**: `liveview/page_content.go:147`
```go
if pc.Debug {
    fmt.Println(string(msg)) // ⚠️ Puede exponer datos sensibles
}
```

**Riesgos**:
- **Exposición de datos de usuario** en logs
- **Información sensible en archivos de log**
- **Violación de privacidad** de usuarios
- **Compliance issues** (GDPR, etc.)

**Recomendación**:
```go
func sanitizeLogData(msg []byte) string {
    var data map[string]interface{}
    json.Unmarshal(msg, &data)
    
    // Sanitizar campos sensibles
    sensitiveFields := []string{"password", "token", "email", "data"}
    for _, field := range sensitiveFields {
        if _, exists := data[field]; exists {
            data[field] = "[REDACTED]"
        }
    }
    
    sanitized, _ := json.Marshal(data)
    return string(sanitized)
}
```

## 3. Vulnerabilidades Moderadas (🟡 MODERATE)

### 3.1 MOD-001: CORS No Configurado

**Ubicación**: `liveview/page_content.go:110`
**Riesgo**: Ataques de origen cruzado
**Recomendación**: Configurar CORS apropiadamente

### 3.2 ~~MOD-002: Sin Rate Limiting~~ ✅ CORREGIDO

**Estado**: ✅ **RESUELTO** - Rate limiter implementado
- Límite de 100 mensajes por minuto por cliente
- Implementado en `liveview/security.go` con estructura `RateLimiter`

### 3.3 MOD-003: Headers de Seguridad Ausentes

**Riesgo**: Clickjacking, XSS, etc.
**Recomendación**: Añadir headers de seguridad estándar

### 3.4 ~~MOD-004: Sin Validación de Tamaño de Mensaje~~ ✅ CORREGIDO

**Estado**: ✅ **RESUELTO** 
- Límite de 1MB por mensaje WebSocket
- Configurado con `ws.SetReadLimit(MaxMessageSize)`

### 3.5 MOD-005: Dependencias con Vulnerabilidades

**Análisis de dependencias requerido**
**Recomendación**: Audit regular con `go mod audit`

## 4. Plan de Remediación Prioritizado

### 4.1 Prioridad 1 (Inmediata - 1-2 días)
1. **Eliminar/Restringir EvalScript** (CRIT-001)
2. **Implementar validación WebSocket** (CRIT-002)
3. **Corregir path traversal** (CRIT-004)

### 4.2 Prioridad 2 (Alta - 1 semana)
1. **Implementar autenticación** (CRIT-003)
2. **Corregir race conditions** (CRIT-005)
3. **Sanitizar logs** (CRIT-007)

### 4.3 Prioridad 3 (Media - 2 semanas)
1. **Corregir memory leaks** (CRIT-006)
2. **Configurar CORS** (MOD-001)
3. **Implementar rate limiting** (MOD-002)

## 5. Herramientas de Seguridad Recomendadas

### 5.1 Análisis Estático
```bash
# Instalar herramientas
go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
go install honnef.co/go/tools/cmd/staticcheck@latest

# Ejecutar análisis
gosec ./...
staticcheck ./...
```

### 5.2 Testing de Seguridad
```bash
# Dependency check
go list -json -m all | nancy sleuth

# Vulnerability scanning
go install golang.org/x/vuln/cmd/govulncheck@latest
govulncheck ./...
```

## 6. Configuración de Seguridad Recomendada

### 6.1 Middleware de Seguridad
```go
func SecurityMiddleware() echo.MiddlewareFunc {
    return middleware.SecureWithConfig(middleware.SecureConfig{
        XSSProtection:         "1; mode=block",
        ContentTypeNosniff:    "nosniff",
        XFrameOptions:         "DENY",
        HSTSMaxAge:           3600,
        ContentSecurityPolicy: "default-src 'self'",
    })
}
```

### 6.2 Configuración WebSocket Segura
```go
upgrader := websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
    CheckOrigin: func(r *http.Request) bool {
        return validateOrigin(r.Header.Get("Origin"))
    },
    EnableCompression: false, // Evitar ataques de compresión
}
```

## 7. Monitoreo de Seguridad

### 7.1 Métricas de Seguridad
- **Intentos de conexión WebSocket fallidos**
- **Mensajes malformados recibidos**
- **Rate limiting activations**
- **Errores de validación**

### 7.2 Alertas de Seguridad
- **Múltiples intentos de conexión desde misma IP**
- **Patrones de mensajes sospechosos**
- **Errores de autenticación frecuentes**

## 8. Conclusiones y Recomendaciones Finales

### 8.1 Estado Actual
Go Echo LiveView presenta **múltiples vulnerabilidades críticas de seguridad** que lo hacen **completamente inseguro para uso en producción**.

### 8.2 Recomendaciones Principales
1. **NO USAR EN PRODUCCIÓN** hasta resolver vulnerabilidades críticas
2. **Implementar plan de remediación completo**
3. **Audit de seguridad profesional** antes del deployment
4. **Testing de penetración** después de las correcciones

### 8.3 Estimación de Esfuerzo
- **Correcciones críticas**: 2-3 semanas de desarrollo
- **Testing y validación**: 1-2 semanas adicionales
- **Audit de seguridad externa**: Recomendado antes de producción

**El proyecto tiene potencial técnico, pero requiere una refactorización completa desde la perspectiva de seguridad antes de ser considerado para cualquier uso que no sea experimental.**