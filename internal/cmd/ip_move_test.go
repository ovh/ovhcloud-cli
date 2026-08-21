// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"github.com/jarcoal/httpmock"
	"github.com/maxatome/go-testdeep/td"
	"github.com/ovh/ovhcloud-cli/internal/cmd"
)

const ipMoveDestinations = `{
	"dedicatedServer": [
		{"service": "ns0000005.ip-203-0-113.eu", "nexthop": []},
		{"service": "ns0000006.ip-203-0-113.eu", "nexthop": []}
	],
	"vps": [{"service": "vps-c924a68c.vps.ovh.net", "nexthop": ["1.2.3.4"]}],
	"dedicatedCloud": [],
	"hostingReseller": []
}`

func registerIpMoveDestinations() {
	httpmock.RegisterResponder("GET",
		"https://eu.api.ovh.com/v1/ip/203.0.113.80%2F32/move",
		httpmock.NewStringResponder(200, ipMoveDestinations))
}

// A family that accepts nothing is returned as an empty list rather than being
// absent. Showing it would be four empty sections in a table whose whole point
// is to say where the IP can go.
func (ms *MockSuite) TestIpDestinationsDropsEmptyFamilies(assert, require *td.T) {
	registerIpMoveDestinations()

	out, err := cmd.Execute("ip", "destinations", "203.0.113.80/32")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("ns0000006.ip-203-0-113.eu"))
	assert.Cmp(out, td.Contains("vps-c924a68c.vps.ovh.net"))
	assert.Cmp(out, td.Not(td.Contains("hostingReseller")), "an empty family is not a destination")
	assert.Cmp(out, td.Not(td.Contains("dedicatedCloud")))
}

// The destination is checked before the request. The API would answer 400
// several seconds later without saying which names would have worked, and the
// CLI is holding that list already.
func (ms *MockSuite) TestIpMoveRefusesAnImpossibleDestinationLocally(assert, require *td.T) {
	registerIpMoveDestinations()

	_, err := cmd.Execute("ip", "move", "203.0.113.80/32", "ns9999999.example", "--yes")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("ns9999999.example"))
	assert.Cmp(err.Error(), td.Contains("2 dedicatedServer"), "the families are counted")
	assert.Cmp(err.Error(), td.Contains("ovhcloud ip destinations"), "and the way out is named")

	assert.Cmp(httpmock.GetTotalCallCount() > 0, true)
	for call := range httpmock.GetCallCountInfo() {
		assert.Cmp(call, td.Not(td.HasPrefix("POST")), "nothing may be sent for a refused move")
	}
}

// --dry-run prints the call and sends nothing.
func (ms *MockSuite) TestIpMoveDryRunSendsNothing(assert, require *td.T) {
	registerIpMoveDestinations()

	out, err := cmd.Execute("ip", "move", "203.0.113.80/32", "ns0000006.ip-203-0-113.eu", "--dry-run")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("Dry run"))
	assert.Cmp(out, td.Contains("ns0000006.ip-203-0-113.eu"))
	for call := range httpmock.GetCallCountInfo() {
		assert.Cmp(call, td.Not(td.HasPrefix("POST")))
	}
}

// Parking an IP that serves nothing is not a failure and not a request: it is
// already in the state being asked for.
func (ms *MockSuite) TestIpParkSaysWhenThereIsNothingToPark(assert, require *td.T) {
	httpmock.RegisterResponder("GET",
		"https://eu.api.ovh.com/v1/ip/203.0.113.80%2F32",
		httpmock.NewStringResponder(200, `{"ip": "203.0.113.80/32", "routedTo": {"serviceName": ""}}`))

	out, err := cmd.Execute("ip", "park", "203.0.113.80/32", "--yes")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("nothing to park"))
	for call := range httpmock.GetCallCountInfo() {
		assert.Cmp(call, td.Not(td.HasPrefix("POST")))
	}
}

// The other half of eb56fab that had no test. "Nothing to park" is the sentence
// an operator reads as "already done, move on", and a failed read must never
// produce it: parking waits for the empty string, and the empty string is also
// what a folded-in error looked like. So an unreadable IP stops the command
// instead of being reported as one that serves nothing.
func (ms *MockSuite) TestIpParkRefusesWhenItCannotTellWhatIsServed(assert, require *td.T) {
	httpmock.RegisterResponder("GET",
		"https://eu.api.ovh.com/v1/ip/203.0.113.80%2F32",
		httpmock.NewStringResponder(500, `{"message": "gateway is having a day"}`))

	_, err := cmd.Execute("ip", "park", "203.0.113.80/32", "--yes")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("failed to read what 203.0.113.80/32 is routed to"))
	assert.Cmp(err.Error(), td.Not(td.Contains("nothing to park")),
		"an unread IP is never an unrouted one")
	for call := range httpmock.GetCallCountInfo() {
		assert.Cmp(call, td.Not(td.HasPrefix("POST")), "and nothing is sent")
	}
}
