// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package veeamenterprise

import (
	_ "embed"

	"github.com/ovh/ovhcloud-cli/internal/flags"
	"github.com/ovh/ovhcloud-cli/internal/services/common"
	"github.com/spf13/cobra"
)

var (
	veeamenterpriseColumnsToDisplay = []string{"serviceName", "activationStatus", "ip", "sourceIp"}

	//go:embed templates/veeamenterprise.tmpl
	veeamenterpriseTemplate string
)

func ListVeeamEnterprise(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/v1/veeam/veeamEnterprise", "", veeamenterpriseColumnsToDisplay, flags.GenericFilters)
}

func GetVeeamEnterprise(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/v1/veeam/veeamEnterprise", args[0], veeamenterpriseTemplate)
}
