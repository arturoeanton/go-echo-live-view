package main

import (
	"fmt"
	"html"
	"time"

	"github.com/arturoeanton/go-echo-live-view/components"
	"github.com/arturoeanton/go-echo-live-view/liveview"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	home := liveview.PageControl{
		Title:  "Example2",
		Lang:   "en",
		Path:   "/",
		Router: e,
	}

	home.Register(func() liveview.LiveDriver {
		document := liveview.NewLayout("home", `
		{{ mount "text1"}}
		<hr/>
		LocalStatus:<div id="div_text_result"></div>
		<div>
			{{mount "button2"}}
			{{mount "button1"}}
		</div>
		<div>
			Resuls:<span id="span_result"></span><br/>
			Attr Text:<span id="span_result_attr"></span>
		</div>
		{{mount "button3"}}

		<div>
			GlobalStatus:<br/><span id="span_global_status"></span>
		</div>
		<br>
		`)

		liveview.New("span_result", &liveview.None{})
		liveview.New("div_text_result", &liveview.None{})
		liveview.New("span_result_attr", &liveview.None{})
		liveview.New("span_global_status", &liveview.None{})

		liveview.New("button3", &components.Button{Caption: "Button 3!!"}).
			SetClick(func(this *components.Button, data interface{}) {
				this.I++
				this.Caption = "Button " + fmt.Sprint(this.I)
				this.Commit()
			})

		liveview.New("button1", &components.Button{Caption: "Sum 1"}).
			SetEvent("Click", func(button1 *components.Button, data interface{}) {
				spanResult := document.GetDriverById("span_result")
				button1.I++
				text1 := document.GetDriverById("text1")
				spanResult.FillValue(fmt.Sprint(button1.I) + " -> " + html.EscapeString(text1.GetValue()))
				document.EvalScript("console.log(1)")
			})

		liveview.New("button2", &components.Button{Caption: "Change ReadOnly"}).
			SetClick(func(button2 *components.Button, data interface{}) {
				spanResultAttr := document.GetDriverById("span_result_attr")
				text1 := document.GetDriverById("text1")
				flag := !(text1.GetPropertie("readOnly") == "true")
				text1.SetPropertie("readOnly", flag)
				spanResultAttr.SetText("ReadOnly of text1 is " + text1.GetPropertie("readOnly"))
			})

		liveview.New("text1", &components.InputText{}).
			SetKeyUp(func(text1 *components.InputText, data interface{}) {
				divTextResult := document.GetDriverById("div_text_result")
				divTextResult.SetText(html.EscapeString(text1.GetValue()))
				//outChannel <- text1.GetValue()
				document.Component.ChanBus <- text1.GetValue()
				divTextResult.FillValue(html.EscapeString(data.(string)))
			})

		go func() {
			for {
				select {
				case data := <-document.Component.ChanIn:
					spanGlobalStatus := document.GetDriverById("span_global_status")
					spanGlobalStatus.FillValue(fmt.Sprint(data))
				case <-time.After(time.Second * 10):
					spanGlobalStatus := document.GetDriverById("span_global_status")
					spanGlobalStatus.FillValue(" shhhhhh")
				}
			}
		}()

		return document

	},
	)

	e.Logger.Fatal(e.Start(":1323"))
}

func SafeSend[T any](ch chan T, value T) (closed bool) {
	defer func() {
		if recover() != nil {
			closed = true
		}
	}()

	ch <- value  // panic if ch is closed
	return false // <=> closed = false; return
}
