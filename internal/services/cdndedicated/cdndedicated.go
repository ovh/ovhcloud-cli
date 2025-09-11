// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cdndedicated

import (
	_ "embed"

	"github.com/ovh/ovhcloud-cli/internal/flags"
	"github.com/ovh/ovhcloud-cli/internal/services/common"
	"github.com/spf13/cobra"
)

var (
	cdndedicatedColumnsToDisplay = []string{"service", "offer", "anycast"}

	//go:embed templates/cdndedicated.tmpl
	cdndedicatedTemplate string
)

func ListCdnDedicated(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/cdn/dedicated", "", cdndedicatedColumnsToDisplay, flags.GenericFilters)
}

func GetCdnDedicated(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/cdn/dedicated", args[0], cdndedicatedTemplate)
}
