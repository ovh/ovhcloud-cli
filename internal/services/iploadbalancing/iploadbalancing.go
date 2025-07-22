package iploadbalancing

import (
	_ "embed"
	"fmt"
	"net/url"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/assets"
	"stash.ovh.net/api/ovh-cli/internal/display"
	"stash.ovh.net/api/ovh-cli/internal/flags"
	"stash.ovh.net/api/ovh-cli/internal/services/common"
)

var (
	iploadbalancingColumnsToDisplay = []string{"serviceName", "displayName", "zone", "state"}

	//go:embed templates/iploadbalancing.tmpl
	iploadbalancingTemplate string

	IPLoadbalancingSpec struct {
		DisplayName      string `json:"displayName,omitempty"`
		SSLConfiguration string `json:"sslConfiguration,omitempty"`
	}
)

func ListIpLoadbalancing(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/ipLoadbalancing", "", iploadbalancingColumnsToDisplay, flags.GenericFilters)
}

func GetIpLoadbalancing(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/ipLoadbalancing", args[0], iploadbalancingTemplate)
}

func EditIpLoadbalancing(cmd *cobra.Command, args []string) {
	if err := common.EditResource(
		cmd,
		"/ipLoadbalancing/{serviceName}",
		fmt.Sprintf("/ipLoadbalancing/%s", url.PathEscape(args[0])),
		IPLoadbalancingSpec,
		assets.IploadbalancingOpenapiSchema,
	); err != nil {
		display.ExitError(err.Error())
		return
	}
}
