package cmd

import (
	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/services/supporttickets"
)

func init() {
	supportticketsCmd := &cobra.Command{
		Use:   "support-tickets",
		Short: "Retrieve information and manage your support tickets",
	}

	// Command to list tickets
	supportticketsListCmd := &cobra.Command{
		Use:   "list",
		Short: "List your support tickets",
		Run:   supporttickets.ListSupportTickets,
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
