// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/ovh/ovhcloud-cli/internal/assets"
	"github.com/ovh/ovhcloud-cli/internal/completion"
	"github.com/ovh/ovhcloud-cli/internal/flags"
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
		Use:               "get <policy_id>",
		Short:             "Get a specific IAM policy",
		ValidArgsFunction: completion.ServiceList("/v2/iam/policy"),
		Run:               iam.GetIAMPolicy,
		Args:              cobra.ExactArgs(1),
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
		Use:               "edit <policy_id>",
		Short:             "Edit specific IAM policy",
		ValidArgsFunction: completion.ServiceList("/v2/iam/policy"),
		Run:               iam.EditIAMPolicy,
		Args:              cobra.ExactArgs(1),
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
		Use:               "delete <policy_id>",
		Short:             "Delete a specific IAM policy",
		ValidArgsFunction: completion.ServiceList("/v2/iam/policy"),
		Run:               iam.DeleteIAMPolicy,
		Args:              cobra.ExactArgs(1),
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
		Use:               "get <permissions_group_id>",
		Short:             "Get a specific IAM permissions group",
		ValidArgsFunction: completion.ServiceList("/v2/iam/permissionsGroup"),
		Run:               iam.GetIAMPermissionsGroup,
		Args:              cobra.ExactArgs(1),
	})

	iamPermissionsGroupEditCmd := &cobra.Command{
		Use:               "edit <permissions_group_id>",
		Short:             "Edit a specific IAM permissions group",
		ValidArgsFunction: completion.ServiceList("/v2/iam/permissionsGroup"),
		Run:               iam.EditIAMPermissionsGroup,
		Args:              cobra.ExactArgs(1),
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
	iamResourceEditCmd.Flags().StringToStringVar(&iam.IAMResourceSpec.Tags, "tag", nil,
		// Pas de backticks ici : cobra les lit comme le NOM du type d argument,
		// et l aide affichait « --tag iam resource tag remove ».
		"Tags to apply, merged with the ones already there (a tag left out is kept; remove one with: iam resource tag remove)")
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
		Use:               "get <resource_group_id>",
		Short:             "Get a specific IAM resource group",
		ValidArgsFunction: completion.ServiceList("/v2/iam/resourceGroup"),
		Run:               iam.GetIAMResourceGroup,
		Args:              cobra.ExactArgs(1),
	})

	iamResourceGroupEditCmd := &cobra.Command{
		Use:               "edit <resource_group_id>",
		Short:             "Edit a specific IAM resource group",
		ValidArgsFunction: completion.ServiceList("/v2/iam/resourceGroup"),
		Run:               iam.EditIAMResourceGroup,
		Args:              cobra.ExactArgs(1),
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
		Use:               "get <user_login>",
		Short:             "Get a specific IAM user",
		ValidArgsFunction: completion.ServiceList("/v1/me/identity/user"),
		Run:               iam.GetUser,
		Args:              cobra.ExactArgs(1),
	})

	iamUserCmd.AddCommand(getUserCreateCmd())
	iamUserCmd.AddCommand(getUserEditCmd())

	iamUserCmd.AddCommand(&cobra.Command{
		Use:               "delete <user_login>",
		Short:             "Delete a specific IAM user",
		ValidArgsFunction: completion.ServiceList("/v1/me/identity/user"),
		Run:               iam.DeleteUser,
		Args:              cobra.ExactArgs(1),
	})

	tokenCmd := &cobra.Command{
		Use:   "token",
		Short: "Manage IAM user tokens",
	}
	iamUserCmd.AddCommand(tokenCmd)

	tokenCmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:               "list <user_login>",
		Aliases:           []string{"ls"},
		Short:             "List tokens of a specific IAM user",
		ValidArgsFunction: completion.ServiceList("/v1/me/identity/user"),
		Run:               iam.ListUserTokens,
		Args:              cobra.ExactArgs(1),
	}))

	tokenCmd.AddCommand(&cobra.Command{
		Use:               "get <user_login> <token_name>",
		Short:             "Get a specific token of an IAM user",
		ValidArgsFunction: completion.ServiceList("/v1/me/identity/user"),
		Run:               iam.GetUserToken,
		Args:              cobra.ExactArgs(2),
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
		Use:               "delete <user_login> <token_name>",
		Short:             "Delete a specific token of an IAM user",
		ValidArgsFunction: completion.ServiceList("/v1/me/identity/user"),
		Run:               iam.DeleteUserToken,
		Args:              cobra.ExactArgs(2),
	})

	// API credentials. They live in the /me catalogue while the rest of IAM is
	// v2; they are placed here because that is where an operator looks for
	// "what can this key do", and because Scaleway, AWS and gcloud all keep
	// API keys under their identity command.
	iamCredentialCmd := &cobra.Command{
		Use:   "credential",
		Short: "Manage the API credentials of your account",
	}
	iamCmd.AddCommand(iamCredentialCmd)

	iamCredentialListCmd := withFilterFlag(&cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List the API credentials of your account",
		Long: "List the API credentials of your account.\n\n" +
			"The scope column says how far a key reaches: a rule on /* covers the whole\n" +
			"API with that verb. Use --unused to find the keys that were never used.",
		Run: iam.ListCredentials,
	})
	iamCredentialListCmd.Flags().StringVar(&iam.CredentialStatus, "status", "", "Keep only the credentials in this state")
	iamCredentialListCmd.Flags().Int64Var(&iam.CredentialApplication, "application", 0, "Keep only the credentials of this application")
	iamCredentialListCmd.Flags().BoolVar(&iam.CredentialUnusedOnly, "unused", false, "Keep only the credentials that were never used")
	iamCredentialListCmd.RegisterFlagCompletionFunc("status", iam.CompleteCredentialStatus)
	iamCredentialCmd.AddCommand(iamCredentialListCmd)

	iamCredentialCmd.AddCommand(&cobra.Command{
		Use:   "get <credential_id>",
		Short: "Get one API credential, with the paths it may call",
		Args:  cobra.ExactArgs(1),
		Run:   iam.GetCredential,
	})

	iamCredentialDeleteCmd := &cobra.Command{
		Use:   "delete <credential_id>",
		Short: "Revoke an API credential",
		Args:  cobra.ExactArgs(1),
		Run:   iam.DeleteCredential,
	}
	addConfirmationFlags(iamCredentialDeleteCmd, "Print the call that would be made without making it")
	iamCredentialCmd.AddCommand(iamCredentialDeleteCmd)

	iamApplicationCmd := &cobra.Command{
		Use:   "application",
		Short: "Manage the applications your API credentials are issued against",
	}
	iamCmd.AddCommand(iamApplicationCmd)

	iamApplicationCmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List your API applications",
		Run:     iam.ListApplications,
	}))

	iamApplicationCmd.AddCommand(&cobra.Command{
		Use:   "get <application_id>",
		Short: "Get one API application",
		Args:  cobra.ExactArgs(1),
		Run:   iam.GetApplication,
	})

	iamApplicationDeleteCmd := &cobra.Command{
		Use:   "delete <application_id>",
		Short: "Delete an application and every credential issued against it",
		Args:  cobra.ExactArgs(1),
		Run:   iam.DeleteApplication,
	}
	addConfirmationFlags(iamApplicationDeleteCmd, "Print the call that would be made without making it")
	iamApplicationCmd.AddCommand(iamApplicationDeleteCmd)

	// Authorization check: a POST that writes nothing, and the only command
	// that answers "why can this identity not do that".
	iamCheckCmd := &cobra.Command{
		Use:   "check <action>...",
		Short: "Check whether the current identity may perform actions on resources",
		Long: "Check whether the current identity may perform actions on resources.\n\n" +
			"This writes nothing. It answers the question a permission error does not:\n" +
			"which of these actions are allowed on which resource.",
		Args: cobra.MinimumNArgs(1),
		Run:  iam.CheckAuthorization,
	}
	iamCheckCmd.Flags().StringArrayVar(&iam.CheckResources, "on", nil, "Resource URN to check against (repeatable)")
	iamCmd.AddCommand(iamCheckCmd)

	// The action reference, which is what makes a policy writable.
	iamReferenceCmd := &cobra.Command{
		Use:   "reference",
		Short: "Read what can be granted by an IAM policy",
	}
	iamCmd.AddCommand(iamReferenceCmd)

	iamReferenceActionsCmd := withFilterFlag(&cobra.Command{
		Use:   "actions",
		Short: "List the actions an IAM policy can grant",
		Long: "List the actions an IAM policy can grant.\n\n" +
			"There are over nine thousand of them across a hundred-odd product\n" +
			"families, so this asks to be narrowed rather than printing them all.",
		Run: iam.ListReferenceActions,
	})
	iamReferenceActionsCmd.Flags().StringVar(&iam.ActionResourceType, "type", "", "Keep only the actions of this resource type")
	iamReferenceActionsCmd.Flags().StringVar(&iam.ActionCategory, "category", "", "Keep only the actions in this category (READ, EDIT, DELETE...)")
	iamReferenceActionsCmd.Flags().StringVar(&iam.ActionSearch, "search", "", "Keep the actions whose name or description contains this")
	iamReferenceActionsCmd.RegisterFlagCompletionFunc("type", iam.CompleteResourceType)
	iamReferenceCmd.AddCommand(iamReferenceActionsCmd)

	iamReferenceCmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:   "resource-types",
		Short: "List the resource types actions are grouped by",
		Run:   iam.ListReferenceResourceTypes,
	}))

	// Fine-grained tagging. `iam resource edit --tag` sends a PUT that replaces
	// the whole map: adding one tag drops the others, silently and with a 200.
	iamResourceTagCmd := &cobra.Command{
		Use:   "tag",
		Short: "Add or remove resource tags without touching the others",
	}
	iamResourceCmd.AddCommand(iamResourceTagCmd)

	iamResourceTagSetCmd := &cobra.Command{
		Use:   "set <resource_urn> <key>=<value>...",
		Short: "Add or update tags, leaving the other tags in place",
		Args:  cobra.MinimumNArgs(2),
		Run:   iam.SetResourceTags,
	}
	// No --yes: setting a tag prompts for nothing, and a flag that answers a
	// question never asked is the defect P4bis found through generated docs.
	iamResourceTagSetCmd.Flags().BoolVar(&flags.DryRun, "dry-run", false,
		"Print the calls that would be made without making them")
	iamResourceTagCmd.AddCommand(iamResourceTagSetCmd)

	iamResourceTagRemoveCmd := &cobra.Command{
		Use:   "remove <resource_urn> <key>...",
		Short: "Remove tags by key, leaving the other tags in place",
		Args:  cobra.MinimumNArgs(2),
		Run:   iam.RemoveResourceTags,
	}
	iamResourceTagRemoveCmd.Flags().BoolVar(&flags.DryRun, "dry-run", false,
		"Print the calls that would be made without making them")
	iamResourceTagCmd.AddCommand(iamResourceTagRemoveCmd)

	iamUserEnableCmd := &cobra.Command{
		Use:   "enable <user_login>",
		Short: "Enable an IAM user",
		Args:  cobra.ExactArgs(1),
		Run:   iam.SetUserState(true),
	}
	iamUserEnableCmd.Flags().BoolVar(&flags.DryRun, "dry-run", false,
		"Print the call that would be made without making it")
	iamUserCmd.AddCommand(iamUserEnableCmd)

	iamUserDisableCmd := &cobra.Command{
		Use:   "disable <user_login>",
		Short: "Disable an IAM user",
		Args:  cobra.ExactArgs(1),
		Run:   iam.SetUserState(false),
	}
	addConfirmationFlags(iamUserDisableCmd, "Print the call that would be made without making it")
	iamUserCmd.AddCommand(iamUserDisableCmd)

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
	addParameterFileFlags(userCreateCmd, false, assets.MeOpenapiSchema, "/me/identity/user", "post", iam.UserCreateExample, nil)
	addInteractiveEditorFlag(userCreateCmd)
	markFlagsMutuallyExclusive(userCreateCmd, "from-file", "editor")

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
		ValidArgsFunction: completion.ServiceList("/v1/me/identity/user"),
		Run:               iam.EditUser,
		Args:              cobra.ExactArgs(1),
	}

	userEditCmd.Flags().StringVar(&iam.UserSpec.Email, "email", "", "Email of the user")
	userEditCmd.Flags().StringVar(&iam.UserSpec.Description, "description", "", "Description of the user")
	userEditCmd.Flags().StringVar(&iam.UserSpec.Group, "group", "", "Group of the user")

	// Common flags for other means to define parameters
	addParameterFileFlags(userEditCmd, false, assets.MeOpenapiSchema, "/me/identity/user", "post", iam.UserEditExample, nil)
	addInteractiveEditorFlag(userEditCmd)
	markFlagsMutuallyExclusive(userEditCmd, "from-file", "editor")

	return userEditCmd
}
