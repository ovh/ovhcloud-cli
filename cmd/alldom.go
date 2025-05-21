package cmd

import (
	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/services/alldom"
)

func init() {
	alldomCmd := &cobra.Command{
		Use:   "alldom",
		Short: "Retrieve information and manage your AllDom services",
	}

	// Command to list AllDom services
	alldomListCmd := &cobra.Command{
		Use:   "list",
		Short: "List your AllDom services",
		Run:   alldom.ListAllDom,
	}
	alldomCmd.AddCommand(withFilterFlag(alldomListCmd))

	// Command to get a single AllDom
	alldomCmd.AddCommand(&cobra.Command{
		Use:   "get <service_name>",
		Short: "Retrieve information of a specific AllDom",
		Args:  cobra.ExactArgs(1),
		Run:   alldom.GetAllDom,
	})

	rootCmd.AddCommand(alldomCmd)
}
