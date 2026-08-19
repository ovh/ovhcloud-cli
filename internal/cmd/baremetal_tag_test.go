// SPDX-FileCopyrightText: 2026 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/jarcoal/httpmock"
	"github.com/maxatome/go-testdeep/td"
	"github.com/ovh/ovhcloud-cli/internal/cmd"
)

const (
	serversV1 = "https://eu.api.ovh.com/v1/dedicated/server"
	serversV2 = "https://eu.api.ovh.com/v2/dedicated/server"
)

// captureTagQuery answers the v2 collection and keeps the filter it was asked
// for, because the filter is the whole point of the flag.
func captureTagQuery(seen *map[string][]map[string]any) {
	httpmock.RegisterResponder(http.MethodGet, serversV2,
		func(req *http.Request) (*http.Response, error) {
			raw := req.URL.Query().Get("iamTags")
			if raw != "" {
				decoded, err := url.QueryUnescape(raw)
				if err != nil {
					return nil, err
				}
				if err := json.Unmarshal([]byte(decoded), seen); err != nil {
					return nil, err
				}
			}
			return httpmock.NewStringResponse(200, `[{"id":"ns1.example"}]`), nil
		})
	httpmock.RegisterResponder(http.MethodGet, serversV1+"/ns1.example",
		httpmock.NewStringResponder(200, `{"name":"ns1.example","datacenter":"rbx8","region":"eu-west-rbx","os":"debian12","state":"ok","iam":{"displayName":"Paperclip"}}`))
}

// The narrowing is asked of the API, in the shape the API takes.
func (ms *MockSuite) TestBaremetalListByTagAsksTheApiToNarrow(assert, require *td.T) {
	var seen map[string][]map[string]any
	captureTagQuery(&seen)

	out, err := cmd.Execute("baremetal", "list", "--tag", "Compliance=PCI-DSS")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("ns1.example"))
	assert.Cmp(seen["Compliance"][0]["operator"], "EQ")
	assert.Cmp(seen["Compliance"][0]["value"], "PCI-DSS")
	assert.Cmp(httpmock.GetCallCountInfo()["GET "+serversV1], 0,
		"the v1 collection is not listed when the API is doing the narrowing")
}

// The servers that come back are read on v1: the v2 object is {id, iam} and
// carries no machine, so a table built from it would lose every column.
func (ms *MockSuite) TestBaremetalListByTagReadsTheServersOnV1(assert, require *td.T) {
	var seen map[string][]map[string]any
	captureTagQuery(&seen)

	out, err := cmd.Execute("baremetal", "list", "--tag", "owner:EXISTS")

	require.CmpNoError(err)
	assert.Cmp(seen["owner"][0]["operator"], "EXISTS")
	assert.Cmp(seen["owner"][0], td.Not(td.ContainsKey("value")), "EXISTS carries no value")
	assert.Cmp(out, td.Contains("debian12"), "the operating system only exists on the v1 object")
	assert.Cmp(httpmock.GetCallCountInfo()["GET "+serversV1+"/ns1.example"], 1)
}

// Without the flag nothing changes: the same v1 collection as before, and the
// v2 one is never touched.
func (ms *MockSuite) TestBaremetalListWithoutTagDoesNotTouchV2(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, serversV1,
		httpmock.NewStringResponder(200, `["ns1.example"]`))
	httpmock.RegisterResponder(http.MethodGet, serversV1+"/ns1.example",
		httpmock.NewStringResponder(200, `{"name":"ns1.example","datacenter":"rbx8","region":"eu-west-rbx","os":"debian12","state":"ok","iam":{"displayName":"Paperclip"}}`))

	out, err := cmd.Execute("baremetal", "list")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("ns1.example"))
	assert.Cmp(httpmock.GetCallCountInfo()["GET "+serversV2], 0)
}

// --tag runs on the API before the servers are read, --filter runs on the table
// after. Both are honoured, and they are not the same thing.
func (ms *MockSuite) TestBaremetalListByTagStillHonoursFilter(assert, require *td.T) {
	var seen map[string][]map[string]any
	captureTagQuery(&seen)

	out, err := cmd.Execute("baremetal", "list", "--tag", "owner:EXISTS", "--filter", `datacenter=="gra1"`)

	require.CmpNoError(err)
	assert.Cmp(seen["owner"][0]["operator"], "EXISTS", "the tag still went to the API")
	assert.Cmp(out, td.Not(td.Contains("ns1.example")), "and the table filter still removed the row")
}

// A filter the API would reject is refused here, with the operators that exist.
func (ms *MockSuite) TestBaremetalListByTagRefusesAnUnknownOperator(assert, require *td.T) {
	_, err := cmd.Execute("baremetal", "list", "--tag", "owner:CONTAINS=Denis")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("NEXISTS"))
	assert.Cmp(httpmock.GetCallCountInfo()["GET "+serversV2], 0, "nothing was asked of the API")
	assert.Cmp(httpmock.GetCallCountInfo()["GET "+serversV1], 0)
}
