package sslgateway

import (
	_ "embed"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/flags"
	"stash.ovh.net/api/ovh-cli/internal/services/common"
)

var (
	sslgatewayColumnsToDisplay = []string{"serviceName", "displayName", "state", "zones"}

	//go:embed templates/sslgateway.tmpl
	sslgatewayTemplate string
)

func ListSslGateway(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/sslGateway", "", sslgatewayColumnsToDisplay, flags.GenericFilters)
}

func GetSslGateway(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/sslGateway", args[0], sslgatewayTemplate)
}
