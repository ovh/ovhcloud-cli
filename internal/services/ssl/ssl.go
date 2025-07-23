package ssl

import (
	_ "embed"

	"github.com/ovh/ovhcloud-cli/internal/flags"
	"github.com/ovh/ovhcloud-cli/internal/services/common"
	"github.com/spf13/cobra"
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
