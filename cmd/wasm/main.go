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
		
		// Debug logging for all messages in verbose mode
		if isVerbose {
			if dataEventIn.Type == "fill" {
				fmt.Printf("[WASM] FILL message - ID: %s, Value length: %d\n", dataEventIn.ID, len(fmt.Sprint(dataEventIn.Value)))
			} else {
				fmt.Printf("[WASM] Received message type: %s, ID: %s, Value: %v\n", dataEventIn.Type, dataEventIn.ID, dataEventIn.Value)
			}
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
	// Create drag state object for generic drag and drop
	dragState := map[string]interface{}{
		"isDragging":     false,
		"draggedElement": "",
		"startX":         0,
		"startY":         0,
		"initX":          0,
		"initY":          0,
		"lastUpdate":     0,
		"componentId":    "", // The component that owns the draggable element
	}
	
	// Set global drag state
	js.Global().Set("dragState", dragState)
	
	// Mousedown handler - generic for any draggable element
	mouseDownHandler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		e := args[0]
		target := e.Get("target")
		
		if isVerbose {
			fmt.Printf("[WASM] Mouse down on element: %s, classes: %s\n", target.Get("tagName").String(), target.Get("className").String())
		}
		
		// Walk up the DOM tree to find a draggable element
		for !target.IsNull() && !target.Equal(document.Get("body")) {
			classList := target.Get("classList")
			
			// Skip elements with pointer-events: none
			style := window.Call("getComputedStyle", target)
			pointerEvents := style.Get("pointerEvents").String()
			if pointerEvents == "none" {
				if isVerbose {
					fmt.Printf("[WASM] Skipping element with pointer-events:none: %s\n", target.Get("tagName").String())
				}
				target = target.Get("parentElement")
				continue
			}
			
			if isVerbose {
				fmt.Printf("[WASM] Checking element: %s, classes: %s, id: %s\n", target.Get("tagName").String(), target.Get("className").String(), target.Get("id").String())
			}
			if !classList.IsUndefined() && (classList.Call("contains", "draggable").Bool() || classList.Call("contains", "draggable-box").Bool()) {
				if isVerbose {
					fmt.Printf("[WASM] Found draggable element: %s\n", target.Get("id").String())
				}
				// Check if dragging is disabled on this element
				if target.Call("hasAttribute", "data-drag-disabled").Bool() {
					return nil
				}
				
				e.Call("preventDefault")
				e.Call("stopPropagation")
				
				// Get element info from data attributes (try both new and old format)
				elementId := target.Call("getAttribute", "data-element-id")
				if elementId.IsNull() {
					// Try old format
					boxId := target.Call("getAttribute", "data-box-id")
					if !boxId.IsNull() {
						elementId = js.ValueOf("box-" + boxId.String())
					} else {
						elementId = target.Get("id")
					}
				}
				
				// Get component ID from data attribute or parent
				componentId := target.Call("getAttribute", "data-component-id")
				if componentId.IsNull() {
					// For old format, default to flow-tool
					if !target.Call("getAttribute", "data-box-id").IsNull() {
						componentId = js.ValueOf("flow-tool")
					} else {
						// Try to find component ID from parent elements
						parent := target.Get("parentElement")
						for !parent.IsNull() && !parent.Equal(document.Get("body")) {
							compId := parent.Call("getAttribute", "data-component-id")
							if !compId.IsNull() {
								componentId = compId
								break
							}
							parent = parent.Get("parentElement")
						}
					}
				}
				
				// Get initial position (try data attributes first, then computed style)
				initX := 0
				initY := 0
				
				// Try old format data attributes first
				boxX := target.Call("getAttribute", "data-box-x")
				boxY := target.Call("getAttribute", "data-box-y")
				if !boxX.IsNull() && !boxY.IsNull() {
					initX, _ = strconv.Atoi(boxX.String())
					initY, _ = strconv.Atoi(boxY.String())
				} else {
					// Fall back to computed style
					style := window.Call("getComputedStyle", target)
					leftStr := style.Get("left").String()
					topStr := style.Get("top").String()
					
					if leftStr != "auto" && leftStr != "" {
						initX, _ = strconv.Atoi(strings.TrimSuffix(leftStr, "px"))
					}
					if topStr != "auto" && topStr != "" {
						initY, _ = strconv.Atoi(strings.TrimSuffix(topStr, "px"))
					}
				}
				
				// Update drag state
				state := js.Global().Get("dragState")
				state.Set("isDragging", true)
				state.Set("draggedElement", elementId.String())
				state.Set("componentId", componentId.String())
				state.Set("startX", e.Get("clientX").Int())
				state.Set("startY", e.Get("clientY").Int())
				state.Set("initX", initX)
				state.Set("initY", initY)
				
				// Send specific or generic drag start event based on element type
				if !componentId.IsNull() && componentId.String() != "" {
					dragData := map[string]interface{}{
						"element": elementId.String(),
						"x":       e.Get("clientX").Int(),
						"y":       e.Get("clientY").Int(),
					}
					
					// Check if this is a flow-tool box for backward compatibility
					if strings.HasPrefix(elementId.String(), "box-") && componentId.String() == "flow-tool" {
						// Send both old-style BoxStartDrag and new generic DragStart
						boxData := map[string]interface{}{
							"id": strings.TrimPrefix(elementId.String(), "box-"),
							"x":  e.Get("clientX").Int(),
							"y":  e.Get("clientY").Int(),
						}
						boxDataJSON, _ := json.Marshal(boxData)
						sendEvent(componentId.String(), "BoxStartDrag", string(boxDataJSON))
					}
					
					// Always send generic event too
					dragDataJSON, _ := json.Marshal(dragData)
					sendEvent(componentId.String(), "DragStart", string(dragDataJSON))
				}
				
				if isVerbose {
					fmt.Printf("[WASM] Started dragging: %s\n", elementId.String())
				}
				
				return false
			}
			target = target.Get("parentElement")
		}
		return nil
	})
	
	// Mousemove handler - generic
	mouseMoveHandler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		e := args[0]
		state := js.Global().Get("dragState")
		
		if state.Get("isDragging").Bool() {
			e.Call("preventDefault")
			
			draggedElement := state.Get("draggedElement").String()
			componentId := state.Get("componentId").String()
			startX := state.Get("startX").Int()
			startY := state.Get("startY").Int()
			initX := state.Get("initX").Int()
			initY := state.Get("initY").Int()
			
			deltaX := e.Get("clientX").Int() - startX
			deltaY := e.Get("clientY").Int() - startY
			newX := initX + deltaX
			newY := initY + deltaY
			
			// Update visual position immediately
			element := document.Call("getElementById", draggedElement)
			if !element.IsNull() {
				style := element.Get("style")
				style.Set("left", fmt.Sprintf("%dpx", newX))
				style.Set("top", fmt.Sprintf("%dpx", newY))
				
				// Update data attributes if they exist (for old format compatibility)
				if !element.Call("getAttribute", "data-box-x").IsNull() {
					element.Call("setAttribute", "data-box-x", newX)
					element.Call("setAttribute", "data-box-y", newY)
				}
			}
			
			// Throttle server updates
			now := js.Global().Get("Date").Call("now").Float()
			lastUpdate := state.Get("lastUpdate").Float()
			if now-lastUpdate > 50 && componentId != "" {
				// Check if this is a flow-tool box for backward compatibility
				if strings.HasPrefix(draggedElement, "box-") && componentId == "flow-tool" {
					// Send old-style BoxDrag event
					boxData := map[string]interface{}{
						"id": strings.TrimPrefix(draggedElement, "box-"),
						"x":  newX,
						"y":  newY,
					}
					boxDataJSON, _ := json.Marshal(boxData)
					sendEvent(componentId, "BoxDrag", string(boxDataJSON))
				}
				
				// Always send generic event too
				moveData := map[string]interface{}{
					"element": draggedElement,
					"x":       newX,
					"y":       newY,
				}
				moveDataJSON, _ := json.Marshal(moveData)
				sendEvent(componentId, "DragMove", string(moveDataJSON))
				state.Set("lastUpdate", now)
				
				if isVerbose {
					fmt.Printf("[WASM] Dragging %s to (%d, %d)\n", draggedElement, newX, newY)
				}
			}
		}
		return nil
	})
	
	// Mouseup handler - generic
	mouseUpHandler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		e := args[0]
		state := js.Global().Get("dragState")
		
		if state.Get("isDragging").Bool() {
			e.Call("preventDefault")
			
			draggedElement := state.Get("draggedElement").String()
			componentId := state.Get("componentId").String()
			
			// Get final position
			element := document.Call("getElementById", draggedElement)
			if !element.IsNull() && componentId != "" {
				style := element.Get("style")
				finalX, _ := strconv.Atoi(strings.TrimSuffix(style.Get("left").String(), "px"))
				finalY, _ := strconv.Atoi(strings.TrimSuffix(style.Get("top").String(), "px"))
				
				// Check if this is a flow-tool box for backward compatibility
				if strings.HasPrefix(draggedElement, "box-") && componentId == "flow-tool" {
					// Send old-style BoxEndDrag event
					boxData := map[string]interface{}{
						"id": strings.TrimPrefix(draggedElement, "box-"),
						"x":  finalX,
						"y":  finalY,
					}
					boxDataJSON, _ := json.Marshal(boxData)
					sendEvent(componentId, "BoxEndDrag", string(boxDataJSON))
				}
				
				// Always send generic event too
				finalData := map[string]interface{}{
					"element": draggedElement,
					"x":       finalX,
					"y":       finalY,
				}
				finalDataJSON, _ := json.Marshal(finalData)
				sendEvent(componentId, "DragEnd", string(finalDataJSON))
			}
			
			if isVerbose {
				fmt.Printf("[WASM] Ended dragging: %s\n", draggedElement)
			}
			
			// Reset drag state
			state.Set("isDragging", false)
			state.Set("draggedElement", "")
			state.Set("componentId", "")
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
