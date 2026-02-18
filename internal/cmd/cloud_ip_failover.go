// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/ovh/ovhcloud-cli/internal/services/cloud"
	"github.com/spf13/cobra"
)

func initCloudIPFailoverCommand(cloudCmd *cobra.Command) {
	ipFailoverCmd := &cobra.Command{
		Use:   "ip-failover",
		Short: "Manage failover IPs in the given cloud project",
	}
	ipFailoverCmd.PersistentFlags().StringVar(&cloud.CloudProject, "cloud-project", "", "Cloud project ID")

	ipFailoverListCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List failover IPs",
		Run:     cloud.ListCloudIPFailovers,
	}
	ipFailoverCmd.AddCommand(withFilterFlag(ipFailoverListCmd))

	ipFailoverCmd.AddCommand(&cobra.Command{
		Use:   "get <failover_ip_id>",
		Short: "Get information about a failover IP",
		Run:   cloud.GetCloudIPFailover,
		Args:  cobra.ExactArgs(1),
	})

	ipFailoverCmd.AddCommand(&cobra.Command{
		Use:   "attach <failover_ip_id> <instance_id>",
		Short: "Attach a failover IP to an instance",
		Run:   cloud.AttachCloudIPFailover,
		Args:  cobra.ExactArgs(2),
	})

	cloudCmd.AddCommand(ipFailoverCmd)
}
