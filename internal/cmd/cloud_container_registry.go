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

	listCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List your container registries",
		Run:     cloud.ListContainerRegistries,
	}
	registryCmd.AddCommand(withFilterFlag(listCmd))

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
	initContainerRegistryIAMCommand(registryCmd)
	initContainerRegistryIPRestrictionsCommand(registryCmd)
	initContainerRegistryOIDCCommand(registryCmd)
	initContainerRegistryPlanCommand(registryCmd)

	cloudCmd.AddCommand(registryCmd)
}

func initContainerRegistryUsersCommand(registryCmd *cobra.Command) {
	usersCmd := &cobra.Command{ //nolint:exhaustruct
		Use:   "users",
		Short: "Manage container registry users",
	}

	listCmd := &cobra.Command{
		Use:     "list <registry_id>",
		Aliases: []string{"ls"},
		Short:   "List your container registry users",
		Run:     cloud.ListContainerRegistryUsers,
	}
	usersCmd.AddCommand(withFilterFlag(listCmd))

	usersCmd.AddCommand(&cobra.Command{
		Use:   "get <registry_id> <user_id>",
		Short: "Get a specific container registry user",
		Run:   cloud.GetContainerRegistryUser,
		Args:  cobra.ExactArgs(2),
	})

	createCmd := &cobra.Command{
		Use:   "create <registry_id>",
		Short: "Create a new container registry user",
		Args:  cobra.ExactArgs(1),
		Run:   cloud.CreateContainerRegistryUser,
	}
	createCmd.Flags().StringVar(&cloud.CloudContainerRegistryUserSpec.Email, "email", "", "User email")
	createCmd.Flags().StringVar(&cloud.CloudContainerRegistryUserSpec.Login, "login", "", "User login")
	addInitParameterFileFlag(createCmd, assets.CloudOpenapiSchema, "/cloud/project/{serviceName}/containerRegistry/{registryId}/users", "post", cloud.CloudContainerRegistryUserCreateSample, nil)
	addInteractiveEditorFlag(createCmd)
	addFromFileFlag(createCmd)
	createCmd.MarkFlagsMutuallyExclusive("from-file", "editor")
	usersCmd.AddCommand(createCmd)

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

func initContainerRegistryIAMCommand(registryCmd *cobra.Command) {
	iamCmd := &cobra.Command{ //nolint:exhaustruct
		Use:   "iam",
		Short: "Manage container registry IAM",
	}

	enableCmd := &cobra.Command{
		Use:   "enable <registry_id>",
		Short: "Enable IAM for the given container registry",
		Args:  cobra.ExactArgs(1),
		Run:   cloud.EnableContainerRegistryIAM,
	}
	enableCmd.Flags().BoolVar(&cloud.CloudContainerRegistryIamSpec.DeleteUsers, "delete-users", false, "Delete existing container registry users when enabling IAM")
	addInitParameterFileFlag(enableCmd, assets.CloudOpenapiSchema, "/cloud/project/{serviceName}/containerRegistry/{registryId}/iam", "post", cloud.CloudContainerRegistryIamEnableSample, nil)
	addInteractiveEditorFlag(enableCmd)
	addFromFileFlag(enableCmd)
	enableCmd.MarkFlagsMutuallyExclusive("from-file", "editor")
	iamCmd.AddCommand(enableCmd)

	iamCmd.AddCommand(&cobra.Command{
		Use:   "disable <registry_id>",
		Short: "Disable IAM for the given container registry",
		Args:  cobra.ExactArgs(1),
		Run:   cloud.DisableContainerRegistryIAM,
	})

	registryCmd.AddCommand(iamCmd)
}

func initContainerRegistryIPRestrictionsCommand(registryCmd *cobra.Command) {
	ipRestrictionsCmd := &cobra.Command{
		Use:   "ip-restrictions",
		Short: "Manage container registry IP restrictions",
	}

	initContainerRegistryIPRestrictionsManagementCommand(ipRestrictionsCmd)
	initContainerRegistryIPRestrictionsRegistryCommand(ipRestrictionsCmd)

	registryCmd.AddCommand(ipRestrictionsCmd)
}

func initContainerRegistryIPRestrictionsManagementCommand(ipRestrictionsCmd *cobra.Command) {
	// Management IP restrictions
	managementCmd := &cobra.Command{
		Use:   "management",
		Short: "Manage IP restrictions for container registry Harbor UI and API access",
	}

	listCmd := &cobra.Command{
		Use:     "list <registry_id>",
		Aliases: []string{"ls"},
		Short:   "List management IP restrictions for a container registry",
		Run:     cloud.ListContainerRegistryIPRestrictionsManagement,
		Args:    cobra.ExactArgs(1),
	}
	managementCmd.AddCommand(withFilterFlag(listCmd))

	addCmd := &cobra.Command{
		Use:   "add <registry_id>",
		Short: "Add a management IP restriction to a container registry",
		Run:   cloud.AddContainerRegistryIPRestrictionsManagement,
		Args:  cobra.ExactArgs(1),
	}
	addCmd.Flags().StringVar(&cloud.ContainerRegistryIPRestrictionsAddSpec.IPBlock, "ip-block", "", "IP block in CIDR notation (e.g., 192.0.2.0/24)")
	addCmd.MarkFlagRequired("ip-block") //nolint:errcheck
	addCmd.Flags().StringVar(&cloud.ContainerRegistryIPRestrictionsAddSpec.Description, "description", "", "Description for the IP restriction (optional)")
	managementCmd.AddCommand(addCmd)

	deleteCmd := &cobra.Command{
		Use:   "delete <registry_id>",
		Short: "Delete a management IP restriction from a container registry",
		Run:   cloud.DeleteContainerRegistryIPRestrictionsManagement,
		Args:  cobra.ExactArgs(1),
	}
	deleteCmd.Flags().StringVar(&cloud.ContainerRegistryIPRestrictionsDeleteSpec.IPBlock, "ip-block", "", "IP block in CIDR notation to delete (e.g., 192.0.2.0/24)")
	deleteCmd.MarkFlagRequired("ip-block") //nolint:errcheck
	managementCmd.AddCommand(deleteCmd)

	ipRestrictionsCmd.AddCommand(managementCmd)
}

func initContainerRegistryIPRestrictionsRegistryCommand(ipRestrictionsCmd *cobra.Command) {
	// Registry IP restrictions
	registryRestrictionsCmd := &cobra.Command{
		Use:   "registry",
		Short: "Manage IP restrictions for container registry artifact manager (Docker, Helm...) access",
	}

	listCmd := &cobra.Command{
		Use:     "list <registry_id>",
		Aliases: []string{"ls"},
		Short:   "List registry IP restrictions for a container registry",
		Run:     cloud.ListContainerRegistryIPRestrictionsRegistry,
		Args:    cobra.ExactArgs(1),
	}
	registryRestrictionsCmd.AddCommand(withFilterFlag(listCmd))

	addCmd := &cobra.Command{
		Use:   "add <registry_id>",
		Short: "Add a registry IP restriction to a container registry",
		Run:   cloud.AddContainerRegistryIPRestrictionsRegistry,
		Args:  cobra.ExactArgs(1),
	}
	addCmd.Flags().StringVar(&cloud.ContainerRegistryIPRestrictionsAddSpec.IPBlock, "ip-block", "", "IP block in CIDR notation (e.g., 192.0.2.0/24)")
	addCmd.MarkFlagRequired("ip-block") //nolint:errcheck
	addCmd.Flags().StringVar(&cloud.ContainerRegistryIPRestrictionsAddSpec.Description, "description", "", "Description for the IP restriction (optional)")
	registryRestrictionsCmd.AddCommand(addCmd)

	deleteCmd := &cobra.Command{
		Use:   "delete <registry_id>",
		Short: "Delete a registry IP restriction from a container registry",
		Run:   cloud.DeleteContainerRegistryIPRestrictionsRegistry,
		Args:  cobra.ExactArgs(1),
	}
	deleteCmd.Flags().StringVar(&cloud.ContainerRegistryIPRestrictionsDeleteSpec.IPBlock, "ip-block", "", "IP block in CIDR notation to delete (e.g., 192.0.2.0/24)")
	deleteCmd.MarkFlagRequired("ip-block") //nolint:errcheck
	registryRestrictionsCmd.AddCommand(deleteCmd)

	ipRestrictionsCmd.AddCommand(registryRestrictionsCmd)
}

func initContainerRegistryOIDCCommand(registryCmd *cobra.Command) {
	oidcCmd := &cobra.Command{ //nolint:exhaustruct
		Use:   "oidc",
		Short: "Manage container registry OIDC integration",
	}

	getCmd := &cobra.Command{
		Use:   "get <registry_id>",
		Short: "Get OIDC configuration for a container registry",
		Args:  cobra.ExactArgs(1),
		Run:   cloud.GetContainerRegistryOIDC,
	}
	oidcCmd.AddCommand(getCmd)

	createCmd := &cobra.Command{
		Use:   "create <registry_id>",
		Short: "Create a new OIDC configuration for a container registry",
		Args:  cobra.ExactArgs(1),
		Run:   cloud.CreateContainerRegistryOIDC,
	}
	createCmd.Flags().BoolVar(&cloud.CloudContainerRegistryOidcCreateSpec.DeleteUsers, "delete-users", false, "Delete existing local users when enabling OIDC")
	createCmd.Flags().StringVar(&cloud.CloudContainerRegistryOidcCreateSpec.Provider.Name, "name", "", "OIDC provider name")
	createCmd.Flags().StringVar(&cloud.CloudContainerRegistryOidcCreateSpec.Provider.Endpoint, "endpoint", "", "OIDC provider endpoint")
	createCmd.Flags().StringVar(&cloud.CloudContainerRegistryOidcCreateSpec.Provider.ClientID, "client-id", "", "OIDC client ID")
	createCmd.Flags().StringVar(&cloud.CloudContainerRegistryOidcCreateSpec.Provider.ClientSecret, "client-secret", "", "OIDC client secret")
	createCmd.Flags().StringVar(&cloud.CloudContainerRegistryOidcCreateSpec.Provider.Scope, "scope", "", "OIDC scopes")
	createCmd.Flags().StringVar(&cloud.CloudContainerRegistryOidcCreateSpec.Provider.AdminGroup, "admin-group", "", "Group granted admin role")
	createCmd.Flags().StringVar(&cloud.CloudContainerRegistryOidcCreateSpec.Provider.GroupFilter, "group-filter", "", "Regex applied to filter groups")
	createCmd.Flags().StringVar(&cloud.CloudContainerRegistryOidcCreateSpec.Provider.GroupsClaim, "groups-claim", "", "OIDC claim containing groups")
	createCmd.Flags().StringVar(&cloud.CloudContainerRegistryOidcCreateSpec.Provider.UserClaim, "user-claim", "", "OIDC claim containing the username")
	createCmd.Flags().BoolVar(&cloud.CloudContainerRegistryOidcCreateSpec.Provider.AutoOnboard, "auto-onboard", false, "Automatically create users on first login")
	createCmd.Flags().BoolVar(&cloud.CloudContainerRegistryOidcCreateSpec.Provider.VerifyCert, "verify-cert", false, "Verify the provider TLS certificate")
	addInitParameterFileFlag(createCmd, assets.CloudOpenapiSchema, "/cloud/project/{serviceName}/containerRegistry/{registryID}/openIdConnect", "post", cloud.CloudContainerRegistryOidcCreateSample, nil)
	addInteractiveEditorFlag(createCmd)
	addFromFileFlag(createCmd)
	createCmd.MarkFlagsMutuallyExclusive("from-file", "editor")
	oidcCmd.AddCommand(createCmd)

	editCmd := &cobra.Command{
		Use:   "edit <registry_id>",
		Short: "Edit the OIDC configuration for a container registry",
		Args:  cobra.ExactArgs(1),
		Run:   cloud.EditContainerRegistryOIDC,
	}
	editCmd.Flags().StringVar(&cloud.CloudContainerRegistryOidcEditSpec.AdminGroup, "admin-group", "", "Group granted admin role")
	editCmd.Flags().StringVar(&cloud.CloudContainerRegistryOidcEditSpec.ClientID, "client-id", "", "OIDC client ID")
	editCmd.Flags().StringVar(&cloud.CloudContainerRegistryOidcEditSpec.ClientSecret, "client-secret", "", "OIDC client secret")
	editCmd.Flags().StringVar(&cloud.CloudContainerRegistryOidcEditSpec.Endpoint, "endpoint", "", "OIDC provider endpoint")
	editCmd.Flags().StringVar(&cloud.CloudContainerRegistryOidcEditSpec.GroupFilter, "group-filter", "", "Regex applied to filter groups")
	editCmd.Flags().StringVar(&cloud.CloudContainerRegistryOidcEditSpec.GroupsClaim, "groups-claim", "", "OIDC claim containing groups")
	editCmd.Flags().StringVar(&cloud.CloudContainerRegistryOidcEditSpec.Name, "name", "", "OIDC provider name")
	editCmd.Flags().StringVar(&cloud.CloudContainerRegistryOidcEditSpec.Scope, "scope", "", "OIDC scopes")
	editCmd.Flags().StringVar(&cloud.CloudContainerRegistryOidcEditSpec.UserClaim, "user-claim", "", "OIDC claim containing the username")
	editCmd.Flags().BoolVar(&cloud.CloudContainerRegistryOidcEditSpec.AutoOnboard, "auto-onboard", false, "Automatically create users on first login")
	editCmd.Flags().BoolVar(&cloud.CloudContainerRegistryOidcEditSpec.VerifyCert, "verify-cert", false, "Verify the provider TLS certificate")
	addInteractiveEditorFlag(editCmd)
	oidcCmd.AddCommand(editCmd)

	deleteCmd := &cobra.Command{
		Use:   "delete <registry_id>",
		Short: "Delete the OIDC configuration for a container registry",
		Args:  cobra.ExactArgs(1),
		Run:   cloud.DeleteContainerRegistryOIDC,
	}
	oidcCmd.AddCommand(deleteCmd)

	registryCmd.AddCommand(oidcCmd)
}

func initContainerRegistryPlanCommand(registryCmd *cobra.Command) {
	planCmd := &cobra.Command{ //nolint:exhaustruct
		Use:   "plan",
		Short: "Manage container registry plans",
	}

	listCapabilitiesCmd := &cobra.Command{
		Use:   "list-capabilities <registry_id>",
		Short: "List available plans for a specific container registry",
		Args:  cobra.ExactArgs(1),
		Run:   cloud.ListContainerRegistryPlanCapabilities,
	}
	planCmd.AddCommand(withFilterFlag(listCapabilitiesCmd))

	upgradeCmd := &cobra.Command{
		Use:   "upgrade <registry_id>",
		Short: "Upgrade a container registry plan",
		Args:  cobra.ExactArgs(1),
		Run:   cloud.UpgradeContainerRegistryPlan,
	}
	upgradeCmd.Flags().StringVar(&cloud.CloudContainerRegistryPlanUpgradeSpec.PlanID, "plan-id", "", "Target plan ID for the registry")
	upgradeCmd.MarkFlagRequired("plan-id") //nolint:errcheck
	planCmd.AddCommand(upgradeCmd)

	registryCmd.AddCommand(planCmd)
}
