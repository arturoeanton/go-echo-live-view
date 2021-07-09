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
		Title:    "Home",
		HeadCode: "head.html",
		Lang:     "en",
		Path:     "/",
		Router:   e,
		//	Debug:    true,
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
		}

		return components.NewLayout("home", "layout.html").Mount(text1).Mount(button1)

	})

	e.Logger.Fatal(e.Start(":1323"))
}
