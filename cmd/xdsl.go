package cmd

import (
	_ "embed"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/services/xdsl"
)

func init() {
	xdslCmd := &cobra.Command{
		Use:   "xdsl",
		Short: "Retrieve information and manage your XDSL services",
	}

	// Command to list Xdsl services
	xdslListCmd := &cobra.Command{
		Use:   "list",
		Short: "List your XDSL services",
		Run:   xdsl.ListXdsl,
	}
	xdslCmd.AddCommand(withFilterFlag(xdslListCmd))

	// Command to get a single Xdsl
	xdslCmd.AddCommand(&cobra.Command{
		Use:   "get <service_name>",
		Short: "Retrieve information of a specific XDSL",
		Args:  cobra.ExactArgs(1),
		Run:   xdsl.GetXdsl,
	})

	// Command to update a single Xdsl
	xdslCmd.AddCommand(&cobra.Command{
		Use:   "edit <service_name>",
		Short: "Edit the given XDSL",
		Run:   xdsl.EditXdsl,
	})

	rootCmd.AddCommand(xdslCmd)
}
