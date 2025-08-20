package liveview_test

import (
	"testing"

	"github.com/arturoeanton/go-echo-live-view/components"
	"github.com/arturoeanton/go-echo-live-view/liveview"
)

func TestPaginationComponent(t *testing.T) {
	t.Run("Create pagination", func(t *testing.T) {
		p := components.NewPagination("test-pagination", 100, 10, 1)

		if p.TotalItems != 100 {
			t.Errorf("Expected 100 total items, got %d", p.TotalItems)
		}

		if p.TotalPages != 10 {
			t.Errorf("Expected 10 total pages, got %d", p.TotalPages)
		}

		if p.CurrentPage != 1 {
			t.Errorf("Expected current page 1, got %d", p.CurrentPage)
		}
	})

	t.Run("Start and End items", func(t *testing.T) {
		p := components.NewPagination("test-pagination", 100, 10, 3)

		if p.StartItem() != 21 {
			t.Errorf("Expected start item 21, got %d", p.StartItem())
		}

		if p.EndItem() != 30 {
			t.Errorf("Expected end item 30, got %d", p.EndItem())
		}
	})

	t.Run("Page navigation", func(t *testing.T) {
		p := components.NewPagination("test-pagination", 100, 10, 5)
		driver := liveview.NewTestDriver("test-pagination", p, nil)
		p.ComponentDriver = driver
		p.Start()

		p.Next(nil)
		if p.CurrentPage != 6 {
			t.Errorf("Expected page 6 after Next, got %d", p.CurrentPage)
		}

		p.Previous(nil)
		if p.CurrentPage != 5 {
			t.Errorf("Expected page 5 after Previous, got %d", p.CurrentPage)
		}

		p.First(nil)
		if p.CurrentPage != 1 {
			t.Errorf("Expected page 1 after First, got %d", p.CurrentPage)
		}

		p.Last(nil)
		if p.CurrentPage != 10 {
			t.Errorf("Expected page 10 after Last, got %d", p.CurrentPage)
		}

		p.GoToPage(float64(7))
		if p.CurrentPage != 7 {
			t.Errorf("Expected page 7 after GoToPage, got %d", p.CurrentPage)
		}
	})

	t.Run("Page numbers generation", func(t *testing.T) {
		tests := []struct {
			name        string
			totalPages  int
			currentPage int
			maxButtons  int
			expected    []int
		}{
			{
				name:        "Few pages",
				totalPages:  5,
				currentPage: 3,
				maxButtons:  7,
				expected:    []int{1, 2, 3, 4, 5},
			},
			{
				name:        "Many pages, at start",
				totalPages:  20,
				currentPage: 2,
				maxButtons:  7,
				expected:    []int{1, 2, 3, 4, 5, -1, 20},
			},
			{
				name:        "Many pages, at end",
				totalPages:  20,
				currentPage: 19,
				maxButtons:  7,
				expected:    []int{1, -1, 16, 17, 18, 19, 20},
			},
			{
				name:        "Many pages, in middle",
				totalPages:  20,
				currentPage: 10,
				maxButtons:  7,
				expected:    []int{1, -1, 9, 10, 11, -1, 20},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				p := &components.Pagination{
					TotalPages:  tt.totalPages,
					CurrentPage: tt.currentPage,
					MaxButtons:  tt.maxButtons,
				}

				result := p.PageNumbers()

				if len(result) != len(tt.expected) {
					t.Errorf("Expected %d page numbers, got %d", len(tt.expected), len(result))
					return
				}

				for i, v := range tt.expected {
					if result[i] != v {
						t.Errorf("At index %d: expected %d, got %d", i, v, result[i])
					}
				}
			})
		}
	})

	t.Run("Update total items", func(t *testing.T) {
		p := components.NewPagination("test-pagination", 100, 10, 5)
		driver := liveview.NewTestDriver("test-pagination", p, nil)
		p.ComponentDriver = driver

		p.UpdateTotal(50)

		if p.TotalItems != 50 {
			t.Errorf("Expected 50 total items, got %d", p.TotalItems)
		}

		if p.TotalPages != 5 {
			t.Errorf("Expected 5 total pages, got %d", p.TotalPages)
		}

		if p.CurrentPage != 5 {
			t.Errorf("Expected current page adjusted to 5, got %d", p.CurrentPage)
		}

		p.UpdateTotal(150)
		if p.TotalPages != 15 {
			t.Errorf("Expected 15 total pages, got %d", p.TotalPages)
		}
	})

	t.Run("OnPageChange callback", func(t *testing.T) {
		pageChanged := false
		newPage := 0

		p := components.NewPagination("test-pagination", 100, 10, 1)
		driver := liveview.NewTestDriver("test-pagination", p, nil)
		p.ComponentDriver = driver
		p.OnPageChange = func(page int) {
			pageChanged = true
			newPage = page
		}
		p.Start()

		p.GoToPage(float64(5))

		if !pageChanged {
			t.Error("Expected OnPageChange to be called")
		}

		if newPage != 5 {
			t.Errorf("Expected new page 5, got %d", newPage)
		}
	})
}

func TestStepperComponent(t *testing.T) {
	t.Run("Create stepper", func(t *testing.T) {
		steps := []components.Step{
			{ID: "step1", Title: "Step 1", Description: "First step"},
			{ID: "step2", Title: "Step 2", Description: "Second step"},
			{ID: "step3", Title: "Step 3", Description: "Third step"},
		}

		s := components.NewStepper("test-stepper", steps)

		if len(s.Steps) != 3 {
			t.Errorf("Expected 3 steps, got %d", len(s.Steps))
		}

		if s.CurrentStep != 0 {
			t.Errorf("Expected current step 0, got %d", s.CurrentStep)
		}

		if !s.Steps[0].Active {
			t.Error("Expected first step to be active")
		}

		if s.Steps[1].Active || s.Steps[2].Active {
			t.Error("Expected only first step to be active")
		}
	})

	t.Run("Step navigation", func(t *testing.T) {
		steps := []components.Step{
			{ID: "step1", Title: "Step 1"},
			{ID: "step2", Title: "Step 2"},
			{ID: "step3", Title: "Step 3"},
		}

		s := components.NewStepper("test-stepper", steps)
		driver := liveview.NewTestDriver("test-stepper", s, nil)
		s.ComponentDriver = driver
		s.Start()

		s.NextStep(nil)
		if s.CurrentStep != 1 {
			t.Errorf("Expected step 1 after NextStep, got %d", s.CurrentStep)
		}
		if !s.Steps[0].Completed {
			t.Error("Expected step 0 to be completed")
		}
		if !s.Steps[1].Active {
			t.Error("Expected step 1 to be active")
		}

		s.PreviousStep(nil)
		if s.CurrentStep != 0 {
			t.Errorf("Expected step 0 after PreviousStep, got %d", s.CurrentStep)
		}
		if s.Steps[0].Completed {
			t.Error("Expected step 0 to not be completed after going back")
		}

		s.GoToStep(float64(2))
		if s.AllowSkip {
			if s.CurrentStep != 2 {
				t.Errorf("Expected step 2 after GoToStep with AllowSkip, got %d", s.CurrentStep)
			}
		} else {
			if s.CurrentStep != 0 {
				t.Errorf("Expected step 0 (no change) without AllowSkip, got %d", s.CurrentStep)
			}
		}
	})

	t.Run("Step error handling", func(t *testing.T) {
		steps := []components.Step{
			{ID: "step1", Title: "Step 1"},
			{ID: "step2", Title: "Step 2"},
		}

		s := components.NewStepper("test-stepper", steps)
		driver := liveview.NewTestDriver("test-stepper", s, nil)
		s.ComponentDriver = driver

		s.SetStepError(0, true)
		if !s.Steps[0].Error {
			t.Error("Expected step 0 to have error")
		}

		s.SetStepError(0, false)
		if s.Steps[0].Error {
			t.Error("Expected step 0 to not have error")
		}

		s.SetStepError(5, true)
	})

	t.Run("Complete current step", func(t *testing.T) {
		steps := []components.Step{
			{ID: "step1", Title: "Step 1"},
			{ID: "step2", Title: "Step 2"},
		}

		s := components.NewStepper("test-stepper", steps)
		driver := liveview.NewTestDriver("test-stepper", s, nil)
		s.ComponentDriver = driver

		s.CompleteCurrentStep()
		if !s.Steps[0].Completed {
			t.Error("Expected current step to be completed")
		}
	})

	t.Run("Reset stepper", func(t *testing.T) {
		steps := []components.Step{
			{ID: "step1", Title: "Step 1", Completed: true},
			{ID: "step2", Title: "Step 2", Active: true, Error: true},
		}

		s := components.NewStepper("test-stepper", steps)
		driver := liveview.NewTestDriver("test-stepper", s, nil)
		s.ComponentDriver = driver
		s.CurrentStep = 1

		s.Reset()

		if s.CurrentStep != 0 {
			t.Errorf("Expected current step 0 after reset, got %d", s.CurrentStep)
		}

		if !s.Steps[0].Active {
			t.Error("Expected first step to be active after reset")
		}

		if s.Steps[0].Completed || s.Steps[0].Error {
			t.Error("Expected first step to be clean after reset")
		}

		if s.Steps[1].Active || s.Steps[1].Completed || s.Steps[1].Error {
			t.Error("Expected second step to be clean after reset")
		}
	})

	t.Run("OnStepChange callback", func(t *testing.T) {
		stepChanged := false
		newStepIndex := -1

		steps := []components.Step{
			{ID: "step1", Title: "Step 1"},
			{ID: "step2", Title: "Step 2"},
		}

		s := components.NewStepper("test-stepper", steps)
		driver := liveview.NewTestDriver("test-stepper", s, nil)
		s.ComponentDriver = driver
		s.OnStepChange = func(stepIndex int) {
			stepChanged = true
			newStepIndex = stepIndex
		}
		s.Start()

		s.NextStep(nil)

		if !stepChanged {
			t.Error("Expected OnStepChange to be called")
		}

		if newStepIndex != 1 {
			t.Errorf("Expected new step index 1, got %d", newStepIndex)
		}
	})
}

func TestSearchBoxComponent(t *testing.T) {
	t.Run("Create search box", func(t *testing.T) {
		sb := components.NewSearchBox("test-search")

		if sb.Placeholder != "Search..." {
			t.Errorf("Expected default placeholder 'Search...', got '%s'", sb.Placeholder)
		}

		if sb.MinChars != 2 {
			t.Errorf("Expected default MinChars 2, got %d", sb.MinChars)
		}

		if sb.DebounceMs != 300 {
			t.Errorf("Expected default DebounceMs 300, got %d", sb.DebounceMs)
		}

		if sb.MaxResults != 10 {
			t.Errorf("Expected default MaxResults 10, got %d", sb.MaxResults)
		}
	})

	t.Run("Update query", func(t *testing.T) {
		sb := components.NewSearchBox("test-search")
		driver := liveview.NewTestDriver("test-search", sb, nil)
		sb.ComponentDriver = driver
		sb.Start()

		sb.UpdateQuery("te")
		if sb.Query != "te" {
			t.Errorf("Expected query 'te', got '%s'", sb.Query)
		}

		if sb.ShowResults {
			t.Error("Expected results not to show with less than MinChars")
		}

		sb.UpdateQuery("test")
		if sb.Query != "test" {
			t.Errorf("Expected query 'test', got '%s'", sb.Query)
		}

		if !sb.ShowResults {
			t.Error("Expected results to show with MinChars or more")
		}

		if !sb.IsLoading {
			t.Error("Expected loading state to be true")
		}
	})

	t.Run("Clear search", func(t *testing.T) {
		sb := components.NewSearchBox("test-search")
		driver := liveview.NewTestDriver("test-search", sb, nil)
		sb.ComponentDriver = driver
		sb.Query = "test query"
		sb.Results = []components.SearchResult{
			{ID: "1", Title: "Result 1"},
		}
		sb.ShowResults = true
		sb.IsLoading = true

		sb.ClearSearch(nil)

		if sb.Query != "" {
			t.Errorf("Expected empty query after clear, got '%s'", sb.Query)
		}

		if len(sb.Results) != 0 {
			t.Errorf("Expected no results after clear, got %d", len(sb.Results))
		}

		if sb.ShowResults {
			t.Error("Expected results to be hidden after clear")
		}

		if sb.IsLoading {
			t.Error("Expected loading to be false after clear")
		}
	})

	t.Run("Select result", func(t *testing.T) {
		resultSelected := false
		var selectedResult components.SearchResult

		sb := components.NewSearchBox("test-search")
		driver := liveview.NewTestDriver("test-search", sb, nil)
		sb.ComponentDriver = driver
		sb.Results = []components.SearchResult{
			{ID: "1", Title: "Result 1", Description: "First result"},
			{ID: "2", Title: "Result 2", Description: "Second result"},
		}
		sb.ShowResults = true
		sb.OnSelect = func(result components.SearchResult) {
			resultSelected = true
			selectedResult = result
		}
		sb.Start()

		sb.SelectResult("2")

		if !resultSelected {
			t.Error("Expected OnSelect to be called")
		}

		if selectedResult.ID != "2" {
			t.Errorf("Expected selected result ID '2', got '%s'", selectedResult.ID)
		}

		if sb.Query != "Result 2" {
			t.Errorf("Expected query to be 'Result 2', got '%s'", sb.Query)
		}

		if sb.ShowResults {
			t.Error("Expected results to be hidden after selection")
		}
	})

	t.Run("Hide results", func(t *testing.T) {
		sb := components.NewSearchBox("test-search")
		driver := liveview.NewTestDriver("test-search", sb, nil)
		sb.ComponentDriver = driver
		sb.ShowResults = true

		sb.HideResults(nil)

		if sb.ShowResults {
			t.Error("Expected results to be hidden")
		}
	})

	t.Run("Set results", func(t *testing.T) {
		sb := components.NewSearchBox("test-search")
		driver := liveview.NewTestDriver("test-search", sb, nil)
		sb.ComponentDriver = driver
		sb.IsLoading = true

		results := []components.SearchResult{
			{ID: "1", Title: "Result 1"},
			{ID: "2", Title: "Result 2"},
		}

		sb.SetResults(results)

		if len(sb.Results) != 2 {
			t.Errorf("Expected 2 results, got %d", len(sb.Results))
		}

		if sb.IsLoading {
			t.Error("Expected loading to be false after setting results")
		}

		if !sb.ShowResults {
			t.Error("Expected results to be shown when results are set")
		}

		sb.SetResults([]components.SearchResult{})
		if sb.ShowResults {
			t.Error("Expected results to be hidden when empty results are set")
		}
	})

	t.Run("Show loading", func(t *testing.T) {
		sb := components.NewSearchBox("test-search")
		driver := liveview.NewTestDriver("test-search", sb, nil)
		sb.ComponentDriver = driver

		sb.ShowLoading(true)
		if !sb.IsLoading {
			t.Error("Expected loading to be true")
		}

		sb.ShowLoading(false)
		if sb.IsLoading {
			t.Error("Expected loading to be false")
		}
	})
}

func TestTabsComponent(t *testing.T) {
	t.Run("Create tabs", func(t *testing.T) {
		tabs := []components.Tab{
			{ID: "tab1", Label: "Tab 1", Content: "Content 1"},
			{ID: "tab2", Label: "Tab 2", Content: "Content 2", Active: true},
			{ID: "tab3", Label: "Tab 3", Content: "Content 3"},
		}

		tc := &components.Tabs{
			Tabs:      tabs,
			ActiveTab: "tab2",
		}

		if tc.ActiveTab != "tab2" {
			t.Errorf("Expected active tab 'tab2', got '%s'", tc.ActiveTab)
		}

		tc.Start()

		if !tc.Tabs[1].Active {
			t.Error("Expected tab2 to be active")
		}
	})

	t.Run("Tab selection", func(t *testing.T) {
		tabChanged := false
		var selectedTabID string

		tabs := []components.Tab{
			{ID: "tab1", Label: "Tab 1", Content: "Content 1", Active: true},
			{ID: "tab2", Label: "Tab 2", Content: "Content 2"},
		}

		tc := &components.Tabs{
			Tabs:      tabs,
			ActiveTab: "tab1",
			OnTabChange: func(tabID string) {
				tabChanged = true
				selectedTabID = tabID
			},
		}

		driver := liveview.NewTestDriver("test-tabs", tc, nil)
		tc.ComponentDriver = driver
		tc.Start()

		tc.SelectTab("tab2")

		if !tabChanged {
			t.Error("Expected OnTabChange to be called")
		}

		if selectedTabID != "tab2" {
			t.Errorf("Expected selected tab ID 'tab2', got '%s'", selectedTabID)
		}

		if tc.ActiveTab != "tab2" {
			t.Errorf("Expected active tab 'tab2', got '%s'", tc.ActiveTab)
		}

		if !tc.Tabs[1].Active {
			t.Error("Expected tab2 to be active")
		}

		if tc.Tabs[0].Active {
			t.Error("Expected tab1 to not be active")
		}
	})

	t.Run("Add tab", func(t *testing.T) {
		tabs := []components.Tab{
			{ID: "tab1", Label: "Tab 1", Content: "Content 1", Active: true},
		}

		tc := &components.Tabs{
			Tabs:      tabs,
			ActiveTab: "tab1",
		}

		driver := liveview.NewTestDriver("test-tabs", tc, nil)
		tc.ComponentDriver = driver

		newTab := components.Tab{
			ID:      "tab2",
			Label:   "Tab 2",
			Content: "Content 2",
		}

		tc.AddTab(newTab)

		if len(tc.Tabs) != 2 {
			t.Errorf("Expected 2 tabs after adding, got %d", len(tc.Tabs))
		}

		if tc.Tabs[1].ID != "tab2" {
			t.Errorf("Expected new tab ID 'tab2', got '%s'", tc.Tabs[1].ID)
		}
	})

	t.Run("Remove tab", func(t *testing.T) {
		tabs := []components.Tab{
			{ID: "tab1", Label: "Tab 1", Content: "Content 1"},
			{ID: "tab2", Label: "Tab 2", Content: "Content 2", Active: true},
			{ID: "tab3", Label: "Tab 3", Content: "Content 3"},
		}

		tc := &components.Tabs{
			Tabs:      tabs,
			ActiveTab: "tab2",
		}

		driver := liveview.NewTestDriver("test-tabs", tc, nil)
		tc.ComponentDriver = driver

		tc.RemoveTab("tab2")

		if len(tc.Tabs) != 2 {
			t.Errorf("Expected 2 tabs after removing, got %d", len(tc.Tabs))
		}

		if tc.ActiveTab != "tab1" {
			t.Errorf("Expected active tab to switch to 'tab1', got '%s'", tc.ActiveTab)
		}

		for _, tab := range tc.Tabs {
			if tab.ID == "tab2" {
				t.Error("Expected tab2 to be removed")
			}
		}
	})

	t.Run("Set active tab", func(t *testing.T) {
		tabs := []components.Tab{
			{ID: "tab1", Label: "Tab 1", Content: "Content 1", Active: true},
			{ID: "tab2", Label: "Tab 2", Content: "Content 2"},
			{ID: "tab3", Label: "Tab 3", Content: "Content 3"},
		}

		tc := &components.Tabs{
			Tabs:      tabs,
			ActiveTab: "tab1",
		}

		driver := liveview.NewTestDriver("test-tabs", tc, nil)
		tc.ComponentDriver = driver

		tc.SetActiveTab("tab3")

		if tc.ActiveTab != "tab3" {
			t.Errorf("Expected active tab 'tab3', got '%s'", tc.ActiveTab)
		}

		if !tc.Tabs[2].Active {
			t.Error("Expected tab3 to be active")
		}

		if tc.Tabs[0].Active || tc.Tabs[1].Active {
			t.Error("Expected only tab3 to be active")
		}
	})
}
