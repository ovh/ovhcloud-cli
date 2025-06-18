package cloud

import (
	"bufio"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"time"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/display"
	"stash.ovh.net/api/ovh-cli/internal/editor"
	"stash.ovh.net/api/ovh-cli/internal/flags"
	httpLib "stash.ovh.net/api/ovh-cli/internal/http"
	"stash.ovh.net/api/ovh-cli/internal/openapi"
	"stash.ovh.net/api/ovh-cli/internal/services/common"
	"stash.ovh.net/api/ovh-cli/internal/utils"
)

var (
	cloudprojectInstanceColumnsToDisplay = []string{"id", "name", "region", "flavor.name", "status"}

	//go:embed templates/cloud_instance.tmpl
	cloudInstanceTemplate string

	//go:embed templates/cloud_instance_interface.tmpl
	cloudInstanceInterfaceTemplate string

	// InstanceRebootType defines the type of reboot to perform on an instance.
	// It is set with a CLI flag.
	InstanceRebootType string

	// InstanceInstallationFile is the path to the installation file for the instance.
	// It is set with a CLI flag.
	InstanceInstallationFile string

	// InstanceInstallViaEditor indicates whether to use an editor to define installation parameters.
	// It is set with a CLI flag.
	InstanceInstallViaEditor bool

	// InstanceInstallationViaInteractiveSelector indicates whether to use an interactive image selector for installation.
	// It is set with a CLI flag.
	InstanceInstallationViaInteractiveSelector bool

	// InstanceFlavorViaInteractiveSelector indicates whether to use an interactive flavor selector for setting the instance flavor.
	// It is set with a CLI flag.
	InstanceFlavorViaInteractiveSelector bool

	// InstanceImage is the image to use for reinstallation or rescue mode.
	// It is set with a CLI flag.
	InstanceImageID string
)

func ListInstances(_ *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}
	common.ManageListRequest(fmt.Sprintf("/cloud/project/%s/instance", projectID), "id", cloudprojectInstanceColumnsToDisplay, flags.GenericFilters)
}

func GetInstance(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}
	common.ManageObjectRequest(fmt.Sprintf("/cloud/project/%s/instance", projectID), args[0], cloudInstanceTemplate)
}

func StartInstance(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	endpoint := fmt.Sprintf("/cloud/project/%s/instance/%s/start", projectID, url.PathEscape(args[0]))

	if err := httpLib.Client.Post(endpoint, nil, nil); err != nil {
		display.ExitError("error starting instance %s: %s\n", args[0], err)
		return
	}

	fmt.Println("\n⚡️ Instance starting ...")
}

func StopInstance(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	endpoint := fmt.Sprintf("/cloud/project/%s/instance/%s/stop", projectID, url.PathEscape(args[0]))

	if err := httpLib.Client.Post(endpoint, nil, nil); err != nil {
		display.ExitError("error stopping instance %s: %s\n", args[0], err)
		return
	}

	fmt.Println("\n⚡️ Instance stopping ...")
}

func ShelveInstance(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	endpoint := fmt.Sprintf("/cloud/project/%s/instance/%s/shelve", projectID, url.PathEscape(args[0]))

	if err := httpLib.Client.Post(endpoint, nil, nil); err != nil {
		display.ExitError("error shelving instance %s: %s\n", args[0], err)
		return
	}

	fmt.Println("\n⚡️ Instance is being shelved ...")
}

func UnshelveInstance(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	endpoint := fmt.Sprintf("/cloud/project/%s/instance/%s/unshelve", projectID, url.PathEscape(args[0]))

	if err := httpLib.Client.Post(endpoint, nil, nil); err != nil {
		display.ExitError("error unshelving instance %s: %s\n", args[0], err)
		return
	}

	fmt.Println("\n⚡️ Instance is being unshelved ...")
}

func ResumeInstance(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	endpoint := fmt.Sprintf("/cloud/project/%s/instance/%s/resume", projectID, url.PathEscape(args[0]))

	if err := httpLib.Client.Post(endpoint, nil, nil); err != nil {
		display.ExitError("error resuming instance %s: %s\n", args[0], err)
		return
	}

	fmt.Println("\n⚡️ Instance is being resumed ...")
}

func RebootInstance(_ *cobra.Command, args []string) {
	if InstanceRebootType != "soft" && InstanceRebootType != "hard" {
		display.ExitError("invalid reboot type: %s. Use 'soft' or 'hard'.\n", InstanceRebootType)
		return
	}

	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	endpoint := fmt.Sprintf("/cloud/project/%s/instance/%s/reboot", projectID, url.PathEscape(args[0]))
	body := map[string]any{
		"type": InstanceRebootType,
	}

	if err := httpLib.Client.Post(endpoint, body, nil); err != nil {
		display.ExitError("error rebooting instance %s: %s\n", args[0], err)
		return
	}

	fmt.Println("\n⚡️ Instance is rebooting ...")
}

func ReinstallInstance(cmd *cobra.Command, args []string) {
	// No instance ID given, print usage and exit
	if len(args) == 0 {
		cmd.Help()
		display.ExitError("reinstall command requires an instance ID as the first argument.\nUsage:\n%s", cmd.UsageString())
		return
	}

	// Get cloud project ID
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	// Create object from parameters given on command line
	jsonCliParameters, err := json.Marshal(struct {
		ImageID string `json:"imageId,omitempty"`
	}{
		ImageID: InstanceImageID,
	})
	if err != nil {
		display.ExitError("failed to prepare arguments from command line: %s", err)
		return
	}
	var cliParameters map[string]any
	if err := json.Unmarshal(jsonCliParameters, &cliParameters); err != nil {
		display.ExitError("failed to parse arguments from command line: %s", err)
		return
	}

	var parameters map[string]any

	if utils.IsInputFromPipe() { // Install data given through a pipe
		var stdin []byte
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			stdin = append(stdin, scanner.Bytes()...)
		}
		if err := scanner.Err(); err != nil {
			display.ExitError(err.Error())
			return
		}

		if err := json.Unmarshal(stdin, &parameters); err != nil {
			display.ExitError("failed to parse given installation data: %s", err)
			return
		}
	} else if InstanceInstallationViaInteractiveSelector { // Install data given through an interactive image selector
		log.Print("Flag --image-selector used, all other flags will be ignored")

		// Fetch instance details to get its region
		endpoint := fmt.Sprintf("/cloud/project/%s/instance/%s", projectID, url.PathEscape(args[0]))
		var instance map[string]any
		if err := httpLib.Client.Get(endpoint, &instance); err != nil {
			display.ExitError("failed to fetch instance details: %s", err)
			return
		}
		region := instance["region"].(string)

		// Fetch available images for the project
		endpoint = fmt.Sprintf("/cloud/project/%s/image?region=%s", projectID, url.QueryEscape(region))
		var images []map[string]any
		if err := httpLib.Client.Get(endpoint, &images); err != nil {
			display.ExitError("failed to fetch available images: %s", err)
			return
		}

		if len(images) == 0 {
			display.ExitWarning("No images found for project %s, exiting...", projectID)
			return
		}

		// Let the user select an image
		imageChoices := make(map[string]string, len(images))
		for _, img := range images {
			name := img["name"].(string)
			imageChoices[name] = img["id"].(string)
		}

		selectedImage, selectedID, err := display.RunGenericChoicePicker("Please select an image to use for reinstallation", imageChoices, 30)
		if err != nil {
			display.ExitError(err.Error())
			return
		}

		if selectedImage == "" {
			display.ExitWarning("No image selected, exiting...")
			return
		}

		parameters = map[string]any{
			"imageId": selectedID,
		}

		log.Printf("Selected image %s with ID: %s", selectedImage, selectedID)
	} else if InstanceInstallViaEditor { // Install data given through an editor
		log.Print("Flag --editor used, all other flags will override the example values")

		examples, err := openapi.GetOperationRequestExamples(CloudOpenapiSchema, "/cloud/project/{serviceName}/instance/{instanceId}/reinstall", "post", cliParameters)
		if err != nil {
			display.ExitError("failed to fetch API call examples: %s", err)
			return
		}

		_, choice, err := display.RunGenericChoicePicker("Please select an installation example", examples, 0)
		if err != nil {
			display.ExitError(err.Error())
			return
		}

		if choice == "" {
			display.ExitWarning("No installation example selected, exiting...")
			return
		}

		newValue, err := editor.EditValueWithEditor([]byte(choice))
		if err != nil {
			display.ExitError("failed to edit installation parameters using editor: %s", err)
			return
		}

		if err := json.Unmarshal(newValue, &parameters); err != nil {
			display.ExitError("failed to parse given installation parameters: %s", err)
			return
		}
	} else if InstanceInstallationFile != "" { // Install data given in a file
		log.Print("Flag --installation-file used, all other flags will override the file values")

		fd, err := os.Open(InstanceInstallationFile)
		if err != nil {
			display.ExitError("failed to open given file: %s", err)
			return
		}
		defer fd.Close()

		content, err := io.ReadAll(fd)
		if err != nil {
			display.ExitError("failed to read installation file: %s", err)
			return
		}

		if err := json.Unmarshal(content, &parameters); err != nil {
			display.ExitError("failed to parse given installation file: %s", err)
			return
		}
	}

	// Only merge CLI parameters with other ones if not in --editor mode.
	// In this case, the CLI parameters have already been merged with the
	// request examples coming from API schemas.
	if !InstanceInstallViaEditor {
		parameters = utils.MergeMaps(cliParameters, parameters)
	}

	// Check if at least an image ID was provided as it is mandatory
	if imageID, ok := parameters["imageId"]; !ok || imageID == "" {
		display.ExitError("image ID parameter is mandatory to trigger a reinstallation")
		return
	}

	out, err := json.MarshalIndent(parameters, "", " ")
	if err != nil {
		display.ExitError("installation parameters cannot be marshalled: %s", err)
		return
	}

	log.Println("Installation parameters: \n" + string(out))

	var task map[string]any
	endpoint := fmt.Sprintf("/cloud/project/%s/instance/%s/reinstall", projectID, url.PathEscape(args[0]))
	if err := httpLib.Client.Post(endpoint, parameters, &task); err != nil {
		display.ExitError("error reinstalling instance %s: %s\n", args[0], err)
		return
	}

	fmt.Println("\n⚡️ Reinstallation started ...")

	if !flags.WaitForTask {
		return
	}

	if err := waitForInstanceStatus(projectID, args[0], "ACTIVE"); err != nil {
		display.ExitError("failed to wait for instance to be reinstalled: %s", err)
		return
	}

	fmt.Println("\n✅ Reinstallation done")
}

func waitForInstanceStatus(cloudProject, instanceID, targetStatus string) error {
	endpoint := fmt.Sprintf("/cloud/project/%s/instance/%s", cloudProject, url.PathEscape(instanceID))

	for range 100 {
		var instance map[string]any

		if err := httpLib.Client.Get(endpoint, &instance); err != nil {
			return fmt.Errorf("failed to fetch instance: %w", err)
		}

		switch instance["status"] {
		case "ACTIVE":
			return nil
		case "ERROR":
			return fmt.Errorf("invalid state for instance %d", instance["status"])
		default:
			log.Printf("Still waiting for instance to be in state 'ACTIVE' (status=%s) ...", instance["status"])
			time.Sleep(30 * time.Second)
		}
	}

	return fmt.Errorf("timeout waiting for instance %s to be in state 'ACTIVE'", instanceID)
}

func ActivateMonthlyBilling(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	endpoint := fmt.Sprintf("/cloud/project/%s/instance/%s/activeMonthlyBilling", projectID, url.PathEscape(args[0]))

	if err := httpLib.Client.Post(endpoint, nil, nil); err != nil {
		display.ExitError("error activating monthly billing for instance %s: %s\n", args[0], err)
		return
	}

	fmt.Println("\n✅ Monthly billing activated for instance")
}

func ListInstanceInterfaces(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	endpoint := fmt.Sprintf("/cloud/project/%s/instance/%s/interface", projectID, url.PathEscape(args[0]))

	common.ManageListRequestNoExpand(endpoint, []string{"id", "type", "macAddress", "networkId", "state"}, flags.GenericFilters)
}

func GetInstanceInterface(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	endpoint := fmt.Sprintf("/cloud/project/%s/instance/%s/interface", projectID, url.PathEscape(args[0]))

	common.ManageObjectRequest(endpoint, args[1], cloudInstanceInterfaceTemplate)
}

func CreateInstanceInterface(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	endpoint := fmt.Sprintf("/cloud/project/%s/instance/%s/interface", projectID, url.PathEscape(args[0]))
	body := map[string]any{
		"networkId": args[1],
	}

	if len(args) > 2 {
		// If a third argument is provided, use it as the IP address
		body["ip"] = args[2]
	}

	if err := httpLib.Client.Post(endpoint, body, nil); err != nil {
		display.ExitError("error creating interface for instance %s: %s\n", args[0], err)
		return
	}

	fmt.Println("\n✅ Interface created successfully")
}

func DeleteInstanceInterface(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	endpoint := fmt.Sprintf("/cloud/project/%s/instance/%s/interface/%s", projectID, url.PathEscape(args[0]), url.PathEscape(args[1]))

	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.ExitError("error deleting interface %s for instance %s: %s\n", args[1], args[0], err)
		return
	}

	fmt.Println("\n✅ Interface deleted successfully")
}

func EnableInstanceInRescueMode(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	endpoint := fmt.Sprintf("/cloud/project/%s/instance/%s/rescueMode", projectID, url.PathEscape(args[0]))
	body := map[string]any{
		"rescue": true,
	}

	if InstanceImageID != "" {
		body["imageId"] = InstanceImageID
	}

	if err := httpLib.Client.Post(endpoint, body, nil); err != nil {
		display.ExitError("error setting instance %s in rescue mode: %s\n", args[0], err)
		return
	}

	fmt.Println("\n⚡️ Instance is being rebooted in rescue mode ...")

	if !flags.WaitForTask {
		return
	}

	if err := waitForInstanceStatus(projectID, args[0], "RESCUE"); err != nil {
		display.ExitError("failed to wait for instance to be in rescue mode %s", err)
		return
	}

	fmt.Println("\n✅ Instance is now in rescue mode")
}

func DisableInstanceRescueMode(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	endpoint := fmt.Sprintf("/cloud/project/%s/instance/%s/rescueMode", projectID, url.PathEscape(args[0]))
	body := map[string]any{
		"rescue": false,
	}

	if err := httpLib.Client.Post(endpoint, body, nil); err != nil {
		display.ExitError("error unsetting instance %s from rescue mode: %s\n", args[0], err)
		return
	}

	fmt.Println("\n⚡️ Instance is exiting rescue mode ...")

	if !flags.WaitForTask {
		return
	}

	if err := waitForInstanceStatus(projectID, args[0], "ACTIVE"); err != nil {
		display.ExitError("failed to wait for instance to exit rescue mode %s", err)
		return
	}

	fmt.Println("\n✅ Instance is no longer in rescue mode")
}

func SetInstanceFlavor(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	var flavor string

	if InstanceFlavorViaInteractiveSelector {
		log.Print("Flag --flavor-selector used, all other flags will be ignored")

		// Fetch instance details to get its region
		endpoint := fmt.Sprintf("/cloud/project/%s/instance/%s", projectID, url.PathEscape(args[0]))
		var instance map[string]any
		if err := httpLib.Client.Get(endpoint, &instance); err != nil {
			display.ExitError("failed to fetch instance details: %s", err)
			return
		}
		region := instance["region"].(string)

		// Fetch available flavors for the project
		endpoint = fmt.Sprintf("/cloud/project/%s/flavor?region=%s", projectID, url.QueryEscape(region))
		var flavors []map[string]any
		if err := httpLib.Client.Get(endpoint, &flavors); err != nil {
			display.ExitError("failed to fetch available flavors: %s", err)
			return
		}

		if len(flavors) == 0 {
			display.ExitWarning("No flavors found for project %s, exiting...", projectID)
			return
		}

		// Let the user select a flavor
		flavorChoices := make(map[string]string, len(flavors))
		for _, flavor := range flavors {
			name := flavor["name"].(string)
			flavorChoices[name] = flavor["id"].(string)
		}

		selectedFlavor, selectedID, err := display.RunGenericChoicePicker("Please select a flavor to set for the instance", flavorChoices, 30)
		if err != nil {
			display.ExitError(err.Error())
			return
		}

		if selectedFlavor == "" {
			display.ExitWarning("No flavor selected, exiting...")
			return
		}

		flavor = selectedID
	} else if len(args) > 1 {
		flavor = args[1]
	} else {
		display.ExitError("Flavor ID is required when not using the --flavor-selector flag")
		return
	}

	log.Printf("Selected flavor %s", flavor)

	endpoint := fmt.Sprintf("/cloud/project/%s/instance/%s/resize", projectID, url.PathEscape(args[0]))
	body := map[string]any{
		"flavorId": flavor,
	}

	if err := httpLib.Client.Post(endpoint, body, nil); err != nil {
		display.ExitError("error setting flavor for instance %s: %s\n", args[0], err)
		return
	}

	fmt.Println("\n⚡️ Migrating instance to the desired flavor ...")

	if !flags.WaitForTask {
		return
	}

	if err := waitForInstanceStatus(projectID, args[0], "ACTIVE"); err != nil {
		display.ExitError("failed to wait for instance to migrate to the desired flavor %s", err)
		return
	}

	fmt.Println("\n✅ Instance correctly migrated to the desired flavor")
}
