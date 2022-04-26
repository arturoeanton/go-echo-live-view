package main

import (
	"fmt"
	"html"

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

		document := components.NewLayout("home", `
		{{ mount "text1"}}
		<hr/>
		<div id="div_text_result"></div>
		<div>
			{{mount "button2"}}
			{{mount "button1"}}
		</div>
		<div>
			<span id="span_result"></span>
		</div>
		`)

		liveview.New("span_result", &liveview.None{})
		liveview.New("div_text_result", &liveview.None{})
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
				spanResult := document.GetDriverById("span_result")
				text1 := document.GetDriverById("text1")
				flag := !(text1.GetPropertie("readOnly") == "true")
				text1.SetPropertie("readOnly", flag)
				spanResult.SetText("ReadOnly of text1 is " + text1.GetPropertie("readOnly"))
			})

		liveview.New("text1", &components.InputText{}).
			SetKeyUp(func(text1 *components.InputText, data interface{}) {
				divTextResult := document.GetDriverById("div_text_result")
				divTextResult.FillValue(html.EscapeString(data.(string)))
			})

		return document

	})

	e.Logger.Fatal(e.Start(":1323"))
}
