# Go Echo LiveView - Best Practices Guide

[English](#english) | [Español](#español)

---

## English

# Best Practices for Go Echo LiveView Development

## Table of Contents
1. [Component Design](#component-design)
2. [State Management](#state-management)
3. [Performance Optimization](#performance-optimization)
4. [Security Best Practices](#security-best-practices)
5. [Testing Strategies](#testing-strategies)
6. [Error Handling](#error-handling)
7. [Code Organization](#code-organization)
8. [Production Deployment](#production-deployment)
9. [Monitoring and Debugging](#monitoring-and-debugging)
10. [Common Pitfalls](#common-pitfalls)

## Component Design

### Single Responsibility Principle

Each component should have a single, well-defined purpose:

✅ **Good:**
```go
// UserProfile handles only user profile display
type UserProfile struct {
    *liveview.ComponentDriver[*UserProfile]
    User User
}

// UserSettings handles only user settings
type UserSettings struct {
    *liveview.ComponentDriver[*UserSettings]
    Settings Settings
}
```

❌ **Bad:**
```go
// UserComponent does too much
type UserComponent struct {
    *liveview.ComponentDriver[*UserComponent]
    User     User
    Settings Settings
    Orders   []Order
    Messages []Message
}
```

### Component Composition

Use composition for complex UIs:

```go
type Dashboard struct {
    *liveview.ComponentDriver[*Dashboard]
    Header    *Header
    Sidebar   *Sidebar
    Analytics *AnalyticsPanel
    Activity  *ActivityFeed
}

func (d *Dashboard) Start() {
    // Initialize child components
    d.Header = &Header{}
    d.Sidebar = &Sidebar{}
    d.Analytics = &AnalyticsPanel{}
    d.Activity = &ActivityFeed{}
    
    // Mount components
    d.Mount(liveview.NewDriver("header", d.Header))
    d.Mount(liveview.NewDriver("sidebar", d.Sidebar))
    d.Mount(liveview.NewDriver("analytics", d.Analytics))
    d.Mount(liveview.NewDriver("activity", d.Activity))
    
    d.Commit()
}
```

### Template Organization

Keep templates clean and maintainable:

```go
func (c *Component) GetTemplate() string {
    return `
    {{define "styles"}}
    <style>
        .component { padding: 20px; }
        .title { font-size: 24px; }
    </style>
    {{end}}
    
    {{define "content"}}
    <div class="component">
        <h1 class="title">{{.Title}}</h1>
        <div class="body">{{.Content}}</div>
    </div>
    {{end}}
    
    {{template "styles" .}}
    {{template "content" .}}
    `
}
```

## State Management

### Initialize State Properly

Always initialize state in the `Start()` method:

```go
func (c *Component) Start() {
    // Initialize collections
    c.Items = make([]Item, 0)
    c.Users = make(map[string]User)
    
    // Set defaults
    c.CurrentPage = 1
    c.PageSize = 20
    c.SortOrder = "asc"
    
    // Load initial data
    c.LoadData()
    
    // Always commit after initialization
    c.Commit()
}
```

### Use Contexts for Cancellation

Manage long-running operations with contexts:

```go
type DataLoader struct {
    *liveview.ComponentDriver[*DataLoader]
    ctx    context.Context
    cancel context.CancelFunc
    Data   []DataPoint
}

func (d *DataLoader) Start() {
    d.ctx, d.cancel = context.WithCancel(context.Background())
    d.LoadDataAsync()
    d.Commit()
}

func (d *DataLoader) LoadDataAsync() {
    go func() {
        select {
        case <-d.ctx.Done():
            return
        case data := <-fetchData():
            d.Data = data
            d.Commit()
        }
    }()
}

func (d *DataLoader) Stop() {
    if d.cancel != nil {
        d.cancel()
    }
}
```

### Avoid State Mutations Without Commit

Always call `Commit()` after state changes:

✅ **Good:**
```go
func (c *Component) UpdateItem(data interface{}) {
    if id, ok := data.(int); ok {
        c.Items[id].Updated = true
        c.Commit() // Always commit changes
    }
}
```

❌ **Bad:**
```go
func (c *Component) UpdateItem(data interface{}) {
    if id, ok := data.(int); ok {
        c.Items[id].Updated = true
        // Missing Commit() - UI won't update!
    }
}
```

## Performance Optimization

### Debounce Frequent Updates

For search boxes and real-time filters:

```go
type SearchComponent struct {
    *liveview.ComponentDriver[*SearchComponent]
    Query     string
    Results   []Result
    debouncer *time.Timer
    mu        sync.Mutex
}

func (s *SearchComponent) Search(data interface{}) {
    s.mu.Lock()
    defer s.mu.Unlock()
    
    if query, ok := data.(string); ok {
        s.Query = query
        
        // Cancel previous search
        if s.debouncer != nil {
            s.debouncer.Stop()
        }
        
        // Debounce for 300ms
        s.debouncer = time.AfterFunc(300*time.Millisecond, func() {
            results := performSearch(s.Query)
            s.Results = results
            s.Commit()
        })
    }
}
```

### Batch Updates

Group multiple updates together:

```go
func (c *Component) ProcessBulkUpdate(items []Item) {
    // Batch all updates
    for _, item := range items {
        c.updateItemInternal(item)
    }
    
    // Single commit for all changes
    c.Commit()
}

func (c *Component) updateItemInternal(item Item) {
    // Update without commit
    c.Items[item.ID] = item
}
```

### Lazy Loading

Load data only when needed:

```go
type TabComponent struct {
    *liveview.ComponentDriver[*TabComponent]
    ActiveTab string
    TabData   map[string]interface{}
    loaded    map[string]bool
}

func (t *TabComponent) SwitchTab(data interface{}) {
    if tabName, ok := data.(string); ok {
        t.ActiveTab = tabName
        
        // Load tab data only once
        if !t.loaded[tabName] {
            t.TabData[tabName] = t.loadTabData(tabName)
            t.loaded[tabName] = true
        }
        
        t.Commit()
    }
}
```

### Virtualization for Large Lists

Implement virtual scrolling for large datasets:

```go
type VirtualList struct {
    *liveview.ComponentDriver[*VirtualList]
    AllItems     []Item
    VisibleItems []Item
    ScrollTop    int
    ItemHeight   int
    ViewHeight   int
}

func (v *VirtualList) UpdateScroll(data interface{}) {
    if scrollData, ok := data.(map[string]interface{}); ok {
        v.ScrollTop = int(scrollData["scrollTop"].(float64))
        
        // Calculate visible range
        startIndex := v.ScrollTop / v.ItemHeight
        endIndex := (v.ScrollTop + v.ViewHeight) / v.ItemHeight
        
        // Update only visible items
        if endIndex > len(v.AllItems) {
            endIndex = len(v.AllItems)
        }
        
        v.VisibleItems = v.AllItems[startIndex:endIndex]
        v.Commit()
    }
}
```

## Security Best Practices

### Input Validation

Always validate user input:

```go
func (c *Component) UpdateEmail(data interface{}) {
    email, ok := data.(string)
    if !ok {
        c.SetError("Invalid input type")
        c.Commit()
        return
    }
    
    // Validate email format
    if !isValidEmail(email) {
        c.SetError("Invalid email format")
        c.Commit()
        return
    }
    
    // Sanitize input
    email = strings.TrimSpace(email)
    email = html.EscapeString(email)
    
    c.Email = email
    c.Commit()
}

func isValidEmail(email string) bool {
    pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
    matched, _ := regexp.MatchString(pattern, email)
    return matched
}
```

### SQL Injection Prevention

Use parameterized queries:

✅ **Good:**
```go
func (c *Component) LoadUser(userID int) {
    query := "SELECT * FROM users WHERE id = ?"
    row := c.db.QueryRow(query, userID)
    row.Scan(&c.User)
}
```

❌ **Bad:**
```go
func (c *Component) LoadUser(userID string) {
    // NEVER DO THIS - SQL Injection vulnerability!
    query := fmt.Sprintf("SELECT * FROM users WHERE id = %s", userID)
    row := c.db.QueryRow(query)
}
```

### Rate Limiting

Implement rate limiting for sensitive operations:

```go
type RateLimitedComponent struct {
    *liveview.ComponentDriver[*RateLimitedComponent]
    rateLimiter *rate.Limiter
}

func (r *RateLimitedComponent) Start() {
    // Allow 10 requests per second
    r.rateLimiter = rate.NewLimiter(10, 1)
    r.Commit()
}

func (r *RateLimitedComponent) HandleAction(data interface{}) {
    if !r.rateLimiter.Allow() {
        r.SetError("Too many requests, please slow down")
        r.Commit()
        return
    }
    
    // Process action
    r.processAction(data)
    r.Commit()
}
```

### Authentication and Authorization

Always check permissions:

```go
func (c *Component) DeleteItem(data interface{}) {
    // Check authentication
    if !c.IsAuthenticated() {
        c.SetError("Please login first")
        c.Commit()
        return
    }
    
    // Check authorization
    if !c.HasPermission("delete:items") {
        c.SetError("Insufficient permissions")
        c.Commit()
        return
    }
    
    // Proceed with deletion
    c.performDelete(data)
    r.Commit()
}
```

## Testing Strategies

### Unit Testing Components

Test each component in isolation:

```go
func TestCounter_Increment(t *testing.T) {
    // Arrange
    counter := &Counter{}
    td := liveview.NewTestDriver(t, counter, "test-counter")
    defer td.Cleanup()
    
    // Act
    td.SimulateEvent("Increment", nil)
    
    // Assert
    assert.Equal(t, 1, counter.Count)
    td.AssertHTML(t, "Count: 1")
}
```

### Integration Testing

Test component interactions:

```go
func TestDashboard_Integration(t *testing.T) {
    suite := liveview.NewIntegrationTestSuite(t)
    
    suite.SetupPage("/dashboard", "Dashboard", func() liveview.LiveDriver {
        dashboard := &Dashboard{}
        driver := liveview.NewDriver("dashboard", dashboard)
        dashboard.ComponentDriver = driver
        return driver
    })
    
    suite.Start()
    
    client := suite.NewClient("test-client")
    err := client.Connect(suite.WebSocketURL)
    require.NoError(t, err)
    
    // Test navigation
    err = client.SendEvent("sidebar", "Navigate", "analytics")
    require.NoError(t, err)
    
    // Verify update
    msg, err := client.WaitForMessage("fill", 2*time.Second)
    require.NoError(t, err)
    assert.Contains(t, msg["value"], "Analytics Panel")
}
```

### Benchmark Critical Paths

```go
func BenchmarkDataTable_Render(b *testing.B) {
    table := &DataTable{
        Rows: generateTestData(1000),
    }
    
    liveview.BenchmarkComponentRender(b, table)
}

func BenchmarkDataTable_Sort(b *testing.B) {
    table := &DataTable{
        Rows: generateTestData(1000),
    }
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        table.Sort("name")
    }
}
```

## Error Handling

### Graceful Error Recovery

```go
func (c *Component) LoadData() {
    defer func() {
        if r := recover(); r != nil {
            c.SetError(fmt.Sprintf("An error occurred: %v", r))
            c.Loading = false
            c.Commit()
        }
    }()
    
    c.Loading = true
    c.Commit()
    
    data, err := fetchDataFromAPI()
    if err != nil {
        c.SetError(fmt.Sprintf("Failed to load data: %v", err))
        c.Loading = false
        c.Commit()
        return
    }
    
    c.Data = data
    c.Loading = false
    c.Commit()
}
```

### User-Friendly Error Messages

```go
func (c *Component) getErrorMessage(err error) string {
    switch {
    case errors.Is(err, ErrNotFound):
        return "The requested item was not found"
    case errors.Is(err, ErrUnauthorized):
        return "You don't have permission to perform this action"
    case errors.Is(err, ErrTimeout):
        return "The operation timed out. Please try again"
    default:
        // Log detailed error for debugging
        log.Printf("Error: %v", err)
        // Return generic message to user
        return "An unexpected error occurred. Please try again later"
    }
}
```

## Code Organization

### Project Structure

```
project/
├── cmd/
│   └── server/
│       └── main.go           # Entry point
├── components/
│   ├── common/               # Shared components
│   │   ├── header.go
│   │   ├── footer.go
│   │   └── navigation.go
│   ├── pages/                # Page components
│   │   ├── home.go
│   │   ├── dashboard.go
│   │   └── settings.go
│   └── widgets/               # Widget components
│       ├── chart.go
│       ├── table.go
│       └── form.go
├── models/                    # Data models
│   ├── user.go
│   └── product.go
├── services/                  # Business logic
│   ├── auth.go
│   └── database.go
├── static/                    # Static assets
│   ├── css/
│   ├── js/
│   └── images/
├── templates/                 # External templates
│   └── email/
├── tests/                     # Test files
│   ├── integration/
│   └── unit/
└── go.mod
```

### Naming Conventions

```go
// Components - PascalCase
type UserProfile struct {}
type NavigationBar struct {}

// Methods - PascalCase for exported, camelCase for internal
func (c *Component) HandleClick(data interface{}) {}
func (c *Component) validateInput(input string) bool {}

// Events - PascalCase
func (c *Component) UpdateUser(data interface{}) {}
func (c *Component) DeleteItem(data interface{}) {}

// Template IDs - kebab-case
<div id="user-profile">
<button id="submit-button">
```

## Production Deployment

### Configuration Management

```go
type Config struct {
    Port        string
    DatabaseURL string
    RedisURL    string
    LogLevel    string
    MaxConns    int
}

func LoadConfig() *Config {
    return &Config{
        Port:        getEnv("PORT", "8080"),
        DatabaseURL: getEnv("DATABASE_URL", "postgres://localhost/app"),
        RedisURL:    getEnv("REDIS_URL", "redis://localhost:6379"),
        LogLevel:    getEnv("LOG_LEVEL", "info"),
        MaxConns:    getEnvInt("MAX_CONNS", 100),
    }
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}
```

### Health Checks

```go
func setupHealthChecks(e *echo.Echo) {
    e.GET("/health", func(c echo.Context) error {
        health := checkHealth()
        
        if !health.Healthy {
            return c.JSON(http.StatusServiceUnavailable, health)
        }
        
        return c.JSON(http.StatusOK, health)
    })
}

type HealthStatus struct {
    Healthy  bool                   `json:"healthy"`
    Services map[string]ServiceStatus `json:"services"`
    Version  string                 `json:"version"`
    Uptime   time.Duration          `json:"uptime"`
}

func checkHealth() HealthStatus {
    status := HealthStatus{
        Healthy:  true,
        Services: make(map[string]ServiceStatus),
        Version:  version,
        Uptime:   time.Since(startTime),
    }
    
    // Check database
    if err := db.Ping(); err != nil {
        status.Healthy = false
        status.Services["database"] = ServiceStatus{
            Healthy: false,
            Error:   err.Error(),
        }
    }
    
    // Check Redis
    if err := redis.Ping(); err != nil {
        status.Healthy = false
        status.Services["redis"] = ServiceStatus{
            Healthy: false,
            Error:   err.Error(),
        }
    }
    
    return status
}
```

### Logging

```go
func setupLogging() *zap.Logger {
    config := zap.NewProductionConfig()
    
    if os.Getenv("ENV") == "development" {
        config = zap.NewDevelopmentConfig()
    }
    
    config.Level = zap.NewAtomicLevelAt(getLogLevel())
    
    logger, _ := config.Build()
    return logger
}

func (c *Component) logEvent(event string, data interface{}) {
    logger.Info("Component event",
        zap.String("component", c.GetIDComponent()),
        zap.String("event", event),
        zap.Any("data", data),
        zap.Time("timestamp", time.Now()),
    )
}
```

## Monitoring and Debugging

### Metrics Collection

```go
var (
    requestCounter = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "liveview_requests_total",
            Help: "Total number of LiveView requests",
        },
        []string{"component", "event"},
    )
    
    responseDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "liveview_response_duration_seconds",
            Help: "Response duration in seconds",
        },
        []string{"component"},
    )
)

func (c *Component) trackMetrics(event string, start time.Time) {
    requestCounter.WithLabelValues(c.GetIDComponent(), event).Inc()
    responseDuration.WithLabelValues(c.GetIDComponent()).Observe(time.Since(start).Seconds())
}
```

### Debug Mode

```go
type DebugComponent struct {
    *liveview.ComponentDriver[*DebugComponent]
    DebugMode bool
}

func (d *DebugComponent) HandleEvent(data interface{}) {
    start := time.Now()
    
    if d.DebugMode {
        log.Printf("[DEBUG] Event received: %+v", data)
    }
    
    // Process event
    d.processEvent(data)
    
    if d.DebugMode {
        log.Printf("[DEBUG] Event processed in %v", time.Since(start))
        log.Printf("[DEBUG] State after: %+v", d.getState())
    }
    
    d.Commit()
}
```

## Common Pitfalls

### 1. Forgetting to Call Commit()

❌ **Problem:**
```go
func (c *Component) Update(data interface{}) {
    c.Value = "updated"
    // Forgot c.Commit() - UI won't update!
}
```

✅ **Solution:**
```go
func (c *Component) Update(data interface{}) {
    c.Value = "updated"
    c.Commit() // Always commit changes
}
```

### 2. Memory Leaks with Goroutines

❌ **Problem:**
```go
func (c *Component) Start() {
    go func() {
        for {
            time.Sleep(1 * time.Second)
            c.updateData()
        }
    }()
}
```

✅ **Solution:**
```go
func (c *Component) Start() {
    c.ctx, c.cancel = context.WithCancel(context.Background())
    
    go func() {
        ticker := time.NewTicker(1 * time.Second)
        defer ticker.Stop()
        
        for {
            select {
            case <-c.ctx.Done():
                return
            case <-ticker.C:
                c.updateData()
            }
        }
    }()
}

func (c *Component) Stop() {
    if c.cancel != nil {
        c.cancel()
    }
}
```

### 3. Race Conditions

❌ **Problem:**
```go
func (c *Component) ConcurrentUpdate(data interface{}) {
    go func() {
        c.Counter++ // Race condition!
        c.Commit()
    }()
}
```

✅ **Solution:**
```go
func (c *Component) ConcurrentUpdate(data interface{}) {
    go func() {
        c.mu.Lock()
        c.Counter++
        c.mu.Unlock()
        c.Commit()
    }()
}
```

### 4. Inefficient Re-rendering

❌ **Problem:**
```go
func (c *Component) UpdateItems(items []Item) {
    for _, item := range items {
        c.Items = append(c.Items, item)
        c.Commit() // Commits on every iteration!
    }
}
```

✅ **Solution:**
```go
func (c *Component) UpdateItems(items []Item) {
    c.Items = append(c.Items, items...)
    c.Commit() // Single commit
}
```

---

## Español

# Mejores Prácticas para el Desarrollo con Go Echo LiveView

## Tabla de Contenidos
1. [Diseño de Componentes](#diseño-de-componentes)
2. [Gestión de Estado](#gestión-de-estado)
3. [Optimización de Rendimiento](#optimización-de-rendimiento)
4. [Mejores Prácticas de Seguridad](#mejores-prácticas-de-seguridad)
5. [Estrategias de Testing](#estrategias-de-testing)
6. [Manejo de Errores](#manejo-de-errores)
7. [Organización del Código](#organización-del-código)
8. [Despliegue en Producción](#despliegue-en-producción)
9. [Monitoreo y Depuración](#monitoreo-y-depuración)
10. [Errores Comunes](#errores-comunes)

## Diseño de Componentes

### Principio de Responsabilidad Única

Cada componente debe tener un propósito único y bien definido:

✅ **Bueno:**
```go
// PerfilUsuario maneja solo la visualización del perfil
type PerfilUsuario struct {
    *liveview.ComponentDriver[*PerfilUsuario]
    Usuario Usuario
}

// ConfiguracionUsuario maneja solo la configuración
type ConfiguracionUsuario struct {
    *liveview.ComponentDriver[*ConfiguracionUsuario]
    Configuracion Configuracion
}
```

❌ **Malo:**
```go
// ComponenteUsuario hace demasiado
type ComponenteUsuario struct {
    *liveview.ComponentDriver[*ComponenteUsuario]
    Usuario       Usuario
    Configuracion Configuracion
    Pedidos       []Pedido
    Mensajes      []Mensaje
}
```

### Composición de Componentes

Usa composición para UIs complejas:

```go
type Tablero struct {
    *liveview.ComponentDriver[*Tablero]
    Cabecera     *Cabecera
    BarraLateral *BarraLateral
    Analiticas   *PanelAnaliticas
    Actividad    *FeedActividad
}

func (t *Tablero) Start() {
    // Inicializar componentes hijos
    t.Cabecera = &Cabecera{}
    t.BarraLateral = &BarraLateral{}
    t.Analiticas = &PanelAnaliticas{}
    t.Actividad = &FeedActividad{}
    
    // Montar componentes
    t.Mount(liveview.NewDriver("cabecera", t.Cabecera))
    t.Mount(liveview.NewDriver("barra-lateral", t.BarraLateral))
    t.Mount(liveview.NewDriver("analiticas", t.Analiticas))
    t.Mount(liveview.NewDriver("actividad", t.Actividad))
    
    t.Commit()
}
```

## Gestión de Estado

### Inicializar Estado Correctamente

Siempre inicializa el estado en el método `Start()`:

```go
func (c *Componente) Start() {
    // Inicializar colecciones
    c.Items = make([]Item, 0)
    c.Usuarios = make(map[string]Usuario)
    
    // Establecer valores por defecto
    c.PaginaActual = 1
    c.TamañoPagina = 20
    c.OrdenSort = "asc"
    
    // Cargar datos iniciales
    c.CargarDatos()
    
    // Siempre hacer commit después de inicializar
    c.Commit()
}
```

### Usar Contextos para Cancelación

Gestiona operaciones de larga duración con contextos:

```go
type CargadorDatos struct {
    *liveview.ComponentDriver[*CargadorDatos]
    ctx    context.Context
    cancel context.CancelFunc
    Datos  []PuntoDatos
}

func (c *CargadorDatos) Start() {
    c.ctx, c.cancel = context.WithCancel(context.Background())
    c.CargarDatosAsync()
    c.Commit()
}

func (c *CargadorDatos) CargarDatosAsync() {
    go func() {
        select {
        case <-c.ctx.Done():
            return
        case datos := <-obtenerDatos():
            c.Datos = datos
            c.Commit()
        }
    }()
}

func (c *CargadorDatos) Stop() {
    if c.cancel != nil {
        c.cancel()
    }
}
```

## Optimización de Rendimiento

### Debounce para Actualizaciones Frecuentes

Para cajas de búsqueda y filtros en tiempo real:

```go
type ComponenteBusqueda struct {
    *liveview.ComponentDriver[*ComponenteBusqueda]
    Consulta    string
    Resultados  []Resultado
    debouncer   *time.Timer
    mu          sync.Mutex
}

func (b *ComponenteBusqueda) Buscar(data interface{}) {
    b.mu.Lock()
    defer b.mu.Unlock()
    
    if consulta, ok := data.(string); ok {
        b.Consulta = consulta
        
        // Cancelar búsqueda anterior
        if b.debouncer != nil {
            b.debouncer.Stop()
        }
        
        // Debounce por 300ms
        b.debouncer = time.AfterFunc(300*time.Millisecond, func() {
            resultados := realizarBusqueda(b.Consulta)
            b.Resultados = resultados
            b.Commit()
        })
    }
}
```

### Actualización por Lotes

Agrupa múltiples actualizaciones:

```go
func (c *Componente) ProcesarActualizacionMasiva(items []Item) {
    // Agrupar todas las actualizaciones
    for _, item := range items {
        c.actualizarItemInterno(item)
    }
    
    // Un solo commit para todos los cambios
    c.Commit()
}

func (c *Componente) actualizarItemInterno(item Item) {
    // Actualizar sin commit
    c.Items[item.ID] = item
}
```

## Mejores Prácticas de Seguridad

### Validación de Entrada

Siempre valida la entrada del usuario:

```go
func (c *Componente) ActualizarEmail(data interface{}) {
    email, ok := data.(string)
    if !ok {
        c.SetError("Tipo de entrada inválido")
        c.Commit()
        return
    }
    
    // Validar formato de email
    if !esEmailValido(email) {
        c.SetError("Formato de email inválido")
        c.Commit()
        return
    }
    
    // Sanitizar entrada
    email = strings.TrimSpace(email)
    email = html.EscapeString(email)
    
    c.Email = email
    c.Commit()
}

func esEmailValido(email string) bool {
    patron := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
    coincide, _ := regexp.MatchString(patron, email)
    return coincide
}
```

### Prevención de Inyección SQL

Usa consultas parametrizadas:

✅ **Bueno:**
```go
func (c *Componente) CargarUsuario(idUsuario int) {
    consulta := "SELECT * FROM usuarios WHERE id = ?"
    fila := c.db.QueryRow(consulta, idUsuario)
    fila.Scan(&c.Usuario)
}
```

❌ **Malo:**
```go
func (c *Componente) CargarUsuario(idUsuario string) {
    // NUNCA HAGAS ESTO - ¡Vulnerabilidad de inyección SQL!
    consulta := fmt.Sprintf("SELECT * FROM usuarios WHERE id = %s", idUsuario)
    fila := c.db.QueryRow(consulta)
}
```

## Estrategias de Testing

### Testing Unitario de Componentes

Prueba cada componente aisladamente:

```go
func TestContador_Incrementar(t *testing.T) {
    // Preparar
    contador := &Contador{}
    td := liveview.NewTestDriver(t, contador, "test-contador")
    defer td.Cleanup()
    
    // Actuar
    td.SimulateEvent("Incrementar", nil)
    
    // Afirmar
    assert.Equal(t, 1, contador.Cuenta)
    td.AssertHTML(t, "Cuenta: 1")
}
```

### Testing de Integración

Prueba las interacciones entre componentes:

```go
func TestTablero_Integracion(t *testing.T) {
    suite := liveview.NewIntegrationTestSuite(t)
    
    suite.SetupPage("/tablero", "Tablero", func() liveview.LiveDriver {
        tablero := &Tablero{}
        driver := liveview.NewDriver("tablero", tablero)
        tablero.ComponentDriver = driver
        return driver
    })
    
    suite.Start()
    
    cliente := suite.NewClient("test-cliente")
    err := cliente.Connect(suite.WebSocketURL)
    require.NoError(t, err)
    
    // Probar navegación
    err = cliente.SendEvent("barra-lateral", "Navegar", "analiticas")
    require.NoError(t, err)
    
    // Verificar actualización
    msg, err := cliente.WaitForMessage("fill", 2*time.Second)
    require.NoError(t, err)
    assert.Contains(t, msg["value"], "Panel de Analíticas")
}
```

## Manejo de Errores

### Recuperación Elegante de Errores

```go
func (c *Componente) CargarDatos() {
    defer func() {
        if r := recover(); r != nil {
            c.SetError(fmt.Sprintf("Ocurrió un error: %v", r))
            c.Cargando = false
            c.Commit()
        }
    }()
    
    c.Cargando = true
    c.Commit()
    
    datos, err := obtenerDatosDeAPI()
    if err != nil {
        c.SetError(fmt.Sprintf("Error al cargar datos: %v", err))
        c.Cargando = false
        c.Commit()
        return
    }
    
    c.Datos = datos
    c.Cargando = false
    c.Commit()
}
```

### Mensajes de Error Amigables

```go
func (c *Componente) obtenerMensajeError(err error) string {
    switch {
    case errors.Is(err, ErrNoEncontrado):
        return "El elemento solicitado no fue encontrado"
    case errors.Is(err, ErrNoAutorizado):
        return "No tienes permiso para realizar esta acción"
    case errors.Is(err, ErrTimeout):
        return "La operación tardó demasiado. Por favor intenta de nuevo"
    default:
        // Registrar error detallado para depuración
        log.Printf("Error: %v", err)
        // Retornar mensaje genérico al usuario
        return "Ocurrió un error inesperado. Por favor intenta más tarde"
    }
}
```

## Organización del Código

### Estructura del Proyecto

```
proyecto/
├── cmd/
│   └── servidor/
│       └── main.go           # Punto de entrada
├── componentes/
│   ├── comunes/              # Componentes compartidos
│   │   ├── cabecera.go
│   │   ├── pie.go
│   │   └── navegacion.go
│   ├── paginas/              # Componentes de página
│   │   ├── inicio.go
│   │   ├── tablero.go
│   │   └── configuracion.go
│   └── widgets/              # Componentes widget
│       ├── grafico.go
│       ├── tabla.go
│       └── formulario.go
├── modelos/                  # Modelos de datos
│   ├── usuario.go
│   └── producto.go
├── servicios/                # Lógica de negocio
│   ├── auth.go
│   └── database.go
├── estaticos/                # Archivos estáticos
│   ├── css/
│   ├── js/
│   └── imagenes/
├── plantillas/               # Plantillas externas
│   └── email/
├── pruebas/                  # Archivos de prueba
│   ├── integracion/
│   └── unitarias/
└── go.mod
```

## Despliegue en Producción

### Gestión de Configuración

```go
type Config struct {
    Puerto      string
    URLDatabase string
    URLRedis    string
    NivelLog    string
    MaxConex    int
}

func CargarConfig() *Config {
    return &Config{
        Puerto:      obtenerEnv("PUERTO", "8080"),
        URLDatabase: obtenerEnv("URL_DATABASE", "postgres://localhost/app"),
        URLRedis:    obtenerEnv("URL_REDIS", "redis://localhost:6379"),
        NivelLog:    obtenerEnv("NIVEL_LOG", "info"),
        MaxConex:    obtenerEnvInt("MAX_CONEX", 100),
    }
}

func obtenerEnv(clave, valorPorDefecto string) string {
    if valor := os.Getenv(clave); valor != "" {
        return valor
    }
    return valorPorDefecto
}
```

## Monitoreo y Depuración

### Recolección de Métricas

```go
var (
    contadorPeticiones = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "liveview_peticiones_total",
            Help: "Número total de peticiones LiveView",
        },
        []string{"componente", "evento"},
    )
    
    duracionRespuesta = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "liveview_duracion_respuesta_segundos",
            Help: "Duración de respuesta en segundos",
        },
        []string{"componente"},
    )
)

func (c *Componente) rastrearMetricas(evento string, inicio time.Time) {
    contadorPeticiones.WithLabelValues(c.GetIDComponent(), evento).Inc()
    duracionRespuesta.WithLabelValues(c.GetIDComponent()).Observe(time.Since(inicio).Seconds())
}
```

## Errores Comunes

### 1. Olvidar Llamar Commit()

❌ **Problema:**
```go
func (c *Componente) Actualizar(data interface{}) {
    c.Valor = "actualizado"
    // ¡Olvidó c.Commit() - La UI no se actualizará!
}
```

✅ **Solución:**
```go
func (c *Componente) Actualizar(data interface{}) {
    c.Valor = "actualizado"
    c.Commit() // Siempre confirmar cambios
}
```

### 2. Fugas de Memoria con Goroutines

❌ **Problema:**
```go
func (c *Componente) Start() {
    go func() {
        for {
            time.Sleep(1 * time.Second)
            c.actualizarDatos()
        }
    }()
}
```

✅ **Solución:**
```go
func (c *Componente) Start() {
    c.ctx, c.cancel = context.WithCancel(context.Background())
    
    go func() {
        ticker := time.NewTicker(1 * time.Second)
        defer ticker.Stop()
        
        for {
            select {
            case <-c.ctx.Done():
                return
            case <-ticker.C:
                c.actualizarDatos()
            }
        }
    }()
}

func (c *Componente) Stop() {
    if c.cancel != nil {
        c.cancel()
    }
}
```

### 3. Condiciones de Carrera

❌ **Problema:**
```go
func (c *Componente) ActualizacionConcurrente(data interface{}) {
    go func() {
        c.Contador++ // ¡Condición de carrera!
        c.Commit()
    }()
}
```

✅ **Solución:**
```go
func (c *Componente) ActualizacionConcurrente(data interface{}) {
    go func() {
        c.mu.Lock()
        c.Contador++
        c.mu.Unlock()
        c.Commit()
    }()
}
```

### 4. Re-renderizado Ineficiente

❌ **Problema:**
```go
func (c *Componente) ActualizarItems(items []Item) {
    for _, item := range items {
        c.Items = append(c.Items, item)
        c.Commit() // ¡Commit en cada iteración!
    }
}
```

✅ **Solución:**
```go
func (c *Componente) ActualizarItems(items []Item) {
    c.Items = append(c.Items, items...)
    c.Commit() // Un solo commit
}
```

---

## Summary / Resumen

Following these best practices will help you build robust, performant, and maintainable LiveView applications. Remember to always prioritize security, test thoroughly, and optimize for user experience.

Siguiendo estas mejores prácticas te ayudará a construir aplicaciones LiveView robustas, eficientes y mantenibles. Recuerda siempre priorizar la seguridad, probar exhaustivamente y optimizar para la experiencia del usuario.