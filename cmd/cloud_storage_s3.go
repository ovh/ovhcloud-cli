package cmd

import (
	_ "embed"
	"fmt"
	"net/url"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/display"
	filtersLib "stash.ovh.net/api/ovh-cli/internal/filters"
)

var (
	cloudprojectStorageS3ColumnsToDisplay = []string{"name", "region", "createdAt"}

	//go:embed templates/cloud_storage_s3.tmpl
	cloudStorageS3Template string
)

func listCloudStorageS3(_ *cobra.Command, _ []string) {
	projectID := url.PathEscape(getConfiguredCloudProject())

	// Fetch regions with network feature available
	regions, err := getCloudRegionsWithFeatureAvailable(projectID, "storage-s3-high-perf", "storage-s3-standard")
	if err != nil {
		display.ExitError("failed to fetch regions with storage feature available: %s", err)
	}

	// Fetch gateways in all regions
	url := fmt.Sprintf("/cloud/project/%s/region", projectID)
	containers, err := fetchObjectsParallel[[]map[string]any](url+"/%s/storage", regions, true)
	if err != nil {
		display.ExitError("failed to fetch storage containers: %s", err)
	}

	// Flatten gateways in a single array
	var allContainers []map[string]any
	for _, regionContainers := range containers {
		allContainers = append(allContainers, regionContainers...)
	}

	// Filter results
	allContainers, err = filtersLib.FilterLines(allContainers, genericFilters)
	if err != nil {
		display.ExitError("failed to filter results: %s", err)
	}

	display.RenderTable(allContainers, cloudprojectStorageS3ColumnsToDisplay, &outputFormatConfig)
}

func getStorageS3(_ *cobra.Command, args []string) {
	projectID := url.PathEscape(getConfiguredCloudProject())

	// Fetch regions with network feature available
	regions, err := getCloudRegionsWithFeatureAvailable(projectID, "storage-s3-high-perf", "storage-s3-standard")
	if err != nil {
		display.ExitError("failed to fetch regions with storage feature available: %s", err)
	}

	// Search for the given container in all regions
	// TODO: speed up with parallel search or by adding a required region argument
	var foundContainer map[string]any
	for _, region := range regions {
		url := fmt.Sprintf("/cloud/project/%s/region/%s/storage/%s",
			projectID, url.PathEscape(region.(string)), url.PathEscape(args[0]))
		if err := client.Get(url, &foundContainer); err == nil {
			break
		}
		foundContainer = nil
	}

	if foundContainer == nil {
		display.ExitError("no storage container found with given ID")
	}

	display.OutputObject(foundContainer, args[0], cloudStorageS3Template, &outputFormatConfig)
}

func initCloudStorageS3Command(cloudCmd *cobra.Command) {
	storageS3Cmd := &cobra.Command{
		Use:   "storage-s3",
		Short: "Manage S3™* compatible storage containers in the given cloud project (* S3 is a trademark filed by Amazon Technologies,Inc. OVHcloud's service is not sponsored by, endorsed by, or otherwise affiliated with Amazon Technologies,Inc.)",
	}
	storageS3Cmd.PersistentFlags().StringVar(&cloudProject, "cloud-project", "", "Cloud project ID")

	storageS3ListCmd := &cobra.Command{
		Use:   "list",
		Short: "List S3™* compatible storage containers (* S3 is a trademark filed by Amazon Technologies,Inc. OVHcloud's service is not sponsored by, endorsed by, or otherwise affiliated with Amazon Technologies,Inc.)",
		Run:   listCloudStorageS3,
	}
	storageS3Cmd.AddCommand(withFilterFlag(storageS3ListCmd))

	storageS3Cmd.AddCommand(&cobra.Command{
		Use:   "get <container_name>",
		Short: "Get a specific S3™* compatible storage container (* S3 is a trademark filed by Amazon Technologies,Inc. OVHcloud's service is not sponsored by, endorsed by, or otherwise affiliated with Amazon Technologies,Inc.)",
		Run:   getStorageS3,
		Args:  cobra.ExactArgs(1),
	})

	cloudCmd.AddCommand(storageS3Cmd)
}
