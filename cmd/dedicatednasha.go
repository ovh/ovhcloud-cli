
package cmd

import (
	"github.com/spf13/cobra"
)

var (
	dedicatednashaColumnsToDisplay = []string{ "serviceName","customName","datacenter" }
)

func listDedicatedNasHA(_ *cobra.Command, _ []string) {
	manageListRequest("/dedicated/nasha", dedicatednashaColumnsToDisplay)
}

func getDedicatedNasHA(_ *cobra.Command, args []string) {
	manageObjectRequest("/dedicated/nasha", args[0], dedicatednashaColumnsToDisplay[0])
}

func init() {
	dedicatednashaCmd := &cobra.Command{
		Use:   "dedicatednasha",
		Short: "Retrieve information and manage your DedicatedNasHA services",
	}

	// Command to list DedicatedNasHA services
	dedicatednashaCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List your DedicatedNasHA services",
		Run:   listDedicatedNasHA,
	})

	// Command to get a single DedicatedNasHA
	dedicatednashaCmd.AddCommand(&cobra.Command{
		Use:        "get",
		Short:      "Retrieve information of a specific DedicatedNasHA",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"service_name"},
		Run:        getDedicatedNasHA,
	})

	rootCmd.AddCommand(dedicatednashaCmd)
}
