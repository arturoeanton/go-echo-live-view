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
	Path     string
	Title    string
	HeadCode string
	Lang     string
	Css      string
	LiveJs   string
	Router   *echo.Echo
	Debug    bool
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
		var loc=window.location,uri="ws:";function send_event(t,e,a){var n=JSON.stringify({type:"data",id:t,event:e,data:a});ws.send(n)}"https:"===loc.protocol&&(uri="wss:"),uri+="//"+loc.host,uri+=loc.pathname+"ws_goliveview",ws=new WebSocket(uri),ws.onopen=function(){console.log("Connected")},ws.onmessage=function(evt){json_data=JSON.parse(evt.data);var out=document.getElementById(json_data.id);"fill"==json_data.type&&(out.innerHTML=json_data.value),"text"==json_data.type&&(out.innerText=json_data.value),"propertie"==json_data.type&&(out[json_data.propertie]=json_data.value),"style"==json_data.type&&(out.style.cssText=json_data.value),"set"==json_data.type&&(out.value=json_data.value),"script"==json_data.type&&eval(json_data.value),"get"==json_data.type&&(str=JSON.stringify({type:"get",id_ret:json_data.id_ret,data:null}),"style"==json_data.sub_type&&(str=JSON.stringify({type:"get",id_ret:json_data.id_ret,data:document.getElementById(json_data.id).style[json_data.value]})),"value"==json_data.sub_type&&(str=JSON.stringify({type:"get",id_ret:json_data.id_ret,data:document.getElementById(json_data.id).value})),"html"==json_data.sub_type&&(str=JSON.stringify({type:"get",id_ret:json_data.id_ret,data:document.getElementById(json_data.id).innerHTML})),"text"==json_data.sub_type&&(str=JSON.stringify({type:"get",id_ret:json_data.id_ret,data:document.getElementById(json_data.id).innerHTML})),"propertie"==json_data.sub_type&&(str=JSON.stringify({type:"get",id_ret:json_data.id_ret,data:document.getElementById(json_data.id)[json_data.value]})),ws.send(str))};		
		</script>
    </body>
</html>
`
)

//Register this method to register in router of Echo page and websocket
func (pc *PageControl) Register(fx func() *ComponentDriver) {
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
		content.SetID("content")

		channel := make(chan (map[string]interface{}))
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
