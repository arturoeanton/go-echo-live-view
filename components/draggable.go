package components

import (
	"fmt"
	"github.com/arturoeanton/go-echo-live-view/liveview"
)

type DragItem struct {
	ID      string
	Content string
	Group   string
	Data    interface{}
}

type Draggable struct {
	*liveview.ComponentDriver[*Draggable]
	Items      []DragItem
	Containers []string
	OnDrop     func(itemID, fromContainer, toContainer string)
	OnReorder  func(items []DragItem)
}

func (d *Draggable) Start() {
	if len(d.Containers) == 0 {
		d.Containers = []string{"container1"}
	}
	d.Commit()
}

func (d *Draggable) GetTemplate() string {
	return `
	<div id="{{.IdComponent}}" class="draggable-area">
		<style>
			.draggable-area { display: flex; gap: 2rem; flex-wrap: wrap; }
			.drag-container { min-width: 250px; min-height: 200px; background: #f5f5f5; border: 2px dashed #ddd; border-radius: 8px; padding: 1rem; }
			.drag-container.drag-over { background: #e8f5e9; border-color: #4CAF50; }
			.drag-item { background: white; padding: 1rem; margin: 0.5rem 0; border-radius: 4px; cursor: move; box-shadow: 0 2px 4px rgba(0,0,0,0.1); transition: all 0.3s; }
			.drag-item:hover { box-shadow: 0 4px 8px rgba(0,0,0,0.15); transform: translateY(-2px); }
			.drag-item.dragging { opacity: 0.5; }
			.container-title { font-weight: 600; margin-bottom: 1rem; color: #666; }
			.empty-state { text-align: center; color: #999; padding: 2rem; }
		</style>
		
		{{range $containerName := .Containers}}
		<div class="drag-container" 
			data-container="{{$containerName}}"
			ondragover="event.preventDefault(); this.classList.add('drag-over')"
			ondragleave="this.classList.remove('drag-over')"
			ondrop="event.preventDefault(); this.classList.remove('drag-over'); 
				var itemId = event.dataTransfer.getData('itemId');
				var fromContainer = event.dataTransfer.getData('fromContainer');
				send_event('{{$.IdComponent}}', 'Drop', JSON.stringify({
					itemId: itemId, 
					from: fromContainer, 
					to: '{{$containerName}}'
				}))"
		>
			<div class="container-title">{{$containerName}}</div>
			{{$currentContainer := $containerName}}
			{{$hasItems := false}}
			{{range $.Items}}
				{{if eq .Group $currentContainer}}
				{{$hasItems = true}}
				<div class="drag-item" 
					draggable="true"
					data-item-id="{{.ID}}"
					ondragstart="event.dataTransfer.setData('itemId', '{{.ID}}'); 
						event.dataTransfer.setData('fromContainer', '{{.Group}}');
						this.classList.add('dragging')"
					ondragend="this.classList.remove('dragging')"
				>
					{{.Content}}
				</div>
				{{end}}
			{{end}}
			{{if not $hasItems}}
			<div class="empty-state">Drop items here</div>
			{{end}}
		</div>
		{{end}}
	</div>
	`
}

func (d *Draggable) GetDriver() liveview.LiveDriver {
	return d
}

func (d *Draggable) Drop(data interface{}) {
	params := make(map[string]string)
	fmt.Sscanf(data.(string), "%v", &params)
	
	itemID := params["itemId"]
	fromContainer := params["from"]
	toContainer := params["to"]
	
	for i, item := range d.Items {
		if item.ID == itemID {
			d.Items[i].Group = toContainer
			break
		}
	}
	
	if d.OnDrop != nil {
		d.OnDrop(itemID, fromContainer, toContainer)
	}
	
	d.Commit()
}

func (d *Draggable) AddItem(id, content, group string, data interface{}) {
	d.Items = append(d.Items, DragItem{
		ID:      id,
		Content: content,
		Group:   group,
		Data:    data,
	})
	d.Commit()
}

func (d *Draggable) RemoveItem(id string) {
	newItems := []DragItem{}
	for _, item := range d.Items {
		if item.ID != id {
			newItems = append(newItems, item)
		}
	}
	d.Items = newItems
	d.Commit()
}

func (d *Draggable) GetItemsByGroup(group string) []DragItem {
	items := []DragItem{}
	for _, item := range d.Items {
		if item.Group == group {
			items = append(items, item)
		}
	}
	return items
}

func (d *Draggable) MoveItem(itemID, toGroup string) {
	for i, item := range d.Items {
		if item.ID == itemID {
			d.Items[i].Group = toGroup
			break
		}
	}
	d.Commit()
}