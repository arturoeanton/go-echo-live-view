package components

import (
	"fmt"
	"github.com/arturoeanton/go-echo-live-view/liveview"
)

type Tab struct {
	ID      string
	Label   string
	Content string
	Active  bool
}

type Tabs struct {
	*liveview.ComponentDriver[*Tabs]
	
	Tabs       []Tab
	ActiveTab  string
	OnTabChange func(tabID string)
}

func (t *Tabs) Start() {
	if t.ActiveTab == "" && len(t.Tabs) > 0 {
		t.ActiveTab = t.Tabs[0].ID
	}
	t.updateTabStates()
	t.Commit()
}

func (t *Tabs) updateTabStates() {
	for i := range t.Tabs {
		t.Tabs[i].Active = (t.Tabs[i].ID == t.ActiveTab)
	}
}

func (t *Tabs) GetTemplate() string {
	return `
	<div id="{{.IdComponent}}" class="tabs-container">
		<style>
			.tabs-header {
				display: flex;
				gap: 1rem;
				margin-bottom: 2rem;
				flex-wrap: wrap;
				border-bottom: 2px solid #eee;
				padding-bottom: 1rem;
			}
			.tab-button {
				padding: 0.75rem 1.5rem;
				border: 1px solid #ddd;
				border-radius: 4px;
				cursor: pointer;
				font-weight: 500;
				transition: all 0.3s ease;
			}
			.tab-button.active {
				background: #4CAF50;
				color: white;
				border-color: #4CAF50;
			}
			.tab-button:not(.active) {
				background: white;
				color: #333;
			}
			.tab-button:hover:not(.active) {
				background: #f5f5f5;
			}
			.tabs-content {
				background: white;
				padding: 2rem;
				border-radius: 8px;
				box-shadow: 0 2px 8px rgba(0,0,0,0.1);
			}
			.tab-panel {
				display: none;
			}
			.tab-panel.active {
				display: block;
			}
		</style>
		
		<div class="tabs-header">
			{{range .Tabs}}
			<button class="tab-button {{if .Active}}active{{end}}" 
					onclick="send_event('{{$.IdComponent}}', 'SelectTab', '{{.ID}}')">
				{{.Label}}
			</button>
			{{end}}
		</div>
		
		<div class="tabs-content">
			{{range .Tabs}}
			<div class="tab-panel {{if .Active}}active{{end}}" data-tab-id="{{.ID}}">
				{{.Content}}
			</div>
			{{end}}
		</div>
	</div>
	`
}

func (t *Tabs) GetDriver() liveview.LiveDriver {
	return t
}

func (t *Tabs) SelectTab(data interface{}) {
	tabID := data.(string)
	t.ActiveTab = tabID
	t.updateTabStates()
	
	liveview.Info("Tab selected: %s", tabID)
	
	// Use smart update to preserve content
	t.smartTabSwitch(tabID)
	
	if t.OnTabChange != nil {
		t.OnTabChange(tabID)
	}
}

func (t *Tabs) smartTabSwitch(tabID string) {
	// Use JavaScript to switch tabs without destroying content
	script := fmt.Sprintf(`
		(function() {
			// Hide all panels
			document.querySelectorAll('#%s .tab-panel').forEach(function(panel) {
				panel.classList.remove('active');
			});
			
			// Show selected panel
			var activePanel = document.querySelector('#%s .tab-panel[data-tab-id="%s"]');
			if (activePanel) {
				activePanel.classList.add('active');
			}
			
			// Update button states
			document.querySelectorAll('#%s .tab-button').forEach(function(btn) {
				btn.classList.remove('active');
			});
			
			// Find and activate the clicked button
			var buttons = document.querySelectorAll('#%s .tab-button');
			var index = %d;
			if (buttons[index]) {
				buttons[index].classList.add('active');
			}
		})();
	`, t.IdComponent, t.IdComponent, tabID, t.IdComponent, t.IdComponent, t.getTabIndex(tabID))
	
	t.EvalScript(script)
}

func (t *Tabs) getTabIndex(tabID string) int {
	for i, tab := range t.Tabs {
		if tab.ID == tabID {
			return i
		}
	}
	return 0
}

func (t *Tabs) SetActiveTab(tabID string) {
	t.ActiveTab = tabID
	t.updateTabStates()
	t.smartTabSwitch(tabID)
}

func (t *Tabs) AddTab(tab Tab) {
	t.Tabs = append(t.Tabs, tab)
	t.updateTabStates()
	t.Commit()
}

func (t *Tabs) RemoveTab(tabID string) {
	newTabs := []Tab{}
	for _, tab := range t.Tabs {
		if tab.ID != tabID {
			newTabs = append(newTabs, tab)
		}
	}
	t.Tabs = newTabs
	
	// If we removed the active tab, select the first one
	if t.ActiveTab == tabID && len(t.Tabs) > 0 {
		t.ActiveTab = t.Tabs[0].ID
	}
	
	t.updateTabStates()
	t.Commit()
}