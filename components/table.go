package components

import (
	"fmt"
	"sort"
	"strings"

	"github.com/arturoeanton/go-echo-live-view/liveview"
)

type Column struct {
	Key      string
	Title    string
	Width    string
	Sortable bool
	Filter   bool
}

type Row map[string]interface{}

type Table struct {
	*liveview.ComponentDriver[*Table]
	Columns       []Column
	Rows          []Row
	FilteredRows  []Row
	CurrentPage   int
	PageSize      int
	TotalPages    int
	SortColumn    string
	SortDirection string
	FilterText    string
	Selectable    bool
	SelectedRows  map[int]bool
	OnRowClick    func(row Row, index int)
	OnSort        func(column string, direction string)
	ShowPagination bool
}

func (t *Table) Start() {
	if t.PageSize == 0 {
		t.PageSize = 10
	}
	if t.CurrentPage == 0 {
		t.CurrentPage = 1
	}
	if t.SelectedRows == nil {
		t.SelectedRows = make(map[int]bool)
	}
	if t.ShowPagination {
		t.updatePagination()
	}
	t.FilteredRows = t.Rows
	t.Commit()
}

func (t *Table) GetTemplate() string {
	return `
	<div id="{{.IdComponent}}" class="table-container">
		<style>
			.table-container {
				width: 100%;
				overflow-x: auto;
			}
			.table-controls {
				display: flex;
				justify-content: space-between;
				align-items: center;
				margin-bottom: 1rem;
				gap: 1rem;
			}
			.table-filter {
				display: flex;
				align-items: center;
				gap: 0.5rem;
			}
			.table-filter input {
				padding: 0.5rem;
				border: 1px solid #ddd;
				border-radius: 4px;
				font-size: 0.875rem;
				width: 250px;
			}
			.data-table {
				width: 100%;
				border-collapse: collapse;
				background: white;
				box-shadow: 0 1px 3px rgba(0,0,0,0.1);
			}
			.data-table th {
				background: #f5f5f5;
				padding: 0.75rem;
				text-align: left;
				font-weight: 600;
				color: #333;
				border-bottom: 2px solid #ddd;
				position: sticky;
				top: 0;
				z-index: 10;
			}
			.data-table th.sortable {
				cursor: pointer;
				user-select: none;
			}
			.data-table th.sortable:hover {
				background: #ececec;
			}
			.sort-indicator {
				display: inline-block;
				margin-left: 0.5rem;
				font-size: 0.75rem;
			}
			.data-table td {
				padding: 0.75rem;
				border-bottom: 1px solid #eee;
			}
			.data-table tr:hover {
				background: #f9f9f9;
			}
			.data-table tr.selected {
				background: #e3f2fd;
			}
			.table-checkbox {
				width: 20px;
				height: 20px;
				cursor: pointer;
			}
			.table-pagination {
				display: flex;
				justify-content: center;
				align-items: center;
				gap: 0.5rem;
				margin-top: 1rem;
			}
			.pagination-button {
				padding: 0.5rem 0.75rem;
				border: 1px solid #ddd;
				background: white;
				cursor: pointer;
				border-radius: 4px;
				font-size: 0.875rem;
			}
			.pagination-button:hover:not(:disabled) {
				background: #f5f5f5;
			}
			.pagination-button:disabled {
				opacity: 0.5;
				cursor: not-allowed;
			}
			.pagination-button.active {
				background: #4CAF50;
				color: white;
				border-color: #4CAF50;
			}
			.pagination-info {
				padding: 0.5rem 1rem;
				font-size: 0.875rem;
				color: #666;
			}
			.empty-state {
				text-align: center;
				padding: 3rem;
				color: #999;
			}
		</style>
		
		<div class="table-controls">
			<div class="table-filter">
				<span>🔍</span>
				<input 
					type="text" 
					placeholder="Filter rows..."
					value="{{.FilterText}}"
					oninput="send_event('{{.IdComponent}}', 'Filter', this.value)"
				>
			</div>
			<div class="pagination-info">
				Showing {{.GetStartIndex}} - {{.GetEndIndex}} of {{len .FilteredRows}} rows
			</div>
		</div>
		
		{{if .FilteredRows}}
		<table class="data-table">
			<thead>
				<tr>
					{{if .Selectable}}
					<th style="width: 40px;">
						<input 
							type="checkbox" 
							class="table-checkbox"
							onchange="send_event('{{.IdComponent}}', 'SelectAll', this.checked)"
						>
					</th>
					{{end}}
					{{range .Columns}}
					<th 
						{{if .Width}}style="width: {{.Width}}"{{end}}
						{{if .Sortable}}
						class="sortable"
						onclick="send_event('{{$.IdComponent}}', 'Sort', '{{.Key}}')"
						{{end}}
					>
						{{.Title}}
						{{if .Sortable}}
							{{if eq $.SortColumn .Key}}
								<span class="sort-indicator">
									{{if eq $.SortDirection "asc"}}▲{{else}}▼{{end}}
								</span>
							{{end}}
						{{end}}
					</th>
					{{end}}
				</tr>
			</thead>
			<tbody>
				{{range $i, $row := .GetPageRows}}
				<tr 
					{{if $.Selectable}}
						{{if index $.SelectedRows $i}}class="selected"{{end}}
					{{end}}
					{{if $.OnRowClick}}
						onclick="send_event('{{$.IdComponent}}', 'RowClick', {{$i}})"
						style="cursor: pointer;"
					{{end}}
				>
					{{if $.Selectable}}
					<td>
						<input 
							type="checkbox" 
							class="table-checkbox"
							{{if index $.SelectedRows $i}}checked{{end}}
							onchange="send_event('{{$.IdComponent}}', 'SelectRow', JSON.stringify({index: {{$i}}, checked: this.checked}))"
							onclick="event.stopPropagation()"
						>
					</td>
					{{end}}
					{{range $.Columns}}
					<td>{{index $row .Key}}</td>
					{{end}}
				</tr>
				{{end}}
			</tbody>
		</table>
		
		{{if .ShowPagination}}
		{{if gt .TotalPages 1}}
		<div class="table-pagination">
			<button 
				class="pagination-button" 
				{{if eq .CurrentPage 1}}disabled{{end}}
				onclick="send_event('{{.IdComponent}}', 'Page', 1)"
			>
				First
			</button>
			<button 
				class="pagination-button" 
				{{if eq .CurrentPage 1}}disabled{{end}}
				onclick="send_event('{{.IdComponent}}', 'Page', {{.CurrentPage}} - 1)"
			>
				Previous
			</button>
			
			{{range .GetPageNumbers}}
			<button 
				class="pagination-button {{if eq . $.CurrentPage}}active{{end}}"
				onclick="send_event('{{$.IdComponent}}', 'Page', {{.}})"
			>
				{{.}}
			</button>
			{{end}}
			
			<button 
				class="pagination-button" 
				{{if eq .CurrentPage .TotalPages}}disabled{{end}}
				onclick="send_event('{{.IdComponent}}', 'Page', {{.CurrentPage}} + 1)"
			>
				Next
			</button>
			<button 
				class="pagination-button" 
				{{if eq .CurrentPage .TotalPages}}disabled{{end}}
				onclick="send_event('{{.IdComponent}}', 'Page', {{.TotalPages}})"
			>
				Last
			</button>
		</div>
		{{end}}
		{{end}}
		{{else}}
		<div class="empty-state">
			<div style="font-size: 3rem; margin-bottom: 1rem;">📊</div>
			<div>No data available</div>
		</div>
		{{end}}
	</div>
	`
}

func (t *Table) GetDriver() liveview.LiveDriver {
	return t
}

// updateTableContent updates only the table content without re-rendering the entire component
func (t *Table) updateTableContent() {
	// Generate only the table body HTML
	bodyHTML := t.generateTableBodyHTML()
	
	// Use JavaScript to update only the table body
	script := fmt.Sprintf(`
		(function() {
			var tableBody = document.querySelector('#%s tbody');
			if (tableBody) {
				tableBody.innerHTML = %s;
			}
			
			// Update pagination info
			var paginationInfo = document.querySelector('#%s .pagination-info');
			if (paginationInfo) {
				paginationInfo.innerHTML = 'Showing %d - %d of %d rows';
			}
			
			// Update pagination buttons if needed
			var currentPageSpan = document.querySelector('#%s .current-page');
			if (currentPageSpan) {
				currentPageSpan.innerHTML = '%d';
			}
		})();
	`, 
		t.IdComponent,
		"`" + bodyHTML + "`",
		t.IdComponent,
		t.GetStartIndex(), t.GetEndIndex(), len(t.FilteredRows),
		t.IdComponent,
		t.CurrentPage,
	)
	
	t.EvalScript(script)
}

// generateTableBodyHTML generates only the tbody content
func (t *Table) generateTableBodyHTML() string {
	var html strings.Builder
	
	for i, row := range t.GetPageRows() {
		html.WriteString("<tr")
		
		if t.Selectable && t.SelectedRows[i] {
			html.WriteString(` class="selected"`)
		}
		
		if t.OnRowClick != nil {
			html.WriteString(fmt.Sprintf(` onclick="send_event('%s', 'RowClick', %d)" style="cursor: pointer;"`, t.IdComponent, i))
		}
		
		html.WriteString(">")
		
		if t.Selectable {
			html.WriteString(`<td><input type="checkbox" class="table-checkbox"`)
			if t.SelectedRows[i] {
				html.WriteString(` checked`)
			}
			html.WriteString(fmt.Sprintf(` onchange="send_event('%s', 'SelectRow', JSON.stringify({index: %d, checked: this.checked}))" onclick="event.stopPropagation()"></td>`, t.IdComponent, i))
		}
		
		for _, col := range t.Columns {
			html.WriteString(fmt.Sprintf("<td>%v</td>", row[col.Key]))
		}
		
		html.WriteString("</tr>")
	}
	
	return html.String()
}

func (t *Table) Filter(data interface{}) {
	t.FilterText = fmt.Sprint(data)
	t.applyFilter()
	t.CurrentPage = 1
	t.updatePagination()
	
	// Don't do full Commit() to preserve input focus
	// Instead, update only the table body and pagination
	t.updateTableContent()
}

func (t *Table) Sort(data interface{}) {
	column := fmt.Sprint(data)
	
	if t.SortColumn == column {
		if t.SortDirection == "asc" {
			t.SortDirection = "desc"
		} else {
			t.SortDirection = "asc"
		}
	} else {
		t.SortColumn = column
		t.SortDirection = "asc"
	}
	
	t.applySort()
	
	if t.OnSort != nil {
		t.OnSort(t.SortColumn, t.SortDirection)
	}
	
	t.Commit()
}

func (t *Table) Page(data interface{}) {
	page := 0
	switch v := data.(type) {
	case float64:
		page = int(v)
	case string:
		fmt.Sscanf(v, "%d", &page)
	}
	
	if page >= 1 && page <= t.TotalPages {
		t.CurrentPage = page
		t.Commit()
	}
}

func (t *Table) SelectRow(data interface{}) {
	params := make(map[string]interface{})
	if _, ok := data.(string); ok {
		params["index"] = 0
		params["checked"] = false
	}
	
	if index, ok := params["index"].(float64); ok {
		if checked, ok := params["checked"].(bool); ok {
			if checked {
				t.SelectedRows[int(index)] = true
			} else {
				delete(t.SelectedRows, int(index))
			}
			t.Commit()
		}
	}
}

func (t *Table) SelectAll(data interface{}) {
	selectAll := false
	if b, ok := data.(bool); ok {
		selectAll = b
	} else if s, ok := data.(string); ok {
		selectAll = s == "true"
	}
	
	if selectAll {
		for i := range t.GetPageRows() {
			actualIndex := (t.CurrentPage-1)*t.PageSize + i
			t.SelectedRows[actualIndex] = true
		}
	} else {
		t.SelectedRows = make(map[int]bool)
	}
	
	t.Commit()
}

func (t *Table) RowClick(data interface{}) {
	index := 0
	switch v := data.(type) {
	case float64:
		index = int(v)
	case string:
		fmt.Sscanf(v, "%d", &index)
	}
	
	actualIndex := (t.CurrentPage-1)*t.PageSize + index
	if actualIndex < len(t.FilteredRows) && t.OnRowClick != nil {
		t.OnRowClick(t.FilteredRows[actualIndex], actualIndex)
	}
}

func (t *Table) applyFilter() {
	if t.FilterText == "" {
		t.FilteredRows = t.Rows
		return
	}
	
	filter := strings.ToLower(t.FilterText)
	t.FilteredRows = []Row{}
	
	for _, row := range t.Rows {
		match := false
		for _, col := range t.Columns {
			if col.Filter {
				value := strings.ToLower(fmt.Sprint(row[col.Key]))
				if strings.Contains(value, filter) {
					match = true
					break
				}
			}
		}
		if match || !hasFilterableColumns(t.Columns) {
			for _, value := range row {
				if strings.Contains(strings.ToLower(fmt.Sprint(value)), filter) {
					match = true
					break
				}
			}
		}
		if match {
			t.FilteredRows = append(t.FilteredRows, row)
		}
	}
}

func (t *Table) applySort() {
	if t.SortColumn == "" {
		return
	}
	
	sort.Slice(t.FilteredRows, func(i, j int) bool {
		val1 := fmt.Sprint(t.FilteredRows[i][t.SortColumn])
		val2 := fmt.Sprint(t.FilteredRows[j][t.SortColumn])
		
		if t.SortDirection == "asc" {
			return val1 < val2
		}
		return val1 > val2
	})
}

func (t *Table) updatePagination() {
	if t.PageSize > 0 {
		t.TotalPages = (len(t.FilteredRows) + t.PageSize - 1) / t.PageSize
		if t.TotalPages == 0 {
			t.TotalPages = 1
		}
	}
}

func (t *Table) GetPageRows() []Row {
	if !t.ShowPagination {
		return t.FilteredRows
	}
	
	start := (t.CurrentPage - 1) * t.PageSize
	end := start + t.PageSize
	
	if start >= len(t.FilteredRows) {
		return []Row{}
	}
	
	if end > len(t.FilteredRows) {
		end = len(t.FilteredRows)
	}
	
	return t.FilteredRows[start:end]
}

func (t *Table) GetStartIndex() int {
	if len(t.FilteredRows) == 0 {
		return 0
	}
	return (t.CurrentPage-1)*t.PageSize + 1
}

func (t *Table) GetEndIndex() int {
	end := t.CurrentPage * t.PageSize
	if end > len(t.FilteredRows) {
		end = len(t.FilteredRows)
	}
	return end
}

func (t *Table) GetPageNumbers() []int {
	pages := []int{}
	maxPages := 5
	
	start := t.CurrentPage - 2
	if start < 1 {
		start = 1
	}
	
	end := start + maxPages - 1
	if end > t.TotalPages {
		end = t.TotalPages
		start = end - maxPages + 1
		if start < 1 {
			start = 1
		}
	}
	
	for i := start; i <= end; i++ {
		pages = append(pages, i)
	}
	
	return pages
}

func (t *Table) UpdateData(rows []Row) {
	t.Rows = rows
	t.FilteredRows = rows
	t.CurrentPage = 1
	t.updatePagination()
	t.Commit()
}

func (t *Table) GetSelectedRows() []Row {
	selected := []Row{}
	for index := range t.SelectedRows {
		if index < len(t.FilteredRows) {
			selected = append(selected, t.FilteredRows[index])
		}
	}
	return selected
}

func (t *Table) ClearSelection() {
	t.SelectedRows = make(map[int]bool)
	t.Commit()
}

func hasFilterableColumns(columns []Column) bool {
	for _, col := range columns {
		if col.Filter {
			return true
		}
	}
	return false
}