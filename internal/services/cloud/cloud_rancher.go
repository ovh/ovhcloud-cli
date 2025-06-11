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
	cloudprojectRancherColumnsToDisplay = []string{"id", "currentState.name name", "currentState.region region", "currentState.version version", "resourceStatus"}

	//go:embed templates/cloud_rancher.tmpl
	cloudRancherTemplate string
)

func ListCloudRanchers(_ *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	common.ManageListRequestNoExpand(fmt.Sprintf("/v2/publicCloud/project/%s/rancher", projectID), cloudprojectRancherColumnsToDisplay, flags.GenericFilters)
}

func GetRancher(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	common.ManageObjectRequest(fmt.Sprintf("/v2/publicCloud/project/%s/rancher", projectID), args[0], cloudRancherTemplate)
}

func EditRancher(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	endpoint := fmt.Sprintf("/v2/publicCloud/project/%s/rancher/%s", projectID, url.PathEscape(args[0]))
	if err := editor.EditResource(httpLib.Client, "/publicCloud/project/{projectId}/rancher/{rancherId}", endpoint, cloudV2OpenapiSchema); err != nil {
		display.ExitError(err.Error())
	}
}
