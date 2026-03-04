// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"encoding/json"
	"os"

	"github.com/jarcoal/httpmock"
	"github.com/maxatome/go-testdeep/td"
	"github.com/maxatome/tdhttpmock"
	"github.com/ovh/ovhcloud-cli/internal/cmd"
)

const (
	testFirewallIPBlock = "198.51.100.42/32"
	testFirewallIP      = "198.51.100.42"
)

func (ms *MockSuite) TestIpFirewallListCmd(assert, require *td.T) {
	httpmock.RegisterResponder("GET",
		"https://eu.api.ovh.com/v1/ip/198.51.100.42%2F32/firewall",
		httpmock.NewStringResponder(200, `["`+testFirewallIP+`"]`).Once())

	httpmock.RegisterResponder("GET",
		"https://eu.api.ovh.com/v1/ip/198.51.100.42%2F32/firewall/198.51.100.42",
		httpmock.NewStringResponder(200, `{
			"ipOnFirewall": "`+testFirewallIP+`",
			"enabled": true,
			"state": "ok"
		}`).Once())

	out, err := cmd.Execute("ip", "firewall", "list", testFirewallIPBlock, "-o", "json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`[
		{
			"ipOnFirewall": "`+testFirewallIP+`",
			"enabled": true,
			"state": "ok"
		}
	]`))
}

func (ms *MockSuite) TestIpFirewallAddCmd(assert, require *td.T) {
	httpmock.RegisterMatcherResponder("POST",
		"https://eu.api.ovh.com/v1/ip/198.51.100.42%2F32/firewall",
		tdhttpmock.JSONBody(td.JSON(`{"ipOnFirewall": "`+testFirewallIP+`"}`)),
		httpmock.NewStringResponder(200, `{
			"ipOnFirewall": "`+testFirewallIP+`",
			"enabled": false,
			"state": "ok"
		}`),
	)

	out, err := cmd.Execute("ip", "firewall", "add", testFirewallIPBlock, testFirewallIP)

	require.CmpNoError(err)
	assert.String(out, "✅ IP "+testFirewallIP+" successfully added to firewall")
}

func (ms *MockSuite) TestIpFirewallGetCmd(assert, require *td.T) {
	httpmock.RegisterResponder("GET",
		"https://eu.api.ovh.com/v1/ip/198.51.100.42%2F32/firewall/198.51.100.42",
		httpmock.NewStringResponder(200, `{
			"ipOnFirewall": "`+testFirewallIP+`",
			"enabled": true,
			"state": "ok"
		}`).Once())

	out, err := cmd.Execute("ip", "firewall", "get", testFirewallIPBlock, testFirewallIP, "-o", "json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{
		"ipOnFirewall": "`+testFirewallIP+`",
		"enabled": true,
		"state": "ok"
	}`))
}

func (ms *MockSuite) TestIpFirewallEnableCmd(assert, require *td.T) {
	httpmock.RegisterMatcherResponder("PUT",
		"https://eu.api.ovh.com/v1/ip/198.51.100.42%2F32/firewall/198.51.100.42",
		tdhttpmock.JSONBody(td.JSON(`{"enabled": true}`)),
		httpmock.NewStringResponder(200, ``),
	)

	out, err := cmd.Execute("ip", "firewall", "enable", testFirewallIPBlock, testFirewallIP)

	require.CmpNoError(err)
	assert.String(out, "✅ Firewall successfully enabled for "+testFirewallIP)
}

func (ms *MockSuite) TestIpFirewallDisableCmd(assert, require *td.T) {
	httpmock.RegisterMatcherResponder("PUT",
		"https://eu.api.ovh.com/v1/ip/198.51.100.42%2F32/firewall/198.51.100.42",
		tdhttpmock.JSONBody(td.JSON(`{"enabled": false}`)),
		httpmock.NewStringResponder(200, ``),
	)

	out, err := cmd.Execute("ip", "firewall", "disable", testFirewallIPBlock, testFirewallIP)

	require.CmpNoError(err)
	assert.String(out, "✅ Firewall successfully disabled for "+testFirewallIP)
}

func (ms *MockSuite) TestIpFirewallDeleteCmd(assert, require *td.T) {
	httpmock.RegisterResponder("DELETE",
		"https://eu.api.ovh.com/v1/ip/198.51.100.42%2F32/firewall/198.51.100.42",
		httpmock.NewStringResponder(200, ``),
	)

	out, err := cmd.Execute("ip", "firewall", "delete", testFirewallIPBlock, testFirewallIP)

	require.CmpNoError(err)
	assert.String(out, "✅ Firewall and all rules successfully removed for "+testFirewallIP)
}

func (ms *MockSuite) TestIpFirewallRuleListCmd(assert, require *td.T) {
	httpmock.RegisterResponder("GET",
		"https://eu.api.ovh.com/v1/ip/198.51.100.42%2F32/firewall/198.51.100.42/rule",
		httpmock.NewStringResponder(200, `[5, 19]`).Once())

	httpmock.RegisterResponder("GET",
		"https://eu.api.ovh.com/v1/ip/198.51.100.42%2F32/firewall/198.51.100.42/rule/5",
		httpmock.NewStringResponder(200, `{
			"sequence": 5,
			"action": "permit",
			"protocol": "tcp",
			"source": "192.0.2.1/32",
			"destination": "`+testFirewallIPBlock+`",
			"destinationPort": "eq 22",
			"sourcePort": null,
			"rule": "permit tcp 192.0.2.1/32 `+testFirewallIPBlock+` eq 22",
			"state": "ok",
			"creationDate": "2026-03-03T19:43:08+01:00",
			"tcpOption": null,
			"fragments": false,
			"l3PacketLength": null
		}`).Once())

	httpmock.RegisterResponder("GET",
		"https://eu.api.ovh.com/v1/ip/198.51.100.42%2F32/firewall/198.51.100.42/rule/19",
		httpmock.NewStringResponder(200, `{
			"sequence": 19,
			"action": "deny",
			"protocol": "tcp",
			"source": "any",
			"destination": "`+testFirewallIPBlock+`",
			"destinationPort": "eq 22",
			"sourcePort": null,
			"rule": "deny tcp any `+testFirewallIPBlock+` eq 22",
			"state": "ok",
			"creationDate": "2025-08-11T13:16:53+02:00",
			"tcpOption": null,
			"fragments": false,
			"l3PacketLength": null
		}`).Once())

	out, err := cmd.Execute("ip", "firewall", "rule", "list",
		testFirewallIPBlock, testFirewallIP, "-o", "json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`[
		{
			"sequence": 5,
			"action": "permit",
			"protocol": "tcp",
			"source": "192.0.2.1/32",
			"destination": "`+testFirewallIPBlock+`",
			"destinationPort": "eq 22",
			"sourcePort": null,
			"rule": "permit tcp 192.0.2.1/32 `+testFirewallIPBlock+` eq 22",
			"state": "ok",
			"creationDate": "2026-03-03T19:43:08+01:00",
			"tcpOption": null,
			"fragments": false,
			"l3PacketLength": null
		},
		{
			"sequence": 19,
			"action": "deny",
			"protocol": "tcp",
			"source": "any",
			"destination": "`+testFirewallIPBlock+`",
			"destinationPort": "eq 22",
			"sourcePort": null,
			"rule": "deny tcp any `+testFirewallIPBlock+` eq 22",
			"state": "ok",
			"creationDate": "2025-08-11T13:16:53+02:00",
			"tcpOption": null,
			"fragments": false,
			"l3PacketLength": null
		}
	]`))
}

func (ms *MockSuite) TestIpFirewallRuleGetCmd(assert, require *td.T) {
	httpmock.RegisterResponder("GET",
		"https://eu.api.ovh.com/v1/ip/198.51.100.42%2F32/firewall/198.51.100.42/rule/5",
		httpmock.NewStringResponder(200, `{
			"sequence": 5,
			"action": "permit",
			"protocol": "tcp",
			"source": "192.0.2.1/32",
			"destination": "`+testFirewallIPBlock+`",
			"destinationPort": "eq 22",
			"sourcePort": null,
			"rule": "permit tcp 192.0.2.1/32 `+testFirewallIPBlock+` eq 22",
			"state": "ok",
			"creationDate": "2026-03-03T19:43:08+01:00",
			"tcpOption": null,
			"fragments": false,
			"l3PacketLength": null
		}`).Once())

	out, err := cmd.Execute("ip", "firewall", "rule", "get",
		testFirewallIPBlock, testFirewallIP, "5", "-o", "json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{
		"sequence": 5,
		"action": "permit",
		"protocol": "tcp",
		"source": "192.0.2.1/32",
		"destination": "`+testFirewallIPBlock+`",
		"destinationPort": "eq 22",
		"sourcePort": null,
		"rule": "permit tcp 192.0.2.1/32 `+testFirewallIPBlock+` eq 22",
		"state": "ok",
		"creationDate": "2026-03-03T19:43:08+01:00",
		"tcpOption": null,
		"fragments": false,
		"l3PacketLength": null
	}`))
}

func (ms *MockSuite) TestIpFirewallRuleCreateCmd(assert, require *td.T) {
	httpmock.RegisterMatcherResponder("POST",
		"https://eu.api.ovh.com/v1/ip/198.51.100.42%2F32/firewall/198.51.100.42/rule",
		tdhttpmock.JSONBody(td.JSON(`{
			"action": "permit",
			"protocol": "tcp",
			"sequence": 0,
			"source": "10.0.0.1/32",
			"destinationPort": 443
		}`)),
		httpmock.NewStringResponder(200, `{
			"sequence": 0,
			"action": "permit",
			"protocol": "tcp",
			"source": "10.0.0.1/32",
			"destination": "`+testFirewallIPBlock+`",
			"destinationPort": "eq 443",
			"rule": "permit tcp 10.0.0.1/32 `+testFirewallIPBlock+` eq 443",
			"state": "creationPending"
		}`),
	)

	out, err := cmd.Execute("ip", "firewall", "rule", "create",
		testFirewallIPBlock, testFirewallIP,
		"--action", "permit",
		"--protocol", "tcp",
		"--sequence", "0",
		"--source", "10.0.0.1/32",
		"--destination-port", "443",
	)

	require.CmpNoError(err)
	assert.String(out, "✅ Rule #0 successfully created")
}

func (ms *MockSuite) TestIpFirewallRuleCreateFromFile(assert, require *td.T) {
	// Create a temp JSON file with rule parameters
	tmpFile, err := os.CreateTemp("", "firewall-rule-*.json")
	require.CmpNoError(err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(`{
		"action": "permit",
		"protocol": "tcp",
		"sequence": 0,
		"source": "10.0.0.1/32",
		"destinationPort": 443
	}`)
	require.CmpNoError(err)
	tmpFile.Close()

	httpmock.RegisterMatcherResponder("POST",
		"https://eu.api.ovh.com/v1/ip/198.51.100.42%2F32/firewall/198.51.100.42/rule",
		tdhttpmock.JSONBody(td.JSON(`{
			"action": "permit",
			"protocol": "tcp",
			"sequence": 0,
			"source": "10.0.0.1/32",
			"destinationPort": 443
		}`)),
		httpmock.NewStringResponder(200, `{
			"sequence": 0,
			"action": "permit",
			"protocol": "tcp",
			"source": "10.0.0.1/32",
			"destination": "`+testFirewallIPBlock+`",
			"destinationPort": "eq 443",
			"rule": "permit tcp 10.0.0.1/32 `+testFirewallIPBlock+` eq 443",
			"state": "creationPending"
		}`),
	)

	out, err := cmd.Execute("ip", "firewall", "rule", "create",
		testFirewallIPBlock, testFirewallIP,
		"--from-file", tmpFile.Name(),
	)

	require.CmpNoError(err)
	assert.String(out, "✅ Rule #0 successfully created")
}

func (ms *MockSuite) TestIpFirewallRuleDeleteCmd(assert, require *td.T) {
	httpmock.RegisterResponder("DELETE",
		"https://eu.api.ovh.com/v1/ip/198.51.100.42%2F32/firewall/198.51.100.42/rule/5",
		httpmock.NewStringResponder(200, ``),
	)

	out, err := cmd.Execute("ip", "firewall", "rule", "delete",
		testFirewallIPBlock, testFirewallIP, "5")

	require.CmpNoError(err)
	assert.String(out, "✅ Rule #5 successfully deleted")
}

func (ms *MockSuite) TestIpFirewallRuleCreatePortExcl(assert, require *td.T) {
	_, err := cmd.Execute("ip", "firewall", "rule", "create",
		testFirewallIPBlock, testFirewallIP,
		"--action", "permit",
		"--protocol", "tcp",
		"--sequence", "0",
		"--destination-port", "443",
		"--destination-port-from", "80",
	)

	require.CmpError(err)
	assert.Contains(err.Error(), "destination-port")
}

func (ms *MockSuite) TestIpFirewallRuleCreateTcpOnly(assert, require *td.T) {
	_, err := cmd.Execute("ip", "firewall", "rule", "create",
		testFirewallIPBlock, testFirewallIP,
		"--action", "deny",
		"--protocol", "udp",
		"--sequence", "1",
		"--tcp-option", "established",
	)

	require.CmpError(err)
	assert.Contains(err.Error(), "tcp-option")
}
