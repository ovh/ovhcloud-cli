package cmd

import (
	"github.com/ovh/ovhcloud-cli/internal/services/dedicatednasha"
	"github.com/spf13/cobra"
)

func init() {
	dedicatednashaCmd := &cobra.Command{
		Use:   "dedicated-nasha",
		Short: "Retrieve information and manage your Dedicated NasHA services",
	}

	// Command to list DedicatedNasHA services
	dedicatednashaListCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List your Dedicated NasHA services",
		Run:     dedicatednasha.ListDedicatedNasHA,
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
	editDedicatednashaCmd := &cobra.Command{
		Use:   "edit <service_name>",
		Short: "Edit the given Dedicated NasHA",
		Args:  cobra.ExactArgs(1),
		Run:   dedicatednasha.EditDedicatedNasHA,
	}
	editDedicatednashaCmd.Flags().StringVar(&dedicatednasha.DedicatedNasHASpec.CustomName, "custom-name", "", "Custom name for the Dedicated NasHA")
	editDedicatednashaCmd.Flags().BoolVar(&dedicatednasha.DedicatedNasHASpec.Monitored, "monitored", false, "Send an email to customer if any issue is detected")
	addInteractiveEditorFlag(editDedicatednashaCmd)
	dedicatednashaCmd.AddCommand(editDedicatednashaCmd)

	rootCmd.AddCommand(dedicatednashaCmd)
}
