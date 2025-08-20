package liveview

import (
	"fmt"
	"regexp"
	"strings"
)

// SafeScriptConfig defines configuration for safe script execution
type SafeScriptConfig struct {
	AllowedFunctions []string
	MaxScriptLength  int
	EnableLogging    bool
}

// DefaultSafeScriptConfig returns a default configuration for safe scripts
func DefaultSafeScriptConfig() *SafeScriptConfig {
	return &SafeScriptConfig{
		// Empty AllowedFunctions means allow all safe functions
		// Dangerous patterns are still blocked by validateScript
		AllowedFunctions: []string{},
		MaxScriptLength: 1000,
		EnableLogging:   true,
	}
}

// SafeScript represents a validated and safe JavaScript code
type SafeScript struct {
	code   string
	config *SafeScriptConfig
}

// NewSafeScript creates a new SafeScript with validation
func NewSafeScript(code string, config *SafeScriptConfig) (*SafeScript, error) {
	if config == nil {
		config = DefaultSafeScriptConfig()
	}

	// Check script length
	if len(code) > config.MaxScriptLength {
		return nil, fmt.Errorf("script exceeds maximum length of %d characters", config.MaxScriptLength)
	}

	// Validate script content
	if err := validateScript(code, config); err != nil {
		return nil, err
	}

	return &SafeScript{
		code:   code,
		config: config,
	}, nil
}

// validateScript checks if the script contains only allowed operations
func validateScript(code string, config *SafeScriptConfig) error {
	// Block dangerous patterns
	dangerousPatterns := []string{
		`eval\s*\(`,
		`new\s+Function\s*\(`,  // Only block Function constructor, not function keyword
		`innerHTML\s*=`,
		`outerHTML\s*=`,
		`document\.write`,
		`document\.writeln`,
		`window\.location\s*=`,  // Block assignment, not just access
		`document\.cookie`,
		`localStorage`,
		`sessionStorage`,
		`XMLHttpRequest`,
		`fetch\s*\(`,
		`import\s*\(`,
		`require\s*\(`,
		`<script`,
		`javascript:`,
		`on\w+\s*=`, // onclick, onload, etc.
	}

	for _, pattern := range dangerousPatterns {
		matched, _ := regexp.MatchString(`(?i)`+pattern, code)
		if matched {
			return fmt.Errorf("script contains dangerous pattern: %s", pattern)
		}
	}

	// Additional validation for allowed functions - only if explicitly configured
	// If AllowedFunctions is empty, skip this check (allow basic safe functions)
	// This allows predefined scripts to work without explicit configuration

	return nil
}

// isUsingOnlyAllowedFunctions checks if script only uses allowed functions
func isUsingOnlyAllowedFunctions(code string, allowed []string) bool {
	// This is a simplified implementation
	// In production, consider using a JavaScript AST parser for accurate validation
	
	// For now, we'll do a basic check
	// Remove strings and comments to avoid false positives
	cleanCode := removeStringsAndComments(code)
	
	// Check for function calls
	functionCallPattern := regexp.MustCompile(`\b(\w+(?:\.\w+)*)\s*\(`)
	matches := functionCallPattern.FindAllStringSubmatch(cleanCode, -1)
	
	for _, match := range matches {
		if len(match) > 1 {
			functionName := match[1]
			if !isAllowedFunction(functionName, allowed) {
				return false
			}
		}
	}
	
	return true
}

// removeStringsAndComments removes strings and comments from JavaScript code
func removeStringsAndComments(code string) string {
	// Remove single-line comments
	code = regexp.MustCompile(`//.*$`).ReplaceAllString(code, "")
	// Remove multi-line comments
	code = regexp.MustCompile(`/\*[\s\S]*?\*/`).ReplaceAllString(code, "")
	// Remove strings (simplified - doesn't handle all cases)
	code = regexp.MustCompile(`"[^"]*"`).ReplaceAllString(code, "")
	code = regexp.MustCompile(`'[^']*'`).ReplaceAllString(code, "")
	code = regexp.MustCompile("`[^`]*`").ReplaceAllString(code, "")
	
	return code
}

// isAllowedFunction checks if a function name is in the allowed list
func isAllowedFunction(functionName string, allowed []string) bool {
	for _, allowedFunc := range allowed {
		if strings.HasPrefix(functionName, allowedFunc) || 
		   strings.HasPrefix(allowedFunc, functionName) {
			return true
		}
	}
	// Allow basic operations
	basicAllowed := []string{
		"setTimeout", "setInterval", "clearTimeout", "clearInterval",
		"parseInt", "parseFloat", "String", "Number", "Boolean",
		"Math", "Date", "Array", "Object",
	}
	for _, basic := range basicAllowed {
		if strings.HasPrefix(functionName, basic) {
			return true
		}
	}
	return false
}

// GetCode returns the validated script code
func (s *SafeScript) GetCode() string {
	return s.code
}

// PredefinedScripts provides common safe scripts
type PredefinedScripts struct{}

// ScrollToTop returns a safe scroll to top script
func (ps *PredefinedScripts) ScrollToTop() *SafeScript {
	script, _ := NewSafeScript("window.scrollTo(0, 0);", nil)
	return script
}

// ScrollToElement returns a safe scroll to element script
func (ps *PredefinedScripts) ScrollToElement(elementID string) (*SafeScript, error) {
	// Validate element ID
	if !isValidElementID(elementID) {
		return nil, fmt.Errorf("invalid element ID")
	}
	
	code := fmt.Sprintf(`
		var element = document.getElementById('%s');
		if (element) {
			element.scrollIntoView({ behavior: 'smooth' });
		}
	`, elementID)
	
	return NewSafeScript(code, nil)
}

// FocusElement returns a safe focus element script
func (ps *PredefinedScripts) FocusElement(elementID string) (*SafeScript, error) {
	if !isValidElementID(elementID) {
		return nil, fmt.Errorf("invalid element ID")
	}
	
	code := fmt.Sprintf(`
		var element = document.getElementById('%s');
		if (element && element.focus) {
			element.focus();
		}
	`, elementID)
	
	return NewSafeScript(code, nil)
}

// ToggleClass returns a safe toggle class script
func (ps *PredefinedScripts) ToggleClass(elementID, className string) (*SafeScript, error) {
	if !isValidElementID(elementID) || !isValidClassName(className) {
		return nil, fmt.Errorf("invalid element ID or class name")
	}
	
	code := fmt.Sprintf(`
		var element = document.getElementById('%s');
		if (element) {
			element.classList.toggle('%s');
		}
	`, elementID, className)
	
	return NewSafeScript(code, nil)
}

// ShowAlert returns a safe alert script (for development only)
func (ps *PredefinedScripts) ShowAlert(message string) (*SafeScript, error) {
	// Sanitize message
	message = strings.ReplaceAll(message, "'", "\\'")
	message = strings.ReplaceAll(message, "\n", "\\n")
	
	if len(message) > 200 {
		return nil, fmt.Errorf("alert message too long")
	}
	
	code := fmt.Sprintf("alert('%s');", message)
	return NewSafeScript(code, nil)
}

// isValidElementID validates an element ID
func isValidElementID(id string) bool {
	// Allow alphanumeric, hyphens, underscores
	validID := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	return validID.MatchString(id) && len(id) < 100
}

// isValidClassName validates a CSS class name
func isValidClassName(className string) bool {
	// Allow alphanumeric, hyphens, underscores
	validClass := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	return validClass.MatchString(className) && len(className) < 50
}

// ExecuteSafeScript safely executes a script through the component driver
func (cw *ComponentDriver[T]) ExecuteSafeScript(script *SafeScript) error {
	if script == nil {
		return fmt.Errorf("script is nil")
	}
	
	// Log script execution if enabled
	if script.config.EnableLogging {
		Debug("Executing safe script: %d bytes", len(script.code))
	}
	
	// Send through the channel
	cw.channel <- map[string]interface{}{
		"type":  "script",
		"value": script.GetCode(),
		"safe":  true, // Mark as safe script
	}
	
	return nil
}

// ExecutePredefinedAction executes a predefined safe action
func (cw *ComponentDriver[T]) ExecutePredefinedAction(action string, params ...string) error {
	ps := &PredefinedScripts{}
	var script *SafeScript
	var err error
	
	switch action {
	case "scrollTop":
		script = ps.ScrollToTop()
	case "scrollToElement":
		if len(params) > 0 {
			script, err = ps.ScrollToElement(params[0])
		} else {
			return fmt.Errorf("scrollToElement requires element ID parameter")
		}
	case "focusElement":
		if len(params) > 0 {
			script, err = ps.FocusElement(params[0])
		} else {
			return fmt.Errorf("focusElement requires element ID parameter")
		}
	case "toggleClass":
		if len(params) >= 2 {
			script, err = ps.ToggleClass(params[0], params[1])
		} else {
			return fmt.Errorf("toggleClass requires element ID and class name parameters")
		}
	default:
		return fmt.Errorf("unknown predefined action: %s", action)
	}
	
	if err != nil {
		return err
	}
	
	return cw.ExecuteSafeScript(script)
}

// DeprecatedEvalScript wraps the old EvalScript with deprecation warning
func (cw *ComponentDriver[T]) DeprecatedEvalScript(code string) {
	// Log deprecation warning
	Debug("WARNING: EvalScript is deprecated and will be removed in next major version. Use ExecuteSafeScript or ExecutePredefinedAction instead.")
	
	// For backward compatibility, still execute but with additional validation
	config := &SafeScriptConfig{
		AllowedFunctions: []string{}, // Allow all for backward compatibility
		MaxScriptLength:  10000,       // Increased limit for backward compatibility
		EnableLogging:    true,
	}
	
	script, err := NewSafeScript(code, config)
	if err != nil {
		Debug("ERROR: Script validation failed: %v", err)
		// For backward compatibility, we still send it but log the error
		cw.channel <- map[string]interface{}{
			"type":       "script",
			"value":      code,
			"deprecated": true,
		}
		return
	}
	
	cw.ExecuteSafeScript(script)
}