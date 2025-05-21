package veeamenterprise

import (
	_ "embed"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/flags"
	"stash.ovh.net/api/ovh-cli/internal/services/common"
)

var (
	veeamenterpriseColumnsToDisplay = []string{"serviceName", "activationStatus", "ip", "sourceIp"}

	//go:embed templates/veeamenterprise.tmpl
	veeamenterpriseTemplate string
)

func ListVeeamEnterprise(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/veeam/veeamEnterprise", "", veeamenterpriseColumnsToDisplay, flags.GenericFilters)
}

func GetVeeamEnterprise(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/veeam/veeamEnterprise", args[0], veeamenterpriseTemplate)
}
