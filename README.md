# go-echo-live-view
Little POC for test the idea  of Phoenix LiveView in Go and Echo (https://github.com/labstack/echo) 


The idea was stolen from  https://github.com/brendonmatos/golive 



## Driver Methods

| Method | Description |
| --- | --- |
| `Remove` | return document.getElementById("$id").remove() |
| `GetHTML` | return document.getElementById("$id").innerHTML |
| `GetText` | return document.getElementById("$id").innerText |
| `GetPropertie` | return document.getElementById("$id")[$propertie] |
| `GetValue` | return document.getElementById("$id").value |
| `GetStyle` | return document.getElementById("$id").style["$propertie"] |
| `GetElementById` | return document.getElementById("$id").value |
| `EvalScript` | execute  eval($code);|
| `FillValue` | document.getElementById("$id").innerHTML = $value |
| `SetHTML` | document.getElementById("$id").innerHTML = $value |
| `SetText` | document.getElementById("$id").innerText = $value|
| `SetPropertie` | document.getElementById("$id")[$propertie] = $value |
| `SetValue` | document.getElementById("$id").value = $value|
| `SetStyle` | document.getElementById("$id").style.cssText = $style |



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


## Interface Component

```golang
type Component interface {
	GetTemplate() string
	Start()
}
```

## Example go-notebook

https://github.com/arturoeanton/go-notebook


![alt text](https://raw.githubusercontent.com/arturoeanton/go-notebook/main/gonote1.gif)

![alt text](https://raw.githubusercontent.com/arturoeanton/go-notebook/main/gonote2.gif)




## More Examples 

### example_todo
![alt text](https://raw.githubusercontent.com/arturoeanton/go-echo-live-view/main/example/example_todo/example_todo.gif)

### example1 
![alt text](https://raw.githubusercontent.com/arturoeanton/go-echo-live-view/main/example/example1/example1.gif)


### Example Style
```golang
package main

import (
	"github.com/arturoeanton/go-echo-live-view/components"
	"github.com/arturoeanton/go-echo-live-view/liveview"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Button struct {
	Driver *liveview.ComponentDriver
}

func (t *Button) Start() {
	t.Driver.Commit()
}

func (t *Button) GetTemplate() string {
	return `<button id="button1" onclick="send_event(this.id, 'Click')" >Change style</button>`
}

func (t *Button) Click(data interface{}) {
	background := t.Driver.GetStyle("button1", "background")
	if background != "red" {
		t.Driver.SetStyle("button1", "background: red")
	} else {
		t.Driver.SetStyle("button1", "background: blue")
	}
}

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
		button1 := liveview.NewDriver("button1", &Button{})
		return components.NewLayout("home", `<div> {{mount "button1"}} </div>`).Mount(button1)
	})
	e.Logger.Fatal(e.Start(":1323"))
}
```
![alt text](https://raw.githubusercontent.com/arturoeanton/go-echo-live-view/main/example/example_style/example_style.gif)






