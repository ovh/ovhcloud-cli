// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/ovh/ovhcloud-cli/internal/services/cloud"
	"github.com/spf13/cobra"
)

func initCloudAICommand(cloudCmd *cobra.Command) {
	aiCmd := &cobra.Command{
		Use:   "ai",
		Short: "Manage AI Endpoints settings for your cloud project",
	}
	aiCmd.PersistentFlags().StringVar(&cloud.CloudProject, "cloud-project", "", "Cloud project ID")

	// AI Endpoints authorization commands
	aiAuthorizationCmd := &cobra.Command{
		Use:   "authorization",
		Short: "Manage AI Endpoints authorization for the project",
	}

	aiAuthorizationCmd.AddCommand(&cobra.Command{
		Use:   "get",
		Short: "Get the AI Endpoints authorization status of the project",
		Run:   cloud.GetAIAuthorization,
	})

	aiAuthorizationCmd.AddCommand(&cobra.Command{
		Use:   "create",
		Short: "Authorize AI Endpoints usage on the project",
		Run:   cloud.CreateAIAuthorization,
	})

	aiCmd.AddCommand(aiAuthorizationCmd)

	cloudCmd.AddCommand(aiCmd)
}
