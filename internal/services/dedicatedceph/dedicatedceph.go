// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package dedicatedceph

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
	dedicatedcephColumnsToDisplay = []string{"serviceName", "region", "state", "status"}

	//go:embed templates/dedicatedceph.tmpl
	dedicatedcephTemplate string

	DedicatedCephSpec struct {
		CrushTunables string `json:"crushTunables,omitempty"`
		Label         string `json:"label,omitempty"`
	}
)

func ListDedicatedCeph(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/v1/dedicated/ceph", "", dedicatedcephColumnsToDisplay, flags.GenericFilters)
}

func GetDedicatedCeph(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/v1/dedicated/ceph", args[0], dedicatedcephTemplate)
}

func EditDedicatedCeph(cmd *cobra.Command, args []string) {
	if err := common.EditResource(
		cmd,
		"/dedicated/ceph/{serviceName}",
		fmt.Sprintf("/v1/dedicated/ceph/%s", url.PathEscape(args[0])),
		DedicatedCephSpec,
		assets.DedicatedcephOpenapiSchema,
	); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}
}
