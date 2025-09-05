// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/ovh/ovhcloud-cli/internal/services/supporttickets"
	"github.com/spf13/cobra"
)

func init() {
	supportticketsCmd := &cobra.Command{
		Use:   "support-tickets",
		Short: "Retrieve information and manage your support tickets",
	}

	// Command to list tickets
	supportticketsListCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List your support tickets",
		Run:     supporttickets.ListSupportTickets,
	}
	supportticketsCmd.AddCommand(withFilterFlag(supportticketsListCmd))

	// Command to get a single support ticket
	supportticketsCmd.AddCommand(&cobra.Command{
		Use:   "get <ticket_id>",
		Short: "Retrieve information of a specific support ticket",
		Args:  cobra.ExactArgs(1),
		Run:   supporttickets.GetSupportTickets,
	})

	rootCmd.AddCommand(supportticketsCmd)
}
