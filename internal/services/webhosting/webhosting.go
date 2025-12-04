// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package webhosting

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
	webhostingColumnsToDisplay = []string{"serviceName", "displayName", "datacenter", "state"}

	//go:embed templates/webhosting.tmpl
	webhostingTemplate string

	WebHostingSpec struct {
		DisplayName string `json:"displayName,omitempty"`
	}
)

func ListWebHosting(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/v1/hosting/web", "", webhostingColumnsToDisplay, flags.GenericFilters)
}

func GetWebHosting(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/v1/hosting/web", args[0], webhostingTemplate)
}

func EditWebHosting(cmd *cobra.Command, args []string) {
	if err := common.EditResource(
		cmd,
		"/hosting/web/{serviceName}",
		fmt.Sprintf("/v1/hosting/web/%s", url.PathEscape(args[0])),
		WebHostingSpec,
		assets.WebhostingOpenapiSchema,
	); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}
}
