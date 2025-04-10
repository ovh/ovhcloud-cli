package cmd

import (
	"github.com/spf13/cobra"
)

var (
	okmsColumnsToDisplay = []string{ "id","region" }
)

func listOkms(_ *cobra.Command, _ []string) {
	manageListRequest("/v2/okms/resource", okmsColumnsToDisplay, genericFilters)
}

func getOkms(_ *cobra.Command, args []string) {
	manageObjectRequest("/v2/okms/resource", args[0], okmsColumnsToDisplay[0])
}

func init() {
	okmsCmd := &cobra.Command{
		Use:   "okms",
		Short: "Retrieve information and manage your Okms services",
	}

	// Command to list Okms services
	okmsListCmd := &cobra.Command{
		Use:   "list",
		Short: "List your Okms services",
		Run:   listOkms,
	}
	okmsListCmd.PersistentFlags().StringArrayVar(
		&genericFilters,
		"filter",
		nil,
		"Filter results by any property using github.com/PaesslerAG/gval syntax'",
	)
	okmsCmd.AddCommand(okmsListCmd)

	// Command to get a single Okms
	okmsCmd.AddCommand(&cobra.Command{
		Use:        "get",
		Short:      "Retrieve information of a specific Okms",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"service_name"},
		Run:        getOkms,
	})

	rootCmd.AddCommand(okmsCmd)
}
