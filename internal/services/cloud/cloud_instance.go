// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cloud

import (
	_ "embed"
	"errors"
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/ovh/ovhcloud-cli/internal/assets"
	"github.com/ovh/ovhcloud-cli/internal/display"
	filtersLib "github.com/ovh/ovhcloud-cli/internal/filters"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
	"github.com/ovh/ovhcloud-cli/internal/openapi"
	"github.com/ovh/ovhcloud-cli/internal/services/common"
	"github.com/spf13/cobra"
)

var (
	cloudprojectInstanceColumnsToDisplay = []string{
		"id",
		"targetSpec.name name",
		"targetSpec.location.region region",
		"targetSpec.powerState powerState",
		"resourceStatus status",
	}

	//go:embed templates/cloud_instance.tmpl
	cloudInstanceTemplate string

	//go:embed templates/cloud_instance_interface.tmpl
	cloudInstanceInterfaceTemplate string

	//go:embed templates/cloud_instance_application_access.tmpl
	cloudInstanceApplicationAccessTemplate string

	//go:embed templates/cloud_instance_autobackup.tmpl
	cloudInstanceAutobackupTemplate string

	autobackupColumnsToDisplay = []string{"id", "name", "instanceId", "cron", "rotation", "nextExecutionTime"}

	AutobackupCreateParams struct {
		InstanceID        string `json:"instanceId"`
		Name              string `json:"name"`
		Cron              string `json:"cron"`
		Rotation          int    `json:"rotation"`
		MaxExecutionCount *int   `json:"maxExecutionCount,omitempty"`
	}

	//go:embed templates/cloud_instance_group.tmpl
	cloudInstanceGroupTemplate string

	instanceGroupColumnsToDisplay = []string{"id", "name", "type", "region", "instance_ids"}

	InstanceGroupType string

	//go:embed parameter-samples/instance-create.json
	CloudInstanceCreationExample string

	// InstanceRebootType defines the type of reboot to perform on an instance.
	// It is set with a CLI flag.
	InstanceRebootType string

	// InstanceImageViaInteractiveSelector indicates whether to use an interactive image selector for installation.
	// It is set with a CLI flag.
	InstanceImageViaInteractiveSelector bool

	// InstanceFlavorViaInteractiveSelector indicates whether to use an interactive flavor selector for setting the instance flavor.
	// It is set with a CLI flag.
	InstanceFlavorViaInteractiveSelector bool

	// InstanceImage is the image to use for reinstallation or rescue mode.
	// It is set with a CLI flag.
	InstanceImageID string

	// InstanceCreationParameters holds the parameters for creating a new instance.
	InstanceCreationParameters = struct {
		Autobackup struct {
			Cron     string `json:"cron,omitempty"`
			Rotation int    `json:"rotation,omitempty"`
		} `json:"autobackup,omitzero"`
		AvailabilityZone string `json:"availabilityZone,omitempty"`
		BillingPeriod    string `json:"billingPeriod,omitempty"`
		BootFrom         struct {
			ImageID  string `json:"imageId,omitempty"`
			VolumeID string `json:"volumeId,omitempty"`
		} `json:"bootFrom,omitzero"`
		Bulk   int `json:"bulk,omitempty"`
		Flavor struct {
			ID string `json:"id,omitempty"`
		} `json:"flavor,omitzero"`
		Group struct {
			ID string `json:"id,omitempty"`
		} `json:"group,omitzero"`
		Name    string `json:"name,omitempty"`
		Network struct {
			Private struct {
				FloatingIp struct {
					ID string `json:"id,omitempty"`
				} `json:"floatingIp,omitzero"`
				FloatingIpCreate struct {
					Description string `json:"description,omitempty"`
				} `json:"floatingIpCreate,omitzero"`
				Gateway struct {
					ID string `json:"id,omitempty"`
				} `json:"gateway,omitzero"`
				GatewayCreate struct {
					Model string `json:"model,omitempty"`
					Name  string `json:"name,omitempty"`
				} `json:"gatewayCreate,omitzero"`
				IP      string `json:"ip,omitempty"`
				Network struct {
					ID       string `json:"id,omitempty"`
					SubnetID string `json:"subnetId,omitempty"`
				} `json:"network,omitzero"`
				NetworkCreate struct {
					Name   string `json:"name,omitempty"`
					Subnet struct {
						CIDR       string `json:"cidr,omitempty"`
						EnableDhcp bool   `json:"enableDhcp,omitempty"`
						IPVersion  int    `json:"ipVersion,omitempty"`
					} `json:"subnet,omitzero"`
					VlanID int `json:"vlanId,omitempty"`
				} `json:"networkCreate,omitzero"`
			} `json:"private,omitzero"`
			Public bool `json:"public,omitempty"`
		} `json:"network,omitzero"`
		SshKey struct {
			Name string `json:"name,omitempty"`
		} `json:"sshKey,omitzero"`
		SshKeyCreate struct {
			Name      string `json:"name,omitempty"`
			PublicKey string `json:"publicKey,omitempty"`
		} `json:"sshKeyCreate,omitzero"`
		UserData string `json:"userData,omitempty"`
	}{}

	InstanceBackupSpec struct {
		SnapshotName        string `json:"snapshotName,omitempty"`
		DistantSnapshotName string `json:"distantSnapshotName,omitempty"`
		DistantRegionName   string `json:"distantRegionName,omitempty"`
	}
)

func instanceV2Endpoint(projectID string) string {
	return fmt.Sprintf("/v2/publicCloud/project/%s/compute/instance", projectID)
}

func ListInstances(_ *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}
	common.ManageListRequestNoExpand(instanceV2Endpoint(projectID), cloudprojectInstanceColumnsToDisplay, flags.GenericFilters)
}

func GetInstance(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}
	common.ManageObjectRequest(instanceV2Endpoint(projectID), args[0], cloudInstanceTemplate)
}

func updateInstanceTargetSpec(projectID, instanceID string, change func(targetSpec map[string]any) error) (map[string]any, error) {
	endpoint := fmt.Sprintf("%s/%s", instanceV2Endpoint(projectID), url.PathEscape(instanceID))

	var instance map[string]any
	if err := httpLib.Client.Get(endpoint, &instance); err != nil {
		return nil, fmt.Errorf("error fetching instance %q: %w", instanceID, err)
	}
	targetSpec, ok := instance["targetSpec"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("instance %q has no targetSpec", instanceID)
	}
	if err := change(targetSpec); err != nil {
		return nil, err
	}

	body, err := openapi.FilterEditableFields(
		assets.CloudV2OpenapiSchema,
		"/publicCloud/project/{projectId}/compute/instance/{instanceId}",
		"put",
		instance,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to extract writable properties: %w", err)
	}

	if err := httpLib.Client.Put(endpoint, body, nil); err != nil {
		return nil, fmt.Errorf("error updating instance %q: %w", instanceID, err)
	}

	return waitForInstanceIfRequested(endpoint)
}

// runInstanceAction triggers an imperative action (reboot, rescue, lock…) on
// the instance. When --wait is set, it returns the instance once ready.
func runInstanceAction(projectID, instanceID, actionType string, parameters map[string]any) (map[string]any, error) {
	endpoint := fmt.Sprintf("%s/%s", instanceV2Endpoint(projectID), url.PathEscape(instanceID))

	body := map[string]any{"type": actionType}
	if len(parameters) > 0 {
		body["parameters"] = parameters
	}

	if err := httpLib.Client.Post(endpoint+"/action", body, nil); err != nil {
		return nil, fmt.Errorf("error running action %s on instance %q: %w", actionType, instanceID, err)
	}

	return waitForInstanceIfRequested(endpoint)
}

func waitForInstanceIfRequested(endpoint string) (map[string]any, error) {
	if !flags.WaitForTask {
		return nil, nil
	}

	ready, err := waitForCloudResourceReady(endpoint, 10*time.Minute)
	if err != nil {
		return nil, fmt.Errorf("failed to wait for instance to be ready: %w", err)
	}

	return ready, nil
}

// outputInstanceResult prints the message matching the --wait mode.
func outputInstanceResult(ready map[string]any, startedMessage, doneMessage string, args ...any) {
	if !flags.WaitForTask {
		display.OutputInfo(&flags.OutputFormatConfig, nil, startedMessage, args...)
		return
	}
	display.OutputInfo(&flags.OutputFormatConfig, ready, doneMessage, args...)
}

func SetInstanceName(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	ready, err := updateInstanceTargetSpec(projectID, args[0], func(targetSpec map[string]any) error {
		targetSpec["name"] = args[1]
		return nil
	})
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	outputInstanceResult(ready, "✅ Instance %s renamed to %s", "✅ Instance %s renamed to %s", args[0], args[1])
}

func setInstancePowerState(args []string, powerState, startedMessage, doneMessage string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	ready, err := updateInstanceTargetSpec(projectID, args[0], func(targetSpec map[string]any) error {
		targetSpec["powerState"] = powerState
		return nil
	})
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	outputInstanceResult(ready, startedMessage, doneMessage, args[0])
}

func StartInstance(_ *cobra.Command, args []string) {
	setInstancePowerState(args, "ACTIVE", "⚡️ Instance %s starting…", "✅ Instance %s started")
}

func StopInstance(_ *cobra.Command, args []string) {
	setInstancePowerState(args, "SHUTOFF", "⚡️ Instance %s stopping…", "✅ Instance %s stopped")
}

func ShelveInstance(_ *cobra.Command, args []string) {
	setInstancePowerState(args, "SHELVED", "⚡️ Instance %s is being shelved…", "✅ Instance %s shelved")
}

func UnshelveInstance(_ *cobra.Command, args []string) {
	setInstancePowerState(args, "ACTIVE", "⚡️ Instance %s is being unshelved…", "✅ Instance %s unshelved")
}

func instanceAction(args []string, actionType string, parameters map[string]any, startedMessage, doneMessage string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	ready, err := runInstanceAction(projectID, args[0], actionType, parameters)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	outputInstanceResult(ready, startedMessage, doneMessage, args[0])
}

func RebootInstance(_ *cobra.Command, args []string) {
	if InstanceRebootType != "soft" && InstanceRebootType != "hard" {
		display.OutputError(&flags.OutputFormatConfig, "invalid reboot type: %q. Use 'soft' or 'hard'.", InstanceRebootType)
		return
	}

	instanceAction(args, "REBOOT", map[string]any{"hard": InstanceRebootType == "hard"},
		"⚡️ Instance %s is rebooting…", "✅ Instance %s rebooted")
}

func LockInstance(_ *cobra.Command, args []string) {
	instanceAction(args, "LOCK", nil, "⚡️ Instance %s is being locked…", "✅ Instance %s locked")
}

func UnlockInstance(_ *cobra.Command, args []string) {
	instanceAction(args, "UNLOCK", nil, "⚡️ Instance %s is being unlocked…", "✅ Instance %s unlocked")
}

func CreateInstance(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		display.OutputError(&flags.OutputFormatConfig, "create command requires a region as the first argument.\n\n%s", cmd.UsageString())
		return
	}

	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to get configured cloud project: %s", err)
		return
	}
	region := args[0]

	// Run interactive image & flavor selectors if the flags are set
	interactiveParams, err := GetInstanceFlavorAndImageInteractiveSelector(cmd, args)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to get interactive parameters: %s", err)
		return
	}
	if interactiveParams != nil {
		if boot, ok := interactiveParams["bootFrom"]; ok {
			InstanceCreationParameters.BootFrom.ImageID = boot.(map[string]any)["imageId"].(string)
		}
		if flavor, ok := interactiveParams["flavor"]; ok {
			InstanceCreationParameters.Flavor.ID = flavor.(map[string]any)["id"].(string)
		}
	}

	if s := &InstanceCreationParameters.Network.Private.NetworkCreate.Subnet; s.IPVersion == 0 && s.CIDR != "" {
		s.IPVersion = ipVersionFromCIDR(s.CIDR)
	}

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/region/%s/instance", projectID, region)
	operation, err := common.CreateResource(
		cmd,
		"/cloud/project/{serviceName}/region/{regionName}/instance",
		endpoint,
		CloudInstanceCreationExample,
		InstanceCreationParameters,
		assets.CloudOpenapiSchema,
		[]string{"name", "flavor", "bootFrom", "network"})
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to create instance: %s", err)
		return
	}

	if !flags.WaitForTask {
		display.OutputInfo(&flags.OutputFormatConfig, nil, "⚡️ Instance creation started")
		return
	}

	log.Println("⚡️ Instance creation started…")

	operationID := operation["id"].(string)
	instanceID, err := waitForCloudOperation(projectID, operationID, "instance#create", time.Hour)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to wait for instance creation: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, map[string]any{"id": instanceID}, "✅ Instance %s created successfully", instanceID)
}

func DeleteInstance(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("%s/%s", instanceV2Endpoint(projectID), url.PathEscape(args[0]))

	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error deleting instance %q: %s", args[0], err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Instance %s is being deleted", args[0])
}

func GetInstanceFlavorAndImageInteractiveSelector(cmd *cobra.Command, args []string) (map[string]any, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("create command requires a region as the first argument.\nUsage:\n%s", cmd.UsageString())
	}
	region := args[0]

	projectID, err := getConfiguredCloudProject()
	if err != nil {
		return nil, err
	}

	params := map[string]any{}

	// Run interactive image selector if the flag is set
	if InstanceImageViaInteractiveSelector {
		selectedImage, selectedID, err := runImageSelector(projectID, region)
		if err != nil {
			return nil, fmt.Errorf("failed to select an image: %w", err)
		}

		if selectedImage == "" {
			return nil, errors.New("no image selected, exiting")
		}

		params["bootFrom"] = map[string]any{
			"imageId": selectedID,
		}
	}

	// Run interactive flavor selector if the flag is set
	if InstanceFlavorViaInteractiveSelector {
		selectedFlavor, selectedID, err := runFlavorSelector(projectID, region)
		if err != nil {
			return nil, fmt.Errorf("failed to select a flavor: %w", err)
		}

		if selectedFlavor == "" {
			return nil, errors.New("no flavor selected, exiting")
		}

		params["flavor"] = map[string]any{
			"id": selectedID,
		}
	}

	return params, nil
}

func ReinstallInstance(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	imageID := InstanceImageID

	if InstanceImageViaInteractiveSelector {
		region, err := instanceRegion(projectID, args[0])
		if err != nil {
			display.OutputError(&flags.OutputFormatConfig, "%s", err)
			return
		}

		// Run interactive image selector
		selectedImage, selectedID, err := runImageSelector(projectID, region)
		if err != nil {
			display.OutputError(&flags.OutputFormatConfig, "failed to select an image: %s", err)
			return
		}

		if selectedImage == "" {
			display.OutputWarning(&flags.OutputFormatConfig, "No image selected, exiting…")
			return
		}

		log.Printf("Selected image %s with ID: %s", selectedImage, selectedID)
		imageID = selectedID
	}

	if imageID == "" {
		display.OutputError(&flags.OutputFormatConfig, "image ID is required to reinstall an instance: use --image or --image-selector")
		return
	}

	ready, err := updateInstanceTargetSpec(projectID, args[0], func(targetSpec map[string]any) error {
		if currentImage, ok := targetSpec["image"].(map[string]any); ok && currentImage["id"] == imageID {
			return fmt.Errorf("instance %q is already running image %s, nothing to reinstall", args[0], imageID)
		}
		targetSpec["image"] = map[string]any{"id": imageID}
		return nil
	})
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	outputInstanceResult(ready, "⚡️ Instance %s reinstallation with image %s started…", "✅ Instance %s reinstalled with image %s", args[0], imageID)
}

func ActivateMonthlyBilling(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/instance/%s/activeMonthlyBilling", projectID, url.PathEscape(args[0]))

	if err := httpLib.Client.Post(endpoint, nil, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error activating monthly billing for instance %q: %s", args[0], err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Monthly billing activated for instance %q", args[0])
}

func ListInstanceInterfaces(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/instance/%s/interface", projectID, url.PathEscape(args[0]))

	common.ManageListRequestNoExpand(endpoint, []string{"id", "type", "macAddress", "networkId", "state"}, flags.GenericFilters)
}

func GetInstanceInterface(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/instance/%s/interface", projectID, url.PathEscape(args[0]))

	common.ManageObjectRequest(endpoint, args[1], cloudInstanceInterfaceTemplate)
}

func CreateInstanceInterface(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/instance/%s/interface", projectID, url.PathEscape(args[0]))
	body := map[string]any{
		"networkId": args[1],
	}

	if len(args) > 2 {
		// If a third argument is provided, use it as the IP address
		body["ip"] = args[2]
	}

	if err := httpLib.Client.Post(endpoint, body, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error creating interface for instance %q: %s", args[0], err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Interface created successfully")
}

func DeleteInstanceInterface(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/instance/%s/interface/%s", projectID, url.PathEscape(args[0]), url.PathEscape(args[1]))

	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error deleting interface %s for instance %q: %s", args[1], args[0], err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Interface deleted successfully")
}

func EnableInstanceInRescueMode(_ *cobra.Command, args []string) {
	var parameters map[string]any
	if InstanceImageID != "" {
		parameters = map[string]any{"imageId": InstanceImageID}
	}

	instanceAction(args, "RESCUE", parameters,
		"⚡️ Instance %s is being rebooted in rescue mode…", "✅ Instance %s is now in rescue mode")
}

func DisableInstanceRescueMode(_ *cobra.Command, args []string) {
	instanceAction(args, "UNRESCUE", nil, "⚡️ Instance %s is exiting rescue mode…", "✅ Instance %s is no longer in rescue mode")
}

// instanceRegion returns the region of the given instance.
func instanceRegion(projectID, instanceID string) (string, error) {
	var instance map[string]any
	endpoint := fmt.Sprintf("%s/%s", instanceV2Endpoint(projectID), url.PathEscape(instanceID))
	if err := httpLib.Client.Get(endpoint, &instance); err != nil {
		return "", fmt.Errorf("failed to fetch instance details: %w", err)
	}

	targetSpec, _ := instance["targetSpec"].(map[string]any)
	location, _ := targetSpec["location"].(map[string]any)
	region, _ := location["region"].(string)
	if region == "" {
		return "", fmt.Errorf("no region found for instance %q", instanceID)
	}

	return region, nil
}

func SetInstanceFlavor(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	var flavor string

	if InstanceFlavorViaInteractiveSelector {
		log.Print("Flag --flavor-selector used, all other flags will be ignored")

		region, err := instanceRegion(projectID, args[0])
		if err != nil {
			display.OutputError(&flags.OutputFormatConfig, "%s", err)
			return
		}

		// Run interactive flavor selector
		selectedFlavor, selectedID, err := runFlavorSelector(projectID, region)
		if err != nil {
			display.OutputError(&flags.OutputFormatConfig, "failed to run flavor selector: %s", err)
			return
		}

		if selectedFlavor == "" {
			display.OutputWarning(&flags.OutputFormatConfig, "No flavor selected, exiting…")
			return
		}

		flavor = selectedID
	} else if len(args) > 1 {
		flavor = args[1]
	} else {
		display.OutputError(&flags.OutputFormatConfig, "Flavor ID is required when not using the --flavor-selector flag")
		return
	}

	log.Printf("Selected flavor %s", flavor)

	ready, err := updateInstanceTargetSpec(projectID, args[0], func(targetSpec map[string]any) error {
		if currentFlavor, ok := targetSpec["flavor"].(map[string]any); ok && currentFlavor["id"] == flavor {
			return fmt.Errorf("instance %q is already running flavor %s, nothing to change", args[0], flavor)
		}
		targetSpec["flavor"] = map[string]any{"id": flavor}
		return nil
	})
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	outputInstanceResult(ready, "⚡️ Instance %s migration to flavor %s started…", "✅ Instance %s migrated to flavor %s", args[0], flavor)
}

func CreateInstanceBackup(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	// Fetch instance details to get its region
	endpoint := fmt.Sprintf("/v1/cloud/project/%s/instance/%s", projectID, url.PathEscape(args[0]))
	var instance map[string]any
	if err := httpLib.Client.Get(endpoint, &instance); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch instance details: %s", err)
		return
	}
	region := instance["region"].(string)

	InstanceBackupSpec.SnapshotName = args[1]

	endpoint = fmt.Sprintf("/v1/cloud/project/%s/region/%s/instance/%s/snapshot", projectID, url.PathEscape(region), url.PathEscape(args[0]))
	var response map[string]any
	if err := httpLib.Client.Post(endpoint, InstanceBackupSpec, &response); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error creating backup for instance %q: %s", args[0], err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, response, "✅ Backup created successfully with ID: %s", response["imageId"])
}

func AbortInstanceBackup(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	// Fetch instance details to get its region
	endpoint := fmt.Sprintf("/v1/cloud/project/%s/instance/%s", projectID, url.PathEscape(args[0]))
	var instance map[string]any
	if err := httpLib.Client.Get(endpoint, &instance); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch instance details: %s", err)
		return
	}
	region := instance["region"].(string)

	// Abort the backup
	endpoint = fmt.Sprintf("/v1/cloud/project/%s/region/%s/instance/%s/abortSnapshot", projectID, url.PathEscape(region), url.PathEscape(args[0]))
	if err := httpLib.Client.Post(endpoint, nil, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error aborting backup for instance %q: %s", args[0], err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Backup aborted successfully")
}

func ListInstanceBackups(_ *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.ManageListRequestNoExpand(fmt.Sprintf("/v1/cloud/project/%s/snapshot", projectID), []string{"id", "name", "type", "status", "region"}, flags.GenericFilters)
}

func GetInstanceBackup(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.ManageObjectRequest(fmt.Sprintf("/v1/cloud/project/%s/snapshot", projectID), args[0], "")
}

func DeleteInstanceBackup(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/snapshot/%s", projectID, url.PathEscape(args[0]))

	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error deleting backup %q: %s", args[0], err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Backup successfully deleted")
}

// Application Access

func GetInstanceApplicationAccess(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/instance/%s/applicationAccess", projectID, url.PathEscape(args[0]))

	var response map[string]any
	if err := httpLib.Client.Post(endpoint, nil, &response); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error getting application access for instance %q: %s", args[0], err)
		return
	}

	display.OutputObject(response, args[0], cloudInstanceApplicationAccessTemplate, &flags.OutputFormatConfig)
}

// Autobackup

func getInstanceRegion(projectID, instanceID string) (string, error) {
	endpoint := fmt.Sprintf("/v1/cloud/project/%s/instance/%s", projectID, url.PathEscape(instanceID))
	var instance map[string]any
	if err := httpLib.Client.Get(endpoint, &instance); err != nil {
		return "", fmt.Errorf("failed to fetch instance details: %w", err)
	}
	region, ok := instance["region"].(string)
	if !ok || region == "" {
		return "", fmt.Errorf("could not determine instance region")
	}
	return region, nil
}

func ListAutobackups(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	region, err := getInstanceRegion(projectID, args[0])
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/region/%s/workflow/backup", projectID, url.PathEscape(region))

	var backups []map[string]any
	if err := httpLib.Client.Get(endpoint, &backups); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error listing autobackups: %s", err)
		return
	}

	// Filter by instance ID
	filtered := make([]map[string]any, 0)
	for _, b := range backups {
		if id, ok := b["instanceId"].(string); ok && id == args[0] {
			filtered = append(filtered, b)
		}
	}

	filtered, err = filtersLib.FilterLines(filtered, flags.GenericFilters)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to filter results: %s", err)
		return
	}

	display.RenderTable(filtered, autobackupColumnsToDisplay, &flags.OutputFormatConfig)
}

func GetAutobackup(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	region, err := getInstanceRegion(projectID, args[0])
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/region/%s/workflow/backup", projectID, url.PathEscape(region))
	common.ManageObjectRequest(endpoint, args[1], cloudInstanceAutobackupTemplate)
}

func CreateAutobackup(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	region, err := getInstanceRegion(projectID, args[0])
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	params := AutobackupCreateParams
	params.InstanceID = args[0]

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/region/%s/workflow/backup", projectID, url.PathEscape(region))

	var response map[string]any
	if err := httpLib.Client.Post(endpoint, params, &response); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error creating autobackup for instance %q: %s", args[0], err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, response, "✅ Autobackup workflow created with ID: %s", response["id"])
}

func DeleteAutobackup(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	region, err := getInstanceRegion(projectID, args[0])
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/region/%s/workflow/backup/%s", projectID, url.PathEscape(region), url.PathEscape(args[1]))

	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error deleting autobackup %q: %s", args[1], err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Autobackup workflow deleted")
}

// Instance Groups

func ListInstanceGroups(_ *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.ManageListRequestNoExpand(fmt.Sprintf("/v1/cloud/project/%s/instance/group", projectID), instanceGroupColumnsToDisplay, flags.GenericFilters)
}

func GetInstanceGroup(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.ManageObjectRequest(fmt.Sprintf("/v1/cloud/project/%s/instance/group", projectID), args[0], cloudInstanceGroupTemplate)
}

func CreateInstanceGroup(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	if InstanceGroupType != "affinity" && InstanceGroupType != "anti-affinity" {
		display.OutputError(&flags.OutputFormatConfig, "invalid group type: %q. Use 'affinity' or 'anti-affinity'.", InstanceGroupType)
		return
	}

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/instance/group", projectID)
	body := map[string]any{
		"name":   args[0],
		"region": args[1],
		"type":   InstanceGroupType,
	}

	var response map[string]any
	if err := httpLib.Client.Post(endpoint, body, &response); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error creating instance group: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, response, "✅ Instance group created with ID: %s", response["id"])
}

func DeleteInstanceGroup(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/instance/group/%s", projectID, url.PathEscape(args[0]))

	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error deleting instance group %q: %s", args[0], err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Instance group successfully deleted")
}
