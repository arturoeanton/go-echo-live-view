package liveview

import (
	"text/template"
)

var (
	FuncMapTemplate = template.FuncMap{
		"mount": func(id string) string {
			return "<span id='mount_span_" + id + "'></span>"
		},
	}
)
