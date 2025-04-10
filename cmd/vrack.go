package cmd

import (
	"github.com/spf13/cobra"
)

var (
	vrackColumnsToDisplay = []string{ "name","description" }
)

func listVrack(_ *cobra.Command, _ []string) {
	manageListRequest("/vrack", vrackColumnsToDisplay, genericFilters)
}

func getVrack(_ *cobra.Command, args []string) {
	manageObjectRequest("/vrack", args[0], vrackColumnsToDisplay[0])
}

func init() {
	vrackCmd := &cobra.Command{
		Use:   "vrack",
		Short: "Retrieve information and manage your Vrack services",
	}

	// Command to list Vrack services
	vrackListCmd := &cobra.Command{
		Use:   "list",
		Short: "List your Vrack services",
		Run:   listVrack,
	}
	vrackListCmd.PersistentFlags().StringArrayVar(
		&genericFilters,
		"filter",
		nil,
		"Filter results by any property using github.com/PaesslerAG/gval syntax'",
	)
	vrackCmd.AddCommand(vrackListCmd)

	// Command to get a single Vrack
	vrackCmd.AddCommand(&cobra.Command{
		Use:        "get",
		Short:      "Retrieve information of a specific Vrack",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"service_name"},
		Run:        getVrack,
	})

	rootCmd.AddCommand(vrackCmd)
}
