package liveview

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"text/template"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

type PageControl struct {
	Path      string
	Title     string
	HeadCode  string
	Lang      string
	Css       string
	LiveJs    string
	AfterCode string
	Router    *echo.Echo
	Debug     bool
}

var (
	templateBase string = `
<html lang="{{.Lang}}">
	<head>
		<title>{{.Title}}</title>
		{{.HeadCode}}
		<style>
			{{.Css}}
		</style>
		<meta charset="utf-8"/>
        <script src="assets/wasm_exec.js"></script>
	</head>
    <body>
		<div id="content"> 
		</div>
		<script>
		// Global send_event stub until WASM loads
		window.send_event = function(id, event, data) {
			console.log('[Pre-WASM] Event queued:', id, event, data);
			// Queue events until WASM is ready
			if (!window._eventQueue) window._eventQueue = [];
			window._eventQueue.push({id, event, data});
		};
		
		const go = new Go();
		WebAssembly.instantiateStreaming(fetch("assets/json.wasm?v=" + Date.now()), go.importObject).then((result) => {
			go.run(result.instance);
			console.log('[WASM] Loaded successfully');
			
			// Process queued events
			if (window._eventQueue && window._eventQueue.length > 0) {
				console.log('[WASM] Processing', window._eventQueue.length, 'queued events');
				window._eventQueue.forEach(e => {
					window.send_event(e.id, e.event, e.data);
				});
				window._eventQueue = [];
			}
		}).catch(err => {
			console.error('[WASM] Failed to load:', err);
		});
		</script>
		{{.AfterCode}}
    </body>
</html>
`
)

// Register this method to register in router of Echo page and websocket
func (pc *PageControl) Register(fx func() LiveDriver) {
	if Exists(pc.AfterCode) {
		pc.AfterCode, _ = FileToString(pc.AfterCode)
	}
	if Exists(pc.HeadCode) {
		pc.HeadCode, _ = FileToString(pc.HeadCode)
	}
	if pc.Lang == "" {
		pc.Lang = "en"
	}
	if Exists("live.js") {
		pc.LiveJs, _ = FileToString("live.js")
	}

	pc.Router.Static("/assets", "assets")
	pc.Router.GET(pc.Path, func(c echo.Context) error {
		t := template.Must(template.New("page_control").Parse(templateBase))
		buf := new(bytes.Buffer)
		_ = t.Execute(buf, pc)
		c.HTML(http.StatusOK, buf.String())

		return nil
	})

	pc.Router.GET(pc.Path+"ws_goliveview", func(c echo.Context) error {

		content := fx()
		defer func() {
			func() {
				MuLayout.Lock()
				defer MuLayout.Unlock()
				id := content.GetIDComponet()
				delete(Layouts, id)
			}()
			func() {
				defer func() {
					if r := recover(); r != nil {
						fmt.Println("Layout has not HandlerEventDestroy method defined", r)
					}
				}()
				handlerEventDestroy := (content.GetComponet().(*Layout)).HandlerEventDestroy
				if handlerEventDestroy != nil {
					(*handlerEventDestroy)(content.GetIDComponet())
				}
			}()

			fmt.Println("Delete Layout:", content.GetIDComponet())
		}()
		for _, v := range componentsDrivers {
			content.Mount(v.GetComponet())
		}

		content.SetID("content")
		//content.SetIDComponent("content")

		// MEM-001: Crear channel con buffer y asegurar cierre
		channel := make(chan map[string]interface{}, 10)
		defer close(channel)
		
		upgrader := websocket.Upgrader{
			// SEC-005: Establecer límites de tamaño de mensaje
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		}
		ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
		if err != nil {
			return err
		}
		defer ws.Close()
		
		// SEC-005: Configurar límite máximo de mensaje
		ws.SetReadLimit(MaxMessageSize)
		
		// Crear rate limiter para este cliente
		rateLimiter := NewRateLimiter(100, 60) // 100 mensajes por minuto
		clientID := c.RealIP()

		drivers := make(map[string]LiveDriver)
		channelIn := make(map[string]chan interface{})
		
		// MEM-001: Usar context para cancelación coordinada
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		go func() {
			defer HandleReover()
			content.StartDriverWithContext(ctx, &drivers, &channelIn, channel)
		}()
		
		// MEM-001: Usar done channel sin buffer
		done := make(chan struct{})
		defer close(done)
		go func() {
			defer HandleReover()
			for {
				select {
				case data, ok := <-channel:
					if !ok {
						// Channel cerrado, salir
						return
					}
					if pc.Debug {
						if dataMap, ok := data["type"]; ok && dataMap == "script" {
							LogWebSocket("Sending", "script message")
						}
					}
					// MEM-004: Agregar timeout para escritura
					ws.SetWriteDeadline(time.Now().Add(10 * time.Second))
					if err := ws.WriteJSON(data); err != nil {
						Warn("Error writing to WebSocket: %v", err)
						return
					}
				case <-ctx.Done():
					return
				case <-done:
					return
				}
			}
		}()

		// MEM-004: Configurar timeouts para el WebSocket
		ws.SetReadDeadline(time.Now().Add(60 * time.Second))
		ws.SetPongHandler(func(string) error {
			ws.SetReadDeadline(time.Now().Add(60 * time.Second))
			return nil
		})
		
		// Ping periódico para mantener la conexión viva
		go func() {
			ticker := time.NewTicker(30 * time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					ws.SetWriteDeadline(time.Now().Add(10 * time.Second))
					if err := ws.WriteMessage(websocket.PingMessage, nil); err != nil {
						return
					}
				case <-ctx.Done():
					return
				}
			}
		}()
		
		for {
			_, msg, err := ws.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					Error("WebSocket error: %v", err)
				}
				return nil
			}
			// Reset read deadline on successful read
			ws.SetReadDeadline(time.Now().Add(60 * time.Second))
			if pc.Debug {
				LogWebSocket("Received", string(msg))
			}
			
			// Verificar rate limiting
			if !rateLimiter.Allow(clientID, time.Now().Unix()) {
				Warn("Rate limit exceeded for client: %s", clientID)
				continue
			}
			
			// SEC-002: Validar mensaje WebSocket
			validatedMsg, err := ValidateWebSocketMessage(msg)
			if err != nil {
				// Log error but don't crash
				Error("Invalid WebSocket message: %v", err)
				continue
			}
			
			if validatedMsg.Type == "data" {
				// Validar que el driver existe antes de ejecutar
				if driver, ok := drivers[validatedMsg.ID]; ok {
					LogEvent(validatedMsg.ID, validatedMsg.Event, validatedMsg.Data)
					driver.ExecuteEvent(validatedMsg.Event, validatedMsg.Data)
				} else {
					Warn("Driver not found: %s", validatedMsg.ID)
				}
			} else if validatedMsg.Type == "get" {
				// Validar que el canal existe
				if ch, ok := channelIn[validatedMsg.IdRet]; ok {
					// MEM-001: Usar select con timeout para evitar bloqueos
					select {
					case ch <- validatedMsg.Data:
					case <-time.After(5 * time.Second):
						Warn("Timeout sending to channel: %s", validatedMsg.IdRet)
					}
				} else {
					Warn("Channel not found: %s", validatedMsg.IdRet)
				}
			}
		}
	})
}
