// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

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

const fakeTerminationToken = "abcd-1234-token"

func registerBaremetalTermination(captured *map[string]any) {
	httpmock.RegisterResponder("POST", "https://eu.api.ovh.com/v1/dedicated/server/fakeBaremetal/terminate",
		httpmock.NewStringResponder(200, `"termination requested"`),
	)
	httpmock.RegisterResponder("POST", "https://eu.api.ovh.com/v1/dedicated/server/fakeBaremetal/confirmTermination",
		func(req *http.Request) (*http.Response, error) {
			body, err := io.ReadAll(req.Body)
			if err != nil {
				return nil, err
			}
			var sent map[string]any
			if err := json.Unmarshal(body, &sent); err != nil {
				return nil, err
			}
			*captured = sent
			return httpmock.NewStringResponse(200, `"ok"`), nil
		},
	)
}

// terminate stops nothing, and the token is emailed rather than returned. An
// operator who reads "termination started: <body>" pastes that body into the
// confirmation and wonders why it is refused.
func (ms *MockSuite) TestBaremetalTerminateSaysWhereTheTokenComesFrom(assert, require *td.T) {
	var sent map[string]any
	registerBaremetalTermination(&sent)

	out, err := cmd.Execute("baremetal", "terminate", "fakeBaremetal", "--yes")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("emailed"), "the operator must be told where the token arrives")
	assert.Cmp(out, td.Contains("confirm-termination fakeBaremetal"), "and what to run next")
}

// The reversible half asks for a yes; the irreversible half asks for the name.
// An unattended run gets neither, so it must reach nothing.
func (ms *MockSuite) TestBaremetalTerminationRefusesWithoutConsent(assert, require *td.T) {
	var sent map[string]any
	registerBaremetalTermination(&sent)

	_, err := cmd.Execute("baremetal", "terminate", "fakeBaremetal")
	require.CmpError(err)

	_, err = cmd.Execute("baremetal", "confirm-termination", "fakeBaremetal", fakeTerminationToken)
	require.CmpError(err)

	assert.Cmp(httpmock.GetTotalCallCount(), 0, "nothing must reach the API without a confirmation")
}

func (ms *MockSuite) TestBaremetalConfirmTerminationSendsTheSurveyFields(assert, require *td.T) {
	var sent map[string]any
	registerBaremetalTermination(&sent)

	_, err := cmd.Execute("baremetal", "confirm-termination", "fakeBaremetal", fakeTerminationToken,
		"--reason", "TOO_EXPENSIVE", "--future-use", "NOT_REPLACING_SERVICE",
		"--commentary", "consolidating racks", "--yes")

	require.CmpNoError(err)
	assert.Cmp(sent["token"], fakeTerminationToken)
	assert.Cmp(sent["reason"], "TOO_EXPENSIVE")
	assert.Cmp(sent["futureUse"], "NOT_REPLACING_SERVICE")
	assert.Cmp(sent["commentary"], "consolidating racks")
}

// A value the API would reject must be named as such here, with the accepted
// ones listed. A 400 from the other side says "invalid value" and stops there.
func (ms *MockSuite) TestBaremetalConfirmTerminationRejectsAnUnknownReason(assert, require *td.T) {
	var sent map[string]any
	registerBaremetalTermination(&sent)

	_, err := cmd.Execute("baremetal", "confirm-termination", "fakeBaremetal", fakeTerminationToken,
		"--reason", "BECAUSE", "--yes")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("TOO_EXPENSIVE"), "the accepted values are listed")
	assert.Cmp(httpmock.GetTotalCallCount(), 0, "and nothing is sent")
}

// The preview must describe the request, and withhold the one value that is a
// single-use credential: --dry-run is run with the output on a screen or in a
// pipeline log.
func (ms *MockSuite) TestBaremetalConfirmTerminationDryRunWithholdsTheToken(assert, require *td.T) {
	var sent map[string]any
	registerBaremetalTermination(&sent)

	out, err := cmd.Execute("baremetal", "confirm-termination", "fakeBaremetal", fakeTerminationToken,
		"--reason", "TOO_EXPENSIVE", "--dry-run")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("confirmTermination"), "the call is named")
	assert.Cmp(out, td.Contains("reason: TOO_EXPENSIVE"), "and so are the fields")
	assert.Cmp(out, td.Not(td.Contains(fakeTerminationToken)), "but never the token itself")
	assert.Cmp(httpmock.GetTotalCallCount(), 0, "and nothing is sent")
}

// Editing one renewal setting must not carry the others along at their zero
// value, on baremetal as anywhere else.
func (ms *MockSuite) TestBaremetalServiceInfoEditSendsOnlyWhatWasAsked(assert, require *td.T) {
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/dedicated/server/fakeBaremetal/serviceInfos",
		httpmock.NewStringResponder(200, `{
			"serviceId": 1,
			"domain": "fakeBaremetal",
			"renew": {"automatic": true, "deleteAtExpiration": false, "forced": false, "manualPayment": false, "period": 1}
		}`),
	)
	var sent map[string]any
	httpmock.RegisterResponder("PUT", "https://eu.api.ovh.com/v1/dedicated/server/fakeBaremetal/serviceInfos",
		func(req *http.Request) (*http.Response, error) {
			body, err := io.ReadAll(req.Body)
			if err != nil {
				return nil, err
			}
			if err := json.Unmarshal(body, &sent); err != nil {
				return nil, err
			}
			return httpmock.NewStringResponse(200, `null`), nil
		},
	)

	_, err := cmd.Execute("baremetal", "service-info", "edit", "fakeBaremetal", "--renew-period", "12")

	require.CmpNoError(err)
	renew, _ := sent["renew"].(map[string]any)
	require.NotNil(renew)
	assert.Cmp(renew["period"], float64(12))
	assert.Cmp(renew["automatic"], true, "automatic renewal must survive untouched")
}

// The termination survey is held in package-level variables bound by cobra. A
// reason typed on one command must not attach itself to the next termination
// of the same process, which would file a survey answer nobody gave.
func (ms *MockSuite) TestBaremetalTerminationReasonDoesNotSurvive(assert, require *td.T) {
	var sent map[string]any
	registerBaremetalTermination(&sent)

	_, err := cmd.Execute("baremetal", "confirm-termination", "fakeBaremetal", fakeTerminationToken,
		"--reason", "TOO_EXPENSIVE", "--yes")
	require.CmpNoError(err)
	require.Cmp(sent["reason"], "TOO_EXPENSIVE")

	cmd.PostExecute()

	_, err = cmd.Execute("baremetal", "confirm-termination", "fakeBaremetal", fakeTerminationToken, "--yes")
	require.CmpNoError(err)
	assert.Cmp(sent["reason"], nil, "the second termination carries no reason")
}

// cobra counts the positional arguments, it does not look at them: an empty
// token satisfies ExactArgs(2). Sending it spends a round trip to learn what
// the CLI already knew.
func (ms *MockSuite) TestBaremetalConfirmTerminationRefusesAnEmptyToken(assert, require *td.T) {
	var sent map[string]any
	registerBaremetalTermination(&sent)

	_, err := cmd.Execute("baremetal", "confirm-termination", "fakeBaremetal", "   ", "--yes")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("no termination token given"))
	assert.Cmp(httpmock.GetTotalCallCount(), 0, "nothing must reach the API")
}

// An error from the API must not be reported as a success, whatever the CLI
// does with the response body.
func (ms *MockSuite) TestBaremetalConfirmTerminationReportsAnApiRefusal(assert, require *td.T) {
	httpmock.RegisterResponder("POST", "https://eu.api.ovh.com/v1/dedicated/server/fakeBaremetal/confirmTermination",
		httpmock.NewStringResponder(400, `{"class":"Client::BadRequest","message":"This token is not valid"}`),
	)

	out, err := cmd.Execute("baremetal", "confirm-termination", "fakeBaremetal", fakeTerminationToken, "--yes")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("This token is not valid"), "the reason the API gave")
	assert.Cmp(out, td.Not(td.Contains("confirmed")), "and no claim that it worked")
}

// The preview must not hide a token the shell mangled: four characters and a
// length answer that without reproducing the credential.
func (ms *MockSuite) TestBaremetalConfirmTerminationDryRunFingerprintsTheToken(assert, require *td.T) {
	var sent map[string]any
	registerBaremetalTermination(&sent)

	out, err := cmd.Execute("baremetal", "confirm-termination", "fakeBaremetal", fakeTerminationToken, "--dry-run")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("abcd…"), "enough to recognise the token")
	assert.Cmp(out, td.Contains(fmt.Sprintf("%d characters", len(fakeTerminationToken))), "and its length")
	assert.Cmp(out, td.Not(td.Contains(fakeTerminationToken)), "never the token itself")
}
