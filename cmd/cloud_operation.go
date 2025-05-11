package cmd

import (
	_ "embed"
	"fmt"
	"net/url"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/display"
	filtersLib "stash.ovh.net/api/ovh-cli/internal/filters"
)

var (
	cloudprojectOperationColumnsToDisplay = []string{"id", "action", "progress", "status"}

	//go:embed templates/cloud_operation.tmpl
	cloudOperationTemplate string
)

func listCloudOperations(_ *cobra.Command, _ []string) {
	projectID := url.PathEscape(getConfiguredCloudProject())

	var operations []map[string]any
	err := client.Get(fmt.Sprintf("/cloud/project/%s/operation", projectID), &operations)
	if err != nil {
		display.ExitError("failed to fetch results: %s", err)
	}

	operations, err = filtersLib.FilterLines(operations, genericFilters)
	if err != nil {
		display.ExitError("failed to filter results: %s", err)
	}

	display.RenderTable(operations, cloudprojectOperationColumnsToDisplay, &outputFormatConfig)
}

func getCloudOperation(_ *cobra.Command, args []string) {
	projectID := url.PathEscape(getConfiguredCloudProject())
	manageObjectRequest(fmt.Sprintf("/cloud/project/%s/operation", projectID), args[0], cloudOperationTemplate)
}

func initCloudOperationCommand(cloudCmd *cobra.Command) {
	operationCmd := &cobra.Command{
		Use:   "operation",
		Short: "List and get operations in the given cloud project",
	}
	operationCmd.PersistentFlags().StringVar(&cloudProject, "cloud-project", "", "Cloud project ID")

	operationListCmd := &cobra.Command{
		Use:   "list",
		Short: "List operations of the given project",
		Run:   listCloudOperations,
	}
	operationCmd.AddCommand(withFilterFlag(operationListCmd))

	operationCmd.AddCommand(&cobra.Command{
		Use:        "get",
		Short:      "Get a specific operation",
		Run:        getCloudOperation,
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"operation_id"},
	})

	cloudCmd.AddCommand(operationCmd)
}
