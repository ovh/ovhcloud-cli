package ssl

import (
	_ "embed"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/flags"
	"stash.ovh.net/api/ovh-cli/internal/services/common"
)

var (
	sslColumnsToDisplay = []string{"serviceName", "type", "authority", "status"}

	//go:embed templates/ssl.tmpl
	sslTemplate string
)

func ListSsl(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/ssl", "", sslColumnsToDisplay, flags.GenericFilters)
}

func GetSsl(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/ssl", args[0], sslTemplate)
}
