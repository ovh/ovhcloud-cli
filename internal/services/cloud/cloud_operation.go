package cloud

import (
	_ "embed"
	"fmt"
	"net/url"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/display"
	filtersLib "stash.ovh.net/api/ovh-cli/internal/filters"
	"stash.ovh.net/api/ovh-cli/internal/flags"
	httpLib "stash.ovh.net/api/ovh-cli/internal/http"
	"stash.ovh.net/api/ovh-cli/internal/services/common"
)

var (
	cloudprojectOperationColumnsToDisplay = []string{"id", "action", "progress", "status"}

	//go:embed templates/cloud_operation.tmpl
	cloudOperationTemplate string
)

func ListCloudOperations(_ *cobra.Command, _ []string) {
	projectID := url.PathEscape(getConfiguredCloudProject())

	var operations []map[string]any
	err := httpLib.Client.Get(fmt.Sprintf("/cloud/project/%s/operation", projectID), &operations)
	if err != nil {
		display.ExitError("failed to fetch results: %s", err)
	}

	operations, err = filtersLib.FilterLines(operations, flags.GenericFilters)
	if err != nil {
		display.ExitError("failed to filter results: %s", err)
	}

	display.RenderTable(operations, cloudprojectOperationColumnsToDisplay, &flags.OutputFormatConfig)
}

func GetCloudOperation(_ *cobra.Command, args []string) {
	projectID := url.PathEscape(getConfiguredCloudProject())
	common.ManageObjectRequest(fmt.Sprintf("/cloud/project/%s/operation", projectID), args[0], cloudOperationTemplate)
}
