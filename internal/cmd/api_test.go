// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"github.com/jarcoal/httpmock"
	"github.com/maxatome/go-testdeep/td"
	"github.com/ovh/ovhcloud-cli/internal/cmd"
)

// The requirement itself: an endpoint the CLI has no command for is reachable.
// /dedicated/server/*/vrack is one of the 107 paths of the dedicated-server API
// that no command covers.
func (ms *MockSuite) TestAPICallReachesAnUncoveredEndpoint(assert, require *td.T) {
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/dedicated/server/fakeBaremetal/vrack",
		httpmock.NewStringResponder(200, `{"vrack": "pn-12345"}`),
	)

	out, err := cmd.Execute("api", "call", "GET", "/dedicated/server/fakeBaremetal/vrack", "-o", "json")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("pn-12345"))
}

// A user reads the API documentation, which shows paths without their version.
func (ms *MockSuite) TestAPICallAcceptsAPathWithoutItsVersion(assert, require *td.T) {
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/dedicated/server",
		httpmock.NewStringResponder(200, `["ns3168421.ip-51-77-12.eu"]`),
	)

	out, err := cmd.Execute("api", "call", "GET", "dedicated/server", "-o", "json")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("ns3168421"))
}

// /v2 is a different API, not a typo to be corrected into /v1.
func (ms *MockSuite) TestAPICallLeavesAV2PathAlone(assert, require *td.T) {
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v2/publicCloud/project",
		httpmock.NewStringResponder(200, `[{"project_id": "abc"}]`),
	)

	out, err := cmd.Execute("api", "call", "GET", "/v2/publicCloud/project", "-o", "json")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("abc"))
	assert.Cmp(httpmock.GetCallCountInfo()["GET https://eu.api.ovh.com/v1/v2/publicCloud/project"], 0,
		"a v2 path must not be prefixed into a v1 one")
}

// This command has none of the guards the product commands carry, so --dry-run
// is the only thing standing between a typed path and a real write.
func (ms *MockSuite) TestAPICallDryRunSendsNothing(assert, require *td.T) {
	httpmock.RegisterResponder("DELETE", "https://eu.api.ovh.com/v1/dedicated/server/fakeBaremetal/secondaryDnsDomains/example.com",
		httpmock.NewStringResponder(200, `{}`),
	)

	out, err := cmd.Execute("api", "call", "DELETE",
		"/dedicated/server/fakeBaremetal/secondaryDnsDomains/example.com", "--dry-run")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("Dry run"))
	assert.Cmp(out, td.Contains("DELETE /v1/dedicated/server/fakeBaremetal/secondaryDnsDomains/example.com"),
		"the request is shown in full, so it can be checked before being sent")
	assert.Cmp(httpmock.GetTotalCallCount(), 0, "nothing must leave")
}

// An unsupported verb must be named, not sent and rejected by the API.
func (ms *MockSuite) TestAPICallRejectsAnUnsupportedMethod(assert, require *td.T) {
	_, err := cmd.Execute("api", "call", "PATCH", "/dedicated/server")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("unsupported method"))
	assert.Cmp(httpmock.GetTotalCallCount(), 0, "an unknown verb must not reach the API")
}

// The escape hatch used to live under `webhosting`; that spelling must keep
// working for whoever scripted it.
func (ms *MockSuite) TestWebhostingAPICallStillWorks(assert, require *td.T) {
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/hosting/web",
		httpmock.NewStringResponder(200, `["mysite.example"]`),
	)

	out, err := cmd.Execute("webhosting", "api", "call", "GET", "/hosting/web", "-o", "json")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("mysite.example"))
}
