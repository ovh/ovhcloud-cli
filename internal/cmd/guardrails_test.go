// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"slices"
	"strings"

	"github.com/jarcoal/httpmock"
	"github.com/maxatome/go-testdeep/td"
	"github.com/ovh/ovhcloud-cli/internal/cmd"
)

// These commands deleted or interrupted something without asking. They were the
// last destructive or disruptive verbs of the CLI with no confirmation, and each
// gets the same three checks as the commands guarded before them.
var guardedCommands = []struct {
	name   string
	args   []string
	method string
	url    string
}{
	{"ipmi reset-sessions", []string{"baremetal", "ipmi", "reset-sessions", "fakeBaremetal"},
		"POST", "https://eu.api.ovh.com/v1/dedicated/server/fakeBaremetal/features/ipmi/resetSessions"},
	{"iam policy delete", []string{"iam", "policy", "delete", "policy-1234"},
		"DELETE", "https://eu.api.ovh.com/v2/iam/policy/policy-1234"},
	{"iam user delete", []string{"iam", "user", "delete", "user1"},
		"DELETE", "https://eu.api.ovh.com/v1/me/identity/user/user1"},
	{"iam user token delete", []string{"iam", "user", "token", "delete", "user1", "token1"},
		"DELETE", "https://eu.api.ovh.com/v1/me/identity/user/user1/token/token1"},
	{"ip reverse delete", []string{"ip", "reverse", "delete", testFirewallIPBlock, testFirewallIP},
		"DELETE", "https://eu.api.ovh.com/v1/ip/198.51.100.42%2F32/reverse/198.51.100.42"},
	{"ip firewall delete", []string{"ip", "firewall", "delete", testFirewallIPBlock, testFirewallIP},
		"DELETE", "https://eu.api.ovh.com/v1/ip/198.51.100.42%2F32/firewall/198.51.100.42"},
	{"ip firewall rule delete", []string{"ip", "firewall", "rule", "delete", testFirewallIPBlock, testFirewallIP, "5"},
		"DELETE", "https://eu.api.ovh.com/v1/ip/198.51.100.42%2F32/firewall/198.51.100.42/rule/5"},
}

func registerGuardedCommands() {
	for _, c := range guardedCommands {
		httpmock.RegisterResponder(c.method, c.url, httpmock.NewStringResponder(200, `{}`))
	}
}

// An unattended run that did not say --yes has no terminal to answer on: it must
// refuse, and nothing may reach the API.
func (ms *MockSuite) TestGuardedCommandsRefuseWithoutConsent(assert, require *td.T) {
	registerGuardedCommands()

	for _, c := range guardedCommands {
		_, err := cmd.Execute(c.args...)
		cmd.PostExecute()

		if assert.CmpError(err, c.name) {
			assert.Cmp(err.Error(), td.Contains("cancelled"), c.name)
		}
	}

	assert.Cmp(httpmock.GetTotalCallCount(), 0, "nothing must reach the API without a confirmation")
}

// --yes is how a pipeline states its intent, and it must be enough.
func (ms *MockSuite) TestGuardedCommandsProceedWithYes(assert, require *td.T) {
	registerGuardedCommands()

	for _, c := range guardedCommands {
		_, err := cmd.Execute(slices.Concat(c.args, []string{"--yes"})...)
		cmd.PostExecute()

		assert.CmpNoError(err, c.name)
		assert.Cmp(httpmock.GetCallCountInfo()[c.method+" "+c.url], 1, c.name)
	}
}

// --dry-run names the call, says what the confirmation would have warned about,
// and sends nothing.
func (ms *MockSuite) TestGuardedCommandsDryRunSendsNothing(assert, require *td.T) {
	registerGuardedCommands()

	for _, c := range guardedCommands {
		out, err := cmd.Execute(slices.Concat(c.args, []string{"--dry-run"})...)
		cmd.PostExecute()

		assert.CmpNoError(err, c.name)
		assert.Cmp(out, td.Contains(c.method+" "+strings.TrimPrefix(c.url, "https://eu.api.ovh.com")), c.name)
		assert.Cmp(out, td.Contains("It would have asked first"), c.name)
	}

	assert.Cmp(httpmock.GetTotalCallCount(), 0, "a dry run must send nothing")
}
