// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"github.com/jarcoal/httpmock"
	"github.com/maxatome/go-testdeep/td"
	"github.com/ovh/ovhcloud-cli/internal/cmd"
)

const testBlock = "https://eu.api.ovh.com/v1/ip/192.0.2.0%2F24"

// registerEmptyBlockLists answers the three block families with nothing, which
// is what all 537 blocks of a real account answered when this lot was measured.
func registerEmptyBlockLists() {
	for _, mechanism := range []string{"antihack", "arp", "spam"} {
		httpmock.RegisterResponder("GET", testBlock+"/"+mechanism,
			httpmock.NewStringResponder(200, `[]`))
	}
}

// "Nothing is blocked" is the answer this command exists to give, so it must
// read as an answer and not as an empty table.
func (ms *MockSuite) TestIpBlockedSaysNothingIsBlocked(assert, require *td.T) {
	registerEmptyBlockLists()

	out, err := cmd.Execute("ip", "blocked", "192.0.2.0/24")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("No address"))
}

// The three mechanisms are read together because the operator hit by one of
// them has no way to know which.
func (ms *MockSuite) TestIpBlockedReadsTheThreeMechanisms(assert, require *td.T) {
	httpmock.RegisterResponder("GET", testBlock+"/antihack",
		httpmock.NewStringResponder(200, `["192.0.2.7"]`))
	httpmock.RegisterResponder("GET", testBlock+"/antihack/192.0.2.7",
		httpmock.NewStringResponder(200, `{"ipBlocked":"192.0.2.7","state":"blocked","blockedSince":"2026-08-19T10:00:00+02:00","time":720,"logs":"brute force on ssh"}`))
	httpmock.RegisterResponder("GET", testBlock+"/arp",
		httpmock.NewStringResponder(200, `[]`))
	httpmock.RegisterResponder("GET", testBlock+"/spam",
		httpmock.NewStringResponder(200, `["192.0.2.9"]`))
	httpmock.RegisterResponder("GET", testBlock+"/spam/192.0.2.9",
		httpmock.NewStringResponder(200, `{"ipSpamming":"192.0.2.9","state":"blockedForSpam","date":"2026-08-18T09:00:00+02:00","time":86400}`))

	out, err := cmd.Execute("ip", "blocked", "192.0.2.0/24")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("192.0.2.7"))
	assert.Cmp(out, td.Contains("192.0.2.9"))
	assert.Cmp(out, td.Contains("antihack"))
	assert.Cmp(out, td.Contains("spam"))
	// The same `time` field means a cooldown on one row and a sentence length
	// on the other; the note has to say which.
	assert.Cmp(out, td.Contains("unblocked in 12m"))
	assert.Cmp(out, td.Contains("blocked for 1d"))
}

// An address nothing holds is not released: the API's own refusal reads as
// "this address does not exist", which is not what happened.
func (ms *MockSuite) TestIpUnblockRefusesWhenNothingBlocksTheAddress(assert, require *td.T) {
	registerEmptyBlockLists()

	_, err := cmd.Execute("ip", "unblock", "192.0.2.0/24", "192.0.2.7", "--yes")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("not blocked"))
	for call := range httpmock.GetCallCountInfo() {
		assert.Cmp(call, td.Not(td.HasPrefix("POST")), "nothing may be sent for a refused unblock")
	}
}

// The API says how long is left before it accepts a release. Sending the
// request anyway costs a 404 saying the address does not exist.
func (ms *MockSuite) TestIpUnblockWaitsOutTheCooldownBeforeSending(assert, require *td.T) {
	httpmock.RegisterResponder("GET", testBlock+"/antihack",
		httpmock.NewStringResponder(200, `["192.0.2.7"]`))
	httpmock.RegisterResponder("GET", testBlock+"/antihack/192.0.2.7",
		httpmock.NewStringResponder(200, `{"ipBlocked":"192.0.2.7","state":"blocked","time":900}`))
	httpmock.RegisterResponder("GET", testBlock+"/arp", httpmock.NewStringResponder(200, `[]`))
	httpmock.RegisterResponder("GET", testBlock+"/spam", httpmock.NewStringResponder(200, `[]`))

	_, err := cmd.Execute("ip", "unblock", "192.0.2.0/24", "192.0.2.7", "--yes")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("15m"))
	for call := range httpmock.GetCallCountInfo() {
		assert.Cmp(call, td.Not(td.HasPrefix("POST")))
	}
}

// --reason goes straight to one mechanism instead of reading the three.
func (ms *MockSuite) TestIpUnblockWithReasonSkipsTheLookup(assert, require *td.T) {
	httpmock.RegisterResponder("GET", testBlock+"/arp/192.0.2.7",
		httpmock.NewStringResponder(200, `{"ipBlocked":"192.0.2.7","state":"blocked","time":0}`))

	out, err := cmd.Execute("ip", "unblock", "192.0.2.0/24", "192.0.2.7",
		"--reason", "arp", "--dry-run")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("/arp/192.0.2.7/unblock"))
	for call := range httpmock.GetCallCountInfo() {
		assert.Cmp(call, td.Not(td.Contains("/antihack")), "--reason must not read the other families")
		assert.Cmp(call, td.Not(td.Contains("/spam")))
	}
}

func (ms *MockSuite) TestIpUnblockRefusesAnUnknownReason(assert, require *td.T) {
	_, err := cmd.Execute("ip", "unblock", "192.0.2.0/24", "192.0.2.7", "--reason", "ddos", "--yes")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("antihack"))
	assert.Cmp(err.Error(), td.Contains("arp"))
	assert.Cmp(err.Error(), td.Contains("spam"))
}

// This route answers 500 for every address hosted outside Europe — 52 of 537
// blocks when measured. An empty table there would say "no phishing reported"
// about a block nobody managed to read.
func (ms *MockSuite) TestIpPhishingReportsAServerErrorAsAnError(assert, require *td.T) {
	httpmock.RegisterResponder("GET", testBlock+"/phishing",
		httpmock.NewStringResponder(500, `{"class":"Server::InternalServerError","message":"Internal server error"}`))

	_, err := cmd.Execute("ip", "phishing", "list", "192.0.2.0/24")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("failed to read the phishing entries"))
}

// Enabling UDP firewall mode on an address with no rule drops every UDP packet
// sent to it. One address of the account measured was exactly one flag away.
func (ms *MockSuite) TestIpGameEditRefusesToBlackholeUdp(assert, require *td.T) {
	httpmock.RegisterResponder("GET", testBlock+"/game/192.0.2.7/rule",
		httpmock.NewStringResponder(200, `[]`))

	_, err := cmd.Execute("ip", "game", "edit", "192.0.2.0/24", "192.0.2.7",
		"--firewall-mode", "--yes")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("every UDP packet"))
	assert.Cmp(err.Error(), td.Contains("ovhcloud ip game rule add"))
	for call := range httpmock.GetCallCountInfo() {
		assert.Cmp(call, td.Not(td.HasPrefix("PUT")), "nothing may be written for a refused change")
	}
}

// The protocols come from the address, not from the enum shipped with the CLI:
// the two disagree in production.
func (ms *MockSuite) TestIpGameRuleAddChecksTheProtocolAgainstTheAddress(assert, require *td.T) {
	httpmock.RegisterResponder("GET", testBlock+"/game/192.0.2.7",
		httpmock.NewStringResponder(200, `{"ipOnGame":"192.0.2.7","state":"ok","maxRules":30,"firewallModeEnabled":true,"supportedProtocols":["arma","hl2Source","teamspeak3"]}`))

	_, err := cmd.Execute("ip", "game", "rule", "add", "192.0.2.0/24", "192.0.2.7",
		"--protocol", "valheim", "--ports", "2456", "--yes")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("hl2Source"), "the refusal lists what this address supports")
	assert.Cmp(err.Error(), td.Not(td.Contains("minecraftJava")), "and nothing it does not")
	for call := range httpmock.GetCallCountInfo() {
		assert.Cmp(call, td.Not(td.HasPrefix("POST")))
	}
}

// maxRules is per address — 30 on one measured, 100 on another. Sending the
// rule anyway costs a round trip to be told the limit the CLI had already read.
func (ms *MockSuite) TestIpGameRuleAddStopsAtTheAddressLimit(assert, require *td.T) {
	httpmock.RegisterResponder("GET", testBlock+"/game/192.0.2.7",
		httpmock.NewStringResponder(200, `{"ipOnGame":"192.0.2.7","maxRules":2,"supportedProtocols":["arma"]}`))
	httpmock.RegisterResponder("GET", testBlock+"/game/192.0.2.7/rule",
		httpmock.NewStringResponder(200, `[1, 2]`))

	_, err := cmd.Execute("ip", "game", "rule", "add", "192.0.2.0/24", "192.0.2.7",
		"--protocol", "arma", "--ports", "2302", "--yes")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("2 rules"))
	assert.Cmp(err.Error(), td.Contains("ovhcloud ip game rule delete"))
	for call := range httpmock.GetCallCountInfo() {
		assert.Cmp(call, td.Not(td.HasPrefix("POST")))
	}
}

// The profile is created or changed depending on whether one exists, because
// an operator setting a delay does not know which of the two API routes applies.
func (ms *MockSuite) TestIpMitigationProfileSetCreatesWhenThereIsNone(assert, require *td.T) {
	httpmock.RegisterResponder("GET", testBlock+"/mitigationProfiles/192.0.2.7",
		httpmock.NewStringResponder(404, `{"message":"The requested object (ipMitigationProfile = 192.0.2.7) does not exist"}`))

	out, err := cmd.Execute("ip", "mitigation-profile", "set", "192.0.2.0/24", "192.0.2.7",
		"--timeout", "360", "--dry-run")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("POST"))
	assert.Cmp(out, td.Contains("/mitigationProfiles"))
}

func (ms *MockSuite) TestIpMitigationProfileSetUpdatesWhenOneExists(assert, require *td.T) {
	httpmock.RegisterResponder("GET", testBlock+"/mitigationProfiles/192.0.2.7",
		httpmock.NewStringResponder(200, `{"ipMitigationProfile":"192.0.2.7","autoMitigationTimeOut":15,"state":"ok"}`))

	out, err := cmd.Execute("ip", "mitigation-profile", "set", "192.0.2.0/24", "192.0.2.7",
		"--timeout", "360", "--dry-run")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("PUT"))
	assert.Cmp(out, td.Contains("/mitigationProfiles/192.0.2.7"))
}

func (ms *MockSuite) TestIpMitigationProfileSetRefusesAnUnacceptedDelay(assert, require *td.T) {
	_, err := cmd.Execute("ip", "mitigation-profile", "set", "192.0.2.0/24", "192.0.2.7",
		"--timeout", "30", "--yes")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("0, 15, 60, 360, 1560"))
	assert.Cmp(httpmock.GetTotalCallCount(), 0, "an impossible delay costs no request")
}
