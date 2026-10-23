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
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
	"github.com/ovh/ovhcloud-cli/internal/services/common"
	"github.com/spf13/cobra"
)

var (
	iploadbalancingColumnsToDisplay = []string{"serviceName", "displayName", "zone", "state"}

	iploadbalancingPendingChangesColumnsToDisplay = []string{"zone", "number"}

	//go:embed templates/iploadbalancing.tmpl
	iploadbalancingTemplate string

	// IPLoadbalancingRefreshZone is the zone to apply the configuration to (all zones when empty)
	IPLoadbalancingRefreshZone string

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

// ListIpLoadbalancingPendingChanges lists the pending changes of the given IpLoadbalancing service, per zone
func ListIpLoadbalancingPendingChanges(_ *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("/v1/ipLoadbalancing/%s/pendingChanges", url.PathEscape(args[0]))
	common.ManageListRequestNoExpand(endpoint, iploadbalancingPendingChangesColumnsToDisplay, flags.GenericFilters)
}

// RefreshIpLoadbalancing applies the pending configuration changes of the given IpLoadbalancing service
func RefreshIpLoadbalancing(_ *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("/v1/ipLoadbalancing/%s/refresh", url.PathEscape(args[0]))

	body := map[string]any{}
	if IPLoadbalancingRefreshZone != "" {
		body["zone"] = IPLoadbalancingRefreshZone
	}

	var task map[string]any
	if err := httpLib.Client.Post(endpoint, body, &task); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to refresh IpLoadbalancing %s: %s", args[0], err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, task, "✅ Configuration refresh started for IpLoadbalancing %s (task ID: %v)", args[0], task["id"])
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
