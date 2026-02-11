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

	//go:embed templates/cloud_loadbalancer_stats.tmpl
	cloudLoadbalancerStatsTemplate string

	//go:embed parameter-samples/loadbalancer-create.json
	LoadbalancerCreationExample string

	//go:embed parameter-samples/loadbalancer-associate-floating-ip.json
	LoadbalancerAssociateFloatingIpExample string

	//go:embed parameter-samples/loadbalancer-create-floating-ip.json
	LoadbalancerCreateFloatingIpExample string

	CloudLoadbalancerUpdateSpec struct {
		Description string `json:"description,omitempty"`
		Name        string `json:"name,omitempty"`
		FlavorId    string `json:"flavorId,omitempty"`
	}

	CloudLoadbalancerCreateSpec struct {
		FlavorId string `json:"flavorId,omitempty"`
		Name     string `json:"name,omitempty"`
		Network  struct {
			Private struct {
				FloatingIp struct {
					Id string `json:"id,omitempty"`
				} `json:"floatingIp,omitzero"`
				Gateway struct {
					Id string `json:"id,omitempty"`
				} `json:"gateway,omitzero"`
				Network struct {
					Id       string `json:"id,omitempty"`
					SubnetId string `json:"subnetId,omitempty"`
				} `json:"network,omitzero"`
			} `json:"private,omitzero"`
		} `json:"network,omitzero"`
	}

	CloudLoadbalancerAssociateFloatingIpSpec struct {
		FloatingIpId string `json:"floatingIpId,omitempty"`
		Ip           string `json:"ip,omitempty"`
	}

	CloudLoadbalancerCreateFloatingIpSpec struct {
		Ip string `json:"ip,omitempty"`
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

// locateLoadbalancingResource searches for a loadbalancing resource by type and ID across all regions.
// resourceType is e.g. "listener", "pool", "healthMonitor", "l7Policy".
func locateLoadbalancingResource(projectID, resourceType, resourceID string) (string, map[string]any, error) {
	regions, err := getCloudRegionsWithFeatureAvailable(projectID, "octavialoadbalancer")
	if err != nil {
		return "", nil, fmt.Errorf("failed to fetch regions with loadbalancer feature available: %w", err)
	}

	for _, region := range regions {
		endpoint := fmt.Sprintf("/v1/cloud/project/%s/region/%s/loadbalancing/%s/%s",
			projectID, url.PathEscape(region.(string)), resourceType, url.PathEscape(resourceID))

		var resource map[string]any
		if err := httpLib.Client.Get(endpoint, &resource); err == nil {
			return region.(string), resource, nil
		}
	}

	return "", nil, fmt.Errorf("no %s found with id %s", resourceType, resourceID)
}

// listLoadbalancingResources fetches a loadbalancing resource type across all regions in parallel.
func listLoadbalancingResources(resourceType string, columnsToDisplay []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	regions, err := getCloudRegionsWithFeatureAvailable(projectID, "octavialoadbalancer")
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch regions with loadbalancer feature available: %s", err)
		return
	}

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/region", projectID)
	containers, err := httpLib.FetchObjectsParallel[[]map[string]any](endpoint+"/%s/loadbalancing/"+resourceType, regions, true)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch %s: %s", resourceType, err)
		return
	}

	var allResources []map[string]any
	for _, regionResources := range containers {
		allResources = append(allResources, regionResources...)
	}

	allResources, err = filtersLib.FilterLines(allResources, flags.GenericFilters)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to filter results: %s", err)
		return
	}

	display.RenderTable(allResources, columnsToDisplay, &flags.OutputFormatConfig)
}

func ListCloudLoadbalancers(_ *cobra.Command, _ []string) {
	listLoadbalancingResources("loadbalancer", cloudprojectLoadbalancerColumnsToDisplay)
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

func CreateCloudLoadbalancer(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	region := args[0]
	endpoint := fmt.Sprintf("/v1/cloud/project/%s/region/%s/loadbalancing/loadbalancer",
		projectID, url.PathEscape(region))

	result, err := common.CreateResource(
		cmd,
		"/cloud/project/{serviceName}/region/{regionName}/loadbalancing/loadbalancer",
		endpoint,
		LoadbalancerCreationExample,
		CloudLoadbalancerCreateSpec,
		assets.CloudOpenapiSchema,
		[]string{"name", "flavorId", "network"},
	)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to create loadbalancer: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, result, "✅ Loadbalancer created successfully (ID: %s)", result["id"])
}

func DeleteCloudLoadbalancer(_ *cobra.Command, args []string) {
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

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/region/%s/loadbalancing/loadbalancer/%s",
		projectID, url.PathEscape(region), url.PathEscape(args[0]))

	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to delete loadbalancer: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Loadbalancer %s deleted successfully", args[0])
}

func GetCloudLoadbalancerStats(_ *cobra.Command, args []string) {
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

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/region/%s/loadbalancing/loadbalancer/%s/stats",
		projectID, url.PathEscape(region), url.PathEscape(args[0]))

	var stats map[string]any
	if err := httpLib.Client.Get(endpoint, &stats); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch loadbalancer stats: %s", err)
		return
	}

	display.OutputObject(stats, args[0], cloudLoadbalancerStatsTemplate, &flags.OutputFormatConfig)
}

func AssociateFloatingIpToLoadbalancer(cmd *cobra.Command, args []string) {
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

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/region/%s/loadbalancing/loadbalancer/%s/associateFloatingIp",
		projectID, url.PathEscape(region), url.PathEscape(args[0]))

	result, err := common.CreateResource(
		cmd,
		"/cloud/project/{serviceName}/region/{regionName}/loadbalancing/loadbalancer/{loadBalancerId}/associateFloatingIp",
		endpoint,
		LoadbalancerAssociateFloatingIpExample,
		CloudLoadbalancerAssociateFloatingIpSpec,
		assets.CloudOpenapiSchema,
		nil,
	)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to associate floating IP: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, result, "✅ Floating IP associated to loadbalancer %s successfully", args[0])
}

func CreateFloatingIpForLoadbalancer(cmd *cobra.Command, args []string) {
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

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/region/%s/loadbalancing/loadbalancer/%s/floatingIp",
		projectID, url.PathEscape(region), url.PathEscape(args[0]))

	result, err := common.CreateResource(
		cmd,
		"/cloud/project/{serviceName}/region/{regionName}/loadbalancing/loadbalancer/{loadBalancerId}/floatingIp",
		endpoint,
		LoadbalancerCreateFloatingIpExample,
		CloudLoadbalancerCreateFloatingIpSpec,
		assets.CloudOpenapiSchema,
		[]string{"ip"},
	)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to create floating IP: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, result, "✅ Floating IP created and attached to loadbalancer %s successfully", args[0])
}
