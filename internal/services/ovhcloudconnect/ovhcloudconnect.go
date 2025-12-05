// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package ovhcloudconnect

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
	ovhcloudconnectColumnsToDisplay = []string{"uuid", "provider", "status", "description"}

	//go:embed templates/ovhcloudconnect.tmpl
	ovhcloudconnectTemplate string

	OvhCloudConnectSpec struct {
		Description string `json:"description,omitempty"`
	}
)

func ListOvhCloudConnect(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/v1/ovhCloudConnect", "", ovhcloudconnectColumnsToDisplay, flags.GenericFilters)
}

func GetOvhCloudConnect(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/v1/ovhCloudConnect", args[0], ovhcloudconnectTemplate)
}

func EditOvhCloudConnect(cmd *cobra.Command, args []string) {
	if err := common.EditResource(
		cmd,
		"/ovhCloudConnect/{serviceName}",
		fmt.Sprintf("/v1/ovhCloudConnect/%s", url.PathEscape(args[0])),
		OvhCloudConnectSpec,
		assets.OvhcloudconnectOpenapiSchema,
	); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}
}
