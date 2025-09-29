// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	_ "embed"

	"github.com/ovh/ovhcloud-cli/internal/services/iploadbalancing"
	"github.com/spf13/cobra"
)

func init() {
	iploadbalancingCmd := &cobra.Command{
		Use:   "iploadbalancing",
		Short: "Retrieve information and manage your IP LoadBalancing services",
	}

	// Command to list IpLoadbalancing services
	iploadbalancingListCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List your IpLoadbalancing services",
		Run:     iploadbalancing.ListIpLoadbalancing,
	}
	iploadbalancingCmd.AddCommand(withFilterFlag(iploadbalancingListCmd))

	// Command to get a single IpLoadbalancing
	iploadbalancingCmd.AddCommand(&cobra.Command{
		Use:   "get <service_name>",
		Short: "Retrieve information of a specific IpLoadbalancing",
		Args:  cobra.ExactArgs(1),
		Run:   iploadbalancing.GetIpLoadbalancing,
	})

	// Command to update a single IpLoadbalancing
	iploadbalancingEditCmd := &cobra.Command{
		Use:   "edit <service_name>",
		Short: "Edit the given IpLoadbalancing",
		Args:  cobra.ExactArgs(1),
		Run:   iploadbalancing.EditIpLoadbalancing,
	}
	iploadbalancingEditCmd.Flags().StringVar(&iploadbalancing.IPLoadbalancingSpec.DisplayName, "display-name", "", "Display name of the load balancer")
	iploadbalancingEditCmd.Flags().StringVar(&iploadbalancing.IPLoadbalancingSpec.SSLConfiguration, "ssl-configuration", "", "SSL configuration of the load balancer (intermediate, modern)")
	addInteractiveEditorFlag(iploadbalancingEditCmd)
	iploadbalancingCmd.AddCommand(iploadbalancingEditCmd)

	rootCmd.AddCommand(iploadbalancingCmd)
}
