// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"net/http"

	"github.com/jarcoal/httpmock"
	"github.com/maxatome/go-testdeep/td"
	"github.com/maxatome/tdhttpmock"
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

// instanceV2Response is a v2 instance as returned by the API (shape taken from
// a real GET /v2/publicCloud/project/{projectId}/compute/instance/{instanceId}).
const instanceV2Response = `{
	"id": "fakeInstanceID",
	"checksum": "abc123",
	"createdAt": "2026-10-05T13:54:47Z",
	"updatedAt": "2026-10-05T14:00:00Z",
	"resourceStatus": "READY",
	"currentTasks": [],
	"targetSpec": {
		"name": "my-instance",
		"flavor": { "id": "flavor-d2-2" },
		"image": { "id": "image-debian-12" },
		"location": { "region": "GRA11" },
		"networks": [
			{ "autoAssignPublicIp": true },
			{ "id": "private-net-id", "subnetId": "private-subnet-id", "ip": "10.1.2.155" }
		],
		"powerState": "ACTIVE",
		"securityGroups": [ { "id": "sg-default" } ],
		"sshKeyName": "my-key",
		"volumes": [ { "id": "volume-1" } ]
	},
	"currentState": {
		"name": "my-instance",
		"flavor": {
			"id": "flavor-d2-2",
			"name": "d2-2",
			"vcpus": 1,
			"ram": 2000,
			"disk": 25,
			"gpus": 0,
			"bandwidth": { "private": 4000, "publicInbound": 100, "publicOutbound": 100 }
		},
		"image": { "id": "image-debian-12", "name": "Debian 12", "defaultUser": "debian", "status": "ACTIVE" },
		"location": { "region": "GRA11" },
		"locked": false,
		"networks": [
			{
				"addresses": [
					{ "ip": "51.210.167.194", "mac": "fa:16:3e:02:2d:c8", "type": "FIXED", "version": 4 },
					{ "ip": "2001:41d0:304:300::66cf", "mac": "fa:16:3e:02:2d:c8", "type": "FIXED", "version": 6 }
				]
			},
			{
				"id": "private-net-id",
				"subnetId": "private-subnet-id",
				"addresses": [
					{ "ip": "10.1.2.155", "mac": "fa:16:3e:12:77:cf", "type": "FIXED", "version": 4 }
				]
			}
		],
		"powerState": "ACTIVE",
		"securityGroups": [ { "id": "sg-default" } ],
		"sshKeyName": "my-key",
		"volumes": [ { "id": "volume-1", "name": "", "size": 0 } ]
	}
}`

func (ms *MockSuite) TestCloudInstanceListCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/compute/instance",
		httpmock.NewStringResponder(200, `[
			`+instanceV2Response+`,
			{
				"id": "creatingInstanceID",
				"resourceStatus": "CREATING",
				"targetSpec": {
					"name": "new-instance",
					"flavor": { "id": "flavor-d2-2" },
					"location": { "region": "SBG5" },
					"powerState": "ACTIVE"
				},
				"currentState": null
			}
		]`))

	out, err := cmd.Execute("cloud", "instance", "list", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("fakeInstanceID"))
	assert.Cmp(out, td.Contains("my-instance"))
	assert.Cmp(out, td.Contains("GRA11"))
	assert.Cmp(out, td.Contains("creatingInstanceID"))
	assert.Cmp(out, td.Contains("new-instance"))
	assert.Cmp(out, td.Contains("SBG5"))
	assert.Cmp(out, td.Contains("CREATING"))
}

func (ms *MockSuite) TestCloudInstanceGetCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/compute/instance/fakeInstanceID",
		httpmock.NewStringResponder(200, instanceV2Response))

	out, err := cmd.Execute("cloud", "instance", "get", "fakeInstanceID", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("my-instance"))
	assert.Cmp(out, td.Re(`Status\**:\s+READY`))
	assert.Cmp(out, td.Re(`Power state\**:\s+ACTIVE`))
	assert.Cmp(out, td.Contains("GRA11"))
	assert.Cmp(out, td.Contains("d2-2"))
	assert.Cmp(out, td.Contains("Debian 12"))
	assert.Cmp(out, td.Contains("51.210.167.194"))
	assert.Cmp(out, td.Contains("2001:41d0:304:300::66cf"))
	assert.Cmp(out, td.Contains("10.1.2.155"))
	assert.Cmp(out, td.Contains("private-net-id"))
	assert.Cmp(out, td.Contains("volume-1"))
	assert.Cmp(out, td.Contains("sg-default"))
	assert.Cmp(out, td.Contains("my-key"))
	assert.Cmp(out, td.Not(td.Contains("<no value>")))
}

func (ms *MockSuite) TestCloudInstanceGetNotFoundCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/compute/instance/missingInstanceID",
		httpmock.NewStringResponder(404, `{"class":"Client::NotFound::InstanceDoesNotExist","message":"Instance not found"}`))

	_, err := cmd.Execute("cloud", "instance", "get", "missingInstanceID", "--cloud-project", "fakeProjectID")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("Client::NotFound::InstanceDoesNotExist"))
}

func (ms *MockSuite) TestCloudInstanceNullImageCmd(assert, require *td.T) {
	// A boot-from-volume instance has no image: the image section must be
	// skipped instead of rendering empty values.
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/compute/instance/fakeInstanceID",
		httpmock.NewStringResponder(200, `{
			"id": "fakeInstanceID",
			"resourceStatus": "READY",
			"createdAt": "2025-09-24T17:21:31Z",
			"targetSpec": {
				"name": "TestInstance",
				"flavor": { "id": "flavor-b2-7" },
				"location": { "region": "GRA9" },
				"powerState": "ACTIVE",
				"volumes": [ { "id": "boot-volume" } ]
			},
			"currentState": {
				"name": "TestInstance",
				"flavor": { "id": "flavor-b2-7", "name": "b2-7", "vcpus": 2, "ram": 7000, "disk": 50 },
				"location": { "region": "GRA9" },
				"powerState": "ACTIVE",
				"networks": [
					{ "addresses": [ { "ip": "1.2.3.4", "type": "FIXED", "version": 4 } ] }
				],
				"volumes": [ { "id": "boot-volume" } ]
			}
		}`))

	out, err := cmd.Execute("cloud", "instance", "get", "fakeInstanceID", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("TestInstance"))
	assert.Cmp(out, td.Contains("b2-7"))
	assert.Cmp(out, td.Contains("1.2.3.4"))
	assert.Cmp(out, td.Not(td.Contains("Image")))
	assert.Cmp(out, td.Not(td.Contains("<no value>")))
}

func (ms *MockSuite) TestCloudInstanceDeleteCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodDelete,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/compute/instance/fakeInstanceID",
		httpmock.NewStringResponder(202, `{"id": "fakeInstanceID", "resourceStatus": "DELETING"}`))

	out, err := cmd.Execute("cloud", "instance", "delete", "fakeInstanceID", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("Instance fakeInstanceID is being deleted"))
}

func (ms *MockSuite) TestCloudInstanceSetNameCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/compute/instance/fakeInstanceID",
		httpmock.NewStringResponder(200, instanceV2Response))

	// The PUT must keep the checksum (optimistic locking) and send back the
	// whole editable targetSpec with the new name; create-only fields
	// (location, sshKeyName, group) are not part of the update.
	httpmock.RegisterMatcherResponder(http.MethodPut,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/compute/instance/fakeInstanceID",
		tdhttpmock.JSONBody(td.JSON(`{
			"checksum": "abc123",
			"targetSpec": {
				"name": "renamed-instance",
				"flavor": { "id": "flavor-d2-2" },
				"image": { "id": "image-debian-12" },
				"networks": [
					{ "autoAssignPublicIp": true },
					{ "id": "private-net-id", "subnetId": "private-subnet-id", "ip": "10.1.2.155" }
				],
				"powerState": "ACTIVE",
				"securityGroups": [ { "id": "sg-default" } ],
				"volumes": [ { "id": "volume-1" } ]
			}
		}`)),
		httpmock.NewStringResponder(202, `{"id": "fakeInstanceID", "resourceStatus": "UPDATING"}`))

	out, err := cmd.Execute("cloud", "instance", "set-name", "fakeInstanceID", "renamed-instance", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("renamed to renamed-instance"))
}

func (ms *MockSuite) TestCloudInstanceSetNameWaitCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/compute/instance/fakeInstanceID",
		httpmock.NewStringResponder(200, instanceV2Response))
	httpmock.RegisterResponder(http.MethodPut,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/compute/instance/fakeInstanceID",
		httpmock.NewStringResponder(202, `{"id": "fakeInstanceID", "resourceStatus": "UPDATING"}`))

	out, err := cmd.Execute("cloud", "instance", "set-name", "fakeInstanceID", "renamed-instance", "--wait", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("renamed to renamed-instance"))
	assert.Cmp(httpmock.GetCallCountInfo()["GET https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/compute/instance/fakeInstanceID"], 2)
}
