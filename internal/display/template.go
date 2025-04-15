package display

import (
	"strings"
	"text/template"
)

var funcMap = template.FuncMap{
	"replace": strings.ReplaceAll,
}
