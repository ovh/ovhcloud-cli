// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"encoding/json"
	"net/http"
	"strings"

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

const instanceV2Endpoint = "https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/compute/instance/fakeInstanceID"

// instanceV2ResponseWithPowerState returns instanceV2Response with both the
// target and current power states set to the given value.
func instanceV2ResponseWithPowerState(powerState string) string {
	return strings.ReplaceAll(instanceV2Response, `"powerState": "ACTIVE"`, `"powerState": "`+powerState+`"`)
}

// expectedInstancePutBody is the PUT body expected when only the power state
// of instanceV2Response changes.
func expectedInstancePutBody(powerState string) td.TestDeep {
	return td.JSON(`{
		"checksum": "abc123",
		"targetSpec": {
			"name": "my-instance",
			"flavor": { "id": "flavor-d2-2" },
			"image": { "id": "image-debian-12" },
			"networks": [
				{ "autoAssignPublicIp": true },
				{ "id": "private-net-id", "subnetId": "private-subnet-id", "ip": "10.1.2.155" }
			],
			"powerState": $1,
			"securityGroups": [ { "id": "sg-default" } ],
			"volumes": [ { "id": "volume-1" } ]
		}
	}`, powerState)
}

func (ms *MockSuite) TestCloudInstancePowerStateCmds(assert, require *td.T) {
	for _, tc := range []struct {
		command      string
		currentPower string
		targetPower  string
		message      string
	}{
		{"start", "SHUTOFF", "ACTIVE", "⚡️ Instance fakeInstanceID starting…"},
		{"stop", "ACTIVE", "SHUTOFF", "⚡️ Instance fakeInstanceID stopping…"},
		{"shelve", "ACTIVE", "SHELVED", "⚡️ Instance fakeInstanceID is being shelved…"},
		{"unshelve", "SHELVED", "ACTIVE", "⚡️ Instance fakeInstanceID is being unshelved…"},
	} {
		assert.RunAssertRequire(tc.command, func(assert, require *td.T) {
			defer cmd.PostExecute()

			httpmock.RegisterResponder(http.MethodGet, instanceV2Endpoint,
				httpmock.NewStringResponder(200, instanceV2ResponseWithPowerState(tc.currentPower)))
			httpmock.RegisterMatcherResponder(http.MethodPut, instanceV2Endpoint,
				tdhttpmock.JSONBody(expectedInstancePutBody(tc.targetPower)),
				httpmock.NewStringResponder(202, `{"id": "fakeInstanceID", "resourceStatus": "UPDATING"}`))

			out, err := cmd.Execute("cloud", "instance", tc.command, "fakeInstanceID", "--cloud-project", "fakeProjectID", "-o", "json")

			require.CmpNoError(err)
			// Exact match: the message must not carry unused format arguments.
			assert.Cmp(json.RawMessage(out), td.JSON(`{"message": $1}`, tc.message))
		})
	}
}

func (ms *MockSuite) TestCloudInstanceStartWaitCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, instanceV2Endpoint,
		httpmock.NewStringResponder(200, instanceV2Response))
	httpmock.RegisterResponder(http.MethodPut, instanceV2Endpoint,
		httpmock.NewStringResponder(202, `{"id": "fakeInstanceID", "resourceStatus": "UPDATING"}`))

	out, err := cmd.Execute("cloud", "instance", "start", "fakeInstanceID", "--wait", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("✅ Instance fakeInstanceID started"))
	assert.Cmp(httpmock.GetCallCountInfo()["GET "+instanceV2Endpoint], 2)
}

func (ms *MockSuite) TestCloudInstanceStopLockedCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, instanceV2Endpoint,
		httpmock.NewStringResponder(200, instanceV2Response))
	httpmock.RegisterResponder(http.MethodPut, instanceV2Endpoint,
		httpmock.NewStringResponder(409, `{"class": "Client::Conflict", "message": "instance is locked — unlock it before making changes: resource in invalid state"}`))

	_, err := cmd.Execute("cloud", "instance", "stop", "fakeInstanceID", "--cloud-project", "fakeProjectID")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("instance is locked"))
}

func (ms *MockSuite) TestCloudInstanceActionCmds(assert, require *td.T) {
	for _, tc := range []struct {
		name    string
		args    []string
		body    string
		message string
	}{
		{"reboot soft", []string{"reboot", "fakeInstanceID"}, `{"type": "REBOOT", "parameters": {"hard": false}}`, "⚡️ Instance fakeInstanceID is rebooting…"},
		{"reboot hard", []string{"reboot", "fakeInstanceID", "--type", "hard"}, `{"type": "REBOOT", "parameters": {"hard": true}}`, "⚡️ Instance fakeInstanceID is rebooting…"},
		{"rescue", []string{"reboot-rescue", "fakeInstanceID"}, `{"type": "RESCUE"}`, "⚡️ Instance fakeInstanceID is being rebooted in rescue mode…"},
		{"rescue with image", []string{"reboot-rescue", "fakeInstanceID", "--image", "rescue-image-id"}, `{"type": "RESCUE", "parameters": {"imageId": "rescue-image-id"}}`, "⚡️ Instance fakeInstanceID is being rebooted in rescue mode…"},
		{"exit rescue", []string{"exit-rescue", "fakeInstanceID"}, `{"type": "UNRESCUE"}`, "⚡️ Instance fakeInstanceID is exiting rescue mode…"},
		{"lock", []string{"lock", "fakeInstanceID"}, `{"type": "LOCK"}`, "⚡️ Instance fakeInstanceID is being locked…"},
		{"unlock", []string{"unlock", "fakeInstanceID"}, `{"type": "UNLOCK"}`, "⚡️ Instance fakeInstanceID is being unlocked…"},
	} {
		assert.RunAssertRequire(tc.name, func(assert, require *td.T) {
			defer cmd.PostExecute()

			httpmock.RegisterMatcherResponder(http.MethodPost, instanceV2Endpoint+"/action",
				tdhttpmock.JSONBody(td.JSON(tc.body)),
				httpmock.NewStringResponder(202, `{"id": "fakeInstanceID", "resourceStatus": "UPDATING"}`))

			args := append([]string{"cloud", "instance"}, tc.args...)
			out, err := cmd.Execute(append(args, "--cloud-project", "fakeProjectID", "-o", "json")...)

			require.CmpNoError(err)
			// Exact match: the message must not carry unused format arguments.
			assert.Cmp(json.RawMessage(out), td.JSON(`{"message": $1}`, tc.message))
		})
	}
}

func (ms *MockSuite) TestCloudInstanceRebootInvalidTypeCmd(assert, require *td.T) {
	_, err := cmd.Execute("cloud", "instance", "reboot", "fakeInstanceID", "--type", "warm", "--cloud-project", "fakeProjectID")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains(`invalid reboot type: "warm"`))
	assert.Cmp(httpmock.GetTotalCallCount(), 0)
}

const instanceV2CollectionEndpoint = "https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/compute/instance"

func (ms *MockSuite) TestCloudInstanceCreateCmd(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(http.MethodPost, instanceV2CollectionEndpoint,
		tdhttpmock.JSONBody(td.JSON(`{
			"targetSpec": {
				"name": "my-instance",
				"flavor": { "id": "flavor-d2-2" },
				"image": { "id": "image-debian-12" },
				"location": { "region": "GRA9", "availabilityZone": "gra9-a" },
				"networks": [
					{ "autoAssignPublicIp": true },
					{ "id": "private-net-id", "subnetId": "private-subnet-id", "ip": "10.1.2.155" }
				],
				"sshKeyName": "my-key",
				"userData": "I2Nsb3VkLWNvbmZpZw=="
			}
		}`)),
		httpmock.NewStringResponder(200, `{"id": "fakeInstanceID", "resourceStatus": "CREATING"}`).Once())

	out, err := cmd.Execute("cloud", "instance", "create", "GRA9",
		"--name", "my-instance",
		"--flavor", "flavor-d2-2",
		"--boot-from.image", "image-debian-12",
		"--availability-zone", "gra9-a",
		"--network.public",
		"--network.private.id", "private-net-id",
		"--network.private.subnet-id", "private-subnet-id",
		"--network.ip", "10.1.2.155",
		"--ssh-key.name", "my-key",
		"--user-data", "#cloud-config",
		"--cloud-project", "fakeProjectID", "-o", "json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{"message": "⚡️ Instance fakeInstanceID creation started", "details": {"id": "fakeInstanceID"}}`))
}

func (ms *MockSuite) TestCloudInstanceCreateWaitCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodPost, instanceV2CollectionEndpoint,
		httpmock.NewStringResponder(200, `{"id": "fakeInstanceID", "resourceStatus": "CREATING"}`).Once())
	httpmock.RegisterResponder(http.MethodGet, instanceV2Endpoint,
		httpmock.NewStringResponder(200, instanceV2Response))

	out, err := cmd.Execute("cloud", "instance", "create", "GRA11",
		"--name", "my-instance", "--flavor", "flavor-d2-2", "--boot-from.image", "image-debian-12",
		"--wait", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("✅ Instance fakeInstanceID created successfully"))
	assert.Cmp(httpmock.GetCallCountInfo()["GET "+instanceV2Endpoint], 1)
}

func (ms *MockSuite) TestCloudInstanceCreateBootVolumeCmd(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(http.MethodPost, instanceV2CollectionEndpoint,
		tdhttpmock.JSONBody(td.JSON(`{
			"targetSpec": {
				"name": "my-instance",
				"flavor": { "id": "flavor-d2-2" },
				"location": { "region": "GRA9" },
				"volumes": [ { "id": "boot-volume-id", "role": "BOOT" } ]
			}
		}`)),
		httpmock.NewStringResponder(200, `{"id": "fakeInstanceID", "resourceStatus": "CREATING"}`).Once())

	out, err := cmd.Execute("cloud", "instance", "create", "GRA9",
		"--name", "my-instance", "--flavor", "flavor-d2-2", "--boot-from.volume", "boot-volume-id",
		"--cloud-project", "fakeProjectID", "-o", "json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{"message": "⚡️ Instance fakeInstanceID creation started", "details": {"id": "fakeInstanceID"}}`))
}

func (ms *MockSuite) TestCloudInstanceCreatePublicIPCmd(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(http.MethodPost, instanceV2CollectionEndpoint,
		tdhttpmock.JSONBody(td.JSON(`{
			"targetSpec": {
				"name": "my-instance",
				"flavor": { "id": "flavor-d2-2" },
				"image": { "id": "image-debian-12" },
				"location": { "region": "GRA9" },
				"networks": [ { "ip": "203.0.113.10" } ]
			}
		}`)),
		httpmock.NewStringResponder(200, `{"id": "fakeInstanceID", "resourceStatus": "CREATING"}`).Once())

	out, err := cmd.Execute("cloud", "instance", "create", "GRA9",
		"--name", "my-instance", "--flavor", "flavor-d2-2", "--boot-from.image", "image-debian-12",
		"--network.ip", "203.0.113.10", "--cloud-project", "fakeProjectID", "-o", "json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{"message": "⚡️ Instance fakeInstanceID creation started", "details": {"id": "fakeInstanceID"}}`))
}

func (ms *MockSuite) TestCloudInstanceCreateSubnetOnlyCmd(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(http.MethodPost, instanceV2CollectionEndpoint,
		tdhttpmock.JSONBody(td.JSON(`{
			"targetSpec": {
				"name": "my-instance",
				"flavor": { "id": "flavor-d2-2" },
				"image": { "id": "image-debian-12" },
				"location": { "region": "GRA9" },
				"networks": [ { "subnetId": "private-subnet-id", "ip": "198.51.100.7" } ]
			}
		}`)),
		httpmock.NewStringResponder(200, `{"id": "fakeInstanceID", "resourceStatus": "CREATING"}`).Once())

	out, err := cmd.Execute("cloud", "instance", "create", "GRA9",
		"--name", "my-instance", "--flavor", "flavor-d2-2", "--boot-from.image", "image-debian-12",
		"--network.private.subnet-id", "private-subnet-id", "--network.ip", "198.51.100.7",
		"--cloud-project", "fakeProjectID", "-o", "json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{"message": "⚡️ Instance fakeInstanceID creation started", "details": {"id": "fakeInstanceID"}}`))
}

func (ms *MockSuite) TestCloudInstanceCreateImageAndBootVolumeCmd(assert, require *td.T) {
	_, err := cmd.Execute("cloud", "instance", "create", "GRA9",
		"--name", "my-instance", "--flavor", "flavor-d2-2",
		"--boot-from.image", "image-debian-12", "--boot-from.volume", "boot-volume-id",
		"--cloud-project", "fakeProjectID")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("[boot-from.image boot-from.volume] are set none of the others can be"))
	assert.Cmp(httpmock.GetTotalCallCount(), 0)
}

// The boot source, an image or a volume with role BOOT, may come from a
// parameters file, so it is left to the API, whose message says what is missing.
func (ms *MockSuite) TestCloudInstanceCreateMissingBootSourceCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodPost, instanceV2CollectionEndpoint,
		httpmock.NewStringResponder(400, `{"class": "Client::BadRequest", "message": "targetSpec.image is required, unless one volume of targetSpec.volumes has role BOOT"}`))

	_, err := cmd.Execute("cloud", "instance", "create", "GRA9",
		"--name", "my-instance", "--flavor", "flavor-d2-2", "--cloud-project", "fakeProjectID")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("targetSpec.image is required, unless one volume of targetSpec.volumes has role BOOT"))
}

func (ms *MockSuite) TestCloudInstanceCreateRemovedFlagCmd(assert, require *td.T) {
	// Flags of the v1 creation body that API v2 cannot express were removed.
	_, err := cmd.Execute("cloud", "instance", "create", "GRA9",
		"--name", "my-instance", "--flavor", "flavor-d2-2", "--boot-from.image", "image-debian-12",
		"--billing-period", "monthly", "--cloud-project", "fakeProjectID")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("unknown flag: --billing-period"))
	assert.Cmp(httpmock.GetTotalCallCount(), 0)
}

// expectedInstanceVolumesPutBody is the PUT body expected when only the
// volumes of instanceV2Response change.
func expectedInstanceVolumesPutBody(volumes string) td.TestDeep {
	return td.JSON(`{
		"checksum": "abc123",
		"targetSpec": {
			"name": "my-instance",
			"flavor": { "id": "flavor-d2-2" },
			"image": { "id": "image-debian-12" },
			"networks": [
				{ "autoAssignPublicIp": true },
				{ "id": "private-net-id", "subnetId": "private-subnet-id", "ip": "10.1.2.155" }
			],
			"powerState": "ACTIVE",
			"securityGroups": [ { "id": "sg-default" } ],
			"volumes": ` + volumes + `
		}
	}`)
}

func (ms *MockSuite) TestCloudStorageBlockVolumeAttachCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, instanceV2Endpoint,
		httpmock.NewStringResponder(200, instanceV2Response))
	httpmock.RegisterMatcherResponder(http.MethodPut, instanceV2Endpoint,
		tdhttpmock.JSONBody(expectedInstanceVolumesPutBody(`[ { "id": "volume-1" }, { "id": "volume-2" } ]`)),
		httpmock.NewStringResponder(202, `{"id": "fakeInstanceID", "resourceStatus": "UPDATING"}`))

	out, err := cmd.Execute("cloud", "storage", "block", "volume", "attach", "volume-2", "fakeInstanceID", "--cloud-project", "fakeProjectID", "-o", "json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{"message": "⚡️ Volume volume-2 is being attached to instance fakeInstanceID…"}`))
}

func (ms *MockSuite) TestCloudStorageBlockVolumeAttachWaitCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, instanceV2Endpoint,
		httpmock.NewStringResponder(200, instanceV2Response))
	httpmock.RegisterResponder(http.MethodPut, instanceV2Endpoint,
		httpmock.NewStringResponder(202, `{"id": "fakeInstanceID", "resourceStatus": "UPDATING"}`))

	out, err := cmd.Execute("cloud", "storage", "block", "volume", "attach", "volume-2", "fakeInstanceID", "--wait", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("✅ Volume volume-2 attached to instance fakeInstanceID"))
	assert.Cmp(httpmock.GetCallCountInfo()["GET "+instanceV2Endpoint], 2)
}

func (ms *MockSuite) TestCloudStorageBlockVolumeAttachAlreadyAttachedCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, instanceV2Endpoint,
		httpmock.NewStringResponder(200, instanceV2Response))

	_, err := cmd.Execute("cloud", "storage", "block", "volume", "attach", "volume-1", "fakeInstanceID", "--cloud-project", "fakeProjectID")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("volume volume-1 is already attached to instance fakeInstanceID"))
	assert.Cmp(httpmock.GetCallCountInfo()["PUT "+instanceV2Endpoint], 0)
}

func (ms *MockSuite) TestCloudStorageBlockVolumeDetachCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, instanceV2Endpoint,
		httpmock.NewStringResponder(200, instanceV2Response))
	httpmock.RegisterMatcherResponder(http.MethodPut, instanceV2Endpoint,
		tdhttpmock.JSONBody(expectedInstanceVolumesPutBody(`[]`)),
		httpmock.NewStringResponder(202, `{"id": "fakeInstanceID", "resourceStatus": "UPDATING"}`))

	out, err := cmd.Execute("cloud", "storage", "block", "volume", "detach", "volume-1", "fakeInstanceID", "--cloud-project", "fakeProjectID", "-o", "json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{"message": "⚡️ Volume volume-1 is being detached from instance fakeInstanceID…"}`))
}

func (ms *MockSuite) TestCloudStorageBlockVolumeDetachNotAttachedCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, instanceV2Endpoint,
		httpmock.NewStringResponder(200, instanceV2Response))

	_, err := cmd.Execute("cloud", "storage", "block", "volume", "detach", "volume-2", "fakeInstanceID", "--cloud-project", "fakeProjectID")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("volume volume-2 is not attached to instance fakeInstanceID"))
	assert.Cmp(httpmock.GetCallCountInfo()["PUT "+instanceV2Endpoint], 0)
}

const instanceGroupsURL = "https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/compute/instanceGroup"

func (ms *MockSuite) TestCloudInstanceGroupListCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, instanceGroupsURL,
		httpmock.NewStringResponder(200, `[
			{
				"id": "group-id-1",
				"resourceStatus": "READY",
				"targetSpec": {"name": "my-group", "policy": "ANTI_AFFINITY", "location": {"region": "GRA9"}},
				"currentState": {
					"name": "my-group",
					"policy": "ANTI_AFFINITY",
					"location": {"region": "GRA9"},
					"members": [{"id": "inst-1"}]
				}
			}
		]`).Once(),
	)

	out, err := cmd.Execute("cloud", "instance", "group", "ls", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	// "members" is the number of instances in the group
	assert.Cmp(out, td.Re(`group-id-1\s+│\s+my-group\s+│\s+ANTI_AFFINITY\s+│\s+GRA9\s+│\s+1\s+│\s+READY`))
}

func (ms *MockSuite) TestCloudInstanceGroupListCreatingCmd(assert, require *td.T) {
	// currentState is null while the group is being created: columns come from targetSpec
	httpmock.RegisterResponder(http.MethodGet, instanceGroupsURL,
		httpmock.NewStringResponder(200, `[
			{
				"id": "group-id-1",
				"resourceStatus": "CREATING",
				"targetSpec": {"name": "my-group", "policy": "AFFINITY", "location": {"region": "GRA9"}},
				"currentState": null
			}
		]`).Once(),
	)

	out, err := cmd.Execute("cloud", "instance", "group", "ls", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	// No member yet: the count is 0, not an empty cell
	assert.Cmp(out, td.Re(`group-id-1\s+│\s+my-group\s+│\s+AFFINITY\s+│\s+GRA9\s+│\s+0\s+│\s+CREATING`))
}

func (ms *MockSuite) TestCloudInstanceGroupListFilterCmd(assert, require *td.T) {
	// --filter runs on the raw API v2 object, so it selects the same rows in
	// the table and in JSON, where the computed "members" column does not exist
	groups := `[
		{
			"id": "group-affinity",
			"resourceStatus": "READY",
			"targetSpec": {"name": "g1", "policy": "AFFINITY", "location": {"region": "GRA9"}},
			"currentState": {"name": "g1", "policy": "AFFINITY", "location": {"region": "GRA9"}, "members": [{"id": "inst-1"}]}
		},
		{
			"id": "group-anti",
			"resourceStatus": "READY",
			"targetSpec": {"name": "g2", "policy": "ANTI_AFFINITY", "location": {"region": "GRA9"}},
			"currentState": {"name": "g2", "policy": "ANTI_AFFINITY", "location": {"region": "GRA9"}, "members": []}
		}
	]`
	httpmock.RegisterResponder(http.MethodGet, instanceGroupsURL, httpmock.NewStringResponder(200, groups))

	table, err := cmd.Execute("cloud", "instance", "group", "ls", "--filter", `targetSpec.policy == "AFFINITY"`, "--cloud-project", "fakeProjectID")
	require.CmpNoError(err)
	assert.Cmp(table, td.All(td.Contains("group-affinity"), td.Not(td.Contains("group-anti"))))

	out, err := cmd.Execute("cloud", "instance", "group", "ls", "--filter", `targetSpec.policy == "AFFINITY"`, "--cloud-project", "fakeProjectID", "-o", "json")
	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`[{"id": "group-affinity", "resourceStatus": $1, "targetSpec": $1, "currentState": $1}]`, td.Ignore()))

	// A custom output expression also gets the raw object: the computed
	// "members" column exists in the plain table only, so it prints nothing
	// (a real field such as "id" prints its value, see the JSON test above)
	out, err = cmd.Execute("cloud", "instance", "group", "ls", "--cloud-project", "fakeProjectID", "-o", "members")
	require.CmpNoError(err)
	assert.Cmp(strings.TrimSpace(out), "")
}

func (ms *MockSuite) TestCloudInstanceGroupListJSONCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, instanceGroupsURL,
		httpmock.NewStringResponder(200, `[
			{
				"id": "group-id-1",
				"resourceStatus": "READY",
				"targetSpec": {"name": "my-group", "policy": "ANTI_AFFINITY", "location": {"region": "GRA9"}},
				"currentState": {
					"name": "my-group",
					"policy": "ANTI_AFFINITY",
					"location": {"region": "GRA9"},
					"members": [{"id": "inst-1"}]
				}
			}
		]`).Once(),
	)

	out, err := cmd.Execute("cloud", "instance", "group", "ls", "--cloud-project", "fakeProjectID", "-o", "json")

	require.CmpNoError(err)
	// JSON output is the raw v2 object: no table column, no computed member count
	assert.Cmp(json.RawMessage(out), td.JSON(`[
		{
			"id": "group-id-1",
			"resourceStatus": "READY",
			"targetSpec": {"name": "my-group", "policy": "ANTI_AFFINITY", "location": {"region": "GRA9"}},
			"currentState": {
				"name": "my-group",
				"policy": "ANTI_AFFINITY",
				"location": {"region": "GRA9"},
				"members": [{"id": "inst-1"}]
			}
		}
	]`))
}

func (ms *MockSuite) TestCloudInstanceGroupGetCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, instanceGroupsURL+"/group-id-1",
		httpmock.NewStringResponder(200, `{
			"id": "group-id-1",
			"resourceStatus": "READY",
			"targetSpec": {"name": "my-group", "policy": "AFFINITY", "location": {"region": "GRA9"}},
			"currentState": {
				"name": "my-group",
				"policy": "AFFINITY",
				"location": {"region": "GRA9"},
				"members": [{"id": "inst-1"}, {"id": "inst-2"}]
			}
		}`).Once(),
	)

	out, err := cmd.Execute("cloud", "instance", "group", "get", "group-id-1", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Contains(out, "my-group")
	assert.Contains(out, "AFFINITY")
	assert.Contains(out, "GRA9")
	assert.Contains(out, "inst-1")
	assert.Contains(out, "inst-2")
}

func (ms *MockSuite) TestCloudInstanceGroupGetJSONCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, instanceGroupsURL+"/group-id-1",
		httpmock.NewStringResponder(200, `{
			"id": "group-id-1",
			"resourceStatus": "READY",
			"targetSpec": {"name": "my-group", "policy": "AFFINITY", "location": {"region": "GRA9"}},
			"currentState": {
				"name": "my-group",
				"policy": "AFFINITY",
				"location": {"region": "GRA9"},
				"members": [{"id": "inst-1"}]
			}
		}`).Once(),
	)

	out, err := cmd.Execute("cloud", "instance", "group", "get", "group-id-1", "--cloud-project", "fakeProjectID", "-o", "json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{
		"id": "group-id-1",
		"resourceStatus": "READY",
		"targetSpec": {"name": "my-group", "policy": "AFFINITY", "location": {"region": "GRA9"}},
		"currentState": {
			"name": "my-group",
			"policy": "AFFINITY",
			"location": {"region": "GRA9"},
			"members": [{"id": "inst-1"}]
		}
	}`))
}

func (ms *MockSuite) TestCloudInstanceGroupGetNotFoundCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, instanceGroupsURL+"/missing",
		httpmock.NewStringResponder(404, `{"class":"Client::NotFound","message":"instance group not found"}`).Once(),
	)

	_, err := cmd.Execute("cloud", "instance", "group", "get", "missing", "--cloud-project", "fakeProjectID")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("instance group not found"))
}

func (ms *MockSuite) TestCloudInstanceGroupGetCreatingCmd(assert, require *td.T) {
	// currentState is null while the group is being created: fall back on targetSpec
	httpmock.RegisterResponder(http.MethodGet, instanceGroupsURL+"/group-id-1",
		httpmock.NewStringResponder(200, `{
			"id": "group-id-1",
			"resourceStatus": "CREATING",
			"targetSpec": {"name": "my-group", "policy": "AFFINITY", "location": {"region": "GRA9"}},
			"currentState": null
		}`).Once(),
	)

	out, err := cmd.Execute("cloud", "instance", "group", "get", "group-id-1", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Contains(out, "my-group")
	assert.Contains(out, "CREATING")
	assert.Contains(out, "No member")
}

func (ms *MockSuite) TestCloudInstanceGroupCreateCmd(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(http.MethodPost, instanceGroupsURL,
		tdhttpmock.JSONBody(td.JSON(`{
			"targetSpec": {"name": "new-group", "location": {"region": "GRA9"}, "policy": "ANTI_AFFINITY"}
		}`)),
		httpmock.NewStringResponder(200, `{"id": "group-id-new", "resourceStatus": "CREATING"}`).Once(),
	)

	out, err := cmd.Execute("cloud", "instance", "group", "create", "new-group", "GRA9", "--type", "anti-affinity", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Contains(out, "group-id-new")
}

func (ms *MockSuite) TestCloudInstanceGroupCreateJSONCmd(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(http.MethodPost, instanceGroupsURL,
		tdhttpmock.JSONBody(td.JSON(`{
			"targetSpec": {"name": "new-group", "location": {"region": "GRA9"}, "policy": "ANTI_AFFINITY"}
		}`)),
		httpmock.NewStringResponder(200, `{"id": "group-id-new", "resourceStatus": "CREATING"}`).Once(),
	)

	out, err := cmd.Execute("cloud", "instance", "group", "create", "new-group", "GRA9", "--type", "anti-affinity", "--cloud-project", "fakeProjectID", "-o", "json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{
		"message": "⚡️ Instance group group-id-new creation started",
		"details": {"id": "group-id-new", "resourceStatus": "CREATING"}
	}`))
}

func (ms *MockSuite) TestCloudInstanceGroupCreateAPIErrorCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodPost, instanceGroupsURL,
		httpmock.NewStringResponder(400, `{"class":"Client::BadRequest","message":"invalid region"}`).Once(),
	)

	_, err := cmd.Execute("cloud", "instance", "group", "create", "new-group", "XX1", "--cloud-project", "fakeProjectID")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.All(td.Contains("error creating instance group"), td.Contains("invalid region")))
}

func (ms *MockSuite) TestCloudInstanceGroupCreateDefaultTypeCmd(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(http.MethodPost, instanceGroupsURL,
		tdhttpmock.JSONBody(td.JSON(`{
			"targetSpec": {"name": "new-group", "location": {"region": "GRA9"}, "policy": "AFFINITY"}
		}`)),
		httpmock.NewStringResponder(200, `{"id": "group-id-new", "resourceStatus": "CREATING"}`).Once(),
	)

	_, err := cmd.Execute("cloud", "instance", "group", "create", "new-group", "GRA9", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
}

func (ms *MockSuite) TestCloudInstanceGroupCreateInvalidTypeCmd(assert, require *td.T) {
	// No responder: the command must fail before calling the API
	_, err := cmd.Execute("cloud", "instance", "group", "create", "new-group", "GRA9", "--type", "soft", "--cloud-project", "fakeProjectID")

	require.CmpError(err)
	assert.Contains(err.Error(), "invalid group type")
	assert.Cmp(httpmock.GetTotalCallCount(), 0)
}

func (ms *MockSuite) TestCloudInstanceGroupCreateUndocumentedTypeCmd(assert, require *td.T) {
	// Only the two documented spellings are accepted, as in API v1
	for _, groupType := range []string{"ANTI_AFFINITY", "anti_affinity", "Affinity"} {
		_, err := cmd.Execute("cloud", "instance", "group", "create", "new-group", "GRA9", "--type", groupType, "--cloud-project", "fakeProjectID")

		require.CmpError(err, "type %q", groupType)
		assert.Contains(err.Error(), "invalid group type")
	}
	assert.Cmp(httpmock.GetTotalCallCount(), 0)
}

func (ms *MockSuite) TestCloudInstanceGroupCreateNoIDCmd(assert, require *td.T) {
	// A 202 without an ID must not be reported as a success, nor polled
	httpmock.RegisterResponder(http.MethodPost, instanceGroupsURL,
		httpmock.NewStringResponder(202, `{"resourceStatus": "CREATING"}`).Once(),
	)

	_, err := cmd.Execute("cloud", "instance", "group", "create", "new-group", "GRA9", "--wait", "--cloud-project", "fakeProjectID")

	require.CmpError(err)
	assert.Contains(err.Error(), "no ID in the API response")
	// Only the POST was sent: nothing was polled
	assert.Cmp(httpmock.GetTotalCallCount(), 1)
}

func (ms *MockSuite) TestCloudInstanceGroupCreateWaitErrorCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodPost, instanceGroupsURL,
		httpmock.NewStringResponder(202, `{"id": "group-id-new", "resourceStatus": "CREATING"}`).Once(),
	)
	httpmock.RegisterResponder(http.MethodGet, instanceGroupsURL+"/group-id-new",
		httpmock.NewStringResponder(200, `{"id": "group-id-new", "resourceStatus": "ERROR"}`).Once(),
	)

	out, err := cmd.Execute("cloud", "instance", "group", "create", "new-group", "GRA9", "--wait", "--cloud-project", "fakeProjectID")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.All(td.Contains("failed to wait for instance group group-id-new creation"), td.Contains("ERROR")))
	assert.Cmp(out, td.Not(td.Contains("created successfully")))
}

func (ms *MockSuite) TestCloudInstanceGroupDeleteCmd(assert, require *td.T) {
	// The API answers 202 with the group as it is after the request
	httpmock.RegisterResponder(http.MethodDelete, instanceGroupsURL+"/group-id-1",
		httpmock.NewStringResponder(202, `{"id": "group-id-1", "resourceStatus": "DELETING"}`).Once(),
	)

	out, err := cmd.Execute("cloud", "instance", "group", "delete", "group-id-1", "--cloud-project", "fakeProjectID", "-o", "json")

	require.CmpNoError(err)
	// details carries the API answer, so a script can read resourceStatus
	assert.Cmp(json.RawMessage(out), td.JSON(`{
		"message": "⚡️ Instance group group-id-1 deletion started",
		"details": {"id": "group-id-1", "resourceStatus": "DELETING"}
	}`))
	// Without --wait, the command does not poll the group
	assert.Cmp(httpmock.GetCallCountInfo()["GET "+instanceGroupsURL+"/group-id-1"], 0)
}

func (ms *MockSuite) TestCloudInstanceGroupDeleteWaitCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodDelete, instanceGroupsURL+"/group-id-1",
		httpmock.NewStringResponder(202, ``).Once(),
	)
	// The group is gone once the API answers 404
	httpmock.RegisterResponder(http.MethodGet, instanceGroupsURL+"/group-id-1",
		httpmock.NewStringResponder(404, `{"class":"Client::NotFound","message":"instance group not found"}`).Once(),
	)

	out, err := cmd.Execute("cloud", "instance", "group", "delete", "group-id-1", "--wait", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Contains(out, "✅ Instance group group-id-1 deleted successfully")
}

func (ms *MockSuite) TestCloudInstanceGroupDeleteWaitErrorCmd(assert, require *td.T) {
	// A group that was already in ERROR before the DELETE still reports ERROR on
	// the first poll, then DELETING, then is gone: that is a success.
	httpmock.RegisterResponder(http.MethodDelete, instanceGroupsURL+"/group-id-1",
		httpmock.NewStringResponder(202, `{"id": "group-id-1", "resourceStatus": "ERROR"}`).Once(),
	)
	httpmock.RegisterResponder(http.MethodGet, instanceGroupsURL+"/group-id-1",
		httpmock.ResponderFromMultipleResponses([]*http.Response{
			httpmock.NewStringResponse(200, `{"id": "group-id-1", "resourceStatus": "ERROR"}`),
			httpmock.NewStringResponse(200, `{"id": "group-id-1", "resourceStatus": "DELETING"}`),
			httpmock.NewStringResponse(404, `{"class":"Client::NotFound","message":"instance group not found"}`),
		}),
	)

	out, err := cmd.Execute("cloud", "instance", "group", "delete", "group-id-1", "--wait", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Contains(out, "✅ Instance group group-id-1 deleted successfully")
	assert.Cmp(httpmock.GetCallCountInfo()["GET "+instanceGroupsURL+"/group-id-1"], 3)
}

func (ms *MockSuite) TestCloudInstanceGroupDeleteWaitFailedCmd(assert, require *td.T) {
	// ERROR after DELETING means the deletion itself failed
	httpmock.RegisterResponder(http.MethodDelete, instanceGroupsURL+"/group-id-1",
		httpmock.NewStringResponder(202, `{"id": "group-id-1", "resourceStatus": "DELETING"}`).Once(),
	)
	httpmock.RegisterResponder(http.MethodGet, instanceGroupsURL+"/group-id-1",
		httpmock.ResponderFromMultipleResponses([]*http.Response{
			httpmock.NewStringResponse(200, `{"id": "group-id-1", "resourceStatus": "DELETING"}`),
			httpmock.NewStringResponse(200, `{"id": "group-id-1", "resourceStatus": "ERROR"}`),
		}),
	)

	_, err := cmd.Execute("cloud", "instance", "group", "delete", "group-id-1", "--wait", "--cloud-project", "fakeProjectID")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.All(td.Contains("failed to wait for instance group group-id-1 deletion"), td.Contains("deletion failed")))
}

func (ms *MockSuite) TestCloudInstanceGroupDeleteWaitFailedReasonCmd(assert, require *td.T) {
	// When the backend tasks say why the deletion failed, the error says it too
	httpmock.RegisterResponder(http.MethodDelete, instanceGroupsURL+"/group-id-1",
		httpmock.NewStringResponder(202, `{"id": "group-id-1", "resourceStatus": "DELETING"}`).Once(),
	)
	httpmock.RegisterResponder(http.MethodGet, instanceGroupsURL+"/group-id-1",
		httpmock.ResponderFromMultipleResponses([]*http.Response{
			httpmock.NewStringResponse(200, `{"id": "group-id-1", "resourceStatus": "DELETING"}`),
			httpmock.NewStringResponse(200, `{
				"id": "group-id-1",
				"resourceStatus": "ERROR",
				"currentTasks": [{"type": "DELETE", "status": "ERROR", "errors": [{"message": "group still has members"}]}]
			}`),
		}),
	)

	_, err := cmd.Execute("cloud", "instance", "group", "delete", "group-id-1", "--wait", "--cloud-project", "fakeProjectID")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("deletion failed (DELETE: group still has members)"))
}

func (ms *MockSuite) TestCloudInstanceGroupDeleteWaitNeverStartsCmd(assert, require *td.T) {
	// ERROR that never turns into DELETING: give up after a few polls, not after the full timeout
	httpmock.RegisterResponder(http.MethodDelete, instanceGroupsURL+"/group-id-1",
		httpmock.NewStringResponder(202, `{"id": "group-id-1", "resourceStatus": "ERROR"}`).Once(),
	)
	httpmock.RegisterResponder(http.MethodGet, instanceGroupsURL+"/group-id-1",
		httpmock.NewStringResponder(200, `{"id": "group-id-1", "resourceStatus": "ERROR"}`),
	)

	_, err := cmd.Execute("cloud", "instance", "group", "delete", "group-id-1", "--wait", "--cloud-project", "fakeProjectID")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.All(td.Contains("failed to wait for instance group group-id-1 deletion"), td.Contains("deletion did not start")))
	assert.Cmp(httpmock.GetCallCountInfo()["GET "+instanceGroupsURL+"/group-id-1"], 3)
}

func (ms *MockSuite) TestCloudInstanceGroupCreateWaitCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodPost, instanceGroupsURL,
		httpmock.NewStringResponder(202, `{"id": "group-id-new", "resourceStatus": "CREATING"}`).Once(),
	)
	httpmock.RegisterResponder(http.MethodGet, instanceGroupsURL+"/group-id-new",
		httpmock.NewStringResponder(200, `{
			"id": "group-id-new",
			"resourceStatus": "READY",
			"targetSpec": {"name": "new-group", "policy": "AFFINITY", "location": {"region": "GRA9"}},
			"currentState": {"name": "new-group", "policy": "AFFINITY", "location": {"region": "GRA9"}, "members": []}
		}`).Once(),
	)

	out, err := cmd.Execute("cloud", "instance", "group", "create", "new-group", "GRA9", "--wait", "--cloud-project", "fakeProjectID", "-o", "json")

	require.CmpNoError(err)
	// details is the group as fetched once READY, not the POST response
	assert.Cmp(json.RawMessage(out), td.JSON(`{
		"message": "✅ Instance group group-id-new created successfully",
		"details": {"id": "group-id-new", "resourceStatus": "READY", "targetSpec": $1, "currentState": $1}
	}`, td.Ignore()))
}

const (
	instanceSnapshotsURL = "https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/compute/snapshot"
	instancesV2URL       = "https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/compute/instance"
)

func registerInstanceV2Region(region string) {
	httpmock.RegisterResponder(http.MethodGet, instancesV2URL+"/fakeInstanceID",
		httpmock.NewStringResponder(200, `{
			"id": "fakeInstanceID",
			"resourceStatus": "READY",
			"targetSpec": {"name": "my-instance", "location": {"region": "`+region+`"}}
		}`))
}

func (ms *MockSuite) TestCloudInstanceSnapshotCreateCmd(assert, require *td.T) {
	registerInstanceV2Region("GRA11")
	httpmock.RegisterMatcherResponder(http.MethodPost, instanceSnapshotsURL,
		tdhttpmock.JSONBody(td.JSON(`{
			"targetSpec": {"name": "my-snap", "instance": {"id": "fakeInstanceID"}, "location": {"region": "GRA11"}}
		}`)),
		httpmock.NewStringResponder(202, `{"id": "snap-001", "resourceStatus": "CREATING"}`).Once(),
	)

	out, err := cmd.Execute("cloud", "instance", "snapshot", "create", "fakeInstanceID", "my-snap", "--cloud-project", "fakeProjectID", "-o", "json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{
		"message": "⚡️ Snapshot snap-001 creation started",
		"details": {"id": "snap-001", "resourceStatus": "CREATING"}
	}`))
	// Without --wait, the command does not poll the snapshot
	assert.Cmp(httpmock.GetCallCountInfo()["GET "+instanceSnapshotsURL+"/snap-001"], 0)
}

func (ms *MockSuite) TestCloudInstanceSnapshotCreateWaitCmd(assert, require *td.T) {
	registerInstanceV2Region("GRA11")
	httpmock.RegisterResponder(http.MethodPost, instanceSnapshotsURL,
		httpmock.NewStringResponder(202, `{"id": "snap-001", "resourceStatus": "CREATING"}`).Once(),
	)
	httpmock.RegisterResponder(http.MethodGet, instanceSnapshotsURL+"/snap-001",
		httpmock.NewStringResponder(200, `{
			"id": "snap-001",
			"resourceStatus": "READY",
			"targetSpec": {"name": "my-snap", "instance": {"id": "fakeInstanceID"}, "location": {"region": "GRA11"}},
			"currentState": {"name": "my-snap", "status": "ACTIVE"}
		}`).Once(),
	)

	out, err := cmd.Execute("cloud", "instance", "snapshot", "create", "fakeInstanceID", "my-snap", "--wait", "--cloud-project", "fakeProjectID", "-o", "json")

	require.CmpNoError(err)
	// details is the snapshot as fetched once READY, not the POST response
	assert.Cmp(json.RawMessage(out), td.JSON(`{
		"message": "✅ Snapshot snap-001 created successfully",
		"details": {"id": "snap-001", "resourceStatus": "READY", "targetSpec": $1, "currentState": $1}
	}`, td.Ignore()))
}

func (ms *MockSuite) TestCloudInstanceSnapshotCreateAPIErrorCmd(assert, require *td.T) {
	registerInstanceV2Region("GRA11")
	httpmock.RegisterResponder(http.MethodPost, instanceSnapshotsURL,
		httpmock.NewStringResponder(400, `{"class":"Client::BadRequest","message":"instance is not active"}`).Once(),
	)

	_, err := cmd.Execute("cloud", "instance", "snapshot", "create", "fakeInstanceID", "my-snap", "--cloud-project", "fakeProjectID")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.All(td.Contains("error creating snapshot for instance"), td.Contains("instance is not active")))
}

func (ms *MockSuite) TestCloudInstanceSnapshotCreateDistantCmd(assert, require *td.T) {
	// A cross-region snapshot has no API v2 equivalent and stays on API v1
	registerInstanceV2Region("GRA11")
	httpmock.RegisterMatcherResponder(http.MethodPost,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/GRA11/instance/fakeInstanceID/snapshot",
		tdhttpmock.JSONBody(td.JSON(`{"snapshotName": "my-snap", "distantSnapshotName": "my-distant-snap", "distantRegionName": "SBG5"}`)),
		httpmock.NewStringResponder(200, `{"imageId": "img-001"}`).Once(),
	)

	out, err := cmd.Execute("cloud", "instance", "snapshot", "create", "fakeInstanceID", "my-snap",
		"--distant-region-name", "SBG5", "--distant-snapshot-name", "my-distant-snap", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Contains(out, "Snapshot created successfully with ID: img-001")
}

func (ms *MockSuite) TestCloudInstanceSnapshotAbortCmd(assert, require *td.T) {
	registerInstanceV2Region("GRA11")
	httpmock.RegisterResponder(http.MethodPost,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/GRA11/instance/fakeInstanceID/abortSnapshot",
		httpmock.NewStringResponder(200, `null`).Once(),
	)

	out, err := cmd.Execute("cloud", "instance", "snapshot", "abort", "fakeInstanceID", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Contains(out, "Snapshot aborted successfully")
}

func (ms *MockSuite) TestCloudInstanceSnapshotListCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, instanceSnapshotsURL,
		httpmock.NewStringResponder(200, `[
			{
				"id": "snap-001",
				"resourceStatus": "READY",
				"targetSpec": {"name": "my-snap", "instance": {"id": "fakeInstanceID"}, "location": {"region": "GRA11"}},
				"currentState": {
					"name": "my-snap",
					"instance": {"id": "fakeInstanceID"},
					"location": {"region": "GRA11"},
					"status": "ACTIVE"
				}
			}
		]`).Once(),
	)

	out, err := cmd.Execute("cloud", "instance", "snapshot", "ls", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.Re(`snap-001\s+│\s+my-snap\s+│\s+fakeInstanceID\s+│\s+GRA11\s+│\s+ACTIVE\s+│\s+READY`))
}

func (ms *MockSuite) TestCloudInstanceSnapshotListCreatingCmd(assert, require *td.T) {
	// currentState is null while the snapshot is being created: columns come from targetSpec
	httpmock.RegisterResponder(http.MethodGet, instanceSnapshotsURL,
		httpmock.NewStringResponder(200, `[
			{
				"id": "snap-001",
				"resourceStatus": "CREATING",
				"targetSpec": {"name": "my-snap", "instance": {"id": "fakeInstanceID"}, "location": {"region": "GRA11"}},
				"currentState": null
			}
		]`).Once(),
	)

	out, err := cmd.Execute("cloud", "instance", "snapshot", "ls", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.Re(`snap-001\s+│\s+my-snap\s+│\s+fakeInstanceID\s+│\s+GRA11\s+│\s+│\s+CREATING`))
}

func (ms *MockSuite) TestCloudInstanceSnapshotGetCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, instanceSnapshotsURL+"/snap-001",
		httpmock.NewStringResponder(200, `{"id": "snap-001", "resourceStatus": "READY", "currentState": {"name": "my-snap"}}`).Once(),
	)

	out, err := cmd.Execute("cloud", "instance", "snapshot", "get", "snap-001", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{"id": "snap-001", "resourceStatus": "READY", "currentState": {"name": "my-snap"}}`))
}

func (ms *MockSuite) TestCloudInstanceSnapshotDeleteCmd(assert, require *td.T) {
	// The API answers 202 with the snapshot as it is after the request
	httpmock.RegisterResponder(http.MethodDelete, instanceSnapshotsURL+"/snap-001",
		httpmock.NewStringResponder(202, `{"id": "snap-001", "resourceStatus": "DELETING"}`).Once(),
	)

	out, err := cmd.Execute("cloud", "instance", "snapshot", "delete", "snap-001", "--cloud-project", "fakeProjectID", "-o", "json")

	require.CmpNoError(err)
	// details carries the API answer, so a script can read resourceStatus
	assert.Cmp(json.RawMessage(out), td.JSON(`{
		"message": "⚡️ Snapshot snap-001 deletion started",
		"details": {"id": "snap-001", "resourceStatus": "DELETING"}
	}`))
	// Without --wait, the command does not poll the snapshot
	assert.Cmp(httpmock.GetCallCountInfo()["GET "+instanceSnapshotsURL+"/snap-001"], 0)
}

func (ms *MockSuite) TestCloudInstanceSnapshotCreateNoIDCmd(assert, require *td.T) {
	// A 202 without an ID must not be reported as a success, nor polled
	registerInstanceV2Region("GRA11")
	httpmock.RegisterResponder(http.MethodPost, instanceSnapshotsURL,
		httpmock.NewStringResponder(202, `{"resourceStatus": "CREATING"}`).Once(),
	)

	_, err := cmd.Execute("cloud", "instance", "snapshot", "create", "fakeInstanceID", "my-snap", "--wait", "--cloud-project", "fakeProjectID")

	require.CmpError(err)
	assert.Contains(err.Error(), "no ID in the API response")
	assert.Cmp(httpmock.GetCallCountInfo()["GET "+instanceSnapshotsURL+"/"], 0)
}

func (ms *MockSuite) TestCloudInstanceSnapshotDeleteWaitCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodDelete, instanceSnapshotsURL+"/snap-001",
		httpmock.NewStringResponder(202, ``).Once(),
	)
	// The snapshot is gone once the API answers 404
	httpmock.RegisterResponder(http.MethodGet, instanceSnapshotsURL+"/snap-001",
		httpmock.NewStringResponder(404, `{"class":"Client::NotFound","message":"snapshot not found"}`).Once(),
	)

	out, err := cmd.Execute("cloud", "instance", "snapshot", "delete", "snap-001", "--wait", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Contains(out, "✅ Snapshot snap-001 deleted successfully")
}

func (ms *MockSuite) TestCloudInstanceSnapshotDeleteAPIErrorCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodDelete, instanceSnapshotsURL+"/missing",
		httpmock.NewStringResponder(404, `{"class":"Client::NotFound","message":"snapshot not found"}`).Once(),
	)

	_, err := cmd.Execute("cloud", "instance", "snapshot", "delete", "missing", "--cloud-project", "fakeProjectID")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.All(td.Contains("error deleting snapshot"), td.Contains("snapshot not found")))
}
