// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/jarcoal/httpmock"
	"github.com/maxatome/go-testdeep/td"
	"github.com/maxatome/tdhttpmock"
	"github.com/ovh/ovhcloud-cli/internal/cmd"
)

const s3BucketURL = "https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/storage/object/bucket/GRA_my-bucket"

func (ms *MockSuite) TestCloudStorageS3ObjectListCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, s3BucketURL+"/object?delimiter=%2F&prefix=logs%2F",
		httpmock.NewStringResponder(200, `[
			{"key": "logs/app.log", "size": 2048, "storageClass": "STANDARD", "lastModifiedAt": "2026-10-20T10:00:00Z"},
			{"key": "logs/2026/", "isCommonPrefix": true}
		]`))

	out, err := cmd.Execute("cloud", "storage", "object", "bucket", "object", "list", "GRA_my-bucket",
		"--cloud-project", "fakeProjectID",
		"--prefix", "logs/",
		"--delimiter", "/")

	require.CmpNoError(err)
	assert.Cmp(out, td.Re(`logs/app.log\s+\S+\s+2048\s+\S+\s+STANDARD\s+\S+\s+2026-10-20T10:00:00Z`))
	assert.Cmp(out, td.Contains("logs/2026/"))
	assert.Cmp(out, td.Not(td.Contains("versionId")))
}

func (ms *MockSuite) TestCloudStorageS3ObjectListWithVersionsCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, s3BucketURL+"/object?withVersions=true",
		httpmock.NewStringResponder(200, `[
			{"key": "a.txt", "size": 1, "versionId": "v2", "isLatest": true},
			{"key": "a.txt", "size": 1, "versionId": "v1", "isLatest": false}
		]`))

	out, err := cmd.Execute("cloud", "storage", "object", "bucket", "object", "list", "GRA_my-bucket",
		"--cloud-project", "fakeProjectID",
		"--with-versions")

	require.CmpNoError(err)
	assert.Cmp(out, td.Re(`a\.txt.*v2\s+\S+\s+true`))
	assert.Cmp(out, td.Re(`a\.txt.*v1\s+\S+\s+false`))
}

func (ms *MockSuite) TestCloudStorageS3ObjectGetCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, s3BucketURL+"/object/dir%2Ffile.txt",
		httpmock.NewStringResponder(200, `{
			"key": "dir/file.txt",
			"size": 12,
			"etag": "abcdef",
			"storageClass": "STANDARD",
			"lastModifiedAt": "2026-10-20T10:00:00Z",
			"legalHold": "ON",
			"lock": {"mode": "GOVERNANCE", "retainUntil": "2027-01-01T00:00:00Z"}
		}`))

	out, err := cmd.Execute("cloud", "storage", "object", "bucket", "object", "get", "GRA_my-bucket", "dir/file.txt",
		"--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.Re(`Size\**:\s+12 bytes`))
	assert.Cmp(out, td.Re(`Legal hold\**:\s+ON`))
	assert.Cmp(out, td.Re(`Mode\**:\s+GOVERNANCE`))
	assert.Cmp(out, td.Contains("2027-01-01T00:00:00Z"))
	assert.Cmp(out, td.Not(td.Contains("<no value>")))
}

func (ms *MockSuite) TestCloudStorageS3ObjectGetMinimalCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, s3BucketURL+"/object/file.txt",
		httpmock.NewStringResponder(200, `{"key": "file.txt", "size": 0}`))

	out, err := cmd.Execute("cloud", "storage", "object", "bucket", "object", "get", "GRA_my-bucket", "file.txt",
		"--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("file.txt"))
	assert.Cmp(out, td.Not(td.Contains("Object lock")))
	assert.Cmp(out, td.Not(td.Contains("<no value>")))
}

func (ms *MockSuite) TestCloudStorageS3ObjectEditCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, s3BucketURL+"/object/dir%2Ffile.txt",
		httpmock.NewStringResponder(200, `{"key": "dir/file.txt", "size": 12, "legalHold": "OFF"}`))

	httpmock.RegisterMatcherResponder(http.MethodPut, s3BucketURL+"/object/dir%2Ffile.txt",
		tdhttpmock.JSONBody(td.JSON(`{
			"legalHold": "ON",
			"lock": {"mode": "COMPLIANCE", "retainUntil": "2027-01-01T00:00:00Z"}
		}`)),
		httpmock.NewStringResponder(200, `{"key": "dir/file.txt"}`))

	out, err := cmd.Execute("cloud", "storage", "object", "bucket", "object", "edit", "GRA_my-bucket", "dir/file.txt",
		"--cloud-project", "fakeProjectID",
		"--legal-hold", "ON",
		"--lock-mode", "COMPLIANCE",
		"--lock-retain-until", "2027-01-01T00:00:00Z")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("updated successfully"))
}

func (ms *MockSuite) TestCloudStorageS3ObjectDeleteCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodDelete, s3BucketURL+"/object/dir%2Ffile.txt",
		httpmock.NewStringResponder(200, `{"key": "dir/file.txt"}`))

	out, err := cmd.Execute("cloud", "storage", "object", "bucket", "object", "delete", "GRA_my-bucket", "dir/file.txt",
		"--cloud-project", "fakeProjectID",
		"-o", "json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{"message": "✅ Object dir/file.txt deleted from bucket GRA_my-bucket"}`))
}

func (ms *MockSuite) TestCloudStorageS3ObjectCopyCmd(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(http.MethodPost, s3BucketURL+"/object/dir%2Ffile.txt/copy",
		tdhttpmock.JSONBody(td.JSON(`{"destinationBucket": "SBG_dest-bucket", "destinationKey": "dest/file.txt", "storageClass": "STANDARD_IA"}`)),
		httpmock.NewStringResponder(200, `{"key": "dest/file.txt", "etag": "abc123"}`))

	out, err := cmd.Execute("cloud", "storage", "object", "bucket", "object", "copy", "GRA_my-bucket", "dir/file.txt",
		"--cloud-project", "fakeProjectID",
		"--destination-bucket", "SBG_dest-bucket",
		"--destination-key", "dest/file.txt",
		"--storage-class", "STANDARD_IA",
		"-o", "json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{
		"message": "✅ Object dir/file.txt copied to dest/file.txt in bucket SBG_dest-bucket",
		"details": {"key": "dest/file.txt", "etag": "abc123"}
	}`))
}

func (ms *MockSuite) TestCloudStorageS3ObjectCopyMissingDestinationCmd(assert, require *td.T) {
	_, err := cmd.Execute("cloud", "storage", "object", "bucket", "object", "copy", "GRA_my-bucket", "file.txt",
		"--cloud-project", "fakeProjectID",
		"--destination-key", "copy.txt")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("destination-bucket"))
	assert.Cmp(httpmock.GetTotalCallCount(), 0)
}

func (ms *MockSuite) TestCloudStorageS3ObjectRestoreCmd(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(http.MethodPost, s3BucketURL+"/object/dir%2Ffile.txt/restore",
		tdhttpmock.JSONBody(td.JSON(`{"days": 7}`)),
		httpmock.NewStringResponder(202, `{"key": "dir/file.txt", "restoreStatus": {"inProgress": true}}`))

	out, err := cmd.Execute("cloud", "storage", "object", "bucket", "object", "restore", "GRA_my-bucket", "dir/file.txt",
		"--cloud-project", "fakeProjectID",
		"--days", "7",
		"-o", "json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{
		"message": "⚡️ Restore of object dir/file.txt started for 7 days",
		"details": {"key": "dir/file.txt", "restoreStatus": {"inProgress": true}}
	}`))
}

func (ms *MockSuite) TestCloudStorageS3ObjectVersionListCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, s3BucketURL+"/object/dir%2Ffile.txt/version",
		httpmock.NewStringResponder(200, `[
			{"key": "dir/file.txt", "versionId": "v2", "size": 20, "storageClass": "STANDARD", "isLatest": true, "lastModifiedAt": "2026-10-21T10:00:00Z"},
			{"key": "dir/file.txt", "versionId": "v1", "size": 12, "storageClass": "STANDARD", "isLatest": false, "lastModifiedAt": "2026-10-20T10:00:00Z"}
		]`))

	out, err := cmd.Execute("cloud", "storage", "object", "bucket", "object", "version", "list", "GRA_my-bucket", "dir/file.txt",
		"--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.Re(`v2\s+\S+\s+20\s+\S+\s+STANDARD\s+\S+\s+true`))
	assert.Cmp(out, td.Re(`v1\s+\S+\s+12\s+\S+\s+STANDARD\s+\S+\s+false`))
}

func (ms *MockSuite) TestCloudStorageS3ObjectVersionGetCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, s3BucketURL+"/object/dir%2Ffile.txt/version/v1",
		httpmock.NewStringResponder(200, `{
			"key": "dir/file.txt",
			"versionId": "v1",
			"size": 12,
			"storageClass": "STANDARD",
			"lastModifiedAt": "2026-10-20T10:00:00Z",
			"legalHold": "ON"
		}`))

	out, err := cmd.Execute("cloud", "storage", "object", "bucket", "object", "version", "get", "GRA_my-bucket", "dir/file.txt", "v1",
		"--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.Re(`Version ID\**:\s+v1`))
	assert.Cmp(out, td.Re(`Modification date\**:\s+2026-10-20T10:00:00Z`))
	assert.Cmp(out, td.Re(`Legal hold\**:\s+ON`))
	assert.Cmp(out, td.Not(td.Contains("<no value>")))
}

func (ms *MockSuite) TestCloudStorageS3ObjectVersionEditCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, s3BucketURL+"/object/dir%2Ffile.txt/version/v1",
		httpmock.NewStringResponder(200, `{"key": "dir/file.txt", "versionId": "v1", "legalHold": "OFF"}`))

	httpmock.RegisterMatcherResponder(http.MethodPut, s3BucketURL+"/object/dir%2Ffile.txt/version/v1",
		tdhttpmock.JSONBody(td.JSON(`{"legalHold": "ON"}`)),
		httpmock.NewStringResponder(200, `{"key": "dir/file.txt", "versionId": "v1", "legalHold": "ON"}`))

	out, err := cmd.Execute("cloud", "storage", "object", "bucket", "object", "version", "edit", "GRA_my-bucket", "dir/file.txt", "v1",
		"--cloud-project", "fakeProjectID",
		"--legal-hold", "ON")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("updated successfully"))
}

func (ms *MockSuite) TestCloudStorageS3ObjectVersionDeleteCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodDelete, s3BucketURL+"/object/dir%2Ffile.txt/version/v1",
		httpmock.NewStringResponder(200, `{"key": "dir/file.txt", "versionId": "v1"}`))

	out, err := cmd.Execute("cloud", "storage", "object", "bucket", "object", "version", "delete", "GRA_my-bucket", "dir/file.txt", "v1",
		"--cloud-project", "fakeProjectID",
		"-o", "json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{"message": "✅ Object dir/file.txt version v1 deleted from bucket GRA_my-bucket"}`))
}

func (ms *MockSuite) TestCloudStorageS3ObjectVersionCopyCmd(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(http.MethodPost, s3BucketURL+"/object/dir%2Ffile.txt/copy",
		tdhttpmock.JSONBody(td.JSON(`{"destinationBucket": "GRA_my-bucket", "destinationKey": "restored.txt", "versionId": "v1"}`)),
		httpmock.NewStringResponder(200, `{"key": "restored.txt", "versionId": "v3"}`))

	out, err := cmd.Execute("cloud", "storage", "object", "bucket", "object", "version", "copy", "GRA_my-bucket", "dir/file.txt", "v1",
		"--cloud-project", "fakeProjectID",
		"--destination-bucket", "GRA_my-bucket",
		"--destination-key", "restored.txt",
		"-o", "json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{
		"message": "✅ Object dir/file.txt (version v1) copied to restored.txt in bucket GRA_my-bucket",
		"details": {"key": "restored.txt", "versionId": "v3"}
	}`))
}

func (ms *MockSuite) TestCloudStorageS3ObjectVersionRestoreCmd(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(http.MethodPost, s3BucketURL+"/object/dir%2Ffile.txt/restore",
		tdhttpmock.JSONBody(td.JSON(`{"days": 14, "versionId": "v1"}`)),
		httpmock.NewStringResponder(202, `{"key": "dir/file.txt", "versionId": "v1"}`))

	out, err := cmd.Execute("cloud", "storage", "object", "bucket", "object", "version", "restore", "GRA_my-bucket", "dir/file.txt", "v1",
		"--cloud-project", "fakeProjectID",
		"--days", "14",
		"-o", "json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{
		"message": "⚡️ Restore of object dir/file.txt (version v1) started for 14 days",
		"details": {"key": "dir/file.txt", "versionId": "v1"}
	}`))
}

func (ms *MockSuite) TestCloudStorageS3BulkDeleteObjectsCmd(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(http.MethodPost, s3BucketURL+"/bulkDeleteObjects",
		tdhttpmock.JSONBody(td.JSON(`{"objects": [{"key": "a.txt"}, {"key": "b.txt", "versionId": "v1"}]}`)),
		httpmock.NewStringResponder(200, `{"deleted": [{"key": "a.txt"}, {"key": "b.txt", "versionId": "v1"}]}`))

	out, err := cmd.Execute("cloud", "storage", "object", "bucket", "bulk-delete", "GRA_my-bucket",
		"--cloud-project", "fakeProjectID",
		"--objects", "a.txt,b.txt:v1",
		"-o", "json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{"message": "✅ 2 objects deleted successfully"}`))
}

func (ms *MockSuite) TestCloudStorageS3BulkDeleteInvalidObjectCmd(assert, require *td.T) {
	_, err := cmd.Execute("cloud", "storage", "object", "bucket", "bulk-delete", "GRA_my-bucket",
		"--cloud-project", "fakeProjectID",
		"--objects", "a.txt:v1:extra")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("invalid object format"))
	assert.Cmp(httpmock.GetTotalCallCount(), 0)
}

func (ms *MockSuite) TestCloudStorageS3BulkDeletePrefixCmd(assert, require *td.T) {
	// The listing is paginated: both pages must be deleted.
	httpmock.RegisterResponder(http.MethodGet, s3BucketURL+"/object?prefix=logs%2F",
		func(req *http.Request) (*http.Response, error) {
			if req.Header.Get("X-Pagination-Cursor") == "" {
				resp := httpmock.NewStringResponse(200, `[{"key": "logs/1.log"}]`)
				resp.Header.Set("X-Pagination-Cursor-Next", "page2")
				return resp, nil
			}
			return httpmock.NewStringResponse(200, `[{"key": "logs/2.log"}]`), nil
		})

	httpmock.RegisterMatcherResponder(http.MethodPost, s3BucketURL+"/bulkDeleteObjects",
		tdhttpmock.JSONBody(td.JSON(`{"objects": [{"key": "logs/1.log"}, {"key": "logs/2.log"}]}`)),
		httpmock.NewStringResponder(200, `{"deleted": [{"key": "logs/1.log"}, {"key": "logs/2.log"}]}`))

	out, err := cmd.Execute("cloud", "storage", "object", "bucket", "bulk-delete", "GRA_my-bucket",
		"--cloud-project", "fakeProjectID",
		"--prefix", "logs/",
		"-o", "json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{"message": "✅ 2 objects deleted successfully"}`))
}

func (ms *MockSuite) TestCloudStorageS3BulkDeleteAllBatchesCmd(assert, require *td.T) {
	// The API takes at most 1000 objects per call.
	objects := make([]string, 1001)
	for i := range objects {
		objects[i] = fmt.Sprintf(`{"key": "o%d"}`, i)
	}
	httpmock.RegisterResponder(http.MethodGet, s3BucketURL+"/object",
		httpmock.NewStringResponder(200, "["+strings.Join(objects, ",")+"]"))

	var batchSizes []int
	httpmock.RegisterResponder(http.MethodPost, s3BucketURL+"/bulkDeleteObjects",
		func(req *http.Request) (*http.Response, error) {
			var body struct {
				Objects []map[string]any `json:"objects"`
			}
			if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
				return nil, err
			}
			batchSizes = append(batchSizes, len(body.Objects))
			deleted, _ := json.Marshal(map[string]any{"deleted": body.Objects})
			return httpmock.NewBytesResponse(200, deleted), nil
		})

	out, err := cmd.Execute("cloud", "storage", "object", "bucket", "bulk-delete", "GRA_my-bucket",
		"--cloud-project", "fakeProjectID",
		"--all",
		"-o", "json")

	require.CmpNoError(err)
	assert.Cmp(batchSizes, []int{1000, 1})
	assert.Cmp(json.RawMessage(out), td.JSON(`{"message": "✅ 1001 objects deleted successfully"}`))
}

func (ms *MockSuite) TestCloudStorageS3BulkDeleteBatchErrorCmd(assert, require *td.T) {
	objects := make([]string, 1001)
	for i := range objects {
		objects[i] = fmt.Sprintf(`{"key": "o%d"}`, i)
	}
	httpmock.RegisterResponder(http.MethodGet, s3BucketURL+"/object",
		httpmock.NewStringResponder(200, "["+strings.Join(objects, ",")+"]"))

	calls := 0
	httpmock.RegisterResponder(http.MethodPost, s3BucketURL+"/bulkDeleteObjects",
		func(req *http.Request) (*http.Response, error) {
			calls++
			if calls > 1 {
				return httpmock.NewStringResponse(500, `{"message": "boom"}`), nil
			}
			var body struct {
				Objects []map[string]any `json:"objects"`
			}
			if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
				return nil, err
			}
			deleted, _ := json.Marshal(map[string]any{"deleted": body.Objects})
			return httpmock.NewBytesResponse(200, deleted), nil
		})

	_, err := cmd.Execute("cloud", "storage", "object", "bucket", "bulk-delete", "GRA_my-bucket",
		"--cloud-project", "fakeProjectID",
		"--all")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("failed to delete objects after 1000 were deleted"))
}

func (ms *MockSuite) TestCloudStorageS3BulkDeletePartialFailureCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodPost, s3BucketURL+"/bulkDeleteObjects",
		httpmock.NewStringResponder(200, `{
			"deleted": [{"key": "a.txt"}],
			"errors": [{"key": "locked.txt", "code": "AccessDenied", "message": "Access Denied because object protected by object lock"}]
		}`))

	_, err := cmd.Execute("cloud", "storage", "object", "bucket", "bulk-delete", "GRA_my-bucket",
		"--cloud-project", "fakeProjectID",
		"--objects", "a.txt,locked.txt")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("1 objects deleted, 1 could not be deleted"))
	assert.Cmp(err.Error(), td.Contains("locked.txt: Access Denied because object protected by object lock"))
}

func (ms *MockSuite) TestCloudStorageS3PresignCmd(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(http.MethodPost, s3BucketURL+"/presign",
		tdhttpmock.JSONBody(td.JSON(`{"key": "dir/file.txt", "method": "PUT", "expiresIn": 600, "storageClass": "STANDARD"}`)),
		httpmock.NewStringResponder(201, `{
			"method": "PUT",
			"key": "dir/file.txt",
			"url": "https://my-bucket.s3.gra.io.cloud.ovh.net/dir/file.txt?X-Amz-Signature=s",
			"expiresAt": "2026-11-02T10:10:00Z",
			"signedHeaders": {"x-amz-storage-class": "STANDARD", "Host": "my-bucket.s3.gra.io.cloud.ovh.net"}
		}`))

	out, err := cmd.Execute("cloud", "storage", "object", "bucket", "presign", "GRA_my-bucket", "dir/file.txt",
		"--cloud-project", "fakeProjectID",
		"--method", "PUT",
		"--expires-in", "600",
		"--storage-class", "STANDARD")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("-> PUT https://my-bucket.s3.gra.io.cloud.ovh.net/dir/file.txt?X-Amz-Signature=s"))
	assert.Cmp(out, td.Contains("-> Expires at: 2026-11-02T10:10:00Z"))
	assert.Cmp(out, td.Re(`(?s)Host: my-bucket\.s3\.gra\.io\.cloud\.ovh\.net.*x-amz-storage-class: STANDARD`))
}

func (ms *MockSuite) TestCloudStorageS3PresignDefaultMethodCmd(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(http.MethodPost, s3BucketURL+"/presign",
		tdhttpmock.JSONBody(td.JSON(`{"key": "file.txt", "method": "GET"}`)),
		httpmock.NewStringResponder(201, `{"method": "GET", "key": "file.txt", "url": "https://u", "expiresAt": "2026-11-02T11:00:00Z"}`))

	out, err := cmd.Execute("cloud", "storage", "object", "bucket", "presign", "GRA_my-bucket", "file.txt",
		"--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.Not(td.Contains("Headers")))
}

func (ms *MockSuite) TestCloudStorageS3LifecycleGetCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, s3BucketURL+"/lifecycle",
		httpmock.NewStringResponder(200, `{
			"rules": [
				{
					"id": "expire-logs",
					"status": "ENABLED",
					"filter": {"prefix": "logs/"},
					"expiration": {"days": 30}
				},
				{
					"id": "archive",
					"status": "DISABLED",
					"transitions": [{"days": 90, "storageClass": "DEEP_ARCHIVE"}]
				}
			]
		}`))

	out, err := cmd.Execute("cloud", "storage", "object", "bucket", "lifecycle", "get", "GRA_my-bucket",
		"--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.Re(`expire-logs\s+\S+\s+ENABLED\s+\S+\s+logs/\s+\S+\s+30`))
	assert.Cmp(out, td.Re(`archive\s+\S+\s+DISABLED.*\s1\s`))
	assert.Cmp(out, td.Not(td.Contains("<no value>")))
}

func (ms *MockSuite) TestCloudStorageS3LifecycleGetEmptyCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, s3BucketURL+"/lifecycle",
		httpmock.NewStringResponder(200, `{"rules": []}`))

	out, err := cmd.Execute("cloud", "storage", "object", "bucket", "lifecycle", "get", "GRA_my-bucket",
		"--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("No lifecycle rule configured."))
}

func (ms *MockSuite) TestCloudStorageS3LifecycleEditFromFileCmd(assert, require *td.T) {
	// The file replaces the rules: they must not be appended to the current ones.
	rulesFile := filepath.Join(require.TempDir(), "rules.json")
	require.CmpNoError(os.WriteFile(rulesFile, []byte(`{"rules": [{"id": "new", "status": "ENABLED", "expiration": {"days": 7}}], "unknown": "dropped"}`), 0o600))

	httpmock.RegisterResponder(http.MethodGet, s3BucketURL+"/lifecycle",
		httpmock.NewStringResponder(200, `{"rules": [{"id": "old", "status": "ENABLED", "expiration": {"days": 1}}]}`))

	httpmock.RegisterMatcherResponder(http.MethodPut, s3BucketURL+"/lifecycle",
		tdhttpmock.JSONBody(td.JSON(`{"rules": [{"id": "new", "status": "ENABLED", "expiration": {"days": 7}}]}`)),
		httpmock.NewStringResponder(202, `{"rules": [{"id": "new", "status": "ENABLED", "expiration": {"days": 7}}]}`))

	out, err := cmd.Execute("cloud", "storage", "object", "bucket", "lifecycle", "edit", "GRA_my-bucket",
		"--cloud-project", "fakeProjectID",
		"--from-file", rulesFile,
		"-o", "json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{"message": "✅ Resource updated successfully"}`))
	assert.Cmp(httpmock.GetCallCountInfo()["GET "+s3BucketURL+"/lifecycle"], 0)
}

func (ms *MockSuite) TestCloudStorageS3LifecycleEditFromPipeCmd(assert, require *td.T) {
	rules := `{"rules": [{"id": "piped", "status": "ENABLED", "expiration": {"days": 3}}]}`
	stdin, stdinWriter, err := os.Pipe()
	require.CmpNoError(err)
	_, err = stdinWriter.WriteString(rules)
	require.CmpNoError(err)
	require.CmpNoError(stdinWriter.Close())

	originalStdin := os.Stdin
	os.Stdin = stdin
	defer func() { os.Stdin = originalStdin }()

	httpmock.RegisterMatcherResponder(http.MethodPut, s3BucketURL+"/lifecycle",
		tdhttpmock.JSONBody(td.JSON(rules)),
		httpmock.NewStringResponder(202, rules))

	out, err := cmd.Execute("cloud", "storage", "object", "bucket", "lifecycle", "edit", "GRA_my-bucket",
		"--cloud-project", "fakeProjectID",
		"-o", "json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{"message": "✅ Resource updated successfully"}`))
}

func (ms *MockSuite) TestCloudStorageS3LifecycleEditNoBucketCmd(assert, require *td.T) {
	_, err := cmd.Execute("cloud", "storage", "object", "bucket", "lifecycle", "edit",
		"--cloud-project", "fakeProjectID",
		"--from-file", "rules.json")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("bucket argument is required"))
	assert.Cmp(httpmock.GetTotalCallCount(), 0)
}

func (ms *MockSuite) TestCloudStorageS3LifecycleDeleteCmd(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(http.MethodPut, s3BucketURL+"/lifecycle",
		tdhttpmock.JSONBody(td.JSON(`{"rules": []}`)),
		httpmock.NewStringResponder(202, `{"rules": []}`))

	out, err := cmd.Execute("cloud", "storage", "object", "bucket", "lifecycle", "delete", "GRA_my-bucket",
		"--cloud-project", "fakeProjectID",
		"-o", "json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{"message": "✅ Bucket GRA_my-bucket lifecycle rules deleted successfully"}`))
}

func (ms *MockSuite) TestCloudStorageS3ReplicationGetCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, s3BucketURL+"/replication",
		httpmock.NewStringResponder(200, `{
			"rules": [
				{
					"id": "to-sbg",
					"priority": 1,
					"status": "ENABLED",
					"deleteMarkerReplication": "DISABLED",
					"filter": {"prefix": "data/"},
					"destination": {"name": "my-backup", "region": "SBG"}
				}
			]
		}`))

	out, err := cmd.Execute("cloud", "storage", "object", "bucket", "replication", "get", "GRA_my-bucket",
		"--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.Re(`to-sbg\s+\S+\s+1\s+\S+\s+ENABLED\s+\S+\s+my-backup \(SBG\)\s+\S+\s+data/\s+\S+\s+DISABLED`))
	assert.Cmp(out, td.Not(td.Contains("<no value>")))
}

func (ms *MockSuite) TestCloudStorageS3ReplicationEditFromFileCmd(assert, require *td.T) {
	rulesFile := filepath.Join(require.TempDir(), "rules.json")
	rules := `{"rules": [{"id": "r", "priority": 1, "status": "ENABLED", "deleteMarkerReplication": "DISABLED", "destination": {"name": "dst", "region": "SBG"}}]}`
	require.CmpNoError(os.WriteFile(rulesFile, []byte(rules), 0o600))

	httpmock.RegisterMatcherResponder(http.MethodPut, s3BucketURL+"/replication",
		tdhttpmock.JSONBody(td.JSON(rules)),
		httpmock.NewStringResponder(202, rules))

	out, err := cmd.Execute("cloud", "storage", "object", "bucket", "replication", "edit", "GRA_my-bucket",
		"--cloud-project", "fakeProjectID",
		"--from-file", rulesFile,
		"-o", "json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{"message": "✅ Resource updated successfully"}`))
}

func (ms *MockSuite) TestCloudStorageS3ReplicationDeleteCmd(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(http.MethodPut, s3BucketURL+"/replication",
		tdhttpmock.JSONBody(td.JSON(`{"rules": []}`)),
		httpmock.NewStringResponder(202, `{"rules": []}`))

	out, err := cmd.Execute("cloud", "storage", "object", "bucket", "replication", "delete", "GRA_my-bucket",
		"--cloud-project", "fakeProjectID",
		"-o", "json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{"message": "✅ Bucket GRA_my-bucket replication rules deleted successfully"}`))
}

func (ms *MockSuite) TestCloudStorageS3ReplicationDestinationListCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, s3BucketURL+"/replication/destination",
		httpmock.NewStringResponder(200, `[
			{"name": "my-backup", "region": "SBG", "eligible": true},
			{"name": "no-versioning", "region": "GRA", "eligible": false, "reason": "VERSIONING_DISABLED"}
		]`))

	out, err := cmd.Execute("cloud", "storage", "object", "bucket", "replication", "destination", "list", "GRA_my-bucket",
		"--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.Re(`my-backup\s+\S+\s+SBG\s+\S+\s+true`))
	assert.Cmp(out, td.Re(`no-versioning\s+\S+\s+GRA\s+\S+\s+false\s+\S+\s+VERSIONING_DISABLED`))
}

func (ms *MockSuite) TestCloudStorageS3ReplicationTriggerListCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, s3BucketURL+"/replication/trigger",
		httpmock.NewStringResponder(200, `[
			{"id": "job-1", "status": "COMPLETED", "total": 10, "completed": 9, "failed": 1, "createdAt": "2026-11-01T10:00:00Z"}
		]`))

	out, err := cmd.Execute("cloud", "storage", "object", "bucket", "replication", "trigger", "list", "GRA_my-bucket",
		"--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.Re(`job-1\s+\S+\s+COMPLETED\s+\S+\s+10\s+\S+\s+9\s+\S+\s+1\s+\S+\s+2026-11-01T10:00:00Z`))
}

func (ms *MockSuite) TestCloudStorageS3ReplicationTriggerGetCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, s3BucketURL+"/replication/trigger/job-1",
		httpmock.NewStringResponder(200, `{"id": "job-1", "status": "PREPARING", "createdAt": "2026-11-01T10:00:00Z"}`))

	out, err := cmd.Execute("cloud", "storage", "object", "bucket", "replication", "trigger", "get", "GRA_my-bucket", "job-1",
		"--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.Re(`Status\**:\s+PREPARING`))
	assert.Cmp(out, td.Re(`Total\**:\s+0`))
	assert.Cmp(out, td.Not(td.Contains("<no value>")))
}

func (ms *MockSuite) TestCloudStorageS3ReplicationTriggerCreateCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodPost, s3BucketURL+"/replication/trigger",
		httpmock.NewStringResponder(201, `{"id": "job-2", "status": "CREATED"}`))

	out, err := cmd.Execute("cloud", "storage", "object", "bucket", "replication", "trigger", "create", "GRA_my-bucket",
		"--cloud-project", "fakeProjectID",
		"-o", "json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{
		"message": "⚡️ Replication job job-2 started on bucket GRA_my-bucket",
		"details": {"id": "job-2", "status": "CREATED"}
	}`))
}
