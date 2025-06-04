package display

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"text/template"

	"github.com/PaesslerAG/gval"
	fxdisplay "github.com/amstuta/fx/display"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/ghodss/yaml"
	"gopkg.in/ini.v1"
	"stash.ovh.net/api/ovh-cli/internal/filters"
)

// Common flags used by all subcommands to control output format (json, yaml)
type OutputFormat struct {
	JsonOutput, YamlOutput, InteractiveOutput bool
	CustomFormat                              string
}

func renderCustomFormat(value any, format string) {
	ev, err := gval.Full(filters.AdditionalEvaluators...).NewEvaluable(format)
	if err != nil {
		ExitError("invalid format given: %s", err)
	}

	switch reflect.TypeOf(value).Kind() {
	case reflect.Slice:
		for _, val := range value.([]map[string]any) {
			out, err := ev(context.Background(), val)
			if err != nil {
				ExitError("couldn't extract data according to given format: %s", err)
			}

			outBytes, err := json.Marshal(out)
			if err != nil {
				ExitError("error marshalling result")
			}
			fmt.Println(string(outBytes))
		}
	default:
		out, err := ev(context.Background(), value)
		if err != nil {
			ExitError("couldn't extract data according to given format: %s", err)
		}

		outBytes, err := json.Marshal(out)
		if err != nil {
			ExitError("error marshalling result")
		}
		fmt.Print(string(outBytes))
	}
}

func RenderTable(values []map[string]any, columnsToDisplay []string, outputFormat *OutputFormat) {
	switch {
	case outputFormat.CustomFormat != "":
		renderCustomFormat(values, outputFormat.CustomFormat)
		return
	case outputFormat.InteractiveOutput:
		bytes, err := json.Marshal(values)
		if err != nil {
			ExitError("error preparing interactive output: %s", err)
		}
		fxdisplay.Display(bytes, "")
		return
	case outputFormat.YamlOutput:
		if err := prettyPrintYAML(values); err != nil {
			ExitError("error displaying YAML results: %s", err)
		}
		return
	case outputFormat.JsonOutput:
		if err := prettyPrintJSON(values); err != nil {
			ExitError("error displaying JSON results: %s", err)
		}
		return
	}

	var (
		rows      [][]string
		selectors gval.Evaluables
	)

	columnsTitles := make([]string, 0, len(columnsToDisplay))
	for _, col := range columnsToDisplay {
		// If column to display contains an alias, use it as column title
		parts := strings.SplitN(col, " ", 2)
		if len(parts) == 2 {
			col = parts[0]
			columnsTitles = append(columnsTitles, parts[1])
		} else {
			columnsTitles = append(columnsTitles, col)
		}

		// Create selector to extract value at given key
		evaluator, err := gval.Base().NewEvaluable(col)
		if err != nil {
			ExitError("invalid column to display %q: %s", col, err)
		}
		selectors = append(selectors, evaluator)
	}

	// Extract values to display
	for _, value := range values {
		var row []string
		for _, selector := range selectors {
			val, err := selector(context.Background(), value)
			if err != nil {
				ExitError("failed to select row field: %s", err)
			}

			switch val.(type) {
			case float32, float64:
				// TODO: default formatting without decimals, may cause issues at some point
				row = append(row, fmt.Sprintf("%.0f", val))
			default:
				row = append(row, fmt.Sprintf("%v", val))
			}
		}

		rows = append(rows, row)
	}

	var (
		purple = lipgloss.Color("99")
		gray   = lipgloss.Color("245")

		headerStyle = lipgloss.NewStyle().Foreground(purple).Bold(true).Align(lipgloss.Center)
		cellStyle   = lipgloss.NewStyle().Padding(0, 1)
		oddRowStyle = cellStyle.Foreground(gray)
	)

	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(purple)).
		StyleFunc(func(row, col int) lipgloss.Style {
			switch {
			case row == table.HeaderRow:
				return headerStyle
			default:
				return oddRowStyle
			}
		}).
		Headers(columnsTitles...).
		Rows(rows...)

	fmt.Println(t)
	fmt.Println("💡 Use option --json or --yaml to get the raw output with all information")
}

func RenderConfigTable(cfg *ini.File) {
	var (
		rows    [][]string
		columns = []string{"section", "key", "value"}
	)

	for _, section := range cfg.Sections() {
		if section.Name() == "DEFAULT" {
			continue
		}

		rows = append(rows, []string{section.Name()})
		for _, key := range section.Keys() {
			rows = append(rows, []string{"", key.Name(), key.Value()})
		}
	}

	var (
		purple    = lipgloss.Color("99")
		gray      = lipgloss.Color("245")
		lightGray = lipgloss.Color("241")

		headerStyle  = lipgloss.NewStyle().Foreground(purple).Bold(true).Align(lipgloss.Center)
		cellStyle    = lipgloss.NewStyle().Padding(0, 1)
		oddRowStyle  = cellStyle.Foreground(gray)
		evenRowStyle = cellStyle.Foreground(lightGray)
	)

	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(purple)).
		StyleFunc(func(row, col int) lipgloss.Style {
			switch {
			case row == table.HeaderRow:
				return headerStyle
			case len(rows[row]) == 1:
				return evenRowStyle
			default:
				return oddRowStyle
			}
		}).
		Headers(columns...).
		Rows(rows...)

	fmt.Println(t)
}

func prettyPrintJSON(value any) error {
	bytesOut, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(bytesOut))

	return nil
}

func prettyPrintYAML(value any) error {
	bytesOut, err := yaml.Marshal(value)
	if err != nil {
		return err
	}

	fmt.Println(string(bytesOut))

	return nil
}

func OutputObject(value map[string]any, serviceName, templateContent string, outputFormat *OutputFormat) {
	// Force JSON rendering if no template defined
	if templateContent == "" && !outputFormat.YamlOutput &&
		!outputFormat.InteractiveOutput && outputFormat.CustomFormat == "" {
		outputFormat.JsonOutput = true
	}

	switch {
	case outputFormat.CustomFormat != "":
		renderCustomFormat(value, outputFormat.CustomFormat)
		return
	case outputFormat.YamlOutput:
		if err := prettyPrintYAML(value); err != nil {
			ExitError("error displaying YAML results: %s", err)
		}
		return
	case outputFormat.JsonOutput:
		if err := prettyPrintJSON(value); err != nil {
			ExitError("error displaying JSON results: %s", err)
		}
		return
	case outputFormat.InteractiveOutput:
		bytes, err := json.Marshal(value)
		if err != nil {
			ExitError("error preparing interactive output: %s", err)
		}
		fxdisplay.Display(bytes, "")
		return
	default:
		var tpl bytes.Buffer
		t := template.Must(template.New("").Funcs(funcMap).Parse(templateContent))
		err := t.Execute(&tpl, map[string]any{
			"ServiceName": serviceName,
			"Result":      value,
		})
		if err != nil {
			ExitError("failed to execute template: %s", err)
		}

		r, err := glamour.NewTermRenderer(
			glamour.WithAutoStyle(),
			glamour.WithPreservedNewLines(),
		)
		if err != nil {
			ExitError("failed to init rendered: %s", err)
		}

		out, err := r.Render(tpl.String())
		if err != nil {
			ExitError("execution failed: %s", err)
		}
		fmt.Print(out)
	}
}

func ExitError(message string, params ...any) {
	fmt.Printf("🛑 "+message+"\n", params...)
	os.Exit(1)
}

func ExitWarning(message string, params ...any) {
	fmt.Printf("🟠 "+message+"\n", params...)
	os.Exit(0)
}
