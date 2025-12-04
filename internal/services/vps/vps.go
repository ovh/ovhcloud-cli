// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package vps

import (
	_ "embed"
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/ovh/go-ovh/ovh"
	"github.com/ovh/ovhcloud-cli/internal/assets"
	"github.com/ovh/ovhcloud-cli/internal/display"
	filtersLib "github.com/ovh/ovhcloud-cli/internal/filters"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
	"github.com/ovh/ovhcloud-cli/internal/services/common"
	"github.com/spf13/cobra"
)

var (
	vpsColumnsToDisplay = []string{"name", "displayName", "state", "zone"}

	//go:embed templates/vps.tmpl
	vpsTemplate string

	//go:embed templates/vps-snapshot.tmpl
	vpsSnapshotTemplate string

	//go:embed parameter-samples/reinstall.json
	VpsReinstallExample string

	VpsSpec struct {
		DisplayName   string `json:"displayName,omitempty"`
		Keymap        string `json:"keymap,omitempty"`
		NetbootMode   string `json:"netbootMode,omitempty"`
		SlaMonitoring bool   `json:"slaMonitoring,omitempty"`
	}

	VpsSnapshotSpec struct {
		Description string `json:"description,omitempty"`
	}

	VpsSnapshotRestoreSpec struct {
		ChangePassword bool   `json:"changePassword"`
		RestorePoint   string `json:"restorePoint,omitempty"`
		Type           string `json:"type,omitempty"`
	}

	VpsBackupRestorePointsState string

	VpsContacts struct {
		ContactAdmin   string `json:"contactAdmin,omitempty"`
		ContactBilling string `json:"contactBilling,omitempty"`
		ContactTech    string `json:"contactTech,omitempty"`
	}

	VpsDiskSpec struct {
		LowFreeSpaceThreshold int  `json:"lowFreeSpaceThreshold,omitempty"`
		Monitoring            bool `json:"monitoring,omitempty"`
	}

	VpsReinstallSpec struct {
		DoNotSendPassword bool   `json:"doNotSendPassword,omitempty"`
		ImageId           string `json:"imageId,omitempty"`
		InstallRTM        bool   `json:"installRTM,omitempty"`
		PublicSshKey      string `json:"publicSshKey,omitempty"`
		SshKey            string `json:"sshKey,omitempty"`
	}

	VpsImageViaInteractiveSelector  bool
	VpsSSHKeyViaInteractiveSelector bool

	VpsSecondaryDNSDomainSpec struct {
		// Creation parameters
		Domain string `json:"domain,omitempty"`
		IP     string `json:"ip,omitempty"`
	}
)

func ListVps(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/v1/vps", "", vpsColumnsToDisplay, flags.GenericFilters)
}

func GetVps(_ *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("/v1/vps/%s", url.PathEscape(args[0]))

	var object map[string]any
	if err := httpLib.Client.Get(endpoint, &object); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error fetching %s: %s", endpoint, err)
		return
	}

	// Fetch datanceter information
	var datacenter map[string]any
	if err := httpLib.Client.Get(endpoint+"/datacenter", &datacenter); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error fetching datacenter information for %s: %s", args[0], err)
		return
	}
	object["datacenter"] = datacenter

	display.OutputObject(object, args[0], vpsTemplate, &flags.OutputFormatConfig)
}

func EditVps(cmd *cobra.Command, args []string) {
	if err := common.EditResource(
		cmd,
		"/vps/{serviceName}",
		fmt.Sprintf("/v1/vps/%s", url.PathEscape(args[0])),
		VpsSpec,
		assets.VpsOpenapiSchema,
	); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}
}

func GetVpsSnapshot(_ *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("/v1/vps/%s/snapshot", url.PathEscape(args[0]))

	var object map[string]any
	if err := httpLib.Client.Get(endpoint, &object); err != nil {
		if ovhErr, ok := err.(*ovh.APIError); ok && ovhErr.Code == 404 {
			display.OutputWarning(&flags.OutputFormatConfig, "VPS %s does not have any snapshot yet", args[0])
			return
		}
		display.OutputError(&flags.OutputFormatConfig, "error fetching %s: %s", endpoint, err)
		return
	}

	display.OutputObject(object, args[0], vpsSnapshotTemplate, &flags.OutputFormatConfig)
}

func EditVpsSnapshot(cmd *cobra.Command, args []string) {
	if err := common.EditResource(
		cmd,
		"/vps/{serviceName}/snapshot",
		fmt.Sprintf("/v1/vps/%s/snapshot", url.PathEscape(args[0])),
		VpsSnapshotSpec,
		assets.VpsOpenapiSchema,
	); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}
}

func CreateVpsSnapshot(_ *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("/v1/vps/%s/createSnapshot", url.PathEscape(args[0]))

	if err := httpLib.Client.Post(endpoint, VpsSnapshotSpec, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error creating snapshot for %s: %s", args[0], err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "⚡️ Snapshot creation started")
}

func DeleteVpsSnapshot(_ *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("/v1/vps/%s/snapshot", url.PathEscape(args[0]))

	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error deleting snapshot for %s: %s", args[0], err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "⚡️ Snapshot deletion started")
}

func AbortVpsSnapshot(_ *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("/v1/vps/%s/abortSnapshot", url.PathEscape(args[0]))

	if err := httpLib.Client.Post(endpoint, nil, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error aborting snapshot for %s: %s", args[0], err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Snapshot aborted")
}

func RestoreVpsSnapshot(_ *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("/v1/vps/%s/snapshot/revert", url.PathEscape(args[0]))

	if err := httpLib.Client.Post(endpoint, nil, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error restoring snapshot for %s: %s", args[0], err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "⚡️ Snapshot restoration started")
}

func DownloadVpsSnapshot(_ *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("/v1/vps/%s/snapshot/download", url.PathEscape(args[0]))

	var response map[string]any
	if err := httpLib.Client.Get(endpoint, &response); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error downloading snapshot for %s: %s", args[0], err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Snapshot download URL: %s (size: %s bytes)", response["url"], response["size"])
}

func GetVpsAutomatedBackup(_ *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("/v1/vps/%s/automatedBackup", url.PathEscape(args[0]))

	var object map[string]any
	if err := httpLib.Client.Get(endpoint, &object); err != nil {
		if ovhErr, ok := err.(*ovh.APIError); ok && ovhErr.Code == 404 {
			display.OutputWarning(&flags.OutputFormatConfig, "VPS %s does not have any automated backup yet", args[0])
			return
		}
		display.OutputError(&flags.OutputFormatConfig, "error fetching %s: %s", endpoint, err)
		return
	}

	display.OutputObject(object, args[0], "", &flags.OutputFormatConfig)
}

func ListVpsAutomatedBackups(_ *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("/v1/vps/%s/automatedBackup/attachedBackup", url.PathEscape(args[0]))
	common.ManageListRequestNoExpand(endpoint, []string{"restorePoint"}, flags.GenericFilters)
}

func DetachVpsAutomatedBackup(_ *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("/v1/vps/%s/automatedBackup/detachBackup", url.PathEscape(args[0]))
	body := map[string]any{
		"restorePoint": args[1],
	}

	if err := httpLib.Client.Post(endpoint, body, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error detaching automated backup for %s: %s", args[0], err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Automated backup detached")
}

func RescheduleVpsAutomatedBackup(_ *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("/v1/vps/%s/automatedBackup/reschedule", url.PathEscape(args[0]))
	body := map[string]any{
		"schedule": args[1],
	}

	if err := httpLib.Client.Post(endpoint, body, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error updating automated backup schedule for %s: %s", args[0], err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Automated backup schedule updated")
}

func RestoreVpsAutomatedBackup(_ *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("/v1/vps/%s/automatedBackup/restore", url.PathEscape(args[0]))
	if err := httpLib.Client.Post(endpoint, VpsSnapshotRestoreSpec, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error restoring automated backup for %s: %s", args[0], err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "⚡️ Automated backup restoration started")
}

func ListVpsAutomatedBackupRestorePoints(_ *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("/v1/vps/%s/automatedBackup/restorePoints?state=%s", url.PathEscape(args[0]), url.QueryEscape(VpsBackupRestorePointsState))

	var restorePoints []string
	if err := httpLib.Client.Get(endpoint, &restorePoints); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error fetching restore points for %s: %s", args[0], err)
		return
	}

	pointsMaps := make([]map[string]any, 0, len(restorePoints))
	for _, point := range restorePoints {
		pointsMaps = append(pointsMaps, map[string]any{
			"restorePoint": point,
		})
	}

	display.RenderTable(pointsMaps, []string{"restorePoint"}, &flags.OutputFormatConfig)
}

func ListVpsAvailableUpgrades(_ *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("/v1/vps/%s/availableUpgrade", url.PathEscape(args[0]))
	common.ManageListRequestNoExpand(endpoint, []string{"name", "offer", "vcore", "memory", "disk"}, flags.GenericFilters)
}

func ChangeVpsContacts(_ *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("/v1/vps/%s/changeContact", url.PathEscape(args[0]))

	if err := httpLib.Client.Post(endpoint, VpsContacts, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error changing contacts for %s: %s", args[0], err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Contacts updated")
}

func GetVpsServiceInfo(_ *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("/v1/vps/%s/serviceInfos", url.PathEscape(args[0]))

	var object map[string]any
	if err := httpLib.Client.Get(endpoint, &object); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error fetching service info for %s: %s", args[0], err)
		return
	}

	display.OutputObject(object, args[0], common.ServiceInfoTemplate, &flags.OutputFormatConfig)
}

func EditVpsServiceInfo(cmd *cobra.Command, args []string) {
	if err := common.EditResource(
		cmd,
		"/vps/{serviceName}/serviceInfos",
		fmt.Sprintf("/v1/vps/%s/serviceInfos", url.PathEscape(args[0])),
		common.ServiceInfoSpec,
		assets.VpsOpenapiSchema,
	); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}
}

func TerminateVps(_ *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("/v1/vps/%s/terminate", url.PathEscape(args[0]))

	var response string
	if err := httpLib.Client.Post(endpoint, nil, &response); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error terminating VPS %s: %s", args[0], err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ VPS %s termination started: %s", args[0], response)
}

func ConfirmVpsTermination(_ *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("/v1/vps/%s/confirmTermination", url.PathEscape(args[0]))

	body := map[string]any{
		"token": args[1],
	}

	if err := httpLib.Client.Post(endpoint, body, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error confirming termination for %s: %s", args[0], err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ VPS %s termination confirmed", args[0])
}

func ListVpsDisks(_ *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("/v1/vps/%s/disks", url.PathEscape(args[0]))
	common.ManageListRequest(endpoint, "", []string{"id", "serviceName", "size", "type", "state"}, flags.GenericFilters)
}

func GetVpsDisk(_ *cobra.Command, args []string) {
	common.ManageObjectRequest(fmt.Sprintf("/v1/vps/%s/disks", url.PathEscape(args[0])), args[1], "")
}

func EditVpsDisk(cmd *cobra.Command, args []string) {
	if err := common.EditResource(
		cmd,
		"/vps/{serviceName}/disks/{id}",
		fmt.Sprintf("/v1/vps/%s/disks/%s", url.PathEscape(args[0]), url.PathEscape(args[1])),
		VpsDiskSpec,
		assets.VpsOpenapiSchema,
	); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}
}

func VpsGetConsoleURL(_ *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("/v1/vps/%s/getConsoleUrl", url.PathEscape(args[0]))

	var consoleURL string
	if err := httpLib.Client.Post(endpoint, nil, &consoleURL); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error fetching console URL for %s: %s", args[0], err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Console URL for VPS %s: %s", args[0], consoleURL)
}

func GetVpsImages(_ *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("/v1/vps/%s/images/available", url.PathEscape(args[0]))

	// Fetch available images
	body, err := httpLib.FetchExpandedArray(endpoint, "")
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch results: %s", err)
		return
	}
	for _, object := range body {
		object["current"] = false
	}

	body, err = filtersLib.FilterLines(body, flags.GenericFilters)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to filter results: %s", err)
		return
	}

	// Fetch current image
	var current map[string]any
	endpoint = fmt.Sprintf("/v1/vps/%s/images/current", url.PathEscape(args[0]))
	if err := httpLib.Client.Get(endpoint, &current); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch current image: %s", err)
		return
	}
	current["current"] = true

	// Prepend current image to the list
	body = append([]map[string]any{current}, body...)

	display.RenderTable(body, []string{"id", "name", "current"}, &flags.OutputFormatConfig)
}

func ListVpsIPs(_ *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("/v1/vps/%s/ips", url.PathEscape(args[0]))
	common.ManageListRequest(endpoint, "", []string{"ipAddress", "reverse", "type", "geolocation", "gateway", "macAddress"}, flags.GenericFilters)
}

func SetVpsIPReverse(_ *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("/v1/vps/%s/ips/%s", url.PathEscape(args[0]), url.PathEscape(args[1]))

	body := map[string]any{
		"reverse": args[2],
	}

	if err := httpLib.Client.Put(endpoint, body, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error setting reverse for IP %s: %s", args[1], err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Reverse for IP %s on VPS %s set to '%s'", args[1], args[0], args[2])
}

func ReleaseVpsIP(_ *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("/v1/vps/%s/ips/%s", url.PathEscape(args[0]), url.PathEscape(args[1]))

	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error releasing IP %s: %s", args[1], err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ IP %s released from VPS %s", args[1], args[0])
}

func ListVPSOptions(_ *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("/v1/vps/%s/option", url.PathEscape(args[0]))
	common.ManageListRequest(endpoint, "", []string{"option", "state"}, flags.GenericFilters)
}

func StartVps(_ *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("/v1/vps/%s/start", url.PathEscape(args[0]))

	var response map[string]any
	if err := httpLib.Client.Post(endpoint, nil, &response); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error starting VPS %s: %s", args[0], err)
		return
	}

	log.Printf("⚡️ VPS %s starting…", args[0])

	if !flags.WaitForTask {
		display.OutputInfo(&flags.OutputFormatConfig, nil, "⚡️ VPS %s starting…", args[0])
		return
	}

	// Wait for the task to complete
	if _, err := waitForVpsTask(args[0], response, 10*time.Minute); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error waiting for start task to complete: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ VPS %s started successfully", args[0])
}

func StopVps(_ *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("/v1/vps/%s/stop", url.PathEscape(args[0]))

	var response map[string]any
	if err := httpLib.Client.Post(endpoint, nil, &response); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error stopping VPS %s: %s", args[0], err)
		return
	}

	log.Printf("⚡️ VPS %s stopping", args[0])

	if !flags.WaitForTask {
		display.OutputInfo(&flags.OutputFormatConfig, nil, "⚡️ VPS %s stopping", args[0])
		return
	}

	// Wait for the task to complete
	if _, err := waitForVpsTask(args[0], response, 10*time.Minute); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error waiting for stop task to complete: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ VPS %s stopped successfully", args[0])
}

func RebootVps(_ *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("/v1/vps/%s/reboot", url.PathEscape(args[0]))

	var response map[string]any
	if err := httpLib.Client.Post(endpoint, nil, &response); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error rebooting VPS %s: %s", args[0], err)
		return
	}

	log.Printf("⚡️ VPS %s reboot started…", args[0])

	if !flags.WaitForTask {
		display.OutputInfo(&flags.OutputFormatConfig, nil, "⚡️ VPS %s reboot started…", args[0])
		return
	}

	// Wait for the task to complete
	if _, err := waitForVpsTask(args[0], response, 10*time.Minute); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error waiting for reboot task to complete: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ VPS %s reboot completed successfully", args[0])
}

func ReinstallVps(cmd *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("/v1/vps/%s/rebuild", url.PathEscape(args[0]))

	if VpsImageViaInteractiveSelector {
		_, id, err := runImageSelector(args[0])
		if err != nil {
			display.OutputError(&flags.OutputFormatConfig, "error selecting image: %s", err)
			return
		}
		VpsReinstallSpec.ImageId = id
	}

	if VpsSSHKeyViaInteractiveSelector {
		keyName, _, err := runSSHKeySelector()
		if err != nil {
			display.OutputError(&flags.OutputFormatConfig, "error selecting SSH key: %s", err)
			return
		}
		VpsReinstallSpec.SshKey = keyName
	}

	response, err := common.CreateResource(
		cmd,
		"/vps/{serviceName}/rebuild",
		endpoint,
		VpsReinstallExample,
		VpsReinstallSpec,
		assets.VpsOpenapiSchema,
		[]string{"imageId"},
	)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error preparing reinstallation: %s", err)
		return
	}

	log.Printf("⚡️ VPS %s reinstallation started", args[0])

	if !flags.WaitForTask {
		display.OutputInfo(&flags.OutputFormatConfig, nil, "⚡️ VPS %s reinstallation started", args[0])
		return
	}

	// Wait for the task to complete
	if _, err := waitForVpsTask(args[0], response, 20*time.Minute); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error waiting for reinstall task to complete: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ VPS %s reinstalled successfully", args[0])
}

func ListVpsSecondaryDNSDomains(_ *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("/v1/vps/%s/secondaryDnsDomains", url.PathEscape(args[0]))
	common.ManageListRequest(endpoint, "", []string{"domain", "dns", "ipMaster", "creationDate"}, flags.GenericFilters)
}

func AddVpsSecondaryDNSDomain(cmd *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("/v1/vps/%s/secondaryDnsDomains", url.PathEscape(args[0]))

	if _, err := common.CreateResource(
		cmd,
		"/vps/{serviceName}/secondaryDnsDomains",
		endpoint,
		"",
		VpsSecondaryDNSDomainSpec,
		assets.VpsOpenapiSchema,
		[]string{"domain", "ip"},
	); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error creating secondary DNS domain for %s: %s", args[0], err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Secondary DNS domain %s created for VPS %s", VpsSecondaryDNSDomainSpec.Domain, args[0])
}

func DeleteVpsSecondaryDNSDomain(_ *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("/v1/vps/%s/secondaryDnsDomains/%s", url.PathEscape(args[0]), url.PathEscape(args[1]))

	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error deleting secondary DNS domain %s: %s", args[1], err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Secondary DNS domain %s deleted from VPS %s", args[1], args[0])
}

func ChangeVpsPassword(_ *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("/v1/vps/%s/setPassword", url.PathEscape(args[0]))

	var response map[string]any
	if err := httpLib.Client.Post(endpoint, nil, &response); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error changing password for %s: %s", args[0], err)
		return
	}

	log.Printf("⚡️ VPS %s process to set the root password has started", args[0])

	if !flags.WaitForTask {
		display.OutputInfo(&flags.OutputFormatConfig, nil, "⚡️ VPS %s process to set the root password has started", args[0])
		return
	}

	// Wait for the task to complete
	if _, err := waitForVpsTask(args[0], response, 20*time.Minute); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error waiting for task to complete: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ VPS %s process to set the root password completed successfully", args[0])
}

func ListVpsTasks(_ *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("/v1/vps/%s/tasks", url.PathEscape(args[0]))
	common.ManageListRequest(endpoint, "", []string{"id", "type", "state", "date", "progress"}, flags.GenericFilters)
}
