package cmd

import (
	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/services/iam"
)

func init() {
	iamCmd := &cobra.Command{
		Use:   "iam",
		Short: "Manage IAM resources, permissions and policies",
	}

	iamPolicyCmd := &cobra.Command{
		Use:   "policy",
		Short: "Manage IAM policies",
	}
	iamCmd.AddCommand(iamPolicyCmd)

	iamPolicyListCmd := withFilterFlag(&cobra.Command{
		Use:   "list",
		Short: "List IAM policies",
		Run:   iam.ListIAMPolicies,
	})
	iamPolicyCmd.AddCommand(iamPolicyListCmd)

	iamPolicyCmd.AddCommand(&cobra.Command{
		Use:   "get <policy_id>",
		Short: "Get a specific IAM policy",
		Run:   iam.GetIAMPolicy,
	})

	iamPolicyCmd.AddCommand(&cobra.Command{
		Use:   "edit <policy_id>",
		Short: "Edit specific IAM policy",
		Run:   iam.EditIAMPolicy,
	})

	iamPermissionsGroupCmd := &cobra.Command{
		Use:   "permissions-group",
		Short: "Manage IAM permissions groups",
	}
	iamCmd.AddCommand(iamPermissionsGroupCmd)

	iamPermissionsGroupCmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:   "list",
		Short: "List IAM permissions groups",
		Run:   iam.ListIAMPermissionsGroups,
	}))

	iamPermissionsGroupCmd.AddCommand(&cobra.Command{
		Use:   "get <permissions_group_id>",
		Short: "Get a specific IAM permissions group",
		Run:   iam.GetIAMPermissionsGroup,
	})

	iamPermissionsGroupCmd.AddCommand(&cobra.Command{
		Use:   "edit <permissions_group_id>",
		Short: "Edit a specific IAM permissions group",
		Run:   iam.EditIAMPermissionsGroup,
	})

	iamResourceCmd := &cobra.Command{
		Use:   "resource",
		Short: "Manage IAM resources",
	}
	iamCmd.AddCommand(iamResourceCmd)

	iamResourceCmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:   "list",
		Short: "List IAM resources",
		Run:   iam.ListIAMResources,
	}))

	iamResourceCmd.AddCommand(&cobra.Command{
		Use:   "get <resource_urn>",
		Short: "Get a specific IAM resource",
		Run:   iam.GetIAMResource,
	})

	iamResourceCmd.AddCommand(&cobra.Command{
		Use:   "edit <resource_urn>",
		Short: "Edit a specific IAM resource",
		Run:   iam.EditIAMResource,
	})

	iamResourceGroupCmd := &cobra.Command{
		Use:   "resource-group",
		Short: "Manage IAM resource groups",
	}
	iamCmd.AddCommand(iamResourceGroupCmd)

	iamResourceGroupCmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:   "list",
		Short: "List IAM resource groups",
		Run:   iam.ListIAMResourceGroups,
	}))

	iamResourceGroupCmd.AddCommand(&cobra.Command{
		Use:   "get <resource_group_id>",
		Short: "Get a specific IAM resource group",
		Run:   iam.GetIAMResourceGroup,
	})

	iamResourceGroupCmd.AddCommand(&cobra.Command{
		Use:   "edit <resource_group_id>",
		Short: "Edit a specific IAM resource group",
		Run:   iam.EditIAMResourceGroup,
	})

	rootCmd.AddCommand(iamCmd)
}
