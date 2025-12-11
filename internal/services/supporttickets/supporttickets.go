// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package supporttickets

import (
	_ "embed"

	"github.com/ovh/ovhcloud-cli/internal/flags"
	"github.com/ovh/ovhcloud-cli/internal/services/common"
	"github.com/spf13/cobra"
)

var (
	supportticketsColumnsToDisplay = []string{"ticketId", "serviceName", "type", "category", "state"}

	//go:embed templates/supporttickets.tmpl
	supportticketsTemplate string
)

func ListSupportTickets(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/v1/support/tickets", "", supportticketsColumnsToDisplay, flags.GenericFilters)
}

func GetSupportTickets(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/v1/support/tickets", args[0], supportticketsTemplate)
}
