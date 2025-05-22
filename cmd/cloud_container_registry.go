package cmd

import (
	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/services/cloud"
)

func initContainerRegistryCommand(cloudCmd *cobra.Command) {
	registryCmd := &cobra.Command{
		Use:   "container-registry",
		Short: "Manage container registries in the given cloud project",
	}
	registryCmd.PersistentFlags().StringVar(&cloud.CloudProject, "cloud-project", "", "Cloud project ID")

	registryListCmd := &cobra.Command{
		Use:   "list",
		Short: "List your container registries",
		Run:   cloud.ListContainerRegistries,
	}
	registryCmd.AddCommand(withFilterFlag(registryListCmd))

	registryCmd.AddCommand(&cobra.Command{
		Use:   "get <registry_id>",
		Short: "Get a specific container registry",
		Run:   cloud.GetContainerRegistry,
		Args:  cobra.ExactArgs(1),
	})

	registryCmd.AddCommand(&cobra.Command{
		Use:   "edit <registry_id>",
		Short: "Edit the given container registry",
		Run:   cloud.EditContainerRegistry,
	})

	cloudCmd.AddCommand(registryCmd)
}
