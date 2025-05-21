package telephony

import (
	_ "embed"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/flags"
	"stash.ovh.net/api/ovh-cli/internal/services/common"
)

var (
	telephonyColumnsToDisplay = []string{"billingAccount", "description", "status"}

	//go:embed templates/telephony.tmpl
	telephonyTemplate string
)

func ListTelephony(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/telephony", "", telephonyColumnsToDisplay, flags.GenericFilters)
}

func GetTelephony(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/telephony", args[0], telephonyTemplate)
}
