// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/ovh/ovhcloud-cli/internal/services/veeamenterprise"
	"github.com/spf13/cobra"
)

func init() {
	veeamenterpriseCmd := &cobra.Command{
		Use:   "veeamenterprise",
		Short: "Retrieve information and manage your VeeamEnterprise services",
	}

	// Command to list VeeamEnterprise services
	veeamenterpriseListCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List your VeeamEnterprise services",
		Run:     veeamenterprise.ListVeeamEnterprise,
	}
	veeamenterpriseCmd.AddCommand(withFilterFlag(veeamenterpriseListCmd))

	// Command to get a single VeeamEnterprise
	veeamenterpriseCmd.AddCommand(&cobra.Command{
		Use:   "get <service_name>",
		Short: "Retrieve information of a specific VeeamEnterprise",
		Args:  cobra.ExactArgs(1),
		Run:   veeamenterprise.GetVeeamEnterprise,
	})

	rootCmd.AddCommand(veeamenterpriseCmd)
}
