// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/ovh/ovhcloud-cli/internal/services/dedicatedcluster"
	"github.com/spf13/cobra"
)

func init() {
	dedicatedclusterCmd := &cobra.Command{
		Use:   "dedicated-cluster",
		Short: "Retrieve information and manage your DedicatedCluster services",
	}

	// Command to list DedicatedCluster services
	dedicatedclusterListCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List your DedicatedCluster services",
		Run:     dedicatedcluster.ListDedicatedCluster,
	}
	dedicatedclusterCmd.AddCommand(withFilterFlag(dedicatedclusterListCmd))

	// Command to get a single DedicatedCluster
	dedicatedclusterCmd.AddCommand(&cobra.Command{
		Use:   "get <service_name>",
		Short: "Retrieve information of a specific DedicatedCluster",
		Args:  cobra.ExactArgs(1),
		Run:   dedicatedcluster.GetDedicatedCluster,
	})

	rootCmd.AddCommand(dedicatedclusterCmd)
}
