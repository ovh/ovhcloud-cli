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
		Short: "List your DedicatedNasHA services",
		Run:   dedicatednasha.ListDedicatedNasHA,
	}
	dedicatednashaCmd.AddCommand(withFilterFlag(dedicatednashaListCmd))

	// Command to get a single DedicatedNasHA
	dedicatednashaCmd.AddCommand(&cobra.Command{
		Use:   "get <service_name>",
		Short: "Retrieve information of a specific DedicatedNasHA",
		Args:  cobra.ExactArgs(1),
		Run:   dedicatednasha.GetDedicatedNasHA,
	})

	rootCmd.AddCommand(dedicatednashaCmd)
}
