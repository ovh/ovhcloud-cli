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
	cloudprojectUserColumnsToDisplay = []string{"id", "username", "description", "status"}

	//go:embed templates/cloud_user.tmpl
	cloudUserTemplate string
)

func ListCloudUsers(_ *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}
	path := fmt.Sprintf("/cloud/project/%s/user", projectID)

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

	display.RenderTable(body, cloudprojectUserColumnsToDisplay, &flags.OutputFormatConfig)
}

func GetCloudUser(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	common.ManageObjectRequest(fmt.Sprintf("/cloud/project/%s/user", projectID), args[0], cloudUserTemplate)
}
