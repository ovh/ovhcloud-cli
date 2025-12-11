// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package nutanix

import (
	_ "embed"

	"github.com/ovh/ovhcloud-cli/internal/flags"
	"github.com/ovh/ovhcloud-cli/internal/services/common"
	"github.com/spf13/cobra"
)

var (
	nutanixColumnsToDisplay = []string{"serviceName", "status"}

	//go:embed templates/nutanix.tmpl
	nutanixTemplate string
)

func ListNutanix(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/v1/nutanix", "", nutanixColumnsToDisplay, flags.GenericFilters)
}

func GetNutanix(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/v1/nutanix", args[0], nutanixTemplate)
}
