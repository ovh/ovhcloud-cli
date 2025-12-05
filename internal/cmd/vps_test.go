// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"encoding/json"

	"github.com/jarcoal/httpmock"
	"github.com/maxatome/go-testdeep/td"
	"github.com/ovh/ovhcloud-cli/internal/cmd"
)

func (ms *MockSuite) TestVpsListCmd(assert, require *td.T) {
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/vps",
		httpmock.NewStringResponder(200, `["vps-12345","vps-67890"]`).Once())

	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/vps/vps-12345",
		httpmock.NewStringResponder(200, `{"name": "vps-12345", "displayName": "VPS 12345", "state": "running", "zone": "Region OpenStack: os-waw2"}`).Once())

	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/vps/vps-67890",
		httpmock.NewStringResponder(200, `{"name": "vps-67890", "displayName": "VPS 67890", "state": "stopped", "zone": "Region OpenStack: os-gra1"}`).Once())

	out, err := cmd.Execute("vps", "ls", "--json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`[
		{
			"displayName": "VPS 12345",
			"name": "vps-12345",
			"state": "running",
			"zone": "Region OpenStack: os-waw2"
		},
		{
			"displayName": "VPS 67890",
			"name": "vps-67890",
			"state": "stopped",
			"zone": "Region OpenStack: os-gra1"
		}
	]`))
}

func (ms *MockSuite) TestVpsGetCmd(assert, require *td.T) {
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/vps/vps-67890",
		httpmock.NewStringResponder(200, `{"name": "vps-67890", "displayName": "VPS 67890", "state": "stopped", "zone": "Region OpenStack: os-gra1"}`).Once())

	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/vps/vps-67890/datacenter",
		httpmock.NewStringResponder(200, `{"country": "fr", "name": "os-gra1", "longName": "Region OpenStack: os-gra1"}`).Once())

	out, err := cmd.Execute("vps", "get", "vps-67890", "--json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{
		"displayName": "VPS 67890",
		"name": "vps-67890",
		"state": "stopped",
		"zone": "Region OpenStack: os-gra1",
		"datacenter": {
			"country": "fr",
			"name": "os-gra1",
			"longName": "Region OpenStack: os-gra1"
		}
	}`))
}
