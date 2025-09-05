// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/ovh/ovhcloud-cli/internal/services/okms"
	"github.com/spf13/cobra"
)

func init() {
	okmsCmd := &cobra.Command{
		Use:   "okms",
		Short: "Retrieve information and manage your Okms services",
	}

	// Command to list Okms services
	okmsListCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List your Okms services",
		Run:     okms.ListOkms,
	}
	okmsCmd.AddCommand(withFilterFlag(okmsListCmd))

	// Command to get a single Okms
	okmsCmd.AddCommand(&cobra.Command{
		Use:   "get <service_name>",
		Short: "Retrieve information of a specific Okms",
		Args:  cobra.ExactArgs(1),
		Run:   okms.GetOkms,
	})

	rootCmd.AddCommand(okmsCmd)
}
