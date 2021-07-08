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

		button1 := components.NewButton("button1", "button1")
		text1 := components.NewInputText("text1")

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

		content := components.NewLayout("home", `
		<div id="div_text"></div>
		<div id="div_text_result"></div>
		<div>
			<span id="span_button"></span>
		</div>
		<div>
			<span id="span_result"></span>
		</div>
		`)
		content.Mount("div_text", text1)
		content.Mount("span_button", button1)

		return content
	})

	e.Logger.Fatal(e.Start(":1323"))
}
