// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package location

import (
	_ "embed"

	"github.com/ovh/ovhcloud-cli/internal/flags"
	"github.com/ovh/ovhcloud-cli/internal/services/common"
	"github.com/spf13/cobra"
)

var (
	locationColumnsToDisplay = []string{"name", "type", "specificType", "location"}

	//go:embed templates/location.tmpl
	locationTemplate string
)

func ListLocation(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/v2/location", "name", locationColumnsToDisplay, flags.GenericFilters)
}

func GetLocation(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/v2/location", args[0], locationTemplate)
}
