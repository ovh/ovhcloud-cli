// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cloud

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/ovh/ovhcloud-cli/internal/assets"
	"github.com/ovh/ovhcloud-cli/internal/display"
	filtersLib "github.com/ovh/ovhcloud-cli/internal/filters"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
	"github.com/ovh/ovhcloud-cli/internal/services/common"
	"github.com/spf13/cobra"
)

var (
	cloudprojectStorageS3ColumnsToDisplay = []string{"name", "region", "createdAt"}

	//go:embed templates/cloud_storage_s3.tmpl
	cloudStorageS3Template string

	//go:embed templates/cloud_storage_s3_object.tmpl
	cloudStorageS3ObjectTemplate string

	//go:embed parameter-samples/storage-s3-create.json
	CloudStorageS3CreationExample string

	//go:embed parameter-samples/storage-s3-presigned-url.json
	CloudStorageS3PresignedURLExample string

	StorageS3Spec struct {
		Name       string `json:"name,omitempty"`
		OwnerId    int    `json:"ownerId,omitempty"`
		Encryption struct {
			SSEAlgorithm string `json:"sseAlgorithm,omitempty"`
		} `json:"encryption,omitzero"`
		ObjectLock struct {
			Rule struct {
				Mode   string `json:"mode,omitempty"`
				Period string `json:"period,omitempty"`
			} `json:"rule,omitzero"`
			Status string `json:"status,omitempty"`
		} `json:"objectLock,omitzero"`
		Replication struct {
			Rules []struct {
				DeleteMarkerReplication string `json:"deleteMarkerReplication,omitempty"`
				Destination             struct {
					Name         string `json:"name,omitempty"`
					Region       string `json:"region,omitempty"`
					StorageClass string `json:"storageClass,omitempty"`
				} `json:"destination,omitzero"`
				Filter struct {
					Prefix string            `json:"prefix,omitempty"`
					Tags   map[string]string `json:"tags,omitempty"`
				} `json:"filter,omitzero"`
				ID       string `json:"id,omitempty"`
				Priority int    `json:"priority,omitempty"`
				Status   string `json:"status,omitempty"`
			} `json:"rules,omitempty"`
		} `json:"replication,omitzero"`
		Tags       map[string]string `json:"tags,omitempty"`
		Versioning struct {
			Status string `json:"status,omitempty"`
		} `json:"versioning,omitzero"`
	}

	StorageS3ObjectsToDelete []string

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
		endpoint := fmt.Sprintf("/cloud/project/%s/region/%s/storage/%s",
			projectID, url.PathEscape(region.(string)), url.PathEscape(containerName))

		var container map[string]any
		if err := httpLib.Client.Get(endpoint, &container); err == nil {
			return endpoint, container, nil
		}
	}

	return "", nil, fmt.Errorf("no storage container found with name %s", containerName)
}

func ListCloudStorageS3(_ *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	// Fetch regions with storage feature available
	regions, err := getCloudRegionsWithFeatureAvailable(projectID, "storage-s3-high-perf", "storage-s3-standard")
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch regions with storage feature available: %s", err)
		return
	}

	// Fetch containers in all regions
	url := fmt.Sprintf("/cloud/project/%s/region", projectID)
	containers, err := httpLib.FetchObjectsParallel[[]map[string]any](url+"/%s/storage", regions, true)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch storage containers: %s", err)
		return
	}

	// Flatten containers in a single array
	var allContainers []map[string]any
	for _, regionContainers := range containers {
		allContainers = append(allContainers, regionContainers...)
	}

	// Filter results
	allContainers, err = filtersLib.FilterLines(allContainers, flags.GenericFilters)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to filter results: %s", err)
		return
	}

	display.RenderTable(allContainers, cloudprojectStorageS3ColumnsToDisplay, &flags.OutputFormatConfig)
}

func GetStorageS3(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	_, foundContainer, err := locateStorageS3Container(projectID, args[0])
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	// Convert used space to float
	usedFloat, err := foundContainer["objectsSize"].(json.Number).Float64()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error parsing used storage: %s", err)
		return
	}
	foundContainer["objectsSize"] = usedFloat

	display.OutputObject(foundContainer, args[0], cloudStorageS3Template, &flags.OutputFormatConfig)
}

func EditStorageS3(cmd *cobra.Command, args []string) {
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
		"/cloud/project/{serviceName}/region/{regionName}/storage/{name}",
		foundURL,
		StorageS3Spec,
		assets.CloudOpenapiSchema,
	); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}
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

	endpoint := fmt.Sprintf("/cloud/project/%s/region/%s/storage", projectID, url.PathEscape(args[0]))
	container, err := common.CreateResource(
		cmd,
		"/cloud/project/{serviceName}/region/{regionName}/storage",
		endpoint,
		CloudStorageS3CreationExample,
		StorageS3Spec,
		assets.CloudOpenapiSchema,
		[]string{"name"})
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to create s3 storage container: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, container, "✅ Container %s created successfully", container["name"])
}

func DeleteStorageS3(_ *cobra.Command, args []string) {
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

	if err := httpLib.Client.Delete(foundURL, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to delete storage container: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Storage container %s deleted successfully", args[0])
}

func StorageS3BulkDeleteObjects(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	if len(StorageS3ObjectsToDelete) == 0 {
		display.OutputWarning(&flags.OutputFormatConfig, "no objects to delete. Use --objects flag to specify objects to delete")
		return
	}

	foundURL, _, err := locateStorageS3Container(projectID, args[0])
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

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

	endpoint := fmt.Sprintf("/cloud/project/%s/user/%s/s3Credentials", projectID, url.PathEscape(args[0]))
	common.ManageListRequestNoExpand(endpoint, []string{"access", "userId", "tenantId"}, flags.GenericFilters)
}

func CreateStorageS3Credentials(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	credentials := map[string]any{}
	endpoint := fmt.Sprintf("/cloud/project/%s/user/%s/s3Credentials", projectID, url.PathEscape(args[0]))
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

	endpoint := fmt.Sprintf("/cloud/project/%s/user/%s/s3Credentials/%s", projectID, url.PathEscape(args[0]), url.PathEscape(args[1]))
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

	endpoint := fmt.Sprintf("/cloud/project/%s/user/%s/s3Credentials/%s", projectID, url.PathEscape(args[0]), url.PathEscape(args[1]))
	var credentials map[string]any
	if err := httpLib.Client.Get(endpoint, &credentials); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to get S3 credentials: %s", err)
		return
	}

	// Fetch credentials secret
	secretEndpoint := fmt.Sprintf("/cloud/project/%s/user/%s/s3Credentials/%s/secret", projectID, url.PathEscape(args[0]), url.PathEscape(args[1]))
	if err := httpLib.Client.Post(secretEndpoint, nil, &credentials); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to get S3 credentials secret: %s", err)
		return
	}

	display.OutputObject(credentials, args[1], "", &flags.OutputFormatConfig)
}
