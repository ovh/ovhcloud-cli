// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package dedicatednasha

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
	dedicatednashaColumnsToDisplay = []string{"serviceName", "customName", "datacenter"}

	//go:embed templates/dedicatednasha.tmpl
	dedicatednashaTemplate string

	DedicatedNasHASpec struct {
		CustomName string `json:"customName,omitempty"`
		Monitored  bool   `json:"monitored,omitempty"`
	}
)

func ListDedicatedNasHA(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/v1/dedicated/nasha", "", dedicatednashaColumnsToDisplay, flags.GenericFilters)
}

func GetDedicatedNasHA(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/v1/dedicated/nasha", args[0], dedicatednashaTemplate)
}

func EditDedicatedNasHA(cmd *cobra.Command, args []string) {
	if err := common.EditResource(
		cmd,
		"/dedicated/nasha/{serviceName}",
		fmt.Sprintf("/v1/dedicated/nasha/%s", url.PathEscape(args[0])),
		DedicatedNasHASpec,
		assets.DedicatednashaOpenapiSchema,
	); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}
}
