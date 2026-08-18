// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cloud

import (
	_ "embed"
	"fmt"
	"net/url"

	"github.com/ovh/ovhcloud-cli/internal/assets"
	"github.com/ovh/ovhcloud-cli/internal/display"
	filtersLib "github.com/ovh/ovhcloud-cli/internal/filters"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
	"github.com/ovh/ovhcloud-cli/internal/services/common"
	"github.com/spf13/cobra"
)

var (
	shareColumnsToDisplay         = []string{"id", "name", "region", "protocol", "size", "status"}
	shareSnapshotColumnsToDisplay = []string{"id", "name", "shareId", "size", "status"}
	shareACLColumnsToDisplay      = []string{"id", "accessLevel", "accessTo", "accessType", "status"}

	//go:embed templates/cloud_storage_file_share.tmpl
	shareTemplate string

	//go:embed templates/cloud_storage_file_share_snapshot.tmpl
	shareSnapshotTemplate string

	//go:embed parameter-samples/storage-file-share-create.json
	ShareCreateExample string

	ShareSpec struct {
		AvailabilityZone string `json:"availabilityZone,omitempty"`
		Description      string `json:"description,omitempty"`
		Name             string `json:"name,omitempty"`
		NetworkId        string `json:"networkId,omitempty"`
		Size             int    `json:"size,omitempty"`
		SnapshotId       string `json:"snapshotId,omitempty"`
		SubnetId         string `json:"subnetId,omitempty"`
		Type             string `json:"type,omitempty"`
	}

	ShareEditSpec struct {
		Description string `json:"description,omitempty"`
		Name        string `json:"name,omitempty"`
		NewSize     int    `json:"newSize,omitempty"`
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

	regions, err := getShareRegions(projectID)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch regions with share feature available: %s", err)
		return
	}

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/region", projectID)
	shares, err := httpLib.FetchObjectsParallel[[]map[string]any](endpoint+"/%s/share", regions, true)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch shares: %s", err)
		return
	}

	var allShares []map[string]any
	for _, regionShares := range shares {
		allShares = append(allShares, regionShares...)
	}

	allShares, err = filtersLib.FilterLines(allShares, flags.GenericFilters)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to filter results: %s", err)
		return
	}

	display.RenderTable(allShares, shareColumnsToDisplay, &flags.OutputFormatConfig)
}

func GetShare(_ *cobra.Command, args []string) {
	_, share, err := findShare(args[0])
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	display.OutputObject(share, args[0], shareTemplate, &flags.OutputFormatConfig)
}

func CreateShare(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/region/%s/share", projectID, url.PathEscape(args[0]))
	task, err := common.CreateResource(
		cmd,
		"/cloud/project/{serviceName}/region/{regionName}/share",
		endpoint,
		ShareCreateExample,
		ShareSpec,
		assets.CloudOpenapiSchema,
		[]string{"type"},
	)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, task, "✅ Share creation started successfully (operation ID: %s)", task["id"])
}

func EditShare(cmd *cobra.Command, args []string) {
	endpoint, _, err := findShare(args[0])
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	if err := common.EditResource(
		cmd,
		"/cloud/project/{serviceName}/region/{regionName}/share/{shareId}",
		endpoint,
		&ShareEditSpec,
		assets.CloudOpenapiSchema,
	); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}
}

func DeleteShare(_ *cobra.Command, args []string) {
	endpoint, _, err := findShare(args[0])
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	var task map[string]any
	if err := httpLib.Client.Delete(endpoint, &task); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to delete share: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, task, "✅ Share %s deletion started successfully (operation ID: %s)", args[0], task["id"])

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
