package cmd

import (
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
		Use:   "get <user>",
		Short: "Get information about a user",
		Run:   cloud.GetCloudUser,
		Args:  cobra.ExactArgs(1),
	})

	cloudCmd.AddCommand(userCmd)
}
