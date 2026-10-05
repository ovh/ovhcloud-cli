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

func (ms *MockSuite) TestCloudStorageS3BucketListCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/storage/object/bucket",
		httpmock.NewStringResponder(200, `[
			{
				"id": "GRA_bucket-1",
				"resourceStatus": "READY",
				"createdAt": "2026-02-01T10:00:00Z",
				"targetSpec": {
					"name": "bucket-1",
					"location": { "region": "GRA" }
				}
			},
			{
				"id": "SBG_bucket-2",
				"resourceStatus": "CREATING",
				"createdAt": "2026-02-02T10:00:00Z",
				"targetSpec": {
					"name": "bucket-2",
					"location": { "region": "SBG" }
				},
				"currentState": null
			}
		]`))

	out, err := cmd.Execute("cloud", "storage", "object", "bucket", "list", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("bucket-1"))
	assert.Cmp(out, td.Contains("GRA"))
	assert.Cmp(out, td.Contains("bucket-2"))
	assert.Cmp(out, td.Contains("CREATING"))
}

func (ms *MockSuite) TestCloudStorageS3BucketGetCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/storage/object/bucket/GRA_my-data-bucket",
		httpmock.NewStringResponder(200, `{
			"id": "GRA_my-data-bucket",
			"checksum": "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4",
			"resourceStatus": "READY",
			"createdAt": "2026-02-01T10:00:00Z",
			"updatedAt": "2026-02-01T10:05:00Z",
			"currentTasks": [],
			"targetSpec": {
				"name": "my-data-bucket",
				"location": { "region": "GRA" },
				"encryption": { "algorithm": "AES256" },
				"versioning": { "status": "ENABLED" },
				"tags": { "env": "production" }
			},
			"currentState": {
				"name": "my-data-bucket",
				"location": { "region": "GRA" },
				"encryption": { "algorithm": "AES256" },
				"versioning": { "status": "ENABLED" },
				"tags": { "env": "production" },
				"objectsCount": 1,
				"objectsSize": 1048576,
				"virtualHost": "my-data-bucket.s3.gra.io.cloud.ovh.net"
			}
		}`))

	out, err := cmd.Execute("cloud", "storage", "object", "bucket", "get", "GRA_my-data-bucket", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.Re(`Name\**:\s+my-data-bucket\b`))
	assert.Cmp(out, td.Contains("GRA"))
	assert.Cmp(out, td.Contains("ENABLED"))
	assert.Cmp(out, td.Contains("AES256"))
	assert.Cmp(out, td.Contains("production"))
	assert.Cmp(out, td.Re(`Objects count\**:\s+1\b`))
	assert.Cmp(out, td.Contains("1.00 MiB"))
	assert.Cmp(out, td.Contains("https://my-data-bucket.s3.gra.io.cloud.ovh.net"))
	// Optional fields absent from the response (objectLock, ownerUserId)
	// must not leak Go's "<no value>" placeholder.
	assert.Cmp(out, td.Not(td.Contains("<no value>")))
}

func (ms *MockSuite) TestCloudStorageS3BucketGetNotFoundCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/storage/object/bucket/GRA_missing-bucket",
		httpmock.NewStringResponder(404, `{"class":"Client::NotFound::BucketDoesNotExist","message":"S3 bucket not found"}`))

	_, err := cmd.Execute("cloud", "storage", "object", "bucket", "get", "GRA_missing-bucket", "--cloud-project", "fakeProjectID")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("failed to fetch bucket"))
	assert.Cmp(err.Error(), td.Contains("Client::NotFound::BucketDoesNotExist"))
}

func (ms *MockSuite) TestCloudStorageS3BucketCreateMissingRegionCmd(assert, require *td.T) {
	_, err := cmd.Execute("cloud", "storage", "object", "bucket", "create",
		"--cloud-project", "fakeProjectID",
		"--name", "my-bucket")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("region argument is required"))
	assert.Cmp(httpmock.GetTotalCallCount(), 0)
}

func (ms *MockSuite) TestCloudStorageS3BucketCreateCmd(assert, require *td.T) {
	// The command must build a v2 "targetSpec" body from the CLI flags and
	// the positional region.
	httpmock.RegisterMatcherResponder(http.MethodPost,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/storage/object/bucket",
		tdhttpmock.JSONBody(td.JSON(`
			{
				"targetSpec": {
					"name": "my-bucket",
					"location": { "region": "GRA" },
					"encryption": { "algorithm": "AES256" },
					"objectLock": { "mode": "COMPLIANCE", "retentionDays": 30 },
					"ownerUserId": "1234",
					"versioning": { "status": "ENABLED" }
				}
			}`),
		),
		httpmock.NewStringResponder(200, `{
			"id": "GRA_my-bucket",
			"resourceStatus": "CREATING"
		}`))

	out, err := cmd.Execute("cloud", "storage", "object", "bucket", "create", "GRA",
		"--cloud-project", "fakeProjectID",
		"--name", "my-bucket",
		"--encryption-algorithm", "AES256",
		"--object-lock-mode", "COMPLIANCE",
		"--object-lock-retention-days", "30",
		"--owner-user-id", "1234",
		"--versioning-status", "ENABLED")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("my-bucket"))
}

func (ms *MockSuite) TestCloudStorageS3BucketCreateWaitCmd(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(http.MethodPost,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/storage/object/bucket",
		tdhttpmock.JSONBody(td.JSON(`
			{
				"targetSpec": {
					"name": "my-bucket",
					"location": { "region": "SBG" }
				}
			}`),
		),
		httpmock.NewStringResponder(200, `{
			"id": "SBG_my-bucket",
			"resourceStatus": "CREATING"
		}`))

	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/storage/object/bucket/SBG_my-bucket",
		httpmock.NewStringResponder(200, `{
			"id": "SBG_my-bucket",
			"resourceStatus": "READY"
		}`))

	out, err := cmd.Execute("cloud", "storage", "object", "bucket", "create", "SBG",
		"--cloud-project", "fakeProjectID",
		"--name", "my-bucket",
		"--wait",
		"-o", "json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{
		"message": "✅ Bucket SBG_my-bucket created successfully",
		"details": {"id": "SBG_my-bucket", "resourceStatus": "READY"}
	}`))
}

func (ms *MockSuite) TestCloudStorageS3BucketEditCmd(assert, require *td.T) {
	// EditResource must GET the bucket, preserve its checksum (optimistic
	// locking) and PUT only the editable targetSpec fields (no name nor
	// location), with the CLI flags overriding the current values.
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/storage/object/bucket/GRA_my-bucket",
		httpmock.NewStringResponder(200, `{
			"id": "GRA_my-bucket",
			"checksum": "abc123",
			"resourceStatus": "READY",
			"targetSpec": {
				"name": "my-bucket",
				"location": { "region": "GRA" },
				"versioning": { "status": "DISABLED" }
			},
			"currentState": {
				"name": "my-bucket",
				"location": { "region": "GRA" },
				"versioning": { "status": "DISABLED" }
			}
		}`))

	httpmock.RegisterMatcherResponder(http.MethodPut,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/storage/object/bucket/GRA_my-bucket",
		tdhttpmock.JSONBody(td.JSON(`
			{
				"checksum": "abc123",
				"targetSpec": {
					"tags": { "env": "prod" },
					"versioning": { "status": "ENABLED" }
				}
			}`),
		),
		httpmock.NewStringResponder(200, ``))

	out, err := cmd.Execute("cloud", "storage", "object", "bucket", "edit", "GRA_my-bucket",
		"--cloud-project", "fakeProjectID",
		"--tag", "env=prod",
		"--versioning-status", "ENABLED")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("updated successfully"))
}

func (ms *MockSuite) TestCloudStorageS3BucketDeleteCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodDelete,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/storage/object/bucket/GRA_my-bucket",
		httpmock.NewStringResponder(200, `{
			"id": "GRA_my-bucket",
			"resourceStatus": "DELETING"
		}`))

	out, err := cmd.Execute("cloud", "storage", "object", "bucket", "delete", "GRA_my-bucket", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("Bucket GRA_my-bucket is being deleted"))
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

	out, err := cmd.Execute("cloud", "storage", "object", "bucket", "bulk-delete", "fakeContainer", "--cloud-project", "fakeProjectID", "--prefix", "logs/", "-o", "json")
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

	out, err := cmd.Execute("cloud", "storage", "object", "bucket", "lifecycle", "get", "fakeContainer", "--cloud-project", "fakeProjectID", "-o", "json")
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

	out, err := cmd.Execute("cloud", "storage", "object", "bucket", "lifecycle", "delete", "fakeContainer", "--cloud-project", "fakeProjectID", "-o", "json")
	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{"message": "✅ Lifecycle configuration for container fakeContainer deleted successfully"}`))
}

func (ms *MockSuite) TestCloudStorageS3ObjectCopyCmd(assert, require *td.T) {
	registerS3ContainerMocks("fakeContainer")

	httpmock.RegisterMatcherResponder(http.MethodPost,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/BHS/storage/fakeContainer/object/myobject.txt/copy",
		tdhttpmock.JSONBody(td.JSON(`{"targetBucket": "destBucket", "targetKey": "dest/myobject.txt"}`)),
		httpmock.NewStringResponder(200, `{"etag": "abc123", "versionId": null}`))

	out, err := cmd.Execute("cloud", "storage", "object", "bucket", "object", "copy", "fakeContainer", "myobject.txt",
		"--cloud-project", "fakeProjectID",
		"--target-bucket", "destBucket",
		"--target-key", "dest/myobject.txt",
		"-o", "json")
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

	out, err := cmd.Execute("cloud", "storage", "object", "bucket", "object", "restore", "fakeContainer", "myobject.txt",
		"--cloud-project", "fakeProjectID",
		"--days", "7",
		"-o", "json")
	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{"message": "✅ Object myobject.txt restore initiated successfully"}`))
}

func (ms *MockSuite) TestCloudStorageS3ObjectVersionCopyCmd(assert, require *td.T) {
	registerS3ContainerMocks("fakeContainer")

	httpmock.RegisterMatcherResponder(http.MethodPost,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/BHS/storage/fakeContainer/object/myobject.txt/version/v1/copy",
		tdhttpmock.JSONBody(td.JSON(`{"targetBucket": "destBucket", "targetKey": "dest/myobject.txt"}`)),
		httpmock.NewStringResponder(200, `{"etag": "abc123", "versionId": "v2"}`))

	out, err := cmd.Execute("cloud", "storage", "object", "bucket", "object", "version", "copy", "fakeContainer", "myobject.txt", "v1",
		"--cloud-project", "fakeProjectID",
		"--target-bucket", "destBucket",
		"--target-key", "dest/myobject.txt",
		"-o", "json")
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

	out, err := cmd.Execute("cloud", "storage", "object", "bucket", "object", "version", "restore", "fakeContainer", "myobject.txt", "v1",
		"--cloud-project", "fakeProjectID",
		"--days", "14",
		"-o", "json")
	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{"message": "✅ Object myobject.txt version v1 restore initiated successfully"}`))
}

func (ms *MockSuite) TestCloudStorageS3ReplicationJobCmd(assert, require *td.T) {
	registerS3ContainerMocks("fakeContainer")

	httpmock.RegisterResponder(http.MethodPost,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/BHS/storage/fakeContainer/job/replication",
		httpmock.NewStringResponder(200, `{"id": "job-123"}`))

	out, err := cmd.Execute("cloud", "storage", "object", "bucket", "replication-job", "create", "fakeContainer", "--cloud-project", "fakeProjectID", "-o", "json")
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

	out, err := cmd.Execute("cloud", "storage", "object", "quota", "get", "BHS", "--cloud-project", "fakeProjectID", "-o", "json")
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

	out, err := cmd.Execute("cloud", "storage", "object", "quota", "edit", "BHS", "--cloud-project", "fakeProjectID", "--quota-bytes", "21474836480", "-o", "json")
	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{"message": "✅ Storage quota for region BHS updated successfully"}`))
}

func (ms *MockSuite) TestCloudStorageS3QuotaDeleteCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodDelete,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/BHS/quota/storage",
		httpmock.NewStringResponder(200, ``))

	out, err := cmd.Execute("cloud", "storage", "object", "quota", "delete", "BHS", "--cloud-project", "fakeProjectID", "-o", "json")
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

	out, err := cmd.Execute("cloud", "storage", "object", "bucket", "bulk-delete", "fakeContainer", "--cloud-project", "fakeProjectID", "--all", "-o", "json")
	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{"message": "✅ Objects deleted successfully"}`))
}
