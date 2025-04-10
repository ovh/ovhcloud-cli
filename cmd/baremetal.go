package cmd

import (
	"github.com/spf13/cobra"
)

var (
	baremetalColumnsToDisplay = []string{"name", "region", "os", "powerState", "state"}
)

func listBaremetal(_ *cobra.Command, _ []string) {
	manageListRequestWithFilters("/dedicated/server", baremetalColumnsToDisplay, genericFilters)
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
	baremetalListCmd := &cobra.Command{
		Use:   "list",
		Short: "List your Baremetal services",
		Run:   listBaremetal,
	}
	baremetalListCmd.PersistentFlags().StringArrayVar(
		&genericFilters,
		"filter",
		nil,
		`Filter results by any property, for example --filter 'iam.displayName==something'`,
	)
	baremetalCmd.AddCommand(baremetalListCmd)

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
