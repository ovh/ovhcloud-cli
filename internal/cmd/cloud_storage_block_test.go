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

func (ms *MockSuite) TestCloudStorageBlockListCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/storage/block/volume",
		httpmock.NewStringResponder(200, `[
			{
				"id": "vol-1",
				"resourceStatus": "READY",
				"currentState": {
					"name": "test-volume",
					"size": 50,
					"volumeType": "HIGH_SPEED_GEN2",
					"status": "AVAILABLE",
					"location": { "region": "EU-WEST-PAR" }
				}
			}
		]`))

	out, err := cmd.Execute("cloud", "storage", "block", "list", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("vol-1"))
	assert.Cmp(out, td.Contains("test-volume"))
}

func (ms *MockSuite) TestCloudStorageBlockGetCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/storage/block/volume/vol-1",
		httpmock.NewStringResponder(200, `{
			"id": "vol-1",
			"resourceStatus": "READY",
			"currentState": {
				"name": "test-volume",
				"size": 50,
				"volumeType": "HIGH_SPEED_GEN2",
				"status": "AVAILABLE",
				"location": { "region": "EU-WEST-PAR" }
			},
			"targetSpec": {
				"name": "test-volume",
				"size": 50,
				"volumeType": "HIGH_SPEED_GEN2"
			}
		}`))

	out, err := cmd.Execute("cloud", "storage", "block", "get", "vol-1", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("vol-1"))
	assert.Cmp(out, td.Contains("test-volume"))
	// Optional fields absent from the response (description, availabilityZone,
	// attachedInstances) must not leak Go's "<no value>" placeholder.
	assert.Cmp(out, td.Not(td.Contains("<no value>")))
}

func (ms *MockSuite) TestCloudStorageBlockDeleteCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodDelete,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/storage/block/volume/vol-1",
		httpmock.NewStringResponder(200, ``))

	out, err := cmd.Execute("cloud", "storage", "block", "delete", "vol-1", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("deleted successfully"))
}

func (ms *MockSuite) TestCloudStorageBlockSnapshotListCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/storage/block/snapshot",
		httpmock.NewStringResponder(200, `[
			{
				"id": "snap-1",
				"resourceStatus": "READY",
				"currentState": {
					"name": "test-snapshot",
					"size": 10,
					"volumeId": "vol-1",
					"location": { "region": "EU-WEST-PAR" }
				}
			}
		]`))

	out, err := cmd.Execute("cloud", "storage", "block", "snapshot", "list", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("snap-1"))
}

func (ms *MockSuite) TestCloudStorageBlockBackupListCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/storage/block/backup",
		httpmock.NewStringResponder(200, `[
			{
				"id": "backup-1",
				"resourceStatus": "READY",
				"currentState": {
					"name": "test-backup",
					"size": 10,
					"volumeId": "vol-1",
					"location": { "region": "EU-WEST-PAR" }
				}
			}
		]`))

	out, err := cmd.Execute("cloud", "storage", "block", "backup", "list", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("backup-1"))
}
