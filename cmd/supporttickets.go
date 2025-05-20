package cmd

import (
	_ "embed"

	"github.com/spf13/cobra"
)

var (
	supportticketsColumnsToDisplay = []string{"ticketId", "serviceName", "type", "category", "state"}

	//go:embed templates/supporttickets.tmpl
	supportticketsTemplate string
)

func listSupportTickets(_ *cobra.Command, _ []string) {
	manageListRequest("/support/tickets", "", supportticketsColumnsToDisplay, genericFilters)
}

func getSupportTickets(_ *cobra.Command, args []string) {
	manageObjectRequest("/support/tickets", args[0], supportticketsTemplate)
}

func init() {
	supportticketsCmd := &cobra.Command{
		Use:   "support-tickets",
		Short: "Retrieve information and manage your support tickets",
	}

	// Command to list tickets
	supportticketsListCmd := &cobra.Command{
		Use:   "list",
		Short: "List your support tickets",
		Run:   listSupportTickets,
	}
	supportticketsCmd.AddCommand(withFilterFlag(supportticketsListCmd))

	// Command to get a single support ticket
	supportticketsCmd.AddCommand(&cobra.Command{
		Use:   "get <ticket_id>",
		Short: "Retrieve information of a specific support ticket",
		Args:  cobra.ExactArgs(1),
		Run:   getSupportTickets,
	})

	rootCmd.AddCommand(supportticketsCmd)
}
