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

func (ms *MockSuite) TestCloudStorageFileShareListCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/storage/file/share",
		httpmock.NewStringResponder(200, `[
			{
				"id": "share-1234",
				"resourceStatus": "READY",
				"currentState": {
					"name": "my-share",
					"protocol": "NFS",
					"size": 100,
					"location": {"region": "GRA1"}
				}
			}
		]`),
	)

	out, err := cmd.Execute("cloud", "storage", "file", "share", "list", "--cloud-project", "fakeProjectID")
	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("my-share"))
	assert.Cmp(out, td.Contains("GRA1"))
}

func (ms *MockSuite) TestCloudStorageFileShareCreateCmd(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(http.MethodPost,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/storage/file/share",
		tdhttpmock.JSONBody(td.JSON(`
			{
				"targetSpec": {
					"name": "my-share",
					"protocol": "NFS",
					"shareType": "STANDARD_1AZ",
					"size": 100,
					"location": {"region": "GRA1"}
				}
			}`),
		),
		httpmock.NewStringResponder(200, `{"id": "share-1234"}`),
	)

	out, err := cmd.Execute("cloud", "storage", "file", "share", "create",
		"--cloud-project", "fakeProjectID",
		"--name", "my-share",
		"--protocol", "NFS",
		"--share-type", "STANDARD_1AZ",
		"--size", "100",
		"--region", "GRA1",
	)
	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("share-1234"))
}

func (ms *MockSuite) TestCloudStorageFileShareCreateCmdError(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodPost,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/storage/file/share",
		httpmock.NewStringResponder(400, `{"class": "Client::BadRequest::InvalidParameter", "message": "Invalid parameter in the request"}`),
	)

	_, err := cmd.Execute("cloud", "storage", "file", "share", "create",
		"--cloud-project", "fakeProjectID",
		"--name", "my-share",
		"--protocol", "NFS",
		"--share-type", "STANDARD_1AZ",
		"--size", "100",
		"--region", "GRA1",
	)
	assert.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("failed to create file storage share"))
}

func (ms *MockSuite) TestCloudStorageFileSnapshotCreateCmd(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(http.MethodPost,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/storage/file/snapshot",
		tdhttpmock.JSONBody(td.JSON(`
			{
				"targetSpec": {
					"name": "my-snapshot",
					"share": {"id": "share-1234"}
				}
			}`),
		),
		httpmock.NewStringResponder(200, `{"id": "snapshot-5678"}`),
	)

	out, err := cmd.Execute("cloud", "storage", "file", "snapshot", "create",
		"--cloud-project", "fakeProjectID",
		"--name", "my-snapshot",
		"--share-id", "share-1234",
	)
	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("snapshot-5678"))
}

func (ms *MockSuite) TestCloudStorageFileNetworkGetCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/storage/file/network/network-9012",
		httpmock.NewStringResponder(200, `{
			"id": "network-9012",
			"resourceStatus": "READY",
			"currentState": {
				"name": "my-share-network",
				"location": {"region": "GRA1"},
				"network": {"id": "priv-net-1"},
				"subnet": {"id": "subnet-1"}
			}
		}`),
	)

	out, err := cmd.Execute("cloud", "storage", "file", "network", "get", "network-9012", "--cloud-project", "fakeProjectID")
	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("my-share-network"))
}

func (ms *MockSuite) TestCloudStorageFileNetworkDeleteCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodDelete,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/storage/file/network/network-9012",
		httpmock.NewStringResponder(204, ``),
	)

	out, err := cmd.Execute("cloud", "storage", "file", "network", "delete", "network-9012", "--cloud-project", "fakeProjectID")
	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("network-9012"))
}
