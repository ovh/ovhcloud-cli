// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cloud

import (
	_ "embed"
	"errors"
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
	volumeColumnsToDisplay = []string{
		"id",
		"currentState.name name",
		"currentState.location.region region",
		"currentState.volumeType type",
		"currentState.size size",
		"currentState.status status",
	}

	snapshotColumnsToDisplay = []string{
		"id",
		"currentState.name name",
		"currentState.location.region region",
		"currentState.size size",
		"currentState.volumeId volumeId",
		"resourceStatus status",
	}

	backupColumnsToDisplay = []string{
		"id",
		"currentState.name name",
		"currentState.location.region region",
		"currentState.size size",
		"currentState.volumeId volumeId",
		"resourceStatus status",
	}

	//go:embed templates/cloud_volume.tmpl
	volumeTemplate string

	//go:embed parameter-samples/volume-create.json
	VolumeCreateExample string

	VolumeSpec struct {
		TargetSpec struct {
			Name       string `json:"name,omitempty"`
			Size       int    `json:"size,omitempty"`
			VolumeType string `json:"volumeType,omitempty"`
			Location   struct {
				Region           string `json:"region,omitempty"`
				AvailabilityZone string `json:"availabilityZone,omitempty"`
			} `json:"location,omitzero"`
			CreateFrom struct {
				BackupId   string `json:"backupId,omitempty"`
				ImageId    string `json:"imageId,omitempty"`
				SnapshotId string `json:"snapshotId,omitempty"`
			} `json:"createFrom,omitzero"`
		} `json:"targetSpec"`
	}

	VolumeEditSpec struct {
		TargetSpec struct {
			Name       string `json:"name,omitempty"`
			Size       int    `json:"size,omitempty"`
			VolumeType string `json:"volumeType,omitempty"`
		} `json:"targetSpec,omitzero"`
	}

	VolumeSnapShotSpec struct {
		Description string `json:"description,omitempty"`
		Name        string `json:"name,omitempty"`
	}
)

// volumeV2Endpoint returns the project-scoped v2 block storage volume endpoint.
func volumeV2Endpoint(projectID string) string {
	return fmt.Sprintf("/v2/publicCloud/project/%s/storage/block/volume", projectID)
}

// getVolumeLocation fetches the location (region/availability zone) of a volume,
// required to create project-scoped snapshots and backups in the v2 API.
func getVolumeLocation(projectID, volumeID string) (map[string]any, error) {
	var volume struct {
		CurrentState struct {
			Location map[string]any `json:"location"`
		} `json:"currentState"`
	}
	endpoint := fmt.Sprintf("%s/%s", volumeV2Endpoint(projectID), url.PathEscape(volumeID))
	if err := httpLib.Client.Get(endpoint, &volume); err != nil {
		return nil, err
	}
	if len(volume.CurrentState.Location) == 0 {
		return nil, fmt.Errorf("could not determine location of volume %s", volumeID)
	}
	return volume.CurrentState.Location, nil
}

func ListCloudVolumes(_ *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.ManageListRequestNoExpand(volumeV2Endpoint(projectID), volumeColumnsToDisplay, flags.GenericFilters)
}

func GetVolume(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.ManageObjectRequest(volumeV2Endpoint(projectID), args[0], volumeTemplate)
}

func EditVolume(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	// EditResource fetches the current volume, merges the given parameters and
	// keeps only the fields writable per the v2 PUT schema. The checksum
	// (optimistic locking) is read from the fetched volume and preserved
	// automatically, so it does not need to be handled here.
	endpoint := fmt.Sprintf("%s/%s", volumeV2Endpoint(projectID), url.PathEscape(args[0]))
	if err := common.EditResource(
		cmd,
		"/publicCloud/project/{projectId}/storage/block/volume/{id}",
		endpoint,
		VolumeEditSpec,
		assets.CloudV2OpenapiSchema,
	); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	if !flags.WaitForTask {
		return
	}

	// The volume ID is known from args[0], so we can poll the same endpoint
	// until the update has been applied and the volume is READY again.
	ready, err := waitForCloudResourceReady(endpoint, 10*time.Minute)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to wait for volume update: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, ready, "✅ Volume %s updated successfully", args[0])
}

func CreateVolume(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	// args[0] is the region; the v2 API expects it in targetSpec.location.
	VolumeSpec.TargetSpec.Location.Region = args[0]

	endpoint := volumeV2Endpoint(projectID)
	response, err := common.CreateResource(
		cmd,
		"/publicCloud/project/{projectId}/storage/block/volume",
		endpoint,
		VolumeCreateExample,
		VolumeSpec,
		assets.CloudV2OpenapiSchema,
		[]string{"targetSpec"},
	)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to create volume: %s", err)
		return
	}

	volumeID, _ := response["id"].(string)

	if !flags.WaitForTask {
		display.OutputInfo(&flags.OutputFormatConfig, response, "⚡️ Volume %s creation started", volumeID)
		return
	}

	ready, err := waitForCloudResourceReady(fmt.Sprintf("%s/%s", volumeV2Endpoint(projectID), url.PathEscape(volumeID)), 10*time.Minute)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to wait for volume creation: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, ready, "✅ Volume %s created successfully", volumeID)
}

func DeleteVolume(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	if !display.Confirm(fmt.Sprintf("Delete volume %s?", args[0]), flags.AssumeYes) {
		display.OutputInfo(&flags.OutputFormatConfig, nil, "Aborted: volume %s was not deleted", args[0])
		return
	}

	endpoint := fmt.Sprintf("%s/%s", volumeV2Endpoint(projectID), url.PathEscape(args[0]))
	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to delete volume: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Volume %s deleted successfully", args[0])
}

// AttachVolumeToInstance still uses the v1 API: the v2 block storage API does not
// expose a dedicated attach endpoint yet.
func AttachVolumeToInstance(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	if err := httpLib.Client.Post(
		fmt.Sprintf("/v1/cloud/project/%s/volume/%s/attach", projectID, url.PathEscape(args[0])),
		map[string]string{"instanceId": args[1]},
		nil,
	); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to attach volume: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Volume %s attached to instance %s successfully", args[0], args[1])
}

// DetachVolumeFromInstance still uses the v1 API: the v2 block storage API does
// not expose a dedicated detach endpoint yet.
func DetachVolumeFromInstance(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	if err := httpLib.Client.Post(
		fmt.Sprintf("/v1/cloud/project/%s/volume/%s/detach", projectID, url.PathEscape(args[0])),
		map[string]string{"instanceId": args[1]},
		nil,
	); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to detach volume: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Volume %s detached from instance %s successfully", args[0], args[1])
}

func CreateVolumeSnapshot(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	location, err := getVolumeLocation(projectID, args[0])
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch volume: %s", err)
		return
	}

	targetSpec := map[string]any{
		"name":     VolumeSnapShotSpec.Name,
		"volumeId": args[0],
		"location": location,
	}
	if VolumeSnapShotSpec.Description != "" {
		targetSpec["description"] = VolumeSnapShotSpec.Description
	}

	var response map[string]any
	endpoint := fmt.Sprintf("/v2/publicCloud/project/%s/storage/block/snapshot", projectID)
	if err := httpLib.Client.Post(endpoint, map[string]any{"targetSpec": targetSpec}, &response); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to create snapshot: %s", err)
		return
	}

	snapshotID, _ := response["id"].(string)

	if !flags.WaitForTask {
		display.OutputInfo(&flags.OutputFormatConfig, response, "⚡️ Snapshot %s creation started for volume %s", snapshotID, args[0])
		return
	}

	ready, err := waitForCloudResourceReady(fmt.Sprintf("%s/%s", endpoint, url.PathEscape(snapshotID)), 10*time.Minute)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to wait for snapshot creation: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, ready, "✅ Snapshot %s created successfully for volume %s", snapshotID, args[0])
}

func ListVolumeSnapshots(cmd *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	volume, err := cmd.Flags().GetString("volume-id")
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}
	if volume != "" {
		flags.GenericFilters = append(flags.GenericFilters, fmt.Sprintf("currentState.volumeId==%q", volume))
	}

	endpoint := fmt.Sprintf("/v2/publicCloud/project/%s/storage/block/snapshot", projectID)
	common.ManageListRequestNoExpand(endpoint, snapshotColumnsToDisplay, flags.GenericFilters)
}

func DeleteVolumeSnapshot(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v2/publicCloud/project/%s/storage/block/snapshot/%s", projectID, url.PathEscape(args[0]))
	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to delete snapshot: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Snapshot %s deleted successfully", args[0])
}

func ListVolumeBackups(_ *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v2/publicCloud/project/%s/storage/block/backup", projectID)
	common.ManageListRequestNoExpand(endpoint, backupColumnsToDisplay, flags.GenericFilters)
}

func GetVolumeBackup(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.ManageObjectRequest(fmt.Sprintf("/v2/publicCloud/project/%s/storage/block/backup", projectID), args[0], "")
}

func DeleteVolumeBackup(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v2/publicCloud/project/%s/storage/block/backup/%s", projectID, url.PathEscape(args[0]))
	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to delete volume backup: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Volume backup %s deleted successfully", args[0])
}

func CreateVolumeBackup(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	location, err := getVolumeLocation(projectID, args[0])
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch volume: %s", err)
		return
	}

	targetSpec := map[string]any{
		"name":     args[1],
		"volumeId": args[0],
		"location": location,
	}

	var response map[string]any
	endpoint := fmt.Sprintf("/v2/publicCloud/project/%s/storage/block/backup", projectID)
	if err := httpLib.Client.Post(endpoint, map[string]any{"targetSpec": targetSpec}, &response); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to create volume backup: %s", err)
		return
	}

	backupID, _ := response["id"].(string)

	if !flags.WaitForTask {
		display.OutputInfo(&flags.OutputFormatConfig, response, "⚡️ Backup %s creation started for volume %s", backupID, args[0])
		return
	}

	ready, err := waitForCloudResourceReady(fmt.Sprintf("%s/%s", endpoint, url.PathEscape(backupID)), 30*time.Minute)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to wait for backup creation: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, ready, "✅ Volume backup %s created successfully for volume %s", backupID, args[0])
}

// findVolumeBackupV1 locates a volume backup across all regions using the v1 API.
// It is still used by operations that have no v2 equivalent yet (restore and
// create-volume-from-backup).
func findVolumeBackupV1(backupId string) (string, error) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		return "", err
	}

	regions, err := getCloudRegionsWithFeatureAvailable(projectID, "volume")
	if err != nil {
		return "", fmt.Errorf("failed to fetch regions with volume feature available: %s", err)
	}

	for _, region := range regions {
		var (
			volumeBackup map[string]any
			endpoint     = fmt.Sprintf("/v1/cloud/project/%s/region/%s/volumeBackup/%s",
				projectID, url.PathEscape(region.(string)), url.PathEscape(backupId))
		)
		if err := httpLib.Client.Get(endpoint, &volumeBackup); err == nil {
			return endpoint, nil
		}
	}

	return "", errors.New("no volume backup found with given ID")
}

// RestoreVolumeBackup still uses the v1 API: the v2 block storage API does not
// expose a restore endpoint yet.
func RestoreVolumeBackup(_ *cobra.Command, args []string) {
	endpoint, err := findVolumeBackupV1(args[0])
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	if err := httpLib.Client.Post(
		endpoint+"/restore",
		map[string]string{"volumeId": args[1]},
		nil,
	); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to restore volume backup: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Volume backup %s is being restored to volume %s", args[0], args[1])
}
