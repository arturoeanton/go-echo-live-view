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

		button1 := components.NewButton("b1", "button1")
		button2 := components.NewButton("b2", "button2")
		text1 := components.NewInputText("t1")

		text1.Events["KeyPress"] = func(data interface{}) {
			fmt.Println("KeyPress:" + data.(string))
		}

		text1.Events["KeyUp"] = func(data interface{}) {
			fmt.Println("KeyUp:" + data.(string))
		}

		text1.Events["Change"] = func(data interface{}) {
			fmt.Println("Change:" + data.(string))
		} //*/

		button1.Events["Click"] = func(data interface{}) {
			button := button1.Component.(*components.Button)
			button.I++
			text := button.Driver.GetElementById("t1")
			button.Driver.FillValue("d6", fmt.Sprint(button.I)+" -> "+text)
			button.Driver.EvalScript("console.log(1)")
		}

		button2.Events["Click"] = func(data interface{}) {
			button := button2.Component.(*components.Button)
			button.I++
			text := button.Driver.GetElementById("t1")
			button.Driver.FillValue("d7", fmt.Sprint(button.I)+" <- "+text)
			button.Driver.EvalScript("console.log(2)")
		} //*/

		content := components.NewLayout("home", `
		<div id="d2"></div>
		<div id="d3"></div>
		<div>
			<span id="d4"></span>
			<span id="d5"></span>
		</div>
		<div>
			<span id="d6"></span>
			<span id="d7"></span>
		</div>
		`)
		content.Mount("d2", components.NewClock("c1"))
		content.Mount("d3", text1)
		content.Mount("d4", button1)
		content.Mount("d5", button2)

		return content
	})

	e.Logger.Fatal(e.Start(":1323"))
}
