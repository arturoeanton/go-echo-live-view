package main

import (
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
		Title:  "Example1",
		Path:   "/",
		Router: e,
	}

	home.Register(func() liveview.LiveDriver {
		liveview.New("clock1", &components.Clock{})
		return liveview.NewLayout(`
		<div id="d2">{{mount "clock1"}}</div>
		`)
	})

	e.Logger.Fatal(e.Start(":1323"))
}
