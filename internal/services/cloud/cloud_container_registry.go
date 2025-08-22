package cloud

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/ovh/ovhcloud-cli/internal/assets"
	"github.com/ovh/ovhcloud-cli/internal/display"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
	"github.com/ovh/ovhcloud-cli/internal/services/common"
	"github.com/spf13/cobra"
)

var (
	cloudprojectContainerRegistryColumnsToDisplay = []string{"id", "name", "region", "status"}

	//go:embed templates/cloud_container_registry.tmpl
	cloudContainerRegistryTemplate string

	//go:embed parameter-samples/container-registry-create.json
	CloudContainerRegistryCreateSample string

	// CloudContainerRegistryName is used to edit the container registry
	CloudContainerRegistryName string

	CloudContainerRegistrySpec struct {
		Name   string `json:"name,omitempty"`
		PlanID string `json:"planID,omitempty"`
		Region string `json:"region,omitempty"`
	}
)

func ListContainerRegistries(_ *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}
	common.ManageListRequest(fmt.Sprintf("/cloud/project/%s/containerRegistry", projectID), "id", cloudprojectContainerRegistryColumnsToDisplay, flags.GenericFilters)
}

func GetContainerRegistry(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	// Fetch registry details
	endpoint := fmt.Sprintf("/cloud/project/%s/containerRegistry/%s", projectID, url.PathEscape(args[0]))
	var object map[string]any
	if err := httpLib.Client.Get(endpoint, &object); err != nil {
		display.ExitError("error fetching %s: %s", endpoint, err)
		return
	}

	// Fetch plan details
	var plan map[string]any
	if err := httpLib.Client.Get(endpoint+"/plan", &plan); err != nil {
		display.ExitError("error fetching plan details: %s", err)
		return
	}
	object["plan"] = plan

	// Calculate and add usage information
	planLimits := plan["registryLimits"].(map[string]any)

	usedFloat, err := object["size"].(json.Number).Float64()
	if err != nil {
		display.ExitError("error parsing used storage: %s", err)
		return
	}
	availableFloat, err := planLimits["imageStorage"].(json.Number).Float64()
	if err != nil {
		display.ExitError("error parsing available storage: %s", err)
		return
	}
	object["usage"] = map[string]any{
		"used":      usedFloat,
		"available": availableFloat,
	}

	display.OutputObject(object, args[0], cloudContainerRegistryTemplate, &flags.OutputFormatConfig)
}

func EditContainerRegistry(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	if err := common.EditResource(
		cmd,
		"/cloud/project/{serviceName}/containerRegistry/{registryID}",
		fmt.Sprintf("/cloud/project/%s/containerRegistry/%s", projectID, url.PathEscape(args[0])),
		map[string]any{"name": CloudContainerRegistryName},
		assets.CloudOpenapiSchema,
	); err != nil {
		display.ExitError(err.Error())
		return
	}
}

func CreateContainerRegistry(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	registry, err := common.CreateResource(
		cmd,
		"/cloud/project/{serviceName}/containerRegistry",
		fmt.Sprintf("/cloud/project/%s/containerRegistry", projectID),
		CloudContainerRegistryCreateSample,
		CloudContainerRegistrySpec,
		assets.CloudOpenapiSchema,
		[]string{"name", "region"},
	)
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	fmt.Printf("✅ Container registry '%s' created successfully\n", registry["id"])
}

func DeleteContainerRegistry(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	endpoint := fmt.Sprintf("/cloud/project/%s/containerRegistry/%s", projectID, url.PathEscape(args[0]))
	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.ExitError("failed to delete container registry: %s", err)
		return
	}

	fmt.Println("✅ Container registry deleted successfully")
}
