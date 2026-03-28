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

	// Command to list messages for a support ticket
	supportticketsMessagesCmd := &cobra.Command{
		Use:     "messages <ticket_id>",
		Aliases: []string{"msgs"},
		Short:   "List messages for a support ticket",
		Args:    cobra.ExactArgs(1),
		Run:     supporttickets.ListSupportTicketMessages,
	}
	supportticketsCmd.AddCommand(withFilterFlag(supportticketsMessagesCmd))

	// Command to reply to a support ticket
	supportticketsReplyCmd := &cobra.Command{
		Use:   "reply <ticket_id>",
		Short: "Reply to a support ticket",
		Args:  cobra.ExactArgs(1),
		Run:   supporttickets.ReplySupportTicket,
	}
	supportticketsReplyCmd.Flags().StringVar(&supporttickets.ReplySpec.Body, "body", "", "Text body of the ticket reply")
	_ = supportticketsReplyCmd.MarkFlagRequired("body")
	supportticketsCmd.AddCommand(supportticketsReplyCmd)

	rootCmd.AddCommand(supportticketsCmd)
}
