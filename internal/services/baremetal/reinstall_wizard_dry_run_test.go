// SPDX-FileCopyrightText: 2026 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package baremetal

import (
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/maxatome/go-testdeep/td"
	"github.com/ovh/ovhcloud-cli/internal/display"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	"github.com/spf13/cobra"
)

const wizardReinstallCall = "POST https://eu.api.ovh.com/v1/dedicated/server/srv/reinstall"

// withConfirmedWizard stands in for an operator who went through the whole
// wizard and typed "yes" on its confirmation step.
func withConfirmedWizard(t *testing.T, dryRun bool) {
	t.Helper()
	withTaskAPI(t, 1, `{}`)
	httpmock.RegisterResponder("POST", "https://eu.api.ovh.com/v1/dedicated/server/srv/reinstall",
		httpmock.NewStringResponder(200, `{"taskId": 156839472}`))

	origWizard, origFlag := reinstallWizard, ReinstallWizard
	origDryRun, origWait := flags.DryRun, flags.WaitForTask
	reinstallWizard = func(string) (map[string]any, bool, string, error) {
		return map[string]any{"operatingSystem": "debian12_64"}, true, "", nil
	}
	ReinstallWizard, flags.DryRun, flags.WaitForTask = true, dryRun, false
	display.ResultError = nil

	t.Cleanup(func() {
		reinstallWizard, ReinstallWizard = origWizard, origFlag
		flags.DryRun, flags.WaitForTask = origDryRun, origWait
		display.ResultError = nil
	})
}

// The wizard posts its body itself, so it used to skip the --dry-run stop that
// CreateResource provides: `reinstall --wizard --dry-run` reinstalled the
// server as soon as the operator confirmed in the wizard.
func TestReinstallWizard_DryRunSendsNothing(t *testing.T) {
	withConfirmedWizard(t, true)

	ReinstallBaremetal(&cobra.Command{}, []string{"srv"})

	td.CmpNoError(t, display.ResultError)
	td.Cmp(t, httpmock.GetCallCountInfo()[wizardReinstallCall], 0,
		"a dry run must not reach the reinstall endpoint")
}

// The positive control: without --dry-run, the same answers do reinstall.
// Without it, a wizard branch that never posted at all would pass the test
// above.
func TestReinstallWizard_PostsWithoutDryRun(t *testing.T) {
	withConfirmedWizard(t, false)

	ReinstallBaremetal(&cobra.Command{}, []string{"srv"})

	td.CmpNoError(t, display.ResultError)
	td.Cmp(t, httpmock.GetCallCountInfo()[wizardReinstallCall], 1)
}
