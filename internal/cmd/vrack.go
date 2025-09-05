// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
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
		Use:   "get <service_name>",
		Short: "Retrieve information of a specific vRack",
		Args:  cobra.ExactArgs(1),
		Run:   vrack.GetVrack,
	})

	// Command to update a single Vrack
	vrackEditCmd := &cobra.Command{
		Use:   "edit <service_name>",
		Short: "Edit the given vRack",
		Args:  cobra.ExactArgs(1),
		Run:   vrack.EditVrack,
	}
	vrackEditCmd.Flags().StringVar(&vrack.VrackSpec.Name, "name", "", "Name of the vRack")
	vrackEditCmd.Flags().StringVar(&vrack.VrackSpec.Description, "description", "", "Description of the vRack")
	addInteractiveEditorFlag(vrackEditCmd)
	vrackCmd.AddCommand(vrackEditCmd)

	rootCmd.AddCommand(vrackCmd)
}
