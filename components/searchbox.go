package components

import (
	"strings"
	"time"

	"github.com/arturoeanton/go-echo-live-view/liveview"
)

type SearchResult struct {
	ID          string
	Title       string
	Description string
	Category    string
	Icon        string
	URL         string
}

type SearchBox struct {
	*liveview.ComponentDriver[*SearchBox]
	
	Query           string
	Results         []SearchResult
	ShowResults     bool
	IsLoading       bool
	Placeholder     string
	MinChars        int
	DebounceMs      int
	MaxResults      int
	OnSearch        func(query string) []SearchResult
	OnSelect        func(result SearchResult)
	lastSearchTime  time.Time
}

func NewSearchBox(id string) *SearchBox {
	return &SearchBox{
		Placeholder: "Search...",
		MinChars:    2,
		DebounceMs:  300,
		MaxResults:  10,
		Results:     []SearchResult{},
	}
}

func (s *SearchBox) Start() {
	// Events are registered directly on the ComponentDriver
	if s.ComponentDriver != nil {
		s.ComponentDriver.Events["UpdateQuery"] = func(c *SearchBox, data interface{}) {
			c.UpdateQuery(data)
		}
		s.ComponentDriver.Events["SelectResult"] = func(c *SearchBox, data interface{}) {
			c.SelectResult(data)
		}
		s.ComponentDriver.Events["ClearSearch"] = func(c *SearchBox, data interface{}) {
			c.ClearSearch(data)
		}
		s.ComponentDriver.Events["HideResults"] = func(c *SearchBox, data interface{}) {
			c.HideResults(data)
		}
	}
}

func (s *SearchBox) GetTemplate() string {
	return `
<div class="search-box-container" id="{{.IdComponent}}">
	<style>
		.search-box-container {
			position: relative;
			width: 100%;
			max-width: 600px;
		}
		
		.search-input-wrapper {
			position: relative;
		}
		
		.search-input {
			width: 100%;
			padding: 0.75rem 3rem 0.75rem 1rem;
			border: 1px solid #d1d5db;
			border-radius: 0.5rem;
			font-size: 1rem;
			transition: all 0.2s;
			background: white;
		}
		
		.search-input:focus {
			outline: none;
			border-color: #3b82f6;
			box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
		}
		
		.search-icons {
			position: absolute;
			right: 0.75rem;
			top: 50%;
			transform: translateY(-50%);
			display: flex;
			gap: 0.5rem;
			align-items: center;
		}
		
		.search-icon, .clear-icon, .loading-icon {
			width: 20px;
			height: 20px;
			color: #6b7280;
			cursor: pointer;
		}
		
		.clear-icon:hover {
			color: #374151;
		}
		
		.loading-icon {
			animation: spin 1s linear infinite;
		}
		
		@keyframes spin {
			from { transform: rotate(0deg); }
			to { transform: rotate(360deg); }
		}
		
		.search-results {
			position: absolute;
			top: calc(100% + 0.5rem);
			left: 0;
			right: 0;
			background: white;
			border: 1px solid #d1d5db;
			border-radius: 0.5rem;
			box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.1);
			max-height: 400px;
			overflow-y: auto;
			z-index: 50;
			display: none;
		}
		
		.search-results.show {
			display: block;
		}
		
		.search-result-item {
			padding: 0.75rem 1rem;
			cursor: pointer;
			border-bottom: 1px solid #f3f4f6;
			transition: background 0.15s;
		}
		
		.search-result-item:hover {
			background: #f9fafb;
		}
		
		.search-result-item:last-child {
			border-bottom: none;
		}
		
		.result-title {
			font-weight: 600;
			color: #111827;
			margin-bottom: 0.25rem;
		}
		
		.result-description {
			font-size: 0.875rem;
			color: #6b7280;
			margin-bottom: 0.25rem;
		}
		
		.result-category {
			display: inline-block;
			font-size: 0.75rem;
			color: #3b82f6;
			background: #eff6ff;
			padding: 0.125rem 0.5rem;
			border-radius: 0.25rem;
		}
		
		.no-results {
			padding: 2rem;
			text-align: center;
			color: #6b7280;
		}
		
		.search-hint {
			padding: 0.5rem 1rem;
			font-size: 0.875rem;
			color: #6b7280;
			background: #f9fafb;
			border-bottom: 1px solid #e5e7eb;
		}
	</style>
	
	<div class="search-input-wrapper">
		<input type="text" 
			   class="search-input" 
			   placeholder="{{.Placeholder}}"
			   value="{{.Query}}"
			   oninput="send_event('{{.IdComponent}}', 'UpdateQuery', this.value)"
			   onfocus="send_event('{{.IdComponent}}', 'ShowResults', null)"
			   autocomplete="off">
		
		<div class="search-icons">
			{{if .IsLoading}}
				<svg class="loading-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" 
						  d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
				</svg>
			{{else if .Query}}
				<svg class="clear-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor"
					 onclick="send_event('{{.IdComponent}}', 'ClearSearch', null)">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" 
						  d="M6 18L18 6M6 6l12 12" />
				</svg>
			{{else}}
				<svg class="search-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" 
						  d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
				</svg>
			{{end}}
		</div>
	</div>
	
	<div class="search-results {{if .ShowResults}}show{{end}}">
		{{if .Query}}
			{{if .IsLoading}}
				<div class="search-hint">Searching...</div>
			{{else if .Results}}
				<div class="search-hint">{{len .Results}} results for "{{.Query}}"</div>
				{{range .Results}}
				<div class="search-result-item" 
					 onclick="send_event('{{$.IdComponent}}', 'SelectResult', '{{.ID}}')">
					<div class="result-title">{{.Title}}</div>
					{{if .Description}}
						<div class="result-description">{{.Description}}</div>
					{{end}}
					{{if .Category}}
						<span class="result-category">{{.Category}}</span>
					{{end}}
				</div>
				{{end}}
			{{else}}
				<div class="no-results">
					No results found for "{{.Query}}"
				</div>
			{{end}}
		{{else if ge (len .Query) .MinChars}}
			<div class="search-hint">Type at least {{.MinChars}} characters to search</div>
		{{end}}
	</div>
	
	<div onclick="send_event('{{.IdComponent}}', 'HideResults', null)" 
		 style="position: fixed; inset: 0; z-index: 40; display: {{if .ShowResults}}block{{else}}none{{end}};">
	</div>
</div>
`
}

func (s *SearchBox) GetDriver() liveview.LiveDriver {
	return s
}

func (s *SearchBox) UpdateQuery(data interface{}) {
	query := strings.TrimSpace(data.(string))
	s.Query = query
	
	if len(query) >= s.MinChars {
		s.ShowResults = true
		s.IsLoading = true
		s.Commit()
		
		go func() {
			time.Sleep(time.Duration(s.DebounceMs) * time.Millisecond)
			
			if s.OnSearch != nil {
				results := s.OnSearch(query)
				if len(results) > s.MaxResults {
					results = results[:s.MaxResults]
				}
				s.Results = results
			}
			
			s.IsLoading = false
			s.Commit()
		}()
	} else {
		s.Results = []SearchResult{}
		s.ShowResults = false
		s.IsLoading = false
		s.Commit()
	}
}

func (s *SearchBox) SelectResult(data interface{}) {
	resultID := data.(string)
	
	for _, result := range s.Results {
		if result.ID == resultID {
			s.Query = result.Title
			s.ShowResults = false
			
			if s.OnSelect != nil {
				s.OnSelect(result)
			}
			
			s.Commit()
			break
		}
	}
}

func (s *SearchBox) ClearSearch(data interface{}) {
	s.Query = ""
	s.Results = []SearchResult{}
	s.ShowResults = false
	s.IsLoading = false
	s.Commit()
}

func (s *SearchBox) HideResults(data interface{}) {
	s.ShowResults = false
	s.Commit()
}

func (s *SearchBox) SetResults(results []SearchResult) {
	s.Results = results
	s.IsLoading = false
	s.ShowResults = len(results) > 0
	s.Commit()
}

func (s *SearchBox) ShowLoading(show bool) {
	s.IsLoading = show
	s.Commit()
}