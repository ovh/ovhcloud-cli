package emailmxplan

import (
	_ "embed"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/flags"
	"stash.ovh.net/api/ovh-cli/internal/services/common"
)

var (
	emailmxplanColumnsToDisplay = []string{"domain", "displayName", "state", "offer"}

	//go:embed templates/emailmxplan.tmpl
	emailmxplanTemplate string
)

func ListEmailMXPlan(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/email/mxplan", "", emailmxplanColumnsToDisplay, flags.GenericFilters)
}

func GetEmailMXPlan(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/email/mxplan", args[0], emailmxplanTemplate)
}
