package cloud

import (
	_ "embed"
	"fmt"
	"net/url"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/assets"
	"stash.ovh.net/api/ovh-cli/internal/display"
	"stash.ovh.net/api/ovh-cli/internal/flags"
	"stash.ovh.net/api/ovh-cli/internal/services/common"
)

var (
	cloudprojectContainerRegistryColumnsToDisplay = []string{"id", "name", "region", "status"}

	//go:embed templates/cloud_container_registry.tmpl
	cloudContainerRegistryTemplate string

	// CloudContainerRegistryName is used to edit the container registry
	CloudContainerRegistryName string
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
	common.ManageObjectRequest(fmt.Sprintf("/cloud/project/%s/containerRegistry", projectID), args[0], cloudContainerRegistryTemplate)
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
