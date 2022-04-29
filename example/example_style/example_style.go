package main

import (
	"github.com/arturoeanton/go-echo-live-view/liveview"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Button struct {
	*liveview.ComponentDriver[*Button]
}

func (t *Button) GetDriver() liveview.LiveDriver {
	return t
}

func (t *Button) Start() {
	t.Commit()
}

func (t *Button) GetTemplate() string {
	return `<button id="button1" onclick="send_event(this.id, 'Click')" >Change style</button>`
}

func (t *Button) Click(data interface{}) {

	background := t.GetStyle("background")
	if background != "red" {
		t.SetStyle("background: red")
	} else {
		t.SetStyle("background: blue")
	}
}

func main() {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	home := liveview.PageControl{
		Title:    "Home",
		HeadCode: "example/example_todo/head.html",
		Lang:     "en",
		Path:     "/",
		Router:   e,
		//	Debug:    true,
	}
	home.Register(func() liveview.LiveDriver {
		document := liveview.NewLayout("home", `<div> {{mount "button1"}} </div>`)
		liveview.New("button1", &Button{})
		return document
	})
	e.Logger.Fatal(e.Start(":1323"))
}
