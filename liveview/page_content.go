package liveview

import (
	"bytes"
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

		channel := make(chan (map[string]interface{}))
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
		channelIn := make(map[string](chan interface{}))

		go func() {
			defer HandleReover()
			content.StartDriver(&drivers, &channelIn, channel)
		}()
		end := make(chan bool)
		defer func() {
			end <- true
		}()
		go func() {
			defer HandleReover()
			for {
				select {
				case data := <-channel:
					if pc.Debug {
						if dataMap, ok := data["type"]; ok && dataMap == "script" {
							LogWebSocket("Sending", "script message")
						}
					}
					ws.WriteJSON(data)
				case <-end:
					return
				}
			}
		}()

		for {
			_, msg, err := ws.ReadMessage()
			if err != nil {
				//c.Logger().Error(err)
				return nil
			}
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
					ch <- validatedMsg.Data
				} else {
					Warn("Channel not found: %s", validatedMsg.IdRet)
				}
			}
		}
	})
}
