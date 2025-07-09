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
	"stash.ovh.net/api/ovh-cli/internal/services/common"
)

var (
	cloudprojectStorageS3ColumnsToDisplay = []string{"name", "region", "createdAt"}

	//go:embed templates/cloud_storage_s3.tmpl
	cloudStorageS3Template string

	StorageS3Spec struct {
		Encryption struct {
			SSEAlgorithm string `json:"sseAlgorithm,omitempty"`
		} `json:"encryption,omitzero"`
		ObjectLock struct {
			Rule struct {
				Mode   string `json:"mode,omitempty"`
				Period string `json:"period,omitempty"`
			} `json:"rule,omitzero"`
			Status string `json:"status,omitempty"`
		} `json:"objectLock,omitzero"`
		Replication struct {
			Rules []struct {
				DeleteMarkerReplication string `json:"deleteMarkerReplication,omitempty"`
				Destination             struct {
					Name         string `json:"name,omitempty"`
					Region       string `json:"region,omitempty"`
					StorageClass string `json:"storageClass,omitempty"`
				} `json:"destination,omitzero"`
				Filter struct {
					Prefix string            `json:"prefix,omitempty"`
					Tags   map[string]string `json:"tags,omitempty"`
				} `json:"filter,omitzero"`
				ID       string `json:"id,omitempty"`
				Priority int    `json:"priority,omitempty"`
				Status   string `json:"status,omitempty"`
			} `json:"rules,omitempty"`
		} `json:"replication,omitzero"`
		Tags       map[string]string `json:"tags,omitempty"`
		Versioning struct {
			Status string `json:"status,omitempty"`
		} `json:"versioning,omitzero"`
	}
)

func ListCloudStorageS3(_ *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	// Fetch regions with network feature available
	regions, err := getCloudRegionsWithFeatureAvailable(projectID, "storage-s3-high-perf", "storage-s3-standard")
	if err != nil {
		display.ExitError("failed to fetch regions with storage feature available: %s", err)
		return
	}

	// Fetch gateways in all regions
	url := fmt.Sprintf("/cloud/project/%s/region", projectID)
	containers, err := httpLib.FetchObjectsParallel[[]map[string]any](url+"/%s/storage", regions, true)
	if err != nil {
		display.ExitError("failed to fetch storage containers: %s", err)
		return
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
		return
	}

	display.RenderTable(allContainers, cloudprojectStorageS3ColumnsToDisplay, &flags.OutputFormatConfig)
}

func GetStorageS3(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	// Fetch regions with network feature available
	regions, err := getCloudRegionsWithFeatureAvailable(projectID, "storage-s3-high-perf", "storage-s3-standard")
	if err != nil {
		display.ExitError("failed to fetch regions with storage feature available: %s", err)
		return
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
		return
	}

	display.OutputObject(foundContainer, args[0], cloudStorageS3Template, &flags.OutputFormatConfig)
}

func EditStorageS3(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	// Fetch regions with network feature available
	regions, err := getCloudRegionsWithFeatureAvailable(projectID, "storage-s3-high-perf", "storage-s3-standard")
	if err != nil {
		display.ExitError("failed to fetch regions with storage feature available: %s", err)
		return
	}

	// Search for the given container in all regions
	// TODO: speed up with parallel search or by adding a required region argument
	var foundURL string
	for _, region := range regions {
		endpoint := fmt.Sprintf("/cloud/project/%s/region/%s/storage/%s",
			projectID, url.PathEscape(region.(string)), url.PathEscape(args[0]))
		if err := httpLib.Client.Get(endpoint, nil); err == nil {
			foundURL = endpoint
			break
		}
	}

	if foundURL == "" {
		display.ExitError("no storage container found with given ID")
		return
	}

	if err := common.EditResource(
		cmd,
		"/cloud/project/{serviceName}/region/{regionName}/storage/{name}",
		foundURL,
		StorageS3Spec,
		CloudOpenapiSchema,
	); err != nil {
		display.ExitError(err.Error())
		return
	}
}
