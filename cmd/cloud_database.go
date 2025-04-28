package cmd

import (
	_ "embed"
	"fmt"
	"net/url"

	"github.com/spf13/cobra"
)

var (
	cloudprojectDatabaseColumnsToDisplay = []string{"id", "engine", "version", "description", "status"}

	//go:embed templates/cloud_database.tmpl
	cloudDatabaseTemplate string
)

func listCloudDatabases(_ *cobra.Command, _ []string) {
	projectID := url.PathEscape(getConfiguredCloudProject())
	manageListRequest(fmt.Sprintf("/cloud/project/%s/database/service", projectID), cloudprojectDatabaseColumnsToDisplay, genericFilters)
}

func getCloudDatabase(_ *cobra.Command, args []string) {
	projectID := url.PathEscape(getConfiguredCloudProject())
	manageObjectRequest(fmt.Sprintf("/cloud/project/%s/database/service", projectID), args[0], cloudDatabaseTemplate)
}

func initCloudDatabaseCommand(cloudCmd *cobra.Command) {
	databaseCmd := &cobra.Command{
		Use:   "database",
		Short: "Manage databases in the given cloud project",
	}
	databaseCmd.PersistentFlags().StringVar(&cloudProject, "cloud-project", "", "Cloud project ID")

	databaseListCmd := &cobra.Command{
		Use:   "list",
		Short: "List your databases",
		Run:   listCloudDatabases,
	}
	databaseCmd.AddCommand(databaseListCmd)

	databaseCmd.AddCommand(&cobra.Command{
		Use:        "get",
		Short:      "Get a specific database",
		Run:        getCloudDatabase,
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"database_id"},
	})

	cloudCmd.AddCommand(databaseCmd)
}
