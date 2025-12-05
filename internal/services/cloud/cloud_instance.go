// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cloud

import (
	"bufio"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"time"

	"github.com/ovh/ovhcloud-cli/internal/assets"
	"github.com/ovh/ovhcloud-cli/internal/display"
	"github.com/ovh/ovhcloud-cli/internal/editor"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
	"github.com/ovh/ovhcloud-cli/internal/openapi"
	"github.com/ovh/ovhcloud-cli/internal/services/common"
	"github.com/ovh/ovhcloud-cli/internal/utils"
	"github.com/spf13/cobra"
)

var (
	cloudprojectInstanceColumnsToDisplay = []string{"id", "name", "region", "flavor.name", "status"}

	//go:embed templates/cloud_instance.tmpl
	cloudInstanceTemplate string

	//go:embed templates/cloud_instance_interface.tmpl
	cloudInstanceInterfaceTemplate string

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

	InstanceSnapshotSpec struct {
		SnapshotName        string `json:"snapshotName,omitempty"`
		DistantSnapshotName string `json:"distantSnapshotName,omitempty"`
		DistantRegionName   string `json:"distantRegionName,omitempty"`
	}
)

func ListInstances(_ *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}
	common.ManageListRequest(fmt.Sprintf("/v1/cloud/project/%s/instance", projectID), "id", cloudprojectInstanceColumnsToDisplay, flags.GenericFilters)
}

func GetInstance(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}
	common.ManageObjectRequest(fmt.Sprintf("/v1/cloud/project/%s/instance", projectID), args[0], cloudInstanceTemplate)
}

func SetInstanceName(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/instance/%s", projectID, url.PathEscape(args[0]))
	body := map[string]any{
		"instanceName": args[1],
	}
	if err := httpLib.Client.Put(endpoint, body, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error renaming instance %q: %s", args[0], err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Instance %s renamed to %s", args[0], args[1])
}

func StartInstance(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/instance/%s/start", projectID, url.PathEscape(args[0]))

	if err := httpLib.Client.Post(endpoint, nil, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error starting instance %q: %s", args[0], err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "⚡️ Instance starting…")
}

func StopInstance(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/instance/%s/stop", projectID, url.PathEscape(args[0]))

	if err := httpLib.Client.Post(endpoint, nil, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error stopping instance %q: %s", args[0], err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "⚡️ Instance stopping…")
}

func ShelveInstance(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/instance/%s/shelve", projectID, url.PathEscape(args[0]))

	if err := httpLib.Client.Post(endpoint, nil, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error shelving instance %q: %s", args[0], err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "⚡️ Instance is being shelved…")
}

func UnshelveInstance(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/instance/%s/unshelve", projectID, url.PathEscape(args[0]))

	if err := httpLib.Client.Post(endpoint, nil, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error unshelving instance %q: %s", args[0], err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "⚡️ Instance is being unshelved…")
}

func ResumeInstance(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/instance/%s/resume", projectID, url.PathEscape(args[0]))

	if err := httpLib.Client.Post(endpoint, nil, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error resuming instance %q: %s", args[0], err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "⚡️ Instance is being resumed…")
}

func RebootInstance(_ *cobra.Command, args []string) {
	if InstanceRebootType != "soft" && InstanceRebootType != "hard" {
		display.OutputError(&flags.OutputFormatConfig, "invalid reboot type: %q. Use 'soft' or 'hard'.", InstanceRebootType)
		return
	}

	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/instance/%s/reboot", projectID, url.PathEscape(args[0]))
	body := map[string]any{
		"type": InstanceRebootType,
	}

	if err := httpLib.Client.Post(endpoint, body, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error rebooting instance %q: %s", args[0], err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "⚡️ Instance is rebooting…")
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

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/instance/%s", projectID, url.PathEscape(args[0]))

	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error deleting instance %q: %s", args[0], err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Instance successfully deleted")
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

func ReinstallInstance(cmd *cobra.Command, args []string) {
	// No instance ID given, print usage and exit
	if len(args) == 0 {
		cmd.Help()
		display.OutputError(&flags.OutputFormatConfig, "reinstall command requires an instance ID as the first argument.\nUsage:\n%s", cmd.UsageString())
		return
	}

	// Get cloud project ID
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	// Create object from parameters given on command line
	jsonCliParameters, err := json.Marshal(struct {
		ImageID string `json:"imageId,omitempty"`
	}{
		ImageID: InstanceImageID,
	})
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to prepare arguments from command line: %s", err)
		return
	}
	var cliParameters map[string]any
	if err := json.Unmarshal(jsonCliParameters, &cliParameters); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to parse arguments from command line: %s", err)
		return
	}

	parameters := make(map[string]any)

	if utils.IsInputFromPipe() { // Install data given through a pipe
		var stdin []byte
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			stdin = append(stdin, scanner.Bytes()...)
		}
		if err := scanner.Err(); err != nil {
			display.OutputError(&flags.OutputFormatConfig, "%s", err)
			return
		}

		if err := json.Unmarshal(stdin, &parameters); err != nil {
			display.OutputError(&flags.OutputFormatConfig, "failed to parse given installation data: %s", err)
			return
		}
	} else if InstanceImageViaInteractiveSelector { // Install data given through an interactive image selector
		log.Print("Flag --image-selector used, all other flags will be ignored")

		// Fetch instance details to get its region
		endpoint := fmt.Sprintf("/v1/cloud/project/%s/instance/%s", projectID, url.PathEscape(args[0]))
		var instance map[string]any
		if err := httpLib.Client.Get(endpoint, &instance); err != nil {
			display.OutputError(&flags.OutputFormatConfig, "failed to fetch instance details: %s", err)
			return
		}
		region := instance["region"].(string)

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

		parameters = map[string]any{
			"imageId": selectedID,
		}

		log.Printf("Selected image %s with ID: %s", selectedImage, selectedID)
	} else if flags.ParametersViaEditor { // Install data given through an editor
		log.Print("Flag --editor used, all other flags will override the example values")

		examples, err := openapi.GetOperationRequestExamples(assets.CloudOpenapiSchema, "/cloud/project/{serviceName}/instance/{instanceId}/reinstall", "post", "", cliParameters)
		if err != nil {
			display.OutputError(&flags.OutputFormatConfig, "failed to fetch API call examples: %s", err)
			return
		}

		_, choice, err := display.RunGenericChoicePicker("Please select an installation example", examples, 0)
		if err != nil {
			display.OutputError(&flags.OutputFormatConfig, "%s", err)
			return
		}

		if choice == "" {
			display.OutputWarning(&flags.OutputFormatConfig, "No installation example selected, exiting…")
			return
		}

		newValue, err := editor.EditValueWithEditor([]byte(choice))
		if err != nil {
			display.OutputError(&flags.OutputFormatConfig, "failed to edit installation parameters using editor: %s", err)
			return
		}

		if err := json.Unmarshal(newValue, &parameters); err != nil {
			display.OutputError(&flags.OutputFormatConfig, "failed to parse given installation parameters: %s", err)
			return
		}
	} else if flags.ParametersFile != "" { // Install data given in a file
		log.Print("Flag --from-file used, all other flags will override the file values")

		fd, err := os.Open(flags.ParametersFile)
		if err != nil {
			display.OutputError(&flags.OutputFormatConfig, "failed to open given file: %s", err)
			return
		}
		defer fd.Close()

		if err := json.NewDecoder(fd).Decode(&parameters); err != nil {
			display.OutputError(&flags.OutputFormatConfig, "failed to parse given installation file: %s", err)
			return
		}
	}

	// Only merge CLI parameters with other ones if not in --editor mode.
	// In this case, the CLI parameters have already been merged with the
	// request examples coming from API schemas.
	if !flags.ParametersViaEditor {
		if err := utils.MergeMaps(parameters, cliParameters); err != nil {
			display.OutputError(&flags.OutputFormatConfig, "failed to merge replace values into example: %s", err)
			return
		}
	}

	// Check if at least an image ID was provided as it is mandatory
	if imageID, ok := parameters["imageId"]; !ok || imageID == "" {
		display.OutputError(&flags.OutputFormatConfig, "image ID parameter is mandatory to trigger a reinstallation")
		return
	}

	out, err := json.MarshalIndent(parameters, "", " ")
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "installation parameters cannot be marshalled: %s", err)
		return
	}

	log.Println("Installation parameters: \n" + string(out))

	var task map[string]any
	endpoint := fmt.Sprintf("/v1/cloud/project/%s/instance/%s/reinstall", projectID, url.PathEscape(args[0]))
	if err := httpLib.Client.Post(endpoint, parameters, &task); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error reinstalling instance %q: %s", args[0], err)
		return
	}

	log.Println("⚡️ Reinstallation started…")

	if !flags.WaitForTask {
		display.OutputInfo(&flags.OutputFormatConfig, nil, "⚡️ Reinstallation started…")
		return
	}

	if err := waitForInstanceStatus(projectID, args[0], "ACTIVE"); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to wait for instance to be reinstalled: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Reinstallation done")
}

func waitForInstanceStatus(cloudProject, instanceID, targetStatus string) error {
	endpoint := fmt.Sprintf("/v1/cloud/project/%s/instance/%s", cloudProject, url.PathEscape(instanceID))

	for range 100 {
		var instance map[string]any

		if err := httpLib.Client.Get(endpoint, &instance); err != nil {
			return fmt.Errorf("failed to fetch instance: %w", err)
		}

		switch instance["status"] {
		case targetStatus:
			return nil
		case "ERROR":
			return fmt.Errorf("invalid state for instance: %s", instance["status"])
		default:
			log.Printf("Still waiting for instance to be in state 'ACTIVE' (status=%s)…", instance["status"])
			time.Sleep(30 * time.Second)
		}
	}

	return fmt.Errorf("timeout waiting for instance %s to be in state 'ACTIVE'", instanceID)
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
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/instance/%s/rescueMode", projectID, url.PathEscape(args[0]))
	body := map[string]any{
		"rescue": true,
	}

	if InstanceImageID != "" {
		body["imageId"] = InstanceImageID
	}

	if err := httpLib.Client.Post(endpoint, body, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error setting instance %q in rescue mode: %s", args[0], err)
		return
	}

	log.Println("⚡️ Instance is being rebooted in rescue mode…")

	if !flags.WaitForTask {
		display.OutputInfo(&flags.OutputFormatConfig, nil, "⚡️ Instance is being rebooted in rescue mode…")
		return
	}

	if err := waitForInstanceStatus(projectID, args[0], "RESCUE"); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to wait for instance to be in rescue mode %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Instance is now in rescue mode")
}

func DisableInstanceRescueMode(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/instance/%s/rescueMode", projectID, url.PathEscape(args[0]))
	body := map[string]any{
		"rescue": false,
	}

	if err := httpLib.Client.Post(endpoint, body, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error unsetting instance %q from rescue mode: %s", args[0], err)
		return
	}

	log.Println("⚡️ Instance is exiting rescue mode…")

	if !flags.WaitForTask {
		display.OutputInfo(&flags.OutputFormatConfig, nil, "⚡️ Instance is exiting rescue mode…")
		return
	}

	if err := waitForInstanceStatus(projectID, args[0], "ACTIVE"); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to wait for instance to exit rescue mode %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Instance is no longer in rescue mode")
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

		// Fetch instance details to get its region
		endpoint := fmt.Sprintf("/v1/cloud/project/%s/instance/%s", projectID, url.PathEscape(args[0]))
		var instance map[string]any
		if err := httpLib.Client.Get(endpoint, &instance); err != nil {
			display.OutputError(&flags.OutputFormatConfig, "failed to fetch instance details: %s", err)
			return
		}
		region := instance["region"].(string)

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

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/instance/%s/resize", projectID, url.PathEscape(args[0]))
	body := map[string]any{
		"flavorId": flavor,
	}

	if err := httpLib.Client.Post(endpoint, body, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error setting flavor for instance %q: %s", args[0], err)
		return
	}

	log.Println("⚡️ Migrating instance to the desired flavor…")

	if !flags.WaitForTask {
		display.OutputInfo(&flags.OutputFormatConfig, nil, "⚡️ Instance migration to the desired flavor started…")
		return
	}

	if err := waitForInstanceStatus(projectID, args[0], "ACTIVE"); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to wait for instance to migrate to the desired flavor: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Instance correctly migrated to the desired flavor")
}

func CreateInstanceSnapshot(_ *cobra.Command, args []string) {
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

	InstanceSnapshotSpec.SnapshotName = args[1]

	endpoint = fmt.Sprintf("/v1/cloud/project/%s/region/%s/instance/%s/snapshot", projectID, url.PathEscape(region), url.PathEscape(args[0]))
	var response map[string]any
	if err := httpLib.Client.Post(endpoint, InstanceSnapshotSpec, &response); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error creating snapshot for instance %q: %s", args[0], err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, response, "✅ Snapshot created successfully with ID: %s", response["imageId"])
}

func AbortInstanceSnapshot(_ *cobra.Command, args []string) {
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

	// Abort the snapshot
	endpoint = fmt.Sprintf("/v1/cloud/project/%s/region/%s/instance/%s/abortSnapshot", projectID, url.PathEscape(region), url.PathEscape(args[0]))
	if err := httpLib.Client.Post(endpoint, nil, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error aborting snapshot for instance %q: %s", args[0], err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Snapshot aborted successfully")
}

func ListInstanceSnapshots(_ *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.ManageListRequestNoExpand(fmt.Sprintf("/v1/cloud/project/%s/snapshot", projectID), []string{"id", "name", "type", "status", "region"}, flags.GenericFilters)
}

func GetInstanceSnapshot(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.ManageObjectRequest(fmt.Sprintf("/v1/cloud/project/%s/snapshot", projectID), args[0], "")
}

func DeleteInstanceSnapshot(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/snapshot/%s", projectID, url.PathEscape(args[0]))

	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error deleting snapshot %q: %s", args[0], err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Snapshot successfully deleted")
}
