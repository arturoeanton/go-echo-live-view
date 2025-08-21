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
	"strconv"
	"strings"
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
	isVerbose bool
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
		
		// Debug logging for non-fill messages
		if isVerbose && dataEventIn.Type != "fill" {
			fmt.Printf("[WASM] Received message type: %s, ID: %s\n", dataEventIn.Type, dataEventIn.ID)
		}
		
		// Scripts don't need an element ID
		if dataEventIn.Type == "script" {
			scriptValue := fmt.Sprint(dataEventIn.Value)
			if isVerbose {
				fmt.Printf("[WASM] Executing script: %d bytes\n", len(scriptValue))
			}
			result := js.Global().Call("eval", scriptValue)
			if isVerbose {
				fmt.Printf("[WASM] Script executed, result: %v\n", result)
			}
			return nil
		}
		
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
	// Configurar logging
	verbose := js.Global().Get("location").Get("search").String()
	isVerbose = strings.Contains(verbose, "verbose=true") || strings.Contains(verbose, "debug=true")
	
	if isVerbose {
		fmt.Println("[WASM] Verbose mode enabled")
	}
	
	document.Call("getElementById", "content").Set("innerHTML", "Disconnected")
	connect()

	js.Global().Call("setInterval", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if ws.Get("readyState").Int() != 1 {
			if isVerbose {
				fmt.Println("[WASM] WebSocket disconnected, reconnecting...")
			}
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
			// Check if the third argument is an object
			if args[2].Type() == js.TypeObject {
				// Convert JavaScript object to JSON string
				jsonData := js.Global().Get("JSON").Call("stringify", args[2])
				data = jsonData.String()
			} else {
				data = args[2].String()
			}
		}
		if isVerbose {
			fmt.Printf("[WASM] send_event: id=%s event=%s data=%s\n", id, event, data)
		}
		sendEvent(id, event, data)
		return nil
	}))
	
	// Initialize drag & drop handling
	initDragAndDrop()
	
	// Log inicial
	if isVerbose {
		fmt.Println("[WASM] LiveView WASM initialized")
	}
	
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

func initDragAndDrop() {
	// Create drag state object
	dragState := map[string]interface{}{
		"isDragging": false,
		"draggedBox": "",
		"startX":     0,
		"startY":     0,
		"initX":      0,
		"initY":      0,
		"lastUpdate": 0,
	}
	
	// Set global drag state
	js.Global().Set("dragState", dragState)
	
	// Mousedown handler
	mouseDownHandler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		e := args[0]
		target := e.Get("target")
		
		// Walk up the DOM tree to find a draggable box
		for !target.IsNull() && !target.Equal(document.Get("body")) {
			classList := target.Get("classList")
			if !classList.IsUndefined() && classList.Call("contains", "draggable-box").Bool() {
				e.Call("preventDefault")
				e.Call("stopPropagation")
				
				// Get box info
				boxId := target.Call("getAttribute", "data-box-id").String()
				boxX, _ := strconv.Atoi(target.Call("getAttribute", "data-box-x").String())
				boxY, _ := strconv.Atoi(target.Call("getAttribute", "data-box-y").String())
				
				// Update drag state
				state := js.Global().Get("dragState")
				state.Set("isDragging", true)
				state.Set("draggedBox", boxId)
				state.Set("startX", e.Get("clientX").Int())
				state.Set("startY", e.Get("clientY").Int())
				state.Set("initX", boxX)
				state.Set("initY", boxY)
				
				// Send start drag event
				dragData := map[string]interface{}{
					"id": boxId,
					"x":  e.Get("clientX").Int(),
					"y":  e.Get("clientY").Int(),
				}
				dragDataJSON, _ := json.Marshal(dragData)
				sendEvent("flow-tool", "BoxStartDrag", string(dragDataJSON))
				
				if isVerbose {
					fmt.Printf("[WASM] Started dragging: %s\n", boxId)
				}
				
				return false
			}
			target = target.Get("parentElement")
		}
		return nil
	})
	
	// Mousemove handler
	mouseMoveHandler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		e := args[0]
		state := js.Global().Get("dragState")
		
		if state.Get("isDragging").Bool() {
			e.Call("preventDefault")
			
			draggedBox := state.Get("draggedBox").String()
			startX := state.Get("startX").Int()
			startY := state.Get("startY").Int()
			initX := state.Get("initX").Int()
			initY := state.Get("initY").Int()
			
			deltaX := e.Get("clientX").Int() - startX
			deltaY := e.Get("clientY").Int() - startY
			newX := initX + deltaX
			newY := initY + deltaY
			
			// Update visual position immediately
			boxEl := document.Call("getElementById", "box-"+draggedBox)
			if !boxEl.IsNull() {
				style := boxEl.Get("style")
				style.Set("left", fmt.Sprintf("%dpx", newX))
				style.Set("top", fmt.Sprintf("%dpx", newY))
				boxEl.Call("setAttribute", "data-box-x", newX)
				boxEl.Call("setAttribute", "data-box-y", newY)
			}
			
			// Throttle server updates
			now := js.Global().Get("Date").Call("now").Float()
			lastUpdate := state.Get("lastUpdate").Float()
			if now-lastUpdate > 50 {
				moveData := map[string]interface{}{
					"id": draggedBox,
					"x":  newX,
					"y":  newY,
				}
				moveDataJSON, _ := json.Marshal(moveData)
				sendEvent("flow-tool", "BoxDrag", string(moveDataJSON))
				state.Set("lastUpdate", now)
				
				if isVerbose {
					fmt.Printf("[WASM] Dragging %s to (%d, %d)\n", draggedBox, newX, newY)
				}
			}
		}
		return nil
	})
	
	// Mouseup handler
	mouseUpHandler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		e := args[0]
		state := js.Global().Get("dragState")
		
		if state.Get("isDragging").Bool() {
			e.Call("preventDefault")
			
			draggedBox := state.Get("draggedBox").String()
			
			// Get final position
			boxEl := document.Call("getElementById", "box-"+draggedBox)
			if !boxEl.IsNull() {
				style := boxEl.Get("style")
				finalX, _ := strconv.Atoi(strings.TrimSuffix(style.Get("left").String(), "px"))
				finalY, _ := strconv.Atoi(strings.TrimSuffix(style.Get("top").String(), "px"))
				
				// Send final position
				finalData := map[string]interface{}{
					"id": draggedBox,
					"x":  finalX,
					"y":  finalY,
				}
				finalDataJSON, _ := json.Marshal(finalData)
				sendEvent("flow-tool", "BoxDrag", string(finalDataJSON))
			}
			
			// Send end drag event
			sendEvent("flow-tool", "BoxEndDrag", draggedBox)
			
			if isVerbose {
				fmt.Printf("[WASM] Ended dragging: %s\n", draggedBox)
			}
			
			// Reset drag state
			state.Set("isDragging", false)
			state.Set("draggedBox", "")
		}
		return nil
	})
	
	// Register event listeners
	document.Call("addEventListener", "mousedown", mouseDownHandler)
	document.Call("addEventListener", "mousemove", mouseMoveHandler)
	document.Call("addEventListener", "mouseup", mouseUpHandler)
	
	if isVerbose {
		fmt.Println("[WASM] Drag & Drop initialized")
	}
}
