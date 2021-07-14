package main

import (
	"fmt"

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
		Title:  "Home",
		Lang:   "en",
		Path:   "/",
		Router: e,
	}

	home.Register(func() *liveview.ComponentDriver {

		button1 := liveview.NewDriver("button1", &components.Button{Caption: "Sum 1"})
		text1 := liveview.NewDriver("text1", &components.InputText{})

		text1.Events["KeyUp"] = func(data interface{}) {
			text1.FillValue("div_text_result", data.(string))
		}

		button1.Events["Click"] = func(data interface{}) {
			button := button1.Component.(*components.Button)
			button.I++
			text := button.Driver.GetElementById("text1")
			button.Driver.FillValue("span_result", fmt.Sprint(button.I)+" -> "+text)
			button.Driver.EvalScript("console.log(1)")
		}
		button2 := liveview.NewDriver("button2", &components.Button{Caption: "Change ReadOnly"})
		button2.Events["Click"] = func(data interface{}) {
			button := button2.Component.(*components.Button)
			text := button.Driver.GetPropertie("text1", "readOnly")
			if text == "true" {
				button.Driver.SetPropertie("text1", "readOnly", false)
			} else {
				button.Driver.SetPropertie("text1", "readOnly", true)
			}
			text = button.Driver.GetPropertie("text1", "readOnly")
			button.Driver.SetText("span_result", "ReadOnly of text1 is "+text)
		}

		return components.NewLayout("home", `
		{{ mount "text1"}}
		<div id="div_text_result"></div>
		<div>
			{{mount "button2"}}
			{{mount "button1"}}
		</div>
		<div>
			<span id="span_result"></span>
		</div>
		`).Mount(text1).Mount(button1).Mount(button2)

	})

	e.Logger.Fatal(e.Start(":1323"))
}
