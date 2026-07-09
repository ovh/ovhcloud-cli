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
	fileStorageShareColumnsToDisplay = []string{
		"id",
		"currentState.name name",
		"currentState.location.region region",
		"currentState.protocol protocol",
		"currentState.size size",
		"resourceStatus",
	}
	fileStorageSnapshotColumnsToDisplay = []string{
		"id",
		"currentState.name name",
		"currentState.share.id shareId",
		"currentState.size size",
		"resourceStatus",
	}
	fileStorageNetworkColumnsToDisplay = []string{
		"id",
		"currentState.name name",
		"currentState.location.region region",
		"currentState.network.id networkId",
		"resourceStatus",
	}

	//go:embed templates/cloud_storage_file_share.tmpl
	fileStorageShareTemplate string

	//go:embed templates/cloud_storage_file_snapshot.tmpl
	fileStorageSnapshotTemplate string

	//go:embed templates/cloud_storage_file_network.tmpl
	fileStorageNetworkTemplate string

	//go:embed parameter-samples/storage-file-share-create.json
	FileStorageShareCreateExample string

	//go:embed parameter-samples/storage-file-snapshot-create.json
	FileStorageSnapshotCreateExample string

	//go:embed parameter-samples/storage-file-network-create.json
	FileStorageNetworkCreateExample string

	// Nested objects are built from these scalar flags before the request is sent.
	ShareLocationRegion           string
	ShareLocationAvailabilityZone string
	ShareNetworkID                string

	SnapshotShareID string

	NetworkLocationRegion           string
	NetworkLocationAvailabilityZone string
	NetworkNetworkID                string
	NetworkSubnetID                 string

	ShareSpec struct {
		TargetSpec struct {
			Name         string                  `json:"name,omitempty"`
			Description  string                  `json:"description,omitempty"`
			Protocol     string                  `json:"protocol,omitempty"`
			ShareType    string                  `json:"shareType,omitempty"`
			Size         int                     `json:"size,omitempty"`
			Location     *FileStorageLocation    `json:"location,omitempty"`
			ShareNetwork *FileStorageRef         `json:"shareNetwork,omitempty"`
			AccessRules  []FileStorageAccessRule `json:"accessRules,omitempty"`
		} `json:"targetSpec"`
	}

	ShareEditSpec struct {
		TargetSpec struct {
			Name        string `json:"name,omitempty"`
			Description string `json:"description,omitempty"`
			Size        int    `json:"size,omitempty"`
		} `json:"targetSpec"`
	}

	SnapshotSpec struct {
		TargetSpec struct {
			Name        string          `json:"name,omitempty"`
			Description string          `json:"description,omitempty"`
			Share       *FileStorageRef `json:"share,omitempty"`
		} `json:"targetSpec"`
	}

	SnapshotEditSpec struct {
		TargetSpec struct {
			Name        string `json:"name,omitempty"`
			Description string `json:"description,omitempty"`
		} `json:"targetSpec"`
	}

	NetworkSpec struct {
		TargetSpec struct {
			Name        string               `json:"name,omitempty"`
			Description string               `json:"description,omitempty"`
			Location    *FileStorageLocation `json:"location,omitempty"`
			Network     *FileStorageRef      `json:"network,omitempty"`
			Subnet      *FileStorageRef      `json:"subnet,omitempty"`
		} `json:"targetSpec"`
	}
)

type (
	FileStorageLocation struct {
		Region           string `json:"region,omitempty"`
		AvailabilityZone string `json:"availabilityZone,omitempty"`
	}

	FileStorageRef struct {
		ID string `json:"id,omitempty"`
	}

	FileStorageAccessRule struct {
		AccessLevel string `json:"accessLevel,omitempty"`
		AccessTo    string `json:"accessTo,omitempty"`
	}
)

// File storage shares

func ListShares(_ *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.ManageListRequestNoExpand(
		fmt.Sprintf("/v2/publicCloud/project/%s/storage/file/share", projectID),
		fileStorageShareColumnsToDisplay,
		flags.GenericFilters,
	)
}

func GetShare(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.ManageObjectRequest(
		fmt.Sprintf("/v2/publicCloud/project/%s/storage/file/share", projectID),
		args[0],
		fileStorageShareTemplate,
	)
}

func CreateShare(cmd *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	ShareSpec.TargetSpec.Location = nil
	if ShareLocationRegion != "" || ShareLocationAvailabilityZone != "" {
		ShareSpec.TargetSpec.Location = &FileStorageLocation{
			Region:           ShareLocationRegion,
			AvailabilityZone: ShareLocationAvailabilityZone,
		}
	}
	ShareSpec.TargetSpec.ShareNetwork = nil
	if ShareNetworkID != "" {
		ShareSpec.TargetSpec.ShareNetwork = &FileStorageRef{ID: ShareNetworkID}
	}

	endpoint := fmt.Sprintf("/v2/publicCloud/project/%s/storage/file/share", projectID)
	share, err := common.CreateResource(
		cmd,
		"/publicCloud/project/{projectId}/storage/file/share",
		endpoint,
		FileStorageShareCreateExample,
		ShareSpec,
		assets.CloudV2OpenapiSchema,
		[]string{"targetSpec"},
	)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to create file storage share: %s", err)
		return
	}

	shareID, _ := share["id"].(string)

	if !flags.WaitForTask {
		display.OutputInfo(&flags.OutputFormatConfig, share, "⚡️ File storage share creation started successfully (id: %s)", shareID)
		return
	}

	if err := waitForCloudResourceReady(fmt.Sprintf("%s/%s", endpoint, url.PathEscape(shareID)), 30*time.Minute); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to wait for file storage share to be ready: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, share, "✅ File storage share %s created successfully", shareID)
}

func EditShare(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	if err := common.EditResource(
		cmd,
		"/publicCloud/project/{projectId}/storage/file/share/{fileStorageId}",
		fmt.Sprintf("/v2/publicCloud/project/%s/storage/file/share/%s", projectID, url.PathEscape(args[0])),
		ShareEditSpec,
		assets.CloudV2OpenapiSchema,
	); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}
}

func DeleteShare(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v2/publicCloud/project/%s/storage/file/share/%s", projectID, url.PathEscape(args[0]))
	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to delete file storage share: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ File storage share %s is being deleted…", args[0])
}

// File storage snapshots

func ListFileStorageSnapshots(_ *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.ManageListRequestNoExpand(
		fmt.Sprintf("/v2/publicCloud/project/%s/storage/file/snapshot", projectID),
		fileStorageSnapshotColumnsToDisplay,
		flags.GenericFilters,
	)
}

func GetFileStorageSnapshot(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.ManageObjectRequest(
		fmt.Sprintf("/v2/publicCloud/project/%s/storage/file/snapshot", projectID),
		args[0],
		fileStorageSnapshotTemplate,
	)
}

func CreateFileStorageSnapshot(cmd *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	SnapshotSpec.TargetSpec.Share = nil
	if SnapshotShareID != "" {
		SnapshotSpec.TargetSpec.Share = &FileStorageRef{ID: SnapshotShareID}
	}

	endpoint := fmt.Sprintf("/v2/publicCloud/project/%s/storage/file/snapshot", projectID)
	snapshot, err := common.CreateResource(
		cmd,
		"/publicCloud/project/{projectId}/storage/file/snapshot",
		endpoint,
		FileStorageSnapshotCreateExample,
		SnapshotSpec,
		assets.CloudV2OpenapiSchema,
		[]string{"targetSpec"},
	)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to create file storage snapshot: %s", err)
		return
	}

	snapshotID, _ := snapshot["id"].(string)

	if !flags.WaitForTask {
		display.OutputInfo(&flags.OutputFormatConfig, snapshot, "⚡️ File storage snapshot creation started successfully (id: %s)", snapshotID)
		return
	}

	if err := waitForCloudResourceReady(fmt.Sprintf("%s/%s", endpoint, url.PathEscape(snapshotID)), 30*time.Minute); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to wait for file storage snapshot to be ready: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, snapshot, "✅ File storage snapshot %s created successfully", snapshotID)
}

func EditFileStorageSnapshot(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	if err := common.EditResource(
		cmd,
		"/publicCloud/project/{projectId}/storage/file/snapshot/{snapshotId}",
		fmt.Sprintf("/v2/publicCloud/project/%s/storage/file/snapshot/%s", projectID, url.PathEscape(args[0])),
		SnapshotEditSpec,
		assets.CloudV2OpenapiSchema,
	); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}
}

func DeleteFileStorageSnapshot(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v2/publicCloud/project/%s/storage/file/snapshot/%s", projectID, url.PathEscape(args[0]))
	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to delete file storage snapshot: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ File storage snapshot %s is being deleted…", args[0])
}

// File storage share networks

func ListFileStorageNetworks(_ *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.ManageListRequestNoExpand(
		fmt.Sprintf("/v2/publicCloud/project/%s/storage/file/network", projectID),
		fileStorageNetworkColumnsToDisplay,
		flags.GenericFilters,
	)
}

func GetFileStorageNetwork(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.ManageObjectRequest(
		fmt.Sprintf("/v2/publicCloud/project/%s/storage/file/network", projectID),
		args[0],
		fileStorageNetworkTemplate,
	)
}

func CreateFileStorageNetwork(cmd *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	NetworkSpec.TargetSpec.Location = nil
	if NetworkLocationRegion != "" || NetworkLocationAvailabilityZone != "" {
		NetworkSpec.TargetSpec.Location = &FileStorageLocation{
			Region:           NetworkLocationRegion,
			AvailabilityZone: NetworkLocationAvailabilityZone,
		}
	}
	NetworkSpec.TargetSpec.Network = nil
	if NetworkNetworkID != "" {
		NetworkSpec.TargetSpec.Network = &FileStorageRef{ID: NetworkNetworkID}
	}
	NetworkSpec.TargetSpec.Subnet = nil
	if NetworkSubnetID != "" {
		NetworkSpec.TargetSpec.Subnet = &FileStorageRef{ID: NetworkSubnetID}
	}

	endpoint := fmt.Sprintf("/v2/publicCloud/project/%s/storage/file/network", projectID)
	network, err := common.CreateResource(
		cmd,
		"/publicCloud/project/{projectId}/storage/file/network",
		endpoint,
		FileStorageNetworkCreateExample,
		NetworkSpec,
		assets.CloudV2OpenapiSchema,
		[]string{"targetSpec"},
	)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to create file storage share network: %s", err)
		return
	}

	networkID, _ := network["id"].(string)

	if !flags.WaitForTask {
		display.OutputInfo(&flags.OutputFormatConfig, network, "⚡️ File storage share network creation started successfully (id: %s)", networkID)
		return
	}

	if err := waitForCloudResourceReady(fmt.Sprintf("%s/%s", endpoint, url.PathEscape(networkID)), 30*time.Minute); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to wait for file storage share network to be ready: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, network, "✅ File storage share network %s created successfully", networkID)
}

func DeleteFileStorageNetwork(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v2/publicCloud/project/%s/storage/file/network/%s", projectID, url.PathEscape(args[0]))
	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to delete file storage share network: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ File storage share network %s is being deleted…", args[0])
}
