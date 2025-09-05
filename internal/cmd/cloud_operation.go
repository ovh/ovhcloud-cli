// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/ovh/ovhcloud-cli/internal/services/cloud"
	"github.com/spf13/cobra"
)

func initCloudOperationCommand(cloudCmd *cobra.Command) {
	operationCmd := &cobra.Command{
		Use:   "operation",
		Short: "List and get operations in the given cloud project",
	}
	operationCmd.PersistentFlags().StringVar(&cloud.CloudProject, "cloud-project", "", "Cloud project ID")

	operationListCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List operations of the given project",
		Run:     cloud.ListCloudOperations,
	}
	operationCmd.AddCommand(withFilterFlag(operationListCmd))

	operationCmd.AddCommand(&cobra.Command{
		Use:   "get <operation_id>",
		Short: "Get a specific operation",
		Run:   cloud.GetCloudOperation,
		Args:  cobra.ExactArgs(1),
	})

	cloudCmd.AddCommand(operationCmd)
}
