// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cloud

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"maps"
	"net/url"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/ovh/ovhcloud-cli/internal/assets"
	"github.com/ovh/ovhcloud-cli/internal/display"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
	"github.com/ovh/ovhcloud-cli/internal/services/common"
	"github.com/spf13/cobra"
)

var (
	cloudprojectStorageS3ColumnsToDisplay = []string{
		"id",
		"targetSpec.name name",
		"targetSpec.location.region region",
		"resourceStatus status",
		"createdAt",
	}

	//go:embed templates/cloud_storage_object.tmpl
	cloudStorageS3Template string

	//go:embed templates/cloud_storage_object_object.tmpl
	cloudStorageS3ObjectTemplate string

	//go:embed templates/cloud_storage_object_lifecycle.tmpl
	cloudStorageS3LifecycleTemplate string

	//go:embed templates/cloud_storage_object_replication.tmpl
	cloudStorageS3ReplicationTemplate string

	//go:embed templates/cloud_storage_object_replication_job.tmpl
	cloudStorageS3ReplicationJobTemplate string

	//go:embed parameter-samples/storage-object-create.json
	CloudStorageS3CreationExample string

	//go:embed parameter-samples/storage-object-lifecycle.json
	CloudStorageS3LifecycleExample string

	//go:embed parameter-samples/storage-object-replication.json
	CloudStorageS3ReplicationExample string

	cloudStorageS3ReplicationDestinationColumnsToDisplay = []string{
		"name",
		"region",
		"eligible",
		"reason",
	}

	cloudStorageS3ReplicationJobColumnsToDisplay = []string{
		"id",
		"status",
		"total",
		"completed",
		"failed",
		"createdAt",
	}

	StorageS3CopySpec struct {
		DestinationBucket string `json:"destinationBucket,omitempty"`
		DestinationKey    string `json:"destinationKey,omitempty"`
		StorageClass      string `json:"storageClass,omitempty"`
		VersionId         string `json:"versionId,omitempty"`
	}

	StorageS3RestoreSpec struct {
		Days      int    `json:"days"`
		VersionId string `json:"versionId,omitempty"`
	}

	StorageS3QuotaSpec struct {
		QuotaBytes int64 `json:"quotaBytes"`
	}

	BucketSpec struct {
		TargetSpec struct {
			Name     string `json:"name,omitempty"`
			Location struct {
				Region string `json:"region,omitempty"`
			} `json:"location,omitzero"`
			Encryption struct {
				Algorithm string `json:"algorithm,omitempty"`
			} `json:"encryption,omitzero"`
			ObjectLock struct {
				Mode          string `json:"mode,omitempty"`
				RetentionDays int    `json:"retentionDays,omitempty"`
			} `json:"objectLock,omitzero"`
			OwnerUserId string            `json:"ownerUserId,omitempty"`
			Tags        map[string]string `json:"tags,omitempty"`
			Versioning  struct {
				Status string `json:"status,omitempty"`
			} `json:"versioning,omitzero"`
		} `json:"targetSpec"`
	}

	BucketEditSpec struct {
		TargetSpec struct {
			Encryption struct {
				Algorithm string `json:"algorithm,omitempty"`
			} `json:"encryption,omitzero"`
			ObjectLock struct {
				Mode          string `json:"mode,omitempty"`
				RetentionDays int    `json:"retentionDays,omitempty"`
			} `json:"objectLock,omitzero"`
			OwnerUserId string            `json:"ownerUserId,omitempty"`
			Tags        map[string]string `json:"tags,omitempty"`
			Versioning  struct {
				Status string `json:"status,omitempty"`
			} `json:"versioning,omitzero"`
		} `json:"targetSpec,omitzero"`
	}

	StorageS3ObjectsToDelete  []string
	StorageS3BulkDeleteAll    bool
	StorageS3BulkDeletePrefix string

	StorageS3ObjectListParams struct {
		Delimiter    string
		Prefix       string
		WithVersions bool
	}

	StorageS3ObjectSpec struct {
		LegalHold string `json:"legalHold,omitempty"`
		Lock      struct {
			Mode        string `json:"mode,omitempty"`
			RetainUntil string `json:"retainUntil,omitempty"`
		} `json:"lock,omitzero"`
	}

	StorageS3PresignSpec struct {
		Key          string `json:"key"`
		Method       string `json:"method"`
		ExpiresIn    int    `json:"expiresIn,omitempty"`
		StorageClass string `json:"storageClass,omitempty"`
		VersionId    string `json:"versionId,omitempty"`
	}
)

// bulkDeleteMaxObjects is the API cap on objects per bulkDeleteObjects call.
const bulkDeleteMaxObjects = 1000

// v1 lookup for add-user: API v2 has no bucket user policy route.
func locateStorageS3Container(projectID, containerName string) (string, map[string]any, error) {
	// Fetch regions with storage feature available
	regions, err := getCloudRegionsWithFeatureAvailable(projectID, "storage-s3-high-perf", "storage-s3-standard")
	if err != nil {
		return "", nil, fmt.Errorf("failed to fetch regions with storage feature available: %w", err)
	}

	// Search for the given container in all regions
	for _, region := range regions {
		endpoint := fmt.Sprintf("/v1/cloud/project/%s/region/%s/storage/%s",
			projectID, url.PathEscape(region.(string)), url.PathEscape(containerName))

		var container map[string]any
		if err := httpLib.Client.Get(endpoint, &container); err == nil {
			return endpoint, container, nil
		}
	}

	return "", nil, fmt.Errorf("no storage container found with name %s", containerName)
}

func bucketV2Endpoint(projectID string) string {
	return fmt.Sprintf("/v2/publicCloud/project/%s/storage/object/bucket", projectID)
}

func bucketEndpoint(projectID, bucketID string) string {
	return fmt.Sprintf("%s/%s", bucketV2Endpoint(projectID), url.PathEscape(bucketID))
}

func bucketObjectEndpoint(projectID, bucketID, key string) string {
	return fmt.Sprintf("%s/object/%s", bucketEndpoint(projectID, bucketID), url.PathEscape(key))
}

func ListCloudStorageS3(_ *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.ManageListRequestNoExpand(bucketV2Endpoint(projectID), cloudprojectStorageS3ColumnsToDisplay, flags.GenericFilters)
}

func GetStorageS3(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	var bucket map[string]any
	endpoint := fmt.Sprintf("%s/%s", bucketV2Endpoint(projectID), url.PathEscape(args[0]))
	if err := httpLib.Client.Get(endpoint, &bucket); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch bucket: %s", err)
		return
	}

	// Convert used space to float
	if state, ok := bucket["currentState"].(map[string]any); ok {
		if size, ok := state["objectsSize"].(json.Number); ok {
			sizeFloat, err := size.Float64()
			if err != nil {
				display.OutputError(&flags.OutputFormatConfig, "error parsing used storage: %s", err)
				return
			}
			state["objectsSize"] = sizeFloat
		}
	}

	display.OutputObject(bucket, args[0], cloudStorageS3Template, &flags.OutputFormatConfig)
}

func EditStorageS3(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("%s/%s", bucketV2Endpoint(projectID), url.PathEscape(args[0]))
	if err := common.EditResource(
		cmd,
		"/publicCloud/project/{projectId}/storage/object/bucket/{bucketName}",
		endpoint,
		BucketEditSpec,
		assets.CloudV2OpenapiSchema,
	); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	if !flags.WaitForTask {
		return
	}

	ready, err := waitForCloudResourceReady(endpoint, 10*time.Minute)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to wait for bucket to be ready: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, ready, "✅ Bucket %s is now ready", args[0])
}

func CreateStorageS3(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		display.OutputError(&flags.OutputFormatConfig, "region argument is required\n\n%s", cmd.UsageString())
		return
	}

	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	BucketSpec.TargetSpec.Location.Region = args[0]
	endpoint := bucketV2Endpoint(projectID)
	bucket, err := common.CreateResource(
		cmd,
		"/publicCloud/project/{projectId}/storage/object/bucket",
		endpoint,
		CloudStorageS3CreationExample,
		BucketSpec,
		assets.CloudV2OpenapiSchema,
		[]string{"targetSpec.name"},
	)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to create bucket: %s", err)
		return
	}

	bucketID, _ := bucket["id"].(string)

	if !flags.WaitForTask {
		display.OutputInfo(&flags.OutputFormatConfig, bucket, "✅ Bucket %s creation started successfully", bucketID)
		return
	}

	ready, err := waitForCloudResourceReady(fmt.Sprintf("%s/%s", endpoint, url.PathEscape(bucketID)), 10*time.Minute)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to wait for bucket creation: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, ready, "✅ Bucket %s created successfully", bucketID)
}

func DeleteStorageS3(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("%s/%s", bucketV2Endpoint(projectID), url.PathEscape(args[0]))
	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to delete bucket: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Bucket %s is being deleted", args[0])
}

func StorageS3BulkDeleteObjects(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	objects, err := objectsToBulkDelete(bucketEndpoint(projectID, args[0]))
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	if len(objects) == 0 {
		display.OutputInfo(&flags.OutputFormatConfig, nil, "🟠 No object to delete in bucket %s", args[0])
		return
	}

	var (
		deleted  []any
		failures []any
	)
	for start := 0; start < len(objects); start += bulkDeleteMaxObjects {
		batch := objects[start:min(start+bulkDeleteMaxObjects, len(objects))]
		log.Printf("Deleting %d objects...", len(batch))

		var result map[string]any
		if err := httpLib.Client.Post(bucketEndpoint(projectID, args[0])+"/bulkDeleteObjects", map[string]any{
			"objects": batch,
		}, &result); err != nil {
			display.OutputError(&flags.OutputFormatConfig, "failed to delete objects: %s", err)
			return
		}

		if batchDeleted, ok := result["deleted"].([]any); ok {
			deleted = append(deleted, batchDeleted...)
		}
		if batchErrors, ok := result["errors"].([]any); ok {
			failures = append(failures, batchErrors...)
		}
	}

	if len(failures) > 0 {
		var sb strings.Builder
		for _, failure := range failures {
			if f, ok := failure.(map[string]any); ok {
				sb.WriteString(fmt.Sprintf("\n- %s: %s", f["key"], f["message"]))
			}
		}
		display.OutputError(&flags.OutputFormatConfig, "%d objects deleted, %d could not be deleted:%s", len(deleted), len(failures), sb.String())
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ %d objects deleted successfully", len(deleted))
}

func objectsToBulkDelete(bucketURL string) ([]map[string]any, error) {
	var objects []map[string]any

	if len(StorageS3ObjectsToDelete) > 0 {
		for _, object := range StorageS3ObjectsToDelete {
			key, versionID, hasVersion := strings.Cut(object, ":")
			if hasVersion && (versionID == "" || strings.Contains(versionID, ":")) {
				return nil, fmt.Errorf("invalid object format: %s. Use <object_key> or <object_key>:<version_id>", object)
			}

			identifier := map[string]any{"key": key}
			if hasVersion {
				identifier["versionId"] = versionID
			}
			objects = append(objects, identifier)
		}
		return objects, nil
	}

	listURL := bucketURL + "/object"
	switch {
	case StorageS3BulkDeletePrefix != "":
		listURL += "?" + url.Values{"prefix": {StorageS3BulkDeletePrefix}}.Encode()
	case StorageS3BulkDeleteAll:
	default:
		return nil, errors.New("nothing to delete, either --objects, --prefix or --all must be specified")
	}

	listed, err := httpLib.FetchArray(listURL, "")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch objects: %w", err)
	}

	for _, item := range listed {
		if object, ok := item.(map[string]any); ok {
			objects = append(objects, map[string]any{"key": object["key"]})
		}
	}

	return objects, nil
}

func ListStorageS3Objects(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	params := make(url.Values)
	if StorageS3ObjectListParams.Delimiter != "" {
		params.Set("delimiter", StorageS3ObjectListParams.Delimiter)
	}
	if StorageS3ObjectListParams.Prefix != "" {
		params.Set("prefix", StorageS3ObjectListParams.Prefix)
	}

	columns := []string{"key", "size", "storageClass", "lastModifiedAt"}
	if StorageS3ObjectListParams.WithVersions {
		params.Set("withVersions", "true")
		columns = append(columns, "versionId", "isLatest", "isDeleteMarker")
	}

	endpoint := bucketEndpoint(projectID, args[0]) + "/object"
	if len(params) > 0 {
		endpoint += "?" + params.Encode()
	}

	common.ManageListRequestNoExpand(endpoint, columns, flags.GenericFilters)
}

func GetStorageS3Object(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.ManageObjectRequest(bucketEndpoint(projectID, args[0])+"/object", args[1], cloudStorageS3ObjectTemplate)
}

func EditStorageS3Object(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	if err := common.EditResource(
		cmd,
		"/publicCloud/project/{projectId}/storage/object/bucket/{bucketName}/object/{key}",
		bucketObjectEndpoint(projectID, args[0], args[1]),
		StorageS3ObjectSpec,
		assets.CloudV2OpenapiSchema,
	); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}
}

func DeleteStorageS3Object(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	if err := httpLib.Client.Delete(bucketObjectEndpoint(projectID, args[0], args[1]), nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to delete object: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Object %s deleted from bucket %s", args[1], args[0])
}

func ListStorageS3ObjectVersions(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.ManageListRequestNoExpand(bucketObjectEndpoint(projectID, args[0], args[1])+"/version",
		[]string{"versionId", "size", "storageClass", "isLatest", "isDeleteMarker", "lastModifiedAt"}, flags.GenericFilters)
}

func GetStorageS3ObjectVersion(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.ManageObjectRequest(bucketObjectEndpoint(projectID, args[0], args[1])+"/version", args[2], cloudStorageS3ObjectTemplate)
}

func EditStorageS3ObjectVersion(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	if err := common.EditResource(
		cmd,
		"/publicCloud/project/{projectId}/storage/object/bucket/{bucketName}/object/{key}/version/{versionId}",
		bucketObjectEndpoint(projectID, args[0], args[1])+"/version/"+url.PathEscape(args[2]),
		StorageS3ObjectSpec,
		assets.CloudV2OpenapiSchema,
	); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}
}

func DeleteStorageS3ObjectVersion(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	if err := httpLib.Client.Delete(bucketObjectEndpoint(projectID, args[0], args[1])+"/version/"+url.PathEscape(args[2]), nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to delete object version: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Object %s version %s deleted from bucket %s", args[1], args[2], args[0])
}

func PresignStorageS3(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	StorageS3PresignSpec.Key = args[1]

	var response map[string]any
	if err := httpLib.Client.Post(bucketEndpoint(projectID, args[0])+"/presign", StorageS3PresignSpec, &response); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to generate presigned URL: %s", err)
		return
	}

	var sb strings.Builder
	sb.WriteString("✅ Presigned URL generated successfully:\n")
	sb.WriteString(fmt.Sprintf("-> %s %s\n", response["method"], response["url"]))
	sb.WriteString(fmt.Sprintf("-> Expires at: %s", response["expiresAt"]))
	if headers, ok := response["signedHeaders"].(map[string]any); ok && len(headers) > 0 {
		sb.WriteString("\n-> Headers to send with the request:")
		for _, key := range slices.Sorted(maps.Keys(headers)) {
			sb.WriteString(fmt.Sprintf("\n   - %s: %s", key, headers[key]))
		}
	}

	display.OutputInfo(&flags.OutputFormatConfig, response, "%s", &sb)
}

func StorageS3AddUser(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	foundURL, _, err := locateStorageS3Container(projectID, args[0])
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	userID := args[1]
	userRole := args[2]
	endpoint := foundURL + "/policy/" + url.PathEscape(userID)

	if err := httpLib.Client.Post(endpoint, map[string]any{
		"roleName": userRole,
	}, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to add user: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ User %s successfully added to the bucket", args[1])
}

func ListStorageS3Credentials(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/user/%s/s3Credentials", projectID, url.PathEscape(args[0]))
	common.ManageListRequestNoExpand(endpoint, []string{"access", "userId", "tenantId"}, flags.GenericFilters)
}

func CreateStorageS3Credentials(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	credentials := map[string]any{}
	endpoint := fmt.Sprintf("/v1/cloud/project/%s/user/%s/s3Credentials", projectID, url.PathEscape(args[0]))
	if err := httpLib.Client.Post(endpoint, nil, &credentials); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to create S3 credentials: %s", err)
		return
	}

	display.OutputObject(credentials, args[0], "", &flags.OutputFormatConfig)
}

func DeleteStorageS3Credentials(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/user/%s/s3Credentials/%s", projectID, url.PathEscape(args[0]), url.PathEscape(args[1]))
	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to delete S3 credentials: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ S3 credentials %s for user %s deleted successfully", args[1], args[0])
}

func GetStorageS3Credentials(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/user/%s/s3Credentials/%s", projectID, url.PathEscape(args[0]), url.PathEscape(args[1]))
	var credentials map[string]any
	if err := httpLib.Client.Get(endpoint, &credentials); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to get S3 credentials: %s", err)
		return
	}

	// Fetch credentials secret
	secretEndpoint := fmt.Sprintf("/v1/cloud/project/%s/user/%s/s3Credentials/%s/secret", projectID, url.PathEscape(args[0]), url.PathEscape(args[1]))
	if err := httpLib.Client.Post(secretEndpoint, nil, &credentials); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to get S3 credentials secret: %s", err)
		return
	}

	display.OutputObject(credentials, args[1], "", &flags.OutputFormatConfig)
}

func GetStorageS3Lifecycle(_ *cobra.Command, args []string) {
	getBucketRules(args[0], "lifecycle", cloudStorageS3LifecycleTemplate)
}

func EditStorageS3Lifecycle(cmd *cobra.Command, args []string) {
	editBucketRules(cmd, args, "lifecycle")
}

func DeleteStorageS3Lifecycle(_ *cobra.Command, args []string) {
	deleteBucketRules(args[0], "lifecycle")
}

func GetStorageS3Replication(_ *cobra.Command, args []string) {
	getBucketRules(args[0], "replication", cloudStorageS3ReplicationTemplate)
}

func EditStorageS3Replication(cmd *cobra.Command, args []string) {
	editBucketRules(cmd, args, "replication")
}

func DeleteStorageS3Replication(_ *cobra.Command, args []string) {
	deleteBucketRules(args[0], "replication")
}

func getBucketRules(bucketID, configuration, template string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	var rules map[string]any
	if err := httpLib.Client.Get(bucketEndpoint(projectID, bucketID)+"/"+configuration, &rules); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to get %s configuration: %s", configuration, err)
		return
	}

	display.OutputObject(rules, bucketID, template, &flags.OutputFormatConfig)
}

func editBucketRules(cmd *cobra.Command, args []string, configuration string) {
	if len(args) == 0 {
		display.OutputError(&flags.OutputFormatConfig, "bucket argument is required\n\n%s", cmd.UsageString())
		return
	}
	bucketID := args[0]

	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := bucketEndpoint(projectID, bucketID) + "/" + configuration

	// EditResource appends the file's rules to the fetched ones, but a PUT replaces all rules.
	if flags.ParametersFile != "" {
		fd, err := os.Open(flags.ParametersFile)
		if err != nil {
			display.OutputError(&flags.OutputFormatConfig, "failed to open given file: %s", err)
			return
		}
		defer fd.Close()

		var rules map[string]any
		if err := json.NewDecoder(fd).Decode(&rules); err != nil {
			display.OutputError(&flags.OutputFormatConfig, "failed to parse given file: %s", err)
			return
		}

		if err := httpLib.Client.Put(endpoint, rules, nil); err != nil {
			display.OutputError(&flags.OutputFormatConfig, "failed to update %s configuration: %s", configuration, err)
			return
		}

		display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Bucket %s %s rules updated successfully", bucketID, configuration)
		return
	}

	if err := common.EditResource(
		cmd,
		"/publicCloud/project/{projectId}/storage/object/bucket/{bucketName}/"+configuration,
		endpoint,
		struct{}{},
		assets.CloudV2OpenapiSchema,
	); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
	}
}

func deleteBucketRules(bucketID, configuration string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	if err := httpLib.Client.Put(bucketEndpoint(projectID, bucketID)+"/"+configuration, map[string]any{"rules": []any{}}, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to delete %s configuration: %s", configuration, err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Bucket %s %s rules deleted successfully", bucketID, configuration)
}

func ListStorageS3ReplicationDestinations(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.ManageListRequestNoExpand(bucketEndpoint(projectID, args[0])+"/replication/destination", cloudStorageS3ReplicationDestinationColumnsToDisplay, flags.GenericFilters)
}

func ListStorageS3ReplicationJobs(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.ManageListRequestNoExpand(bucketEndpoint(projectID, args[0])+"/replication/trigger", cloudStorageS3ReplicationJobColumnsToDisplay, flags.GenericFilters)
}

func GetStorageS3ReplicationJob(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.ManageObjectRequest(bucketEndpoint(projectID, args[0])+"/replication/trigger", args[1], cloudStorageS3ReplicationJobTemplate)
}

func CreateStorageS3ReplicationJob(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	var job map[string]any
	if err := httpLib.Client.Post(bucketEndpoint(projectID, args[0])+"/replication/trigger", nil, &job); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to create replication job: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, job, "⚡️ Replication job %s started on bucket %s", job["id"], args[0])
}

func CopyStorageS3Object(_ *cobra.Command, args []string) {
	copyStorageS3Object(args[0], args[1], "")
}

func CopyStorageS3ObjectVersion(_ *cobra.Command, args []string) {
	copyStorageS3Object(args[0], args[1], args[2])
}

func copyStorageS3Object(bucketID, key, versionID string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	StorageS3CopySpec.VersionId = versionID

	var result map[string]any
	if err := httpLib.Client.Post(bucketObjectEndpoint(projectID, bucketID, key)+"/copy", StorageS3CopySpec, &result); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to copy object: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, result, "✅ Object %s copied to %s in bucket %s",
		objectWithVersion(key, versionID), StorageS3CopySpec.DestinationKey, StorageS3CopySpec.DestinationBucket)
}

func RestoreStorageS3Object(_ *cobra.Command, args []string) {
	restoreStorageS3Object(args[0], args[1], "")
}

func RestoreStorageS3ObjectVersion(_ *cobra.Command, args []string) {
	restoreStorageS3Object(args[0], args[1], args[2])
}

func restoreStorageS3Object(bucketID, key, versionID string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	StorageS3RestoreSpec.VersionId = versionID

	var result map[string]any
	if err := httpLib.Client.Post(bucketObjectEndpoint(projectID, bucketID, key)+"/restore", StorageS3RestoreSpec, &result); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to restore object: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, result, "⚡️ Restore of object %s started for %d days", objectWithVersion(key, versionID), StorageS3RestoreSpec.Days)
}

func objectWithVersion(key, versionID string) string {
	if versionID == "" {
		return key
	}
	return key + " (version " + versionID + ")"
}

// Storage quota

func GetStorageS3Quota(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/region/%s/quota/storage", projectID, url.PathEscape(args[0]))
	var quota map[string]any
	if err := httpLib.Client.Get(endpoint, &quota); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to get storage quota: %s", err)
		return
	}

	display.OutputObject(quota, args[0], "", &flags.OutputFormatConfig)
}

func EditStorageS3Quota(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/region/%s/quota/storage", projectID, url.PathEscape(args[0]))
	if err := httpLib.Client.Put(endpoint, StorageS3QuotaSpec, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to update storage quota: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Storage quota for region %s updated successfully", args[0])
}

func DeleteStorageS3Quota(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/region/%s/quota/storage", projectID, url.PathEscape(args[0]))
	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to delete storage quota: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Storage quota for region %s deleted successfully", args[0])
}
