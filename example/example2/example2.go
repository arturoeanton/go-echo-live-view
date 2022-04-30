package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/arturoeanton/go-echo-live-view/components"
	"github.com/arturoeanton/go-echo-live-view/liveview"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var (
	// LiveView is the live view instance
	userMutex                   = &sync.Mutex{}
	users     map[string]string = make(map[string]string)
	usersById map[string]string = make(map[string]string)
)

func main() {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	home := liveview.PageControl{
		Title:  "Example2",
		Lang:   "en",
		Path:   "/",
		Router: e,
	}

	home.Register(func() liveview.LiveDriver {
		document := liveview.NewLayout(`
		
		<div> Nickname: {{ mount "text_nickname" }} <span id="span_text_nickname"></span>
		<hr/>
		<div id="div_general_chat">
		</div>
		<hr/>
		<div> Message: {{ mount "text_msg" }} to {{ mount "text_to" }} {{mount "button_send"}}</div>
		<hr/>
		<br>
		<div id="div_status"></div>
		`)

		liveview.New("span_text_nickname", &liveview.None{})
		liveview.New("div_general_chat", &liveview.None{})
		liveview.New("span_result", &liveview.None{})
		liveview.New("div_status", &liveview.None{})

		liveview.New("text_msg", &components.InputText{})
		liveview.New("text_to", &components.InputText{})

		liveview.New("text_nickname", &components.InputText{}).
			SetEvent("Change", func(this *components.InputText, data interface{}) {
				userMutex.Lock()
				defer userMutex.Unlock()
				users[fmt.Sprint(data)] = document.Component.UUID
				usersById[document.Component.UUID] = fmt.Sprint(data)
				spanTextNickname := document.GetDriverById("span_text_nickname")
				spanTextNickname.FillValue(fmt.Sprint(data))
				textTo := document.GetDriverById("text_to")
				textTo.SetValue("*")
			})

		liveview.New("button_send", &components.Button{Caption: "Send"}).
			SetClick(func(this *components.Button, data interface{}) {
				nickname := usersById[document.Component.UUID]
				textMsg := document.GetDriverById("text_msg")
				textTo := document.GetDriverById("text_to")
				if textTo.GetValue() == "*" || textTo.GetValue() == "" {
					liveview.SendToAllLayouts(fmt.Sprint(nickname, "[Public]:", textMsg.GetValue()))
					return
				}
				liveview.SendToLayouts(fmt.Sprint(nickname, " to ", textTo.GetValue(), "[Private]:", textMsg.GetValue()), users[textTo.GetValue()], document.Component.UUID)
			})

		go func() {
			for {
				select {
				case data := <-document.Component.ChanIn:
					divGeneralChat := document.GetDriverById("div_general_chat")
					history := divGeneralChat.GetHTML() + "<br/>"
					divGeneralChat.FillValue(fmt.Sprint(history, data))
				case <-time.After(time.Second * 10):
					spanGlobalStatus := document.GetDriverById("div_status")
					spanGlobalStatus.FillValue("online")
				}
			}
		}()

		return document

	},
	)

	e.Logger.Fatal(e.Start(":1323"))
}
