package components

import (
	"github.com/arturoeanton/go-echo-live-view/liveview"
)

type Modal struct {
	*liveview.ComponentDriver[*Modal]
	IsOpen      bool
	Title       string
	Content     string
	Size        string
	Closable    bool
	ShowFooter  bool
	OkText      string
	CancelText  string
	OnOk        func()
	OnCancel    func()
	OnClose     func()
}

func (m *Modal) Start() {
	if m.Size == "" {
		m.Size = "medium"
	}
	if m.OkText == "" {
		m.OkText = "OK"
	}
	if m.CancelText == "" {
		m.CancelText = "Cancel"
	}
	if m.Closable {
		m.Closable = true
	}
	m.Commit()
}

func (m *Modal) GetTemplate() string {
	return `
	<div id="{{.IdComponent}}" class="modal-wrapper">
		<style>
			.modal-overlay {
				position: fixed;
				top: 0;
				left: 0;
				right: 0;
				bottom: 0;
				background: rgba(0, 0, 0, 0.5);
				display: flex;
				align-items: center;
				justify-content: center;
				z-index: 1000;
				animation: fadeIn 0.2s ease;
			}
			.modal-overlay.hidden {
				display: none;
			}
			@keyframes fadeIn {
				from { opacity: 0; }
				to { opacity: 1; }
			}
			@keyframes slideUp {
				from { 
					transform: translateY(20px);
					opacity: 0;
				}
				to { 
					transform: translateY(0);
					opacity: 1;
				}
			}
			.modal-container {
				background: white;
				border-radius: 8px;
				box-shadow: 0 4px 20px rgba(0, 0, 0, 0.15);
				max-height: 90vh;
				display: flex;
				flex-direction: column;
				animation: slideUp 0.3s ease;
			}
			.modal-container.small {
				width: 400px;
			}
			.modal-container.medium {
				width: 600px;
			}
			.modal-container.large {
				width: 900px;
			}
			.modal-container.full {
				width: 90vw;
				height: 90vh;
			}
			.modal-header {
				padding: 1.5rem;
				border-bottom: 1px solid #e0e0e0;
				display: flex;
				justify-content: space-between;
				align-items: center;
			}
			.modal-title {
				font-size: 1.25rem;
				font-weight: 600;
				color: #333;
				margin: 0;
			}
			.modal-close {
				background: none;
				border: none;
				font-size: 1.5rem;
				color: #999;
				cursor: pointer;
				padding: 0;
				width: 32px;
				height: 32px;
				display: flex;
				align-items: center;
				justify-content: center;
				border-radius: 4px;
				transition: all 0.2s;
			}
			.modal-close:hover {
				background: #f5f5f5;
				color: #333;
			}
			.modal-body {
				padding: 1.5rem;
				overflow-y: auto;
				flex: 1;
			}
			.modal-footer {
				padding: 1rem 1.5rem;
				border-top: 1px solid #e0e0e0;
				display: flex;
				justify-content: flex-end;
				gap: 0.75rem;
			}
			.modal-button {
				padding: 0.5rem 1rem;
				border-radius: 4px;
				font-size: 0.875rem;
				font-weight: 500;
				cursor: pointer;
				transition: all 0.2s;
				min-width: 80px;
			}
			.modal-button-cancel {
				background: white;
				border: 1px solid #ddd;
				color: #666;
			}
			.modal-button-cancel:hover {
				background: #f5f5f5;
				border-color: #999;
			}
			.modal-button-ok {
				background: #4CAF50;
				border: 1px solid #4CAF50;
				color: white;
			}
			.modal-button-ok:hover {
				background: #45a049;
				border-color: #45a049;
			}
		</style>
		
		{{if .IsOpen}}
		<div class="modal-overlay" onclick="send_event('{{.IdComponent}}', 'OverlayClick', '')">
			<div class="modal-container {{.Size}}" onclick="event.stopPropagation()">
				<div class="modal-header">
					<h2 class="modal-title">{{.Title}}</h2>
					{{if .Closable}}
					<button class="modal-close" onclick="send_event('{{.IdComponent}}', 'Close', '')">
						×
					</button>
					{{end}}
				</div>
				<div class="modal-body">
					{{.Content}}
				</div>
				{{if .ShowFooter}}
				<div class="modal-footer">
					<button class="modal-button modal-button-cancel" onclick="send_event('{{.IdComponent}}', 'Cancel', '')">
						{{.CancelText}}
					</button>
					<button class="modal-button modal-button-ok" onclick="send_event('{{.IdComponent}}', 'Ok', '')">
						{{.OkText}}
					</button>
				</div>
				{{end}}
			</div>
		</div>
		{{end}}
	</div>
	`
}

func (m *Modal) GetDriver() liveview.LiveDriver {
	return m
}

func (m *Modal) Open() {
	m.IsOpen = true
	m.Commit()
}

func (m *Modal) Close(data interface{}) {
	m.IsOpen = false
	if m.OnClose != nil {
		m.OnClose()
	}
	m.Commit()
}

func (m *Modal) OverlayClick(data interface{}) {
	if m.Closable {
		m.Close(nil)
	}
}

func (m *Modal) Ok(data interface{}) {
	if m.OnOk != nil {
		m.OnOk()
	}
	m.IsOpen = false
	m.Commit()
}

func (m *Modal) Cancel(data interface{}) {
	if m.OnCancel != nil {
		m.OnCancel()
	}
	m.IsOpen = false
	m.Commit()
}

func (m *Modal) SetContent(content string) {
	m.Content = content
	m.Commit()
}

func (m *Modal) SetTitle(title string) {
	m.Title = title
	m.Commit()
}

func (m *Modal) Show(title, content string) {
	m.Title = title
	m.Content = content
	m.IsOpen = true
	m.Commit()
}

func (m *Modal) Hide() {
	m.IsOpen = false
	m.Commit()
}

func (m *Modal) Toggle() {
	m.IsOpen = !m.IsOpen
	m.Commit()
}