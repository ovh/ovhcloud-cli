package cmd

import (
	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/services/cloud"
)

func initCloudUserCommand(cloudCmd *cobra.Command) {
	userCmd := &cobra.Command{
		Use:   "user",
		Short: "Manage users in the given cloud project",
	}
	userCmd.PersistentFlags().StringVar(&cloud.CloudProject, "cloud-project", "", "Cloud project ID")

	userListCmd := &cobra.Command{
		Use:   "list",
		Short: "List users",
		Run:   cloud.ListCloudUsers,
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
