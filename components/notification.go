package components

import (
	"fmt"
	"time"

	"github.com/arturoeanton/go-echo-live-view/liveview"
)

type NotificationType string

const (
	NotificationSuccess NotificationType = "success"
	NotificationError   NotificationType = "error"
	NotificationWarning NotificationType = "warning"
	NotificationInfo    NotificationType = "info"
)

type Notification struct {
	ID       string
	Type     NotificationType
	Title    string
	Message  string
	Duration int
	Show     bool
}

type NotificationSystem struct {
	*liveview.ComponentDriver[*NotificationSystem]
	Notifications []Notification
	Position      string
	MaxVisible    int
}

func (n *NotificationSystem) Start() {
	if n.Position == "" {
		n.Position = "top-right"
	}
	if n.MaxVisible == 0 {
		n.MaxVisible = 5
	}
	n.Commit()
}

func (n *NotificationSystem) GetTemplate() string {
	return `
	<div id="{{.IdComponent}}" class="notification-system {{.Position}}">
		<style>
			.notification-system {
				position: fixed;
				z-index: 2000;
				pointer-events: none;
				display: flex;
				flex-direction: column;
				gap: 0.75rem;
				padding: 1rem;
			}
			.notification-system.top-right {
				top: 0;
				right: 0;
			}
			.notification-system.top-left {
				top: 0;
				left: 0;
			}
			.notification-system.bottom-right {
				bottom: 0;
				right: 0;
				flex-direction: column-reverse;
			}
			.notification-system.bottom-left {
				bottom: 0;
				left: 0;
				flex-direction: column-reverse;
			}
			.notification-system.top-center {
				top: 0;
				left: 50%;
				transform: translateX(-50%);
			}
			.notification-system.bottom-center {
				bottom: 0;
				left: 50%;
				transform: translateX(-50%);
				flex-direction: column-reverse;
			}
			.notification-item {
				background: white;
				border-radius: 8px;
				box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
				padding: 1rem;
				min-width: 300px;
				max-width: 400px;
				pointer-events: auto;
				display: flex;
				align-items: flex-start;
				gap: 0.75rem;
				animation: slideIn 0.3s ease;
				position: relative;
				border-left: 4px solid;
			}
			@keyframes slideIn {
				from {
					transform: translateX(100%);
					opacity: 0;
				}
				to {
					transform: translateX(0);
					opacity: 1;
				}
			}
			@keyframes slideOut {
				from {
					transform: translateX(0);
					opacity: 1;
				}
				to {
					transform: translateX(100%);
					opacity: 0;
				}
			}
			.notification-item.hiding {
				animation: slideOut 0.3s ease;
			}
			.notification-item.success {
				border-left-color: #4CAF50;
			}
			.notification-item.error {
				border-left-color: #f44336;
			}
			.notification-item.warning {
				border-left-color: #ff9800;
			}
			.notification-item.info {
				border-left-color: #2196F3;
			}
			.notification-icon {
				font-size: 1.25rem;
				flex-shrink: 0;
			}
			.notification-item.success .notification-icon {
				color: #4CAF50;
			}
			.notification-item.error .notification-icon {
				color: #f44336;
			}
			.notification-item.warning .notification-icon {
				color: #ff9800;
			}
			.notification-item.info .notification-icon {
				color: #2196F3;
			}
			.notification-content {
				flex: 1;
			}
			.notification-title {
				font-weight: 600;
				margin-bottom: 0.25rem;
				color: #333;
			}
			.notification-message {
				font-size: 0.875rem;
				color: #666;
				line-height: 1.4;
			}
			.notification-close {
				position: absolute;
				top: 0.5rem;
				right: 0.5rem;
				background: none;
				border: none;
				color: #999;
				cursor: pointer;
				font-size: 1.25rem;
				padding: 0.25rem;
				line-height: 1;
				transition: color 0.2s;
			}
			.notification-close:hover {
				color: #333;
			}
			.notification-progress {
				position: absolute;
				bottom: 0;
				left: 0;
				right: 0;
				height: 3px;
				background: rgba(0, 0, 0, 0.1);
				overflow: hidden;
			}
			.notification-progress-bar {
				height: 100%;
				background: currentColor;
				animation: progress linear forwards;
			}
			@keyframes progress {
				from { width: 100%; }
				to { width: 0%; }
			}
		</style>
		
		{{range $i, $notif := .GetVisibleNotifications}}
		{{if $notif.Show}}
		<div class="notification-item {{$notif.Type}}" id="notif_{{$notif.ID}}">
			<div class="notification-icon">
				{{if eq $notif.Type "success"}}✓{{end}}
				{{if eq $notif.Type "error"}}✕{{end}}
				{{if eq $notif.Type "warning"}}⚠{{end}}
				{{if eq $notif.Type "info"}}ℹ{{end}}
			</div>
			<div class="notification-content">
				{{if $notif.Title}}
				<div class="notification-title">{{$notif.Title}}</div>
				{{end}}
				{{if $notif.Message}}
				<div class="notification-message">{{$notif.Message}}</div>
				{{end}}
			</div>
			<button class="notification-close" onclick="send_event('{{$.IdComponent}}', 'Close', '{{$notif.ID}}')">
				×
			</button>
			{{if gt $notif.Duration 0}}
			<div class="notification-progress">
				<div class="notification-progress-bar" style="animation-duration: {{$notif.Duration}}ms; color: {{$.GetProgressColor $notif.Type}}"></div>
			</div>
			{{end}}
		</div>
		{{end}}
		{{end}}
	</div>
	`
}

func (n *NotificationSystem) GetDriver() liveview.LiveDriver {
	return n
}

func (n *NotificationSystem) Show(notifType NotificationType, title, message string, duration int) {
	id := fmt.Sprintf("notif_%d", time.Now().UnixNano())
	
	notification := Notification{
		ID:       id,
		Type:     notifType,
		Title:    title,
		Message:  message,
		Duration: duration,
		Show:     true,
	}
	
	n.Notifications = append(n.Notifications, notification)
	
	if duration > 0 {
		go func(notifID string) {
			time.Sleep(time.Duration(duration) * time.Millisecond)
			n.hideNotification(notifID)
		}(id)
	}
	
	n.Commit()
}

func (n *NotificationSystem) Close(data interface{}) {
	id := fmt.Sprint(data)
	n.hideNotification(id)
}

func (n *NotificationSystem) hideNotification(id string) {
	for i, notif := range n.Notifications {
		if notif.ID == id {
			n.Notifications[i].Show = false
			n.Commit()
			
			go func() {
				time.Sleep(300 * time.Millisecond)
				n.removeNotification(id)
			}()
			break
		}
	}
}

func (n *NotificationSystem) removeNotification(id string) {
	newNotifications := []Notification{}
	for _, notif := range n.Notifications {
		if notif.ID != id {
			newNotifications = append(newNotifications, notif)
		}
	}
	n.Notifications = newNotifications
	n.Commit()
}

func (n *NotificationSystem) GetVisibleNotifications() []Notification {
	visible := []Notification{}
	count := 0
	
	for i := len(n.Notifications) - 1; i >= 0 && count < n.MaxVisible; i-- {
		if n.Notifications[i].Show {
			visible = append([]Notification{n.Notifications[i]}, visible...)
			count++
		}
	}
	
	return visible
}

func (n *NotificationSystem) GetProgressColor(notifType NotificationType) string {
	switch notifType {
	case NotificationSuccess:
		return "#4CAF50"
	case NotificationError:
		return "#f44336"
	case NotificationWarning:
		return "#ff9800"
	case NotificationInfo:
		return "#2196F3"
	default:
		return "#999"
	}
}

func (n *NotificationSystem) Success(title, message string) {
	n.Show(NotificationSuccess, title, message, 5000)
}

func (n *NotificationSystem) Error(title, message string) {
	n.Show(NotificationError, title, message, 0)
}

func (n *NotificationSystem) Warning(title, message string) {
	n.Show(NotificationWarning, title, message, 7000)
}

func (n *NotificationSystem) Info(title, message string) {
	n.Show(NotificationInfo, title, message, 5000)
}

func (n *NotificationSystem) Clear() {
	n.Notifications = []Notification{}
	n.Commit()
}

func (n *NotificationSystem) SetPosition(position string) {
	n.Position = position
	n.Commit()
}