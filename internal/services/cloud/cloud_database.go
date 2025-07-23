package cloud

import (
	_ "embed"
	"fmt"

	"github.com/ovh/ovhcloud-cli/internal/display"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	"github.com/ovh/ovhcloud-cli/internal/services/common"
	"github.com/spf13/cobra"
)

var (
	cloudprojectDatabaseColumnsToDisplay = []string{"id", "engine", "version", "description", "status"}

	//go:embed templates/cloud_database.tmpl
	cloudDatabaseTemplate string
)

func ListCloudDatabases(_ *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	common.ManageListRequest(fmt.Sprintf("/cloud/project/%s/database/service", projectID), "", cloudprojectDatabaseColumnsToDisplay, flags.GenericFilters)
}

func GetCloudDatabase(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	common.ManageObjectRequest(fmt.Sprintf("/cloud/project/%s/database/service", projectID), args[0], cloudDatabaseTemplate)
}
