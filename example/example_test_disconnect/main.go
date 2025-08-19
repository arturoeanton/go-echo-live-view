package main

import (
	"fmt"
	"github.com/arturoeanton/go-echo-live-view/liveview"
	"github.com/labstack/echo/v4"
)

type TestDisconnect struct {
	*liveview.ComponentDriver[*TestDisconnect]
	Counter int
}

func (c *TestDisconnect) Start() {
	c.Counter = 0
	c.Commit()
}

func (c *TestDisconnect) GetTemplate() string {
	// IMPORTANTE: El div más externo debe tener id="content" 
	// para que el WASM pueda mostrar "Disconnected" cuando se pierde la conexión
	return fmt.Sprintf(`
	<div id="content">
		<div style="padding: 2rem; font-family: Arial, sans-serif;">
			<h1>Test Disconnect Behavior</h1>
			<p>This example shows what happens when the WebSocket disconnects.</p>
			
			<div style="background: #f0f0f0; padding: 2rem; border-radius: 8px; margin: 2rem 0;">
				<h2>Counter: %d</h2>
				<button onclick="send_event('%s', 'Increment', '')" style="padding: 0.5rem 1rem; background: #4CAF50; color: white; border: none; border-radius: 4px; cursor: pointer;">
					Increment
				</button>
			</div>
			
			<div style="background: #fff3cd; padding: 1rem; border-radius: 4px; border: 1px solid #ffc107;">
				<strong>Instructions:</strong>
				<ol>
					<li>Click the Increment button a few times to verify it works</li>
					<li>Stop the server (Ctrl+C in terminal)</li>
					<li>The page should show "Disconnected"</li>
				</ol>
			</div>
		</div>
	</div>
	`, c.Counter, c.IdComponent)
}

func (c *TestDisconnect) Increment(data interface{}) {
	c.Counter++
	c.Commit()
}

func (c *TestDisconnect) GetDriver() liveview.LiveDriver {
	return c
}

func main() {
	liveview.InitLogger(true)
	fmt.Println("Starting Test Disconnect Server...")
	
	e := echo.New()
	e.Static("/assets", "assets")
	
	home := liveview.PageControl{
		Title:  "Test Disconnect",
		Lang:   "en",
		Path:   "/",
		Router: e,
		Debug:  true,
	}
	
	home.Register(func() liveview.LiveDriver {
		return liveview.NewDriver("test_disconnect", &TestDisconnect{})
	})
	
	fmt.Println("===========================================")
	fmt.Println("Server: http://localhost:8090")
	fmt.Println("===========================================")
	fmt.Println("Test disconnect behavior:")
	fmt.Println("1. Open browser to http://localhost:8090")
	fmt.Println("2. Stop server with Ctrl+C")
	fmt.Println("3. Page should show 'Disconnected'")
	fmt.Println()
	
	e.Logger.Fatal(e.Start(":8090"))
}