// Package main provides the WebAssembly module for Go Echo LiveView framework.
// This module handles client-side WebSocket communication, DOM manipulation,
// and drag-and-drop functionality for real-time web applications.
//
// Build Instructions:
//   cd cmd/wasm/
//   GOOS=js GOARCH=wasm go build -o ../../assets/json.wasm
//   cd -
//
// The WASM module is automatically loaded by the framework and provides:
//   - WebSocket connection management with auto-reconnection
//   - DOM manipulation and event handling
//   - Generic drag-and-drop support for any draggable element
//   - Real-time bidirectional communication between client and server
package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"syscall/js"
)

// Global JavaScript objects and configuration
var (
	// DOM and browser API references
	document  js.Value = js.Global().Get("document")  // Document object for DOM manipulation
	window    js.Value = js.Global().Get("window")    // Window object for browser APIs
	console   js.Value = js.Global().Get("console")   // Console for debugging output
	webSocket js.Value = js.Global().Get("WebSocket") // WebSocket constructor
	loc       js.Value = window.Get("location")       // Current page location
	
	// WebSocket connection variables
	uri       string   = "ws:"                         // WebSocket URI scheme (ws: or wss:)
	ws        js.Value                                 // Active WebSocket connection
	protocol  string = loc.Get("protocol").String()   // Current protocol (http: or https:)
	
	// Configuration flags
	isVerbose bool                                     // Enable verbose logging when ?verbose=true in URL
)

// MsgEvent represents an outgoing event message from client to server.
// These messages are sent when user interactions occur in the browser.
type MsgEvent struct {
	Type  string `json:"type"`  // Message type (always "data" for user events)
	ID    string `json:"id"`    // Component ID that triggered the event
	Event string `json:"event"`  // Event name (e.g., "Click", "DragStart")
	Data  string `json:"data"`   // JSON-encoded event data
}

// DataEventIn represents an incoming command from the server.
// The server sends these messages to update the DOM or request data.
type DataEventIn struct {
	ID        string      `json:"id"`        // DOM element ID to target
	IdRet     string      `json:"id_ret"`    // Return ID for response messages
	Type      string      `json:"type"`      // Operation type (fill, text, style, script, etc.)
	Value     interface{} `json:"value"`     // Value to apply to the element
	Propertie string      `json:"propertie"` // Property name for "propertie" type operations
	SubType   string      `json:"sub_type"`  // Sub-type for "get" operations
}

// DataEventOut represents a response message from client to server.
// These are sent in response to "get" type requests from the server.
type DataEventOut struct {
	Type  string      `json:"type"`   // Response type (always "get" for responses)
	IdRet string      `json:"id_ret"` // Return ID matching the request
	Data  interface{} `json:"data"`   // Requested data from the DOM
}

// connect establishes a WebSocket connection to the LiveView server.
// It sets up all necessary event handlers for WebSocket lifecycle and message processing.
// The connection URL is automatically determined from the current page location.
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

// main is the entry point for the WebAssembly module.
// It initializes the WebSocket connection, sets up global JavaScript functions,
// configures auto-reconnection, and initializes the drag-and-drop system.
// The module runs indefinitely, maintaining the connection and handling events.
func main() {
	// Configure verbose logging based on URL parameters
	// Enable with ?verbose=true or ?debug=true in the URL
	verbose := js.Global().Get("location").Get("search").String()
	isVerbose = strings.Contains(verbose, "verbose=true") || strings.Contains(verbose, "debug=true")
	
	if isVerbose {
		fmt.Println("[WASM] Verbose mode enabled")
	}
	
	document.Call("getElementById", "content").Set("innerHTML", "Disconnected")
	connect()

	// Set up auto-reconnection mechanism
	// Checks WebSocket connection status every second and reconnects if needed
	// ReadyState: 0=CONNECTING, 1=OPEN, 2=CLOSING, 3=CLOSED
	js.Global().Call("setInterval", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if ws.Get("readyState").Int() != 1 { // Not OPEN
			if isVerbose {
				fmt.Println("[WASM] WebSocket disconnected, reconnecting...")
			}
			connect()
		}
		return nil
	}), 1000) // Check every 1 second

	// Expose WebSocket and utility functions to global JavaScript scope
	// This allows debugging and manual control from browser console
	js.Global().Set("ws", ws) // Access WebSocket via window.ws
	js.Global().Set("connect", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		connect()
		return nil
	})) // Manual reconnect via window.connect()

	// Register the send_event function for DOM elements to communicate with server
	// Usage: send_event('component-id', 'EventName', data)
	// The data parameter can be a string or JavaScript object (automatically serialized)
	js.Global().Set("send_event", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		id := args[0].String()    // Component ID
		event := args[1].String() // Event name
		data := ""               // Optional event data
		
		if len(args) == 3 {
			// Handle both string and object data types
			if args[2].Type() == js.TypeObject {
				// Serialize JavaScript objects to JSON for transmission
				jsonData := js.Global().Get("JSON").Call("stringify", args[2])
				data = jsonData.String()
			} else {
				// Use string data as-is
				data = args[2].String()
			}
		}
		
		if isVerbose {
			fmt.Printf("[WASM] send_event: id=%s event=%s data=%s\n", id, event, data)
		}
		sendEvent(id, event, data)
		return nil
	}))
	
	// Initialize the generic drag & drop system
	// This sets up event listeners for all elements with class="draggable"
	initDragAndDrop()
	
	// Log successful initialization
	if isVerbose {
		fmt.Println("[WASM] LiveView WASM initialized successfully")
	}
	
	// Keep the WebAssembly module running indefinitely
	// This is required for WASM modules to maintain event listeners
	<-make(chan struct{})
}

// GetValue converts a JavaScript value to its Go equivalent.
// This function handles type conversion between JavaScript and Go types,
// ensuring proper data representation when retrieving DOM properties.
//
// Supported conversions:
//   - Boolean -> bool
//   - Number -> int
//   - String -> string
//   - Null/Undefined -> nil
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

// sendEvent transmits a user event to the server via WebSocket.
// This is the primary mechanism for client-to-server communication in LiveView.
//
// Parameters:
//   - id: The component ID that triggered the event
//   - event: The event name (e.g., "Click", "Input", "DragStart")
//   - data: Optional event data as a JSON string
//
// The message is sent as JSON over the WebSocket connection.
func sendEvent(id string, event string, data string) {
	msgEvent := MsgEvent{
		Type:  "data",  // Always "data" for user-triggered events
		ID:    id,
		Event: event,
		Data:  data,
	}
	jsonMsg, _ := json.Marshal(&msgEvent)
	ws.Call("send", string(jsonMsg))
}

// initDragAndDrop initializes the generic drag-and-drop system.
// This function sets up event listeners for mousedown, mousemove, and mouseup events
// to enable dragging functionality for any element with the "draggable" class.
//
// The system supports:
//   - Generic dragging for any element with class="draggable"
//   - Automatic position updates during drag
//   - Throttled server updates to prevent overwhelming the WebSocket
//   - Z-index management to ensure draggable elements stay on top
//   - Backward compatibility with legacy "BoxDrag" events
//
// Elements become draggable by adding:
//   - class="draggable" (required)
//   - data-element-id="unique-id" (required)
//   - data-component-id="owner-component" (required)
//   - data-drag-disabled="true" (optional, to temporarily disable dragging)
func initDragAndDrop() {
	// Initialize drag state tracking object
	// This maintains the current drag state across mouse events
	dragState := map[string]interface{}{
		"isDragging":     false, // Whether a drag operation is in progress
		"draggedElement": "",    // ID of the element being dragged
		"startX":         0,     // Initial mouse X position
		"startY":         0,     // Initial mouse Y position
		"initX":          0,     // Initial element X position
		"initY":          0,     // Initial element Y position
		"lastUpdate":     0,     // Timestamp of last server update (for throttling)
		"componentId":    "",    // ID of the component that owns the draggable element
	}
	
	// Expose drag state globally for debugging and external access
	js.Global().Set("dragState", dragState)
	
	// mouseDownHandler handles the initial mouse down event to start dragging.
	// It walks up the DOM tree to find a draggable element, checks if dragging
	// is enabled, captures initial positions, and sends DragStart events to the server.
	mouseDownHandler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		e := args[0]
		target := e.Get("target")
		
		if isVerbose {
			fmt.Printf("[WASM] Mouse down on element: %s, classes: %s\n", target.Get("tagName").String(), target.Get("className").String())
		}
		
		// Walk up the DOM tree to find a draggable element
		// This allows clicks on child elements to still trigger dragging on the parent
		for !target.IsNull() && !target.Equal(document.Get("body")) {
			classList := target.Get("classList")
			
			// Skip elements with pointer-events: none
			// These are typically decorative elements that shouldn't be interactive
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
				// Check if dragging is temporarily disabled
				// This is useful for elements that should be draggable sometimes but not always
				// (e.g., disabled during connection mode in flow diagrams)
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
			
			// Throttle server updates - reduced to 16ms for smoother updates (60 FPS)
			now := js.Global().Get("Date").Call("now").Float()
			lastUpdate := state.Get("lastUpdate").Float()
			if now-lastUpdate > 16 && componentId != "" {
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
