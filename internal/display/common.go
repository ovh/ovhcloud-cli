// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package display

var (
	ResultError  error
	ResultString string
)

// // Common flags used by all subcommands to control output format (json, yaml)
type OutputFormat struct {
	JsonOutput, YamlOutput, InteractiveOutput bool
	CustomFormat                              string
}
