package main

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/arturoeanton/go-echo-live-view/components"
	"github.com/arturoeanton/go-echo-live-view/liveview"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var (
	userMutex = &sync.Mutex{}
	bUser     *liveview.BiMap[string, string]
)

func main() {
	e := echo.New()
	e.Use(middleware.Logger(), middleware.Recover())
	bUser = liveview.NewBiMap[string, string]()
	home := liveview.PageControl{
		Title:  "Example2",
		Lang:   "en",
		Path:   "/",
		Router: e,
	}
	bUser.Set("*", "Todos")
	home.Register(func() liveview.LiveDriver {
		document := liveview.NewLayout(`
			<div> Nickname: {{ mount "text_nickname" }} <span id="span_text_nickname"></span>
			<hr/>
			<div id="div_general_chat"></div>
			<hr/>
			<div> Message: {{ mount "text_msg" }} to  {{ mount "select_to"}}  {{mount "button_send"}}</div>
			<hr/>
			<div id="div_status"></div>`)
		liveview.NewWithTemplate("select_to",
			`<select   onchange="send_event(this.id,'Change',this.value)" id="{{.IdComponent}}"   >
				{{range $index, $element := .GetDriver.Data}}
				<option value="{{$index}}" >{{$element}}</option>
				{{end}}
			</select>`)
		liveview.New("text_msg", &components.InputText{})
		liveview.New("text_nickname", &components.InputText{}).
			SetEvent("Change", func(this *components.InputText, data interface{}) {
				userMutex.Lock()
				defer userMutex.Unlock()
				if _, ok := bUser.GetByValue(this.GetValue()); ok {
					this.SetValue("")
					return
				}
				bUser.Set(document.Component.UUID, this.GetValue())
				spanTextNickname := document.GetDriverById("span_text_nickname")
				spanTextNickname.FillValue(fmt.Sprint(data))
				liveview.SendToAllLayouts("NEW_USER")
			})
		liveview.New("button_send", &components.Button{Caption: "Send"}).
			SetClick(func(this *components.Button, data interface{}) {
				userMutex.Lock()
				defer userMutex.Unlock()
				if nickname, ok := bUser.Get(document.Component.UUID); ok {
					textMsg := document.GetDriverById("text_msg").GetValue()
					idTo := document.GetDriverById("select_to").GetValue()
					if idTo == "*" {
						liveview.SendToAllLayouts("MSG|" + fmt.Sprint(nickname, "[Public]:", textMsg))
						return
					}
					if userTo, ok := bUser.Get(idTo); ok {
						liveview.SendToLayouts("MSG|"+fmt.Sprint(nickname, " to ", userTo, "[Private]:", textMsg), idTo, document.Component.UUID)
					}
				}
			})
		document.Component.SetHandlerEventIn(func(data interface{}) {
			msg := data.(string)
			if strings.HasPrefix(msg, "MSG|") {
				msg = strings.TrimPrefix(msg, "MSG|")
				divGeneralChat := document.GetDriverById("div_general_chat")
				divGeneralChat.FillValue(fmt.Sprint(divGeneralChat.GetHTML(), msg, "<br/>"))
			}
			if strings.HasPrefix(msg, "NEW_USER") {
				selectTo := document.GetDriverById("select_to")
				idTemp := selectTo.GetValue()
				selectTo.SetData(bUser.GetAll())
				selectTo.Commit()
				if _, ok := bUser.Get(idTemp); ok {
					selectTo.SetValue(idTemp)
				}
			}
		})
		document.Component.SetHandlerEventTime(time.Second*5, func() {
			spanGlobalStatus := document.GetDriverById("div_status")
			spanGlobalStatus.FillValue("online")
		})

		document.Component.SetHandlerEventDestroy(func(id string) {
			bUser.Delete(id)
			liveview.SendToAllLayouts("NEW_USER")
		})
		return document
	})
	e.Logger.Fatal(e.Start(":1323"))
}
