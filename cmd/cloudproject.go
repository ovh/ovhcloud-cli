package cmd

import (
	"github.com/spf13/cobra"
)

var (
	cloudprojectColumnsToDisplay = []string{"project_id", "projectName", "status", "description"}
)

func listCloudProject(_ *cobra.Command, _ []string) {
	manageListRequest("/cloud/project", cloudprojectColumnsToDisplay)
}

func getCloudProject(_ *cobra.Command, args []string) {
	manageObjectRequest("/cloud/project", args[0], cloudprojectColumnsToDisplay[0])
}

func init() {
	cloudCmd := &cobra.Command{
		Use:   "cloud",
		Short: "Manage your projects and services in the Public Cloud universe",
	}

	cloudprojectCmd := &cobra.Command{
		Use:   "project",
		Short: "Retrieve information and manage your CloudProject services",
	}

	// Command to list CloudProject services
	cloudprojectCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List your cloud projects",
		Run:   listCloudProject,
	})

	// Command to get a single CloudProject
	cloudprojectCmd.AddCommand(&cobra.Command{
		Use:        "get",
		Short:      "Retrieve information of a specific cloud project",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"service_name"},
		Run:        getCloudProject,
	})

	initKubeCommand(cloudCmd)

	cloudCmd.AddCommand(cloudprojectCmd)
	rootCmd.AddCommand(cloudCmd)
}
