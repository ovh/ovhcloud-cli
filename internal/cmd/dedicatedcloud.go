// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/ovh/ovhcloud-cli/internal/services/dedicatedcloud"
	"github.com/spf13/cobra"
)

func init() {
	dedicatedcloudCmd := &cobra.Command{
		Use:   "dedicated-cloud",
		Short: "Retrieve information and manage your DedicatedCloud services",
	}

	// Command to list DedicatedCloud services
	dedicatedcloudListCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List your DedicatedCloud services",
		Run:     dedicatedcloud.ListDedicatedCloud,
	}
	dedicatedcloudCmd.AddCommand(withFilterFlag(dedicatedcloudListCmd))

	// Command to get a single DedicatedCloud
	dedicatedcloudCmd.AddCommand(&cobra.Command{
		Use:   "get <service_name>",
		Short: "Retrieve information of a specific DedicatedCloud",
		Args:  cobra.ExactArgs(1),
		Run:   dedicatedcloud.GetDedicatedCloud,
	})

	rootCmd.AddCommand(dedicatedcloudCmd)
}
