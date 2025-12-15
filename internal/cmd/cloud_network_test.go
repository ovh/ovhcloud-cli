// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"encoding/json"
	"net/http"

	"github.com/jarcoal/httpmock"
	"github.com/maxatome/go-testdeep/td"
	"github.com/maxatome/tdhttpmock"
	"github.com/ovh/ovhcloud-cli/internal/cmd"
)

func (ms *MockSuite) TestCloudPrivateNetworkCreateCmd(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(http.MethodPost,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/BHS5/network",
		tdhttpmock.JSONBody(td.JSON(`
			{
				"gateway": {
					"model": "s",
					"name": "TestFromTheCLI"
				},
				"name": "TestFromTheCLI",
				"subnet": {
					"cidr": "10.0.0.2/24",
					"enableDhcp": false,
					"enableGatewayIp": true,
					"ipVersion": 4
				}
			}`),
		),
		httpmock.NewStringResponder(200, `{"id": "operation-12345"}`),
	)

	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/operation/operation-12345",
		httpmock.NewStringResponder(200, `
		{
			"id": "6610ec10-9b09-11f0-a8ac-0050568ce122",
			"action": "network#create",
			"createdAt": "2025-09-26T20:43:14.376907+02:00",
			"startedAt": "2025-09-26T20:43:14.376907+02:00",
			"completedAt": "2025-09-26T20:43:36.631086+02:00",
			"progress": 0,
			"regions": [
				"BHS5"
			],
			"resourceId": "80c1de3e-9b09-11f0-993b-0050568ce122",
			"status": "completed",
			"subOperations": [
				{
					"id": "8c0806ba-9b09-11f0-9a54-0050568ce122",
					"action": "gateway#create",
					"startedAt": "2025-09-26T20:43:14.376907+02:00",
					"completedAt": "2025-09-26T20:43:36.631086+02:00",
					"progress": 0,
					"regions": [
						"BHS5"
					],
					"resourceId": "97a2703c-9b09-11f0-9b6c-0050568ce122",
					"status": "completed"
				}
			]
		}`),
	)

	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/network/private",
		httpmock.NewStringResponder(200, `[
			{
				"id": "pn-example",
				"name": "TestFromTheCLI",
				"vlanId": 1234,
				"regions": [
					{
						"region": "BHS5",
						"status": "ACTIVE",
						"openstackId": "80c1de3e-9b09-11f0-993b-0050568ce122"
					}
				],
				"type": "private",
				"status": "ACTIVE"
			}
		]`),
	)

	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/BHS5/network/80c1de3e-9b09-11f0-993b-0050568ce122/subnet",
		httpmock.NewStringResponder(200, `[
			{
				"id": "c59a3fdc-9b0f-11f0-ac97-0050568ce122",
				"name": "TestFromTheCLI",
				"cidr": "10.0.0.0/24",
				"ipVersion": 4,
				"dhcpEnabled": false,
				"gatewayIp": "10.0.0.1",
				"allocationPools": [
					{
						"start": "10.0.0.2",
						"end": "10.0.0.254"
					}
				]
			}
		]`),
	)

	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/BHS5/gateway?subnetId=c59a3fdc-9b0f-11f0-ac97-0050568ce122",
		httpmock.NewStringResponder(200, `[
			{
				"id": "e7045f34-8f2b-41a4-a734-97b7b0e323de",
				"status": "active",
				"name": "TestFromTheCLI",
				"interfaces": [
					{
						"id": "56d17852-9b11-11f0-8d13-0050568ce122",
						"ip": "10.0.0.1",
						"subnetId": "56d17852-9b11-11f0-8d13-0050568ce122",
						"networkId": "c59a3fdc-9b0f-11f0-ac97-0050568ce122"
					},
					{
						"id": "56d17852-9b11-11f0-8d13-0050568ce122",
						"ip": "10.0.0.218",
						"subnetId": "56d17852-9b11-11f0-8d13-0050568ce122",
						"networkId": "c59a3fdc-9b0f-11f0-ac97-0050568ce122"
					}
				],
				"externalInformation": {
					"ips": [
						{
							"ip": "1.2.3.4",
							"subnetId": "981c226c-57da-4766-966b-3b45db0cfc84"
						}
					],
					"networkId": "c59a3fdc-9b0f-11f0-ac97-0050568ce122"
				},
				"region": "BHS5",
				"model": "s"
			}
		]`),
	)

	out, err := cmd.Execute("cloud", "network", "private", "create", "BHS5", "--cloud-project", "fakeProjectID",
		"--gateway-model", "s", "--gateway-name", "TestFromTheCLI", "--name", "TestFromTheCLI", "--subnet-cidr",
		"10.0.0.2/24", "--subnet-ip-version", "4", "--wait", "--subnet-enable-gateway-ip", "--yaml")
	require.CmpNoError(err)
	assert.String(out, `details:
  id: pn-example
  openstackId: 80c1de3e-9b09-11f0-993b-0050568ce122
  region: BHS5
  subnets:
  - gateways:
    - id: e7045f34-8f2b-41a4-a734-97b7b0e323de
      name: TestFromTheCLI
    id: c59a3fdc-9b0f-11f0-ac97-0050568ce122
    name: TestFromTheCLI
message: '✅ Network pn-example created successfully (Openstack ID: 80c1de3e-9b09-11f0-993b-0050568ce122)'
`)
}

func (ms *MockSuite) TestCloudPrivateNetworkSubnetCreateCmd(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(http.MethodPost,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/network/private/pn-123456/subnet",
		tdhttpmock.JSONBody(td.JSON(`
			{
				"dhcp": false,
				"end": "192.168.1.24",
				"network": "192.168.1.0/24",
				"noGateway": false,
				"region": "BHS5",
				"start": "192.168.1.12"
			}`),
		),
		httpmock.NewStringResponder(200, `
			{
				"cidr": "192.168.1.0/24",
				"gatewayIp": "192.168.1.1",
				"id": "5e625f90-9ec3-11f0-9f75-0050568ce122",
				"ipPools": [
					{
						"dhcp": false,
						"end": "192.168.1.24",
						"network": "192.168.1.0/24",
						"region": "BHS5",
						"start": "192.168.1.12"
					}
				]
			}`,
		),
	)

	out, err := cmd.Execute("cloud", "network", "private", "subnet", "create", "pn-123456", "--cloud-project", "fakeProjectID",
		"--network", "192.168.1.0/24", "--start", "192.168.1.12", "--end", "192.168.1.24", "--region", "BHS5", "--json")
	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{
		"message": "✅ Subnet 5e625f90-9ec3-11f0-9f75-0050568ce122 created successfully",
		"details": {
			"cidr": "192.168.1.0/24",
			"gatewayIp": "192.168.1.1",
			"id": "5e625f90-9ec3-11f0-9f75-0050568ce122",
			"ipPools": [
				{
					"dhcp": false,
					"end": "192.168.1.24",
					"network": "192.168.1.0/24",
					"region": "BHS5",
					"start": "192.168.1.12"
				}
			]
		}
	}`))
}

func (ms *MockSuite) TestCloudLoadbalancerGetCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, "https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region",
		httpmock.NewStringResponder(200, `["GRA11", "SBG5", "BHS5"]`))

	httpmock.RegisterResponder(http.MethodGet, "https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/GRA11",
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

	httpmock.RegisterResponder(http.MethodGet, "https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/SBG5",
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

	httpmock.RegisterResponder(http.MethodGet, "https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/BHS5",
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
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/SBG5/loadbalancing/loadbalancer/fakeLB",
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
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/SBG5/loadbalancing/flavor/f862fa22-6275-4f8f-885e-66a8faf5e44e",
		httpmock.NewStringResponder(200, `{
			"id": "f862fa22-6275-4f8f-885e-66a8faf5e44e",
			"name": "medium",
			"description": "Medium Load Balancer Flavor"
		}`))

	out, err := cmd.Execute("cloud", "network", "loadbalancer", "get", "fakeLB", "--cloud-project", "fakeProjectID")
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
