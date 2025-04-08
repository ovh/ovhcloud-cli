
package cmd

import (
	"github.com/spf13/cobra"
)

var (
	xdslColumnsToDisplay = []string{ "accessName","accessType","provider","role","status" }
)

func listXdsl(_ *cobra.Command, _ []string) {
	manageListRequest("/xdsl", xdslColumnsToDisplay)
}

func getXdsl(_ *cobra.Command, args []string) {
	manageObjectRequest("/xdsl", args[0], xdslColumnsToDisplay[0])
}

func init() {
	xdslCmd := &cobra.Command{
		Use:   "xdsl",
		Short: "Retrieve information and manage your Xdsl services",
	}

	// Command to list Xdsl services
	xdslCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List your Xdsl services",
		Run:   listXdsl,
	})

	// Command to get a single Xdsl
	xdslCmd.AddCommand(&cobra.Command{
		Use:        "get",
		Short:      "Retrieve information of a specific Xdsl",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"service_name"},
		Run:        getXdsl,
	})

	rootCmd.AddCommand(xdslCmd)
}
