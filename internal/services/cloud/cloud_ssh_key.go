package cloud

import (
	_ "embed"
	"fmt"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/display"
	filtersLib "stash.ovh.net/api/ovh-cli/internal/filters"
	"stash.ovh.net/api/ovh-cli/internal/flags"
	httpLib "stash.ovh.net/api/ovh-cli/internal/http"
	"stash.ovh.net/api/ovh-cli/internal/services/common"
)

var (
	cloudprojectSSHKeyColumnsToDisplay = []string{"id", "name", "regions"}

	//go:embed templates/cloud_ssh_key.tmpl
	cloudSSHKeyTemplate string
)

func ListCloudSSHKeys(_ *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}
	path := fmt.Sprintf("/cloud/project/%s/sshkey", projectID)

	var body []map[string]any
	if err := httpLib.Client.Get(path, &body); err != nil {
		display.ExitError("failed to fetch SSH keys: %s", err)
		return
	}

	body, err = filtersLib.FilterLines(body, flags.GenericFilters)
	if err != nil {
		display.ExitError("failed to filter results: %s", err)
		return
	}

	display.RenderTable(body, cloudprojectSSHKeyColumnsToDisplay, &flags.OutputFormatConfig)
}

func GetCloudSSHKey(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	common.ManageObjectRequest(fmt.Sprintf("/cloud/project/%s/sshkey", projectID), args[0], cloudSSHKeyTemplate)
}
