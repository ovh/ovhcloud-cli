package cmd

import (
	"github.com/ovh/ovhcloud-cli/internal/assets"
	"github.com/ovh/ovhcloud-cli/internal/services/cloud"
	"github.com/spf13/cobra"
)

func initCloudUserCommand(cloudCmd *cobra.Command) {
	userCmd := &cobra.Command{
		Use:   "user",
		Short: "Manage users in the given cloud project",
	}
	userCmd.PersistentFlags().StringVar(&cloud.CloudProject, "cloud-project", "", "Cloud project ID")

	userListCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List users",
		Run:     cloud.ListCloudUsers,
	}
	userCmd.AddCommand(withFilterFlag(userListCmd))

	userCmd.AddCommand(&cobra.Command{
		Use:   "get <user_id>",
		Short: "Get information about a user",
		Run:   cloud.GetCloudUser,
		Args:  cobra.ExactArgs(1),
	})

	userCreateCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new user",
		Run:   cloud.CreateCloudUser,
	}
	userCreateCmd.Flags().StringVar(&cloud.UserSpec.Description, "description", "", "Description of the user")
	userCreateCmd.Flags().StringArrayVar(&cloud.UserSpec.Roles, "roles", nil, "Roles assigned to the user")
	addInitParameterFileFlag(userCreateCmd, assets.CloudOpenapiSchema, "/cloud/project/{serviceName}/user", "post", cloud.UserCreateExample, nil)
	addInteractiveEditorFlag(userCreateCmd)
	addFromFileFlag(userCreateCmd)
	userCreateCmd.MarkFlagsMutuallyExclusive("from-file", "editor")
	userCmd.AddCommand(userCreateCmd)

	userCmd.AddCommand(&cobra.Command{
		Use:   "delete <user_id>",
		Short: "Delete the given user",
		Run:   cloud.DeleteCloudUser,
		Args:  cobra.ExactArgs(1),
	})

	cloudCmd.AddCommand(userCmd)
}
