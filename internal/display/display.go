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
	"io"
	"os"
	"reflect"
	"strings"
	"text/template"

	"github.com/PaesslerAG/gval"
	fxdisplay "github.com/amstuta/fx/display"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/term"
	"github.com/ghodss/yaml"
	"github.com/ovh/ovhcloud-cli/internal/filters"
	"gopkg.in/ini.v1"
)

const (
	maxCellWidth = 50
)

var (
	ExitFunc = os.Exit
)

// writeTo writes command data to the given stream and records it so that
// Execute can return it to the caller.
func writeTo(w io.Writer, message string, params ...any) {
	ResultString = fmt.Sprintf(message, params...)
	fmt.Fprintln(w, ResultString)
}

// errorf writes an error or a warning to stderr. Diagnostics must never be
// mixed with command data: a redirected stdout has to stay parsable.
func errorf(message string, params ...any) {
	fmt.Fprintln(os.Stderr, fmt.Sprintf(message, params...))
}

// hintf writes a non-essential hint to stderr, and only when stdout is a
// terminal: a hint is useless to a script and pollutes its input.
func hintf(message string, params ...any) {
	if !term.IsTerminal(os.Stdout.Fd()) {
		return
	}
	fmt.Fprintln(os.Stderr, fmt.Sprintf(message, params...))
}

func renderCustomFormat(value any, format string, w io.Writer) error {
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
		fmt.Fprint(w, ResultString)
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
		fmt.Fprint(w, ResultString)
	}

	return nil
}

func RenderTable(values []map[string]any, columnsToDisplay []string, outputFormat *OutputFormat) {
	switch {
	case outputFormat.CustomFormat() != "":
		if err := renderCustomFormat(values, outputFormat.CustomFormat(), os.Stdout); err != nil {
			exitError("error rendering custom format: %s", err)
		}
		return
	case outputFormat.IsInteractive():
		displayInteractive(values)
		return
	case outputFormat.IsYaml():
		if err := prettyPrintYAML(values, os.Stdout); err != nil {
			exitError("error displaying YAML results: %s", err)
		}
		return
	case outputFormat.IsJson():
		if err := prettyPrintJSON(values, os.Stdout); err != nil {
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

			if val == nil {
				row = append(row, "")
				continue
			}

			var cellValue string
			switch val.(type) {
			case float32, float64:
				cellValue = fmt.Sprintf("%.0f", val)
			default:
				cellValue = fmt.Sprintf("%v", val)
			}

			// Truncate content if it exceeds max width
			cellValue = ansi.Truncate(cellValue, maxCellWidth, "…")
			row = append(row, cellValue)
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

	outputf("%s", t)
	hintf("💡 Use option -o json or -o yaml to get the raw output with all information")
}

func RenderConfigTable(cfg *ini.File, outputformat *OutputFormat) {
	// if a custom output format is passed, it is used instead of rendering the Config Table
	if outputformat.IsJson() || outputformat.IsYaml() || outputformat.IsInteractive() || outputformat.CustomFormat() != "" {
		result := map[string]any{}
		for _, section := range cfg.Sections() {
			if section.Name() == "DEFAULT" {
				continue
			}
			sectionMap := map[string]any{}
			for _, key := range section.Keys() {
				sectionMap[key.Name()] = key.Value()
			}
			result[section.Name()] = sectionMap
		}
		OutputObject(result, "", "", outputformat)
		return
	}
	// otherwise, render the config as a table
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
			value := ansi.Truncate(key.Value(), maxCellWidth, "…")
			rows = append(rows, []string{"", key.Name(), value})
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

func prettyPrintJSON(value any, w io.Writer) error {
	bytesOut, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}

	writeTo(w, "%s", bytesOut)

	return nil
}

func prettyPrintYAML(value any, w io.Writer) error {
	bytesOut, err := yaml.Marshal(value)
	if err != nil {
		return err
	}

	writeTo(w, "%s", bytesOut)

	return nil
}

func OutputObject(value map[string]any, serviceName, templateContent string, outputFormat *OutputFormat) {
	// Force JSON rendering if no template defined
	if templateContent == "" && !outputFormat.IsYaml() &&
		!outputFormat.IsInteractive() && outputFormat.CustomFormat() == "" {
		outputFormat.Output = "json"
	}

	switch {
	case outputFormat.CustomFormat() != "":
		if err := renderCustomFormat(value, outputFormat.CustomFormat(), os.Stdout); err != nil {
			exitError("error rendering custom format: %s", err)
		}
		return
	case outputFormat.IsYaml():
		if err := prettyPrintYAML(value, os.Stdout); err != nil {
			exitError("error displaying YAML results: %s", err)
		}
		return
	case outputFormat.IsJson():
		if err := prettyPrintJSON(value, os.Stdout); err != nil {
			exitError("error displaying JSON results: %s", err)
		}
		return
	case outputFormat.IsInteractive():
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

		// Define word wrap for the renderer.
		// Use 80 characters by default, or the terminal width if available.
		wordWrap := 80
		if termFd := os.Stdout.Fd(); term.IsTerminal(termFd) {
			if termWidth, _, _ := term.GetSize(termFd); termWidth > 0 {
				wordWrap = termWidth
			}
		}

		r, err := glamour.NewTermRenderer(
			glamour.WithAutoStyle(),
			glamour.WithPreservedNewLines(),
			glamour.WithWordWrap(wordWrap),
		)
		if err != nil {
			exitError("failed to init rendered: %s", err)
		}

		out, err := r.Render(tpl.String())
		if err != nil {
			exitError("execution failed: %s", err)
		}
		fmt.Fprint(os.Stdout, out)
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
	errorf("%s", resultString)
	ResultError = errors.New(resultString)
	ExitFunc(1)
}

func outputf(message string, params ...any) {
	writeTo(os.Stdout, "%s", fmt.Sprintf(message, params...))
}

func OutputWithFormat(msg *OutputMessage, outputFormat *OutputFormat) {
	// Errors and warnings go to stderr whatever the output format, so that a
	// redirected stdout only ever holds command data: `cmd -o json > out.json`
	// must leave out.json empty when the call failed.
	dst := io.Writer(os.Stdout)
	if msg.Error || msg.Warning {
		dst = os.Stderr
	}

	switch {
	case outputFormat.CustomFormat() != "":
		data, err := json.Marshal(msg)
		if err != nil {
			exitError("error marshalling message: %s", err)
		}
		var m map[string]any
		if err := json.Unmarshal(data, &m); err != nil {
			exitError("error unmarshalling message: %s", err)
		}

		if err := renderCustomFormat(m, outputFormat.CustomFormat(), dst); err != nil {
			exitError("error rendering custom format: %s", err)
		}

	case outputFormat.IsYaml():
		if err := prettyPrintYAML(msg, dst); err != nil {
			exitError("error displaying YAML results: %s", err)
		}

	case outputFormat.IsJson():
		if err := prettyPrintJSON(msg, dst); err != nil {
			exitError(err.Error())
		}

	case outputFormat.IsInteractive():
		// A diagnostic is not data to browse. The interactive viewer starts a
		// full-screen program on stdout, which would put the message back on
		// the stream this function exists to keep clean.
		if msg.Error || msg.Warning {
			writeTo(dst, "%s", msg.Message)
		} else {
			displayInteractive(msg)
		}

	default:
		writeTo(dst, "%s", msg.Message)
	}

	if msg.Error {
		ResultError = errors.New(msg.Message)
		ExitFunc(1)
	} else if msg.Warning {
		ResultError = errors.New(msg.Message)
		ExitFunc(0)
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
