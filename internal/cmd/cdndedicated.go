// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/ovh/ovhcloud-cli/internal/completion"
	"github.com/ovh/ovhcloud-cli/internal/services/cdndedicated"
	"github.com/spf13/cobra"
)

func init() {
	cdndedicatedCmd := &cobra.Command{
		Use:   "cdn-dedicated",
		Short: "Manage Dedicated CDN (domains, cache)",
	}

	// Command to list CdnDedicated services
	cdndedicatedListCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List your dedicated CDN services",
		Run:     cdndedicated.ListCdnDedicated,
	}
	cdndedicatedCmd.AddCommand(withFilterFlag(cdndedicatedListCmd))

	// Command to get a single CdnDedicated
	cdndedicatedCmd.AddCommand(&cobra.Command{
		Use:               "get <service_name>",
		Short:             "Retrieve information of a specific CDN",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.ServiceList("/v1/cdn/dedicated"),
		Run:               cdndedicated.GetCdnDedicated,
	})

	rootCmd.AddCommand(cdndedicatedCmd)
}
