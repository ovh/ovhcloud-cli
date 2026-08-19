// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package display

var (
	ResultError  error
	ResultString string
)

// OutputFormat controls the output format of the CLI.
// The Output field can be "json", "yaml", "interactive", or a custom gval expression.
type OutputFormat struct {
	Output string
}

func (o *OutputFormat) IsJson() bool        { return o.Output == "json" }
func (o *OutputFormat) IsYaml() bool        { return o.Output == "yaml" }
func (o *OutputFormat) IsInteractive() bool { return o.Output == "interactive" }

// IsPlain reports whether the caller asked for borderless, uncoloured
// columns, the shape text tools such as awk and cut can consume.
func (o *OutputFormat) IsPlain() bool { return o.Output == "plain" }

func (o *OutputFormat) CustomFormat() string {
	if o.Output != "" && !o.IsJson() && !o.IsYaml() && !o.IsInteractive() && !o.IsPlain() {
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
