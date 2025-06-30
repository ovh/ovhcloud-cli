package display

import (
	"strings"
	"text/template"
)

var funcMap = template.FuncMap{
	"replace": strings.ReplaceAll,
	"join": func(elements []any) string {
		strElements := make([]string, 0, len(elements))
		for _, elem := range elements {
			if str, ok := elem.(string); ok {
				strElements = append(strElements, str)
			} else {
				return ""
			}
		}

		return strings.Join(strElements, ", ")
	},
}
