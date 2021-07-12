# go-echo-live-view
Little POC for test the idea  of Phoenix LiveView in Go and Echo (https://github.com/labstack/echo) 


The idea was stolen from  https://github.com/brendonmatos/golive 


## Example 

```golang
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

		return components.NewLayout("home", `
		{{ mount "text1"}}
		<div id="div_text_result"></div>
		<div>
			{{mount "button1"}}
		</div>
		<div>
			<span id="span_result"></span>
		</div>
		`).Mount(text1).Mount(button1)

	})

	e.Logger.Fatal(e.Start(":1323"))
}
```

![alt text](https://raw.githubusercontent.com/arturoeanton/go-echo-live-view/main/example/example2/example2.gif)



## Examples 

![alt text](https://raw.githubusercontent.com/arturoeanton/go-echo-live-view/main/example/example_todo/example_todo.gif)
![alt text](https://raw.githubusercontent.com/arturoeanton/go-echo-live-view/main/example/example1/example1.gif)








