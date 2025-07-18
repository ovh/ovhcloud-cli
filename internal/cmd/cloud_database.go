package cmd

import (
	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/services/cloud"
)

func initCloudDatabaseCommand(cloudCmd *cobra.Command) {
	databaseCmd := &cobra.Command{
		Use:   "database",
		Short: "Manage databases in the given cloud project",
	}
	databaseCmd.PersistentFlags().StringVar(&cloud.CloudProject, "cloud-project", "", "Cloud project ID")

	databaseListCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List your databases",
		Run:     cloud.ListCloudDatabases,
	}
	databaseCmd.AddCommand(withFilterFlag(databaseListCmd))

	databaseCmd.AddCommand(&cobra.Command{
		Use:   "get <database_id>",
		Short: "Get a specific database",
		Run:   cloud.GetCloudDatabase,
		Args:  cobra.ExactArgs(1),
	})

	cloudCmd.AddCommand(databaseCmd)
}
