package okms

import (
	_ "embed"

	"github.com/ovh/ovhcloud-cli/internal/flags"
	"github.com/ovh/ovhcloud-cli/internal/services/common"
	"github.com/spf13/cobra"
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
