package cloud

import (
	_ "embed"
	"fmt"
	"net/url"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/display"
	"stash.ovh.net/api/ovh-cli/internal/editor"
	"stash.ovh.net/api/ovh-cli/internal/flags"
	httpLib "stash.ovh.net/api/ovh-cli/internal/http"
	"stash.ovh.net/api/ovh-cli/internal/services/common"
)

var (
	cloudprojectVolumeColumnsToDisplay = []string{"id", "name", "region", "type", "status"}

	//go:embed templates/cloud_volume.tmpl
	cloudVolumeTemplate string
)

func ListCloudVolumes(_ *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	common.ManageListRequestNoExpand(fmt.Sprintf("/cloud/project/%s/volume", projectID), cloudprojectVolumeColumnsToDisplay, flags.GenericFilters)
}

func GetVolume(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	common.ManageObjectRequest(fmt.Sprintf("/cloud/project/%s/volume", projectID), args[0], cloudVolumeTemplate)
}

func EditVolume(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	endpoint := fmt.Sprintf("/cloud/project/%s/volume/%s", projectID, url.PathEscape(args[0]))
	if err := editor.EditResource(httpLib.Client, "/cloud/project/{serviceName}/volume/{volumeId}", endpoint, cloudOpenapiSchema); err != nil {
		display.ExitError(err.Error())
	}
}
