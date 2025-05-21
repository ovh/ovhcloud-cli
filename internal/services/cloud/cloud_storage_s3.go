package cloud

import (
	_ "embed"
	"fmt"
	"net/url"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/display"
	filtersLib "stash.ovh.net/api/ovh-cli/internal/filters"
	"stash.ovh.net/api/ovh-cli/internal/flags"
	httpLib "stash.ovh.net/api/ovh-cli/internal/http"
)

var (
	cloudprojectStorageS3ColumnsToDisplay = []string{"name", "region", "createdAt"}

	//go:embed templates/cloud_storage_s3.tmpl
	cloudStorageS3Template string
)

func ListCloudStorageS3(_ *cobra.Command, _ []string) {
	projectID := url.PathEscape(getConfiguredCloudProject())

	// Fetch regions with network feature available
	regions, err := getCloudRegionsWithFeatureAvailable(projectID, "storage-s3-high-perf", "storage-s3-standard")
	if err != nil {
		display.ExitError("failed to fetch regions with storage feature available: %s", err)
	}

	// Fetch gateways in all regions
	url := fmt.Sprintf("/cloud/project/%s/region", projectID)
	containers, err := httpLib.FetchObjectsParallel[[]map[string]any](url+"/%s/storage", regions, true)
	if err != nil {
		display.ExitError("failed to fetch storage containers: %s", err)
	}

	// Flatten gateways in a single array
	var allContainers []map[string]any
	for _, regionContainers := range containers {
		allContainers = append(allContainers, regionContainers...)
	}

	// Filter results
	allContainers, err = filtersLib.FilterLines(allContainers, flags.GenericFilters)
	if err != nil {
		display.ExitError("failed to filter results: %s", err)
	}

	display.RenderTable(allContainers, cloudprojectStorageS3ColumnsToDisplay, &flags.OutputFormatConfig)
}

func GetStorageS3(_ *cobra.Command, args []string) {
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
		if err := httpLib.Client.Get(url, &foundContainer); err == nil {
			break
		}
		foundContainer = nil
	}

	if foundContainer == nil {
		display.ExitError("no storage container found with given ID")
	}

	display.OutputObject(foundContainer, args[0], cloudStorageS3Template, &flags.OutputFormatConfig)
}
