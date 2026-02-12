// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cloud

import (
	_ "embed"
	"errors"
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
	shareColumnsToDisplay = []string{"id", "name", "region", "protocol", "type", "status", "size"}

	//go:embed templates/cloud_share.tmpl
	shareTemplate string

	//go:embed templates/cloud_share_acl.tmpl
	shareAclTemplate string

	//go:embed templates/cloud_share_snapshot.tmpl
	shareSnapshotTemplate string

	//go:embed parameter-samples/share-create.json
	ShareCreateExample string

	//go:embed parameter-samples/share-acl-create.json
	ShareAclCreateExample string

	//go:embed parameter-samples/share-snapshot-create.json
	ShareSnapshotCreateExample string

	ShareSpec struct {
		Description string `json:"description,omitempty"`
		Name        string `json:"name,omitempty"`
		NewSize     int    `json:"newSize,omitempty"`
		Size        int    `json:"size,omitempty"`
		Type        string `json:"type,omitempty"`
	}

	ShareAclSpec struct {
		AccessLevel string `json:"accessLevel,omitempty"`
		AccessTo    string `json:"accessTo,omitempty"`
		AccessType  string `json:"accessType,omitempty"`
	}

	ShareSnapshotSpec struct {
		Description string `json:"description,omitempty"`
		Name        string `json:"name,omitempty"`
	}
)

// Share CRUD operations

func ListCloudShares(_ *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	// Fetch regions with share feature available
	regions, err := getCloudRegionsWithFeatureAvailable(projectID, "share")
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch regions with share feature available: %s", err)
		return
	}

	// Fetch shares in all regions
	endpoint := fmt.Sprintf("/v1/cloud/project/%s/region", projectID)
	shares, err := httpLib.FetchObjectsParallel[[]map[string]any](endpoint+"/%s/share", regions, true)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch shares: %s", err)
		return
	}

	// Flatten shares in a single array
	var allShares []map[string]any
	for _, regionShares := range shares {
		allShares = append(allShares, regionShares...)
	}

	// Filter results
	allShares, err = filtersLib.FilterLines(allShares, flags.GenericFilters)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to filter results: %s", err)
		return
	}

	display.RenderTable(allShares, shareColumnsToDisplay, &flags.OutputFormatConfig)
}

func findShare(shareId string) (string, map[string]any, error) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		return "", nil, err
	}

	// Fetch regions with share feature available
	regions, err := getCloudRegionsWithFeatureAvailable(projectID, "share")
	if err != nil {
		return "", nil, fmt.Errorf("failed to fetch regions with share feature available: %s", err)
	}

	// Search for the given share in all regions
	for _, region := range regions {
		var (
			share    map[string]any
			endpoint = fmt.Sprintf("/v1/cloud/project/%s/region/%s/share/%s",
				projectID, url.PathEscape(region.(string)), url.PathEscape(shareId))
		)
		if err := httpLib.Client.Get(endpoint, &share); err == nil {
			return endpoint, share, nil
		}
	}

	return "", nil, errors.New("no share found with given ID")
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
		[]string{"name", "size", "type"},
	)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, task, "✅ Share %s created successfully", task["id"])
}

func EditShare(cmd *cobra.Command, args []string) {
	endpoint, _, err := findShare(args[0])
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	if err := common.EditResource(
		cmd,
		"/cloud/project/{serviceName}/region/{regionName}/share/{id}",
		endpoint,
		ShareSpec,
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

	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to delete share: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Share %s deleted successfully", args[0])
}

// Share ACL operations

func ListShareAcls(_ *cobra.Command, args []string) {
	endpoint, _, err := findShare(args[0])
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.ManageListRequestNoExpand(endpoint+"/acl", []string{"id", "accessType", "accessTo", "accessLevel", "status"}, flags.GenericFilters)
}

func GetShareAcl(_ *cobra.Command, args []string) {
	endpoint, _, err := findShare(args[0])
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.ManageObjectRequest(endpoint+"/acl", args[1], shareAclTemplate)
}

func CreateShareAcl(cmd *cobra.Command, args []string) {
	endpoint, _, err := findShare(args[0])
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	acl, err := common.CreateResource(
		cmd,
		"/cloud/project/{serviceName}/region/{regionName}/share/{id}/acl",
		endpoint+"/acl",
		ShareAclCreateExample,
		ShareAclSpec,
		assets.CloudOpenapiSchema,
		[]string{"accessType", "accessTo", "accessLevel"},
	)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, acl, "✅ Share ACL %s created successfully", acl["id"])
}

func DeleteShareAcl(_ *cobra.Command, args []string) {
	endpoint, _, err := findShare(args[0])
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	aclEndpoint := fmt.Sprintf("%s/acl/%s", endpoint, url.PathEscape(args[1]))
	if err := httpLib.Client.Delete(aclEndpoint, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to delete share ACL: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Share ACL %s deleted successfully", args[1])
}

// Share Snapshot operations

func ListShareSnapshots(_ *cobra.Command, args []string) {
	endpoint, _, err := findShare(args[0])
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.ManageListRequestNoExpand(endpoint+"/snapshot", []string{"id", "name", "shareId", "status", "size"}, flags.GenericFilters)
}

func GetShareSnapshot(_ *cobra.Command, args []string) {
	endpoint, _, err := findShare(args[0])
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.ManageObjectRequest(endpoint+"/snapshot", args[1], shareSnapshotTemplate)
}

func CreateShareSnapshot(cmd *cobra.Command, args []string) {
	endpoint, _, err := findShare(args[0])
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	snapshot, err := common.CreateResource(
		cmd,
		"/cloud/project/{serviceName}/region/{regionName}/share/{id}/snapshot",
		endpoint+"/snapshot",
		ShareSnapshotCreateExample,
		ShareSnapshotSpec,
		assets.CloudOpenapiSchema,
		[]string{},
	)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, snapshot, "✅ Share snapshot %s created successfully", snapshot["id"])
}

func DeleteShareSnapshot(_ *cobra.Command, args []string) {
	endpoint, _, err := findShare(args[0])
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	snapshotEndpoint := fmt.Sprintf("%s/snapshot/%s", endpoint, url.PathEscape(args[1]))
	if err := httpLib.Client.Delete(snapshotEndpoint, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to delete share snapshot: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Share snapshot %s deleted successfully", args[1])
}
