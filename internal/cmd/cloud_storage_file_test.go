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

const shareSnapshotListResponse = `[
	{
		"id": "snap-1",
		"resourceStatus": "READY",
		"targetSpec": {
			"name": "first-snapshot",
			"share": { "id": "share-1" }
		}
	},
	{
		"id": "snap-2",
		"resourceStatus": "READY",
		"targetSpec": {
			"name": "second-snapshot",
			"share": { "id": "share-2" }
		}
	}
]`

func (ms *MockSuite) TestCloudStorageFileSnapshotListCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/storage/file/snapshot",
		httpmock.NewStringResponder(200, shareSnapshotListResponse))

	out, err := cmd.Execute("cloud", "storage", "file", "snapshot", "list", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("snap-1"))
	assert.Cmp(out, td.Contains("snap-2"))
}

func (ms *MockSuite) TestCloudStorageFileSnapshotListFilteredByShareCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/storage/file/snapshot",
		httpmock.NewStringResponder(200, shareSnapshotListResponse))

	out, err := cmd.Execute("cloud", "storage", "file", "snapshot", "list", "--share-id", "share-1", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("snap-1"))
	assert.Cmp(out, td.Not(td.Contains("snap-2")))
}

func (ms *MockSuite) TestCloudStorageFileSnapshotGetCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/storage/file/snapshot/snap-1",
		httpmock.NewStringResponder(200, `{
			"id": "snap-1",
			"resourceStatus": "READY",
			"targetSpec": {
				"name": "first-snapshot",
				"share": { "id": "share-1" }
			},
			"currentState": {
				"name": "first-snapshot",
				"share": { "id": "share-1" },
				"location": { "region": "EU-WEST-PAR" },
				"size": 10
			}
		}`))

	out, err := cmd.Execute("cloud", "storage", "file", "snapshot", "get", "snap-1", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("first-snapshot"))
	assert.Cmp(out, td.Contains("share-1"))
}

func (ms *MockSuite) TestCloudStorageFileSnapshotDeleteCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodDelete,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/storage/file/snapshot/snap-1",
		httpmock.NewStringResponder(200, ``))

	out, err := cmd.Execute("cloud", "storage", "file", "snapshot", "delete", "snap-1", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("Snapshot snap-1 is being deleted"))
}
