// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"github.com/ovh/ovhcloud-cli/internal/display"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	"github.com/ovh/ovhcloud-cli/internal/utils"
)

// Severity says how much an action costs when it was not the one intended.
// The two levels exist because proportionality is what makes a guardrail
// survive daily use: asking an operator to type a server name before every
// reboot teaches them to reach for --yes, and a habit of --yes protects
// nothing on the day it matters.
type Severity int

const (
	// Disruptive interrupts a running service. Nothing is lost, but somebody
	// notices.
	Disruptive Severity = iota

	// Destructive loses data, or the resource itself. There is no retry.
	Destructive
)

// ConfirmAction gates an action behind a confirmation and reports whether the
// caller may proceed.
//
// It answers true without asking anything in two cases: --yes, which is how a
// pipeline states its intent, and --dry-run, where the caller is about to
// print the request instead of sending it. A dry run that still demanded a
// confirmation would refuse to describe what it will not do.
//
// When it answers false it has already told the operator why, so the caller
// only has to stop.
func ConfirmAction(severity Severity, resource, warning string) bool {
	if flags.AssumeYes || flags.DryRun {
		return true
	}

	switch severity {
	case Destructive:
		return utils.ConfirmByName(resource, warning)
	default:
		return utils.ConfirmYesNo(warning)
	}
}

// ReportDryRun prints the call a command would have made, and reports true so
// the caller can stop on the same line it tested the flag.
//
// The endpoint is shown in full because these commands have no other preview:
// the whole point of --dry-run here is to let somebody read the path before it
// is acted on.
func ReportDryRun(method, endpoint string) bool {
	if !flags.DryRun {
		return false
	}

	display.OutputInfo(&flags.OutputFormatConfig,
		map[string]any{"method": method, "endpoint": endpoint},
		"🔍 Dry run: nothing was sent. This would have been called:\n  %s %s", method, endpoint)

	return true
}
