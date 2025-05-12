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
	cloudprojectUserColumnsToDisplay = []string{"id", "username", "description", "status"}

	//go:embed templates/cloud_user.tmpl
	cloudUserTemplate string
)

func listCloudUsers(_ *cobra.Command, _ []string) {
	projectID := url.PathEscape(getConfiguredCloudProject())
	path := fmt.Sprintf("/cloud/project/%s/user", projectID)

	var (
		body []map[string]any
		err  error
	)
	if err := client.Get(path, &body); err != nil {
		display.ExitError("failed to fetch SSH keys: %s", err)
	}

	body, err = filtersLib.FilterLines(body, genericFilters)
	if err != nil {
		display.ExitError("failed to filter results: %s", err)
	}

	display.RenderTable(body, cloudprojectUserColumnsToDisplay, &outputFormatConfig)
}

func getCloudUser(_ *cobra.Command, args []string) {
	projectID := url.PathEscape(getConfiguredCloudProject())
	manageObjectRequest(fmt.Sprintf("/cloud/project/%s/user", projectID), args[0], cloudUserTemplate)
}

func initCloudUserCommand(cloudCmd *cobra.Command) {
	userCmd := &cobra.Command{
		Use:   "user",
		Short: "Manage users in the given cloud project",
	}
	userCmd.PersistentFlags().StringVar(&cloudProject, "cloud-project", "", "Cloud project ID")

	userListCmd := &cobra.Command{
		Use:   "list",
		Short: "List users",
		Run:   listCloudUsers,
	}
	userCmd.AddCommand(withFilterFlag(userListCmd))

	userCmd.AddCommand(&cobra.Command{
		Use:   "get",
		Short: "Get information about a user",
		Run:   getCloudUser,
		Args:  cobra.ExactArgs(1),
	})

	cloudCmd.AddCommand(userCmd)
}
