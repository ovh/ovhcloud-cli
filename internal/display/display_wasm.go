// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

//go:build js && wasm

package display

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/PaesslerAG/gval"
	"github.com/ovh/ovhcloud-cli/internal/filters"
	"gopkg.in/ini.v1"
)

func renderCustomFormat(value any, format string) {
	ev, err := gval.Full(filters.AdditionalEvaluators...).NewEvaluable(format)
	if err != nil {
		exitError("invalid format given: %s", err)
		return
	}

	switch reflect.TypeOf(value).Kind() {
	case reflect.Slice:
		for _, val := range value.([]map[string]any) {
			out, err := ev(context.Background(), val)
			if err != nil {
				exitError("couldn't extract data according to given format: %s", err)
				return
			}

			outBytes, err := json.Marshal(out)
			if err != nil {
				exitError("error marshalling result")
				return
			}
			ResultString = string(outBytes)
		}
	default:
		out, err := ev(context.Background(), value)
		if err != nil {
			exitError("couldn't extract data according to given format: %s", err)
			return
		}

		outBytes, err := json.Marshal(out)
		if err != nil {
			exitError("error marshalling result")
			return
		}
		ResultString = string(outBytes)
	}
}

func RenderTable(values []map[string]any, columnsToDisplay []string, outputFormat *OutputFormat) {
	if outputFormat.CustomFormat != "" {
		renderCustomFormat(values, outputFormat.CustomFormat)
		return
	}

	// If not JSON or YAML output, extract only the columns to display
	// as the values will be displayed in a table
	if !outputFormat.JsonOutput && !outputFormat.YamlOutput {
		var (
			selectors       gval.Evaluables
			extractedValues = make([]map[string]any, 0, len(values))
			columnsTitles   = make([]string, 0, len(columnsToDisplay))
		)

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
			row := make(map[string]any)
			for i, selector := range selectors {
				val, err := selector(context.Background(), value)
				if err != nil {
					exitError("failed to select row field: %s", err)
				}
				row[columnsTitles[i]] = val
			}
			extractedValues = append(extractedValues, row)
		}

		values = extractedValues
	}

	if err := prettyPrintJSON(values); err != nil {
		exitError("error displaying JSON results: %s", err)
		return
	}
}

func RenderConfigTable(cfg *ini.File) {
	// TODO: untested
	output := map[string]any{}
	if err := cfg.MapTo(&output); err != nil {
		exitError("failed to extract config to map: %s", err)
		return
	}

	if err := prettyPrintJSON(output); err != nil {
		exitError("error displaying JSON results: %s", err)
		return
	}
}

func prettyPrintJSON(value any) error {
	bytesOut, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}

	ResultString = string(bytesOut)

	return nil
}

func OutputObject(value map[string]any, serviceName, templateContent string, outputFormat *OutputFormat) {
	if outputFormat.CustomFormat != "" {
		renderCustomFormat(value, outputFormat.CustomFormat)
		return
	}

	if err := prettyPrintJSON(value); err != nil {
		exitError("error displaying JSON results: %s", err)
		return
	}
}

func exitError(message string, params ...any) {
	ResultError = fmt.Errorf("🛑 "+message+"\n", params...)
}

func outputf(message string, params ...any) {
	valueOut := &OutputMessage{
		Message: fmt.Sprintf(message, params...),
	}

	if err := prettyPrintJSON(valueOut); err != nil {
		exitError("error displaying JSON results: %s", err)
		return
	}
}

func OutputInfo(outputFormat *OutputFormat, details any, message string, params ...any) {
	outputf(message, params...)
}

func OutputError(outputFormat *OutputFormat, message string, params ...any) {
	exitError(message, params...)
}

func OutputWarning(outputFormat *OutputFormat, message string, params ...any) {
	exitError(message, params...)
}
