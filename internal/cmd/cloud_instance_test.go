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

func (ms *MockSuite) TestCloudInstanceNullImageCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/1.0/cloud/project/fakeProjectID/instance/fakeInstanceID",
		httpmock.NewStringResponder(200, `
			{
				"id": "fakeInstanceID",
				"name": "TestInstance",
				"ipAddresses": [
					{
						"ip": "1.2.3.4",
						"type": "public",
						"version": 4,
						"networkId": "bc63b98d13fbba642b2653711cc9d156ca7b404f009f7227172d37b5280a6",
						"gatewayIp": "1.2.3.4"
					},
					{
						"ip": "2001:db8::1",
						"type": "public",
						"version": 6,
						"networkId": "bc63b98d13fbba642b2653711cc9d156ca7b404f009f7227172d37b5280a6",
						"gatewayIp": "2001:db8::ff"
					}
				],
				"status": "ACTIVE",
				"created": "2025-09-24T17:21:31Z",
				"region": "GRA9",
				"flavor": {
					"id": "906e8259-0340-4856-95b5-4ea2d26fe377",
					"name": "b2-7",
					"region": "GRA9",
					"ram": 7,
					"disk": 50,
					"vcpus": 2,
					"type": "ovh.ssd.eg",
					"osType": "linux",
					"inboundBandwidth": 250,
					"outboundBandwidth": 250,
					"available": true,
					"planCodes": {
						"monthly": "b2-7.monthly.postpaid",
						"hourly": "b2-7.consumption",
						"license": null
					},
					"capabilities": [
						{
							"name": "resize",
							"enabled": true
						},
						{
							"name": "snapshot",
							"enabled": true
						},
						{
							"name": "volume",
							"enabled": true
						},
						{
							"name": "failoverip",
							"enabled": true
						}
					],
					"quota": 791
				},
				"image": null,
				"sshKey": null,
				"monthlyBilling": null,
				"planCode": "b2-7.consumption",
				"licensePlanCode": null,
				"operationIds": [],
				"currentMonthOutgoingTraffic": null,
				"rescuePassword": null,
				"availabilityZone": null
			}`,
		),
	)

	out, err := cmd.Execute("cloud", "instance", "get", "fakeInstanceID", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(cleanWhitespacesHelper(out), `
  # 🚀 Instance fakeInstanceID

  *TestInstance*

  ## General information

  **Region**:            GRA9
  **Availability zone**:
  **Status**:            ACTIVE
  **Creation date**:     2025-09-24T17:21:31Z

  IP addresses:

   IP                     | Type                   | Gateway IP
  ------------------------|------------------------|------------------------
   1.2.3.4                | public                 | 1.2.3.4
   2001:db8::1            | public                 | 2001:db8::ff

  ## Flavor details

  **Name**:                   b2-7
  **Operating system**:       linux
  **Storage**:                50 GB
  **RAM**:                    7 GB
  **vCPUs**:                  2
  **Max inbound bandwidth**:  250 Mbit/s
  **Max outbound bandwidth**: 250 Mbit/s

  💡 Use option --json or --yaml to get the raw output with all information

`)
}
