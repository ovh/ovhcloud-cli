// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/ovh/ovhcloud-cli/internal/services/cloud"
	"github.com/spf13/cobra"
)

func init() {
	cloudCmd := &cobra.Command{
		Use:   "cloud",
		Short: "Manage your projects and services in the Public Cloud universe (MKS, MPR, MRS, Object Storage...)",
	}

	cloudprojectCmd := &cobra.Command{
		Use:   "project",
		Short: "Retrieve information and manage your CloudProject services",
	}

	// Command to list CloudProject services
	cloudprojectCmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List your cloud projects",
		Run:     cloud.ListCloudProject,
	}))

	// Command to get a single CloudProject
	cloudprojectCmd.AddCommand(&cobra.Command{
		Use:   "get <project_id>",
		Short: "Retrieve information of a specific cloud project",
		Args:  cobra.ExactArgs(1),
		Run:   cloud.GetCloudProject,
	})

	editCloudProjectCmd := &cobra.Command{
		Use:   "edit <project_id>",
		Short: "Edit the given cloud project",
		Args:  cobra.ExactArgs(1),
		Run:   cloud.EditCloudProject,
	}
	editCloudProjectCmd.Flags().StringVar(&cloud.CloudProjectSpec.Description, "description", "", "Description of the project")
	editCloudProjectCmd.Flags().BoolVar(&cloud.CloudProjectSpec.ManualQuota, "manual-quota", false, "Prevent automatic quota upgrade")
	addInteractiveEditorFlag(editCloudProjectCmd)
	cloudprojectCmd.AddCommand(editCloudProjectCmd)

	initKubeCommand(cloudCmd)
	initContainerRegistryCommand(cloudCmd)
	initCloudDatabaseCommand(cloudCmd)
	initInstanceCommand(cloudCmd)
	initCloudNetworkCommand(cloudCmd)
	initCloudOperationCommand(cloudCmd)
	initCloudQuotaCommand(cloudCmd)
	initCloudRegionCommand(cloudCmd)
	initCloudSSHKeyCommand(cloudCmd)
	initCloudUserCommand(cloudCmd)
	initCloudStorageS3Command(cloudCmd)
	initCloudStorageSwiftCommand(cloudCmd)
	initCloudVolumeCommand(cloudCmd)
	initCloudRancherCommand(cloudCmd)
	initCloudReferenceCmd(cloudCmd)

	cloudCmd.AddCommand(cloudprojectCmd)
	rootCmd.AddCommand(cloudCmd)
}
