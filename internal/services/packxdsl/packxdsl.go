// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package packxdsl

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
	packxdslColumnsToDisplay = []string{"packName", "description"}

	//go:embed templates/packxdsl.tmpl
	packxdslTemplate string

	PackXDSLSpec struct {
		Description string `json:"description,omitempty"`
	}
)

func ListPackXDSL(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/v1/pack/xdsl", "", packxdslColumnsToDisplay, flags.GenericFilters)
}

func GetPackXDSL(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/v1/pack/xdsl", args[0], packxdslTemplate)
}

func EditPackXDSL(cmd *cobra.Command, args []string) {
	if err := common.EditResource(
		cmd,
		"/pack/xdsl/{packName}",
		fmt.Sprintf("/v1/pack/xdsl/%s", url.PathEscape(args[0])),
		PackXDSLSpec,
		assets.PackxdslOpenapiSchema,
	); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}
}
