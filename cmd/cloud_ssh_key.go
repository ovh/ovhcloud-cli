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
	cloudprojectSSHKeyColumnsToDisplay = []string{"id", "name", "regions"}

	//go:embed templates/cloud_ssh_key.tmpl
	cloudSSHKeyTemplate string
)

func listCloudSSHKeys(_ *cobra.Command, _ []string) {
	projectID := url.PathEscape(getConfiguredCloudProject())
	path := fmt.Sprintf("/cloud/project/%s/sshkey", projectID)

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

	display.RenderTable(body, cloudprojectSSHKeyColumnsToDisplay, &outputFormatConfig)
}

func getCloudSSHKey(_ *cobra.Command, args []string) {
	projectID := url.PathEscape(getConfiguredCloudProject())
	manageObjectRequest(fmt.Sprintf("/cloud/project/%s/sshkey", projectID), args[0], cloudSSHKeyTemplate)
}

func initCloudSSHKeyCommand(cloudCmd *cobra.Command) {
	sshKeyCmd := &cobra.Command{
		Use:   "ssh-key",
		Short: "Manage SSH keys in the given cloud project",
	}
	sshKeyCmd.PersistentFlags().StringVar(&cloudProject, "cloud-project", "", "Cloud project ID")

	sshKeyListCmd := &cobra.Command{
		Use:   "list",
		Short: "List SSH keys",
		Run:   listCloudSSHKeys,
	}
	sshKeyCmd.AddCommand(withFilterFlag(sshKeyListCmd))

	sshKeyCmd.AddCommand(&cobra.Command{
		Use:   "get",
		Short: "Get information about a SSH key",
		Run:   getCloudSSHKey,
		Args:  cobra.ExactArgs(1),
	})

	cloudCmd.AddCommand(sshKeyCmd)
}
