package vps

import (
	_ "embed"
	"fmt"
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

		// Update parameter
		IPMaster string `json:"ipMaster,omitempty"`
	}
)

func ListVps(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/vps", "", vpsColumnsToDisplay, flags.GenericFilters)
}

func GetVps(_ *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("/vps/%s", url.PathEscape(args[0]))

	var object map[string]any
	if err := httpLib.Client.Get(endpoint, &object); err != nil {
		display.ExitError("error fetching %s: %s", endpoint, err)
		return
	}

	// Fetch datanceter information
	var datacenter map[string]any
	if err := httpLib.Client.Get(endpoint+"/datacenter", &datacenter); err != nil {
		display.ExitError("error fetching datacenter information for %s: %s", args[0], err)
		return
	}
	object["datacenter"] = datacenter

	display.OutputObject(object, args[0], vpsTemplate, &flags.OutputFormatConfig)
}

func EditVps(cmd *cobra.Command, args []string) {
	if err := common.EditResource(
		cmd,
		"/vps/{serviceName}",
		fmt.Sprintf("/vps/%s", url.PathEscape(args[0])),
		VpsSpec,
		assets.VpsOpenapiSchema,
	); err != nil {
		display.ExitError(err.Error())
		return
	}
}

func GetVpsSnapshot(_ *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("/vps/%s/snapshot", url.PathEscape(args[0]))

	var object map[string]any
	if err := httpLib.Client.Get(endpoint, &object); err != nil {
		if ovhErr, ok := err.(*ovh.APIError); ok && ovhErr.Code == 404 {
			display.ExitWarning("VPS %s does not have any snapshot yet", args[0])
			return
		}
		display.ExitError("error fetching %s: %s", endpoint, err)
		return
	}

	display.OutputObject(object, args[0], vpsSnapshotTemplate, &flags.OutputFormatConfig)
}

func EditVpsSnapshot(cmd *cobra.Command, args []string) {
	if err := common.EditResource(
		cmd,
		"/vps/{serviceName}/snapshot",
		fmt.Sprintf("/vps/%s/snapshot", url.PathEscape(args[0])),
		VpsSnapshotSpec,
		assets.VpsOpenapiSchema,
	); err != nil {
		display.ExitError(err.Error())
		return
	}
}

func CreateVpsSnapshot(_ *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("/vps/%s/createSnapshot", url.PathEscape(args[0]))

	if err := httpLib.Client.Post(endpoint, VpsSnapshotSpec, nil); err != nil {
		display.ExitError("error creating snapshot for %s: %s", args[0], err)
		return
	}

	fmt.Println("\n⚡️ Snapshot creation started")
}

func DeleteVpsSnapshot(_ *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("/vps/%s/snapshot", url.PathEscape(args[0]))

	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.ExitError("error deleting snapshot for %s: %s", args[0], err)
		return
	}

	fmt.Println("\n⚡️ Snapshot deletion started")
}

func AbortVpsSnapshot(_ *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("/vps/%s/abortSnapshot", url.PathEscape(args[0]))

	if err := httpLib.Client.Post(endpoint, nil, nil); err != nil {
		display.ExitError("error aborting snapshot for %s: %s", args[0], err)
		return
	}

	fmt.Println("✅ Snapshot aborted")
}

func RestoreVpsSnapshot(_ *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("/vps/%s/snapshot/revert", url.PathEscape(args[0]))

	if err := httpLib.Client.Post(endpoint, nil, nil); err != nil {
		display.ExitError("error restoring snapshot for %s: %s", args[0], err)
		return
	}

	fmt.Println("\n⚡️ Snapshot restoration started")
}

func DownloadVpsSnapshot(_ *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("/vps/%s/snapshot/download", url.PathEscape(args[0]))

	var response map[string]any
	if err := httpLib.Client.Get(endpoint, &response); err != nil {
		display.ExitError("error downloading snapshot for %s: %s", args[0], err)
		return
	}

	fmt.Printf("✅ Snapshot download URL: %s (size: %s bytes)\n", response["url"], response["size"])
}

func GetVpsAutomatedBackup(_ *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("/vps/%s/automatedBackup", url.PathEscape(args[0]))

	var object map[string]any
	if err := httpLib.Client.Get(endpoint, &object); err != nil {
		if ovhErr, ok := err.(*ovh.APIError); ok && ovhErr.Code == 404 {
			display.ExitWarning("VPS %s does not have any automated backup yet", args[0])
			return
		}
		display.ExitError("error fetching %s: %s", endpoint, err)
		return
	}

	display.OutputObject(object, args[0], "", &flags.OutputFormatConfig)
}

func ListVpsAutomatedBackups(_ *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("/vps/%s/automatedBackup/attachedBackup", url.PathEscape(args[0]))
	common.ManageListRequestNoExpand(endpoint, []string{"restorePoint"}, flags.GenericFilters)
}

func DetachVpsAutomatedBackup(_ *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("/vps/%s/automatedBackup/detachBackup", url.PathEscape(args[0]))
	body := map[string]any{
		"restorePoint": args[1],
	}

	if err := httpLib.Client.Post(endpoint, body, nil); err != nil {
		display.ExitError("error detaching automated backup for %s: %s", args[0], err)
		return
	}

	fmt.Println("✅ Automated backup detached")
}

func RescheduleVpsAutomatedBackup(_ *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("/vps/%s/automatedBackup/reschedule", url.PathEscape(args[0]))
	body := map[string]any{
		"schedule": args[1],
	}

	if err := httpLib.Client.Post(endpoint, body, nil); err != nil {
		display.ExitError("error updating automated backup schedule for %s: %s", args[0], err)
		return
	}

	fmt.Println("✅ Automated backup schedule updated")
}

func RestoreVpsAutomatedBackup(_ *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("/vps/%s/automatedBackup/restore", url.PathEscape(args[0]))
	if err := httpLib.Client.Post(endpoint, VpsSnapshotRestoreSpec, nil); err != nil {
		display.ExitError("error restoring automated backup for %s: %s", args[0], err)
		return
	}

	fmt.Println("\n⚡️ Automated backup restoration started")
}

func ListVpsAutomatedBackupRestorePoints(_ *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("/vps/%s/automatedBackup/restorePoints?state=%s", url.PathEscape(args[0]), url.QueryEscape(VpsBackupRestorePointsState))

	var restorePoints []string
	if err := httpLib.Client.Get(endpoint, &restorePoints); err != nil {
		display.ExitError("error fetching restore points for %s: %s", args[0], err)
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
	endpoint := fmt.Sprintf("/vps/%s/availableUpgrade", url.PathEscape(args[0]))
	common.ManageListRequestNoExpand(endpoint, []string{"name", "offer", "vcore", "memory", "disk"}, flags.GenericFilters)
}

func ChangeVpsContacts(_ *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("/vps/%s/changeContact", url.PathEscape(args[0]))

	if err := httpLib.Client.Post(endpoint, VpsContacts, nil); err != nil {
		display.ExitError("error changing contacts for %s: %s", args[0], err)
		return
	}

	fmt.Println("✅ Contacts updated")
}

func GetVpsServiceInfo(_ *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("/vps/%s/serviceInfos", url.PathEscape(args[0]))

	var object map[string]any
	if err := httpLib.Client.Get(endpoint, &object); err != nil {
		display.ExitError("error fetching service info for %s: %s", args[0], err)
		return
	}

	display.OutputObject(object, args[0], common.ServiceInfoTemplate, &flags.OutputFormatConfig)
}

func EditVpsServiceInfo(cmd *cobra.Command, args []string) {
	if err := common.EditResource(
		cmd,
		"/vps/{serviceName}/serviceInfos",
		fmt.Sprintf("/vps/%s/serviceInfos", url.PathEscape(args[0])),
		common.ServiceInfoSpec,
		assets.VpsOpenapiSchema,
	); err != nil {
		display.ExitError(err.Error())
		return
	}
}

func TerminateVps(_ *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("/vps/%s/terminate", url.PathEscape(args[0]))

	var response string
	if err := httpLib.Client.Post(endpoint, nil, &response); err != nil {
		display.ExitError("error terminating VPS %s: %s", args[0], err)
		return
	}

	fmt.Printf("✅ VPS %s termination started: %s\n", args[0], response)
}

func ConfirmVpsTermination(_ *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("/vps/%s/confirmTermination", url.PathEscape(args[0]))

	body := map[string]any{
		"token": args[1],
	}

	if err := httpLib.Client.Post(endpoint, body, nil); err != nil {
		display.ExitError("error confirming termination for %s: %s", args[0], err)
		return
	}

	fmt.Printf("✅ VPS %s termination confirmed\n", args[0])
}

func ListVpsDisks(_ *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("/vps/%s/disks", url.PathEscape(args[0]))
	common.ManageListRequest(endpoint, "", []string{"id", "serviceName", "size", "type", "state"}, flags.GenericFilters)
}

func GetVpsDisk(_ *cobra.Command, args []string) {
	common.ManageObjectRequest(fmt.Sprintf("/vps/%s/disks", url.PathEscape(args[0])), args[1], "")
}

func EditVpsDisk(cmd *cobra.Command, args []string) {
	if err := common.EditResource(
		cmd,
		"/vps/{serviceName}/disks/{id}",
		fmt.Sprintf("/vps/%s/disks/%s", url.PathEscape(args[0]), url.PathEscape(args[1])),
		VpsDiskSpec,
		assets.VpsOpenapiSchema,
	); err != nil {
		display.ExitError(err.Error())
		return
	}
}

func VpsGetConsoleURL(_ *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("/vps/%s/getConsoleUrl", url.PathEscape(args[0]))

	var consoleURL string
	if err := httpLib.Client.Post(endpoint, nil, &consoleURL); err != nil {
		display.ExitError("error fetching console URL for %s: %s", args[0], err)
		return
	}

	fmt.Printf("✅ Console URL for VPS %s: %s\n", args[0], consoleURL)
}

func GetVpsImages(_ *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("/vps/%s/images/available", url.PathEscape(args[0]))

	// Fetch available images
	body, err := httpLib.FetchExpandedArray(endpoint, "")
	if err != nil {
		display.ExitError("failed to fetch results: %s", err)
		return
	}
	for _, object := range body {
		object["current"] = false
	}

	body, err = filtersLib.FilterLines(body, flags.GenericFilters)
	if err != nil {
		display.ExitError("failed to filter results: %s", err)
		return
	}

	// Fetch current image
	var current map[string]any
	endpoint = fmt.Sprintf("/vps/%s/images/current", url.PathEscape(args[0]))
	if err := httpLib.Client.Get(endpoint, &current); err != nil {
		display.ExitError("failed to fetch current image: %s", err)
		return
	}
	current["current"] = true

	// Prepend current image to the list
	body = append([]map[string]any{current}, body...)

	display.RenderTable(body, []string{"id", "name", "current"}, &flags.OutputFormatConfig)
}

func ListVpsIPs(_ *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("/vps/%s/ips", url.PathEscape(args[0]))
	common.ManageListRequest(endpoint, "", []string{"ipAddress", "reverse", "type", "geolocation", "gateway", "macAddress"}, flags.GenericFilters)
}

func SetVpsIPReverse(_ *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("/vps/%s/ips/%s", url.PathEscape(args[0]), url.PathEscape(args[1]))

	body := map[string]any{
		"reverse": args[2],
	}

	if err := httpLib.Client.Put(endpoint, body, nil); err != nil {
		display.ExitError("error setting reverse for IP %s: %s", args[1], err)
		return
	}

	fmt.Printf("✅ Reverse for IP %s on VPS %s set to '%s'\n", args[1], args[0], args[2])
}

func ReleaseVpsIP(_ *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("/vps/%s/ips/%s", url.PathEscape(args[0]), url.PathEscape(args[1]))

	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.ExitError("error releasing IP %s: %s", args[1], err)
		return
	}

	fmt.Printf("✅ IP %s released from VPS %s\n", args[1], args[0])
}

func ListVPSOptions(_ *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("/vps/%s/option", url.PathEscape(args[0]))
	common.ManageListRequest(endpoint, "", []string{"option", "state"}, flags.GenericFilters)
}

func StartVps(_ *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("/vps/%s/start", url.PathEscape(args[0]))

	var response map[string]any
	if err := httpLib.Client.Post(endpoint, nil, &response); err != nil {
		display.ExitError("error starting VPS %s: %s", args[0], err)
		return
	}

	fmt.Printf("\n⚡️ VPS %s starting…\n", args[0])

	if !flags.WaitForTask {
		return
	}

	// Wait for the task to complete
	if _, err := waitForVpsTask(args[0], response, 10*time.Minute); err != nil {
		display.ExitError("error waiting for start task to complete: %s", err)
		return
	}

	fmt.Printf("✅ VPS %s started successfully\n", args[0])
}

func StopVps(_ *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("/vps/%s/stop", url.PathEscape(args[0]))

	var response map[string]any
	if err := httpLib.Client.Post(endpoint, nil, &response); err != nil {
		display.ExitError("error stopping VPS %s: %s", args[0], err)
		return
	}

	fmt.Printf("\n⚡️ VPS %s stopping\n", args[0])

	if !flags.WaitForTask {
		return
	}

	// Wait for the task to complete
	if _, err := waitForVpsTask(args[0], response, 10*time.Minute); err != nil {
		display.ExitError("error waiting for stop task to complete: %s", err)
		return
	}

	fmt.Printf("✅ VPS %s stopped successfully\n", args[0])
}

func RebootVps(_ *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("/vps/%s/reboot", url.PathEscape(args[0]))

	var response map[string]any
	if err := httpLib.Client.Post(endpoint, nil, &response); err != nil {
		display.ExitError("error rebooting VPS %s: %s", args[0], err)
		return
	}

	fmt.Printf("\n⚡️ VPS %s reboot started…\n", args[0])

	if !flags.WaitForTask {
		return
	}

	// Wait for the task to complete
	if _, err := waitForVpsTask(args[0], response, 10*time.Minute); err != nil {
		display.ExitError("error waiting for reboot task to complete: %s", err)
		return
	}

	fmt.Printf("✅ VPS %s reboot completed successfully\n", args[0])
}

func ReinstallVps(cmd *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("/vps/%s/rebuild", url.PathEscape(args[0]))

	if VpsImageViaInteractiveSelector {
		_, id, err := runImageSelector(args[0])
		if err != nil {
			display.ExitError("error selecting image: %s", err)
			return
		}
		VpsReinstallSpec.ImageId = id
	}

	if VpsSSHKeyViaInteractiveSelector {
		keyName, _, err := runSSHKeySelector()
		if err != nil {
			display.ExitError("error selecting SSH key: %s", err)
			return
		}
		VpsReinstallSpec.SshKey = keyName
	}

	response, err := common.CreateResource(
		"/vps/{serviceName}/rebuild",
		endpoint,
		VpsReinstallExample,
		VpsReinstallSpec,
		assets.VpsOpenapiSchema,
		[]string{"imageId"},
	)
	if err != nil {
		display.ExitError("error preparing reinstallation: %s", err)
		return
	}

	fmt.Printf("\n⚡️ VPS %s reinstallation started\n", args[0])

	if !flags.WaitForTask {
		return
	}

	// Wait for the task to complete
	if _, err := waitForVpsTask(args[0], response, 20*time.Minute); err != nil {
		display.ExitError("error waiting for reinstall task to complete: %s", err)
		return
	}

	fmt.Printf("✅ VPS %s reinstalled successfully\n", args[0])
}

func ListVpsSecondaryDNSDomains(_ *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("/vps/%s/secondaryDnsDomains", url.PathEscape(args[0]))
	common.ManageListRequest(endpoint, "", []string{"domain", "dns", "ipMaster", "creationDate"}, flags.GenericFilters)
}

func EditVpsSecondaryDNSDomain(cmd *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("/vps/%s/secondaryDnsDomains/%s", url.PathEscape(args[0]), url.PathEscape(args[1]))

	if err := common.EditResource(
		cmd,
		"/vps/{serviceName}/secondaryDnsDomains/{domain}",
		endpoint,
		VpsSecondaryDNSDomainSpec,
		assets.VpsOpenapiSchema,
	); err != nil {
		display.ExitError(err.Error())
		return
	}
}

func AddVpsSecondaryDNSDomain(_ *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("/vps/%s/secondaryDnsDomains", url.PathEscape(args[0]))

	if _, err := common.CreateResource(
		"/vps/{serviceName}/secondaryDnsDomains",
		endpoint,
		"",
		VpsSecondaryDNSDomainSpec,
		assets.VpsOpenapiSchema,
		[]string{"domain", "ip"},
	); err != nil {
		display.ExitError("error creating secondary DNS domain for %s: %s", args[0], err)
		return
	}

	fmt.Printf("✅ Secondary DNS domain %s created for VPS %s\n", VpsSecondaryDNSDomainSpec.Domain, args[0])
}

func DeleteVpsSecondaryDNSDomain(_ *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("/vps/%s/secondaryDnsDomains/%s", url.PathEscape(args[0]), url.PathEscape(args[1]))

	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.ExitError("error deleting secondary DNS domain %s: %s", args[1], err)
		return
	}

	fmt.Printf("✅ Secondary DNS domain %s deleted from VPS %s\n", args[1], args[0])
}

func ChangeVpsPassword(_ *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("/vps/%s/setPassword", url.PathEscape(args[0]))

	var response map[string]any
	if err := httpLib.Client.Post(endpoint, nil, &response); err != nil {
		display.ExitError("error changing password for %s: %s", args[0], err)
		return
	}

	fmt.Printf("\n⚡️ VPS %s process to set the root password has started\n", args[0])

	if !flags.WaitForTask {
		return
	}

	// Wait for the task to complete
	if _, err := waitForVpsTask(args[0], response, 20*time.Minute); err != nil {
		display.ExitError("error waiting for task to complete: %s", err)
		return
	}

	fmt.Printf("✅ VPS %s process to set the root password completed successfully\n", args[0])
}

func ListVpsTasks(_ *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("/vps/%s/tasks", url.PathEscape(args[0]))
	common.ManageListRequest(endpoint, "", []string{"id", "type", "state", "date", "progress"}, flags.GenericFilters)
}
