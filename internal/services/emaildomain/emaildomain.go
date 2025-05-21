package emaildomain

import (
	_ "embed"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/flags"
	"stash.ovh.net/api/ovh-cli/internal/services/common"
)

var (
	emaildomainColumnsToDisplay = []string{"domain", "status", "offer"}

	//go:embed templates/emaildomain.tmpl
	emaildomainTemplate string
)

func ListEmailDomain(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/email/domain", "", emaildomainColumnsToDisplay, flags.GenericFilters)
}

func GetEmailDomain(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/email/domain", args[0], emaildomainTemplate)
}
