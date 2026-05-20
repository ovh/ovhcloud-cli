// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"

	"github.com/ovh/ovhcloud-cli/internal/services/cloud"
	"github.com/spf13/cobra"
	"github.com/ovh/ovhcloud-cli/internal/completion"
)

// cloudIPType holds the value of the persistent --type flag for `cloud ip` commands.
// Allowed values: "floating", "failover".
var cloudIPType string

const (
	cloudIPTypeFloating = "floating"
	cloudIPTypeFailover = "failover"
)

func validateCloudIPType(allowed ...string) error {
	for _, a := range allowed {
		if cloudIPType == a {
			return nil
		}
	}
	if cloudIPType == "" {
		return fmt.Errorf("--type flag is required (allowed values: %v)", allowed)
	}
	return fmt.Errorf("invalid --type %q (allowed values for this command: %v)", cloudIPType, allowed)
}

func initCloudIPCommand(cloudCmd *cobra.Command) {
	ipCmd := &cobra.Command{
		Use:   "ip",
		Short: "Manage public IPs (floating and failover) in the given cloud project",
	}
	ipCmd.PersistentFlags().StringVar(&cloud.CloudProject, "cloud-project", "", "Cloud project ID")
	ipCmd.RegisterFlagCompletionFunc("cloud-project", completion.CloudProjects) //nolint:errcheck
	ipCmd.PersistentFlags().StringVar(&cloudIPType, "type", "", "Type of IP to manage (floating or failover)")

	// region flag is only relevant for floating IPs (failover IPs are project-scoped)
	ipCmd.PersistentFlags().StringVar(&cloud.CloudFloatingIPRegionFilter, "region", "", "Filter by region or specify the region of the floating IP (only used when --type=floating)")

	// list
	listCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List public IPs (both floating and failover when --type is not specified)",
		PreRunE: func(_ *cobra.Command, _ []string) error {
			// --type is optional for list. If provided, validate it.
			if cloudIPType == "" {
				return nil
			}
			return validateCloudIPType(cloudIPTypeFloating, cloudIPTypeFailover)
		},
		Run: func(cmd *cobra.Command, args []string) {
			switch cloudIPType {
			case cloudIPTypeFloating:
				cloud.ListFloatingIPs(cmd, args)
			case cloudIPTypeFailover:
				cloud.ListCloudIPFailovers(cmd, args)
			default:
				cloud.ListAllCloudIPs(cmd, args)
			}
		},
	}
	ipCmd.AddCommand(withFilterFlag(listCmd))

	// get
	ipCmd.AddCommand(&cobra.Command{
		Use:   "get <ip_id>",
		Short: "Get information about a public IP",
		Args:  cobra.ExactArgs(1),
		PreRunE: func(_ *cobra.Command, _ []string) error {
			return validateCloudIPType(cloudIPTypeFloating, cloudIPTypeFailover)
		},
		Run: func(cmd *cobra.Command, args []string) {
			switch cloudIPType {
			case cloudIPTypeFloating:
				cloud.GetFloatingIP(cmd, args)
			case cloudIPTypeFailover:
				cloud.GetCloudIPFailover(cmd, args)
			}
		},
	})

	// delete (floating only)
	ipCmd.AddCommand(&cobra.Command{
		Use:   "delete <ip_id>",
		Short: "Delete a public IP (only supported for --type=floating)",
		Args:  cobra.ExactArgs(1),
		PreRunE: func(_ *cobra.Command, _ []string) error {
			return validateCloudIPType(cloudIPTypeFloating)
		},
		Run: cloud.DeleteFloatingIP,
	})

	// attach (failover only)
	ipCmd.AddCommand(&cobra.Command{
		Use:   "attach <ip_id> <instance_id>",
		Short: "Attach a public IP to an instance (only supported for --type=failover)",
		Args:  cobra.ExactArgs(2),
		PreRunE: func(_ *cobra.Command, _ []string) error {
			return validateCloudIPType(cloudIPTypeFailover)
		},
		Run: cloud.AttachCloudIPFailover,
	})

	cloudCmd.AddCommand(ipCmd)
}
