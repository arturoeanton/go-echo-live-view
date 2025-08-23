# Simple Kanban Board with JSON Persistence

A lightweight kanban board implementation using Go Echo LiveView with automatic JSON file persistence.

## Features

- ✅ **Simple and Clean**: Minimal kanban board with 3 columns (To Do, In Progress, Done)
- 💾 **Automatic JSON Persistence**: All changes are saved to `kanban_board.json`
- 🎯 **Drag & Drop**: Move cards between columns with smooth drag and drop
- ➕ **Add/Delete Cards**: Easy card management with prompts
- 🔄 **Real-time Updates**: Changes are reflected immediately
- 🎨 **Beautiful UI**: Gradient background with clean card design

## How to Run

```bash
cd example/kanban_simple
go run .
```

Then open http://localhost:8080 in your browser.

## File Structure

- `main.go` - Entry point and server setup
- `simple_kanban.go` - Kanban board component with all logic
- `kanban_board.json` - Persistent storage (created automatically)

## JSON Storage Format

The board state is saved in `kanban_board.json`:

```json
{
  "cards": [
    {
      "id": "card_1234567890",
      "title": "Task Title",
      "content": "Task description",
      "column_id": "todo",
      "order": 0,
      "color": "#fff",
      "created": "2024-01-01T10:00:00Z"
    }
  ],
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
  ]
}
```

## Component Methods

### Event Handlers
- `DragStart(data)` - Handles start of drag operation
- `MoveCard(data)` - Moves card to new column
- `AddCard(data)` - Creates new card in specified column
- `DeleteCard(data)` - Removes card from board
- `SaveState()` - Persists current state to JSON

### Helper Methods
- `GetCardsForColumn(columnID)` - Returns sorted cards for a column
- `GetCardCount(columnID)` - Returns number of cards in column

## Customization

### Add New Columns
Edit the `createDefaultBoard()` method in `simple_kanban.go`:

```go
Columns: []Column{
    {ID: "backlog", Title: "Backlog", Color: "#f0f0f0", Order: 0},
    {ID: "todo", Title: "To Do", Color: "#e3e8ef", Order: 1},
    // Add more columns here
},
```

### Change Styling
Modify the CSS in the `GetTemplate()` method for custom styling.

## Architecture

This example demonstrates:
- Component-based architecture with LiveView
- JSON file persistence layer
- Mutex-based thread safety for concurrent access
- Event-driven updates with WebSocket communication
- Shared board instance across connections

## Production Considerations

For production use, consider:
1. Adding connection wrappers for proper multi-user support
2. Implementing user authentication
3. Adding database persistence instead of JSON files
4. Adding card metadata (assignee, due dates, labels)
5. Implementing column WIP limits
6. Adding search and filter functionality