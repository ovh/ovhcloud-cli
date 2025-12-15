// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cloud

import (
	_ "embed"
	"fmt"
	"net/url"

	"github.com/ovh/ovhcloud-cli/internal/assets"
	"github.com/ovh/ovhcloud-cli/internal/display"
	filtersLib "github.com/ovh/ovhcloud-cli/internal/filters"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
	"github.com/ovh/ovhcloud-cli/internal/services/common"
	"github.com/spf13/cobra"
)

var (
	cloudprojectLoadbalancerColumnsToDisplay = []string{"id", "name", "region", "provisioningStatus", "operatingStatus"}

	//go:embed templates/cloud_loadbalancer.tmpl
	cloudLoadbalancerTemplate string

	CloudLoadbalancerUpdateSpec struct {
		Description string `json:"description,omitempty"`
		Name        string `json:"name,omitempty"`
		FlavorId    string `json:"flavorId,omitempty"`
	}
)

func locateLoadbalancer(projectID, loadbalancerID string) (string, map[string]any, error) {
	// Fetch regions with loadbalancer feature available
	regions, err := getCloudRegionsWithFeatureAvailable(projectID, "octavialoadbalancer")
	if err != nil {
		return "", nil, fmt.Errorf("failed to fetch regions with loadbalancer feature available: %w", err)
	}

	// Search for the given loadbalancer in all regions
	for _, region := range regions {
		endpoint := fmt.Sprintf("/v1/cloud/project/%s/region/%s/loadbalancing/loadbalancer/%s",
			projectID, url.PathEscape(region.(string)), url.PathEscape(loadbalancerID))

		var loadbalancer map[string]any
		if err := httpLib.Client.Get(endpoint, &loadbalancer); err == nil {
			return region.(string), loadbalancer, nil
		}
	}

	return "", nil, fmt.Errorf("no loadbalancer found with id %s", loadbalancerID)
}

func ListCloudLoadbalancers(_ *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	// Fetch regions with loadbalancer feature available
	regions, err := getCloudRegionsWithFeatureAvailable(projectID, "octavialoadbalancer")
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch regions with loadbalancer feature available: %s", err)
		return
	}

	// Fetch loadbalancers in all regions
	endpoint := fmt.Sprintf("/v1/cloud/project/%s/region", projectID)
	containers, err := httpLib.FetchObjectsParallel[[]map[string]any](endpoint+"/%s/loadbalancing/loadbalancer", regions, true)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch loadbalancers: %s", err)
		return
	}

	// Flatten loadbalancers in a single array
	var allLoadbalancers []map[string]any
	for _, regionLoadbalancers := range containers {
		allLoadbalancers = append(allLoadbalancers, regionLoadbalancers...)
	}

	// Filter results
	allLoadbalancers, err = filtersLib.FilterLines(allLoadbalancers, flags.GenericFilters)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to filter results: %s", err)
		return
	}

	display.RenderTable(allLoadbalancers, cloudprojectLoadbalancerColumnsToDisplay, &flags.OutputFormatConfig)
}

func GetCloudLoadbalancer(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	// Find and fetch the loadbalancer
	region, lb, err := locateLoadbalancer(projectID, args[0])
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	// Fetch details about the flavor
	if flavorID, ok := lb["flavorId"].(string); ok && flavorID != "" {
		endpoint := fmt.Sprintf("/v1/cloud/project/%s/region/%s/loadbalancing/flavor/%s",
			projectID, url.PathEscape(region), url.PathEscape(flavorID))

		var flavor map[string]any
		if err := httpLib.Client.Get(endpoint, &flavor); err == nil {
			lb["flavor"] = flavor
		}
	}

	display.OutputObject(lb, args[0], cloudLoadbalancerTemplate, &flags.OutputFormatConfig)
}

func EditCloudLoadbalancer(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	region, _, err := locateLoadbalancer(projectID, args[0])
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	if err := common.EditResource(
		cmd,
		"/cloud/project/{serviceName}/region/{regionName}/loadbalancing/loadbalancer/{loadBalancerId}",
		fmt.Sprintf("/v1/cloud/project/%s/region/%s/loadbalancing/loadbalancer/%s", projectID, url.PathEscape(region), url.PathEscape(args[0])),
		CloudLoadbalancerUpdateSpec,
		assets.CloudOpenapiSchema,
	); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}
}
