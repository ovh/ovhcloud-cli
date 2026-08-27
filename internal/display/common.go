// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package display

import (
	"encoding/json"
	"fmt"
	"reflect"
)

var (
	ResultError  error
	ResultString string
)

// OutputFormat controls the output format of the CLI.
// The Output field can be "json", "yaml", "interactive", or a custom gval expression.
type OutputFormat struct {
	Output string
}

// formatCustomValue renders a value extracted by a custom format (-o '<expr>').
// Scalar values (string, number, bool) are returned without JSON quoting
// (strings as-is, numbers and booleans in their natural form) so they can be
// used directly in scripts; complex values (objects, arrays) fall back to JSON.
func formatCustomValue(out any) (string, error) {
	if out == nil {
		return "", nil
	}
	if s, ok := out.(string); ok {
		return s, nil
	}
	switch reflect.TypeOf(out).Kind() {
	case reflect.Map, reflect.Slice, reflect.Array:
		// Complex value: fall back to JSON below.
	default:
		return fmt.Sprint(out), nil
	}

	outBytes, err := json.Marshal(out)
	if err != nil {
		return "", err
	}
	return string(outBytes), nil
}

func (o *OutputFormat) IsJson() bool        { return o.Output == "json" }
func (o *OutputFormat) IsYaml() bool        { return o.Output == "yaml" }
func (o *OutputFormat) IsInteractive() bool { return o.Output == "interactive" }
func (o *OutputFormat) CustomFormat() string {
	if o.Output != "" && !o.IsJson() && !o.IsYaml() && !o.IsInteractive() {
		return o.Output
	}
	return ""
}

type OutputMessage struct {
	Message string `json:"message,omitempty"`
	Error   bool   `json:"error,omitempty"`
	Warning bool   `json:"warning,omitempty"`
	Details any    `json:"details,omitempty"`
}
