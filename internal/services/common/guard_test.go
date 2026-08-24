// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"testing"

	"github.com/maxatome/go-testdeep/td"
	"github.com/ovh/ovhcloud-cli/internal/flags"
)

// A caller that forgets to state a severity must get the strictest guard.
// Written as an assertion on the constant rather than on behaviour, because
// the failure this prevents is somebody reordering the block: the zero value
// would quietly become the weak confirmation and every call site that relied
// on the default would be downgraded without a single test noticing.
func TestSeverity_ZeroValueIsTheStrictestGuard(t *testing.T) {
	var unset Severity

	td.Cmp(t, unset, Destructive,
		"the zero value must be Destructive: an omitted severity is a bug, and it must fail safe")
}

// --yes is how a pipeline states its intent, for either severity.
func TestConfirmAction_AssumeYesSkipsBothLevels(t *testing.T) {
	orig := flags.AssumeYes
	defer func() { flags.AssumeYes = orig }()
	flags.AssumeYes = true

	td.Cmp(t, ConfirmAction(Destructive, "srv", "wipes the disks"), true)
	td.Cmp(t, ConfirmAction(Disruptive, "srv", "interrupts the service"), true)
}

// A dry run prints instead of acting, so it must never be gated by a
// confirmation: refusing to describe a call it will not make is absurd.
func TestConfirmAction_DryRunNeverPrompts(t *testing.T) {
	orig := flags.DryRun
	defer func() { flags.DryRun = orig }()
	flags.DryRun = true

	td.Cmp(t, ConfirmAction(Destructive, "srv", "wipes the disks"), true)
}

// ReportDryRun is inert unless --dry-run was given: it must not print over the
// output of a command that is really running.
func TestReportDryRun_IsInertWithoutTheFlag(t *testing.T) {
	orig := flags.DryRun
	defer func() { flags.DryRun = orig }()
	flags.DryRun = false

	td.Cmp(t, ReportDryRun(Call{Method: "POST", Endpoint: "/v1/whatever"}), false)
}
