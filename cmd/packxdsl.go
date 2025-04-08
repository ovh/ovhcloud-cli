
package cmd

import (
	"github.com/spf13/cobra"
)

var (
	packxdslColumnsToDisplay = []string{ "packName","description" }
)

func listPackXDSL(_ *cobra.Command, _ []string) {
	manageListRequest("/pack/xdsl", packxdslColumnsToDisplay)
}

func getPackXDSL(_ *cobra.Command, args []string) {
	manageObjectRequest("/pack/xdsl", args[0], packxdslColumnsToDisplay[0])
}

func init() {
	packxdslCmd := &cobra.Command{
		Use:   "packxdsl",
		Short: "Retrieve information and manage your PackXDSL services",
	}

	// Command to list PackXDSL services
	packxdslCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List your PackXDSL services",
		Run:   listPackXDSL,
	})

	// Command to get a single PackXDSL
	packxdslCmd.AddCommand(&cobra.Command{
		Use:        "get",
		Short:      "Retrieve information of a specific PackXDSL",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"service_name"},
		Run:        getPackXDSL,
	})

	rootCmd.AddCommand(packxdslCmd)
}
