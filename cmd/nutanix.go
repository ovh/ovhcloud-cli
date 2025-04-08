
package cmd

import (
	"github.com/spf13/cobra"
)

var (
	nutanixColumnsToDisplay = []string{ "serviceName","status" }
)

func listNutanix(_ *cobra.Command, _ []string) {
	manageListRequest("/nutanix", nutanixColumnsToDisplay)
}

func getNutanix(_ *cobra.Command, args []string) {
	manageObjectRequest("/nutanix", args[0], nutanixColumnsToDisplay[0])
}

func init() {
	nutanixCmd := &cobra.Command{
		Use:   "nutanix",
		Short: "Retrieve information and manage your Nutanix services",
	}

	// Command to list Nutanix services
	nutanixCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List your Nutanix services",
		Run:   listNutanix,
	})

	// Command to get a single Nutanix
	nutanixCmd.AddCommand(&cobra.Command{
		Use:        "get",
		Short:      "Retrieve information of a specific Nutanix",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"service_name"},
		Run:        getNutanix,
	})

	rootCmd.AddCommand(nutanixCmd)
}
