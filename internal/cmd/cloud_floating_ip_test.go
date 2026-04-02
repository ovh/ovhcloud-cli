// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"encoding/json"
	"net/http"

	"github.com/jarcoal/httpmock"
	"github.com/maxatome/go-testdeep/td"
	"github.com/ovh/ovhcloud-cli/internal/cmd"
)

func registerFloatingIPRegionMocks() {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region",
		httpmock.NewStringResponder(200, `["GRA11", "SBG5", "BHS5"]`))

	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/GRA11",
		httpmock.NewStringResponder(200, `{
			"name": "GRA11",
			"type": "region",
			"status": "UP",
			"services": [{"name": "network", "status": "UP"}]
		}`))

	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/SBG5",
		httpmock.NewStringResponder(200, `{
			"name": "SBG5",
			"type": "region",
			"status": "UP",
			"services": [{"name": "network", "status": "UP"}]
		}`))

	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/BHS5",
		httpmock.NewStringResponder(200, `{
			"name": "BHS5",
			"type": "region",
			"status": "UP",
			"services": []
		}`))
}

// ---------------------------------------------------------------------------
// Floating IP – list
// ---------------------------------------------------------------------------

func (ms *MockSuite) TestCloudFloatingIPListCmd(assert, require *td.T) {
	registerFloatingIPRegionMocks()

	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/GRA11/floatingip",
		httpmock.NewStringResponder(200, `[
			{
				"id": "fip-gra-001",
				"ip": "1.2.3.4",
				"status": "active",
				"region": "GRA11",
				"networkId": "net-001",
				"associatedEntity": null
			}
		]`))

	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/SBG5/floatingip",
		httpmock.NewStringResponder(200, `[
			{
				"id": "fip-sbg-001",
				"ip": "5.6.7.8",
				"status": "active",
				"region": "SBG5",
				"networkId": "net-002",
				"associatedEntity": {"id": "port-001", "type": "instance"}
			}
		]`))

	out, err := cmd.Execute("cloud", "floating-ip", "ls", "--cloud-project", "fakeProjectID", "-o", "json")
	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`[
		{
			"id": "fip-gra-001",
			"ip": "1.2.3.4",
			"status": "active",
			"region": "GRA11",
			"networkId": "net-001",
			"associatedEntity": null
		},
		{
			"id": "fip-sbg-001",
			"ip": "5.6.7.8",
			"status": "active",
			"region": "SBG5",
			"networkId": "net-002",
			"associatedEntity": {"id": "port-001", "type": "instance"}
		}
	]`))
}

func (ms *MockSuite) TestCloudFloatingIPListWithRegionFilterCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/GRA11/floatingip",
		httpmock.NewStringResponder(200, `[
			{
				"id": "fip-gra-001",
				"ip": "1.2.3.4",
				"status": "active",
				"region": "GRA11",
				"networkId": "net-001",
				"associatedEntity": null
			}
		]`))

	out, err := cmd.Execute("cloud", "floating-ip", "ls", "--cloud-project", "fakeProjectID", "--region", "GRA11", "-o", "json")
	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`[
		{
			"id": "fip-gra-001",
			"ip": "1.2.3.4",
			"status": "active",
			"region": "GRA11",
			"networkId": "net-001",
			"associatedEntity": null
		}
	]`))
}

// ---------------------------------------------------------------------------
// Floating IP – get
// ---------------------------------------------------------------------------

func (ms *MockSuite) TestCloudFloatingIPGetCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/GRA11/floatingip/fip-gra-001",
		httpmock.NewStringResponder(200, `{
			"id": "fip-gra-001",
			"ip": "1.2.3.4",
			"status": "active",
			"region": "GRA11",
			"networkId": "net-001",
			"associatedEntity": null
		}`))

	out, err := cmd.Execute("cloud", "floating-ip", "get", "fip-gra-001", "--cloud-project", "fakeProjectID", "--region", "GRA11", "-o", "json")
	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{
		"id": "fip-gra-001",
		"ip": "1.2.3.4",
		"status": "active",
		"region": "GRA11",
		"networkId": "net-001",
		"associatedEntity": null
	}`))
}

// ---------------------------------------------------------------------------
// Floating IP – delete
// ---------------------------------------------------------------------------

func (ms *MockSuite) TestCloudFloatingIPDeleteCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodDelete,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/GRA11/floatingip/fip-gra-001",
		httpmock.NewStringResponder(200, `null`))

	out, err := cmd.Execute("cloud", "floating-ip", "delete", "fip-gra-001", "--cloud-project", "fakeProjectID", "--region", "GRA11", "-o", "json")
	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{
		"message": "✅ Floating IP fip-gra-001 deleted successfully"
	}`))
}
