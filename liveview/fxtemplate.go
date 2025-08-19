package liveview

import (
	"text/template"
)

var (
	FuncMapTemplate = template.FuncMap{
		"mount": func(id string) string {
			// Add data-mount attribute to preserve during updates
			return "<span id='mount_span_" + id + "' data-mount='" + id + "'></span>"
		},
		"eqInt": func(value1, value2 int) bool { return value1 == value2 },
	}
)
