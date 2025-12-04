// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package hostingprivatedatabase

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
	hostingprivatedatabaseColumnsToDisplay = []string{"serviceName", "displayName", "type", "version", "state"}

	//go:embed templates/hostingprivatedatabase.tmpl
	hostingprivatedatabaseTemplate string

	// HostingPrivateDatabaseDisplayName is the display name of the HostingPrivateDatabase
	HostingPrivateDatabaseDisplayName string
)

func ListHostingPrivateDatabase(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/v1/hosting/privateDatabase", "", hostingprivatedatabaseColumnsToDisplay, flags.GenericFilters)
}

func GetHostingPrivateDatabase(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/v1/hosting/privateDatabase", args[0], hostingprivatedatabaseTemplate)
}

func EditHostingPrivateDatabase(cmd *cobra.Command, args []string) {
	if err := common.EditResource(
		cmd,
		"/hosting/privateDatabase/{serviceName}",
		fmt.Sprintf("/v1/hosting/privateDatabase/%s", url.PathEscape(args[0])),
		map[string]any{
			"displayName": HostingPrivateDatabaseDisplayName,
		},
		assets.HostingprivatedatabaseOpenapiSchema,
	); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}
}
