// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package veeamcloudconnect

import (
	_ "embed"

	"github.com/ovh/ovhcloud-cli/internal/flags"
	"github.com/ovh/ovhcloud-cli/internal/services/common"
	"github.com/spf13/cobra"
)

var (
	veeamcloudconnectColumnsToDisplay = []string{"serviceName", "productOffer", "location", "vmCount"}

	//go:embed templates/veeamcloudconnect.tmpl
	veeamcloudconnectTemplate string
)

func ListVeeamCloudConnect(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/v1/veeamCloudConnect", "", veeamcloudconnectColumnsToDisplay, flags.GenericFilters)
}

func GetVeeamCloudConnect(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/v1/veeamCloudConnect", args[0], veeamcloudconnectTemplate)
}
