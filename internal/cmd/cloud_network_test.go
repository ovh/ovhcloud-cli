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
				"name": "TestFromTheCLI"
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

	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/BHS5/network/80c1de3e-9b09-11f0-993b-0050568ce122",
		httpmock.NewStringResponder(200, `{
			"id": "80c1de3e-9b09-11f0-993b-0050568ce122",
			"name": "TestFromTheCLI",
			"region": "BHS5",
			"visibility": "private",
			"vlanId": 1234
		}`),
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
		"--name", "TestFromTheCLI", "--wait", "-o", "yaml")
	require.CmpNoError(err)
	assert.String(out, `message: ✅ Network 80c1de3e-9b09-11f0-993b-0050568ce122 created successfully in region
  BHS5
`)
}

func (ms *MockSuite) TestCloudPrivateNetworkSubnetCreateCmd(assert, require *td.T) {
	// findNetwork will look up the network by region
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/BHS5/network/pn-123456",
		httpmock.NewStringResponder(200, `{
			"id": "pn-123456",
			"name": "test-network",
			"region": "BHS5",
			"visibility": "private",
			"vlanId": 0
		}`),
	)

	httpmock.RegisterMatcherResponder(http.MethodPost,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/BHS5/network/pn-123456/subnet",
		tdhttpmock.JSONBody(td.JSON(`
			{
				"name": "my-subnet",
				"cidr": "192.168.1.0/24",
				"ipVersion": 4,
				"enableDhcp": true,
				"enableGatewayIp": true
			}`),
		),
		httpmock.NewStringResponder(200, `
			{
				"cidr": "192.168.1.0/24",
				"gatewayIp": "192.168.1.1",
				"id": "5e625f90-9ec3-11f0-9f75-0050568ce122",
				"name": "my-subnet",
				"ipVersion": 4,
				"dhcpEnabled": true,
				"allocationPools": [
					{
						"start": "192.168.1.2",
						"end": "192.168.1.254"
					}
				]
			}`,
		),
	)

	out, err := cmd.Execute("cloud", "network", "private", "subnet", "create", "pn-123456", "--region", "BHS5", "--cloud-project", "fakeProjectID",
		"--name", "my-subnet", "--cidr", "192.168.1.0/24", "--ip-version", "4", "--enable-dhcp", "--enable-gateway-ip", "-o", "json")
	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{
		"message": "✅ Subnet 5e625f90-9ec3-11f0-9f75-0050568ce122 created successfully",
		"details": {
			"cidr": "192.168.1.0/24",
			"gatewayIp": "192.168.1.1",
			"id": "5e625f90-9ec3-11f0-9f75-0050568ce122",
			"name": "my-subnet",
			"ipVersion": 4,
			"dhcpEnabled": true,
			"allocationPools": [
				{
					"start": "192.168.1.2",
					"end": "192.168.1.254"
				}
			]
		}
	}`))
}

func (ms *MockSuite) TestCloudPrivateNetworkSubnetCreateCmdInferIPVersion(assert, require *td.T) {
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/BHS5/network/pn-123456",
		httpmock.NewStringResponder(200, `{
			"id": "pn-123456",
			"name": "test-network",
			"region": "BHS5",
			"visibility": "private",
			"vlanId": 0
		}`),
	)

	httpmock.RegisterMatcherResponder(http.MethodPost,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/BHS5/network/pn-123456/subnet",
		tdhttpmock.JSONBody(td.JSON(`
			{
				"name": "my-subnet",
				"cidr": "192.168.1.0/24",
				"ipVersion": 4,
				"enableDhcp": true,
				"enableGatewayIp": true
			}`),
		),
		httpmock.NewStringResponder(200, `
			{
				"cidr": "192.168.1.0/24",
				"gatewayIp": "192.168.1.1",
				"id": "subnet-inferred-v4",
				"name": "my-subnet",
				"ipVersion": 4,
				"dhcpEnabled": true,
				"allocationPools": [
					{
						"start": "192.168.1.2",
						"end": "192.168.1.254"
					}
				]
			}`,
		),
	)

	// Note: --ip-version is NOT provided, it should be inferred from the CIDR
	out, err := cmd.Execute("cloud", "network", "private", "subnet", "create", "pn-123456", "--region", "BHS5", "--cloud-project", "fakeProjectID",
		"--name", "my-subnet", "--cidr", "192.168.1.0/24", "--enable-dhcp", "--enable-gateway-ip", "-o", "json")
	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{
		"message": "✅ Subnet subnet-inferred-v4 created successfully",
		"details": {
			"cidr": "192.168.1.0/24",
			"gatewayIp": "192.168.1.1",
			"id": "subnet-inferred-v4",
			"name": "my-subnet",
			"ipVersion": 4,
			"dhcpEnabled": true,
			"allocationPools": [
				{
					"start": "192.168.1.2",
					"end": "192.168.1.254"
				}
			]
		}
	}`))
}

