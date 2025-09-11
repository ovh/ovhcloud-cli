// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cloud

import (
	_ "embed"
	"errors"
	"fmt"
	"net/url"
	"strconv"
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
	volumeColumnsToDisplay = []string{"id", "name", "region", "type", "status"}

	//go:embed templates/cloud_volume.tmpl
	volumeTemplate string

	//go:embed parameter-samples/volume-create.json
	VolumeCreateExample string

	VolumeSpec struct {
		AvailabilityZone string `json:"availabilityZone,omitempty"`
		BackupId         string `json:"backupId,omitempty"`
		Description      string `json:"description,omitempty"`
		ImageId          string `json:"imageId,omitempty"`
		InstanceId       string `json:"instanceId,omitempty"`
		Name             string `json:"name,omitempty"`
		Size             int    `json:"size,omitempty"`
		SnapshotId       string `json:"snapshotId,omitempty"`
		Type             string `json:"type,omitempty"`
	}

	VolumeSnapShotSpec struct {
		Description string `json:"description,omitempty"`
		Name        string `json:"name,omitempty"`
	}
)

func ListCloudVolumes(_ *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	common.ManageListRequestNoExpand(fmt.Sprintf("/cloud/project/%s/volume", projectID), volumeColumnsToDisplay, flags.GenericFilters)
}

func GetVolume(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	common.ManageObjectRequest(fmt.Sprintf("/cloud/project/%s/volume", projectID), args[0], volumeTemplate)
}

func EditVolume(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	if err := common.EditResource(
		cmd,
		"/cloud/project/{serviceName}/volume/{volumeId}",
		fmt.Sprintf("/cloud/project/%s/volume/%s", projectID, url.PathEscape(args[0])),
		VolumeSpec,
		assets.CloudOpenapiSchema,
	); err != nil {
		display.ExitError(err.Error())
		return
	}
}

func CreateVolume(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	endpoint := fmt.Sprintf("/cloud/project/%s/region/%s/volume", projectID, url.PathEscape(args[0]))
	task, err := common.CreateResource(
		cmd,
		"/cloud/project/{serviceName}/region/{regionName}/volume",
		endpoint,
		VolumeCreateExample,
		VolumeSpec,
		assets.CloudOpenapiSchema,
		[]string{"name", "size", "type"},
	)
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	if !flags.WaitForTask {
		fmt.Printf("\n⚡️ Volume creation started successfully (operation ID: %s)\n", task["id"])
		fmt.Printf("You can check the status of the operation with: `ovhcloud cloud operation get %s`\n", task["id"])
		return
	}

	volumeID, err := waitForCloudOperation(projectID, task["id"].(string), "ablockstorage.CreateVolume", 10*time.Minute)
	if err != nil {
		display.ExitError("failed to wait for volume creation: %s", err)
		return
	}

	fmt.Printf("✅ Volume %s created successfully\n", volumeID)
}

func DeleteVolume(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	endpoint := fmt.Sprintf("/cloud/project/%s/volume/%s", projectID, url.PathEscape(args[0]))
	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.ExitError("failed to delete volume: %s", err)
		return
	}

	fmt.Printf("✅ Volume %s deleted successfully\n", args[0])
}

func AttachVolumeToInstance(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	if err := httpLib.Client.Post(
		fmt.Sprintf("/cloud/project/%s/volume/%s/attach", projectID, url.PathEscape(args[0])),
		map[string]string{"instanceId": args[1]},
		nil,
	); err != nil {
		display.ExitError("failed to attach volume: %s", err)
		return
	}

	fmt.Printf("✅ Volume %s attached to instance %s successfully\n", args[0], args[1])
}

func DetachVolumeFromInstance(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	if err := httpLib.Client.Post(
		fmt.Sprintf("/cloud/project/%s/volume/%s/detach", projectID, url.PathEscape(args[0])),
		map[string]string{"instanceId": args[1]},
		nil,
	); err != nil {
		display.ExitError("failed to detach volume: %s", err)
		return
	}

	fmt.Printf("✅ Volume %s detached from instance %s successfully\n", args[0], args[1])
}

func CreateVolumeSnapshot(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	var (
		endpoint = fmt.Sprintf("/cloud/project/%s/volume/%s/snapshot", projectID, url.PathEscape(args[0]))
		response map[string]any
	)

	if err := httpLib.Client.Post(endpoint, VolumeSnapShotSpec, &response); err != nil {
		display.ExitError("failed to create snapshot: %s", err)
		return
	}

	fmt.Printf("✅ Snapshot for volume %s created successfully, id : %s\n", args[0], response["id"])
}

func ListVolumeSnapshots(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	endpoint := fmt.Sprintf("/cloud/project/%s/volume/snapshot", projectID)

	volume, err := cmd.Flags().GetString("volume-id")
	if err != nil {
		display.ExitError(err.Error())
		return
	}
	if volume != "" {
		flags.GenericFilters = append(flags.GenericFilters, fmt.Sprintf("volumeId==%q", volume))
	}

	common.ManageListRequestNoExpand(endpoint, []string{"id", "name", "region", "description", "status"}, flags.GenericFilters)
}

func DeleteVolumeSnapshot(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	endpoint := fmt.Sprintf("/cloud/project/%s/volume/snapshot/%s", projectID, url.PathEscape(args[0]))
	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.ExitError("failed to delete snapshot: %s", err)
		return
	}

	fmt.Printf("✅ Snapshot %s deleted successfully\n", args[0])
}

func UpsizeVolume(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	size, err := strconv.Atoi(args[1])
	if err != nil {
		display.ExitError(err.Error())
		return
	}
	if size <= 0 {
		display.ExitError("size must be a positive integer")
		return
	}

	endpoint := fmt.Sprintf("/cloud/project/%s/volume/%s/upsize", projectID, url.PathEscape(args[0]))
	if err := httpLib.Client.Post(
		endpoint,
		map[string]int{"size": size},
		nil,
	); err != nil {
		display.ExitError("failed to upsize volume: %s", err)
		return
	}

	fmt.Printf("✅ Volume %s upsized successfully to %dGB\n", args[0], size)
}

func findVolumeBackup(backupId string) (string, map[string]any, error) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		return "", nil, err
	}

	// Fetch regions with volume feature available
	regions, err := getCloudRegionsWithFeatureAvailable(projectID, "volume")
	if err != nil {
		return "", nil, fmt.Errorf("failed to fetch regions with volume feature available: %s", err)
	}

	// Search for the given backup in all regions
	// TODO: speed up with parallel search or by adding a required region argument
	for _, region := range regions {
		var (
			volumeBackup map[string]any
			endpoint     = fmt.Sprintf("/cloud/project/%s/region/%s/volumeBackup/%s",
				projectID, url.PathEscape(region.(string)), url.PathEscape(backupId))
		)
		if err := httpLib.Client.Get(endpoint, &volumeBackup); err == nil {
			return endpoint, volumeBackup, nil
		}
	}

	return "", nil, errors.New("no volume backup found with given ID")
}

func ListVolumeBackups(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	// Fetch regions with volume feature available
	regions, err := getCloudRegionsWithFeatureAvailable(projectID, "volume")
	if err != nil {
		display.ExitError("failed to fetch regions with volume feature available: %s", err)
		return
	}

	// Fetch volumes in all regions
	endpoint := fmt.Sprintf("/cloud/project/%s/region", projectID)
	volumeBackups, err := httpLib.FetchObjectsParallel[[]map[string]any](endpoint+"/%s/volumeBackup", regions, true)
	if err != nil {
		display.ExitError("failed to fetch volume backups: %s", err)
		return
	}

	// Flatten volumes in a single array
	var allVolumeBackups []map[string]any
	for _, regionVolumes := range volumeBackups {
		allVolumeBackups = append(allVolumeBackups, regionVolumes...)
	}

	// Filter results
	allVolumeBackups, err = filtersLib.FilterLines(allVolumeBackups, flags.GenericFilters)
	if err != nil {
		display.ExitError("failed to filter results: %s", err)
		return
	}

	display.RenderTable(allVolumeBackups, []string{"id", "name", "region", "status"}, &flags.OutputFormatConfig)
}

func GetVolumeBackup(_ *cobra.Command, args []string) {
	_, backup, err := findVolumeBackup(args[0])
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	display.OutputObject(backup, args[0], "", &flags.OutputFormatConfig)
}

func DeleteVolumeBackup(_ *cobra.Command, args []string) {
	endpoint, _, err := findVolumeBackup(args[0])
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.ExitError("failed to delete volume backup: %s", err)
		return
	}

	fmt.Printf("✅ Volume backup %s deleted successfully\n", args[0])
}

func CreateVolumeBackup(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	// Fetch volume to get its region
	var volume map[string]any
	if err := httpLib.Client.Get(
		fmt.Sprintf("/cloud/project/%s/volume/%s", projectID, url.PathEscape(args[0])),
		&volume,
	); err != nil {
		display.ExitError("failed to fetch volume: %s", err)
		return
	}

	endpoint := fmt.Sprintf("/cloud/project/%s/region/%s/volumeBackup", projectID, url.PathEscape(volume["region"].(string)))

	var (
		response map[string]any
		body     = map[string]string{
			"volumeId": args[0],
			"name":     args[1],
		}
	)
	if err := httpLib.Client.Post(endpoint, body, &response); err != nil {
		display.ExitError("failed to create volume backup: %s", err)
		return
	}

	fmt.Printf("✅ Volume backup for volume %s created successfully, id : %s\n", args[0], response["id"])
}

func RestoreVolumeBackup(_ *cobra.Command, args []string) {
	endpoint, _, err := findVolumeBackup(args[0])
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	if err := httpLib.Client.Post(
		endpoint+"/restore",
		map[string]string{"volumeId": args[1]},
		nil,
	); err != nil {
		display.ExitError("failed to restore volume backup: %s", err)
		return
	}

	fmt.Printf("✅ Volume backup %s is being restored to volume %s\n", args[0], args[1])
}

func CreateVolumeFromBackup(cmd *cobra.Command, args []string) {
	endpoint, _, err := findVolumeBackup(args[0])
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	body := map[string]string{
		"name": args[1],
	}

	var response map[string]any
	if err := httpLib.Client.Post(
		endpoint+"/volume",
		body,
		&response,
	); err != nil {
		display.ExitError("failed to create volume from backup: %s", err)
		return
	}

	fmt.Printf("✅ Volume %s created successfully from backup %s\n", response["id"], args[0])
}
