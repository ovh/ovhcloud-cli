// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package iploadbalancing

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
	iploadbalancingColumnsToDisplay = []string{"serviceName", "displayName", "zone", "state"}

	//go:embed templates/iploadbalancing.tmpl
	iploadbalancingTemplate string

	IPLoadbalancingSpec struct {
		DisplayName      string `json:"displayName,omitempty"`
		SSLConfiguration string `json:"sslConfiguration,omitempty"`
	}
)

func ListIpLoadbalancing(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/v1/ipLoadbalancing", "", iploadbalancingColumnsToDisplay, flags.GenericFilters)
}

func GetIpLoadbalancing(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/v1/ipLoadbalancing", args[0], iploadbalancingTemplate)
}

func EditIpLoadbalancing(cmd *cobra.Command, args []string) {
	if err := common.EditResource(
		cmd,
		"/ipLoadbalancing/{serviceName}",
		fmt.Sprintf("/v1/ipLoadbalancing/%s", url.PathEscape(args[0])),
		IPLoadbalancingSpec,
		assets.IploadbalancingOpenapiSchema,
	); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}
}
