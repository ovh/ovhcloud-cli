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
