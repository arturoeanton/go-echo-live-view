package main

/*
cd cmd/wasm/
#cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" ../../assets/
GOOS=js GOARCH=wasm go build -o  ../../assets/json.wasm
cd -
*/

import (
	"encoding/json"
	"fmt"
	"syscall/js"
)

var (
	document  js.Value = js.Global().Get("document")
	window    js.Value = js.Global().Get("window")
	console   js.Value = js.Global().Get("console")
	webSocket js.Value = js.Global().Get("WebSocket")
	loc       js.Value = window.Get("location")
	uri       string   = "ws:"
	ws        js.Value
	protocol  string = loc.Get("protocol").String()
)

type MsgEvent struct {
	Type  string `json:"type"`
	ID    string `json:"id"`
	Event string `json:"event"`
	Data  string `json:"data"`
}

type DataEventIn struct {
	ID        string      `json:"id"`
	IdRet     string      `json:"id_ret"`
	Type      string      `json:"type"`
	Value     interface{} `json:"value"`
	Propertie string      `json:"propertie"`
	SubType   string      `json:"sub_type"`
}

type DataEventOut struct {
	Type  string      `json:"type"`
	IdRet string      `json:"id_ret"`
	Data  interface{} `json:"data"`
}

func connect() {
	document = js.Global().Get("document")
	window = js.Global().Get("window")
	console = js.Global().Get("console")
	webSocket = js.Global().Get("WebSocket")
	loc = window.Get("location")
	uri = "ws:"
	protocol = loc.Get("protocol").String()

	fmt.Println("Go Web LiveView")
	if protocol == "https:" {
		uri = "wss:"
	}
	fmt.Println("protocol: " + protocol + " uri: " + uri)
	uri += "//" + loc.Get("host").String()
	uri += loc.Get("pathname").String() + "ws_goliveview"
	ws = webSocket.New(uri)

	handlerOnOpen := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		fmt.Println(ws.Get("readyState").Int())
		fmt.Println("Connected...ok!!")
		return nil
	})

	handlerOnClose := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("Recovered in f", r)
			}
		}()
		fmt.Println(ws)
		fmt.Println("Disconnected...ok")
		document.Call("getElementById", "content").Set("innerHTML", "Disconnected")
		return nil
	})

	handlerOnMessage := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		evtData := args[0].Get("data").String()
		var dataEventIn DataEventIn
		json.Unmarshal([]byte(evtData), &dataEventIn)
		currentElement := document.Call("getElementById", dataEventIn.ID)

		if currentElement.IsNull() {
			return nil
		}

		if dataEventIn.Type == "fill" {
			fmt.Println("fill")
			currentElement.Set("innerHTML", dataEventIn.Value)
			return nil
		}

		if dataEventIn.Type == "remove" {
			currentElement.Call("remove")
		}

		if dataEventIn.Type == "addNode" {
			var d = document.Call("createElement", "div")
			currentElement.Set("innerHTML", fmt.Sprint(dataEventIn.Value))
			currentElement.Call("appendChild", d)
		}

		if dataEventIn.Type == "text" {
			if dataEventIn.Value != "" {
				currentElement.Set("innerText", dataEventIn.Value)
			}
		}

		if dataEventIn.Type == "style" {
			currentElement.Get("style").Set("cssText", dataEventIn.Value)
		}

		if dataEventIn.Type == "set" {
			currentElement.Set("value", dataEventIn.Value)
		}

		if dataEventIn.Type == "script" {
			currentElement.Call("eval", dataEventIn.Value)
		}

		if dataEventIn.Type == "propertie" {
			currentElement.Set(dataEventIn.Propertie, dataEventIn.Value)
		}

		if dataEventIn.Type == "get" {
			dataEventOut := DataEventOut{}
			dataEventOut.Type = "get"
			dataEventOut.IdRet = dataEventIn.IdRet
			if dataEventIn.SubType == "value" {
				value := currentElement.Get("value")
				dataEventOut.Data = GetValue(value)
			}
			if dataEventIn.SubType == "html" {
				value := currentElement.Get("innerHTML")
				dataEventOut.Data = GetValue(value)
			}
			if dataEventIn.SubType == "text" {
				value := currentElement.Get("innerText")
				dataEventOut.Data = GetValue(value)
			}
			if dataEventIn.SubType == "style" {
				value := currentElement.Get("style").Get(fmt.Sprint(dataEventIn.Value))
				dataEventOut.Data = GetValue(value)
			}

			if dataEventIn.SubType == "propertie" {
				prop := currentElement.Get(fmt.Sprint(dataEventIn.Value))
				dataEventOut.Data = GetValue(prop)
			}
			jsonBytes, _ := json.Marshal(&dataEventOut)
			ws.Call("send", string(jsonBytes))
		}
		return nil
	})

	fmt.Println("Set handlers...??")
	ws.Set("onclose", handlerOnClose)
	ws.Set("onopen", handlerOnOpen)
	ws.Set("onmessage", handlerOnMessage)
	fmt.Println("Set handlers...ok!!")

}

func main() {
	document.Call("getElementById", "content").Set("innerHTML", "Disconnected")
	connect()

	js.Global().Call("setInterval", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if ws.Get("readyState").Int() != 1 {
			connect()
		}
		return nil
	}), 1000)

	js.Global().Set("ws", ws)
	js.Global().Set("connect", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		connect()
		return nil
	}))

	js.Global().Set("send_event", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		id := args[0].String()
		event := args[1].String()
		data := ""
		if len(args) == 3 {
			data = args[2].String()
		}
		sendEvent(id, event, data)
		return nil
	}))
	<-make(chan struct{})
}

func GetValue(prop js.Value) interface{} {
	switch prop.Type() {
	case js.TypeBoolean:
		return prop.Bool()
	case js.TypeNumber:
		return prop.Int()
	case js.TypeNull:
		return nil
	case js.TypeString:
		return prop.String()
	case js.TypeUndefined:
		return nil
	}
	return nil
}

func sendEvent(id string, event string, data string) {
	msgEvent := MsgEvent{
		Type:  "data",
		ID:    id,
		Event: event,
		Data:  data,
	}
	jsonMsg, _ := json.Marshal(&msgEvent)
	ws.Call("send", string(jsonMsg))
}
