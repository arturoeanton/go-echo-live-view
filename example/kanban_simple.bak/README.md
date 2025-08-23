# Simple Kanban Board

A real-time, collaborative Kanban board built with Go Echo LiveView framework. Features drag-and-drop functionality, persistent storage, and automatic synchronization across multiple users.

## Features

- **Real-time Collaboration**: Changes are instantly synchronized across all connected users via WebSocket
- **Drag & Drop**: 
  - Move cards between columns
  - Reorder columns by dragging their headers
- **Card Management**:
  - Create, edit, and delete cards
  - Assign priority levels (Low, Medium, High, Urgent)
  - Add story points for estimation
  - Write descriptions
- **Column Management**:
  - Create and edit columns
  - Custom colors for visual organization
  - Automatic card and point totals
- **Persistent Storage**: All data is saved to `kanban_board.json`
- **Clean UI**: Modern, responsive design with smooth animations

## Installation

### Prerequisites

- Go 1.19 or higher
- The go-echo-live-view framework (parent project)

### Setup

1. Navigate to the example directory:
```bash
cd example/kanban_simple
```

2. Run the application:
```bash
go run .
```

3. Open your browser and navigate to:
```
http://localhost:8080
```

## Usage

### Managing Cards

- **Add a Card**: Click the "+ Add Card" button in any column
- **Edit a Card**: Click on any card to open the edit modal
- **Move Cards**: Drag and drop cards between columns
- **Card Priority**: Set urgency levels with color-coded badges
- **Story Points**: Assign effort estimation (0-100 points)

### Card Properties

- **Title**: The main card title (required)
- **Description**: Additional details about the task
- **Priority**: Set urgency level (Low, Medium, High, Urgent)
- **Points**: Story points for effort estimation (0-100)
- **Column**: Which column the card belongs to

### Managing Columns

- **Add a Column**: Click the "+ Add Column" button
- **Edit a Column**: Double-click on any column header
- **Reorder Columns**: Drag column headers to rearrange (swap positions)
- **Column Colors**: Customize colors for better visual organization

### Visual Indicators

- **Points Badge**: Blue badge showing story points (bottom-right of cards)
- **Priority Badges**: Color-coded priority indicators
  - Gray: Low priority
  - Orange: Medium priority
  - Red: High priority
  - Purple: Urgent priority
- **Column Stats**: Header shows total cards and points

## Data Storage

The board automatically saves all changes to `kanban_board.json`. This file contains:

- Column definitions (id, title, color, order)
- Card data (id, title, description, column, priority, points, timestamps)

### Example Data Structure

```json
{
  "columns": [
    {
      "id": "todo",
      "title": "To Do",
      "color": "#e3e8ef",
      "order": 0
    },
    {
      "id": "doing",
      "title": "In Progress",
      "color": "#ffd4a3",
      "order": 1
    },
    {
      "id": "done",
      "title": "Done",
      "color": "#a3e4d7",
      "order": 2
    }
  ],
  "cards": [
    {
      "id": "card_1755897826",
      "title": "Example Task",
      "description": "Task description here",
      "column_id": "todo",
      "priority": "medium",
      "points": 5,
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ]
}
```

## Technical Details

### Architecture

- **Backend**: Go with Echo web framework
- **Real-time**: WebSocket connections via LiveView
- **Frontend**: Server-side rendered HTML with real-time DOM updates
- **Storage**: JSON file-based persistence with mutex protection
- **Synchronization**: Global state management across all connected clients

### Key Components

1. **SimpleKanbanModal**: Main component handling all board logic
2. **KanbanBoardData**: Data structure for columns and cards
3. **Global State Management**: Synchronized state across all connected clients
4. **Event Handlers**: Processes user interactions (drag, drop, click, etc.)

### WebSocket Events

- `MoveCard`: Triggered when cards are dragged between columns
- `EditCard`: Opens the card editing modal
- `AddCard`: Creates a new card in a column
- `EditColumn`: Opens column editing modal
- `AddColumn`: Creates a new column
- `ReorderColumns`: Handles column drag and drop (swap)
- `SaveModal`: Persists modal form changes
- `UpdateFormField`: Real-time form field updates
- `CloseModal`: Closes the active modal

## Development

### File Structure

```
kanban_simple/
├── main.go                 # Application entry point
├── simple_kanban_modal.go  # Main Kanban component
├── kanban_board.json      # Persistent data storage
├── README.md              # This file (English)
└── README.es.md           # Spanish documentation
```

### Code Structure

The main component (`SimpleKanbanModal`) includes:

- **Data Structures**: `KanbanColumn`, `KanbanCard`, `KanbanBoardData`
- **State Management**: Global mutex-protected state
- **Event Handlers**: All user interaction handlers
- **Template**: Complete HTML/CSS/JS in `GetTemplate()`
- **Helper Methods**: `GetCardsForColumn()`, `GetCardCount()`, `GetColumnPoints()`, etc.

### Extending the Application

To add new features:

1. Add event handlers to the `Events` map in `Start()`
2. Implement the handler method on `SimpleKanbanModal`
3. Update the template to include UI elements
4. Add necessary fields to data structures
5. Update JSON persistence if needed

### Running in Development

For automatic reload during development:

```bash
# Install gomon if not already installed
go install github.com/c9s/gomon@latest

# Run with auto-reload
gomon
```

## Browser Compatibility

- Chrome/Edge (recommended)
- Firefox
- Safari
- Any modern browser with WebSocket support

## Key Features Explained

### Column Reordering
Columns can be reordered by dragging their headers. The system uses a simple swap mechanism - when you drop a column on another, they exchange positions.

### Story Points
Each card can have story points (0-100) for effort estimation. The total points per column are displayed in the column header.

### Real-time Synchronization
All changes are immediately broadcast to all connected users. The system uses a global state manager with mutex protection to ensure data consistency.

### Persistent Storage
Every change triggers an automatic save to `kanban_board.json`. The system loads this file on startup, ensuring data persistence across server restarts.

## Known Limitations

- File-based storage (not suitable for high-traffic production use)
- No user authentication/authorization
- No card archiving/deletion (cards remain in the system)
- Single board instance (no multi-board support)
- No undo/redo functionality
- No search or filtering capabilities

## Performance Considerations

- Suitable for small to medium teams (up to ~50 concurrent users)
- JSON file can handle thousands of cards efficiently
- WebSocket connections are lightweight and responsive
- Mutex protection ensures thread safety but may impact performance under heavy load

## Contributing

Feel free to submit issues and enhancement requests! Some ideas for contributions:

- Add user authentication
- Implement card archiving/deletion
- Add search and filter functionality
- Create board templates
- Add card attachments support
- Implement activity logging

## License

This example is part of the go-echo-live-view project and follows the same license terms.