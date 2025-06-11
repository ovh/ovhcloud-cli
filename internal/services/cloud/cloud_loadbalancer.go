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
	cloudprojectLoadbalancerColumnsToDisplay = []string{"id", "openstackRegion", "size", "status"}

	//go:embed templates/cloud_loadbalancer.tmpl
	cloudLoadbalancerTemplate string
)

func ListCloudLoadbalancers(_ *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	common.ManageListRequest(fmt.Sprintf("/cloud/project/%s/loadbalancer", projectID), "", cloudprojectLoadbalancerColumnsToDisplay, flags.GenericFilters)
}

func GetCloudLoadbalancer(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	common.ManageObjectRequest(fmt.Sprintf("/cloud/project/%s/loadbalancer", projectID), args[0], cloudLoadbalancerTemplate)
}

func EditCloudLoadbalancer(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	endpoint := fmt.Sprintf("/cloud/project/%s/loadbalancer/%s", projectID, url.PathEscape(args[0]))
	if err := editor.EditResource(httpLib.Client, "/cloud/project/{serviceName}/loadbalancer/{loadBalancerId}", endpoint, cloudOpenapiSchema); err != nil {
		display.ExitError(err.Error())
	}
}
