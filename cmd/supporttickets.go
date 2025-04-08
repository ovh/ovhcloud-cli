
package cmd

import (
	"github.com/spf13/cobra"
)

var (
	supportticketsColumnsToDisplay = []string{ "ticketId","serviceName","type","category","state" }
)

func listSupportTickets(_ *cobra.Command, _ []string) {
	manageListRequest("/support/tickets", supportticketsColumnsToDisplay)
}

func getSupportTickets(_ *cobra.Command, args []string) {
	manageObjectRequest("/support/tickets", args[0], supportticketsColumnsToDisplay[0])
}

func init() {
	supportticketsCmd := &cobra.Command{
		Use:   "supporttickets",
		Short: "Retrieve information and manage your SupportTickets services",
	}

	// Command to list SupportTickets services
	supportticketsCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List your SupportTickets services",
		Run:   listSupportTickets,
	})

	// Command to get a single SupportTickets
	supportticketsCmd.AddCommand(&cobra.Command{
		Use:        "get",
		Short:      "Retrieve information of a specific SupportTickets",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"service_name"},
		Run:        getSupportTickets,
	})

	rootCmd.AddCommand(supportticketsCmd)
}
