package sslgateway

import (
	_ "embed"
	"fmt"
	"net/url"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/assets"
	"stash.ovh.net/api/ovh-cli/internal/display"
	"stash.ovh.net/api/ovh-cli/internal/flags"
	"stash.ovh.net/api/ovh-cli/internal/services/common"
)

var (
	sslgatewayColumnsToDisplay = []string{"serviceName", "displayName", "state", "zones"}

	//go:embed templates/sslgateway.tmpl
	sslgatewayTemplate string

	SSLGatewaySpec struct {
		AllowedSource    []string `json:"allowedSource,omitempty"`
		DisplayName      string   `json:"displayName,omitempty"`
		Hsts             bool     `json:"hsts,omitempty"`
		HttpsRedirect    bool     `json:"httpsRedirect,omitempty"`
		Reverse          string   `json:"reverse,omitempty"`
		ServerHttps      bool     `json:"serverHttps,omitempty"`
		SslConfiguration string   `json:"sslConfiguration,omitempty"`
	}
)

func ListSslGateway(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/sslGateway", "", sslgatewayColumnsToDisplay, flags.GenericFilters)
}

func GetSslGateway(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/sslGateway", args[0], sslgatewayTemplate)
}

func EditSslGateway(cmd *cobra.Command, args []string) {
	if err := common.EditResource(
		cmd,
		"/sslGateway/{serviceName}",
		fmt.Sprintf("/sslGateway/%s", url.PathEscape(args[0])),
		SSLGatewaySpec,
		assets.SslgatewayOpenapiSchema,
	); err != nil {
		display.ExitError(err.Error())
		return
	}
}
