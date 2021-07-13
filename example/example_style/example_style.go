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
