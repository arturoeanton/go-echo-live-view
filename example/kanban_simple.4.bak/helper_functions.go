package main

import (
	"time"
	"fmt"
)

// GetCardsForColumn returns all cards that belong to a specific column
func (k *SimpleKanbanModal) GetCardsForColumn(columnID string) []KanbanCard {
	var cards []KanbanCard
	for _, card := range k.Cards {
		if card.ColumnID == columnID {
			cards = append(cards, card)
		}
	}
	return cards
}

// GetCardCount returns the number of cards in a specific column
func (k *SimpleKanbanModal) GetCardCount(columnID string) int {
	count := 0
	for _, card := range k.Cards {
		if card.ColumnID == columnID {
			count++
		}
	}
	return count
}

// GetColumnPoints calculates and returns the total story points for all cards in a column
func (k *SimpleKanbanModal) GetColumnPoints(columnID string) int {
	total := 0
	for _, card := range k.Cards {
		if card.ColumnID == columnID {
			total += card.Points
		}
	}
	return total
}

// GetOrderedColumns returns a sorted copy of columns based on their Order field
func (k *SimpleKanbanModal) GetOrderedColumns() []KanbanColumn {
	columns := make([]KanbanColumn, len(k.Columns))
	copy(columns, k.Columns)
	
	for i := 0; i < len(columns)-1; i++ {
		for j := 0; j < len(columns)-i-1; j++ {
			if columns[j].Order > columns[j+1].Order {
				columns[j], columns[j+1] = columns[j+1], columns[j]
			}
		}
	}
	return columns
}

// IsOverdue checks if a due date has passed
func (k *SimpleKanbanModal) IsOverdue(dueDate *time.Time) bool {
	if dueDate == nil {
		return false
	}
	return dueDate.Before(time.Now())
}

// FormatFileSize formats bytes to human readable format
func (k *SimpleKanbanModal) FormatFileSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

// CountCheckedItems counts the number of checked items in a checklist
func (k *SimpleKanbanModal) CountCheckedItems(checklist []ChecklistItem) int {
	count := 0
	for _, item := range checklist {
		if item.Checked {
			count++
		}
	}
	return count
}