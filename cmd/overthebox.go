package cmd

import (
	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/services/overthebox"
)

func init() {
	overtheboxCmd := &cobra.Command{
		Use:   "overthebox",
		Short: "Retrieve information and manage your OverTheBox services",
	}

	// Command to list OverTheBox services
	overtheboxListCmd := &cobra.Command{
		Use:   "list",
		Short: "List your OverTheBox services",
		Run:   overthebox.ListOverTheBox,
	}
	overtheboxCmd.AddCommand(withFilterFlag(overtheboxListCmd))

	// Command to get a single OverTheBox
	overtheboxCmd.AddCommand(&cobra.Command{
		Use:   "get <service_name>",
		Short: "Retrieve information of a specific OverTheBox",
		Args:  cobra.ExactArgs(1),
		Run:   overthebox.GetOverTheBox,
	})

	// Command to update a single OverTheBox
	overtheboxCmd.AddCommand(&cobra.Command{
		Use:   "edit <service_name>",
		Short: "Edit the given OverTheBox",
		Run:   overthebox.EditOverTheBox,
	})

	rootCmd.AddCommand(overtheboxCmd)
}
