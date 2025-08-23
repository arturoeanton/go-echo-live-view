package main

import (
	"fmt"

	"github.com/arturoeanton/go-echo-live-view/components"
	"github.com/arturoeanton/go-echo-live-view/liveview"
	"github.com/labstack/echo/v4"
)

type TableFilterTest struct {
	*liveview.ComponentDriver[*TableFilterTest]
	Table *components.Table
}

func (c *TableFilterTest) Start() {
	// Create a table with many rows to test filtering
	rows := []components.Row{}
	for i := 1; i <= 50; i++ {
		rows = append(rows, components.Row{
			"id":     fmt.Sprintf("%d", i),
			"name":   fmt.Sprintf("User %d", i),
			"email":  fmt.Sprintf("user%d@example.com", i),
			"status": map[string]string{"Active": "Active", "Inactive": "Inactive"}[[]string{"Active", "Inactive"}[i%2]],
		})
	}
	
	c.Table = liveview.New("test_table", &components.Table{
		Columns: []components.Column{
			{Key: "id", Title: "ID", Width: "80px", Sortable: true},
			{Key: "name", Title: "Name", Sortable: true, Filter: true},
			{Key: "email", Title: "Email", Sortable: true, Filter: true},
			{Key: "status", Title: "Status", Width: "120px", Filter: true},
		},
		Rows:           rows,
		PageSize:       10,
		ShowPagination: true,
		Selectable:     true,
		OnRowClick: func(row components.Row, index int) {
			fmt.Printf("Row clicked: %v\n", row)
		},
	})
	
	c.Mount(c.Table)
	c.Commit()
}

func (c *TableFilterTest) GetTemplate() string {
	return `
	<div style="font-family: Arial, sans-serif; padding: 2rem; max-width: 1200px; margin: 0 auto;">
		<h1>Table Filter Focus Test</h1>
		<p>Type in the filter box below. The input should maintain focus while filtering:</p>
		
		{{mount "test_table"}}
		
		<div style="margin-top: 2rem; padding: 1rem; background: #f0f0f0; border-radius: 8px;">
			<h3>Test Instructions:</h3>
			<ol>
				<li>Click on the filter input box</li>
				<li>Start typing to filter the table (try "User 1" or "user2")</li>
				<li>The input should maintain focus and cursor position</li>
				<li>Table content should update without losing focus</li>
			</ol>
		</div>
	</div>
	`
}

func (c *TableFilterTest) GetDriver() liveview.LiveDriver {
	return c
}

func main() {
	liveview.InitLogger(true)
	liveview.Info("Starting Table Filter Test Server...")
	
	e := echo.New()
	e.Static("/assets", "assets")
	
	home := liveview.PageControl{
		Title:  "Table Filter Test",
		Lang:   "en",
		Path:   "/",
		Router: e,
		Debug:  true,
	}
	
	home.Register(func() liveview.LiveDriver {
		return liveview.NewDriver("table_filter_test", &TableFilterTest{})
	})
	
	fmt.Println("===========================================")
	fmt.Println("Server: http://localhost:8088")
	fmt.Println("===========================================")
	fmt.Println("Test that filter maintains focus while typing")
	fmt.Println()
	
	e.Logger.Fatal(e.Start(":8088"))
}