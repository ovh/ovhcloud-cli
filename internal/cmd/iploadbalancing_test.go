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

func (ms *MockSuite) TestIpLoadbalancingSSLListCmd(assert, require *td.T) {
	httpmock.RegisterResponder("GET",
		"https://eu.api.ovh.com/v1/ipLoadbalancing/loadbalancer-12345/ssl",
		httpmock.NewStringResponder(200, `[1, 2, 3]`).Once())

	out, err := cmd.Execute("iploadbalancing", "ssl", "list", "loadbalancer-12345", "-o", "json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`[{"id": 1}, {"id": 2}, {"id": 3}]`))
}

func (ms *MockSuite) TestIpLoadbalancingSSLListCmdEmpty(assert, require *td.T) {
	httpmock.RegisterResponder("GET",
		"https://eu.api.ovh.com/v1/ipLoadbalancing/loadbalancer-12345/ssl",
		httpmock.NewStringResponder(200, `[]`).Once())

	out, err := cmd.Execute("iploadbalancing", "ssl", "list", "loadbalancer-12345", "-o", "json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`[]`))
}

func (ms *MockSuite) TestIpLoadbalancingSSLListCmdAPIError(assert, require *td.T) {
	httpmock.RegisterResponder("GET",
		"https://eu.api.ovh.com/v1/ipLoadbalancing/loadbalancer-12345/ssl",
		httpmock.NewStringResponder(404, `{"message": "Service not found"}`).Once())

	_, err := cmd.Execute("iploadbalancing", "ssl", "list", "loadbalancer-12345")

	assert.String(err, `🛑 failed to fetch SSL certificates: OVHcloud API error (status code 404): "Service not found"`)
}

func (ms *MockSuite) TestIpLoadbalancingSSLListCmdMissingArg(assert, require *td.T) {
	_, err := cmd.Execute("iploadbalancing", "ssl", "list")

	assert.String(err, "accepts 1 arg(s), received 0")
}

func (ms *MockSuite) TestIpLoadbalancingSSLGetCmd(assert, require *td.T) {
	httpmock.RegisterResponder("GET",
		"https://eu.api.ovh.com/v1/ipLoadbalancing/loadbalancer-12345/ssl/42",
		httpmock.NewStringResponder(200, `{
			"id": 42,
			"displayName": "my-cert",
			"type": "custom",
			"subject": "CN=example.com",
			"san": ["example.com", "www.example.com"],
			"fingerprint": "AA:BB:CC",
			"serial": "0123",
			"expireDate": "2027-01-01T00:00:00+00:00"
		}`).Once())

	out, err := cmd.Execute("iploadbalancing", "ssl", "get", "loadbalancer-12345", "42", "-o", "json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{
		"id": 42,
		"displayName": "my-cert",
		"type": "custom",
		"subject": "CN=example.com",
		"san": ["example.com", "www.example.com"],
		"fingerprint": "AA:BB:CC",
		"serial": "0123",
		"expireDate": "2027-01-01T00:00:00+00:00"
	}`))
}

func (ms *MockSuite) TestIpLoadbalancingSSLGetCmdTemplate(assert, require *td.T) {
	httpmock.RegisterResponder("GET",
		"https://eu.api.ovh.com/v1/ipLoadbalancing/loadbalancer-12345/ssl/42",
		httpmock.NewStringResponder(200, `{
			"id": 42,
			"displayName": "my-cert",
			"type": "custom",
			"subject": "CN=example.com",
			"san": ["example.com", "www.example.com"],
			"fingerprint": "AA:BB:CC",
			"serial": "0123",
			"expireDate": "2027-01-01T00:00:00+00:00"
		}`).Once())

	out, err := cmd.Execute("iploadbalancing", "ssl", "get", "loadbalancer-12345", "42")

	require.CmpNoError(err)
	assert.Cmp(out, td.All(
		td.Contains("SSL certificate 42"),
		td.Contains("CN=example.com"),
		td.Contains("www.example.com"),
		td.Contains("AA:BB:CC"),
	))
}

func (ms *MockSuite) TestIpLoadbalancingSSLGetCmdInvalidID(assert, require *td.T) {
	_, err := cmd.Execute("iploadbalancing", "ssl", "get", "loadbalancer-12345", "abc")

	assert.String(err, "🛑 invalid SSL ID given, expected a number")
}

func (ms *MockSuite) TestIpLoadbalancingSSLGetCmdAPIError(assert, require *td.T) {
	httpmock.RegisterResponder("GET",
		"https://eu.api.ovh.com/v1/ipLoadbalancing/loadbalancer-12345/ssl/42",
		httpmock.NewStringResponder(404, `{"message": "SSL certificate not found"}`).Once())

	_, err := cmd.Execute("iploadbalancing", "ssl", "get", "loadbalancer-12345", "42")

	assert.Cmp(err, td.Smuggle(func(e error) string { return e.Error() },
		td.HasPrefix("🛑 error fetching /v1/ipLoadbalancing/loadbalancer-12345/ssl/42:")))
}

func (ms *MockSuite) TestIpLoadbalancingSSLGetCmdMissingArg(assert, require *td.T) {
	_, err := cmd.Execute("iploadbalancing", "ssl", "get", "loadbalancer-12345")

	assert.String(err, "accepts 2 arg(s), received 1")
}

func (ms *MockSuite) TestIpLoadbalancingSSLCreateCmd(assert, require *td.T) {
	httpmock.RegisterMatcherResponder("POST",
		"https://eu.api.ovh.com/v1/ipLoadbalancing/loadbalancer-12345/ssl",
		tdhttpmock.JSONBody(td.JSON(`{
			"certificate": "CERT",
			"chain": "CHAIN",
			"displayName": "my-cert",
			"key": "KEY"
		}`)),
		httpmock.NewStringResponder(200, `{"id": 42, "displayName": "my-cert", "type": "custom"}`),
	)

	out, err := cmd.Execute("iploadbalancing", "ssl", "create", "loadbalancer-12345",
		"--certificate", "CERT", "--chain", "CHAIN", "--display-name", "my-cert", "--key", "KEY")

	require.CmpNoError(err)
	assert.String(out, "✅ SSL certificate 42 created successfully")
}

func (ms *MockSuite) TestIpLoadbalancingSSLCreateCmdFromFile(assert, require *td.T) {
	tmpFile, err := os.CreateTemp("", "iploadbalancing-ssl-*.json")
	require.CmpNoError(err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(`{
		"certificate": "-----BEGIN CERTIFICATE-----\nCERT\n-----END CERTIFICATE-----\n",
		"displayName": "from-file",
		"key": "-----BEGIN PRIVATE KEY-----\nKEY\n-----END PRIVATE KEY-----\n"
	}`)
	require.CmpNoError(err)
	tmpFile.Close()

	httpmock.RegisterMatcherResponder("POST",
		"https://eu.api.ovh.com/v1/ipLoadbalancing/loadbalancer-12345/ssl",
		tdhttpmock.JSONBody(td.JSON(`{
			"certificate": "-----BEGIN CERTIFICATE-----\nCERT\n-----END CERTIFICATE-----\n",
			"displayName": "overridden",
			"key": "-----BEGIN PRIVATE KEY-----\nKEY\n-----END PRIVATE KEY-----\n"
		}`)),
		httpmock.NewStringResponder(200, `{"id": 43, "displayName": "overridden", "type": "custom"}`),
	)

	out, err := cmd.Execute("iploadbalancing", "ssl", "create", "loadbalancer-12345",
		"--from-file", tmpFile.Name(), "--display-name", "overridden")

	require.CmpNoError(err)
	assert.String(out, "✅ SSL certificate 43 created successfully")
}

func (ms *MockSuite) TestIpLoadbalancingSSLCreateCmdMissingKey(assert, require *td.T) {
	out, err := cmd.Execute("iploadbalancing", "ssl", "create", "loadbalancer-12345", "--certificate", "CERT")

	assert.CmpError(err)
	assert.Cmp(out, td.Contains(`failed to create SSL certificate: mandatory field "key" is missing in the parameters`))
}

func (ms *MockSuite) TestIpLoadbalancingSSLCreateCmdAPIError(assert, require *td.T) {
	httpmock.RegisterResponder("POST",
		"https://eu.api.ovh.com/v1/ipLoadbalancing/loadbalancer-12345/ssl",
		httpmock.NewStringResponder(400, `{"message": "Invalid certificate"}`),
	)

	out, err := cmd.Execute("iploadbalancing", "ssl", "create", "loadbalancer-12345",
		"--certificate", "CERT", "--key", "KEY")

	assert.CmpError(err)
	assert.Cmp(out, td.Contains("failed to create SSL certificate"))
	assert.Cmp(out, td.Contains("Invalid certificate"))
}

func (ms *MockSuite) TestIpLoadbalancingSSLCreateCmdMissingArg(assert, require *td.T) {
	_, err := cmd.Execute("iploadbalancing", "ssl", "create", "--certificate", "CERT", "--key", "KEY")

	assert.String(err, "accepts 1 arg(s), received 0")
}

func (ms *MockSuite) TestIpLoadbalancingPendingChangesCmd(assert, require *td.T) {
	httpmock.RegisterResponder("GET",
		"https://eu.api.ovh.com/v1/ipLoadbalancing/loadbalancer-12345/pendingChanges",
		httpmock.NewStringResponder(200, `[
			{"zone": "gra", "number": 2},
			{"zone": "rbx", "number": 0}
		]`).Once())

	out, err := cmd.Execute("iploadbalancing", "pending-changes", "loadbalancer-12345", "-o", "json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`[
		{"zone": "gra", "number": 2},
		{"zone": "rbx", "number": 0}
	]`))
}

func (ms *MockSuite) TestIpLoadbalancingPendingChangesCmdFilter(assert, require *td.T) {
	httpmock.RegisterResponder("GET",
		"https://eu.api.ovh.com/v1/ipLoadbalancing/loadbalancer-12345/pendingChanges",
		httpmock.NewStringResponder(200, `[
			{"zone": "gra", "number": 2},
			{"zone": "rbx", "number": 0}
		]`).Once())

	out, err := cmd.Execute("iploadbalancing", "pending-changes", "loadbalancer-12345", "--filter", "number > 0", "-o", "json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`[{"zone": "gra", "number": 2}]`))
}

func (ms *MockSuite) TestIpLoadbalancingPendingChangesCmdAPIError(assert, require *td.T) {
	httpmock.RegisterResponder("GET",
		"https://eu.api.ovh.com/v1/ipLoadbalancing/loadbalancer-12345/pendingChanges",
		httpmock.NewStringResponder(404, `{"message": "Service not found"}`).Once())

	out, err := cmd.Execute("iploadbalancing", "pending-changes", "loadbalancer-12345")

	assert.CmpError(err)
	assert.Cmp(out, td.Contains("failed to fetch results"))
}

func (ms *MockSuite) TestIpLoadbalancingPendingChangesCmdMissingArg(assert, require *td.T) {
	_, err := cmd.Execute("iploadbalancing", "pending-changes")

	assert.String(err, "accepts 1 arg(s), received 0")
}

func (ms *MockSuite) TestIpLoadbalancingRefreshCmd(assert, require *td.T) {
	httpmock.RegisterMatcherResponder("POST",
		"https://eu.api.ovh.com/v1/ipLoadbalancing/loadbalancer-12345/refresh",
		tdhttpmock.JSONBody(td.JSON(`{}`)),
		httpmock.NewStringResponder(200, `{"id": 789, "action": "refreshIplb", "status": "todo", "zones": ["gra", "rbx"]}`),
	)

	out, err := cmd.Execute("iploadbalancing", "refresh", "loadbalancer-12345")

	require.CmpNoError(err)
	assert.String(out, "✅ Configuration refresh started for IpLoadbalancing loadbalancer-12345 (task ID: 789)")
}

func (ms *MockSuite) TestIpLoadbalancingRefreshCmdZone(assert, require *td.T) {
	httpmock.RegisterMatcherResponder("POST",
		"https://eu.api.ovh.com/v1/ipLoadbalancing/loadbalancer-12345/refresh",
		tdhttpmock.JSONBody(td.JSON(`{"zone": "gra"}`)),
		httpmock.NewStringResponder(200, `{"id": 790, "action": "refreshIplb", "status": "todo", "zones": ["gra"]}`),
	)

	out, err := cmd.Execute("iploadbalancing", "refresh", "loadbalancer-12345", "--zone", "gra", "-o", "json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{
		"message": "✅ Configuration refresh started for IpLoadbalancing loadbalancer-12345 (task ID: 790)",
		"details": {"id": 790, "action": "refreshIplb", "status": "todo", "zones": ["gra"]}
	}`))
}

func (ms *MockSuite) TestIpLoadbalancingRefreshCmdAPIError(assert, require *td.T) {
	httpmock.RegisterResponder("POST",
		"https://eu.api.ovh.com/v1/ipLoadbalancing/loadbalancer-12345/refresh",
		httpmock.NewStringResponder(400, `{"message": "No pending changes"}`),
	)

	out, err := cmd.Execute("iploadbalancing", "refresh", "loadbalancer-12345")

	assert.CmpError(err)
	assert.Cmp(out, td.Contains("failed to refresh IpLoadbalancing loadbalancer-12345"))
	assert.Cmp(out, td.Contains("No pending changes"))
}

func (ms *MockSuite) TestIpLoadbalancingRefreshCmdMissingArg(assert, require *td.T) {
	_, err := cmd.Execute("iploadbalancing", "refresh")

	assert.String(err, "accepts 1 arg(s), received 0")
}

func (ms *MockSuite) TestIpLoadbalancingSSLDeleteCmd(assert, require *td.T) {
	httpmock.RegisterResponder("DELETE",
		"https://eu.api.ovh.com/v1/ipLoadbalancing/loadbalancer-12345/ssl/42",
		httpmock.NewStringResponder(200, ``).Once())

	out, err := cmd.Execute("iploadbalancing", "ssl", "delete", "loadbalancer-12345", "42")

	require.CmpNoError(err)
	assert.String(out, "✅ SSL certificate 42 deleted successfully")
}

func (ms *MockSuite) TestIpLoadbalancingSSLDeleteCmdInvalidID(assert, require *td.T) {
	_, err := cmd.Execute("iploadbalancing", "ssl", "delete", "loadbalancer-12345", "abc")

	assert.String(err, "🛑 invalid SSL ID given, expected a number")
}

func (ms *MockSuite) TestIpLoadbalancingSSLDeleteCmdAPIError(assert, require *td.T) {
	httpmock.RegisterResponder("DELETE",
		"https://eu.api.ovh.com/v1/ipLoadbalancing/loadbalancer-12345/ssl/42",
		httpmock.NewStringResponder(404, `{"message": "SSL certificate not found"}`).Once())

	_, err := cmd.Execute("iploadbalancing", "ssl", "delete", "loadbalancer-12345", "42")

	assert.String(err, `🛑 failed to delete SSL certificate 42: OVHcloud API error (status code 404): "SSL certificate not found"`)
}

func (ms *MockSuite) TestIpLoadbalancingSSLDeleteCmdMissingArg(assert, require *td.T) {
	_, err := cmd.Execute("iploadbalancing", "ssl", "delete", "loadbalancer-12345")

	assert.String(err, "accepts 2 arg(s), received 1")
}
