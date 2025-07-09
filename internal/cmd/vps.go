package cmd

import (
	_ "embed"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/flags"
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
	vpsEditCmd := &cobra.Command{
		Use:   "edit <service_name>",
		Short: "Edit the given VPS",
		Args:  cobra.ExactArgs(1),
		Run:   vps.EditVps,
	}
	vpsEditCmd.Flags().StringVar(&vps.VPSSpec.DisplayName, "display-name", "", "Display name of the VPS")
	vpsEditCmd.Flags().StringVar(&vps.VPSSpec.Keymap, "keymap", "", "Keymap of the VPS (fr, us)")
	vpsEditCmd.Flags().StringVar(&vps.VPSSpec.NetbootMode, "netboot-mode", "", "Netboot mode of the VPS (local, rescue)")
	vpsEditCmd.Flags().BoolVar(&vps.VPSSpec.SlaMonitoring, "sla-monitoring", false, "Enable or disable SLA monitoring for the VPS")
	vpsEditCmd.Flags().BoolVar(&flags.ParametersViaEditor, "editor", false, "Use a text editor to define parameters")
	vpsCmd.AddCommand(vpsEditCmd)

	rootCmd.AddCommand(vpsCmd)
}
