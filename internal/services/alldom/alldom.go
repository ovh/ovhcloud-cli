// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package alldom

import (
	_ "embed"

	"github.com/ovh/ovhcloud-cli/internal/flags"
	"github.com/ovh/ovhcloud-cli/internal/services/common"
	"github.com/spf13/cobra"
)

var (
	alldomColumnsToDisplay = []string{"name", "type", "offer"}

	//go:embed templates/alldom.tmpl
	alldomTemplate string
)

func ListAllDom(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/v1/allDom", "", alldomColumnsToDisplay, flags.GenericFilters)
}

func GetAllDom(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/v1/allDom", args[0], alldomTemplate)
}
