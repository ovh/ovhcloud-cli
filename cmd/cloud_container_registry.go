package cmd

import (
	_ "embed"
	"fmt"
	"net/url"

	"github.com/spf13/cobra"
)

var (
	cloudprojectContainerRegistryColumnsToDisplay = []string{"id", "name", "region", "status"}

	//go:embed templates/cloud_container_registry.tmpl
	cloudContainerRegistryTemplate string
)

func listContainerRegistries(_ *cobra.Command, _ []string) {
	projectID := url.PathEscape(getConfiguredCloudProject())
	manageListRequest(fmt.Sprintf("/cloud/project/%s/containerRegistry", projectID), cloudprojectContainerRegistryColumnsToDisplay, genericFilters)
}

func getContainerRegistry(_ *cobra.Command, args []string) {
	projectID := url.PathEscape(getConfiguredCloudProject())
	manageObjectRequest(fmt.Sprintf("/cloud/project/%s/containerRegistry", projectID), args[0], cloudContainerRegistryTemplate)
}

func initContainerRegistryCommand(cloudCmd *cobra.Command) {
	registryCmd := &cobra.Command{
		Use:   "container-registry",
		Short: "Manage container registries in the given cloud project",
	}
	registryCmd.PersistentFlags().StringVar(&cloudProject, "cloud-project", "", "Cloud project ID")

	registryListCmd := &cobra.Command{
		Use:   "list",
		Short: "List your container registries",
		Run:   listContainerRegistries,
	}
	registryCmd.AddCommand(registryListCmd)

	registryCmd.AddCommand(&cobra.Command{
		Use:        "get",
		Short:      "Get a specific container registry",
		Run:        getContainerRegistry,
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"registry_id"},
	})

	cloudCmd.AddCommand(registryCmd)
}
