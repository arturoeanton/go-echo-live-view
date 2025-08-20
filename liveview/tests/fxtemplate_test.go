package liveview_test

import (
	"strings"
	"testing"
	"text/template"

	"github.com/arturoeanton/go-echo-live-view/liveview"
)

func TestMountFunction(t *testing.T) {
	// Get the template function map
	funcMap := liveview.FuncMapTemplate
	
	// Test that mount function exists
	mountFunc, exists := funcMap["mount"]
	if !exists {
		t.Fatal("mount function not found in FuncMapTemplate")
	}

	// Test mount function with different IDs
	result := mountFunc.(func(string) string)("test-component")
	expected := `<span id='mount_span_test-component'></span>`
	
	if result != expected {
		t.Errorf("mount() = %s, want %s", result, expected)
	}

	// Test with different IDs
	result = mountFunc.(func(string) string)("my-fancy-component")
	expected = `<span id='mount_span_my-fancy-component'></span>`
	
	if result != expected {
		t.Errorf("mount() = %s, want %s", result, expected)
	}
}

func TestMountWithEmptyID(t *testing.T) {
	funcMap := liveview.FuncMapTemplate
	mountFunc := funcMap["mount"].(func(string) string)
	
	result := mountFunc("")
	expected := `<span id='mount_span_'></span>`
	
	if result != expected {
		t.Errorf("mount with empty ID = %s, want %s", result, expected)
	}
}

func TestMountWithSpecialCharacters(t *testing.T) {
	funcMap := liveview.FuncMapTemplate
	mountFunc := funcMap["mount"].(func(string) string)
	
	// Test with special characters that might need escaping
	testCases := []struct {
		input    string
		expected string
	}{
		{
			input:    "component-with-dash",
			expected: `<span id='mount_span_component-with-dash'></span>`,
		},
		{
			input:    "component_with_underscore",
			expected: `<span id='mount_span_component_with_underscore'></span>`,
		},
		{
			input:    "component123",
			expected: `<span id='mount_span_component123'></span>`,
		},
	}

	for _, tc := range testCases {
		result := mountFunc(tc.input)
		if result != tc.expected {
			t.Errorf("mount(%q) = %s, want %s", tc.input, result, tc.expected)
		}
	}
}

func TestEqIntFunction(t *testing.T) {
	funcMap := liveview.FuncMapTemplate
	
	// Test that eqInt function exists
	eqIntFunc, exists := funcMap["eqInt"]
	if !exists {
		t.Fatal("eqInt function not found in FuncMapTemplate")
	}

	// Test eqInt function
	testCases := []struct {
		val1     int
		val2     int
		expected bool
	}{
		{1, 1, true},
		{1, 2, false},
		{0, 0, true},
		{-1, -1, true},
		{-1, 1, false},
	}

	for _, tc := range testCases {
		result := eqIntFunc.(func(int, int) bool)(tc.val1, tc.val2)
		if result != tc.expected {
			t.Errorf("eqInt(%d, %d) = %v, want %v", tc.val1, tc.val2, result, tc.expected)
		}
	}
}

func TestTemplateFunctions(t *testing.T) {
	// Test that FuncMapTemplate contains expected functions
	funcs := liveview.FuncMapTemplate
	
	// Check that mount function exists
	if _, exists := funcs["mount"]; !exists {
		t.Error("mount function not found in template functions")
	}

	// Check that eqInt function exists
	if _, exists := funcs["eqInt"]; !exists {
		t.Error("eqInt function not found in template functions")
	}

	// The function map should contain at least these functions
	if len(funcs) < 2 {
		t.Error("Template functions map should contain at least 2 functions")
	}
}

func TestMountInTemplate(t *testing.T) {
	// Test using mount in an actual template
	tmplText := `
	<div class="container">
		{{mount "header"}}
		<main>
			{{mount "content"}}
		</main>
		{{mount "footer"}}
	</div>
	`

	tmpl, err := template.New("test").Funcs(liveview.FuncMapTemplate).Parse(tmplText)
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	var buf strings.Builder
	err = tmpl.Execute(&buf, nil)
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	result := buf.String()

	// Check that the template contains the expected mount results
	if !strings.Contains(result, `<span id='mount_span_header'></span>`) {
		t.Error("Template should contain header mount")
	}
	if !strings.Contains(result, `<span id='mount_span_content'></span>`) {
		t.Error("Template should contain content mount")
	}
	if !strings.Contains(result, `<span id='mount_span_footer'></span>`) {
		t.Error("Template should contain footer mount")
	}
}

func TestEqIntInTemplate(t *testing.T) {
	// Test using eqInt in an actual template
	tmplText := `
	{{if eqInt .Value 42}}
		The answer!
	{{else}}
		Not the answer
	{{end}}
	`

	tmpl, err := template.New("test").Funcs(liveview.FuncMapTemplate).Parse(tmplText)
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	// Test with value 42
	var buf strings.Builder
	err = tmpl.Execute(&buf, map[string]int{"Value": 42})
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	result := buf.String()
	if !strings.Contains(result, "The answer!") {
		t.Error("Template should show 'The answer!' for value 42")
	}

	// Test with different value
	buf.Reset()
	err = tmpl.Execute(&buf, map[string]int{"Value": 10})
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	result = buf.String()
	if !strings.Contains(result, "Not the answer") {
		t.Error("Template should show 'Not the answer' for value != 42")
	}
}

func TestMountHTMLValidity(t *testing.T) {
	funcMap := liveview.FuncMapTemplate
	mountFunc := funcMap["mount"].(func(string) string)
	
	// Test that mount produces valid HTML
	result := mountFunc("test")
	
	// Check for basic HTML structure
	if !strings.HasPrefix(result, "<span") {
		t.Error("Mount result should start with <span")
	}
	
	if !strings.HasSuffix(result, "</span>") {
		t.Error("Mount result should end with </span>")
	}
	
	if !strings.Contains(result, `id='mount_span_test'`) {
		t.Error("Mount result should contain id attribute")
	}
}

func TestMultipleMounts(t *testing.T) {
	funcMap := liveview.FuncMapTemplate
	mountFunc := funcMap["mount"].(func(string) string)
	
	// Test multiple mount calls to ensure they produce unique results
	mounts := []string{
		mountFunc("comp1"),
		mountFunc("comp2"),
		mountFunc("comp3"),
	}

	// All mounts should be different
	for i := 0; i < len(mounts); i++ {
		for j := i + 1; j < len(mounts); j++ {
			if mounts[i] == mounts[j] {
				t.Errorf("Mount results should be unique, but mount[%d] == mount[%d]", i, j)
			}
		}
	}

	// Each should contain its respective ID
	for i, mount := range mounts {
		expectedID := string(rune('1' + i))
		if !strings.Contains(mount, `id='mount_span_comp`+expectedID+`'`) {
			t.Errorf("Mount %d should contain comp%s", i, expectedID)
		}
	}
}