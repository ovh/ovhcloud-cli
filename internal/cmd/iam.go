// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/ovh/ovhcloud-cli/internal/assets"
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

	iamPolicyCreateCmd := getGenericCreateCmd(
		"policy", "iam policy create",
		"--name MyPolicy --allow 'domain:apiovh:get' --identity 'urn:v1:eu:identity:account:aa1-ovh' --resource 'urn:v1:eu:resource:domain:*'",
		"/iam/policy", iam.IAMPolicyCreateExample,
		assets.IamOpenapiSchema, nil, iam.CreateIAMPolicy,
	)
	iamPolicyCreateCmd.Flags().StringVar(&iam.IAMPolicySpec.Name, "name", "", "Name of the policy")
	iamPolicyCreateCmd.Flags().StringVar(&iam.IAMPolicySpec.Description, "description", "", "Description of the policy")
	iamPolicyCreateCmd.Flags().StringVar(&iam.IAMPolicySpec.ExpiredAt, "expiredAt", "", "Expiration date of the policy (RFC3339 format), after this date it will no longer be applied")
	iamPolicyCreateCmd.Flags().StringSliceVar(&iam.IAMPolicySpec.Identities, "identity", nil, "Identities to which the policy applies")
	iamPolicyCreateCmd.Flags().StringSliceVar(&iam.IAMPolicySpec.PermissionsAllowed, "allow", nil, "List of allowed actions")
	iamPolicyCreateCmd.Flags().StringSliceVar(&iam.IAMPolicySpec.PermissionsDenied, "deny", nil, "List of denied actions")
	iamPolicyCreateCmd.Flags().StringSliceVar(&iam.IAMPolicySpec.PermissionsExcept, "except", nil, "List of actions to filter from the allowed list")
	iamPolicyCreateCmd.Flags().StringSliceVar(&iam.IAMPolicySpec.PermissionsGroupsURNs, "permissions-group", nil, "Permissions group URNs")
	iamPolicyCreateCmd.Flags().StringSliceVar(&iam.IAMPolicySpec.ResourcesURNs, "resource", nil, "Resource URNs")
	iamPolicyCmd.AddCommand(iamPolicyCreateCmd)

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

	iamPolicyCmd.AddCommand(&cobra.Command{
		Use:   "delete <policy_id>",
		Short: "Delete a specific IAM policy",
		Run:   iam.DeleteIAMPolicy,
		Args:  cobra.ExactArgs(1),
	})

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

	// Users
	iamUserCmd := &cobra.Command{
		Use:   "user",
		Short: "Manage IAM users",
	}
	iamCmd.AddCommand(iamUserCmd)

	iamUserCmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List IAM users",
		Run:     iam.ListUsers,
	}))

	iamUserCmd.AddCommand(&cobra.Command{
		Use:   "get <user_login>",
		Short: "Get a specific IAM user",
		Run:   iam.GetUser,
		Args:  cobra.ExactArgs(1),
	})

	iamUserCmd.AddCommand(getUserCreateCmd())
	iamUserCmd.AddCommand(getUserEditCmd())

	iamUserCmd.AddCommand(&cobra.Command{
		Use:   "delete <user_login>",
		Short: "Delete a specific IAM user",
		Run:   iam.DeleteUser,
		Args:  cobra.ExactArgs(1),
	})

	tokenCmd := &cobra.Command{
		Use:   "token",
		Short: "Manage IAM user tokens",
	}
	iamUserCmd.AddCommand(tokenCmd)

	tokenCmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:     "list <user_login>",
		Aliases: []string{"ls"},
		Short:   "List tokens of a specific IAM user",
		Run:     iam.ListUserTokens,
		Args:    cobra.ExactArgs(1),
	}))

	tokenCmd.AddCommand(&cobra.Command{
		Use:   "get <user_login> <token_name>",
		Short: "Get a specific token of an IAM user",
		Run:   iam.GetUserToken,
		Args:  cobra.ExactArgs(2),
	})

	tokenCreateCmd := getGenericCreateCmd(
		"token", "iam user token create", "--name Token --description Desc",
		"/me/identity/user/{user}/token", iam.TokenCreateExample,
		assets.MeOpenapiSchema, []string{"user_login"}, iam.CreateUserToken,
	)
	tokenCreateCmd.Flags().StringVar(&iam.TokenSpec.Name, "name", "", "Name of the token")
	tokenCreateCmd.Flags().StringVar(&iam.TokenSpec.Description, "description", "", "Description of the token")
	tokenCreateCmd.Flags().StringVar(&iam.TokenSpec.ExpiredAt, "expiredAt", "", "Expiration date of the token (RFC3339 format)")
	tokenCreateCmd.Flags().IntVar(&iam.TokenSpec.ExpiresIn, "expiresIn", 0, "Number of seconds before the token expires")
	tokenCmd.AddCommand(tokenCreateCmd)

	tokenCmd.AddCommand(&cobra.Command{
		Use:   "delete <user_login> <token_name>",
		Short: "Delete a specific token of an IAM user",
		Run:   iam.DeleteUserToken,
		Args:  cobra.ExactArgs(2),
	})

	rootCmd.AddCommand(iamCmd)
}

func getUserCreateCmd() *cobra.Command {
	userCreateCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new user",
		Long: `Use this command to create a new IAM user.
There are three ways to define the creation parameters:

1. Using only CLI flags:

	ovhcloud iam user create --login my_user --password 'MyStrongPassword123!' --email fake.email@ovhcloud.com

2. Using a configuration file:

  First you can generate an example of parameters file using the following command:

	ovhcloud iam user create --init-file ./params.json

  You will be able to choose from several examples of parameters. Once an example has been selected, the content is written in the given file.
  After editing the file to set the correct creation parameters, run:

	ovhcloud iam user create --from-file ./params.json

  Note that you can also pipe the content of the parameters file, like the following:

	cat ./params.json | ovhcloud iam user create

  In both cases, you can override the parameters in the given file using command line flags, for example:

	ovhcloud iam user create --from-file ./params.json --login nameoverriden

3. Using your default text editor:

	ovhcloud iam user create --editor

  You will be able to choose from several examples of parameters. Once an example has been selected, the CLI will open your
  default text editor to update the parameters. When saving the file, the creation will start.

  Note that it is also possible to override values in the presented examples using command line flags like the following:

	ovhcloud iam user create --editor --login nameoverriden
`,
		Run:  iam.CreateUser,
		Args: cobra.NoArgs,
	}

	userCreateCmd.Flags().StringVar(&iam.UserSpec.Login, "login", "", "Login of the user")
	userCreateCmd.Flags().StringVar(&iam.UserSpec.Email, "email", "", "Email of the user")
	userCreateCmd.Flags().StringVar(&iam.UserSpec.Description, "description", "", "Description of the user")
	userCreateCmd.Flags().StringVar(&iam.UserSpec.Group, "group", "", "Group of the user")
	userCreateCmd.Flags().StringVar(&iam.UserSpec.Password, "password", "", "Password of the user")
	userCreateCmd.Flags().StringVar(&iam.UserSpec.Type, "type", "", "Type of the user (ROOT, SERVICE, USER)")

	// Common flags for other means to define parameters
	addInitParameterFileFlag(userCreateCmd, assets.MeOpenapiSchema, "/me/identity/user", "post", iam.UserCreateExample, nil)
	addInteractiveEditorFlag(userCreateCmd)
	addFromFileFlag(userCreateCmd)
	userCreateCmd.MarkFlagsMutuallyExclusive("from-file", "editor")

	return userCreateCmd
}

func getUserEditCmd() *cobra.Command {
	userEditCmd := &cobra.Command{
		Use:   "edit <user_login>",
		Short: "Edit an existing user",
		Long: `Use this command to edit an existing IAM user.
There are three ways to define the editing parameters:

1. Using only CLI flags:

	ovhcloud iam user edit <user_login> --email fake.email+replaced@ovhcloud.com

2. Using a configuration file:

  First you can generate an example of parameters file using the following command:

	ovhcloud iam user edit --init-file ./params.json

  You will be able to choose from several examples of parameters. Once an example has been selected, the content is written in the given file.
  After editing the file to set the correct parameters, run:

	ovhcloud iam user edit <user_login> --from-file ./params.json

  Note that you can also pipe the content of the parameters file, like the following:

	cat ./params.json | ovhcloud iam user edit <user_login>

  In both cases, you can override the parameters in the given file using command line flags, for example:

	ovhcloud iam user edit <user_login> --from-file ./params.json --email fake.email+overriden@ovhcloud.com

3. Using your default text editor:

	ovhcloud iam user edit <user_login> --editor

  You will be able to choose from several examples of parameters. Once an example has been selected, the CLI will open your
  default text editor to update the parameters. When saving the file, the creation will start.

  Note that it is also possible to override values in the presented examples using command line flags like the following:

	ovhcloud iam user edit <user_login> --editor --description "New description"
`,
		Run:  iam.EditUser,
		Args: cobra.ExactArgs(1),
	}

	userEditCmd.Flags().StringVar(&iam.UserSpec.Email, "email", "", "Email of the user")
	userEditCmd.Flags().StringVar(&iam.UserSpec.Description, "description", "", "Description of the user")
	userEditCmd.Flags().StringVar(&iam.UserSpec.Group, "group", "", "Group of the user")

	// Common flags for other means to define parameters
	addInitParameterFileFlag(userEditCmd, assets.MeOpenapiSchema, "/me/identity/user", "post", iam.UserEditExample, nil)
	addInteractiveEditorFlag(userEditCmd)
	addFromFileFlag(userEditCmd)
	userEditCmd.MarkFlagsMutuallyExclusive("from-file", "editor")

	return userEditCmd
}
