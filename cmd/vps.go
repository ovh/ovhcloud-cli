
package cmd

import (
	"github.com/spf13/cobra"
)

var (
	vpsColumnsToDisplay = []string{ "name","displayName","state","zone" }
)

func listVps(_ *cobra.Command, _ []string) {
	manageListRequest("/vps", vpsColumnsToDisplay)
}

func getVps(_ *cobra.Command, args []string) {
	manageObjectRequest("/vps", args[0], vpsColumnsToDisplay[0])
}

func init() {
	vpsCmd := &cobra.Command{
		Use:   "vps",
		Short: "Retrieve information and manage your Vps services",
	}

	// Command to list Vps services
	vpsCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List your Vps services",
		Run:   listVps,
	})

	// Command to get a single Vps
	vpsCmd.AddCommand(&cobra.Command{
		Use:        "get",
		Short:      "Retrieve information of a specific Vps",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"service_name"},
		Run:        getVps,
	})

	rootCmd.AddCommand(vpsCmd)
}
