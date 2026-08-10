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
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/volume",
		httpmock.NewStringResponder(200, `[
			{
				"id": "vol-1",
				"name": "test-volume",
				"region": "GRA9",
				"type": "high-speed-gen2",
				"status": "available"
			}
		]`))

	out, err := cmd.Execute("cloud", "storage", "block", "volume", "list", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("vol-1"))
	assert.Cmp(out, td.Contains("test-volume"))
}

func (ms *MockSuite) TestCloudStorageBlockGetCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/volume/vol-1",
		httpmock.NewStringResponder(200, `{
			"id": "vol-1",
			"name": "test-volume",
			"region": "GRA9",
			"type": "high-speed-gen2",
			"status": "available",
			"size": 50
		}`))

	out, err := cmd.Execute("cloud", "storage", "block", "volume", "get", "vol-1", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("vol-1"))
	assert.Cmp(out, td.Contains("test-volume"))
}

func (ms *MockSuite) TestCloudStorageBlockDeleteCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodDelete,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/volume/vol-1",
		httpmock.NewStringResponder(200, ``))

	out, err := cmd.Execute("cloud", "storage", "block", "volume", "delete", "vol-1", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("deleted successfully"))
}

func (ms *MockSuite) TestCloudStorageBlockSnapshotListCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/volume/snapshot",
		httpmock.NewStringResponder(200, `[
			{
				"id": "snap-1",
				"name": "test-snapshot",
				"region": "GRA9",
				"description": "test",
				"status": "available"
			}
		]`))

	out, err := cmd.Execute("cloud", "storage", "block", "volume", "snapshot", "list", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("snap-1"))
}

func (ms *MockSuite) TestCloudStorageBlockBackupListCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region",
		httpmock.NewStringResponder(200, `["GRA9"]`))

	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/GRA9",
		httpmock.NewStringResponder(200, `{
			"name": "GRA9",
			"type": "region",
			"status": "UP",
			"services": [{"name": "volume", "status": "UP"}]
		}`))

	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/GRA9/volumeBackup",
		httpmock.NewStringResponder(200, `[
			{
				"id": "backup-1",
				"name": "test-backup",
				"region": "GRA9",
				"status": "ok"
			}
		]`))

	out, err := cmd.Execute("cloud", "storage", "block", "volume", "backup", "list", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("backup-1"))
}
