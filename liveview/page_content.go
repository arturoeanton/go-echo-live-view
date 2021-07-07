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
	Path     string
	Title    string
	HeadCode string
	Lang     string
	Css      string
	Router   *echo.Echo
}

var (
	//websockets map[string]websocket.Upgrader = make(map[string]websocket.Upgrader)
	html string = `
<html lang="{{.Lang}}">
	<head>
		<title>{{.Title}}</title>
		{{.HeadCode}}

		<style>
			{{.Css}}
		</style>
	</head>
    <body>
		<div id="content"> 
		</div>
		<script>
			var loc = window.location;
			var uri = 'ws:';

			if (loc.protocol === 'https:') {
				uri = 'wss:';
			}
			uri += '//' + loc.host;
			uri += loc.pathname + 'ws_goliveview';

			ws = new WebSocket(uri)

			ws.onopen = function() {
				console.log('Connected')
			}

			ws.onmessage = function(evt) {
				json_data = JSON.parse(evt.data)
				var out = document.getElementById(json_data.id);

				if (json_data.type == 'fill'){
					out.innerHTML = json_data.value ;
				}

				if (json_data.type == 'set'){
					out.value = json_data.value ;
				}


				if (json_data.type == 'script'){
					eval(json_data.value);
				}

				if (json_data.type == 'get'){
					var str = JSON.stringify({"type":"get", "id_ret": json_data.id_ret , "data":document.getElementById(json_data.id).value})
					ws.send(str)
				}

				


			}	

			function send_event (id, event, data) {
				var str = JSON.stringify({"type":"data","id": id, "event":event, "data":data})
				ws.send(str)
			}
		</script>
    </body>
</html>
`
)

func (pc *PageControl) Register(fx func() *ComponentDriver) {
	pc.Router.GET(pc.Path, func(c echo.Context) error {
		t := template.Must(template.New("page_control").Parse(html))
		buf := new(bytes.Buffer)
		_ = t.Execute(buf, pc)
		c.HTML(http.StatusOK, buf.String())

		return nil
	})

	pc.Router.GET(pc.Path+"ws_goliveview", func(c echo.Context) error {

		content := fx()
		content.SetID("content")

		channel := make(chan (map[string]string))
		upgrader := websocket.Upgrader{}
		ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
		if err != nil {
			return err
		}
		defer ws.Close()

		drivers := make(map[string]*ComponentDriver)
		channelIn := make(map[string](chan interface{}))

		go content.Start(&drivers, &channelIn, channel)

		go func() {
			for {
				data := <-channel
				ws.WriteJSON(data)
			}
		}()

		for {
			_, msg, err := ws.ReadMessage()
			if err != nil {
				c.Logger().Error(err)
			}
			fmt.Println(string(msg))

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
