// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package xdsl

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
	xdslColumnsToDisplay = []string{"accessName", "accessType", "provider", "role", "status"}

	//go:embed templates/xdsl.tmpl
	xdslTemplate string

	XdslSpec struct {
		Description  string `json:"description,omitempty"`
		LnsRateLimit int    `json:"lnsRateLimit,omitempty"`
		Monitoring   bool   `json:"monitoring,omitempty"`
	}
)

func ListXdsl(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/v1/xdsl", "", xdslColumnsToDisplay, flags.GenericFilters)
}

func GetXdsl(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/v1/xdsl", args[0], xdslTemplate)
}

func EditXdsl(cmd *cobra.Command, args []string) {
	if err := common.EditResource(
		cmd,
		"/xdsl/{serviceName}",
		fmt.Sprintf("/v1/xdsl/%s", url.PathEscape(args[0])),
		XdslSpec,
		assets.XdslOpenapiSchema,
	); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}
}
