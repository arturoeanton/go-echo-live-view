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
		Title:    "Example3",
		HeadCode: "example/example3/head.html",
		Lang:     "en",
		Path:     "/",
		Router:   e,
		//	Debug:    true,
	}

	home.Register(func() liveview.LiveDriver {
		document := liveview.NewLayout("home", "example/example3/layout.html")
		liveview.New("span_result", &liveview.None{})
		liveview.New("div_text_result", &liveview.None{})
		liveview.New("text1", &components.InputText{}).SetKeyUp(func(text1 *components.InputText, data interface{}) {
			divTextResult := document.GetDriverById("div_text_result")
			divTextResult.FillValue(text1.GetValue())
		})

		liveview.New("button1", &components.Button{Caption: "Sum 1"}).SetClick(func(button1 *components.Button, data interface{}) {
			button1.I++
			spanResult := document.GetDriverById("span_result")
			text1 := document.GetDriverById("text1")
			spanResult.FillValue(fmt.Sprint(button1.I) + " -> " + text1.GetValue())
		})
		return document
	})

	e.Logger.Fatal(e.Start(":1323"))
}
