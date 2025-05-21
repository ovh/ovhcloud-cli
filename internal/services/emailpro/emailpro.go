package emailpro

import (
	_ "embed"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/flags"
	"stash.ovh.net/api/ovh-cli/internal/services/common"
)

var (
	emailproColumnsToDisplay = []string{"domain", "displayName", "state", "offer"}

	//go:embed templates/emailpro.tmpl
	emailproTemplate string
)

func ListEmailPro(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/email/pro", "", emailproColumnsToDisplay, flags.GenericFilters)
}

func GetEmailPro(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/email/pro", args[0], emailproTemplate)
}
