// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cloud

import (
	_ "embed"
	"fmt"
	"net/url"
	"time"

	"github.com/ovh/ovhcloud-cli/internal/assets"
	"github.com/ovh/ovhcloud-cli/internal/display"
	filtersLib "github.com/ovh/ovhcloud-cli/internal/filters"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
	"github.com/ovh/ovhcloud-cli/internal/services/common"
	"github.com/spf13/cobra"
)

var (
	shareColumnsToDisplay = []string{
		"id",
		"currentState.name name",
		"currentState.location.region region",
		"currentState.protocol type",
		"currentState.size size",
		"resourceStatus status",
	}
	shareSnapshotColumnsToDisplay = []string{"id", "name", "shareId", "size", "status"}
	shareACLColumnsToDisplay      = []string{"id", "accessLevel", "accessTo", "accessType", "status"}

	//go:embed templates/cloud_storage_file_share.tmpl
	shareTemplate string

	//go:embed templates/cloud_storage_file_share_snapshot.tmpl
	shareSnapshotTemplate string

	//go:embed parameter-samples/storage-file-share-create.json
	ShareCreateExample string

	ShareSpec struct {
		TargetSpec struct {
			Description  string `json:"description,omitempty"`
			Name         string `json:"name,omitempty"`
			ShareNetwork struct {
				Id string `json:"id,omitempty"`
			} `json:"shareNetwork,omitzero"`
			Size      int    `json:"size,omitempty"`
			SubnetId  string `json:"subnetId,omitempty"`
			Protocol  string `json:"protocol,omitempty"`
			ShareType string `json:"shareType,omitempty"`
			Location  struct {
				Region string `json:"region,omitempty"`
			} `json:"location,omitzero"`
		} `json:"targetSpec"`
	}

	ShareEditSpec struct {
		TargetSpec struct {
			Description string `json:"description,omitempty"`
			Name        string `json:"name,omitempty"`
			Size        int    `json:"size,omitempty"`
		} `json:"targetSpec,omitzero"`
	}

	ShareSnapshotSpec struct {
		Description string `json:"description,omitempty"`
		Name        string `json:"name,omitempty"`
	}

	ShareACLSpec struct {
		AccessLevel string `json:"accessLevel,omitempty"`
		AccessTo    string `json:"accessTo,omitempty"`
	}

	ShareRegion string
)

func shareV2Endpoint(projectID string) string {
	return fmt.Sprintf("/v2/publicCloud/project/%s/storage/file/share", projectID)
}

// getShareRegions returns a single-element slice if --region is set,
// otherwise discovers all regions with the share feature available.
func getShareRegions(projectID string) ([]any, error) {
	if ShareRegion != "" {
		return []any{ShareRegion}, nil
	}
	return getCloudRegionsWithFeatureAvailable(projectID, "share")
}

// findShare searches for a share across all regions and returns its endpoint and data.
func findShare(shareID string) (string, map[string]any, error) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		return "", nil, err
	}

	regions, err := getShareRegions(projectID)
	if err != nil {
		return "", nil, fmt.Errorf("failed to fetch regions with share feature available: %w", err)
	}

	for _, region := range regions {
		var (
			share    map[string]any
			endpoint = fmt.Sprintf("/v1/cloud/project/%s/region/%s/share/%s",
				projectID, url.PathEscape(region.(string)), url.PathEscape(shareID))
		)
		if err := httpLib.Client.Get(endpoint, &share); err == nil {
			return endpoint, share, nil
		}
	}

	return "", nil, fmt.Errorf("no share found with ID %s", shareID)
}

func ListShares(_ *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.ManageListRequestNoExpand(shareV2Endpoint(projectID), shareColumnsToDisplay, flags.GenericFilters)
}

func GetShare(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.ManageObjectRequest(shareV2Endpoint(projectID), args[0], shareTemplate)
}

func CreateShare(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	ShareSpec.TargetSpec.Location.Region = args[0]
	endpoint := shareV2Endpoint(projectID)
	task, err := common.CreateResource(
		cmd,
		"/publicCloud/project/{projectId}/storage/file/share",
		endpoint,
		ShareCreateExample,
		ShareSpec,
		assets.CloudV2OpenapiSchema,
		[]string{"targetSpec"},
	)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, task, "✅ Share creation started successfully (operation ID: %s)", task["id"])
}

func EditShare(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("%s/%s", shareV2Endpoint(projectID), url.PathEscape(args[0]))
	if err := common.EditResource(
		cmd,
		"/publicCloud/project/{projectId}/storage/file/share/{fileStorageId}",
		endpoint,
		ShareEditSpec,
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
		display.OutputError(&flags.OutputFormatConfig, "failed to wait for share to be ready: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, ready, "✅ Share %s is now ready", args[0])
}

func DeleteShare(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("%s/%s", shareV2Endpoint(projectID), url.PathEscape(args[0]))
	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to delete share: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Share %s deleted successfully", args[0])
}

// ACL commands

func ListShareACLs(_ *cobra.Command, args []string) {
	endpoint, _, err := findShare(args[0])
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	var acls []map[string]any
	if err := httpLib.Client.Get(endpoint+"/acl", &acls); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch share ACLs: %s", err)
		return
	}

	acls, err = filtersLib.FilterLines(acls, flags.GenericFilters)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to filter results: %s", err)
		return
	}

	display.RenderTable(acls, shareACLColumnsToDisplay, &flags.OutputFormatConfig)
}

func GetShareACL(_ *cobra.Command, args []string) {
	endpoint, _, err := findShare(args[0])
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	var acl map[string]any
	if err := httpLib.Client.Get(fmt.Sprintf("%s/acl/%s", endpoint, url.PathEscape(args[1])), &acl); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch share ACL: %s", err)
		return
	}

	display.OutputObject(acl, args[1], "", &flags.OutputFormatConfig)
}

func CreateShareACL(_ *cobra.Command, args []string) {
	endpoint, _, err := findShare(args[0])
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	var response map[string]any
	if err := httpLib.Client.Post(endpoint+"/acl", ShareACLSpec, &response); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to create share ACL: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, response, "✅ ACL created successfully for share %s (id: %s)", args[0], response["id"])
}

func DeleteShareACL(_ *cobra.Command, args []string) {
	endpoint, _, err := findShare(args[0])
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	if err := httpLib.Client.Delete(fmt.Sprintf("%s/acl/%s", endpoint, url.PathEscape(args[1])), nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to delete share ACL: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ ACL %s deleted successfully from share %s", args[1], args[0])
}

// Snapshot commands

func ListShareSnapshots(_ *cobra.Command, args []string) {
	endpoint, _, err := findShare(args[0])
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	var snapshots []map[string]any
	if err := httpLib.Client.Get(endpoint+"/snapshot", &snapshots); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch share snapshots: %s", err)
		return
	}

	snapshots, err = filtersLib.FilterLines(snapshots, flags.GenericFilters)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to filter results: %s", err)
		return
	}

	display.RenderTable(snapshots, shareSnapshotColumnsToDisplay, &flags.OutputFormatConfig)
}

func GetShareSnapshot(_ *cobra.Command, args []string) {
	endpoint, _, err := findShare(args[0])
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	var snapshot map[string]any
	if err := httpLib.Client.Get(fmt.Sprintf("%s/snapshot/%s", endpoint, url.PathEscape(args[1])), &snapshot); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch share snapshot: %s", err)
		return
	}

	display.OutputObject(snapshot, args[1], shareSnapshotTemplate, &flags.OutputFormatConfig)
}

func CreateShareSnapshot(_ *cobra.Command, args []string) {
	endpoint, _, err := findShare(args[0])
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	var response map[string]any
	if err := httpLib.Client.Post(endpoint+"/snapshot", ShareSnapshotSpec, &response); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to create share snapshot: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, response, "✅ Snapshot created successfully for share %s (id: %s)", args[0], response["id"])
}

func DeleteShareSnapshot(_ *cobra.Command, args []string) {
	endpoint, _, err := findShare(args[0])
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	if err := httpLib.Client.Delete(fmt.Sprintf("%s/snapshot/%s", endpoint, url.PathEscape(args[1])), nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to delete share snapshot: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Snapshot %s deleted successfully from share %s", args[1], args[0])
}
