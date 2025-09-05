// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/ovh/ovhcloud-cli/internal/services/cloud"
	"github.com/spf13/cobra"
)

func initCloudDatabaseCommand(cloudCmd *cobra.Command) {
	databaseCmd := &cobra.Command{
		Use:   "database",
		Short: "Manage databases in the given cloud project",
	}
	databaseCmd.PersistentFlags().StringVar(&cloud.CloudProject, "cloud-project", "", "Cloud project ID")

	databaseListCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List your databases",
		Run:     cloud.ListCloudDatabases,
	}
	databaseCmd.AddCommand(withFilterFlag(databaseListCmd))

	databaseCmd.AddCommand(&cobra.Command{
		Use:   "get <database_id>",
		Short: "Get a specific database",
		Run:   cloud.GetCloudDatabase,
		Args:  cobra.ExactArgs(1),
	})

	cloudCmd.AddCommand(databaseCmd)
}
