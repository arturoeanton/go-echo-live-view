package liveview

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"text/template"

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
		const go = new Go();
		WebAssembly.instantiateStreaming(fetch("assets/json.wasm"), go.importObject).then((result) => {
			go.run(result.instance);
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
				delete(Layaouts, id)
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
		upgrader := websocket.Upgrader{}
		ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
		if err != nil {
			return err
		}
		defer ws.Close()

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
				fmt.Println(string(msg))
			}
			var data map[string]interface{}
			json.Unmarshal(msg, &data)
			if mtype, ok := data["type"]; ok {
				if mtype == "data" {
					param := data["data"]
					drivers[data["id"].(string)].ExecuteEvent(data["event"].(string), param)
				}
				if mtype == "get" {
					param := data["data"]
					channelIn[data["id_ret"].(string)] <- param
				}
			}
		}
	})
}
