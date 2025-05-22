package iploadbalancing

import (
	_ "embed"
	"fmt"
	"net/url"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/editor"
	"stash.ovh.net/api/ovh-cli/internal/flags"
	httpLib "stash.ovh.net/api/ovh-cli/internal/http"
	"stash.ovh.net/api/ovh-cli/internal/services/common"
)

var (
	iploadbalancingColumnsToDisplay = []string{"serviceName", "displayName", "zone", "state"}

	//go:embed templates/iploadbalancing.tmpl
	iploadbalancingTemplate string

	//go:embed api-schemas/iploadbalancing.json
	iploadbalancingOpenapiSchema []byte
)

func ListIpLoadbalancing(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/ipLoadbalancing", "", iploadbalancingColumnsToDisplay, flags.GenericFilters)
}

func GetIpLoadbalancing(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/ipLoadbalancing", args[0], iploadbalancingTemplate)
}

func EditIpLoadbalancing(_ *cobra.Command, args []string) {
	url := fmt.Sprintf("/ipLoadbalancing/%s", url.PathEscape(args[0]))
	editor.EditResource(httpLib.Client, "/ipLoadbalancing/{serviceName}", url, iploadbalancingOpenapiSchema)
}
