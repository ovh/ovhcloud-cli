// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/ovh/ovhcloud-cli/internal/services/veeamcloudconnect"
	"github.com/spf13/cobra"
)

func init() {
	veeamcloudconnectCmd := &cobra.Command{
		Use:   "veeamcloudconnect",
		Short: "Retrieve information and manage your VeeamCloudConnect services",
	}

	// Command to list VeeamCloudConnect services
	veeamcloudconnectListCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List your VeeamCloudConnect services",
		Run:     veeamcloudconnect.ListVeeamCloudConnect,
	}
	veeamcloudconnectCmd.AddCommand(withFilterFlag(veeamcloudconnectListCmd))

	// Command to get a single VeeamCloudConnect
	veeamcloudconnectCmd.AddCommand(&cobra.Command{
		Use:   "get <service_name>",
		Short: "Retrieve information of a specific VeeamCloudConnect",
		Args:  cobra.ExactArgs(1),
		Run:   veeamcloudconnect.GetVeeamCloudConnect,
	})

	rootCmd.AddCommand(veeamcloudconnectCmd)
}
