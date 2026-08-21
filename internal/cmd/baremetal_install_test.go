// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"github.com/jarcoal/httpmock"
	"github.com/maxatome/go-testdeep/td"
	"github.com/ovh/ovhcloud-cli/internal/cmd"
)

// The schemes depend on the template, so the template travels in the query
// rather than being dropped: asking without it returns the schemes of no
// particular OS, which is not a question anybody has.
func (ms *MockSuite) TestBaremetalPartitionSchemesAreAskedForOneTemplate(assert, require *td.T) {
	httpmock.RegisterResponder("GET",
		"https://eu.api.ovh.com/v1/dedicated/server/fakeBaremetal/install/compatibleTemplatePartitionSchemes?templateName=debian12_64",
		httpmock.NewStringResponder(200, `["default", "custom"]`))

	out, err := cmd.Execute("baremetal", "list-partition-schemes", "fakeBaremetal", "--os", "debian12_64")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("default"))
	assert.Cmp(out, td.Contains("custom"))
}

// --os is what makes the answer mean something, so it is required rather than
// silently defaulted.
func (ms *MockSuite) TestBaremetalPartitionSchemesRequireATemplate(assert, require *td.T) {
	_, err := cmd.Execute("baremetal", "list-partition-schemes", "fakeBaremetal")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("os"))
}

// Most servers have no RAID controller — every one of the 14 checked on a real
// account. The API says so with a 403, which is an answer, not a failure: it
// tells the operator to reach for software RAID in the partitioning layout.
// Reporting it as an error would send them looking for a broken command.
func (ms *MockSuite) TestBaremetalRaidProfileTreatsNoControllerAsAnAnswer(assert, require *td.T) {
	httpmock.RegisterResponder("GET",
		"https://eu.api.ovh.com/v1/dedicated/server/fakeBaremetal/install/hardwareRaidProfile",
		httpmock.NewStringResponder(403, `{"message": "Hardware RAID is not supported by this server"}`))

	out, err := cmd.Execute("baremetal", "raid-profile", "fakeBaremetal")

	require.CmpNoError(err, "a server without a controller is not a failed command")
	assert.Cmp(out, td.Contains("no hardware RAID controller"))
	assert.Cmp(out, td.Contains("raidLevel"), "and it names what to use instead")
}

// A genuine failure must stay one. The message match only ever softens the
// no-controller case; everything else is reported as the error it is.
func (ms *MockSuite) TestBaremetalRaidProfileStillReportsRealFailures(assert, require *td.T) {
	httpmock.RegisterResponder("GET",
		"https://eu.api.ovh.com/v1/dedicated/server/fakeBaremetal/install/hardwareRaidProfile",
		httpmock.NewStringResponder(403, `{"message": "This call has not been granted"}`))

	_, err := cmd.Execute("baremetal", "raid-profile", "fakeBaremetal")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("not been granted"))
}

func (ms *MockSuite) TestBaremetalRaidProfileListsControllersAndDisks(assert, require *td.T) {
	httpmock.RegisterResponder("GET",
		"https://eu.api.ovh.com/v1/dedicated/server/fakeBaremetal/install/hardwareRaidProfile",
		httpmock.NewStringResponder(200, `{"controllers": [{"model": "PERC H730P", "type": "Hardware Raid",
			"disks": [{"diskGroupId": 1, "names": ["sda", "sdb", "sdc", "sdd"], "capacity": "3726GiB", "type": "HDD", "speed": "7200rpm"}]}]}`))

	out, err := cmd.Execute("baremetal", "raid-profile", "fakeBaremetal")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("PERC H730P"))
	assert.Cmp(out, td.Contains("3726GiB"))
}

// Outside an installation the route answers 404 with a sentence, not a
// payload. That is the state most servers are in most of the time, and it is
// an answer: reporting it as a failure would send somebody looking for a
// broken command.
func (ms *MockSuite) TestBaremetalInstallStatusTreatsIdleAsAnAnswer(assert, require *td.T) {
	httpmock.RegisterResponder("GET",
		"https://eu.api.ovh.com/v1/dedicated/server/fakeBaremetal/install/status",
		httpmock.NewStringResponder(404,
			`{"message": "Server is not being installed or reinstalled at the moment"}`))

	out, err := cmd.Execute("baremetal", "install-status", "fakeBaremetal")

	require.CmpNoError(err, "an idle server is not a failed command")
	assert.Cmp(out, td.Contains("not being installed"))
}

// A genuine failure stays one: the message match only ever softens the idle
// case.
func (ms *MockSuite) TestBaremetalInstallStatusStillReportsRealFailures(assert, require *td.T) {
	httpmock.RegisterResponder("GET",
		"https://eu.api.ovh.com/v1/dedicated/server/fakeBaremetal/install/status",
		httpmock.NewStringResponder(403, `{"message": "This call has not been granted"}`))

	_, err := cmd.Execute("baremetal", "install-status", "fakeBaremetal")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("not been granted"))
}

// The point of the command is to say which step is running and for how long.
// A list of steps without either is the same non-answer --wait used to give.
func (ms *MockSuite) TestBaremetalInstallStatusNamesTheRunningStep(assert, require *td.T) {
	httpmock.RegisterResponder("GET",
		"https://eu.api.ovh.com/v1/dedicated/server/fakeBaremetal/install/status",
		httpmock.NewStringResponder(200, `{
			"elapsedTime": 754,
			"progress": [
				{"comment": "Initialising Installation process", "status": "done"},
				{"comment": "Installing operating system", "status": "doing"},
				{"comment": "Rebooting", "status": "todo"}
			]
		}`))

	out, err := cmd.Execute("baremetal", "install-status", "fakeBaremetal")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("step 2 of 3"), "the position has to be visible")
	assert.Cmp(out, td.Contains("Installing operating system"))
	assert.Cmp(out, td.Contains("12m34s"), "754 seconds, not 754")
	assert.Cmp(out, td.Contains("API reports"),
		"the figure is the API's: measured shifting origin mid-install")
	assert.Cmp(out, td.Contains("Rebooting"), "and the steps still to come")
}

// The error of a failed step is what says why, so it travels with it.
func (ms *MockSuite) TestBaremetalInstallStatusShowsWhyAStepFailed(assert, require *td.T) {
	httpmock.RegisterResponder("GET",
		"https://eu.api.ovh.com/v1/dedicated/server/fakeBaremetal/install/status",
		httpmock.NewStringResponder(200, `{
			"elapsedTime": 42,
			"progress": [
				{"comment": "Partitioning", "status": "error", "error": "disk 2 is not present"}
			]
		}`))

	out, err := cmd.Execute("baremetal", "install-status", "fakeBaremetal")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("disk 2 is not present"))
}

// --filter is registered on both commands, so it has to do something. A test
// that only checks the kept row would pass just as well with the filtering
// removed, which is why the assertion that matters is the absence of the other
// one.
func (ms *MockSuite) TestBaremetalPartitionSchemesAreFiltered(assert, require *td.T) {
	httpmock.RegisterResponder("GET",
		"https://eu.api.ovh.com/v1/dedicated/server/fakeBaremetal/install/compatibleTemplatePartitionSchemes?templateName=debian12_64",
		httpmock.NewStringResponder(200, `["default", "custom"]`))

	out, err := cmd.Execute("baremetal", "list-partition-schemes", "fakeBaremetal",
		"--os", "debian12_64", "--filter", `name=="custom"`)

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("custom"))
	assert.Cmp(out, td.Not(td.Contains("default")), "the row the filter excludes must not be printed")
}

func (ms *MockSuite) TestBaremetalRaidProfileIsFiltered(assert, require *td.T) {
	httpmock.RegisterResponder("GET",
		"https://eu.api.ovh.com/v1/dedicated/server/fakeBaremetal/install/hardwareRaidProfile",
		httpmock.NewStringResponder(200, `{"controllers": [
			{"model": "PERC H730P", "type": "Hardware Raid", "disks": [
				{"diskGroupId": 1, "names": ["sda"], "type": "HDD"}]},
			{"model": "PERC H740P", "type": "Hardware Raid", "disks": [
				{"diskGroupId": 2, "names": ["nvme0n1"], "type": "SSD"}]}]}`))

	out, err := cmd.Execute("baremetal", "raid-profile", "fakeBaremetal",
		"--filter", `diskType=="SSD"`)

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("H740P"))
	assert.Cmp(out, td.Not(td.Contains("H730P")), "the HDD group the filter excludes must not be printed")
}

// capacity and speed are complexType.UnitAndValue in the schema, not scalars.
// Left untyped they reached the table as the decoded map and rendered as
// "map[unit:GB value:1000]" — on every server that actually has a controller,
// which is none of the fourteen measured. Hence a defect nobody could see.
func (ms *MockSuite) TestBaremetalRaidProfileRendersQuantitiesWithTheirUnit(assert, require *td.T) {
	httpmock.RegisterResponder("GET",
		"https://eu.api.ovh.com/v1/dedicated/server/fakeBaremetal/install/hardwareRaidProfile",
		httpmock.NewStringResponder(200, `{"controllers":[{"model":"PERC H755","type":"Hardware","disks":[
			{"capacity":{"unit":"GB","value":1000},"diskGroupId":1,"names":["disk0","disk1"],
			 "speed":{"unit":"rpm","value":7200},"type":"SATA"}]}]}`))

	out, err := cmd.Execute("baremetal", "raid-profile", "fakeBaremetal")

	require.CmpNoError(err)
	assert.Cmp(out, td.Not(td.Contains("map[")), "the decoded object must not reach a cell")
	assert.Cmp(out, td.Contains("1000 GB"))
	assert.Cmp(out, td.Contains("7200 rpm"))
}

// An absent quantity renders as nothing, not as a zero the API never sent.
func (ms *MockSuite) TestBaremetalRaidProfileLeavesAnAbsentQuantityEmpty(assert, require *td.T) {
	httpmock.RegisterResponder("GET",
		"https://eu.api.ovh.com/v1/dedicated/server/fakeBaremetal/install/hardwareRaidProfile",
		httpmock.NewStringResponder(200, `{"controllers":[{"model":"PERC H755","type":"Hardware","disks":[
			{"diskGroupId":1,"names":["disk0"],"type":"SATA"}]}]}`))

	out, err := cmd.Execute("baremetal", "raid-profile", "fakeBaremetal")

	require.CmpNoError(err)
	assert.Cmp(out, td.Not(td.Contains("0 ")), "no invented zero")
}
