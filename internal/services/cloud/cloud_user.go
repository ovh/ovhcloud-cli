package cloud

import (
	_ "embed"
	"fmt"

	"github.com/ovh/ovhcloud-cli/internal/display"
	filtersLib "github.com/ovh/ovhcloud-cli/internal/filters"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
	"github.com/ovh/ovhcloud-cli/internal/services/common"
	"github.com/spf13/cobra"
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
