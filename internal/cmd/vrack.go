// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"time"

	"github.com/ovh/ovhcloud-cli/internal/completion"
	"github.com/ovh/ovhcloud-cli/internal/services/vrack"
	"github.com/spf13/cobra"
)

func init() {
	vrackCmd := &cobra.Command{
		Use:   "vrack",
		Short: "Retrieve information and manage your vRack services",
	}

	// Command to list Vrack services
	vrackListCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List your vRackservices",
		Run:     vrack.ListVrack,
	}
	vrackCmd.AddCommand(withFilterFlag(vrackListCmd))

	// Command to get a single Vrack
	vrackCmd.AddCommand(&cobra.Command{
		Use:               "get <service_name>",
		Short:             "Retrieve information of a specific vRack",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.ServiceList("/v1/vrack"),
		Run:               vrack.GetVrack,
	})

	// Command to update a single Vrack
	vrackEditCmd := &cobra.Command{
		Use:               "edit <service_name>",
		Short:             "Edit the given vRack",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.ServiceList("/v1/vrack"),
		Run:               vrack.EditVrack,
	}
	vrackEditCmd.Flags().StringVar(&vrack.VrackSpec.Name, "name", "", "Name of the vRack")
	vrackEditCmd.Flags().StringVar(&vrack.VrackSpec.Description, "description", "", "Description of the vRack")
	addInteractiveEditorFlag(vrackEditCmd)
	vrackCmd.AddCommand(vrackEditCmd)

	// Attach and detach: the two gestures that made this domain read-only.
	vrackAttachCmd := &cobra.Command{
		Use:   "attach <service_name> <server_or_interface>",
		Short: "Attach a dedicated server to the given vRack",
		Long: `Attach a dedicated server to the given vRack.

The server can be named by its service name, in which case its vRack interface
is looked up, or by the interface UUID directly. A server with several
interfaces requires --interface, because picking one would sometimes be the
wrong network.`,
		Args:              cobra.ExactArgs(2),
		ValidArgsFunction: completion.ServiceList("/v1/vrack"),
		Run:               vrack.AttachToVrack,
	}
	addVrackTaskFlags(vrackAttachCmd)
	vrackAttachCmd.Flags().StringVar(&vrack.VrackInterface, "interface", "",
		"Interface to attach, when the server has several")
	addConfirmationFlags(vrackAttachCmd, "Print the call that would be made without making it")
	vrackCmd.AddCommand(vrackAttachCmd)

	vrackDetachCmd := &cobra.Command{
		Use:               "detach <service_name> <server_or_interface>",
		Short:             "Detach a dedicated server from the given vRack",
		Args:              cobra.ExactArgs(2),
		ValidArgsFunction: completion.ServiceList("/v1/vrack"),
		Run:               vrack.DetachFromVrack,
	}
	addVrackTaskFlags(vrackDetachCmd)
	addConfirmationFlags(vrackDetachCmd, "Print the call that would be made without making it")
	vrackCmd.AddCommand(vrackDetachCmd)

	rootCmd.AddCommand(vrackCmd)
}

// addVrackTaskFlags gives a command the pair that governs waiting.
//
// --timeout only means something with --wait, and they are added together so a
// command cannot end up offering one without the other.
func addVrackTaskFlags(cmd *cobra.Command) {
	cmd.Flags().BoolVar(&vrack.VrackWait, "wait", false,
		"Wait until the vRack reports the change before exiting")
	cmd.Flags().DurationVar(&vrack.VrackTimeout, "timeout", 10*time.Minute,
		"How long --wait waits")
}
