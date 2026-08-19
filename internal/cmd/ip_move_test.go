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
		{"service": "ns3018397.ip-57-128-116.eu", "nexthop": []},
		{"service": "ns3118333.ip-51-68-100.eu", "nexthop": []}
	],
	"vps": [{"service": "vps-c924a68c.vps.ovh.net", "nexthop": ["1.2.3.4"]}],
	"dedicatedCloud": [],
	"hostingReseller": []
}`

func registerIpMoveDestinations() {
	httpmock.RegisterResponder("GET",
		"https://eu.api.ovh.com/v1/ip/135.125.71.80%2F32/move",
		httpmock.NewStringResponder(200, ipMoveDestinations))
}

// A family that accepts nothing is returned as an empty list rather than being
// absent. Showing it would be four empty sections in a table whose whole point
// is to say where the IP can go.
func (ms *MockSuite) TestIpDestinationsDropsEmptyFamilies(assert, require *td.T) {
	registerIpMoveDestinations()

	out, err := cmd.Execute("ip", "destinations", "135.125.71.80/32")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("ns3118333.ip-51-68-100.eu"))
	assert.Cmp(out, td.Contains("vps-c924a68c.vps.ovh.net"))
	assert.Cmp(out, td.Not(td.Contains("hostingReseller")), "an empty family is not a destination")
	assert.Cmp(out, td.Not(td.Contains("dedicatedCloud")))
}

// The destination is checked before the request. The API would answer 400
// several seconds later without saying which names would have worked, and the
// CLI is holding that list already.
func (ms *MockSuite) TestIpMoveRefusesAnImpossibleDestinationLocally(assert, require *td.T) {
	registerIpMoveDestinations()

	_, err := cmd.Execute("ip", "move", "135.125.71.80/32", "ns9999999.example", "--yes")

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

	out, err := cmd.Execute("ip", "move", "135.125.71.80/32", "ns3118333.ip-51-68-100.eu", "--dry-run")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("Dry run"))
	assert.Cmp(out, td.Contains("ns3118333.ip-51-68-100.eu"))
	for call := range httpmock.GetCallCountInfo() {
		assert.Cmp(call, td.Not(td.HasPrefix("POST")))
	}
}

// Parking an IP that serves nothing is not a failure and not a request: it is
// already in the state being asked for.
func (ms *MockSuite) TestIpParkSaysWhenThereIsNothingToPark(assert, require *td.T) {
	httpmock.RegisterResponder("GET",
		"https://eu.api.ovh.com/v1/ip/135.125.71.80%2F32",
		httpmock.NewStringResponder(200, `{"ip": "135.125.71.80/32", "routedTo": {"serviceName": ""}}`))

	out, err := cmd.Execute("ip", "park", "135.125.71.80/32", "--yes")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("nothing to park"))
	for call := range httpmock.GetCallCountInfo() {
		assert.Cmp(call, td.Not(td.HasPrefix("POST")))
	}
}
