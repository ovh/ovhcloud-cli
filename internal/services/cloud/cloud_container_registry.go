package cloud

import (
	_ "embed"
	"fmt"
	"net/url"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/editor"
	"stash.ovh.net/api/ovh-cli/internal/flags"
	httpLib "stash.ovh.net/api/ovh-cli/internal/http"
	"stash.ovh.net/api/ovh-cli/internal/services/common"
)

var (
	cloudprojectContainerRegistryColumnsToDisplay = []string{"id", "name", "region", "status"}

	//go:embed templates/cloud_container_registry.tmpl
	cloudContainerRegistryTemplate string
)

func ListContainerRegistries(_ *cobra.Command, _ []string) {
	projectID := url.PathEscape(getConfiguredCloudProject())
	common.ManageListRequest(fmt.Sprintf("/cloud/project/%s/containerRegistry", projectID), "id", cloudprojectContainerRegistryColumnsToDisplay, flags.GenericFilters)
}

func GetContainerRegistry(_ *cobra.Command, args []string) {
	projectID := url.PathEscape(getConfiguredCloudProject())
	common.ManageObjectRequest(fmt.Sprintf("/cloud/project/%s/containerRegistry", projectID), args[0], cloudContainerRegistryTemplate)
}

func EditContainerRegistry(_ *cobra.Command, args []string) {
	projectID := url.PathEscape(getConfiguredCloudProject())
	url := fmt.Sprintf("/cloud/project/%s/containerRegistry/%s", projectID, url.PathEscape(args[0]))
	editor.EditResource(httpLib.Client, "/cloud/project/{serviceName}/containerRegistry/{registryID}", url, cloudOpenapiSchema)
}
