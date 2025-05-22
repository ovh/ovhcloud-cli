package sslgateway

import (
	_ "embed"
	"fmt"
	"net/url"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/editor"
	"stash.ovh.net/api/ovh-cli/internal/flags"
	httpLib "stash.ovh.net/api/ovh-cli/internal/http"
	"stash.ovh.net/api/ovh-cli/internal/services/common"
)

var (
	sslgatewayColumnsToDisplay = []string{"serviceName", "displayName", "state", "zones"}

	//go:embed templates/sslgateway.tmpl
	sslgatewayTemplate string

	//go:embed api-schemas/sslgateway.json
	sslgatewayOpenapiSchema []byte
)

func ListSslGateway(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/sslGateway", "", sslgatewayColumnsToDisplay, flags.GenericFilters)
}

func GetSslGateway(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/sslGateway", args[0], sslgatewayTemplate)
}

func EditSslGateway(_ *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("/sslGateway/%s", url.PathEscape(args[0]))
	editor.EditResource(httpLib.Client, "/sslGateway/{serviceName}", endpoint, sslgatewayOpenapiSchema)
}
