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

// compatibleOSNames lists the OS names the fake compatibleTemplates responder
// declares available for "fakeBaremetal", shared across the compatible-OS tests.
var compatibleOSNames = []string{
	"alma8-cpanel-latest_64",
	"alma8-plesk18_64",
	"alma8_64",
	"alma9-cpanel-latest_64",
	"alma9-plesk18_64",
	"alma9_64",
	"byoi_64",
	"byolinux_64",
}

// templateInfosFixture is a GET /dedicated/installationTemplate/templateInfos
// response covering every name in compatibleOSNames, plus one extra template
// ("debian12_64") that is NOT compatible, to prove it gets filtered out.
const templateInfosFixture = `[
	{"templateName": "alma8-cpanel-latest_64", "description": "cPanel (AlmaLinux 8)", "category": "management", "family": "linux", "subfamily": "alma", "endOfInstall": "2029-03-06"},
	{"templateName": "alma8-plesk18_64", "description": "Linux Plesk Obsidian (AlmaLinux 8)", "category": "management", "family": "linux", "subfamily": "alma", "endOfInstall": "2029-03-06"},
	{"templateName": "alma8_64", "description": "AlmaLinux 8", "category": "basic", "family": "linux", "subfamily": "alma", "endOfInstall": "2029-03-06"},
	{"templateName": "alma9-cpanel-latest_64", "description": "cPanel (AlmaLinux 9)", "category": "management", "family": "linux", "subfamily": "alma", "endOfInstall": "2032-06-01"},
	{"templateName": "alma9-plesk18_64", "description": "Linux Plesk Obsidian (AlmaLinux 9)", "category": "management", "family": "linux", "subfamily": "alma", "endOfInstall": "2032-06-01"},
	{"templateName": "alma9_64", "description": "AlmaLinux 9", "category": "basic", "family": "linux", "subfamily": "alma", "endOfInstall": "2032-06-01"},
	{"templateName": "byoi_64", "description": "Bring Your Own Image", "category": "customer", "family": "custom", "subfamily": "byoi", "endOfInstall": "2999-12-31"},
	{"templateName": "byolinux_64", "description": "Bring Your Own Linux", "category": "customer", "family": "custom", "subfamily": "byolinux", "endOfInstall": "2999-12-31"},
	{"templateName": "debian12_64", "description": "Debian 12 (Bookworm)", "category": "basic", "family": "linux", "subfamily": "debian", "endOfInstall": "2028-07-04"}
]`

// registerCompatibleTemplatesResponder wires GET .../install/compatibleTemplates
// for "fakeBaremetal" to return compatibleOSNames under the "ovh" key.
func registerCompatibleTemplatesResponder() {
	names := `"` + compatibleOSNames[0] + `"`
	for _, name := range compatibleOSNames[1:] {
		names += `, "` + name + `"`
	}
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/dedicated/server/fakeBaremetal/install/compatibleTemplates",
		httpmock.NewStringResponder(200, `{"ovh": [`+names+`]}`),
	)
}

func (ms *MockSuite) TestBaremetalListCompatibleOSCmd(assert, require *td.T) {
	registerCompatibleTemplatesResponder()
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/dedicated/installationTemplate/templateInfos",
		httpmock.NewStringResponder(200, templateInfosFixture),
	)

	out, err := cmd.Execute("baremetal", "list-compatible-os", "fakeBaremetal")

	require.CmpNoError(err)
	assert.String(out, `
┌────────────────────────┬────────────────────────────────────┬────────────┬────────┬───────────┬──────────────┐
│          name          │            description             │  category  │ family │ subfamily │ endOfInstall │
├────────────────────────┼────────────────────────────────────┼────────────┼────────┼───────────┼──────────────┤
│ alma8-cpanel-latest_64 │ cPanel (AlmaLinux 8)               │ management │ linux  │ alma      │ 2029-03-06   │
│ alma8-plesk18_64       │ Linux Plesk Obsidian (AlmaLinux 8) │ management │ linux  │ alma      │ 2029-03-06   │
│ alma8_64               │ AlmaLinux 8                        │ basic      │ linux  │ alma      │ 2029-03-06   │
│ alma9-cpanel-latest_64 │ cPanel (AlmaLinux 9)               │ management │ linux  │ alma      │ 2032-06-01   │
│ alma9-plesk18_64       │ Linux Plesk Obsidian (AlmaLinux 9) │ management │ linux  │ alma      │ 2032-06-01   │
│ alma9_64               │ AlmaLinux 9                        │ basic      │ linux  │ alma      │ 2032-06-01   │
│ byoi_64                │ Bring Your Own Image               │ customer   │ custom │ byoi      │ 2999-12-31   │
│ byolinux_64            │ Bring Your Own Linux               │ customer   │ custom │ byolinux  │ 2999-12-31   │
└────────────────────────┴────────────────────────────────────┴────────────┴────────┴───────────┴──────────────┘
💡 Use option -o json or -o yaml to get the raw output with all information`[1:])
}

// -o name only needs the OS names: it must not trigger the extra
// installationTemplate/templateInfos call. No responder is registered for it,
// so the test would fail with "no responder found" if that call happened.
func (ms *MockSuite) TestBaremetalListCompatibleOSCmdNameOnlySkipsDetails(assert, require *td.T) {
	registerCompatibleTemplatesResponder()

	out, err := cmd.Execute("baremetal", "list-compatible-os", "fakeBaremetal", "-o", "name")

	require.CmpNoError(err)
	assert.String(out, `alma8-cpanel-latest_64
alma8-plesk18_64
alma8_64
alma9-cpanel-latest_64
alma9-plesk18_64
alma9_64
byoi_64
byolinux_64
`)
}

// A filter referencing a detail-only field (here "category") must still
// trigger the templateInfos call even when the output format is name-only.
func (ms *MockSuite) TestBaremetalListCompatibleOSCmdFilterOnDetailFieldFetchesDetails(assert, require *td.T) {
	registerCompatibleTemplatesResponder()
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/dedicated/installationTemplate/templateInfos",
		httpmock.NewStringResponder(200, templateInfosFixture),
	)

	out, err := cmd.Execute("baremetal", "list-compatible-os", "fakeBaremetal", "-o", "name", "--filter", `category=="customer"`)

	require.CmpNoError(err)
	assert.String(out, `byoi_64
byolinux_64
`)
}

func (ms *MockSuite) TestBaremetalListOsCmd(assert, require *td.T) {
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/dedicated/installationTemplate/templateInfos",
		httpmock.NewStringResponder(200, templateInfosFixture),
	)

	out, err := cmd.Execute("baremetal", "list-os")

	require.CmpNoError(err)
	assert.String(out, `
┌────────────────────────┬────────────────────────────────────┬────────────┬────────┬───────────┬──────────────┐
│          name          │            description             │  category  │ family │ subfamily │ endOfInstall │
├────────────────────────┼────────────────────────────────────┼────────────┼────────┼───────────┼──────────────┤
│ alma8-cpanel-latest_64 │ cPanel (AlmaLinux 8)               │ management │ linux  │ alma      │ 2029-03-06   │
│ alma8-plesk18_64       │ Linux Plesk Obsidian (AlmaLinux 8) │ management │ linux  │ alma      │ 2029-03-06   │
│ alma8_64               │ AlmaLinux 8                        │ basic      │ linux  │ alma      │ 2029-03-06   │
│ alma9-cpanel-latest_64 │ cPanel (AlmaLinux 9)               │ management │ linux  │ alma      │ 2032-06-01   │
│ alma9-plesk18_64       │ Linux Plesk Obsidian (AlmaLinux 9) │ management │ linux  │ alma      │ 2032-06-01   │
│ alma9_64               │ AlmaLinux 9                        │ basic      │ linux  │ alma      │ 2032-06-01   │
│ byoi_64                │ Bring Your Own Image               │ customer   │ custom │ byoi      │ 2999-12-31   │
│ byolinux_64            │ Bring Your Own Linux               │ customer   │ custom │ byolinux  │ 2999-12-31   │
│ debian12_64            │ Debian 12 (Bookworm)               │ basic      │ linux  │ debian    │ 2028-07-04   │
└────────────────────────┴────────────────────────────────────┴────────────┴────────┴───────────┴──────────────┘
💡 Use option -o json or -o yaml to get the raw output with all information`[1:])
}

// -o name only needs the OS names: it must use the cheap
// GET /dedicated/installationTemplate (plain name list) rather than
// templateInfos. No responder is registered for templateInfos, so the test
// would fail with "no responder found" if that call happened.
func (ms *MockSuite) TestBaremetalListOsCmdNameOnlySkipsDetails(assert, require *td.T) {
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/dedicated/installationTemplate",
		httpmock.NewStringResponder(200, `["alma8_64", "debian12_64"]`),
	)

	out, err := cmd.Execute("baremetal", "list-os", "-o", "name")

	require.CmpNoError(err)
	assert.String(out, `alma8_64
debian12_64
`)
}

// A filter referencing a detail-only field must still trigger the
// templateInfos call even when the output format is name-only.
func (ms *MockSuite) TestBaremetalListOsCmdFilterOnDetailFieldFetchesDetails(assert, require *td.T) {
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/dedicated/installationTemplate/templateInfos",
		httpmock.NewStringResponder(200, templateInfosFixture),
	)

	out, err := cmd.Execute("baremetal", "list-os", "-o", "name", "--filter", `category=="customer"`)

	require.CmpNoError(err)
	assert.String(out, `byoi_64
byolinux_64
`)
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
