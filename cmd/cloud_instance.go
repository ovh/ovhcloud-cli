package cmd

import (
	_ "embed"
	"fmt"
	"net/url"

	"github.com/spf13/cobra"
)

var (
	cloudprojectInstanceColumnsToDisplay = []string{"id", "name", "region", "flavor.name", "status"}

	//go:embed templates/cloud_instance.tmpl
	cloudInstanceTemplate string
)

func listInstances(_ *cobra.Command, _ []string) {
	projectID := url.PathEscape(getConfiguredCloudProject())
	manageListRequest(fmt.Sprintf("/cloud/project/%s/instance", projectID), "id", cloudprojectInstanceColumnsToDisplay, genericFilters)
}

func getInstance(_ *cobra.Command, args []string) {
	projectID := url.PathEscape(getConfiguredCloudProject())
	manageObjectRequest(fmt.Sprintf("/cloud/project/%s/instance", projectID), args[0], cloudInstanceTemplate)
}

func initInstanceCommand(cloudCmd *cobra.Command) {
	instanceCmd := &cobra.Command{
		Use:   "instance",
		Short: "Manage instances in the given cloud project",
	}
	instanceCmd.PersistentFlags().StringVar(&cloudProject, "cloud-project", "", "Cloud project ID")

	instanceListCmd := &cobra.Command{
		Use:   "list",
		Short: "List your instances",
		Run:   listInstances,
	}
	instanceListCmd.PersistentFlags().StringArrayVar(
		&genericFilters,
		"filter",
		nil,
		"Filter results by any property using github.com/PaesslerAG/gval syntax'",
	)
	instanceCmd.AddCommand(instanceListCmd)

	instanceCmd.AddCommand(&cobra.Command{
		Use:        "get",
		Short:      "Get a specific instance",
		Run:        getInstance,
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"instance_id"},
	})

	cloudCmd.AddCommand(instanceCmd)
}
