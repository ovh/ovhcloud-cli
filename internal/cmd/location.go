package cmd

import (
	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/services/location"
)

func init() {
	locationCmd := &cobra.Command{
		Use:   "location",
		Short: "Retrieve information and manage your Location services",
	}

	// Command to list Location services
	locationListCmd := &cobra.Command{
		Use:   "list",
		Short: "List your Location services",
		Run:   location.ListLocation,
	}
	locationCmd.AddCommand(withFilterFlag(locationListCmd))

	// Command to get a single Location
	locationCmd.AddCommand(&cobra.Command{
		Use:   "get <location_name>",
		Short: "Retrieve information of a specific Location",
		Args:  cobra.ExactArgs(1),
		Run:   location.GetLocation,
	})

	rootCmd.AddCommand(locationCmd)
}
