package veeamcloudconnect

import (
	_ "embed"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/flags"
	"stash.ovh.net/api/ovh-cli/internal/services/common"
)

var (
	veeamcloudconnectColumnsToDisplay = []string{"serviceName", "productOffer", "location", "vmCount"}

	//go:embed templates/veeamcloudconnect.tmpl
	veeamcloudconnectTemplate string
)

func ListVeeamCloudConnect(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/veeamCloudConnect", "", veeamcloudconnectColumnsToDisplay, flags.GenericFilters)
}

func GetVeeamCloudConnect(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/veeamCloudConnect", args[0], veeamcloudconnectTemplate)
}
