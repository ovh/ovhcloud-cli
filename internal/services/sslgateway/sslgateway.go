// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package sslgateway

import (
	_ "embed"
	"fmt"
	"net/url"

	"github.com/ovh/ovhcloud-cli/internal/assets"
	"github.com/ovh/ovhcloud-cli/internal/display"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	"github.com/ovh/ovhcloud-cli/internal/services/common"
	"github.com/spf13/cobra"
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
	common.ManageListRequest("/v1/sslGateway", "", sslgatewayColumnsToDisplay, flags.GenericFilters)
}

func GetSslGateway(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/v1/sslGateway", args[0], sslgatewayTemplate)
}

func EditSslGateway(cmd *cobra.Command, args []string) {
	if err := common.EditResource(
		cmd,
		"/sslGateway/{serviceName}",
		fmt.Sprintf("/v1/sslGateway/%s", url.PathEscape(args[0])),
		SSLGatewaySpec,
		assets.SslgatewayOpenapiSchema,
	); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}
}
