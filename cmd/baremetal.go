
package cmd

import (
	"github.com/spf13/cobra"
)

var (
	baremetalColumnsToDisplay = []string{ "serverId","name","region","os" }
)

func listBaremetal(_ *cobra.Command, _ []string) {
	manageListRequest("/dedicated/server", baremetalColumnsToDisplay)
}

func getBaremetal(_ *cobra.Command, args []string) {
	manageObjectRequest("/dedicated/server", args[0], baremetalColumnsToDisplay[0])
}

func init() {
	baremetalCmd := &cobra.Command{
		Use:   "baremetal",
		Short: "Retrieve information and manage your Baremetal services",
	}

	// Command to list Baremetal services
	baremetalCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List your Baremetal services",
		Run:   listBaremetal,
	})

	// Command to get a single Baremetal
	baremetalCmd.AddCommand(&cobra.Command{
		Use:        "get",
		Short:      "Retrieve information of a specific Baremetal",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"service_name"},
		Run:        getBaremetal,
	})

	rootCmd.AddCommand(baremetalCmd)
}
