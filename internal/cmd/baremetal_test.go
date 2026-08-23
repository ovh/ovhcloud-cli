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

	_, err := cmd.Execute("baremetal", "reinstall", "fakeBaremetal", "--os", "debian12_64", "--wait")

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

	_, err := cmd.Execute("baremetal", "reinstall", "fakeBaremetal", "--os", "debian12_64", "--wait")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("156839472"))
	assert.Cmp(err.Error(), td.Not(td.Contains("%!d")), "the identifier must be readable")
	assert.Cmp(err.Error(), td.Contains("list-tasks fakeBaremetal"), "and the message says how to follow up")
}

// A status this CLI does not know is not a failure of the task: the API enum
// can grow, and guessing would report a success or a failure that never was.
func (ms *MockSuite) TestBaremetalReinstallReportsAnUnknownStatus(assert, require *td.T) {
	registerReinstallTask(`{"taskId": 156839472, "function": "reinstallServer", "status": "quarantined"}`)

	_, err := cmd.Execute("baremetal", "reinstall", "fakeBaremetal", "--os", "debian12_64", "--wait")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("unexpected status quarantined"))
}

// A task that completes normally must still report success.
func (ms *MockSuite) TestBaremetalReinstallSucceedsOnDoneTask(assert, require *td.T) {
	registerReinstallTask(`{"taskId": 156839472, "function": "reinstallServer", "status": "done"}`)
	httpmock.RegisterResponder("POST", "https://eu.api.ovh.com/v1/dedicated/server/fakeBaremetal/authenticationSecret",
		httpmock.NewStringResponder(200, `[]`),
	)

	_, err := cmd.Execute("baremetal", "reinstall", "fakeBaremetal", "--os", "debian12_64", "--wait")

	require.CmpNoError(err)
}

func mockCredentialsResponders() {
	httpmock.RegisterResponder("POST", "https://eu.api.ovh.com/v1/dedicated/server/fakeBaremetal/authenticationSecret",
		httpmock.NewStringResponder(200, `[{"type":"password","user":"root","password":"secret-id-1","expiration":"2026-08-18T03:41:00+02:00"}]`),
	)
	httpmock.RegisterResponder("POST", "https://eu.api.ovh.com/v1/secret/retrieve",
		httpmock.NewStringResponder(200, `{"secret":"Tk9uY2VQYXNzMTIz"}`),
	)
}

// A secret must not be printed unless it was explicitly asked for.
func (ms *MockSuite) TestBaremetalCredentialsCreateMasksByDefault(assert, require *td.T) {
	mockCredentialsResponders()

	out, err := cmd.Execute("baremetal", "credentials", "create", "fakeBaremetal")

	require.CmpNoError(err)
	assert.Cmp(out, td.Not(td.Contains("Tk9uY2VQYXNzMTIz")), "the secret value must not be printed")
	assert.Cmp(out, td.Contains("••••"))
}

// Masking applies to machine-readable formats too: a secret written into a
// JSON file in a pipeline leaks as much as one scrolled in a terminal.
func (ms *MockSuite) TestBaremetalCredentialsCreateMasksJSONOutput(assert, require *td.T) {
	mockCredentialsResponders()

	out, err := cmd.Execute("baremetal", "credentials", "create", "fakeBaremetal", "-o", "json")

	require.CmpNoError(err)
	assert.Cmp(out, td.Not(td.Contains("Tk9uY2VQYXNzMTIz")))
}

func (ms *MockSuite) TestBaremetalCredentialsCreateRevealsOnDemand(assert, require *td.T) {
	mockCredentialsResponders()

	out, err := cmd.Execute("baremetal", "credentials", "create", "fakeBaremetal", "--reveal")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("Tk9uY2VQYXNzMTIz"))
}

// The former name keeps working, so existing scripts do not break.
func (ms *MockSuite) TestBaremetalListSecretsAliasStillWorks(assert, require *td.T) {
	mockCredentialsResponders()

	out, err := cmd.Execute("baremetal", "list-secrets", "fakeBaremetal")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("••••"))
}

// The command must fail loudly when the secret cannot be retrieved, and must
// not print anything that looks like credential material on the way out.
func (ms *MockSuite) TestBaremetalCredentialsCreateFailsWithoutLeaking(assert, require *td.T) {
	httpmock.RegisterResponder("POST", "https://eu.api.ovh.com/v1/dedicated/server/fakeBaremetal/authenticationSecret",
		httpmock.NewStringResponder(200, `[{"type":"password","user":"root","password":"secret-id-1"}]`),
	)
	httpmock.RegisterResponder("POST", "https://eu.api.ovh.com/v1/secret/retrieve",
		httpmock.NewStringResponder(500, `{"message":"internal error"}`),
	)

	out, err := cmd.Execute("baremetal", "credentials", "create", "fakeBaremetal", "--reveal")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("failed to retrieve secret value"))
	assert.Cmp(out, td.Not(td.Contains("secret-id-1")), "not even the secret identifier is printed")
}

// The generation call failing must be reported the same way.
func (ms *MockSuite) TestBaremetalCredentialsCreateReportsGenerationFailure(assert, require *td.T) {
	httpmock.RegisterResponder("POST", "https://eu.api.ovh.com/v1/dedicated/server/fakeBaremetal/authenticationSecret",
		httpmock.NewStringResponder(403, `{"message":"This call has not been granted"}`),
	)

	_, err := cmd.Execute("baremetal", "credentials", "create", "fakeBaremetal")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("failed to fetch secrets IDs"))
}
