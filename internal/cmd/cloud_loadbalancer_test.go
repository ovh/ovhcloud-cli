// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"net/http"

	"github.com/jarcoal/httpmock"
	"github.com/maxatome/go-testdeep/td"
	"github.com/ovh/ovhcloud-cli/internal/cmd"
)

func (ms *MockSuite) TestCloudLoadbalancerGetCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, "https://eu.api.ovh.com/1.0/cloud/project/fakeProjectID/region",
		httpmock.NewStringResponder(200, `["GRA11", "SBG5", "BHS5"]`))

	httpmock.RegisterResponder(http.MethodGet, "https://eu.api.ovh.com/1.0/cloud/project/fakeProjectID/region/GRA11",
		httpmock.NewStringResponder(200, `{
			"name": "GRA11",
			"type": "region",
			"status": "UP",
			"services": [
				{
					"name": "octavialoadbalancer",
					"status": "UP"
				}
			],
			"countryCode": "fr",
			"ipCountries": [],
			"continentCode": "NA",
			"availabilityZones": [],
			"datacenterLocation": "GRA11"
		}`))

	httpmock.RegisterResponder(http.MethodGet, "https://eu.api.ovh.com/1.0/cloud/project/fakeProjectID/region/SBG5",
		httpmock.NewStringResponder(200, `{
			"name": "SBG5",
			"type": "region",
			"status": "UP",
			"services": [
				{
					"name": "octavialoadbalancer",
					"status": "UP"
				}
			],
			"countryCode": "fr",
			"ipCountries": [],
			"continentCode": "NA",
			"availabilityZones": [],
			"datacenterLocation": "SBG5"
		}`))

	httpmock.RegisterResponder(http.MethodGet, "https://eu.api.ovh.com/1.0/cloud/project/fakeProjectID/region/BHS5",
		httpmock.NewStringResponder(200, `{
			"name": "BHS5",
			"type": "region",
			"status": "UP",
			"services": [],
			"countryCode": "ca",
			"ipCountries": [],
			"continentCode": "NA",
			"availabilityZones": [],
			"datacenterLocation": "BHS5"
		}`))

	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/1.0/cloud/project/fakeProjectID/region/SBG5/loadbalancing/loadbalancer/fakeLB",
		httpmock.NewStringResponder(200, `{
			"createdAt": "2024-07-30T08:26:51Z",
			"flavorId": "f862fa22-6275-4f8f-885e-66a8faf5e44e",
			"floatingIp": null,
			"id": "334fc97e-a8db-11f0-944d-0050568ce122",
			"name": "loadbalancer-sbg5-2024-07-30",
			"operatingStatus": "online",
			"provisioningStatus": "active",
			"region": "SBG5",
			"updatedAt": "2025-10-14T08:48:33Z",
			"vipAddress": "1.2.3.4",
			"vipNetworkId": "3f29f530-a8db-11f0-9ab2-0050568ce122",
			"vipSubnetId": "44a869c4-a8db-11f0-899f-0050568ce122"
		}`))

	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/1.0/cloud/project/fakeProjectID/region/SBG5/loadbalancing/flavor/f862fa22-6275-4f8f-885e-66a8faf5e44e",
		httpmock.NewStringResponder(200, `{
			"id": "f862fa22-6275-4f8f-885e-66a8faf5e44e",
			"name": "medium",
			"description": "Medium Load Balancer Flavor"
		}`))

	out, err := cmd.Execute("cloud", "loadbalancer", "get", "fakeLB", "--cloud-project", "fakeProjectID")
	require.CmpNoError(err)
	assert.Cmp(cleanWhitespacesHelper(out), `
  # 🚀 Load balancer fakeLB

  *loadbalancer-sbg5-2024-07-30*

  ## General information

  **Region**:              SBG5
  **Operating status**:    online
  **Provisioning status**: active
  **Flavor**:              medium (ID: f862fa22-6275-4f8f-885e-66a8faf5e44e)
  **Creation date**:       2024-07-30T08:26:51Z

  ## Technical information

  **VIP address**:        1.2.3.4
  **VIP network ID**:     3f29f530-a8db-11f0-9ab2-0050568ce122
  **VIP subnet ID**:      44a869c4-a8db-11f0-899f-0050568ce122

  💡 Use option --json or --yaml to get the raw output with all information

`)
}
