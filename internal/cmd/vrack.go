package cmd

import (
	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/services/vrack"
)

func init() {
	vrackCmd := &cobra.Command{
		Use:   "vrack",
		Short: "Retrieve information and manage your vRack services",
	}

	// Command to list Vrack services
	vrackListCmd := &cobra.Command{
		Use:   "list",
		Short: "List your vRackservices",
		Run:   vrack.ListVrack,
	}
	vrackCmd.AddCommand(withFilterFlag(vrackListCmd))

	// Command to get a single Vrack
	vrackCmd.AddCommand(&cobra.Command{
		Use:   "get <service_name>",
		Short: "Retrieve information of a specific vRack",
		Args:  cobra.ExactArgs(1),
		Run:   vrack.GetVrack,
	})

	// Command to update a single Vrack
	vrackCmd.AddCommand(&cobra.Command{
		Use:   "edit <service_name>",
		Short: "Edit the given vRack",
		Run:   vrack.EditVrack,
	})

	rootCmd.AddCommand(vrackCmd)
}
