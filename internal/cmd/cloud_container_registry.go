package cmd

import (
	"github.com/ovh/ovhcloud-cli/internal/assets"
	"github.com/ovh/ovhcloud-cli/internal/services/cloud"
	"github.com/spf13/cobra"
)

func initContainerRegistryCommand(cloudCmd *cobra.Command) {
	registryCmd := &cobra.Command{
		Use:   "container-registry",
		Short: "Manage container registries in the given cloud project",
	}
	registryCmd.PersistentFlags().StringVar(&cloud.CloudProject, "cloud-project", "", "Cloud project ID")

	registryListCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List your container registries",
		Run:     cloud.ListContainerRegistries,
	}
	registryCmd.AddCommand(withFilterFlag(registryListCmd))

	registryCmd.AddCommand(&cobra.Command{
		Use:   "get <registry_id>",
		Short: "Get a specific container registry",
		Run:   cloud.GetContainerRegistry,
		Args:  cobra.ExactArgs(1),
	})

	editCmd := &cobra.Command{
		Use:   "edit <registry_id>",
		Short: "Edit the given container registry",
		Args:  cobra.ExactArgs(1),
		Run:   cloud.EditContainerRegistry,
	}
	editCmd.Flags().StringVar(&cloud.CloudContainerRegistryName, "name", "", "New name for the container registry")
	addInteractiveEditorFlag(editCmd)
	registryCmd.AddCommand(editCmd)

	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new container registry",
		Run:   cloud.CreateContainerRegistry,
	}
	createCmd.Flags().StringVar(&cloud.CloudContainerRegistrySpec.Name, "name", "", "Name of the container registry")
	createCmd.Flags().StringVar(&cloud.CloudContainerRegistrySpec.PlanID, "plan-id", "", "Plan ID for the container registry. Available plans can be listed with 'ovhcloud cloud reference container-registry list-plans'")
	createCmd.Flags().StringVar(&cloud.CloudContainerRegistrySpec.Region, "region", "", "Region for the container registry (e.g., DE, GRA, BHS)")
	addInitParameterFileFlag(createCmd, assets.CloudOpenapiSchema, "/cloud/project/{serviceName}/containerRegistry", "post", cloud.CloudContainerRegistryCreateSample, nil)
	addInteractiveEditorFlag(createCmd)
	addFromFileFlag(createCmd)
	createCmd.MarkFlagsMutuallyExclusive("from-file", "editor")
	registryCmd.AddCommand(createCmd)

	registryCmd.AddCommand(&cobra.Command{
		Use:   "delete <registry_id>",
		Short: "Delete a specific container registry",
		Run:   cloud.DeleteContainerRegistry,
		Args:  cobra.ExactArgs(1),
	})

	cloudCmd.AddCommand(registryCmd)
}
