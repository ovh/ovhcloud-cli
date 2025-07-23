package cmd

import (
	"github.com/ovh/ovhcloud-cli/internal/services/nutanix"
	"github.com/spf13/cobra"
)

func init() {
	nutanixCmd := &cobra.Command{
		Use:   "nutanix",
		Short: "Retrieve information and manage your Nutanix services",
	}

	// Command to list Nutanix services
	nutanixListCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List your Nutanix services",
		Run:     nutanix.ListNutanix,
	}
	nutanixCmd.AddCommand(withFilterFlag(nutanixListCmd))

	// Command to get a single Nutanix
	nutanixCmd.AddCommand(&cobra.Command{
		Use:   "get <service_name>",
		Short: "Retrieve information of a specific Nutanix",
		Args:  cobra.ExactArgs(1),
		Run:   nutanix.GetNutanix,
	})

	rootCmd.AddCommand(nutanixCmd)
}
