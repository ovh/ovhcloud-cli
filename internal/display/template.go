package display

import (
	"fmt"
	"math"
	"strings"
	"text/template"
)

var (
	funcMap = template.FuncMap{
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
		"formatByteSize": formatByteSize,
	}

	// unit holds the units for formatting bytes
	units = map[int][]string{
		1000: {"B", "KB", "MB", "GB", "TB", "PB", "EB", "ZB", "YB"},
		1024: {"B", "KiB", "MiB", "GiB", "TiB", "PiB", "EiB", "ZiB", "YiB"},
	}
)

// formatByteSize formats a byte value into a human-readable string.
// decimals: number of decimal places
// format: 1000 (SI) or 1024 (IEC)
func formatByteSize(bytes float64, decimals int, format int) string {
	if bytes == 0 {
		return "0"
	}

	unitList, ok := units[format]
	if !ok {
		unitList = units[1000]
	}

	i := int(math.Floor(math.Log(bytes) / math.Log(float64(format))))
	if i >= len(unitList) {
		i = len(unitList) - 1
	}
	value := bytes / math.Pow(float64(format), float64(i))

	return fmt.Sprintf("%.*f %s", decimals, value, unitList[i])
}
