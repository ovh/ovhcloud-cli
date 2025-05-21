package cloud

import (
	_ "embed"
	"fmt"
	"net/url"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/flags"
	"stash.ovh.net/api/ovh-cli/internal/services/common"
)

var (
	cloudprojectDatabaseColumnsToDisplay = []string{"id", "engine", "version", "description", "status"}

	//go:embed templates/cloud_database.tmpl
	cloudDatabaseTemplate string
)

func ListCloudDatabases(_ *cobra.Command, _ []string) {
	projectID := url.PathEscape(getConfiguredCloudProject())
	common.ManageListRequest(fmt.Sprintf("/cloud/project/%s/database/service", projectID), "", cloudprojectDatabaseColumnsToDisplay, flags.GenericFilters)
}

func GetCloudDatabase(_ *cobra.Command, args []string) {
	projectID := url.PathEscape(getConfiguredCloudProject())
	common.ManageObjectRequest(fmt.Sprintf("/cloud/project/%s/database/service", projectID), args[0], cloudDatabaseTemplate)
}
