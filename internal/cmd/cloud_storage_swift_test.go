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

const swiftContainerURL = "https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/storage/swift/container"

func (ms *MockSuite) TestCloudStorageSwiftContainerListCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, swiftContainerURL,
		httpmock.NewStringResponder(200, `[
			{
				"id": "GRA_my-container",
				"resourceStatus": "READY",
				"targetSpec": {
					"name": "my-container",
					"location": { "region": "GRA" },
					"type": "PUBLIC"
				}
			}
		]`))

	out, err := cmd.Execute("cloud", "storage", "swift", "container", "list", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.Re(`GRA_my-container\s+\S+\s+my-container\s+\S+\s+GRA\s+\S+\s+PUBLIC\s+\S+\s+READY`))
}

func (ms *MockSuite) TestCloudStorageSwiftContainerGetCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, swiftContainerURL+"/GRA_my-container",
		httpmock.NewStringResponder(200, `{
			"id": "GRA_my-container",
			"checksum": "abc123",
			"resourceStatus": "READY",
			"createdAt": "2026-10-20T10:00:00Z",
			"updatedAt": "2026-10-20T10:05:00Z",
			"currentTasks": [],
			"targetSpec": {
				"name": "my-container",
				"location": { "region": "GRA" },
				"archive": false,
				"type": "PUBLIC",
				"cors": ["https://a.example.com", "https://b.example.com"]
			},
			"currentState": {
				"name": "my-container",
				"location": { "region": "GRA" },
				"archive": false,
				"type": "PUBLIC",
				"cors": ["https://a.example.com", "https://b.example.com"],
				"objectsCount": 3,
				"objectsSize": 1048576,
				"staticUrl": "https://storage.gra.cloud.ovh.net/v1/AUTH_x/my-container"
			}
		}`))

	out, err := cmd.Execute("cloud", "storage", "swift", "container", "get", "GRA_my-container", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.Re(`Name\**:\s+my-container\b`))
	assert.Cmp(out, td.Re(`Type\**:\s+PUBLIC\b`))
	assert.Cmp(out, td.Contains("https://a.example.com, https://b.example.com"))
	assert.Cmp(out, td.Re(`Objects count\**:\s+3\b`))
	assert.Cmp(out, td.Contains("1.00 MiB"))
	assert.Cmp(out, td.Contains("https://storage.gra.cloud.ovh.net/v1/AUTH_x/my-container"))
	assert.Cmp(out, td.Not(td.Contains("<no value>")))
}

func (ms *MockSuite) TestCloudStorageSwiftContainerGetCreatingCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, swiftContainerURL+"/GRA_my-container",
		httpmock.NewStringResponder(200, `{
			"id": "GRA_my-container",
			"resourceStatus": "CREATING",
			"targetSpec": {
				"name": "my-container",
				"location": { "region": "GRA" },
				"archive": false,
				"type": "PRIVATE"
			}
		}`))

	out, err := cmd.Execute("cloud", "storage", "swift", "container", "get", "GRA_my-container", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("CREATING"))
	assert.Cmp(out, td.Not(td.Contains("Current state")))
	assert.Cmp(out, td.Not(td.Contains("<no value>")))
}

func (ms *MockSuite) TestCloudStorageSwiftContainerCreateNoRegionCmd(assert, require *td.T) {
	_, err := cmd.Execute("cloud", "storage", "swift", "container", "create",
		"--cloud-project", "fakeProjectID",
		"--name", "my-container")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("region argument is required"))
	assert.Cmp(httpmock.GetTotalCallCount(), 0)
}

func (ms *MockSuite) TestCloudStorageSwiftContainerCreateCmd(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(http.MethodPost, swiftContainerURL,
		tdhttpmock.JSONBody(td.JSON(`
			{
				"targetSpec": {
					"name": "my-container",
					"location": { "region": "GRA" },
					"archive": true,
					"type": "STATIC",
					"cors": ["https://a.example.com", "https://b.example.com"]
				}
			}`),
		),
		httpmock.NewStringResponder(202, `{
			"id": "GRA_my-container",
			"resourceStatus": "CREATING"
		}`))

	out, err := cmd.Execute("cloud", "storage", "swift", "container", "create", "GRA",
		"--cloud-project", "fakeProjectID",
		"--name", "my-container",
		"--archive",
		"--type", "STATIC",
		"--cors", "https://a.example.com,https://b.example.com",
		"-o", "json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{
		"message": "⚡️ Container GRA_my-container creation started",
		"details": {"id": "GRA_my-container", "resourceStatus": "CREATING"}
	}`))
}

func (ms *MockSuite) TestCloudStorageSwiftContainerCreateWaitCmd(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(http.MethodPost, swiftContainerURL,
		tdhttpmock.JSONBody(td.JSON(`{"targetSpec": {"name": "my-container", "location": { "region": "SBG" }}}`)),
		httpmock.NewStringResponder(202, `{"id": "SBG_my-container", "resourceStatus": "CREATING"}`))

	httpmock.RegisterResponder(http.MethodGet, swiftContainerURL+"/SBG_my-container",
		httpmock.NewStringResponder(200, `{"id": "SBG_my-container", "resourceStatus": "READY"}`))

	out, err := cmd.Execute("cloud", "storage", "swift", "container", "create", "SBG",
		"--cloud-project", "fakeProjectID",
		"--name", "my-container",
		"--wait",
		"-o", "json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{
		"message": "✅ Container SBG_my-container created successfully",
		"details": {"id": "SBG_my-container", "resourceStatus": "READY"}
	}`))
}

func (ms *MockSuite) TestCloudStorageSwiftContainerEditCmd(assert, require *td.T) {
	// The PUT replaces type and cors, so both must be carried over from the
	// fetched container; --cors adds to the existing origins.
	httpmock.RegisterResponder(http.MethodGet, swiftContainerURL+"/GRA_my-container",
		httpmock.NewStringResponder(200, `{
			"id": "GRA_my-container",
			"checksum": "abc123",
			"resourceStatus": "READY",
			"targetSpec": {
				"name": "my-container",
				"location": { "region": "GRA" },
				"archive": false,
				"type": "PRIVATE",
				"cors": ["https://a.example.com"]
			},
			"currentState": {
				"name": "my-container",
				"location": { "region": "GRA" },
				"archive": false,
				"type": "PRIVATE",
				"cors": ["https://a.example.com"]
			}
		}`))

	httpmock.RegisterMatcherResponder(http.MethodPut, swiftContainerURL+"/GRA_my-container",
		tdhttpmock.JSONBody(td.JSON(`
			{
				"checksum": "abc123",
				"targetSpec": {
					"type": "PUBLIC",
					"cors": ["https://a.example.com", "https://b.example.com"]
				}
			}`),
		),
		httpmock.NewStringResponder(202, ``))

	out, err := cmd.Execute("cloud", "storage", "swift", "container", "edit", "GRA_my-container",
		"--cloud-project", "fakeProjectID",
		"--type", "PUBLIC",
		"--cors", "https://b.example.com")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("updated successfully"))
}

func (ms *MockSuite) TestCloudStorageSwiftContainerDeleteCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodDelete, swiftContainerURL+"/GRA_my-container",
		httpmock.NewStringResponder(202, `{"id": "GRA_my-container", "resourceStatus": "DELETING"}`))

	out, err := cmd.Execute("cloud", "storage", "swift", "container", "delete", "GRA_my-container", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("Container GRA_my-container is being deleted"))
}

func (ms *MockSuite) TestCloudStorageSwiftContainerPresignCmd(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(http.MethodPost, swiftContainerURL+"/GRA_my-container/presign",
		tdhttpmock.JSONBody(td.JSON(`{"key": "dir/file.txt", "expiresIn": 600}`)),
		httpmock.NewStringResponder(201, `{
			"method": "GET",
			"key": "dir/file.txt",
			"url": "https://storage.gra.cloud.ovh.net/v1/AUTH_x/my-container/dir/file.txt?temp_url_sig=s",
			"expiresAt": "2026-11-02T10:10:00Z"
		}`))

	out, err := cmd.Execute("cloud", "storage", "swift", "container", "presign", "GRA_my-container", "dir/file.txt",
		"--cloud-project", "fakeProjectID",
		"--expires-in", "600")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("-> GET https://storage.gra.cloud.ovh.net/v1/AUTH_x/my-container/dir/file.txt?temp_url_sig=s"))
	assert.Cmp(out, td.Contains("-> Expires at: 2026-11-02T10:10:00Z"))
}

func (ms *MockSuite) TestCloudStorageSwiftContainerPresignDefaultExpirationCmd(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(http.MethodPost, swiftContainerURL+"/GRA_my-container/presign",
		tdhttpmock.JSONBody(td.JSON(`{"key": "file.txt"}`)),
		httpmock.NewStringResponder(201, `{"method": "GET", "key": "file.txt", "url": "https://u", "expiresAt": "2026-11-02T11:00:00Z"}`))

	_, err := cmd.Execute("cloud", "storage", "swift", "container", "presign", "GRA_my-container", "file.txt",
		"--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
}

func (ms *MockSuite) TestCloudStorageSwiftObjectListCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, swiftContainerURL+"/GRA_my-container/object?prefix=dir%2F",
		httpmock.NewStringResponder(200, `[
			{"key": "dir/file.txt", "size": 12, "contentType": "text/plain", "etag": "e", "lastModifiedAt": "2026-10-20T10:00:00Z"}
		]`))

	out, err := cmd.Execute("cloud", "storage", "swift", "container", "object", "list", "GRA_my-container",
		"--cloud-project", "fakeProjectID",
		"--prefix", "dir/")

	require.CmpNoError(err)
	assert.Cmp(out, td.Re(`dir/file.txt\s+\S+\s+12\s+\S+\s+text/plain\s+\S+\s+2026-10-20T10:00:00Z`))
}

func (ms *MockSuite) TestCloudStorageSwiftObjectGetCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, swiftContainerURL+"/GRA_my-container/object/dir%2Ffile.txt",
		httpmock.NewStringResponder(200, `{
			"key": "dir/file.txt",
			"size": 12,
			"contentType": "text/plain",
			"etag": "abcdef"
		}`))

	out, err := cmd.Execute("cloud", "storage", "swift", "container", "object", "get", "GRA_my-container", "dir/file.txt",
		"--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("abcdef"))
	assert.Cmp(out, td.Re(`Size\**:\s+12 bytes`))
	assert.Cmp(out, td.Not(td.Contains("Retrieval")))
	assert.Cmp(out, td.Not(td.Contains("<no value>")))
}

func (ms *MockSuite) TestCloudStorageSwiftObjectDeleteCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodDelete, swiftContainerURL+"/GRA_my-container/object/dir%2Ffile.txt",
		httpmock.NewStringResponder(204, ``))

	out, err := cmd.Execute("cloud", "storage", "swift", "container", "object", "delete", "GRA_my-container", "dir/file.txt",
		"--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("Object dir/file.txt deleted from container GRA_my-container"))
}
