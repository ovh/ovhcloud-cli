// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/ovh/ovhcloud-cli/internal/completion"
	"github.com/ovh/ovhcloud-cli/internal/services/location"
	"github.com/spf13/cobra"
)

func init() {
	locationCmd := &cobra.Command{
		Use:   "location",
		Short: "Browse OVHcloud locations (datacenters and regions)",
	}

	// Command to list Location services
	locationListCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List OVHcloud locations",
		Run:     location.ListLocation,
	}
	locationCmd.AddCommand(withFilterFlag(locationListCmd))

	// Command to get a single Location
	locationCmd.AddCommand(&cobra.Command{
		Use:               "get <location_name>",
		Short:             "Get details of a specific OVHcloud location",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.ServiceList("/v2/location"),
		Run:               location.GetLocation,
	})

	rootCmd.AddCommand(locationCmd)
}
