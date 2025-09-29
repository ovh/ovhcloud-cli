// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/ovh/ovhcloud-cli/internal/services/ip"
	"github.com/spf13/cobra"
)

func init() {
	ipCmd := &cobra.Command{
		Use:   "ip",
		Short: "Retrieve information and manage your IP services",
	}

	// Command to list Ip services
	ipListCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List your Ip services",
		Run:     ip.ListIp,
	}
	ipCmd.AddCommand(withFilterFlag(ipListCmd))

	// Command to get a single Ip
	ipCmd.AddCommand(&cobra.Command{
		Use:   "get <service_name>",
		Short: "Retrieve information of a specific Ip",
		Args:  cobra.ExactArgs(1),
		Run:   ip.GetIp,
	})

	// Command to update a single Ip
	ipEditCmd := &cobra.Command{
		Use:   "edit <service_name>",
		Short: "Edit the given IP",
		Args:  cobra.ExactArgs(1),
		Run:   ip.EditIp,
	}
	ipEditCmd.Flags().StringVar(&ip.IPSpec.Description, "description", "", "Description of the IP")
	addInteractiveEditorFlag(ipEditCmd)
	ipCmd.AddCommand(ipEditCmd)

	ipReverseCmd := &cobra.Command{
		Use:   "reverse",
		Short: "Manage reverses on the given IP",
	}
	ipCmd.AddCommand(ipReverseCmd)

	ipReverseSetCmd := &cobra.Command{
		Use:   "set <service_name> <ip> <reverse>",
		Short: "Set reverse on the given IP",
		Args:  cobra.ExactArgs(3),
		Run:   ip.IpSetReverse,
	}
	ipReverseCmd.AddCommand(ipReverseSetCmd)

	ipReverseGetCmd := &cobra.Command{
		Use:   "get <service_name>",
		Short: "List reverse on the given IP range",
		Args:  cobra.ExactArgs(1),
		Run:   ip.IpGetReverse,
	}
	ipReverseCmd.AddCommand(ipReverseGetCmd)

	ipReverseDeleteCmd := &cobra.Command{
		Use:   "delete <service_name> <ip>",
		Short: "Delete reverse on the given IP",
		Args:  cobra.ExactArgs(2),
		Run:   ip.IpDeleteReverse,
	}
	ipReverseCmd.AddCommand(ipReverseDeleteCmd)

	rootCmd.AddCommand(ipCmd)
}
