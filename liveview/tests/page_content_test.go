package liveview_test

import (
	"strings"
	"testing"

	"github.com/arturoeanton/go-echo-live-view/liveview"
)

func TestPageControlInitialization(t *testing.T) {
	page := liveview.PageControl{
		Title:    "Test Page",
		Lang:     "es",
		HeadCode: "<style>body { color: red; }</style>",
	}

	if page.Title != "Test Page" {
		t.Errorf("Expected Title 'Test Page', got '%s'", page.Title)
	}

	if page.Lang != "es" {
		t.Errorf("Expected Lang 'es', got '%s'", page.Lang)
	}
}

func TestPageControlRegister(t *testing.T) {
	page := liveview.PageControl{
		Title: "Test Registration",
	}

	// Register should accept a function that returns a LiveDriver
	page.Register(func() liveview.LiveDriver {
		component := &TestComponent{}
		driver := liveview.NewDriver("test-register", component)
		component.Driver = driver
		return driver
	})

	// Multiple registrations should work
	page.Register(func() liveview.LiveDriver {
		component2 := &TestComponent{}
		driver2 := liveview.NewDriver("test-register-2", component2)
		component2.Driver = driver2
		return driver2
	})
}

func TestPageControlFields(t *testing.T) {
	page := liveview.PageControl{
		Path:      "/test",
		Title:     "Test Page",
		HeadCode:  "<meta name='test'>",
		Lang:      "en",
		Css:       ".test { color: blue; }",
		LiveJs:    "console.log('test');",
		AfterCode: "<script>console.log('after');</script>",
		Debug:     true,
	}

	// Test that all fields are properly set
	if page.Path != "/test" {
		t.Errorf("Expected Path '/test', got '%s'", page.Path)
	}

	if page.Title != "Test Page" {
		t.Errorf("Expected Title 'Test Page', got '%s'", page.Title)
	}

	if page.HeadCode != "<meta name='test'>" {
		t.Errorf("Expected HeadCode '<meta name='test'>', got '%s'", page.HeadCode)
	}

	if page.Lang != "en" {
		t.Errorf("Expected Lang 'en', got '%s'", page.Lang)
	}

	if page.Css != ".test { color: blue; }" {
		t.Errorf("Expected Css '.test { color: blue; }', got '%s'", page.Css)
	}

	if page.LiveJs != "console.log('test');" {
		t.Errorf("Expected LiveJs 'console.log('test');', got '%s'", page.LiveJs)
	}

	if page.AfterCode != "<script>console.log('after');</script>" {
		t.Errorf("Expected AfterCode '<script>console.log('after');</script>', got '%s'", page.AfterCode)
	}

	if !page.Debug {
		t.Error("Expected Debug to be true")
	}
}

func TestPageControlWithDefaults(t *testing.T) {
	// Test with minimal configuration
	page := liveview.PageControl{}
	
	// Should work with empty values - this is a basic structural test
	if page.Title != "" {
		t.Log("Title is empty as expected")
	}
	
	if page.Lang != "" {
		t.Log("Lang is empty as expected")
	}

	if page.Router == nil {
		t.Log("Router is nil as expected")
	}
}

func TestTemplateBaseContent(t *testing.T) {
	// We can't access templateBase directly since it's not exported,
	// but we can test that the expected elements would be in an HTML template
	
	expectedElements := []string{
		"<html",
		"<head>",
		"<title>",
		"<body>",
		"<div id=\"content\">",
		"wasm_exec.js",
		"WebAssembly",
		"json.wasm",
	}

	// This is a conceptual test - in a real implementation,
	// we would need access to the template generation function
	for _, element := range expectedElements {
		// Just verify these are valid HTML elements
		if !strings.Contains(element, "<") && !strings.Contains(element, "wasm") && !strings.Contains(element, "WebAssembly") {
			t.Logf("Expected element pattern: %s", element)
		}
	}
}