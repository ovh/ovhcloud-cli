
package cmd

import (
	"github.com/spf13/cobra"
)

var (
	overtheboxColumnsToDisplay = []string{ "serviceName","offer","status","bandwidth" }
)

func listOverTheBox(_ *cobra.Command, _ []string) {
	manageListRequest("/overTheBox", overtheboxColumnsToDisplay)
}

func getOverTheBox(_ *cobra.Command, args []string) {
	manageObjectRequest("/overTheBox", args[0], overtheboxColumnsToDisplay[0])
}

func init() {
	overtheboxCmd := &cobra.Command{
		Use:   "overthebox",
		Short: "Retrieve information and manage your OverTheBox services",
	}

	// Command to list OverTheBox services
	overtheboxCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List your OverTheBox services",
		Run:   listOverTheBox,
	})

	// Command to get a single OverTheBox
	overtheboxCmd.AddCommand(&cobra.Command{
		Use:        "get",
		Short:      "Retrieve information of a specific OverTheBox",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"service_name"},
		Run:        getOverTheBox,
	})

	rootCmd.AddCommand(overtheboxCmd)
}
