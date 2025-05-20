package cmd

import (
	_ "embed"

	"github.com/spf13/cobra"
)

var (
	locationColumnsToDisplay = []string{"name", "type", "specificType", "location"}

	//go:embed templates/location.tmpl
	locationTemplate string
)

func listLocation(_ *cobra.Command, _ []string) {
	manageListRequest("/v2/location", "name", locationColumnsToDisplay, genericFilters)
}

func getLocation(_ *cobra.Command, args []string) {
	manageObjectRequest("/v2/location", args[0], locationTemplate)
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
	locationCmd.AddCommand(withFilterFlag(locationListCmd))

	// Command to get a single Location
	locationCmd.AddCommand(&cobra.Command{
		Use:   "get <location_name>",
		Short: "Retrieve information of a specific Location",
		Args:  cobra.ExactArgs(1),
		Run:   getLocation,
	})

	rootCmd.AddCommand(locationCmd)
}
