package cmd

import (
	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/flags"
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
	overtheboxEditCmd := &cobra.Command{
		Use:   "edit <service_name>",
		Short: "Edit the given OverTheBox",
		Args:  cobra.ExactArgs(1),
		Run:   overthebox.EditOverTheBox,
	}
	overtheboxEditCmd.Flags().BoolVar(&overthebox.OverTheBoxSpec.AutoUpgrade, "auto-upgrade", false, "Enable device auto upgrade")
	overtheboxEditCmd.Flags().StringVar(&overthebox.OverTheBoxSpec.CustomerDescription, "customer-description", "", "Customer description")
	overtheboxEditCmd.Flags().StringVar(&overthebox.OverTheBoxSpec.ReleaseChannel, "release-channel", "", "Release channel")
	overtheboxEditCmd.Flags().BoolVar(&flags.ParametersViaEditor, "editor", false, "Use a text editor to define parameters")
	overtheboxCmd.AddCommand(overtheboxEditCmd)

	rootCmd.AddCommand(overtheboxCmd)
}
