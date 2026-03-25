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

func (ms *MockSuite) TestCloudInstanceApplicationAccessCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodPost,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/instance/fakeInstanceID/applicationAccess",
		httpmock.NewStringResponder(200, `{
			"status": "ok",
			"accesses": [
				{
					"type": "webadmin",
					"url": "https://1.2.3.4/admin",
					"login": "admin",
					"password": "s3cret",
					"database": ""
				}
			]
		}`).Once(),
	)

	out, err := cmd.Execute("cloud", "instance", "application-access", "fakeInstanceID", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(cleanWhitespacesHelper(out), `
  # 🚀 Application Access fakeInstanceID

  **Status**: ok

  Credentials:

   Type       | URL                   | Login      | Password   | Database
  ------------|-----------------------|------------|------------|-----------
   webadmin   | https://1.2.3.4/admin | admin      | s3cret     |

  💡 Use option -o json or -o yaml to get the raw output with all information

`)
}

func (ms *MockSuite) TestCloudInstanceGroupListCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/instance/group",
		httpmock.NewStringResponder(200, `[
			{
				"id": "group-id-1",
				"name": "my-group",
				"type": "affinity",
				"region": "GRA9",
				"instance_ids": ["inst-1", "inst-2"]
			}
		]`).Once(),
	)

	out, err := cmd.Execute("cloud", "instance", "group", "ls", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Contains(out, "my-group")
	assert.Contains(out, "group-id-1")
}

func (ms *MockSuite) TestCloudInstanceGroupDeleteCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodDelete,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/instance/group/group-id-1",
		httpmock.NewStringResponder(204, ``).Once(),
	)

	out, err := cmd.Execute("cloud", "instance", "group", "delete", "group-id-1", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Contains(out, "Instance group successfully deleted")
}

func (ms *MockSuite) TestCloudInstanceAutobackupListCmd(assert, require *td.T) {
	// First call to get instance region
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/instance/fakeInstanceID",
		httpmock.NewStringResponder(200, `{
			"id": "fakeInstanceID",
			"region": "GRA9",
			"status": "ACTIVE"
		}`).Once(),
	)

	// Then list autobackups for that region
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/GRA9/workflow/backup",
		httpmock.NewStringResponder(200, `[
			{
				"id": "backup-id-1",
				"name": "daily-backup",
				"instanceId": "fakeInstanceID",
				"cron": "0 0 * * *",
				"rotation": 7,
				"nextExecutionTime": "2026-04-03T00:00:00Z"
			}
		]`).Once(),
	)

	out, err := cmd.Execute("cloud", "instance", "autobackup", "ls", "fakeInstanceID", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Contains(out, "daily-backup")
	assert.Contains(out, "backup-id-1")
}

func (ms *MockSuite) TestCloudInstanceAutobackupDeleteCmd(assert, require *td.T) {
	// First call to get instance region
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/instance/fakeInstanceID",
		httpmock.NewStringResponder(200, `{
			"id": "fakeInstanceID",
			"region": "GRA9",
			"status": "ACTIVE"
		}`).Once(),
	)

	httpmock.RegisterResponder(http.MethodDelete,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/GRA9/workflow/backup/backup-id-1",
		httpmock.NewStringResponder(204, ``).Once(),
	)

	out, err := cmd.Execute("cloud", "instance", "autobackup", "delete", "fakeInstanceID", "backup-id-1", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Contains(out, "Autobackup workflow deleted")
}

func (ms *MockSuite) TestCloudInstanceGroupGetCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/instance/group/group-id-1",
		httpmock.NewStringResponder(200, `{
			"id": "group-id-1",
			"name": "my-group",
			"type": "affinity",
			"region": "GRA9",
			"instance_ids": ["inst-1", "inst-2"]
		}`).Once(),
	)

	out, err := cmd.Execute("cloud", "instance", "group", "get", "group-id-1", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Contains(out, "my-group")
	assert.Contains(out, "group-id-1")
}

func (ms *MockSuite) TestCloudInstanceGroupCreateCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodPost,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/instance/group",
		httpmock.NewStringResponder(200, `{
			"id": "group-id-new",
			"name": "new-group",
			"type": "affinity",
			"region": "GRA9",
			"instance_ids": []
		}`).Once(),
	)

	out, err := cmd.Execute("cloud", "instance", "group", "create", "new-group", "GRA9", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Contains(out, "group-id-new")
}

func (ms *MockSuite) TestCloudInstanceAutobackupGetCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/instance/fakeInstanceID",
		httpmock.NewStringResponder(200, `{
			"id": "fakeInstanceID",
			"region": "GRA9",
			"status": "ACTIVE"
		}`).Once(),
	)

	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/GRA9/workflow/backup/backup-id-1",
		httpmock.NewStringResponder(200, `{
			"id": "backup-id-1",
			"name": "daily-backup",
			"instanceId": "fakeInstanceID",
			"cron": "0 0 * * *",
			"rotation": 7
		}`).Once(),
	)

	out, err := cmd.Execute("cloud", "instance", "autobackup", "get", "fakeInstanceID", "backup-id-1", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Contains(out, "daily-backup")
	assert.Contains(out, "backup-id-1")
}

func (ms *MockSuite) TestCloudInstanceAutobackupCreateCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/instance/fakeInstanceID",
		httpmock.NewStringResponder(200, `{
			"id": "fakeInstanceID",
			"region": "GRA9",
			"status": "ACTIVE"
		}`).Once(),
	)

	httpmock.RegisterResponder(http.MethodPost,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/GRA9/workflow/backup",
		httpmock.NewStringResponder(200, `{
			"id": "backup-new-id",
			"name": "my-backup",
			"instanceId": "fakeInstanceID",
			"cron": "0 0 * * *",
			"rotation": 7
		}`).Once(),
	)

	out, err := cmd.Execute("cloud", "instance", "autobackup", "create", "fakeInstanceID",
		"--cron", "0 0 * * *", "--rotation", "7", "--name", "my-backup",
		"--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Contains(out, "backup-new-id")
}

func (ms *MockSuite) TestCloudInstanceNullImageCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/instance/fakeInstanceID",
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

  💡 Use option -o json or -o yaml to get the raw output with all information

`)
}
