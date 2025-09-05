// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/ovh/ovhcloud-cli/internal/services/iam"
	"github.com/spf13/cobra"
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
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List IAM policies",
		Run:     iam.ListIAMPolicies,
	})
	iamPolicyCmd.AddCommand(iamPolicyListCmd)

	iamPolicyCmd.AddCommand(&cobra.Command{
		Use:   "get <policy_id>",
		Short: "Get a specific IAM policy",
		Run:   iam.GetIAMPolicy,
		Args:  cobra.ExactArgs(1),
	})

	iamPolicyEditCmd := &cobra.Command{
		Use:   "edit <policy_id>",
		Short: "Edit specific IAM policy",
		Run:   iam.EditIAMPolicy,
		Args:  cobra.ExactArgs(1),
	}
	iamPolicyEditCmd.Flags().StringVar(&iam.IAMPolicySpec.Name, "name", "", "Name of the policy")
	iamPolicyEditCmd.Flags().StringVar(&iam.IAMPolicySpec.Description, "description", "", "Description of the policy")
	iamPolicyEditCmd.Flags().StringVar(&iam.IAMPolicySpec.ExpiredAt, "expiredAt", "", "Expiration date of the policy (RFC3339 format), after this date it will no longer be applied")
	iamPolicyEditCmd.Flags().StringSliceVar(&iam.IAMPolicySpec.Identities, "identity", nil, "Identities to which the policy applies")
	iamPolicyEditCmd.Flags().StringSliceVar(&iam.IAMPolicySpec.PermissionsAllowed, "allow", nil, "List of allowed actions")
	iamPolicyEditCmd.Flags().StringSliceVar(&iam.IAMPolicySpec.PermissionsDenied, "deny", nil, "List of denied actions")
	iamPolicyEditCmd.Flags().StringSliceVar(&iam.IAMPolicySpec.PermissionsExcept, "except", nil, "List of actions to filter from the allowed list")
	iamPolicyEditCmd.Flags().StringSliceVar(&iam.IAMPolicySpec.PermissionsGroupsURNs, "permissions-group", nil, "Permissions group URNs")
	iamPolicyEditCmd.Flags().StringSliceVar(&iam.IAMPolicySpec.ResourcesURNs, "resource", nil, "Resource URNs")
	addInteractiveEditorFlag(iamPolicyEditCmd)
	iamPolicyCmd.AddCommand(iamPolicyEditCmd)

	iamPermissionsGroupCmd := &cobra.Command{
		Use:   "permissions-group",
		Short: "Manage IAM permissions groups",
	}
	iamCmd.AddCommand(iamPermissionsGroupCmd)

	iamPermissionsGroupCmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List IAM permissions groups",
		Run:     iam.ListIAMPermissionsGroups,
	}))

	iamPermissionsGroupCmd.AddCommand(&cobra.Command{
		Use:   "get <permissions_group_id>",
		Short: "Get a specific IAM permissions group",
		Run:   iam.GetIAMPermissionsGroup,
		Args:  cobra.ExactArgs(1),
	})

	iamPermissionsGroupEditCmd := &cobra.Command{
		Use:   "edit <permissions_group_id>",
		Short: "Edit a specific IAM permissions group",
		Run:   iam.EditIAMPermissionsGroup,
		Args:  cobra.ExactArgs(1),
	}
	iamPermissionsGroupEditCmd.Flags().StringVar(&iam.IAMPolicySpec.Name, "name", "", "Name of the policy")
	iamPermissionsGroupEditCmd.Flags().StringVar(&iam.IAMPolicySpec.Description, "description", "", "Description of the policy")
	iamPermissionsGroupEditCmd.Flags().StringSliceVar(&iam.IAMPolicySpec.PermissionsAllowed, "allow", nil, "List of allowed actions")
	iamPermissionsGroupEditCmd.Flags().StringSliceVar(&iam.IAMPolicySpec.PermissionsDenied, "deny", nil, "List of denied actions")
	iamPermissionsGroupEditCmd.Flags().StringSliceVar(&iam.IAMPolicySpec.PermissionsExcept, "except", nil, "List of actions to filter from the allowed list")
	addInteractiveEditorFlag(iamPermissionsGroupEditCmd)
	iamPermissionsGroupCmd.AddCommand(iamPermissionsGroupEditCmd)

	iamResourceCmd := &cobra.Command{
		Use:   "resource",
		Short: "Manage IAM resources",
	}
	iamCmd.AddCommand(iamResourceCmd)

	iamResourceCmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List IAM resources",
		Run:     iam.ListIAMResources,
	}))

	iamResourceCmd.AddCommand(&cobra.Command{
		Use:   "get <resource_urn>",
		Short: "Get a specific IAM resource",
		Run:   iam.GetIAMResource,
		Args:  cobra.ExactArgs(1),
	})

	iamResourceEditCmd := &cobra.Command{
		Use:   "edit <resource_urn>",
		Short: "Edit a specific IAM resource",
		Run:   iam.EditIAMResource,
		Args:  cobra.ExactArgs(1),
	}
	iamResourceEditCmd.Flags().StringToStringVar(&iam.IAMResourceSpec.Tags, "tag", nil, "Tags to apply to the resource")
	addInteractiveEditorFlag(iamResourceEditCmd)
	iamResourceCmd.AddCommand(iamResourceEditCmd)

	iamResourceGroupCmd := &cobra.Command{
		Use:   "resource-group",
		Short: "Manage IAM resource groups",
	}
	iamCmd.AddCommand(iamResourceGroupCmd)

	iamResourceGroupCmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List IAM resource groups",
		Run:     iam.ListIAMResourceGroups,
	}))

	iamResourceGroupCmd.AddCommand(&cobra.Command{
		Use:   "get <resource_group_id>",
		Short: "Get a specific IAM resource group",
		Run:   iam.GetIAMResourceGroup,
		Args:  cobra.ExactArgs(1),
	})

	iamResourceGroupEditCmd := &cobra.Command{
		Use:   "edit <resource_group_id>",
		Short: "Edit a specific IAM resource group",
		Run:   iam.EditIAMResourceGroup,
		Args:  cobra.ExactArgs(1),
	}
	iamResourceGroupEditCmd.Flags().StringVar(&iam.IAMPolicySpec.Name, "name", "", "Name of the resource group")
	iamResourceGroupEditCmd.Flags().StringSliceVar(&iam.IAMPolicySpec.ResourcesURNs, "resource", nil, "List of resource URNs to include in the group")
	addInteractiveEditorFlag(iamResourceGroupEditCmd)
	iamResourceGroupCmd.AddCommand(iamResourceGroupEditCmd)

	rootCmd.AddCommand(iamCmd)
}
