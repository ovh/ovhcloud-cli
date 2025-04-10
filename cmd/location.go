package cmd

import (
	"github.com/spf13/cobra"
)

var (
	locationColumnsToDisplay = []string{ "name","type","specificType","location" }
)

func listLocation(_ *cobra.Command, _ []string) {
	manageListRequest("/v2/location", locationColumnsToDisplay, genericFilters)
}

func getLocation(_ *cobra.Command, args []string) {
	manageObjectRequest("/v2/location", args[0], locationColumnsToDisplay[0])
}

func init() {
	locationCmd := &cobra.Command{
		Use:   "location",
		Short: "Retrieve information and manage your Location services",
	}

	// Command to list Location services
	locationListCmd := &cobra.Command{
		Use:   "list",
		Short: "List your Location services",
		Run:   listLocation,
	}
	locationListCmd.PersistentFlags().StringArrayVar(
		&genericFilters,
		"filter",
		nil,
		"Filter results by any property using github.com/PaesslerAG/gval syntax'",
	)
	locationCmd.AddCommand(locationListCmd)

	// Command to get a single Location
	locationCmd.AddCommand(&cobra.Command{
		Use:        "get",
		Short:      "Retrieve information of a specific Location",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"service_name"},
		Run:        getLocation,
	})

	rootCmd.AddCommand(locationCmd)
}
