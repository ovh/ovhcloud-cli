// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"github.com/jarcoal/httpmock"
	"github.com/maxatome/go-testdeep/td"
	"github.com/ovh/ovhcloud-cli/internal/cmd"
)

func (ms *MockSuite) TestBaremetalIPMIResetSessionsCmd(assert, require *td.T) {
	httpmock.RegisterResponder("POST", "https://eu.api.ovh.com/v1/dedicated/server/fakeBaremetal/features/ipmi/resetSessions",
		httpmock.NewStringResponder(200, `{}`),
	)

	out, err := cmd.Execute("baremetal", "ipmi", "reset-sessions", "fakeBaremetal")

	require.CmpNoError(err)
	assert.Contains(out, "IPMI sessions reset")
}

func (ms *MockSuite) TestBaremetalListCompatibleOSCmd(assert, require *td.T) {
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/dedicated/server/fakeBaremetal/install/compatibleTemplates",
		httpmock.NewStringResponder(200, `{
			"ovh": [
				"alma8-cpanel-latest_64",
				"alma8-plesk18_64",
				"alma8_64",
				"alma9-cpanel-latest_64",
				"alma9-plesk18_64",
				"alma9_64",
				"byoi_64",
				"byolinux_64"
			]
		}`),
	)

	out, err := cmd.Execute("baremetal", "list-compatible-os", "fakeBaremetal")

	require.CmpNoError(err)
	assert.String(out, `
┌────────┬────────────────────────┐
│ source │          name          │
├────────┼────────────────────────┤
│ ovh    │ alma8-cpanel-latest_64 │
│ ovh    │ alma8-plesk18_64       │
│ ovh    │ alma8_64               │
│ ovh    │ alma9-cpanel-latest_64 │
│ ovh    │ alma9-plesk18_64       │
│ ovh    │ alma9_64               │
│ ovh    │ byoi_64                │
│ ovh    │ byolinux_64            │
└────────┴────────────────────────┘
💡 Use option -o json or -o yaml to get the raw output with all information`[1:])
}

// registerReinstallTask wires a reinstall whose task ends in the given state.
//
// These tests pass --yes: they are about what the CLI says when a task fails,
// not about the confirmation, and without it the reinstall never starts.
func registerReinstallTask(task string) {
	httpmock.RegisterResponder("POST", "https://eu.api.ovh.com/v1/dedicated/server/fakeBaremetal/reinstall",
		httpmock.NewStringResponder(200, `{"taskId": 156839472}`),
	)
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/dedicated/server/fakeBaremetal/task/156839472",
		httpmock.NewStringResponder(200, task),
	)
}

// A failed task must say what failed and why. "invalid state" told the
// operator the CLI had not understood, when their reinstallation had failed.
func (ms *MockSuite) TestBaremetalReinstallReportsTheFailureReason(assert, require *td.T) {
	registerReinstallTask(`{"taskId": 156839472, "function": "reinstallServer",
		"status": "customerError", "comment": "partitioning scheme incompatible with this hardware"}`)

	_, err := cmd.Execute("baremetal", "reinstall", "fakeBaremetal", "--os", "debian12_64", "--wait", "--yes")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("customerError"), "the real status is named")
	assert.Cmp(err.Error(), td.Contains("partitioning scheme incompatible"), "and the reason the API gave")
	assert.Cmp(err.Error(), td.Contains("reinstallServer"), "along with the operation that failed")
	assert.Cmp(err.Error(), td.Not(td.Contains("invalid state")))
}

// The task identifier is decoded as a json.Number, and %d used to render it as
// %!d(json.Number=…) — in the one message written to explain a failure.
func (ms *MockSuite) TestBaremetalReinstallPrintsAReadableTaskID(assert, require *td.T) {
	registerReinstallTask(`{"taskId": 156839472, "function": "reinstallServer", "status": "ovhError"}`)

	_, err := cmd.Execute("baremetal", "reinstall", "fakeBaremetal", "--os", "debian12_64", "--wait", "--yes")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("156839472"))
	assert.Cmp(err.Error(), td.Not(td.Contains("%!d")), "the identifier must be readable")
	assert.Cmp(err.Error(), td.Contains("list-tasks fakeBaremetal"), "and the message says how to follow up")
}

// A status this CLI does not know is not a failure of the task: the API enum
// can grow, and guessing would report a success or a failure that never was.
func (ms *MockSuite) TestBaremetalReinstallReportsAnUnknownStatus(assert, require *td.T) {
	registerReinstallTask(`{"taskId": 156839472, "function": "reinstallServer", "status": "quarantined"}`)

	_, err := cmd.Execute("baremetal", "reinstall", "fakeBaremetal", "--os", "debian12_64", "--wait", "--yes")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("unexpected status quarantined"))
}

// A task that completes normally must still report success.
func (ms *MockSuite) TestBaremetalReinstallSucceedsOnDoneTask(assert, require *td.T) {
	registerReinstallTask(`{"taskId": 156839472, "function": "reinstallServer", "status": "done"}`)
	httpmock.RegisterResponder("POST", "https://eu.api.ovh.com/v1/dedicated/server/fakeBaremetal/authenticationSecret",
		httpmock.NewStringResponder(200, `[]`),
	)

	_, err := cmd.Execute("baremetal", "reinstall", "fakeBaremetal", "--os", "debian12_64", "--wait", "--yes")

	require.CmpNoError(err)
}

// Reinstalling wipes the disks: an unattended run must not do it silently.
func (ms *MockSuite) TestBaremetalReinstallRefusesWithoutConfirmation(assert, require *td.T) {
	httpmock.RegisterResponder("POST", "https://eu.api.ovh.com/v1/dedicated/server/fakeBaremetal/reinstall",
		httpmock.NewStringResponder(200, `{"taskId": 123}`),
	)

	_, err := cmd.Execute("baremetal", "reinstall", "fakeBaremetal", "--os", "debian12_64")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("cancelled"))
	assert.Cmp(httpmock.GetCallCountInfo()["POST https://eu.api.ovh.com/v1/dedicated/server/fakeBaremetal/reinstall"], 0,
		"no reinstall call must reach the API")
}

func (ms *MockSuite) TestBaremetalReinstallProceedsWithYes(assert, require *td.T) {
	httpmock.RegisterResponder("POST", "https://eu.api.ovh.com/v1/dedicated/server/fakeBaremetal/reinstall",
		httpmock.NewStringResponder(200, `{"taskId": 123}`),
	)

	out, err := cmd.Execute("baremetal", "reinstall", "fakeBaremetal", "--os", "debian12_64", "--yes")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("Reinstallation is started"))
	assert.Cmp(httpmock.GetCallCountInfo()["POST https://eu.api.ovh.com/v1/dedicated/server/fakeBaremetal/reinstall"], 1)
}

// --dry-run shows what would be sent and sends nothing.
func (ms *MockSuite) TestBaremetalReinstallDryRun(assert, require *td.T) {
	httpmock.RegisterResponder("POST", "https://eu.api.ovh.com/v1/dedicated/server/fakeBaremetal/reinstall",
		httpmock.NewStringResponder(200, `{"taskId": 123}`),
	)

	out, err := cmd.Execute("baremetal", "reinstall", "fakeBaremetal", "--os", "debian12_64", "--dry-run")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("Dry run"))
	// The parameters must be in the message itself: the default output prints
	// the message alone, so a payload living only in the structured details
	// would leave stdout with a promise and no request.
	assert.Cmp(out, td.Contains(`"operatingSystem": "debian12_64"`),
		"the request that would be sent is printed")
	assert.Cmp(out, td.Contains("/v1/dedicated/server/fakeBaremetal/reinstall"),
		"and the endpoint it would go to")
	assert.Cmp(httpmock.GetCallCountInfo()["POST https://eu.api.ovh.com/v1/dedicated/server/fakeBaremetal/reinstall"], 0,
		"no reinstall call must reach the API")
}

// The point of the guardrail: an unattended run that did not say --yes must not
// interrupt a production server, and must not reach the API at all.
func (ms *MockSuite) TestBaremetalRebootRefusesWithoutConsent(assert, require *td.T) {
	httpmock.RegisterResponder("POST", "https://eu.api.ovh.com/v1/dedicated/server/fakeBaremetal/reboot",
		httpmock.NewStringResponder(200, `{}`),
	)

	_, err := cmd.Execute("baremetal", "reboot", "fakeBaremetal")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("cancelled"))
	assert.Cmp(httpmock.GetTotalCallCount(), 0, "nothing must reach the API without a confirmation")
}

// --yes is how a pipeline states its intent, and it must be enough.
func (ms *MockSuite) TestBaremetalRebootProceedsWithYes(assert, require *td.T) {
	httpmock.RegisterResponder("POST", "https://eu.api.ovh.com/v1/dedicated/server/fakeBaremetal/reboot",
		httpmock.NewStringResponder(200, `{}`),
	)

	_, err := cmd.Execute("baremetal", "reboot", "fakeBaremetal", "--yes")

	require.CmpNoError(err)
	assert.Cmp(httpmock.GetCallCountInfo()["POST https://eu.api.ovh.com/v1/dedicated/server/fakeBaremetal/reboot"], 1)
}

// --dry-run shows the call and makes none, so it must never need a confirmation
// of its own: refusing to describe what it will not do would be absurd.
func (ms *MockSuite) TestBaremetalRebootDryRunSendsNothing(assert, require *td.T) {
	httpmock.RegisterResponder("POST", "https://eu.api.ovh.com/v1/dedicated/server/fakeBaremetal/reboot",
		httpmock.NewStringResponder(200, `{}`),
	)

	out, err := cmd.Execute("baremetal", "reboot", "fakeBaremetal", "--dry-run")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("POST /v1/dedicated/server/fakeBaremetal/reboot"))
	assert.Cmp(httpmock.GetTotalCallCount(), 0, "a dry run must send nothing")
}

// reboot-rescue is three calls, and the middle one — writing the rescue boot to
// the server — is the change that outlives the reboot. A preview showing the
// reboot alone would hide exactly what the operator needs to weigh.
func (ms *MockSuite) TestBaremetalRebootRescueDryRunShowsTheBootChange(assert, require *td.T) {
	out, err := cmd.Execute("baremetal", "reboot-rescue", "fakeBaremetal", "--dry-run")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("GET /v1/dedicated/server/fakeBaremetal/boot?bootType=rescue"))
	assert.Cmp(out, td.Contains("PUT /v1/dedicated/server/fakeBaremetal"), "the boot write must be announced")
	assert.Cmp(out, td.Contains("POST /v1/dedicated/server/fakeBaremetal/reboot"))
	assert.Cmp(httpmock.GetTotalCallCount(), 0, "a dry run must send nothing")
}

// One call per interface: the preview must list them rather than imply a single
// request.
func (ms *MockSuite) TestBaremetalOlaResetDryRunListsEveryInterface(assert, require *td.T) {
	out, err := cmd.Execute("baremetal", "vni", "ola-reset", "fakeBaremetal",
		"--interface", "uuid-1", "--interface", "uuid-2", "--dry-run")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("uuid-1"))
	assert.Cmp(out, td.Contains("uuid-2"))
	assert.Cmp(httpmock.GetTotalCallCount(), 0, "a dry run must send nothing")
}

// --yes belongs to the command it was typed on. If it survived into the next
// command of the same process — the WASM build runs several — a confirmation
// would be skipped that nobody granted.
func (ms *MockSuite) TestBaremetalYesDoesNotSurviveIntoTheNextCommand(assert, require *td.T) {
	httpmock.RegisterResponder("POST", "https://eu.api.ovh.com/v1/dedicated/server/fakeBaremetal/reboot",
		httpmock.NewStringResponder(200, `{}`),
	)

	_, err := cmd.Execute("baremetal", "reboot", "fakeBaremetal", "--yes")
	require.CmpNoError(err)

	cmd.PostExecute()

	_, err = cmd.Execute("baremetal", "reboot", "fakeBaremetal")
	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("cancelled"), "the second command must ask again")
	assert.Cmp(httpmock.GetCallCountInfo()["POST https://eu.api.ovh.com/v1/dedicated/server/fakeBaremetal/reboot"], 1,
		"only the consented reboot must have happened")
}

// reboot-rescue leaves the server in rescue until somebody sets the boot back,
// so it is guarded like the reboot it performs.
func (ms *MockSuite) TestBaremetalRebootRescueRefusesWithoutConsent(assert, require *td.T) {
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/dedicated/server/fakeBaremetal/boot?bootType=rescue",
		httpmock.NewStringResponder(200, `[1122]`),
	)

	_, err := cmd.Execute("baremetal", "reboot-rescue", "fakeBaremetal")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("cancelled"))
	assert.Cmp(httpmock.GetTotalCallCount(), 0, "not even the boot lookup must happen")
}

// Resetting an aggregation takes the interfaces down; on a server reached over
// that link the operator is cutting the branch they sit on.
func (ms *MockSuite) TestBaremetalOlaResetRefusesWithoutConsent(assert, require *td.T) {
	httpmock.RegisterResponder("POST", "https://eu.api.ovh.com/v1/dedicated/server/fakeBaremetal/ola/reset",
		httpmock.NewStringResponder(200, `{}`),
	)

	_, err := cmd.Execute("baremetal", "vni", "ola-reset", "fakeBaremetal", "--interface", "uuid-1")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("cancelled"))
	assert.Cmp(httpmock.GetTotalCallCount(), 0, "no interface must be reset without a confirmation")
}
