package iploadbalancing

import (
	_ "embed"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/flags"
	"stash.ovh.net/api/ovh-cli/internal/services/common"
)

var (
	iploadbalancingColumnsToDisplay = []string{"serviceName", "displayName", "zone", "state"}

	//go:embed templates/iploadbalancing.tmpl
	iploadbalancingTemplate string
)

func ListIpLoadbalancing(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/ipLoadbalancing", "", iploadbalancingColumnsToDisplay, flags.GenericFilters)
}

func GetIpLoadbalancing(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/ipLoadbalancing", args[0], iploadbalancingTemplate)
}
