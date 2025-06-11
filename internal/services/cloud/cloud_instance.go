package cloud

import (
	_ "embed"
	"fmt"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/display"
	"stash.ovh.net/api/ovh-cli/internal/flags"
	"stash.ovh.net/api/ovh-cli/internal/services/common"
)

var (
	cloudprojectInstanceColumnsToDisplay = []string{"id", "name", "region", "flavor.name", "status"}

	//go:embed templates/cloud_instance.tmpl
	cloudInstanceTemplate string
)

func ListInstances(_ *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}
	common.ManageListRequest(fmt.Sprintf("/cloud/project/%s/instance", projectID), "id", cloudprojectInstanceColumnsToDisplay, flags.GenericFilters)
}

func GetInstance(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}
	common.ManageObjectRequest(fmt.Sprintf("/cloud/project/%s/instance", projectID), args[0], cloudInstanceTemplate)
}
