package main

import (
	"fmt"
	"strings"
	"sync"

	"github.com/arturoeanton/go-echo-live-view/components"
	"github.com/arturoeanton/go-echo-live-view/liveview"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var (
	counterMutex  = &sync.Mutex{}
	globalCounter = 0
	clientCounter = 0
)

func main() {
	e := echo.New()
	e.Use(middleware.Logger(), middleware.Recover())
	
	home := liveview.PageControl{
		Title:  "Shared Counter Example",
		Lang:   "en",
		Path:   "/",
		Router: e,
	}
	
	home.Register(func() liveview.LiveDriver {
		counterMutex.Lock()
		clientCounter++
		clientID := fmt.Sprintf("Client_%d", clientCounter)
		counterMutex.Unlock()
		
		// Use Layout for broadcast capability
		layout := liveview.NewLayout(clientID, `
			<div style="font-family: Arial, sans-serif; padding: 2rem; max-width: 800px; margin: 0 auto;">
				<h1>🌐 Shared Counter Example</h1>
				
				<div style="background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; padding: 2rem; border-radius: 12px; margin: 2rem 0;">
					<div style="text-align: center;">
						<div style="font-size: 1.2rem; margin-bottom: 1rem;">Your ID: {{.UUID}}</div>
						<div style="font-size: 4rem; font-weight: bold;" id="counter_display">0</div>
						<div style="font-size: 1rem; margin-top: 1rem;">Global Counter</div>
					</div>
				</div>
				
				<div style="display: flex; gap: 1rem; justify-content: center; margin: 2rem 0;">
					{{mount "button_increment"}}
					{{mount "button_decrement"}}
					{{mount "button_reset"}}
				</div>
				
				<div style="background: #f0f0f0; padding: 1.5rem; border-radius: 8px;">
					<h3>Send Message to All:</h3>
					<div style="display: flex; gap: 1rem;">
						{{mount "input_message"}}
						{{mount "button_send"}}
					</div>
				</div>
				
				<div id="messages" style="margin-top: 2rem; padding: 1rem; background: white; border: 1px solid #ddd; border-radius: 8px; min-height: 100px;">
					<strong>Messages:</strong><br/>
				</div>
			</div>`)
		
		// Create components
		liveview.New("button_increment", &components.Button{Caption: "➕ Increment"}).
			SetClick(func(btn *components.Button, data interface{}) {
				counterMutex.Lock()
				globalCounter++
				newValue := globalCounter
				counterMutex.Unlock()
				
				// Broadcast counter update to all layouts
				liveview.SendToAllLayouts(fmt.Sprintf("COUNTER|%d", newValue))
			})
		
		liveview.New("button_decrement", &components.Button{Caption: "➖ Decrement"}).
			SetClick(func(btn *components.Button, data interface{}) {
				counterMutex.Lock()
				globalCounter--
				newValue := globalCounter
				counterMutex.Unlock()
				
				// Broadcast counter update to all layouts
				liveview.SendToAllLayouts(fmt.Sprintf("COUNTER|%d", newValue))
			})
		
		liveview.New("button_reset", &components.Button{Caption: "🔄 Reset"}).
			SetClick(func(btn *components.Button, data interface{}) {
				counterMutex.Lock()
				globalCounter = 0
				counterMutex.Unlock()
				
				// Broadcast counter update and reset message
				liveview.SendToAllLayouts("COUNTER|0")
				liveview.SendToAllLayouts(fmt.Sprintf("MESSAGE|%s reset the counter!", clientID))
			})
		
		inputMessage := liveview.New("input_message", &components.InputText{})
		
		liveview.New("button_send", &components.Button{Caption: "Send Message"}).
			SetClick(func(btn *components.Button, data interface{}) {
				message := inputMessage.GetValue()
				if message != "" {
					// Broadcast message to all layouts
					liveview.SendToAllLayouts(fmt.Sprintf("MESSAGE|%s: %s", clientID, message))
					inputMessage.SetValue("")
				}
			})
		
		// Handle incoming broadcast messages
		layout.Component.SetHandlerEventIn(func(data interface{}) {
			msg := data.(string)
			
			// Handle counter updates
			if strings.HasPrefix(msg, "COUNTER|") {
				value := strings.TrimPrefix(msg, "COUNTER|")
				counterDisplay := layout.GetDriverById("counter_display")
				counterDisplay.FillValue(value)
			}
			
			// Handle messages
			if strings.HasPrefix(msg, "MESSAGE|") {
				message := strings.TrimPrefix(msg, "MESSAGE|")
				messagesDiv := layout.GetDriverById("messages")
				currentHTML := messagesDiv.GetHTML()
				messagesDiv.FillValue(fmt.Sprintf("%s%s<br/>", currentHTML, message))
			}
			
			// Handle user joined
			if strings.HasPrefix(msg, "JOINED|") {
				user := strings.TrimPrefix(msg, "JOINED|")
				messagesDiv := layout.GetDriverById("messages")
				currentHTML := messagesDiv.GetHTML()
				messagesDiv.FillValue(fmt.Sprintf("%s<span style='color: green;'>✅ %s joined</span><br/>", currentHTML, user))
			}
			
			// Handle user left
			if strings.HasPrefix(msg, "LEFT|") {
				user := strings.TrimPrefix(msg, "LEFT|")
				messagesDiv := layout.GetDriverById("messages")
				currentHTML := messagesDiv.GetHTML()
				messagesDiv.FillValue(fmt.Sprintf("%s<span style='color: red;'>❌ %s left</span><br/>", currentHTML, user))
			}
		})
		
		// Set initial counter value
		layout.Component.SetHandlerFirstTime(func() {
			counterDisplay := layout.GetDriverById("counter_display")
			counterDisplay.FillValue(fmt.Sprint(globalCounter))
			
			// Notify others that this user joined
			liveview.SendToAllLayouts(fmt.Sprintf("JOINED|%s", clientID))
		})
		
		// Handle disconnect
		layout.Component.SetHandlerEventDestroy(func(id string) {
			// Notify others that this user left
			liveview.SendToAllLayouts(fmt.Sprintf("LEFT|%s", clientID))
		})
		
		return layout
	})
	
	fmt.Println("===========================================")
	fmt.Println("Server: http://localhost:8092")
	fmt.Println("===========================================")
	fmt.Println("Open multiple browser windows to see shared state!")
	fmt.Println()
	
	e.Logger.Fatal(e.Start(":8092"))
}