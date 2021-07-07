package components

import "github.com/arturoeanton/go-echo-live-view/liveview"

type InputText struct {
	driver *liveview.ComponentDriver
	Id     string
}

func NewInputText(id string) *liveview.ComponentDriver {
	c := &InputText{Id: id}
	c.driver = liveview.NewDriver(c)
	return c.driver
}

func (t *InputText) Start() {
	t.driver.Commit()
}

func (t *InputText) GetTemplate() string {
	return `<input type="text" 
	onkeypress="send_event(this.id,'KeyPress',this.value)"
	onchange="send_event(this.id,'Change',this.value)"
	onkeyup="send_event(this.id,'KeyUp',this.value)"
	id="{{.Id}}"  />`
}

func (t *InputText) GetID() string {
	return t.Id
}

func (t *InputText) KeyPress(data interface{}) {

}

func (t *InputText) KeyUp(data interface{}) {}

func (t *InputText) Change(data interface{}) {}
