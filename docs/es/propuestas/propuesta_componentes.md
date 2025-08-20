# Propuesta de Biblioteca de Componentes - Go Echo LiveView

## Visión General

Esta propuesta describe una biblioteca completa de componentes reutilizables para Go Echo LiveView, diseñada para acelerar el desarrollo y mantener consistencia en las aplicaciones. Los componentes seguirán principios de diseño modernos y serán totalmente personalizables.

## 1. Componentes de Formulario

### 1.1 Input Avanzado

```go
type Input struct {
    liveview.LiveViewComponentWrapper[Input]
    
    // Propiedades
    Type        InputType    // text, email, password, number, etc.
    Value       string
    Placeholder string
    Label       string
    Error       string
    Required    bool
    Disabled    bool
    Icon        string
    Mask        string       // Para formato de entrada
    Validator   Validator    // Validación en tiempo real
    
    // Eventos
    OnChange    func(string)
    OnFocus     func()
    OnBlur      func()
    OnValidate  func(string) error
}

// Uso
input := &Input{
    Type:        InputEmail,
    Label:       "Correo Electrónico",
    Placeholder: "usuario@ejemplo.com",
    Required:    true,
    Validator:   EmailValidator{},
}
```

### 1.2 Select Mejorado

```go
type Select struct {
    liveview.LiveViewComponentWrapper[Select]
    
    Options      []Option
    Selected     []string      // Soporte multi-select
    Searchable   bool         // Búsqueda en opciones
    Async        bool         // Carga async de opciones
    GroupBy      string       // Agrupar opciones
    Multiple     bool
    Placeholder  string
    
    // Personalización
    OptionTemplate string     // Template personalizado para opciones
    ChipTemplate   string     // Para multi-select
}

type Option struct {
    Value    string
    Label    string
    Icon     string
    Disabled bool
    Group    string
    Data     map[string]interface{} // Datos adicionales
}
```

### 1.3 DatePicker

```go
type DatePicker struct {
    liveview.LiveViewComponentWrapper[DatePicker]
    
    Value        time.Time
    Mode         DatePickerMode  // date, time, datetime, range
    Format       string
    MinDate      time.Time
    MaxDate      time.Time
    DisabledDates []time.Time
    Locale       string
    FirstDayOfWeek int
    
    // Características avanzadas
    ShowWeekNumbers bool
    Shortcuts      []DateShortcut  // "Hoy", "Ayer", "Última semana"
    TimezoneSelect bool
}
```

### 1.4 FileUpload

```go
type FileUpload struct {
    liveview.LiveViewComponentWrapper[FileUpload]
    
    Accept       []string      // Tipos MIME permitidos
    Multiple     bool
    MaxSize      int64         // Bytes
    MaxFiles     int
    DragDrop     bool
    
    // Estado
    Files        []UploadedFile
    Progress     map[string]int // Progreso por archivo
    
    // Características
    Preview      bool          // Vista previa de imágenes
    Compression  bool          // Comprimir imágenes
    ChunkSize    int           // Para uploads grandes
    
    // Eventos
    OnUpload     func([]UploadedFile)
    OnProgress   func(string, int)
    OnError      func(error)
}
```

### 1.5 FormBuilder Dinámico

```go
type FormBuilder struct {
    liveview.LiveViewComponentWrapper[FormBuilder]
    
    Schema      FormSchema
    Values      map[string]interface{}
    Errors      map[string]string
    Layout      FormLayout
    
    // Validación
    Validators  map[string]Validator
    AsyncValidation bool
    
    // Características
    AutoSave    bool
    SaveDraft   bool
    Wizard      bool  // Formulario multi-paso
}

type FormField struct {
    Name        string
    Type        string
    Label       string
    Required    bool
    Conditions  []Condition  // Mostrar/ocultar condicionalmente
    Validators  []Validator
}
```

## 2. Componentes de Visualización de Datos

### 2.1 DataTable Avanzado

```go
type DataTable struct {
    liveview.LiveViewComponentWrapper[DataTable]
    
    // Datos
    Columns     []Column
    Data        []map[string]interface{}
    
    // Características
    Sortable    bool
    Filterable  bool
    Paginated   bool
    Selectable  bool
    Expandable  bool
    Editable    bool
    
    // Estado
    SortBy      string
    SortOrder   SortOrder
    Filters     map[string]Filter
    Page        int
    PageSize    int
    Selected    []string
    
    // Personalización
    RowTemplate string
    EmptyState  Component
    
    // Acciones
    Actions     []TableAction
    BulkActions []BulkAction
}

type Column struct {
    Key         string
    Title       string
    Type        ColumnType    // text, number, date, boolean, custom
    Sortable    bool
    Filterable  bool
    Width       string
    Align       Alignment
    Format      Formatter
    Editable    bool
    Component   Component     // Para celdas custom
}
```

### 2.2 Chart (Gráficos)

```go
type Chart struct {
    liveview.LiveViewComponentWrapper[Chart]
    
    Type        ChartType     // line, bar, pie, donut, area, scatter
    Data        ChartData
    Options     ChartOptions
    
    // Interactividad
    Interactive bool
    Tooltips    bool
    Zoom        bool
    Pan         bool
    Export      bool          // Exportar como imagen/PDF
    
    // Actualizaciones en tiempo real
    RealTime    bool
    UpdateInterval time.Duration
}

type ChartData struct {
    Labels   []string
    Datasets []Dataset
}

type Dataset struct {
    Label           string
    Data            []float64
    BackgroundColor string
    BorderColor     string
    Type            string    // Para gráficos mixtos
}
```

### 2.3 Card

```go
type Card struct {
    liveview.LiveViewComponentWrapper[Card]
    
    // Contenido
    Title       string
    Subtitle    string
    Content     Component
    Image       string
    
    // Características
    Hoverable   bool
    Clickable   bool
    Draggable   bool
    Collapsible bool
    
    // Slots
    Header      Component
    Footer      Component
    Actions     []Action
    
    // Estado
    Collapsed   bool
    Selected    bool
}
```

### 2.4 Timeline

```go
type Timeline struct {
    liveview.LiveViewComponentWrapper[Timeline]
    
    Items       []TimelineItem
    Orientation TimelineOrientation  // vertical, horizontal
    Mode        TimelineMode        // left, right, alternate
    
    // Características
    Connectors  bool
    Animations  bool
    Interactive bool
}

type TimelineItem struct {
    Time        time.Time
    Title       string
    Content     string
    Icon        string
    Color       string
    Component   Component    // Contenido personalizado
    Clickable   bool
}
```

## 3. Componentes de Navegación

### 3.1 Menu Avanzado

```go
type Menu struct {
    liveview.LiveViewComponentWrapper[Menu]
    
    Items       []MenuItem
    Mode        MenuMode     // horizontal, vertical, inline
    Theme       MenuTheme
    
    // Estado
    ActiveKey   string
    OpenKeys    []string     // Para submenús
    Collapsed   bool         // Para menú lateral
    
    // Características
    Accordion   bool         // Un solo submenú abierto
    Searchable  bool         // Búsqueda en items
    Responsive  bool         // Cambiar a móvil automáticamente
}

type MenuItem struct {
    Key         string
    Label       string
    Icon        string
    Link        string
    Children    []MenuItem
    Disabled    bool
    Badge       string       // Notificación/contador
    Component   Component    // Item personalizado
}
```

### 3.2 Breadcrumb

```go
type Breadcrumb struct {
    liveview.LiveViewComponentWrapper[Breadcrumb]
    
    Items       []BreadcrumbItem
    Separator   string
    MaxItems    int          // Colapsar items intermedios
    
    // Características
    Responsive  bool         // Adaptar a móvil
    HomeIcon    string
}
```

### 3.3 Tabs

```go
type Tabs struct {
    liveview.LiveViewComponentWrapper[Tabs]
    
    Tabs        []Tab
    ActiveKey   string
    Type        TabType      // line, card, button
    Position    TabPosition  // top, bottom, left, right
    
    // Características
    Animated    bool
    Closeable   bool         // Permitir cerrar tabs
    Draggable   bool         // Reordenar tabs
    LazyLoad    bool         // Cargar contenido solo cuando se activa
    
    // Eventos
    OnChange    func(string)
    OnClose     func(string)
    OnReorder   func([]string)
}
```

### 3.4 Stepper

```go
type Stepper struct {
    liveview.LiveViewComponentWrapper[Stepper]
    
    Steps       []Step
    Current     int
    Direction   StepperDirection  // horizontal, vertical
    
    // Características
    Linear      bool         // Forzar orden secuencial
    Clickable   bool         // Permitir click en pasos
    ShowError   bool         // Mostrar pasos con error
    
    // Validación
    Validator   StepValidator
}

type Step struct {
    Title       string
    Subtitle    string
    Icon        string
    Content     Component
    Status      StepStatus   // wait, process, finish, error
    Disabled    bool
}
```

## 4. Componentes de Feedback

### 4.1 Modal Mejorado

```go
type Modal struct {
    liveview.LiveViewComponentWrapper[Modal]
    
    // Contenido
    Title       string
    Content     Component
    Footer      Component
    
    // Configuración
    Size        ModalSize    // small, medium, large, fullscreen
    Closeable   bool
    Keyboard    bool         // Cerrar con ESC
    Backdrop    bool         // Click fuera para cerrar
    Centered    bool
    
    // Animación
    Animation   ModalAnimation
    Duration    time.Duration
    
    // Estado
    Visible     bool
    Loading     bool
}
```

### 4.2 Notification

```go
type Notification struct {
    liveview.LiveViewComponentWrapper[Notification]
    
    Type        NotificationType  // success, error, warning, info
    Title       string
    Message     string
    Duration    time.Duration     // 0 para manual
    Position    NotificationPosition
    
    // Características
    Closeable   bool
    Icon        string
    Actions     []Action
    Progress    bool             // Barra de progreso
}

// Sistema global de notificaciones
type NotificationSystem struct {
    Stack       []Notification
    MaxStack    int
    Position    NotificationPosition
}
```

### 4.3 Progress

```go
type Progress struct {
    liveview.LiveViewComponentWrapper[Progress]
    
    Type        ProgressType     // line, circle, dashboard
    Percent     int
    Status      ProgressStatus   // active, exception, success
    
    // Personalización
    StrokeColor string
    TrailColor  string
    Width       int              // Para circle/dashboard
    
    // Características
    ShowInfo    bool
    Format      func(int) string // Formatear texto
    Animated    bool
}
```

### 4.4 Skeleton

```go
type Skeleton struct {
    liveview.LiveViewComponentWrapper[Skeleton]
    
    Type        SkeletonType    // text, avatar, button, input, image
    Active      bool            // Animación
    Loading     bool
    
    // Configuración
    Rows        int             // Para tipo paragraph
    Width       string
    Height      string
    Shape       SkeletonShape   // default, circle, square
}
```

## 5. Componentes de Layout

### 5.1 Grid System

```go
type Grid struct {
    liveview.LiveViewComponentWrapper[Grid]
    
    Columns     int
    Gap         string
    Responsive  map[string]int   // Breakpoints
    
    // Características
    AutoFlow    GridAutoFlow
    AlignItems  Alignment
    JustifyContent Justification
}

type GridItem struct {
    Span        int
    Offset      int
    Order       int
    Responsive  map[string]GridItemConfig
}
```

### 5.2 Split Pane

```go
type SplitPane struct {
    liveview.LiveViewComponentWrapper[SplitPane]
    
    Direction   SplitDirection   // horizontal, vertical
    Sizes       []float64        // Tamaños iniciales
    MinSizes    []float64
    MaxSizes    []float64
    
    // Características
    Resizable   bool
    Collapsible []bool           // Por panel
    OnResize    func([]float64)
}
```

### 5.3 Drawer

```go
type Drawer struct {
    liveview.LiveViewComponentWrapper[Drawer]
    
    Placement   DrawerPlacement  // left, right, top, bottom
    Width       string
    Height      string
    
    // Estado
    Visible     bool
    
    // Características
    Mask        bool
    Keyboard    bool
    Push        bool             // Empujar contenido
}
```

## 6. Componentes Especializados

### 6.1 Calendar

```go
type Calendar struct {
    liveview.LiveViewComponentWrapper[Calendar]
    
    View        CalendarView     // month, week, day, year
    Date        time.Time
    Events      []CalendarEvent
    
    // Características
    Draggable   bool            // Arrastrar eventos
    Editable    bool
    ShowWeekends bool
    FirstDay    int
    
    // Eventos
    OnEventClick func(CalendarEvent)
    OnDateClick  func(time.Time)
    OnEventDrop  func(CalendarEvent, time.Time)
}

type CalendarEvent struct {
    ID          string
    Title       string
    Start       time.Time
    End         time.Time
    AllDay      bool
    Color       string
    Recurring   RecurrenceRule
}
```

### 6.2 Editor de Texto Rico

```go
type RichTextEditor struct {
    liveview.LiveViewComponentWrapper[RichTextEditor]
    
    Value       string          // HTML o Markdown
    Format      EditorFormat    // html, markdown
    
    // Toolbar
    Tools       []EditorTool
    CustomTools []CustomTool
    
    // Características
    AutoSave    bool
    Mentions    []Mention       // @menciones
    Hashtags    bool
    CodeHighlight bool
    ImageUpload bool
    
    // Límites
    MaxLength   int
    MaxImages   int
}
```

### 6.3 TreeView

```go
type TreeView struct {
    liveview.LiveViewComponentWrapper[TreeView]
    
    Data        []TreeNode
    
    // Características
    Checkable   bool
    Draggable   bool
    Searchable  bool
    LazyLoad    bool
    ShowLine    bool
    ShowIcon    bool
    
    // Estado
    ExpandedKeys []string
    CheckedKeys  []string
    SelectedKeys []string
    
    // Eventos
    OnExpand    func([]string)
    OnCheck     func([]string)
    OnSelect    func([]string)
    OnDrop      func(DragInfo)
}
```

### 6.4 Kanban Board

```go
type KanbanBoard struct {
    liveview.LiveViewComponentWrapper[KanbanBoard]
    
    Columns     []KanbanColumn
    
    // Características
    Draggable   bool
    Sortable    bool
    Collapsible bool
    AddCard     bool
    AddColumn   bool
    
    // Límites
    MaxCards    map[string]int  // Por columna
    
    // Eventos
    OnCardMove  func(card, fromCol, toCol string)
    OnCardAdd   func(card, column string)
    OnColumnAdd func(column string)
}

type KanbanColumn struct {
    ID          string
    Title       string
    Cards       []KanbanCard
    Color       string
    Collapsed   bool
}
```

## 7. Temas y Personalización

### 7.1 Sistema de Temas

```go
type Theme struct {
    Name        string
    Colors      ColorPalette
    Typography  Typography
    Spacing     SpacingScale
    Borders     BorderConfig
    Shadows     ShadowScale
    Animations  AnimationConfig
}

// Aplicar tema globalmente
theme := &Theme{
    Name: "custom",
    Colors: ColorPalette{
        Primary:   "#1890ff",
        Secondary: "#52c41a",
        Error:     "#f5222d",
        Warning:   "#faad14",
        Info:      "#1890ff",
    },
}

app.SetTheme(theme)
```

### 7.2 Variantes de Componentes

```go
// Registrar variantes personalizadas
RegisterVariant("button", "gradient", ButtonStyle{
    Background: "linear-gradient(45deg, #FE6B8B 30%, #FF8E53 90%)",
    Border:     "0",
    Color:      "white",
    Shadow:     "0 3px 5px 2px rgba(255, 105, 135, .3)",
})

// Usar variante
button := &Button{
    Variant: "gradient",
    Text:    "Click me",
}
```

## 8. Utilidades y Helpers

### 8.1 Validadores

```go
var CommonValidators = struct {
    Required    Validator
    Email       Validator
    URL         Validator
    MinLength   func(int) Validator
    MaxLength   func(int) Validator
    Pattern     func(string) Validator
    Custom      func(ValidatorFunc) Validator
}{}

// Validador compuesto
validator := CombineValidators(
    CommonValidators.Required,
    CommonValidators.Email,
    CommonValidators.Custom(func(value string) error {
        if strings.Contains(value, "example.com") {
            return errors.New("No se permiten emails de ejemplo")
        }
        return nil
    }),
)
```

### 8.2 Formatters

```go
var Formatters = struct {
    Date        func(time.Time, string) string
    Number      func(float64, int) string
    Currency    func(float64, string) string
    Percentage  func(float64) string
    FileSize    func(int64) string
    Duration    func(time.Duration) string
}{}
```

### 8.3 Animaciones

```go
type Animation struct {
    Type        AnimationType
    Duration    time.Duration
    Delay       time.Duration
    Easing      EasingFunction
}

// Animaciones predefinidas
var Animations = struct {
    FadeIn      Animation
    FadeOut     Animation
    SlideIn     Animation
    SlideOut    Animation
    Bounce      Animation
    Shake       Animation
}{}
```

## 9. Integración y Uso

### 9.1 Instalación

```bash
go get github.com/go-echo-live-view/components
```

### 9.2 Ejemplo de Uso Completo

```go
package main

import (
    "github.com/arturoeanton/go-echo-live-view/liveview"
    "github.com/go-echo-live-view/components"
)

type Dashboard struct {
    liveview.LiveViewComponentWrapper[Dashboard]
    
    dataTable   *components.DataTable
    chart       *components.Chart
    dateFilter  *components.DatePicker
}

func (d *Dashboard) Start() {
    // Configurar tabla
    d.dataTable = &components.DataTable{
        Columns: []components.Column{
            {Key: "name", Title: "Nombre", Sortable: true},
            {Key: "sales", Title: "Ventas", Type: components.ColumnNumber},
            {Key: "date", Title: "Fecha", Type: components.ColumnDate},
        },
        Sortable:   true,
        Filterable: true,
        Paginated:  true,
    }
    
    // Configurar gráfico
    d.chart = &components.Chart{
        Type: components.ChartLine,
        Data: d.loadChartData(),
        Options: components.ChartOptions{
            Responsive: true,
            Animations: true,
        },
    }
    
    // Montar componentes
    d.Mount("data-table", d.dataTable)
    d.Mount("sales-chart", d.chart)
}

func (d *Dashboard) GetTemplate() string {
    return `
    <div class="dashboard">
        <h1>Dashboard de Ventas</h1>
        
        <div class="filters">
            {{mount "date-filter"}}
        </div>
        
        <div class="charts">
            {{mount "sales-chart"}}
        </div>
        
        <div class="data">
            {{mount "data-table"}}
        </div>
    </div>
    `
}
```

## 10. Roadmap de Desarrollo

### Fase 1: Componentes Core (0-2 meses)
- Input, Button, Select
- Card, Modal, Notification
- Grid, Layout básico

### Fase 2: Componentes de Datos (2-4 meses)
- DataTable
- Charts básicos
- Forms dinámicos

### Fase 3: Componentes Avanzados (4-6 meses)
- Calendar
- TreeView
- Kanban Board
- Editor Rico

### Fase 4: Temas y Personalización (6-8 meses)
- Sistema de temas completo
- Variantes personalizadas
- Constructor de temas visual

## Conclusión

Esta biblioteca de componentes proporcionará una base sólida para el desarrollo rápido de aplicaciones con Go Echo LiveView, ofreciendo componentes modernos, accesibles y altamente personalizables que cubren la mayoría de necesidades de desarrollo web.