package cmd

import (
	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/services/cloud"
)

func init() {
	cloudCmd := &cobra.Command{
		Use:   "cloud",
		Short: "Manage your projects and services in the Public Cloud universe",
	}

	cloudprojectCmd := &cobra.Command{
		Use:   "project",
		Short: "Retrieve information and manage your CloudProject services",
	}

	// Command to list CloudProject services
	cloudprojectCmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:   "list",
		Short: "List your cloud projects",
		Run:   cloud.ListCloudProject,
	}))

	// Command to get a single CloudProject
	cloudprojectCmd.AddCommand(&cobra.Command{
		Use:   "get <project_id>",
		Short: "Retrieve information of a specific cloud project",
		Args:  cobra.ExactArgs(1),
		Run:   cloud.GetCloudProject,
	})

	cloudprojectCmd.AddCommand(&cobra.Command{
		Use:   "edit <project_id>",
		Short: "Edit the given cloud project",
		Run:   cloud.EditCloudProject,
	})

	initKubeCommand(cloudCmd)
	initContainerRegistryCommand(cloudCmd)
	initCloudDatabaseCommand(cloudCmd)
	initInstanceCommand(cloudCmd)
	initCloudLoadbalancerCommand(cloudCmd)
	initCloudNetworkCommand(cloudCmd)
	initCloudOperationCommand(cloudCmd)
	initCloudQuotaCommand(cloudCmd)
	initCloudRegionCommand(cloudCmd)
	initCloudSSHKeyCommand(cloudCmd)
	initCloudUserCommand(cloudCmd)
	initCloudStorageS3Command(cloudCmd)
	initCloudStorageSwiftCommand(cloudCmd)
	initCloudVolumeCommand(cloudCmd)

	cloudCmd.AddCommand(cloudprojectCmd)
	rootCmd.AddCommand(cloudCmd)
}
