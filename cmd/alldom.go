
package cmd

import (
	"github.com/spf13/cobra"
)

var (
	alldomColumnsToDisplay = []string{ "name","type","offer" }
)

func listAllDom(_ *cobra.Command, _ []string) {
	manageListRequest("/allDom", alldomColumnsToDisplay)
}

func getAllDom(_ *cobra.Command, args []string) {
	manageObjectRequest("/allDom", args[0], alldomColumnsToDisplay[0])
}

func init() {
	alldomCmd := &cobra.Command{
		Use:   "alldom",
		Short: "Retrieve information and manage your AllDom services",
	}

	// Command to list AllDom services
	alldomCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List your AllDom services",
		Run:   listAllDom,
	})

	// Command to get a single AllDom
	alldomCmd.AddCommand(&cobra.Command{
		Use:        "get",
		Short:      "Retrieve information of a specific AllDom",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"service_name"},
		Run:        getAllDom,
	})

	rootCmd.AddCommand(alldomCmd)
}
