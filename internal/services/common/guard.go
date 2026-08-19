// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"fmt"

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
	// Destructive loses data, or the resource itself. There is no retry.
	//
	// It is deliberately the zero value: a caller that forgets to state a
	// severity gets the strictest guard rather than the weakest one, and finds
	// out immediately instead of on the day it mattered.
	Destructive Severity = iota

	// Disruptive interrupts a running service. Nothing is lost, but somebody
	// notices.
	Disruptive
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
	case Disruptive:
		return utils.ConfirmYesNo(warning)
	default:
		return utils.ConfirmByName(resource, warning)
	}
}

// Call is one request a command would send.
type Call struct {
	Method   string
	Endpoint string

	// Detail says what the request carries, when the path alone does not.
	//
	// It is a separate field because Endpoint is read by machines as well as
	// people: callers used to append prose to the path, which made --dry-run
	// -o json report "/dedicated/server/x  (bootId of the rescue entry)" as an
	// endpoint. A string that is not a path has no business in a field named
	// after one.
	Detail string
}

// ReportDryRun prints every call a command would have made, and reports true so
// the caller can stop on the same line it tested the flag.
//
// It takes the whole sequence rather than one call because several of these
// commands are not one request. `reboot-rescue` reads the rescue boot, writes
// it to the server and only then reboots — and the write is the part that
// outlives the reboot. A preview that showed the reboot alone would hide
// exactly the change an operator needs to see before agreeing to it.
//
// Endpoints are shown in full: these commands have no other preview, and the
// point is to let somebody read the paths before they are acted on.
func ReportDryRun(calls ...Call) bool {
	if !flags.DryRun {
		return false
	}

	details := make([]map[string]any, 0, len(calls))
	message := "🔍 Dry run: nothing was sent. This would have been called:"
	for _, c := range calls {
		detail := map[string]any{"method": c.Method, "endpoint": c.Endpoint}
		line := fmt.Sprintf("\n  %s %s", c.Method, c.Endpoint)
		if c.Detail != "" {
			detail["detail"] = c.Detail
			line += "  (" + c.Detail + ")"
		}
		details = append(details, detail)
		message += line
	}

	display.OutputInfo(&flags.OutputFormatConfig, map[string]any{"calls": details}, "%s", message)

	return true
}
