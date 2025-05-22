package cmd

import (
	_ "embed"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/services/vps"
)

func init() {
	vpsCmd := &cobra.Command{
		Use:   "vps",
		Short: "Retrieve information and manage your VPS services",
	}

	// Command to list VPS services
	vpsListCmd := &cobra.Command{
		Use:   "list",
		Short: "List your VPS services",
		Run:   vps.ListVps,
	}
	vpsCmd.AddCommand(withFilterFlag(vpsListCmd))

	// Command to get a single VPS
	vpsCmd.AddCommand(&cobra.Command{
		Use:   "get <service_name>",
		Short: "Retrieve information of a specific VPS",
		Args:  cobra.ExactArgs(1),
		Run:   vps.GetVps,
	})

	// Command to update a single VPS
	vpsCmd.AddCommand(&cobra.Command{
		Use:   "edit <service_name>",
		Short: "Edit the given VPS",
		Run:   vps.EditVps,
	})

	rootCmd.AddCommand(vpsCmd)
}
