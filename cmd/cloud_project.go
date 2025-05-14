package cmd

import (
	"fmt"
	"slices"

	"github.com/spf13/cobra"

	"stash.ovh.net/api/ovh-cli/internal/config"
	"stash.ovh.net/api/ovh-cli/internal/display"
)

var (
	cloudprojectColumnsToDisplay = []string{"project_id", "projectName", "status", "description"}

	// Cloud project set by CLI flags
	cloudProject string
)

func listCloudProject(_ *cobra.Command, _ []string) {
	manageListRequest("/cloud/project", "", cloudprojectColumnsToDisplay, genericFilters)
}

func getCloudProject(_ *cobra.Command, args []string) {
	manageObjectRequest("/cloud/project", args[0], cloudprojectColumnsToDisplay[0])
}

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
		Run:   listCloudProject,
	}))

	// Command to get a single CloudProject
	cloudprojectCmd.AddCommand(&cobra.Command{
		Use:        "get",
		Short:      "Retrieve information of a specific cloud project",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"service_name"},
		Run:        getCloudProject,
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

func getConfiguredCloudProject() string {
	if cloudProject != "" {
		return cloudProject
	}

	projectID, err := config.GetConfigValue(cliConfig, "", "default_cloud_project")
	if err != nil {
		display.ExitError("failed to fetch default cloud project: %s", err)
	}
	if projectID == "" {
		display.ExitError("no project ID configured, please use --cloud-project <id> or set a default cloud project in your configuration")
	}

	return projectID
}

func getCloudRegionsWithFeatureAvailable(projectID string, features ...string) ([]any, error) {
	url := fmt.Sprintf("/cloud/project/%s/region", projectID)

	// List regions available in the cloud project
	regions, err := fetchExpandedArray(url, "")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch regions: %w", err)
	}

	// Filter regions having given feature available
	var regionIDs []any
	for _, region := range regions {
		if region["status"] != "UP" {
			continue
		}

		services := region["services"].([]any)
		for _, service := range services {
			service := service.(map[string]any)

			if slices.Contains(features, service["name"].(string)) && service["status"] == "UP" {
				regionIDs = append(regionIDs, region["name"])
				break
			}
		}
	}

	return regionIDs, nil
}
