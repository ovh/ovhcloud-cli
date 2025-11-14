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

	// Datacenter commands
	dedicatedcloudDatacenterCmd := &cobra.Command{
		Use:   "datacenter",
		Short: "Manage datacenters of a DedicatedCloud",
	}
	dedicatedcloudCmd.AddCommand(dedicatedcloudDatacenterCmd)

	dedicatedcloudDatacenterCmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:     "list <service_name>",
		Aliases: []string{"ls"},
		Short:   "List datacenters of a DedicatedCloud",
		Args:    cobra.ExactArgs(1),
		Run:     dedicatedcloud.ListDatacenter,
	}))

	datacenterGetCmd := &cobra.Command{
		Use:   "get <service_name> <datacenter_id>",
		Short: "Get information about a specific datacenter of a DedicatedCloud",
		Args:  cobra.ExactArgs(2),
		Run:   dedicatedcloud.GetDatacenter,
	}
	dedicatedcloudDatacenterCmd.AddCommand(datacenterGetCmd)

	dedicatedcloudDatacenterCmd.AddCommand(&cobra.Command{
		Use:   "hosts <service_name> <datacenter_id>",
		Short: "Get hosts information for a specific datacenter",
		Args:  cobra.ExactArgs(2),
		Run:   dedicatedcloud.GetDatacenterHosts,
	})

	dedicatedcloudDatacenterCmd.AddCommand(&cobra.Command{
		Use:   "filers <service_name> <datacenter_id>",
		Short: "Get filers information for a specific datacenter",
		Args:  cobra.ExactArgs(2),
		Run:   dedicatedcloud.GetDatacenterFilers,
	})

	dedicatedcloudDatacenterCmd.AddCommand(&cobra.Command{
		Use:   "clusters <service_name> <datacenter_id>",
		Short: "Get clusters information for a specific datacenter",
		Args:  cobra.ExactArgs(2),
		Run:   dedicatedcloud.GetDatacenterClusters,
	})

	rootCmd.AddCommand(dedicatedcloudCmd)
}
