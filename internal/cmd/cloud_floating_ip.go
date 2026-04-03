// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/ovh/ovhcloud-cli/internal/services/cloud"
	"github.com/spf13/cobra"
)

func initCloudFloatingIPCommand(cloudCmd *cobra.Command) {
	floatingIPCmd := &cobra.Command{
		Use:   "floating-ip",
		Short: "Manage floating IPs in the given cloud project",
	}
	floatingIPCmd.PersistentFlags().StringVar(&cloud.CloudProject, "cloud-project", "", "Cloud project ID")
	floatingIPCmd.PersistentFlags().StringVar(&cloud.CloudFloatingIPRegionFilter, "region", "", "Filter by region or specify the region of the floating IP")

	floatingIPListCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List floating IPs",
		Run:     cloud.ListFloatingIPs,
	}
	floatingIPCmd.AddCommand(withFilterFlag(floatingIPListCmd))

	floatingIPCmd.AddCommand(&cobra.Command{
		Use:   "get <floating_ip_id>",
		Short: "Get information about a floating IP",
		Run:   cloud.GetFloatingIP,
		Args:  cobra.ExactArgs(1),
	})

	floatingIPCmd.AddCommand(&cobra.Command{
		Use:   "delete <floating_ip_id>",
		Short: "Delete a floating IP",
		Run:   cloud.DeleteFloatingIP,
		Args:  cobra.ExactArgs(1),
	})

	cloudCmd.AddCommand(floatingIPCmd)
}
