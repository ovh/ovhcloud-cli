package cmd

import (
	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/services/dedicatednasha"
)

func init() {
	dedicatednashaCmd := &cobra.Command{
		Use:   "dedicated-nasha",
		Short: "Retrieve information and manage your Dedicated NasHA services",
	}

	// Command to list DedicatedNasHA services
	dedicatednashaListCmd := &cobra.Command{
		Use:   "list",
		Short: "List your Dedicated NasHA services",
		Run:   dedicatednasha.ListDedicatedNasHA,
	}
	dedicatednashaCmd.AddCommand(withFilterFlag(dedicatednashaListCmd))

	// Command to get a single DedicatedNasHA
	dedicatednashaCmd.AddCommand(&cobra.Command{
		Use:   "get <service_name>",
		Short: "Retrieve information of a specific Dedicated NasHA",
		Args:  cobra.ExactArgs(1),
		Run:   dedicatednasha.GetDedicatedNasHA,
	})

	// Command to update a single DedicatedNasHA
	dedicatednashaCmd.AddCommand(&cobra.Command{
		Use:   "edit <service_name>",
		Short: "Edit the given Dedicated NasHA",
		Run:   dedicatednasha.EditDedicatedNasHA,
	})

	rootCmd.AddCommand(dedicatednashaCmd)
}
