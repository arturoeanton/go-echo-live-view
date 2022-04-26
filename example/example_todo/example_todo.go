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
	*liveview.ComponentDriver[*Todo]
	ActualTime string
	code       string
	Tasks      map[string]Task
}

func (t *Todo) GetDriver() liveview.LiveDriver {
	return t
}

func (t *Todo) Start() {
	tasksString, _ := utils.FileToString("example/example_todo/tasks.json")
	t.Tasks = make(map[string]Task)
	json.Unmarshal([]byte(tasksString), &(t.Tasks))
	t.Commit()
}

func (t *Todo) GetTemplate() string {
	if t.code == "" {
		t.code, _ = utils.FileToString("example/example_todo/todo.html")
	}
	return t.code
}

func (t *Todo) Add(data interface{}) {
	name := t.GetElementById("new_name")
	stateStr := t.GetElementById("new_state")
	state, _ := strconv.Atoi(stateStr)
	id := uuid.NewString()
	task := Task{
		Name:  &name,
		State: &state,
	}
	t.Tasks[id] = task
	content, _ := json.Marshal(t.Tasks)
	utils.StringToFile("tasks.json", string(content))
	t.Commit()
}

func (t *Todo) RemoveTask(data interface{}) {
	id := data.(string)
	delete(t.Tasks, id)
	content, _ := json.Marshal(t.Tasks)
	utils.StringToFile("tasks.json", string(content))
	t.Commit()
}

func (t *Todo) Change(data interface{}) {
	id := data.(string)
	name := t.GetElementById("name_" + id)
	stateStr := t.GetElementById("state_" + id)
	state, _ := strconv.Atoi(stateStr)
	task := Task{
		Name:  &name,
		State: &state,
	}
	t.Tasks[id] = task
	content, _ := json.Marshal(t.Tasks)
	utils.StringToFile("tasks.json", string(content))
	t.Commit()
}

func main() {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	home := liveview.PageControl{
		Title:    "Todo",
		HeadCode: "example/example_todo/head.html",
		Lang:     "en",
		Path:     "/",
		Router:   e,
		//	Debug:    true,
	}
	home.Register(func() liveview.LiveDriver {
		liveview.New("todo", &Todo{})
		return components.NewLayout("home", `<div> {{mount "todo"}} </div>`)
	})
	e.Logger.Fatal(e.Start(":1323"))
}
