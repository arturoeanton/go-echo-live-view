package liveview

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"text/template"
)

// SmartCommit renders and updates only the parts that changed, preserving mounted components
func (cw *ComponentDriver[T]) SmartCommit() {
	defer func() {
		if r := recover(); r != nil {
			Error("Recovered in SmartCommit: %v", r)
		}
	}()
	
	LogComponent(cw.IdComponent, "SmartCommit", "Starting")
	
	// Get current template
	rawTemplate := cw.Component.GetTemplate()
	
	// Parse template
	t := template.Must(template.New("component").Funcs(FuncMapTemplate).Parse(rawTemplate))
	buf := new(bytes.Buffer)
	err := t.Execute(buf, cw.Component)
	if err != nil {
		Error("Template execution error: %v", err)
		return
	}
	
	html := buf.String()
	LogTemplate(cw.IdComponent, "SmartRendered", fmt.Sprintf("%d bytes", len(html)))
	
	// Instead of replacing all HTML, update only non-mount sections
	cw.smartUpdate(html)
}

func (cw *ComponentDriver[T]) smartUpdate(newHTML string) {
	// Find all mount points in the new HTML
	mountRegex := regexp.MustCompile(`<span[^>]*id="mount_span_[^"]*"[^>]*>.*?</span>`)
	mounts := mountRegex.FindAllString(newHTML, -1)
	
	// Store mount IDs to preserve
	mountIDs := make(map[string]bool)
	for _, mount := range mounts {
		idRegex := regexp.MustCompile(`id="(mount_span_[^"]*)"`)
		if matches := idRegex.FindStringSubmatch(mount); len(matches) > 1 {
			mountIDs[matches[1]] = true
		}
	}
	
	// Create a script to update only the necessary parts
	var scriptBuilder strings.Builder
	scriptBuilder.WriteString("(function() {\n")
	
	// Update non-mount content
	scriptBuilder.WriteString(fmt.Sprintf(`
		var container = document.getElementById('%s');
		if (!container) return;
		
		// Save mounted components
		var mounts = {};
		var mountElements = container.querySelectorAll('[id^="mount_span_"]');
		mountElements.forEach(function(el) {
			mounts[el.id] = el.cloneNode(true);
		});
		
		// Update HTML
		var temp = document.createElement('div');
		temp.innerHTML = %s;
		
		// Restore mounted components
		for (var id in mounts) {
			var placeholder = temp.querySelector('#' + id);
			if (placeholder && mounts[id]) {
				placeholder.parentNode.replaceChild(mounts[id], placeholder);
			}
		}
		
		// Replace content
		container.innerHTML = temp.innerHTML;
	`, cw.GetID(), fmt.Sprintf("`%s`", strings.ReplaceAll(newHTML, "`", "\\`"))))
	
	scriptBuilder.WriteString("\n})();")
	
	// Execute the smart update script
	cw.EvalScript(scriptBuilder.String())
}

// CommitPreserveMounts commits but preserves mounted components
func (cw *ComponentDriver[T]) CommitPreserveMounts() {
	cw.SmartCommit()
}