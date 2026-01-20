// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

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

	initContainerRegistryUsersCommand(registryCmd)

	cloudCmd.AddCommand(registryCmd)
}

func initContainerRegistryUsersCommand(registryCmd *cobra.Command) {
	usersCmd := &cobra.Command{ //nolint:exhaustruct
		Use:   "users",
		Short: "Manage container registry users in the given cloud project",
	}

	registryListCmd := &cobra.Command{
		Use:     "list <registry_id>",
		Aliases: []string{"ls"},
		Short:   "List your container registry users",
		Run:     cloud.ListContainerRegistryUsers,
	}
	usersCmd.AddCommand(withFilterFlag(registryListCmd))

	usersCmd.AddCommand(&cobra.Command{
		Use:   "get <registry_id> <user_id>",
		Short: "Get a specific container registry user",
		Run:   cloud.GetContainerRegistryUser,
		Args:  cobra.ExactArgs(2),
	})

	usersCreateCmd := &cobra.Command{
		Use:   "create <registry_id>",
		Short: "Create a new container registry user",
		Args:  cobra.ExactArgs(1),
		Run:   cloud.CreateContainerRegistryUser,
	}
	usersCreateCmd.Flags().StringVar(&cloud.CloudContainerUserSpec.Email, "email", "", "User email")
	usersCreateCmd.Flags().StringVar(&cloud.CloudContainerUserSpec.Login, "login", "", "User login")
	addInitParameterFileFlag(usersCreateCmd, assets.CloudOpenapiSchema, "/cloud/project/{serviceName}/containerRegistry/{registryId}/users", "post", cloud.CloudContainerRegistryUserCreateSample, nil)
	addInteractiveEditorFlag(usersCreateCmd)
	addFromFileFlag(usersCreateCmd)
	usersCreateCmd.MarkFlagsMutuallyExclusive("from-file", "editor")
	usersCmd.AddCommand(usersCreateCmd)

	usersCmd.AddCommand(&cobra.Command{
		Use:   "set-as-admin <registry_id> <user_id>",
		Short: "Set a specific container registry user as admin",
		Run:   cloud.SetContainerRegistryUserAsAdmin,
		Args:  cobra.ExactArgs(2),
	})

	usersCmd.AddCommand(&cobra.Command{
		Use:   "delete <registry_id> <user_id>",
		Short: "Delete a specific container registry user",
		Run:   cloud.DeleteContainerRegistryUser,
		Args:  cobra.ExactArgs(2),
	})

	registryCmd.AddCommand(usersCmd)
}
