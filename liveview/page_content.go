package liveview

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"text/template"

	"github.com/arturoeanton/gocommons/utils"
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
		var uri = "ws:";
		
		if (loc.protocol === "https:") {
			uri = "wss:";
		}
		uri += "//" + loc.host;
		uri += loc.pathname + "ws_goliveview";
		
		ws = new WebSocket(uri);
		
		ws.onopen = function () {
			console.log("Connected");
		};
		
		ws.onmessage = function (evt) {
			json_data = JSON.parse(evt.data);
			var out = document.getElementById(json_data.id);
		
			if (json_data.type == "fill") {
				try{
				out.innerHTML = json_data.value;
				}catch(e){
					console.log(out);
					console.log(json_data.id);
				}
			}
		
			if (json_data.type == "remove") {
				out.remove();
			}
		
			if (json_data.type == "addNode") {
				var d = document.createElement("div");
				d.innerHTML = json_data.value;
				out.appendChild(d);
			}
		
			if (json_data.type == "text") {
				out.innerText = json_data.value;
			}
		
			if (json_data.type == "propertie") {
				out[json_data.propertie] = json_data.value;
			}
		
			if (json_data.type == "style") {
				out.style.cssText = json_data.value;
			}
		
			if (json_data.type == "set") {
				out.value = json_data.value;
			}
		
			if (json_data.type == "script") {
				eval(json_data.value);
			}
		
			if (json_data.type == "get") {
				str = JSON.stringify({
					type: "get",
					id_ret: json_data.id_ret,
					data: null,
				});
				if (json_data.sub_type == "style") {
					str = JSON.stringify({
						type: "get",
						id_ret: json_data.id_ret,
						data: document.getElementById(json_data.id).style[json_data.value],
					});
				}
				if (json_data.sub_type == "value") {
					str = JSON.stringify({
						type: "get",
						id_ret: json_data.id_ret,
						data: document.getElementById(json_data.id).value,
					});
				}
				if (json_data.sub_type == "html") {
					str = JSON.stringify({
						type: "get",
						id_ret: json_data.id_ret,
						data: document.getElementById(json_data.id).innerHTML,
					});
				}
				if (json_data.sub_type == "text") {
					str = JSON.stringify({
						type: "get",
						id_ret: json_data.id_ret,
						data: document.getElementById(json_data.id).innerHTML,
					});
				}
				if (json_data.sub_type == "propertie") {
					str = JSON.stringify({
						type: "get",
						id_ret: json_data.id_ret,
						data: document.getElementById(json_data.id)[json_data.value],
					});
				}
				ws.send(str);
			}
		};
		
		function send_event(id, event, data) {
			var str = JSON.stringify({
				type: "data",
				id: id,
				event: event,
				data: data,
			});
			ws.send(str);
		}		</script>
		{{.AfterCode}}
    </body>
</html>
`
)

//Register this method to register in router of Echo page and websocket
func (pc *PageControl) Register(fx func() LiveDriver) {
	if utils.Exists(pc.AfterCode) {
		pc.AfterCode, _ = utils.FileToString(pc.AfterCode)
	}
	if utils.Exists(pc.HeadCode) {
		pc.HeadCode, _ = utils.FileToString(pc.HeadCode)
	}
	if pc.Lang == "" {
		pc.Lang = "en"
	}
	if utils.Exists("live.js") {
		pc.LiveJs, _ = utils.FileToString("live.js")
	}
	pc.Router.GET(pc.Path, func(c echo.Context) error {
		t := template.Must(template.New("page_control").Parse(html))
		buf := new(bytes.Buffer)
		_ = t.Execute(buf, pc)
		c.HTML(http.StatusOK, buf.String())

		return nil
	})

	pc.Router.GET(pc.Path+"ws_goliveview", func(c echo.Context) error {

		content := fx()
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

		go content.StartDriver(&drivers, &channelIn, channel)

		go func() {
			for {
				data := <-channel
				ws.WriteJSON(data)
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
