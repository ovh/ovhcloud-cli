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
	"github.com/ovh/ovhcloud-cli/internal/flags"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
	"github.com/ovh/ovhcloud-cli/internal/services/common"
	"github.com/spf13/cobra"
)

var (
	shareColumnsToDisplay = []string{
		"id",
		"targetSpec.name name",
		"targetSpec.location.region region",
		"targetSpec.protocol type",
		"targetSpec.size size",
		"resourceStatus status",
	}
	shareNetworkColumnsToDisplay = []string{
		"id",
		"targetSpec.name name",
		"targetSpec.location.region region",
		"targetSpec.network.id networkId",
		"targetSpec.subnet.id subnetId",
		"resourceStatus status",
	}
	shareSnapshotColumnsToDisplay = []string{
		"id",
		"targetSpec.name name",
		"targetSpec.share.id shareId",
		"resourceStatus status",
	}
	shareACLColumnsToDisplay = []string{
		"id",
		"currentState.accessLevel accessLevel",
		"currentState.accessTo accessTo",
		"resourceStatus status",
	}

	//go:embed templates/cloud_storage_file_share.tmpl
	shareTemplate string

	//go:embed templates/cloud_storage_file_share_snapshot.tmpl
	shareSnapshotTemplate string

	//go:embed parameter-samples/storage-file-share-create.json
	ShareCreateExample string

	//go:embed parameter-samples/storage-file-share-network-create.json
	ShareNetworkCreateExample string

	//go:embed parameter-samples/storage-file-share-snapshot-create.json
	ShareSnapshotCreateExample string

	ShareSpec struct {
		TargetSpec struct {
			Description  string `json:"description,omitempty"`
			Name         string `json:"name,omitempty"`
			ShareNetwork struct {
				Id string `json:"id,omitempty"`
			} `json:"shareNetwork,omitzero"`
			Size      int    `json:"size,omitempty"`
			Protocol  string `json:"protocol,omitempty"`
			ShareType string `json:"shareType,omitempty"`
			Location  struct {
				AvailabilityZone string `json:"availabilityZone,omitempty"`
				Region           string `json:"region,omitempty"`
			} `json:"location,omitzero"`
		} `json:"targetSpec"`
	}

	ShareNetworkSpec struct {
		TargetSpec struct {
			Description string `json:"description,omitempty"`
			Name        string `json:"name,omitempty"`
			Location    struct {
				AvailabilityZone string `json:"availabilityZone,omitempty"`
				Region           string `json:"region,omitempty"`
			} `json:"location,omitzero"`
			Network struct {
				Id string `json:"id,omitempty"`
			} `json:"network,omitzero"`
			Subnet struct {
				Id string `json:"id,omitempty"`
			} `json:"subnet,omitzero"`
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
		TargetSpec struct {
			Description string `json:"description,omitempty"`
			Name        string `json:"name,omitempty"`
			Share       struct {
				Id string `json:"id,omitempty"`
			} `json:"share,omitzero"`
		} `json:"targetSpec"`
	}

	ShareSnapshotEditSpec struct {
		TargetSpec struct {
			Description string `json:"description,omitempty"`
			Name        string `json:"name,omitempty"`
		} `json:"targetSpec,omitzero"`
	}

	ShareACLSpec struct {
		TargetSpec struct {
			AccessLevel string `json:"accessLevel,omitempty"`
			AccessTo    string `json:"accessTo,omitempty"`
		} `json:"targetSpec"`
	}
)

func shareV2Endpoint(projectID string) string {
	return fmt.Sprintf("/v2/publicCloud/project/%s/storage/file/share", projectID)
}

func shareNetworkV2Endpoint(projectID string) string {
	return fmt.Sprintf("/v2/publicCloud/project/%s/storage/file/network", projectID)
}

func shareSnapshotV2Endpoint(projectID string) string {
	return fmt.Sprintf("/v2/publicCloud/project/%s/storage/file/snapshot", projectID)
}

func ListShares(_ *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.ManageListRequestNoExpand(shareV2Endpoint(projectID), shareColumnsToDisplay, flags.GenericFilters)
}

func ListShareNetworks(_ *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.ManageListRequestNoExpand(shareNetworkV2Endpoint(projectID), shareNetworkColumnsToDisplay, flags.GenericFilters)
}

func GetShareNetwork(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.ManageObjectRequest(shareNetworkV2Endpoint(projectID), args[0], "")
}

func CreateShareNetwork(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	ShareNetworkSpec.TargetSpec.Location.Region = args[0]
	resource, err := common.CreateResource(
		cmd,
		"/publicCloud/project/{projectId}/storage/file/network",
		shareNetworkV2Endpoint(projectID),
		ShareNetworkCreateExample,
		ShareNetworkSpec,
		assets.CloudV2OpenapiSchema,
		[]string{"targetSpec.subnet.id", "targetSpec.network.id", "targetSpec.location.region", "targetSpec.name"},
	)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, resource, "✅ Share network creation started successfully (id: %s)", resource["id"])
}

func DeleteShareNetwork(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("%s/%s", shareNetworkV2Endpoint(projectID), url.PathEscape(args[0]))
	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to delete share network: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Share network %s is being deleted", args[0])
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
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("%s/%s/acl", shareV2Endpoint(projectID), url.PathEscape(args[0]))
	common.ManageListRequestNoExpand(endpoint, shareACLColumnsToDisplay, flags.GenericFilters)
}

func GetShareACL(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("%s/%s/acl/%s", shareV2Endpoint(projectID), url.PathEscape(args[0]), url.PathEscape(args[1]))
	var acl map[string]any
	if err := httpLib.Client.Get(endpoint, &acl); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch share ACL: %s", err)
		return
	}

	display.OutputObject(acl, args[1], "", &flags.OutputFormatConfig)
}

func CreateShareACL(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("%s/%s/acl", shareV2Endpoint(projectID), url.PathEscape(args[0]))
	var response map[string]any
	if err := httpLib.Client.Post(endpoint, ShareACLSpec, &response); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to create share ACL: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, response, "✅ ACL created successfully for share %s (id: %s)", args[0], response["id"])
}

func DeleteShareACL(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("%s/%s/acl/%s", shareV2Endpoint(projectID), url.PathEscape(args[0]), url.PathEscape(args[1]))
	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to delete share ACL: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ ACL %s deleted successfully from share %s", args[1], args[0])
}

// Snapshot commands

// shareSnapshotParentID returns the identifier of the share a snapshot was
// taken from. The parent share is set at creation and never changes, so
// targetSpec is authoritative and is also the only one available while the
// snapshot is still being created.
func shareSnapshotParentID(snapshot map[string]any) string {
	for _, state := range []string{"targetSpec", "currentState"} {
		stateValue, ok := snapshot[state].(map[string]any)
		if !ok {
			continue
		}
		share, ok := stateValue["share"].(map[string]any)
		if !ok {
			continue
		}
		if id, ok := share["id"].(string); ok && id != "" {
			return id
		}
	}

	return ""
}

// fetchShareSnapshot fetches a snapshot and checks that it really belongs to
// the given share. Snapshot routes are project scoped in the v2 API, so
// without this check any snapshot of the project would be reachable (and
// deletable) through any share ID.
func fetchShareSnapshot(projectID, shareID, snapshotID string) (map[string]any, error) {
	var snapshot map[string]any
	endpoint := fmt.Sprintf("%s/%s", shareSnapshotV2Endpoint(projectID), url.PathEscape(snapshotID))
	if err := httpLib.Client.Get(endpoint, &snapshot); err != nil {
		return nil, fmt.Errorf("failed to fetch share snapshot: %w", err)
	}

	switch parent := shareSnapshotParentID(snapshot); parent {
	case shareID:
		return snapshot, nil
	case "":
		return nil, fmt.Errorf("failed to determine the parent share of snapshot %s", snapshotID)
	default:
		return nil, fmt.Errorf("snapshot %s belongs to share %s, not to share %s", snapshotID, parent, shareID)
	}
}

func ListShareSnapshots(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	filters := append([]string{}, flags.GenericFilters...)
	filters = append(filters, fmt.Sprintf("targetSpec.share.id==%q", args[0]))
	common.ManageListRequestNoExpand(shareSnapshotV2Endpoint(projectID), shareSnapshotColumnsToDisplay, filters)
}

func GetShareSnapshot(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	snapshot, err := fetchShareSnapshot(projectID, args[0], args[1])
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	display.OutputObject(snapshot, args[1], shareSnapshotTemplate, &flags.OutputFormatConfig)
}

func CreateShareSnapshot(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	ShareSnapshotSpec.TargetSpec.Share.Id = args[0]
	endpoint := shareSnapshotV2Endpoint(projectID)
	snapshot, err := common.CreateResource(
		cmd,
		"/publicCloud/project/{projectId}/storage/file/snapshot",
		endpoint,
		ShareSnapshotCreateExample,
		ShareSnapshotSpec,
		assets.CloudV2OpenapiSchema,
		[]string{"targetSpec.share.id"},
	)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	snapshotID, _ := snapshot["id"].(string)

	if !flags.WaitForTask {
		display.OutputInfo(&flags.OutputFormatConfig, snapshot, "✅ Snapshot creation started successfully for share %s (id: %s)", args[0], snapshotID)
		return
	}

	ready, err := waitForCloudResourceReady(fmt.Sprintf("%s/%s", endpoint, url.PathEscape(snapshotID)), 10*time.Minute)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to wait for snapshot creation: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, ready, "✅ Snapshot %s created successfully for share %s", snapshotID, args[0])
}

func EditShareSnapshot(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	if _, err := fetchShareSnapshot(projectID, args[0], args[1]); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("%s/%s", shareSnapshotV2Endpoint(projectID), url.PathEscape(args[1]))
	if err := common.EditResource(
		cmd,
		"/publicCloud/project/{projectId}/storage/file/snapshot/{snapshotId}",
		endpoint,
		ShareSnapshotEditSpec,
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
		display.OutputError(&flags.OutputFormatConfig, "failed to wait for snapshot to be ready: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, ready, "✅ Snapshot %s is now ready", args[1])
}

func DeleteShareSnapshot(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	if _, err := fetchShareSnapshot(projectID, args[0], args[1]); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("%s/%s", shareSnapshotV2Endpoint(projectID), url.PathEscape(args[1]))
	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to delete share snapshot: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Snapshot %s of share %s is being deleted", args[1], args[0])
}
