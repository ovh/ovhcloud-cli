package nutanix

import (
	_ "embed"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/flags"
	"stash.ovh.net/api/ovh-cli/internal/services/common"
)

var (
	nutanixColumnsToDisplay = []string{"serviceName", "status"}

	//go:embed templates/nutanix.tmpl
	nutanixTemplate string
)

func ListNutanix(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/nutanix", "", nutanixColumnsToDisplay, flags.GenericFilters)
}

func GetNutanix(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/nutanix", args[0], nutanixTemplate)
}
