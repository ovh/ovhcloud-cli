// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"strings"

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

func (ms *MockSuite) TestBaremetalBootSetDiskCmd(assert, require *td.T) {
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/dedicated/server/fakeBaremetal/boot?bootType=harddisk",
		httpmock.NewStringResponder(200, `[1]`),
	)
	httpmock.RegisterResponder("PUT", "https://eu.api.ovh.com/v1/dedicated/server/fakeBaremetal",
		httpmock.NewStringResponder(200, `null`),
	)

	out, err := cmd.Execute("baremetal", "boot", "set-disk", "fakeBaremetal")

	require.CmpNoError(err)
	assert.Contains(out, "set to boot on its hard disk")
}

func (ms *MockSuite) TestBaremetalBootSetDiskCmdNoEntry(assert, require *td.T) {
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/dedicated/server/fakeBaremetal/boot?bootType=harddisk",
		httpmock.NewStringResponder(200, `[]`),
	)

	_, err := cmd.Execute("baremetal", "boot", "set-disk", "fakeBaremetal")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("no hard disk boot entry found"))
}

// A server left with the rescue boot entry comes back in rescue at its next
// reboot: the sheet must say so.
func (ms *MockSuite) TestBaremetalGetCmdWarnsAboutRescueBoot(assert, require *td.T) {
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/dedicated/server/fakeBaremetal",
		httpmock.NewStringResponder(200, `{
			"name": "fakeBaremetal",
			"bootId": 1122,
			"os": "debian12_64",
			"state": "ok",
			"powerState": "poweron",
			"ip": "1.2.3.4",
			"reverse": "fake.example.net",
			"region": "eu-west",
			"availabilityZone": "eu-west-gra-a",
			"datacenter": "gra3",
			"rack": "G123A01",
			"iam": {"displayName": "fake", "urn": "urn:v1:eu:resource:dedicatedServer:fakeBaremetal"}
		}`),
	)
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/dedicated/server/fakeBaremetal/task",
		httpmock.NewStringResponder(200, `[]`),
	)
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/dedicated/server/fakeBaremetal/specifications/network",
		httpmock.NewStringResponder(200, `{"routing": {"ipv6": null}}`),
	)
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/dedicated/server/fakeBaremetal/serviceInfos",
		httpmock.NewStringResponder(200, `{"expiration": "2026-11-04", "renew": {"automatic": true}}`),
	)
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/dedicated/server/fakeBaremetal/boot/1122",
		httpmock.NewStringResponder(200, `{"bootId": 1122, "bootType": "rescue", "description": "rescue64-pro", "kernel": "rescue"}`),
	)

	out, err := cmd.Execute("baremetal", "get", "fakeBaremetal")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("rescue"))
	assert.Cmp(out, td.Contains("boot set-disk"))
}

// registerBootListResponders wires the calls made by `baremetal boot list`:
// the list of boot identifiers, one object per entry, and the options of each.
func registerBootListResponders() {
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/dedicated/server/fakeBaremetal/boot",
		httpmock.NewStringResponder(200, `[1, 1122]`),
	)
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/dedicated/server/fakeBaremetal/boot/1",
		httpmock.NewStringResponder(200, `{"bootId": 1, "bootType": "harddisk", "description": "Boot on hard disk", "kernel": ""}`),
	)
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/dedicated/server/fakeBaremetal/boot/1122",
		httpmock.NewStringResponder(200, `{"bootId": 1122, "bootType": "rescue", "description": "rescue64-pro", "kernel": "rescue"}`),
	)
	for _, id := range []string{"1", "1122"} {
		httpmock.RegisterResponder("GET",
			"https://eu.api.ovh.com/v1/dedicated/server/fakeBaremetal/boot/"+id+"/option",
			httpmock.NewStringResponder(200, `[]`),
		)
	}
}

// The marker is the whole point of the column: without it, a server left in
// rescue mode looks exactly like one booting on its disk.
func (ms *MockSuite) TestBaremetalBootListCmdMarksTheActiveEntry(assert, require *td.T) {
	registerBootListResponders()
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/dedicated/server/fakeBaremetal",
		httpmock.NewStringResponder(200, `{"name": "fakeBaremetal", "bootId": 1122}`),
	)

	out, err := cmd.Execute("baremetal", "boot", "list", "fakeBaremetal")

	require.CmpNoError(err)
	// The identifier comes first, as in every other service list.
	assert.Cmp(out, td.Contains("bootId"))
	// Only the rescue entry, which the server is set to boot on, is marked.
	rescue := lineContaining(out, "rescue64-pro")
	disk := lineContaining(out, "Boot on hard disk")
	assert.Cmp(rescue, td.Contains("→"), "the active entry carries the marker")
	assert.Cmp(disk, td.Not(td.Contains("→")), "the other entries do not")
}

// Reading the current boot is best effort: if the server cannot be fetched,
// the listing must still answer, without claiming an entry is active.
func (ms *MockSuite) TestBaremetalBootListCmdSurvivesServerLookupFailure(assert, require *td.T) {
	registerBootListResponders()
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/dedicated/server/fakeBaremetal",
		httpmock.NewStringResponder(403, `{"message": "This call has not been granted"}`),
	)

	out, err := cmd.Execute("baremetal", "boot", "list", "fakeBaremetal")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("rescue64-pro"), "the listing still answers")
	assert.Cmp(out, td.Not(td.Contains("→")), "no entry is claimed active")
}

// lineContaining returns the first line of out holding needle, or "".
func lineContaining(out, needle string) string {
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, needle) {
			return line
		}
	}
	return ""
}
