package main

import (
	"encoding/json"
	"strconv"

	"github.com/arturoeanton/go-echo-live-view/components"
	"github.com/arturoeanton/go-echo-live-view/liveview"
	"github.com/arturoeanton/gocommons/utils"
	"github.com/google/uuid"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Task struct {
	Name  *string `json:"name,omitempty"`
	State *int    `json:"state,omitempty"`
}

type Todo struct {
	Driver     *liveview.ComponentDriver
	ActualTime string
	code       string
	Tasks      map[string]Task
}

func (t *Todo) Start() {
	tasksString, _ := utils.FileToString("tasks.json")
	t.Tasks = make(map[string]Task)
	json.Unmarshal([]byte(tasksString), &(t.Tasks))
	t.Driver.Commit()
}

func (t *Todo) GetTemplate() string {
	if t.code == "" {
		t.code, _ = utils.FileToString("todo.html")
	}
	return t.code
}

func (t *Todo) Add(data interface{}) {
	name := t.Driver.GetElementById("new_name")
	stateStr := t.Driver.GetElementById("new_state")
	state, _ := strconv.Atoi(stateStr)
	id := uuid.NewString()
	task := Task{
		Name:  &name,
		State: &state,
	}
	t.Tasks[id] = task
	content, _ := json.Marshal(t.Tasks)
	utils.StringToFile("tasks.json", string(content))
	t.Driver.Commit()
}

func (t *Todo) Remove(data interface{}) {
	id := data.(string)
	delete(t.Tasks, id)
	content, _ := json.Marshal(t.Tasks)
	utils.StringToFile("tasks.json", string(content))
	t.Driver.Commit()
}

func (t *Todo) Change(data interface{}) {
	id := data.(string)
	name := t.Driver.GetElementById("name_" + id)
	stateStr := t.Driver.GetElementById("state_" + id)
	state, _ := strconv.Atoi(stateStr)
	task := Task{
		Name:  &name,
		State: &state,
	}
	t.Tasks[id] = task
	content, _ := json.Marshal(t.Tasks)
	utils.StringToFile("tasks.json", string(content))
	t.Driver.Commit()
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
		todo := liveview.NewDriver("todo", &Todo{})
		return components.NewLayout("home", `<div> {{mount "todo"}} </div>`).Mount(todo)
	})
	e.Logger.Fatal(e.Start(":1323"))
}
