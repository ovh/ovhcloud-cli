// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/ovh/ovhcloud-cli/internal/assets"
	"github.com/ovh/ovhcloud-cli/internal/services/cloud"
	"github.com/spf13/cobra"
)

func initCloudQuotaCommand(cloudCmd *cobra.Command) {
	quotaCmd := &cobra.Command{
		Use:   "quota",
		Short: "Manage quotas in the given cloud project",
	}
	quotaCmd.PersistentFlags().StringVar(&cloud.CloudProject, "cloud-project", "", "Cloud project ID")

	quotaCmd.AddCommand(&cobra.Command{
		Use:   "get",
		Short: "Get the project quota",
		Run:   cloud.GetCloudQuota,
		Args:  cobra.NoArgs,
	})

	quotaEditCmd := &cobra.Command{
		Use:   "edit",
		Short: "Update the project quota (target quota profile per region)",
		Example: `  # Edit the project quota interactively (choose the target profile per region)
  ovhcloud cloud quota edit --cloud-project <project_id> --editor`,
		Run:  cloud.EditCloudQuota,
		Args: cobra.NoArgs,
	}
	addParameterFileFlags(quotaEditCmd, false, assets.CloudV2OpenapiSchema, "/publicCloud/project/{projectId}/quota", "put", cloud.CloudQuotaEditExample, nil)
	addInteractiveEditorFlag(quotaEditCmd)
	markFlagsMutuallyExclusive(quotaEditCmd, "from-file", "editor")
	quotaCmd.AddCommand(quotaEditCmd)

	cloudCmd.AddCommand(quotaCmd)
}
