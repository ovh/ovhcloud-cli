// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/ovh/ovhcloud-cli/internal/services/cloud"
	"github.com/spf13/cobra"
)

func initCloudQuotaCommand(cloudCmd *cobra.Command) {
	quotaCmd := &cobra.Command{
		Use:   "quota",
		Short: "Check quotas in the given cloud project",
	}
	quotaCmd.PersistentFlags().StringVar(&cloud.CloudProject, "cloud-project", "", "Cloud project ID")

	quotaCmd.AddCommand(&cobra.Command{
		Use:   "get <region>",
		Short: "Get quotas for a specific region",
		Run:   cloud.GetCloudQuota,
		Args:  cobra.ExactArgs(1),
	})

	cloudCmd.AddCommand(quotaCmd)
}
