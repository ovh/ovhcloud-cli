package okms

import (
	_ "embed"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/flags"
	"stash.ovh.net/api/ovh-cli/internal/services/common"
)

var (
	okmsColumnsToDisplay = []string{"id", "region"}

	//go:embed templates/okms.tmpl
	okmsTemplate string
)

func ListOkms(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/v2/okms/resource", "id", okmsColumnsToDisplay, flags.GenericFilters)
}

func GetOkms(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/v2/okms/resource", args[0], okmsTemplate)
}
