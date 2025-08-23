package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// MoveCard handles dragging and dropping cards between columns
func (k *SimpleKanbanModal) MoveCard(data interface{}) {
	var event map[string]interface{}
	if jsonData, ok := data.(string); ok {
		json.Unmarshal([]byte(jsonData), &event)
		
		cardID := event["cardId"].(string)
		columnID := event["columnId"].(string)
		
		updatedCards := make([]KanbanCard, len(k.Cards))
		copy(updatedCards, k.Cards)
		
		for i := range updatedCards {
			if updatedCards[i].ID == cardID {
				if updatedCards[i].ColumnID != columnID {
					fmt.Printf("Moving card %s to column %s\n", cardID, columnID)
					updatedCards[i].ColumnID = columnID
					updatedCards[i].UpdatedAt = time.Now()
				}
				break
			}
		}
		
		k.updateGlobalState(k.Columns, updatedCards)
	}
}

// EditCard opens the modal dialog for editing an existing card
func (k *SimpleKanbanModal) EditCard(data interface{}) {
	cardID := ""
	if id, ok := data.(string); ok {
		cardID = id
	}
	
	for _, card := range k.Cards {
		if card.ID == cardID {
			k.ShowModal = true
			k.ModalType = "edit_card"
			k.ModalTitle = "Edit Card"
			k.FormCardID = card.ID
			k.FormCardTitle = card.Title
			k.FormCardDesc = card.Description
			k.FormCardColumn = card.ColumnID
			k.FormCardPriority = card.Priority
			k.FormCardPoints = card.Points
			k.FormCardAttachments = card.Attachments
			k.FormCardLinks = card.Links
			k.FormCardTags = card.Tags
			k.FormCardChecklist = card.Checklist
			if card.DueDate != nil {
				k.FormCardDueDate = card.DueDate.Format("2006-01-02")
			} else {
				k.FormCardDueDate = ""
			}
			if k.FormCardPriority == "" {
				k.FormCardPriority = "medium"
			}
			break
		}
	}
	k.Commit()
}

// AddCard opens the modal dialog for creating a new card
func (k *SimpleKanbanModal) AddCard(data interface{}) {
	columnID := ""
	if id, ok := data.(string); ok {
		columnID = id
	}
	
	k.ShowModal = true
	k.ModalType = "add_card"
	k.ModalTitle = "Add New Card"
	k.FormCardID = ""
	k.FormCardTitle = ""
	k.FormCardDesc = ""
	k.FormCardColumn = columnID
	k.FormCardPriority = "medium"
	k.FormCardPoints = 0
	k.FormCardAttachments = []Attachment{}
	k.FormCardLinks = []ExternalLink{}
	k.FormCardTags = []string{}
	k.FormCardDueDate = ""
	k.FormCardChecklist = []ChecklistItem{}
	k.Commit()
}

// EditColumn opens the modal for editing a column
func (k *SimpleKanbanModal) EditColumn(data interface{}) {
	columnID := ""
	if id, ok := data.(string); ok {
		columnID = id
	}
	
	for _, col := range k.Columns {
		if col.ID == columnID {
			k.ShowModal = true
			k.ModalType = "edit_column"
			k.ModalTitle = "Edit Column"
			k.FormColumnID = col.ID
			k.FormColumnTitle = col.Title
			k.FormColumnColor = col.Color
			break
		}
	}
	k.Commit()
}

// AddColumn opens the modal for adding a new column
func (k *SimpleKanbanModal) AddColumn(data interface{}) {
	k.ShowModal = true
	k.ModalType = "add_column"
	k.ModalTitle = "Add New Column"
	k.FormColumnID = ""
	k.FormColumnTitle = ""
	k.FormColumnColor = "#e3e8ef"
	k.Commit()
}

// CloseModal closes the modal dialog
func (k *SimpleKanbanModal) CloseModal(data interface{}) {
	k.ShowModal = false
	k.Commit()
}

// SaveModal saves the modal form data
func (k *SimpleKanbanModal) SaveModal(data interface{}) {
	wasAddCard := k.ModalType == "add_card"
	
	if k.ModalType == "edit_card" || k.ModalType == "add_card" {
		k.saveCard()
	} else if k.ModalType == "edit_column" || k.ModalType == "add_column" {
		k.saveColumn()
	}
	
	// Don't close modal if it was a new card (to allow file uploads)
	if !wasAddCard {
		k.ShowModal = false
	}
	
	k.updateGlobalState(k.Columns, k.Cards)
	k.Commit()
}

func (k *SimpleKanbanModal) saveCard() {
	if k.ModalType == "add_card" {
		newCardID := fmt.Sprintf("card_%d", time.Now().UnixNano())
		newCard := KanbanCard{
			ID:          newCardID,
			Title:       k.FormCardTitle,
			Description: k.FormCardDesc,
			ColumnID:    k.FormCardColumn,
			Priority:    k.FormCardPriority,
			Points:      k.FormCardPoints,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Attachments: k.FormCardAttachments,
			Links:       k.FormCardLinks,
			Tags:        k.FormCardTags,
			Checklist:   k.FormCardChecklist,
		}
		
		if k.FormCardDueDate != "" {
			if dueDate, err := time.Parse("2006-01-02", k.FormCardDueDate); err == nil {
				newCard.DueDate = &dueDate
			}
		}
		
		k.Cards = append(k.Cards, newCard)
		// Update FormCardID so attachments can be uploaded
		k.FormCardID = newCardID
		k.ModalType = "edit_card" // Change to edit mode after saving
	} else {
		for i := range k.Cards {
			if k.Cards[i].ID == k.FormCardID {
				k.Cards[i].Title = k.FormCardTitle
				k.Cards[i].Description = k.FormCardDesc
				k.Cards[i].ColumnID = k.FormCardColumn
				k.Cards[i].Priority = k.FormCardPriority
				k.Cards[i].Points = k.FormCardPoints
				k.Cards[i].UpdatedAt = time.Now()
				k.Cards[i].Attachments = k.FormCardAttachments
				k.Cards[i].Links = k.FormCardLinks
				k.Cards[i].Tags = k.FormCardTags
				k.Cards[i].Checklist = k.FormCardChecklist
				
				if k.FormCardDueDate != "" {
					if dueDate, err := time.Parse("2006-01-02", k.FormCardDueDate); err == nil {
						k.Cards[i].DueDate = &dueDate
					}
				} else {
					k.Cards[i].DueDate = nil
				}
				break
			}
		}
	}
}

func (k *SimpleKanbanModal) saveColumn() {
	if k.ModalType == "add_column" {
		newColumn := KanbanColumn{
			ID:    fmt.Sprintf("col_%d", time.Now().UnixNano()),
			Title: k.FormColumnTitle,
			Color: k.FormColumnColor,
			Order: len(k.Columns),
		}
		k.Columns = append(k.Columns, newColumn)
	} else {
		for i := range k.Columns {
			if k.Columns[i].ID == k.FormColumnID {
				k.Columns[i].Title = k.FormColumnTitle
				k.Columns[i].Color = k.FormColumnColor
				break
			}
		}
	}
}

// DeleteCard deletes a card
func (k *SimpleKanbanModal) DeleteCard(data interface{}) {
	updatedCards := []KanbanCard{}
	for _, card := range k.Cards {
		if card.ID != k.FormCardID {
			updatedCards = append(updatedCards, card)
		}
	}
	k.Cards = updatedCards
	k.ShowModal = false
	k.updateGlobalState(k.Columns, k.Cards)
	k.Commit()
}

// DeleteColumn deletes a column and all its cards
func (k *SimpleKanbanModal) DeleteColumn(data interface{}) {
	updatedColumns := []KanbanColumn{}
	for _, col := range k.Columns {
		if col.ID != k.FormColumnID {
			updatedColumns = append(updatedColumns, col)
		}
	}
	
	updatedCards := []KanbanCard{}
	for _, card := range k.Cards {
		if card.ColumnID != k.FormColumnID {
			updatedCards = append(updatedCards, card)
		}
	}
	
	k.Columns = updatedColumns
	k.Cards = updatedCards
	k.ShowModal = false
	k.updateGlobalState(k.Columns, k.Cards)
	k.Commit()
}

// UpdateFormField updates a form field value
func (k *SimpleKanbanModal) UpdateFormField(data interface{}) {
	var field map[string]interface{}
	if jsonData, ok := data.(string); ok {
		json.Unmarshal([]byte(jsonData), &field)
		
		fieldName := field["field"].(string)
		value := field["value"]
		
		switch fieldName {
		case "card_title":
			k.FormCardTitle = value.(string)
		case "card_desc":
			k.FormCardDesc = value.(string)
		case "card_column":
			k.FormCardColumn = value.(string)
		case "card_priority":
			k.FormCardPriority = value.(string)
		case "card_points":
			k.FormCardPoints = int(value.(float64))
		case "card_due_date":
			k.FormCardDueDate = value.(string)
		case "column_title":
			k.FormColumnTitle = value.(string)
		case "column_color":
			k.FormColumnColor = value.(string)
		case "board_name":
			k.FormBoardName = value.(string)
		}
	}
}

// Tag management
func (k *SimpleKanbanModal) AddTag(data interface{}) {
	tag := ""
	if t, ok := data.(string); ok {
		tag = strings.TrimSpace(t)
	}
	
	if tag != "" {
		// Check if tag already exists
		for _, existingTag := range k.FormCardTags {
			if existingTag == tag {
				return
			}
		}
		k.FormCardTags = append(k.FormCardTags, tag)
		k.Commit()
	}
}

func (k *SimpleKanbanModal) RemoveTag(data interface{}) {
	tag := ""
	if t, ok := data.(string); ok {
		tag = t
	}
	
	newTags := []string{}
	for _, t := range k.FormCardTags {
		if t != tag {
			newTags = append(newTags, t)
		}
	}
	k.FormCardTags = newTags
	k.Commit()
}

// Link management
func (k *SimpleKanbanModal) AddLink(data interface{}) {
	var linkData map[string]interface{}
	if jsonData, ok := data.(string); ok {
		json.Unmarshal([]byte(jsonData), &linkData)
		
		newLink := ExternalLink{
			ID:    fmt.Sprintf("link_%d", time.Now().UnixNano()),
			Title: linkData["title"].(string),
			URL:   linkData["url"].(string),
		}
		
		k.FormCardLinks = append(k.FormCardLinks, newLink)
		k.Commit()
	}
}

func (k *SimpleKanbanModal) RemoveLink(data interface{}) {
	linkID := ""
	if id, ok := data.(string); ok {
		linkID = id
	}
	
	newLinks := []ExternalLink{}
	for _, link := range k.FormCardLinks {
		if link.ID != linkID {
			newLinks = append(newLinks, link)
		}
	}
	k.FormCardLinks = newLinks
	k.Commit()
}

// Checklist management
func (k *SimpleKanbanModal) AddChecklistItem(data interface{}) {
	text := ""
	if t, ok := data.(string); ok {
		text = strings.TrimSpace(t)
	}
	
	if text != "" {
		newItem := ChecklistItem{
			ID:      fmt.Sprintf("check_%d", time.Now().UnixNano()),
			Text:    text,
			Checked: false,
		}
		k.FormCardChecklist = append(k.FormCardChecklist, newItem)
		k.Commit()
	}
}

func (k *SimpleKanbanModal) RemoveChecklistItem(data interface{}) {
	itemID := ""
	if id, ok := data.(string); ok {
		itemID = id
	}
	
	newChecklist := []ChecklistItem{}
	for _, item := range k.FormCardChecklist {
		if item.ID != itemID {
			newChecklist = append(newChecklist, item)
		}
	}
	k.FormCardChecklist = newChecklist
	k.Commit()
}

func (k *SimpleKanbanModal) ToggleChecklistItem(data interface{}) {
	itemID := ""
	if id, ok := data.(string); ok {
		itemID = id
	}
	
	for i := range k.FormCardChecklist {
		if k.FormCardChecklist[i].ID == itemID {
			k.FormCardChecklist[i].Checked = !k.FormCardChecklist[i].Checked
			break
		}
	}
	k.Commit()
}

// Other handlers that were in the main file
func (k *SimpleKanbanModal) DismissAlert(data interface{}) {
	k.ShowAlert = false
	k.Commit()
}

func (k *SimpleKanbanModal) SwitchBoard(data interface{}) {
	boardName := ""
	if name, ok := data.(string); ok {
		boardName = name
	}
	
	if boardName != "" && boardName != k.CurrentBoard {
		k.CurrentBoard = boardName
		
		globalMutex.Lock()
		if globalBoards == nil {
			globalBoards = make(map[string]*KanbanBoardData)
		}
		
		if globalBoards[boardName] == nil {
			boardData := loadBoardData(boardName)
			if boardData != nil {
				globalBoards[boardName] = boardData
			}
		}
		
		if globalBoards[boardName] != nil {
			currentBoard := globalBoards[boardName]
			k.Columns = make([]KanbanColumn, len(currentBoard.Columns))
			copy(k.Columns, currentBoard.Columns)
			k.Cards = make([]KanbanCard, len(currentBoard.Cards))
			copy(k.Cards, currentBoard.Cards)
		}
		globalMutex.Unlock()
		
		k.Commit()
	}
}

func (k *SimpleKanbanModal) NewBoard(data interface{}) {
	k.ShowModal = true
	k.ModalType = "new_board"
	k.ModalTitle = "Create New Board"
	k.FormBoardName = ""
	k.Commit()
}

func (k *SimpleKanbanModal) CreateBoard(data interface{}) {
	if k.FormBoardName != "" {
		newBoard, err := createNewBoard(k.FormBoardName)
		if err == nil {
			globalMutex.Lock()
			globalBoards[k.FormBoardName] = newBoard
			globalMutex.Unlock()
			
			k.BoardsList = getAvailableBoards()
			k.CurrentBoard = k.FormBoardName
			k.Columns = newBoard.Columns
			k.Cards = newBoard.Cards
			k.ShowModal = false
			k.Commit()
		}
	}
}

func (k *SimpleKanbanModal) RefreshBoards(data interface{}) {
	k.BoardsList = getAvailableBoards()
	k.Commit()
}

func (k *SimpleKanbanModal) ArchiveBoard(data interface{}) {
	// Implementation for archiving boards
	k.Commit()
}

func (k *SimpleKanbanModal) ReorderColumns(data interface{}) {
	fmt.Println("📋 ReorderColumns event received")
	// Implementation for column reordering
	k.Commit()
}

// File attachment handlers (placeholders)
func (k *SimpleKanbanModal) UploadFiles(data interface{}) {
	// Implementation for file upload
	k.Commit()
}

func (k *SimpleKanbanModal) RemoveAttachment(data interface{}) {
	attachmentID := ""
	if id, ok := data.(string); ok {
		attachmentID = id
	}
	
	// Find the attachment to remove and delete its file
	var filenameToDelete string
	newAttachments := []Attachment{}
	for _, att := range k.FormCardAttachments {
		if att.ID != attachmentID {
			newAttachments = append(newAttachments, att)
		} else {
			filenameToDelete = att.Name
		}
	}
	
	// Delete the physical file if found
	if filenameToDelete != "" && k.FormCardID != "" {
		filePath := filepath.Join("attachments", k.CurrentBoard, k.FormCardID, filenameToDelete)
		if err := os.Remove(filePath); err != nil {
			fmt.Printf("⚠️ Could not delete file %s: %v\n", filePath, err)
		} else {
			fmt.Printf("🗑️ Deleted attachment file: %s\n", filePath)
		}
	}
	
	k.FormCardAttachments = newAttachments
	
	// Also update the card in the Cards array
	for i := range k.Cards {
		if k.Cards[i].ID == k.FormCardID {
			k.Cards[i].Attachments = k.FormCardAttachments
			break
		}
	}
	
	k.Commit()
}

func (k *SimpleKanbanModal) RefreshAttachments(data interface{}) {
	// Parse the refresh data
	var refreshData map[string]interface{}
	if jsonData, ok := data.(string); ok {
		json.Unmarshal([]byte(jsonData), &refreshData)
		
		cardID := refreshData["cardID"].(string)
		files := refreshData["files"].([]interface{})
		
		// Update attachments for the current card being edited
		if k.FormCardID == cardID {
			// Convert files to Attachment structs
			for _, file := range files {
				fileMap := file.(map[string]interface{})
				// Get the original filename and the full path with ID
				originalName := ""
				if name, ok := fileMap["name"].(string); ok {
					originalName = name
				}
				
				// Build the full filename with attachment ID prefix
				attachmentID := fileMap["id"].(string)
				fullFilename := attachmentID + "_" + originalName
				
				attachment := Attachment{
					ID:          attachmentID,
					Name:        fullFilename, // Store full filename for download
					DisplayName: originalName, // Store original name for display
					Size:        int64(fileMap["size"].(float64)),
				}
				k.FormCardAttachments = append(k.FormCardAttachments, attachment)
			}
			
			// Also update the card in the Cards array
			for i := range k.Cards {
				if k.Cards[i].ID == cardID {
					k.Cards[i].Attachments = k.FormCardAttachments
					break
				}
			}
		}
	}
	k.Commit()
}

func (k *SimpleKanbanModal) handleFileUpload(files []interface{}) error {
	// Placeholder for file upload handling
	return nil
}