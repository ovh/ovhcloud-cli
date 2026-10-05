// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cloud

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
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

	//go:embed parameter-samples/storage-object-create.json
	CloudStorageS3CreationExample string

	//go:embed parameter-samples/storage-object-presigned-url.json
	CloudStorageS3PresignedURLExample string

	//go:embed parameter-samples/storage-object-lifecycle.json
	CloudStorageS3LifecycleExample string

	StorageS3LifecycleSpec struct {
		Rules []struct {
			AbortIncompleteMultipartUpload *struct {
				DaysAfterInitiation int `json:"daysAfterInitiation,omitempty"`
			} `json:"abortIncompleteMultipartUpload,omitempty"`
			Expiration *struct {
				Days                      int    `json:"days,omitempty"`
				Date                      string `json:"date,omitempty"`
				ExpiredObjectDeleteMarker bool   `json:"expiredObjectDeleteMarker,omitempty"`
			} `json:"expiration,omitempty"`
			Filter *struct {
				Prefix string            `json:"prefix,omitempty"`
				Tags   map[string]string `json:"tags,omitempty"`
			} `json:"filter,omitempty"`
			ID                          string `json:"id,omitempty"`
			NoncurrentVersionExpiration *struct {
				NoncurrentDays int `json:"noncurrentDays,omitempty"`
			} `json:"noncurrentVersionExpiration,omitempty"`
			Status      string `json:"status"`
			Transitions []struct {
				Days         int    `json:"days,omitempty"`
				StorageClass string `json:"storageClass,omitempty"`
			} `json:"transitions,omitempty"`
		} `json:"rules,omitempty"`
	}

	StorageS3CopySpec struct {
		TargetBucket string `json:"targetBucket,omitempty"`
		TargetKey    string `json:"targetKey,omitempty"`
		StorageClass string `json:"storageClass,omitempty"`
	}

	StorageS3RestoreDays int

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

	StorageS3ListParams struct {
		KeyMarker       string
		Limit           int
		Prefix          string
		VersionIdMarker string
		WithVersions    bool
	}

	StorageS3ObjectSpec struct {
		LegalHold string `json:"legalHold,omitempty"`
		Lock      struct {
			Mode        string `json:"mode,omitempty"`
			RetainUntil string `json:"retainUntil,omitempty"`
		} `json:"lock,omitzero"`
	}

	StorageS3PresignedURLParams struct {
		Expire       int    `json:"expire,omitempty"`
		Method       string `json:"method,omitempty"`
		Object       string `json:"object,omitempty"`
		StorageClass string `json:"storageClass,omitempty"`
		VersionId    string `json:"versionId,omitempty"`
	}
)

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

	foundURL, _, err := locateStorageS3Container(projectID, args[0])
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	// List of objects to delete given, process them
	if len(StorageS3ObjectsToDelete) > 0 {
		var objectsToDelete []map[string]any
		for _, object := range StorageS3ObjectsToDelete {
			parts := strings.Split(object, ":")

			switch len(parts) {
			case 1:
				// Object name only
				objectsToDelete = append(objectsToDelete, map[string]any{"key": parts[0]})
			case 2:
				// Object name with version ID
				objectsToDelete = append(objectsToDelete, map[string]any{"key": parts[0], "versionId": parts[1]})
			default:
				display.OutputError(&flags.OutputFormatConfig, "invalid object format: %s. Use <object_name> or <object_name>:<version_id>", object)
				return
			}
		}

		if err := httpLib.Client.Post(foundURL+"/bulkDeleteObjects", map[string]any{
			"objects": objectsToDelete,
		}, nil); err != nil {
			display.OutputError(&flags.OutputFormatConfig, "failed to delete objects: %s", err)
			return
		}

		display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Objects deleted successfully")
		return
	}

	var request *http.Request
	switch {
	case StorageS3BulkDeletePrefix != "":
		endpoint := foundURL + "/object?prefix=" + url.QueryEscape(StorageS3BulkDeletePrefix)
		request, err = httpLib.Client.NewRequest(http.MethodGet, endpoint, nil, true)
	case StorageS3BulkDeleteAll:
		request, err = httpLib.Client.NewRequest(http.MethodGet, foundURL+"/object", nil, true)
	default:
		display.OutputError(&flags.OutputFormatConfig, "Nothing to delete, either --objects, --prefix or --all must be specified")
		return
	}

	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to create objects listing request: %s", err)
		return
	}

	for {
		// Fetch objects in the container (batches of 1000)
		resp, err := httpLib.Client.Do(request)
		if err != nil {
			display.OutputError(&flags.OutputFormatConfig, "failed to fetch objects: %s", err)
			return
		}

		var objects []map[string]any
		if err := httpLib.Client.UnmarshalResponse(resp, &objects); err != nil {
			display.OutputError(&flags.OutputFormatConfig, "failed to parse objects response: %s", err)
			return
		}

		// No objects found, we are done
		if len(objects) == 0 {
			break
		}

		// Prepare objects to delete
		var objectsToDelete []map[string]any
		for _, object := range objects {
			objectsToDelete = append(objectsToDelete, map[string]any{"key": object["key"]})
		}

		// Delete objects
		log.Printf("Deleting %d objects...", len(objectsToDelete))
		if err := httpLib.Client.Post(foundURL+"/bulkDeleteObjects", map[string]any{
			"objects": objectsToDelete,
		}, nil); err != nil {
			display.OutputError(&flags.OutputFormatConfig, "failed to delete objects: %s", err)
			return
		}
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Objects deleted successfully")
}

func ListStorageS3Objects(_ *cobra.Command, args []string) {
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

	params := make(url.Values)
	if StorageS3ListParams.KeyMarker != "" {
		params.Set("keyMarker", StorageS3ListParams.KeyMarker)
	}
	if StorageS3ListParams.Limit > 0 {
		params.Set("limit", strconv.Itoa(StorageS3ListParams.Limit))
	}
	if StorageS3ListParams.Prefix != "" {
		params.Set("prefix", StorageS3ListParams.Prefix)
	}
	if StorageS3ListParams.VersionIdMarker != "" {
		params.Set("versionIdMarker", StorageS3ListParams.VersionIdMarker)
	}
	if StorageS3ListParams.WithVersions {
		params.Set("withVersions", "true")
	}

	endpoint := fmt.Sprintf("%s/object?%s", foundURL, params.Encode())

	common.ManageListRequestNoExpand(endpoint, []string{"key", "size"}, flags.GenericFilters)
}

func GetStorageS3Object(_ *cobra.Command, args []string) {
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

	common.ManageObjectRequest(foundURL+"/object", args[1], cloudStorageS3ObjectTemplate)
}

func EditStorageS3Object(cmd *cobra.Command, args []string) {
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

	if err := common.EditResource(
		cmd,
		"/cloud/project/{serviceName}/region/{regionName}/storage/{name}/object/{key}",
		foundURL+"/object/"+url.PathEscape(args[1]),
		StorageS3ObjectSpec,
		assets.CloudOpenapiSchema,
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

	foundURL, _, err := locateStorageS3Container(projectID, args[0])
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	if err := httpLib.Client.Delete(foundURL+"/object/"+url.PathEscape(args[1]), nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to delete object: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Object %s deleted successfully", args[1])
}

func ListStorageS3ObjectVersions(_ *cobra.Command, args []string) {
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

	params := make(url.Values)
	if StorageS3ListParams.VersionIdMarker != "" {
		params.Set("versionIdMarker", StorageS3ListParams.VersionIdMarker)
	}
	if StorageS3ListParams.Limit > 0 {
		params.Set("limit", strconv.Itoa(StorageS3ListParams.Limit))
	}

	endpoint := fmt.Sprintf("%s/object/%s/version?%s", foundURL, url.PathEscape(args[1]), params.Encode())

	common.ManageListRequestNoExpand(endpoint, []string{"versionId", "size", "isLatest"}, flags.GenericFilters)
}

func GetStorageS3ObjectVersion(_ *cobra.Command, args []string) {
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

	endpoint := fmt.Sprintf("%s/object/%s/version", foundURL, url.PathEscape(args[1]))

	common.ManageObjectRequest(endpoint, args[2], cloudStorageS3ObjectTemplate)
}

func EditStorageS3ObjectVersion(cmd *cobra.Command, args []string) {
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

	if err := common.EditResource(
		cmd,
		"/cloud/project/{serviceName}/region/{regionName}/storage/{name}/object/{key}/version/{versionId}",
		foundURL+"/object/"+url.PathEscape(args[1])+"/version/"+url.PathEscape(args[2]),
		StorageS3ObjectSpec,
		assets.CloudOpenapiSchema,
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

	foundURL, _, err := locateStorageS3Container(projectID, args[0])
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	if err := httpLib.Client.Delete(foundURL+"/object/"+url.PathEscape(args[1])+"/version/"+url.PathEscape(args[2]), nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to delete object version: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Object version %s deleted successfully", args[2])
}

func StorageS3GeneratePresignedURL(cmd *cobra.Command, args []string) {
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

	response, err := common.CreateResource(
		cmd,
		"/cloud/project/{serviceName}/region/{regionName}/storage/{name}/presign",
		foundURL+"/presign",
		CloudStorageS3PresignedURLExample,
		StorageS3PresignedURLParams,
		assets.CloudOpenapiSchema,
		nil)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to generate presigned URL: %s", err)
		return
	}

	var sb strings.Builder
	sb.WriteString("✅ Presigned URL generated successfully:\n")
	sb.WriteString(fmt.Sprintf("-> %s %s\n", response["method"], response["url"]))
	if headers, ok := response["signedHeaders"].(map[string]any); ok {
		sb.WriteString("-> Headers:\n")
		for key, value := range headers {
			sb.WriteString(fmt.Sprintf("   - %s: %s\n", key, value))
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

// Lifecycle management

func GetStorageS3Lifecycle(_ *cobra.Command, args []string) {
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

	var lifecycle map[string]any
	if err := httpLib.Client.Get(foundURL+"/lifecycle", &lifecycle); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to get lifecycle configuration: %s", err)
		return
	}

	display.OutputObject(lifecycle, args[0], "", &flags.OutputFormatConfig)
}

func EditStorageS3Lifecycle(cmd *cobra.Command, args []string) {
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

	if err := common.EditResource(
		cmd,
		"/cloud/project/{serviceName}/region/{regionName}/storage/{name}/lifecycle",
		foundURL+"/lifecycle",
		StorageS3LifecycleSpec,
		assets.CloudOpenapiSchema,
	); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}
}

func DeleteStorageS3Lifecycle(_ *cobra.Command, args []string) {
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

	if err := httpLib.Client.Delete(foundURL+"/lifecycle", nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to delete lifecycle configuration: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Lifecycle configuration for container %s deleted successfully", args[0])
}

// Object copy and restore

func CopyStorageS3Object(_ *cobra.Command, args []string) {
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

	endpoint := foundURL + "/object/" + url.PathEscape(args[1]) + "/copy"
	var result map[string]any
	if err := httpLib.Client.Post(endpoint, StorageS3CopySpec, &result); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to copy object: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, result, "✅ Object %s copied successfully", args[1])
}

func RestoreStorageS3Object(_ *cobra.Command, args []string) {
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

	endpoint := foundURL + "/object/" + url.PathEscape(args[1]) + "/restore"
	if err := httpLib.Client.Post(endpoint, map[string]any{"days": int(StorageS3RestoreDays)}, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to restore object: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Object %s restore initiated successfully", args[1])
}

// Object version copy and restore

func CopyStorageS3ObjectVersion(_ *cobra.Command, args []string) {
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

	endpoint := foundURL + "/object/" + url.PathEscape(args[1]) + "/version/" + url.PathEscape(args[2]) + "/copy"
	var result map[string]any
	if err := httpLib.Client.Post(endpoint, StorageS3CopySpec, &result); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to copy object version: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, result, "✅ Object %s version %s copied successfully", args[1], args[2])
}

func RestoreStorageS3ObjectVersion(_ *cobra.Command, args []string) {
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

	endpoint := foundURL + "/object/" + url.PathEscape(args[1]) + "/version/" + url.PathEscape(args[2]) + "/restore"
	if err := httpLib.Client.Post(endpoint, map[string]any{"days": int(StorageS3RestoreDays)}, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to restore object version: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Object %s version %s restore initiated successfully", args[1], args[2])
}

// Replication job

func CreateStorageS3ReplicationJob(_ *cobra.Command, args []string) {
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

	var result map[string]any
	if err := httpLib.Client.Post(foundURL+"/job/replication", nil, &result); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to create replication job: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, result, "✅ Replication job created successfully (ID: %s)", result["id"])
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
