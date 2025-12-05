// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package overthebox

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
	overtheboxColumnsToDisplay = []string{"serviceName", "offer", "status", "bandwidth"}

	//go:embed templates/overthebox.tmpl
	overtheboxTemplate string

	OverTheBoxSpec struct {
		AutoUpgrade         bool   `json:"autoUpgrade,omitempty"`
		CustomerDescription string `json:"customerDescription,omitempty"`
		ReleaseChannel      string `json:"releaseChannel,omitempty"`
	}
)

func ListOverTheBox(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/v1/overTheBox", "", overtheboxColumnsToDisplay, flags.GenericFilters)
}

func GetOverTheBox(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/v1/overTheBox", args[0], overtheboxTemplate)
}

func EditOverTheBox(cmd *cobra.Command, args []string) {
	if err := common.EditResource(
		cmd,
		"/overTheBox/{serviceName}",
		fmt.Sprintf("/v1/overTheBox/%s", url.PathEscape(args[0])),
		OverTheBoxSpec,
		assets.OvertheboxOpenapiSchema,
	); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}
}
