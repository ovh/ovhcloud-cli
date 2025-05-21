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
	cloudprojectLoadbalancerColumnsToDisplay = []string{"id", "openstackRegion", "size", "status"}

	//go:embed templates/cloud_loadbalancer.tmpl
	cloudLoadbalancerTemplate string
)

func ListCloudLoadbalancers(_ *cobra.Command, _ []string) {
	projectID := url.PathEscape(getConfiguredCloudProject())
	common.ManageListRequest(fmt.Sprintf("/cloud/project/%s/loadbalancer", projectID), "", cloudprojectLoadbalancerColumnsToDisplay, flags.GenericFilters)
}

func GetCloudLoadbalancer(_ *cobra.Command, args []string) {
	projectID := url.PathEscape(getConfiguredCloudProject())
	common.ManageObjectRequest(fmt.Sprintf("/cloud/project/%s/loadbalancer", projectID), args[0], cloudLoadbalancerTemplate)
}
