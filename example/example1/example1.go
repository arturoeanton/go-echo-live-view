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
		Title:  "Home",
		Lang:   "en",
		Path:   "/",
		Router: e,
	}

	home.Register(func() *liveview.ComponentDriver {
		clock1 := liveview.NewDriver("clock1", &components.Clock{})
		return components.NewLayout("home", `
		<div id="d2">{{mount "clock1"}}</div>
		`).Mount(clock1)
	})

	e.Logger.Fatal(e.Start(":1323"))
}
