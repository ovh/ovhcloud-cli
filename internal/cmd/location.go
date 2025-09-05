// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/ovh/ovhcloud-cli/internal/services/location"
	"github.com/spf13/cobra"
)

func init() {
	locationCmd := &cobra.Command{
		Use:   "location",
		Short: "Retrieve information and manage your Location services",
	}

	// Command to list Location services
	locationListCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List your Location services",
		Run:     location.ListLocation,
	}
	locationCmd.AddCommand(withFilterFlag(locationListCmd))

	// Command to get a single Location
	locationCmd.AddCommand(&cobra.Command{
		Use:   "get <location_name>",
		Short: "Retrieve information of a specific Location",
		Args:  cobra.ExactArgs(1),
		Run:   location.GetLocation,
	})

	rootCmd.AddCommand(locationCmd)
}
