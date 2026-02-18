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

// registerS3ContainerMocks registers the standard HTTP mocks needed to locate a container by name.
func registerS3ContainerMocks(containerName string) {
	httpmock.RegisterResponder(http.MethodGet, "https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region",
		httpmock.NewStringResponder(200, `["BHS"]`))

	httpmock.RegisterResponder(http.MethodGet, "https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/BHS",
		httpmock.NewStringResponder(200, `{
			"name": "BHS",
			"type": "region",
			"status": "UP",
			"services": [
				{"name": "storage", "status": "UP"},
				{"name": "storage-s3-high-perf", "status": "UP"},
				{"name": "storage-s3-standard", "status": "UP"}
			],
			"countryCode": "ca",
			"ipCountries": [],
			"continentCode": "NA",
			"availabilityZones": [],
			"datacenterLocation": "BHS"
		}`))

	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/BHS/storage/"+containerName,
		httpmock.NewStringResponder(200, `{
			"name": "`+containerName+`",
			"virtualHost": "https://`+containerName+`.test.ovh.net/",
			"ownerId": 0,
			"objectsCount": 0,
			"objectsSize": 0,
			"region": "BHS",
			"createdAt": "2025-02-10T14:24:12Z"
		}`))
}

func (ms *MockSuite) TestCloudStorageS3BulkDeletePrefixCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, "https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region",
		httpmock.NewStringResponder(200, `["BHS"]`))

	httpmock.RegisterResponder(http.MethodGet, "https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/BHS",
		httpmock.NewStringResponder(200, `{
			"name": "BHS",
			"type": "region",
			"status": "UP",
			"services": [
				{
					"name": "storage",
					"status": "UP"
				},
				{
					"name": "storage-s3-high-perf",
					"status": "UP"
				},
				{
					"name": "storage-s3-standard",
					"status": "UP"
				}
			],
			"countryCode": "ca",
			"ipCountries": [],
			"continentCode": "NA",
			"availabilityZones": [],
			"datacenterLocation": "BHS"
		}`))

	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/BHS/storage/fakeContainer",
		httpmock.NewStringResponder(200, `{
			"name": "fakeContainer",
			"virtualHost": "https://fakeContainer.test.ovh.net/",
			"ownerId": 0,
			"objectsCount": 15,
			"objectsSize": 4147089,
			"objects": [
				{"key": "logs/log1.txt"},
				{"key": "logs/log2.txt"},
				{"key": "images/img1.png"}
			],
			"region": "BHS",
			"createdAt": "2025-02-10T14:24:12Z"
		}`))

	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/BHS/storage/fakeContainer/object?prefix=logs%2F",
		httpmock.NewStringResponder(200, `[
			{"key": "logs/log1.txt"},
			{"key": "logs/log2.txt"}
		]`).Then(httpmock.NewStringResponder(200, `[]`)),
	)

	httpmock.RegisterMatcherResponder(http.MethodPost,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/BHS/storage/fakeContainer/bulkDeleteObjects",
		tdhttpmock.JSONBody(td.JSON(`
			{
				"objects": [
					{"key": "logs/log1.txt"},
					{"key": "logs/log2.txt"}
				]
			}`),
		),
		httpmock.NewStringResponder(200, ``),
	)

	out, err := cmd.Execute("cloud", "storage-s3", "bulk-delete", "fakeContainer", "--cloud-project", "fakeProjectID", "--prefix", "logs/", "--json")
	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{"message": "✅ Objects deleted successfully"}`))
}

func (ms *MockSuite) TestCloudStorageS3LifecycleGetCmd(assert, require *td.T) {
	registerS3ContainerMocks("fakeContainer")

	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/BHS/storage/fakeContainer/lifecycle",
		httpmock.NewStringResponder(200, `{
			"rules": [
				{
					"id": "expire-logs",
					"status": "enabled",
					"filter": {"prefix": "logs/"},
					"expiration": {"days": 30}
				}
			]
		}`))

	out, err := cmd.Execute("cloud", "storage-s3", "lifecycle", "get", "fakeContainer", "--cloud-project", "fakeProjectID", "--json")
	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{
		"rules": [
			{
				"id": "expire-logs",
				"status": "enabled",
				"filter": {"prefix": "logs/"},
				"expiration": {"days": 30}
			}
		]
	}`))
}

func (ms *MockSuite) TestCloudStorageS3LifecycleDeleteCmd(assert, require *td.T) {
	registerS3ContainerMocks("fakeContainer")

	httpmock.RegisterResponder(http.MethodDelete,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/BHS/storage/fakeContainer/lifecycle",
		httpmock.NewStringResponder(200, ``))

	out, err := cmd.Execute("cloud", "storage-s3", "lifecycle", "delete", "fakeContainer", "--cloud-project", "fakeProjectID", "--json")
	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{"message": "✅ Lifecycle configuration for container fakeContainer deleted successfully"}`))
}

func (ms *MockSuite) TestCloudStorageS3ObjectCopyCmd(assert, require *td.T) {
	registerS3ContainerMocks("fakeContainer")

	httpmock.RegisterMatcherResponder(http.MethodPost,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/BHS/storage/fakeContainer/object/myobject.txt/copy",
		tdhttpmock.JSONBody(td.JSON(`{"targetBucket": "destBucket", "targetKey": "dest/myobject.txt"}`)),
		httpmock.NewStringResponder(200, `{"etag": "abc123", "versionId": null}`))

	out, err := cmd.Execute("cloud", "storage-s3", "object", "copy", "fakeContainer", "myobject.txt",
		"--cloud-project", "fakeProjectID",
		"--target-bucket", "destBucket",
		"--target-key", "dest/myobject.txt",
		"--json")
	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{
		"message": "✅ Object myobject.txt copied successfully",
		"details": {"etag": "abc123", "versionId": null}
	}`))
}

func (ms *MockSuite) TestCloudStorageS3ObjectRestoreCmd(assert, require *td.T) {
	registerS3ContainerMocks("fakeContainer")

	httpmock.RegisterMatcherResponder(http.MethodPost,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/BHS/storage/fakeContainer/object/myobject.txt/restore",
		tdhttpmock.JSONBody(td.JSON(`{"days": 7}`)),
		httpmock.NewStringResponder(200, ``))

	out, err := cmd.Execute("cloud", "storage-s3", "object", "restore", "fakeContainer", "myobject.txt",
		"--cloud-project", "fakeProjectID",
		"--days", "7",
		"--json")
	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{"message": "✅ Object myobject.txt restore initiated successfully"}`))
}

func (ms *MockSuite) TestCloudStorageS3ObjectVersionCopyCmd(assert, require *td.T) {
	registerS3ContainerMocks("fakeContainer")

	httpmock.RegisterMatcherResponder(http.MethodPost,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/BHS/storage/fakeContainer/object/myobject.txt/version/v1/copy",
		tdhttpmock.JSONBody(td.JSON(`{"targetBucket": "destBucket", "targetKey": "dest/myobject.txt"}`)),
		httpmock.NewStringResponder(200, `{"etag": "abc123", "versionId": "v2"}`))

	out, err := cmd.Execute("cloud", "storage-s3", "object", "version", "copy", "fakeContainer", "myobject.txt", "v1",
		"--cloud-project", "fakeProjectID",
		"--target-bucket", "destBucket",
		"--target-key", "dest/myobject.txt",
		"--json")
	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{
		"message": "✅ Object myobject.txt version v1 copied successfully",
		"details": {"etag": "abc123", "versionId": "v2"}
	}`))
}

func (ms *MockSuite) TestCloudStorageS3ObjectVersionRestoreCmd(assert, require *td.T) {
	registerS3ContainerMocks("fakeContainer")

	httpmock.RegisterMatcherResponder(http.MethodPost,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/BHS/storage/fakeContainer/object/myobject.txt/version/v1/restore",
		tdhttpmock.JSONBody(td.JSON(`{"days": 14}`)),
		httpmock.NewStringResponder(200, ``))

	out, err := cmd.Execute("cloud", "storage-s3", "object", "version", "restore", "fakeContainer", "myobject.txt", "v1",
		"--cloud-project", "fakeProjectID",
		"--days", "14",
		"--json")
	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{"message": "✅ Object myobject.txt version v1 restore initiated successfully"}`))
}

func (ms *MockSuite) TestCloudStorageS3ReplicationJobCmd(assert, require *td.T) {
	registerS3ContainerMocks("fakeContainer")

	httpmock.RegisterResponder(http.MethodPost,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/BHS/storage/fakeContainer/job/replication",
		httpmock.NewStringResponder(200, `{"id": "job-123"}`))

	out, err := cmd.Execute("cloud", "storage-s3", "replication-job", "create", "fakeContainer", "--cloud-project", "fakeProjectID", "--json")
	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{
		"message": "✅ Replication job created successfully (ID: job-123)",
		"details": {"id": "job-123"}
	}`))
}

func (ms *MockSuite) TestCloudStorageS3QuotaGetCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/BHS/quota/storage",
		httpmock.NewStringResponder(200, `{
			"bytesUsed": 1048576,
			"quotaBytes": 10737418240,
			"containerCount": 3,
			"objectCount": 42
		}`))

	out, err := cmd.Execute("cloud", "storage-s3", "quota", "get", "BHS", "--cloud-project", "fakeProjectID", "--json")
	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{
		"bytesUsed": 1048576,
		"quotaBytes": 10737418240,
		"containerCount": 3,
		"objectCount": 42
	}`))
}

func (ms *MockSuite) TestCloudStorageS3QuotaEditCmd(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(http.MethodPut,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/BHS/quota/storage",
		tdhttpmock.JSONBody(td.JSON(`{"quotaBytes": 21474836480}`)),
		httpmock.NewStringResponder(200, ``))

	out, err := cmd.Execute("cloud", "storage-s3", "quota", "edit", "BHS", "--cloud-project", "fakeProjectID", "--quota-bytes", "21474836480", "--json")
	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{"message": "✅ Storage quota for region BHS updated successfully"}`))
}

func (ms *MockSuite) TestCloudStorageS3QuotaDeleteCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodDelete,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/BHS/quota/storage",
		httpmock.NewStringResponder(200, ``))

	out, err := cmd.Execute("cloud", "storage-s3", "quota", "delete", "BHS", "--cloud-project", "fakeProjectID", "--json")
	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{"message": "✅ Storage quota for region BHS deleted successfully"}`))
}

func (ms *MockSuite) TestCloudStorageS3BulkDeleteAllCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, "https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region",
		httpmock.NewStringResponder(200, `["BHS"]`))

	httpmock.RegisterResponder(http.MethodGet, "https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/BHS",
		httpmock.NewStringResponder(200, `{
			"name": "BHS",
			"type": "region",
			"status": "UP",
			"services": [
				{
					"name": "storage",
					"status": "UP"
				},
				{
					"name": "storage-s3-high-perf",
					"status": "UP"
				},
				{
					"name": "storage-s3-standard",
					"status": "UP"
				}
			],
			"countryCode": "ca",
			"ipCountries": [],
			"continentCode": "NA",
			"availabilityZones": [],
			"datacenterLocation": "BHS"
		}`))

	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/BHS/storage/fakeContainer",
		httpmock.NewStringResponder(200, `{
			"name": "fakeContainer",
			"virtualHost": "https://fakeContainer.test.ovh.net/",
			"ownerId": 0,
			"objectsCount": 15,
			"objectsSize": 4147089,
			"objects": [
				{"key": "logs/log1.txt"},
				{"key": "logs/log2.txt"},
				{"key": "images/img1.png"}
			],
			"region": "BHS",
			"createdAt": "2025-02-10T14:24:12Z"
		}`))

	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/BHS/storage/fakeContainer/object",
		httpmock.NewStringResponder(200, `[
			{"key": "logs/log1.txt"},
			{"key": "logs/log2.txt"},
			{"key": "images/img1.png"}
		]`).Then(httpmock.NewStringResponder(200, `[]`)),
	)

	httpmock.RegisterMatcherResponder(http.MethodPost,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/BHS/storage/fakeContainer/bulkDeleteObjects",
		tdhttpmock.JSONBody(td.JSON(`
			{
				"objects": [
					{"key": "logs/log1.txt"},
					{"key": "logs/log2.txt"},
					{"key": "images/img1.png"}
				]
			}`),
		),
		httpmock.NewStringResponder(200, ``),
	)

	out, err := cmd.Execute("cloud", "storage-s3", "bulk-delete", "fakeContainer", "--cloud-project", "fakeProjectID", "--all", "--json")
	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{"message": "✅ Objects deleted successfully"}`))
}
