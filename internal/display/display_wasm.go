//go:build js && wasm

package display

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/PaesslerAG/gval"
	"github.com/ovh/ovhcloud-cli/internal/filters"
	"gopkg.in/ini.v1"
)

func renderCustomFormat(value any, format string) {
	ev, err := gval.Full(filters.AdditionalEvaluators...).NewEvaluable(format)
	if err != nil {
		ExitError("invalid format given: %s", err)
		return
	}

	switch reflect.TypeOf(value).Kind() {
	case reflect.Slice:
		for _, val := range value.([]map[string]any) {
			out, err := ev(context.Background(), val)
			if err != nil {
				ExitError("couldn't extract data according to given format: %s", err)
				return
			}

			outBytes, err := json.Marshal(out)
			if err != nil {
				ExitError("error marshalling result")
				return
			}
			ResultString = string(outBytes)
		}
	default:
		out, err := ev(context.Background(), value)
		if err != nil {
			ExitError("couldn't extract data according to given format: %s", err)
			return
		}

		outBytes, err := json.Marshal(out)
		if err != nil {
			ExitError("error marshalling result")
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

	if err := prettyPrintJSON(values); err != nil {
		ExitError("error displaying JSON results: %s", err)
		return
	}
}

func RenderConfigTable(cfg *ini.File) {
	// TODO: untested
	output := map[string]any{}
	if err := cfg.MapTo(&output); err != nil {
		ExitError("failed to extract config to map: %s", err)
		return
	}

	if err := prettyPrintJSON(output); err != nil {
		ExitError("error displaying JSON results: %s", err)
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
		ExitError("error displaying JSON results: %s", err)
		return
	}
}

func ExitError(message string, params ...any) {
	ResultError = fmt.Errorf("🛑 "+message+"\n", params...)
}

func ExitWarning(message string, params ...any) {
	ResultError = fmt.Errorf("🟠 "+message+"\n", params...)
}
