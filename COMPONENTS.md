# Go Echo LiveView Components

This document describes all available components in the Go Echo LiveView framework.

## Core Components

### 1. Form Validation Component (`components/form.go`)
A comprehensive form component with built-in validation rules.

**Features:**
- Field validation (required, email, phone, regex)
- Real-time validation feedback
- Custom validation rules
- Error display
- Form submission handling

**Usage:**
```go
form := &components.Form{
    Fields: []components.FormField{
        {
            Name: "email",
            Label: "Email",
            Type: "email",
            Required: true,
            Rules: []components.ValidationRule{
                {Type: "email", Message: "Invalid email"},
            },
        },
    },
    OnSubmit: func(data map[string]string) error {
        // Handle form submission
        return nil
    },
}
```

### 2. File Upload Component (`components/fileupload.go`)
Drag-and-drop file upload with preview support.

**Features:**
- Drag & drop interface
- Multiple file support
- File type filtering
- Size limits
- Image preview
- Upload progress

**Usage:**
```go
upload := &components.FileUpload{
    Multiple: true,
    Accept: "image/*,.pdf",
    MaxSize: 5 * 1024 * 1024, // 5MB
    OnUpload: func(files []components.FileInfo) error {
        // Handle file upload
        return nil
    },
}
```

### 3. Table/DataGrid Component (`components/table.go`)
Feature-rich data table with sorting, filtering, and pagination.

**Features:**
- Column sorting
- Text filtering
- Pagination
- Row selection
- Custom column widths
- Click events

**Usage:**
```go
table := &components.Table{
    Columns: []components.Column{
        {Key: "id", Title: "ID", Sortable: true},
        {Key: "name", Title: "Name", Sortable: true},
    },
    Rows: []components.Row{
        {"id": 1, "name": "John Doe"},
    },
    PageSize: 10,
    ShowPagination: true,
}
```

### 4. Modal/Dialog Component (`components/modal.go`)
Customizable modal dialog.

**Features:**
- Multiple sizes (small, medium, large, full)
- Closable overlay
- Custom content
- Footer buttons
- Animation effects

**Usage:**
```go
modal := &components.Modal{
    Title: "Confirm Action",
    Content: "Are you sure?",
    Size: "medium",
    ShowFooter: true,
    OnOk: func() {
        // Handle OK
    },
}
```

### 5. Notification System (`components/notification.go`)
Toast-style notification system.

**Features:**
- Multiple types (success, error, warning, info)
- Auto-dismiss with timer
- Custom positioning
- Progress indicator
- Stack management

**Usage:**
```go
notifications := &components.NotificationSystem{
    Position: "top-right",
    MaxVisible: 5,
}
// Show notification
notifications.Success("Title", "Message")
notifications.Error("Error", "Something went wrong")
```

## Advanced Components

### 6. Chart Component (`components/chart.go`)
Interactive charts and visualizations.

**Features:**
- Bar charts
- Pie charts
- Interactive tooltips
- Custom colors
- Legends

**Usage:**
```go
chart := &components.Chart{
    Type: components.ChartBar,
    Title: "Sales Data",
    Data: []components.ChartData{
        {Label: "Jan", Value: 100, Color: "#4CAF50"},
        {Label: "Feb", Value: 150, Color: "#2196F3"},
    },
}
```

### 7. Rich Text Editor (`components/richeditor.go`)
WYSIWYG text editor.

**Features:**
- Text formatting (bold, italic, underline)
- Headings
- Lists (ordered/unordered)
- Quotes
- Link insertion
- Clear formatting

**Usage:**
```go
editor := &components.RichEditor{
    Content: "<p>Initial content</p>",
    Height: "300px",
    OnChange: func(content string) {
        // Handle content change
    },
}
```

### 8. Calendar/Date Picker (`components/calendar.go`)
Interactive calendar component.

**Features:**
- Month navigation
- Date selection
- Today highlighting
- Min/Max date limits
- Custom date formatting

**Usage:**
```go
calendar := &components.Calendar{
    SelectedDate: time.Now(),
    OnSelect: func(date time.Time) {
        // Handle date selection
    },
}
```

### 9. Drag & Drop Component (`components/draggable.go`)
Kanban-style drag and drop interface.

**Features:**
- Multiple containers
- Drag between containers
- Visual feedback
- Reorder items
- Custom data handling

**Usage:**
```go
draggable := &components.Draggable{
    Containers: []string{"To Do", "In Progress", "Done"},
    Items: []components.DragItem{
        {ID: "1", Content: "Task 1", Group: "To Do"},
    },
    OnDrop: func(itemID, from, to string) {
        // Handle drop event
    },
}
```

### 10. Animation Framework (`components/animation.go`)
Built-in animation effects.

**Features:**
- Multiple animation types (fade, slide, bounce, rotate, pulse, shake)
- Customizable duration
- Iteration count
- Animation controls

**Usage:**
```go
animation := &components.Animation{
    Content: "<div>Animated content</div>",
    Type: components.AnimationBounce,
    Duration: "1s",
    IterationCount: "infinite",
}
```

## Running the Component Showcase

To see all components in action, run the showcase example:

```bash
# Build WASM (if needed)
cd cmd/wasm/
GOOS=js GOARCH=wasm go build -o ../../assets/json.wasm
cd ../..

# Run the showcase
go run example/example_components/example_components.go
```

Then visit http://localhost:8080 to interact with all components.

## Component Development Guidelines

When creating new components:

1. **Implement the Component interface:**
   ```go
   type Component interface {
       GetTemplate() string
       Start()
       GetDriver() LiveDriver
   }
   ```

2. **Embed ComponentDriver:**
   ```go
   type MyComponent struct {
       *liveview.ComponentDriver[*MyComponent]
       // Your fields
   }
   ```

3. **Initialize in Start():**
   ```go
   func (c *MyComponent) Start() {
       // Initialize defaults
       c.Commit() // Render component
   }
   ```

4. **Handle events:**
   ```go
   func (c *MyComponent) EventName(data interface{}) {
       // Handle event
       c.Commit() // Update UI
   }
   ```

5. **Use templates with proper IDs:**
   ```go
   func (c *MyComponent) GetTemplate() string {
       return `<div id="{{.IdComponent}}">...</div>`
   }
   ```

## Best Practices

1. **Always call Commit()** after state changes to update the UI
2. **Use proper event delegation** for dynamic content
3. **Implement proper cleanup** for resources
4. **Use CSS-in-template** for component-specific styles
5. **Handle edge cases** (empty data, invalid input)
6. **Provide sensible defaults** in Start()
7. **Document public methods** and configuration options
8. **Test with various data sizes** and edge cases

## Contributing

To contribute a new component:

1. Create the component file in `components/`
2. Implement all required interfaces
3. Add comprehensive styling
4. Create an example usage
5. Update this documentation
6. Test thoroughly

## License

See the main LICENSE file in the repository root.