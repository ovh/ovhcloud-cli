// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

//go:build !(js && wasm)

package display

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
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
	"github.com/ovh/ovhcloud-cli/internal/filters"
	"gopkg.in/ini.v1"
)

func renderCustomFormat(value any, format string) error {
	ev, err := gval.Full(filters.AdditionalEvaluators...).NewEvaluable(format)
	if err != nil {
		return fmt.Errorf("invalid format given: %w", err)
	}

	switch reflect.TypeOf(value).Kind() {
	case reflect.Slice:
		var output strings.Builder
		for _, val := range value.([]map[string]any) {
			out, err := ev(context.Background(), val)
			if err != nil {
				return fmt.Errorf("couldn't extract data according to given format: %w", err)
			}

			outBytes, err := json.Marshal(out)
			if err != nil {
				return fmt.Errorf("error marshalling result: %w", err)
			}
			output.Write(outBytes)
			output.WriteString("\n")
		}
		ResultString = output.String()
		fmt.Print(ResultString)
	default:
		out, err := ev(context.Background(), value)
		if err != nil {
			return fmt.Errorf("couldn't extract data according to given format: %w", err)
		}

		outBytes, err := json.Marshal(out)
		if err != nil {
			return fmt.Errorf("error marshalling result: %w", err)
		}
		ResultString = string(outBytes)
		fmt.Print(string(outBytes))
	}

	return nil
}

func RenderTable(values []map[string]any, columnsToDisplay []string, outputFormat *OutputFormat) {
	switch {
	case outputFormat.CustomFormat != "":
		if err := renderCustomFormat(values, outputFormat.CustomFormat); err != nil {
			exitError("error rendering custom format: %s", err)
		}
		return
	case outputFormat.InteractiveOutput:
		displayInteractive(values)
		return
	case outputFormat.YamlOutput:
		if err := prettyPrintYAML(values); err != nil {
			exitError("error displaying YAML results: %s", err)
		}
		return
	case outputFormat.JsonOutput:
		if err := prettyPrintJSON(values); err != nil {
			exitError("error displaying JSON results: %s", err)
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
		colName, alias, ok := strings.Cut(col, " ")
		if ok {
			col = colName
			columnsTitles = append(columnsTitles, alias)
		} else {
			columnsTitles = append(columnsTitles, col)
		}

		// Create selector to extract value at given key
		evaluator, err := gval.Base().NewEvaluable(col)
		if err != nil {
			exitError("invalid column to display %q: %s", col, err)
		}
		selectors = append(selectors, evaluator)
	}

	// Extract values to display
	for _, value := range values {
		var row []string
		for _, selector := range selectors {
			val, err := selector(context.Background(), value)
			if err != nil {
				exitError("failed to select row field: %s", err)
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

	outputf("%s%s", t, "\n💡 Use option --json or --yaml to get the raw output with all information")
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

	outputf("%s", t)
}

func prettyPrintJSON(value any) error {
	bytesOut, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}

	outputf("%s", bytesOut)

	return nil
}

func prettyPrintYAML(value any) error {
	bytesOut, err := yaml.Marshal(value)
	if err != nil {
		return err
	}

	outputf("%s", bytesOut)

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
		if err := renderCustomFormat(value, outputFormat.CustomFormat); err != nil {
			exitError("error rendering custom format: %s", err)
		}
		return
	case outputFormat.YamlOutput:
		if err := prettyPrintYAML(value); err != nil {
			exitError("error displaying YAML results: %s", err)
		}
		return
	case outputFormat.JsonOutput:
		if err := prettyPrintJSON(value); err != nil {
			exitError("error displaying JSON results: %s", err)
		}
		return
	case outputFormat.InteractiveOutput:
		displayInteractive(value)
		return
	default:
		var tpl bytes.Buffer
		t := template.Must(template.New("").Funcs(funcMap).Parse(templateContent))
		err := t.Execute(&tpl, map[string]any{
			"ServiceName": serviceName,
			"Result":      value,
		})
		if err != nil {
			exitError("failed to execute template: %s", err)
		}

		r, err := glamour.NewTermRenderer(
			glamour.WithAutoStyle(),
			glamour.WithPreservedNewLines(),
		)
		if err != nil {
			exitError("failed to init rendered: %s", err)
		}

		out, err := r.Render(tpl.String())
		if err != nil {
			exitError("execution failed: %s", err)
		}
		fmt.Print(out)
		ResultString = out
	}
}

func displayInteractive(value any) {
	bytes, err := json.Marshal(value)
	if err != nil {
		exitError("error preparing interactive output: %s", err)
	}
	fxdisplay.Display(bytes, "")
}

func exitError(message string, params ...any) {
	resultString := fmt.Sprintf("🛑 "+message, params...)
	fmt.Println(resultString)
	ResultError = errors.New(resultString)
	os.Exit(1)
}

func outputf(message string, params ...any) {
	ResultString = fmt.Sprintf(message, params...)
	fmt.Println(ResultString)
}

func OutputWithFormat(msg *OutputMessage, outputFormat *OutputFormat) {
	switch {
	case outputFormat.CustomFormat != "":
		data, err := json.Marshal(msg)
		if err != nil {
			exitError("error marshalling message: %s", err)
		}
		var m map[string]any
		if err := json.Unmarshal(data, &m); err != nil {
			exitError("error unmarshalling message: %s", err)
		}

		if err := renderCustomFormat(m, outputFormat.CustomFormat); err != nil {
			exitError("error rendering custom format: %s", err)
		}

	case outputFormat.YamlOutput:
		if err := prettyPrintYAML(msg); err != nil {
			exitError("error displaying YAML results: %s", err)
		}

	case outputFormat.JsonOutput:
		if err := prettyPrintJSON(msg); err != nil {
			exitError(err.Error())
		}

	case outputFormat.InteractiveOutput:
		displayInteractive(msg)

	default:
		outputf("%s", msg.Message)
	}

	if msg.Error {
		ResultError = errors.New(msg.Message)
		os.Exit(1)
	} else if msg.Warning {
		ResultError = errors.New(msg.Message)
		os.Exit(0)
	}
}

func OutputInfo(outputFormat *OutputFormat, details any, message string, params ...any) {
	OutputWithFormat(&OutputMessage{
		Message: fmt.Sprintf(message, params...),
		Details: details,
	}, outputFormat)
}

func OutputError(outputFormat *OutputFormat, message string, params ...any) {
	resultString := fmt.Sprintf("🛑 "+message, params...)
	OutputWithFormat(&OutputMessage{
		Message: resultString,
		Error:   true,
	}, outputFormat)
}

func OutputWarning(outputFormat *OutputFormat, message string, params ...any) {
	resultString := fmt.Sprintf("🟠 "+message, params...)
	OutputWithFormat(&OutputMessage{
		Message: resultString,
		Warning: true,
	}, outputFormat)
}
