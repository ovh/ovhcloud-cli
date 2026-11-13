// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cloud

import (
	"fmt"
	"net/url"
	"slices"
	"strings"

	"github.com/ovh/ovhcloud-cli/internal/display"
	filtersLib "github.com/ovh/ovhcloud-cli/internal/filters"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
	"github.com/ovh/ovhcloud-cli/internal/services/common"
	"github.com/spf13/cobra"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// instanceReferenceV2Endpoint returns the v2 endpoint listing the instance
// flavors or images ("flavor" or "image") of the project, optionally
// restricted to one region.
func instanceReferenceV2Endpoint(projectID, kind, region string) string {
	endpoint := fmt.Sprintf("/v2/publicCloud/project/%s/reference/instance/%s", projectID, kind)
	if region != "" {
		endpoint += "?region=" + url.QueryEscape(region)
	}
	return endpoint
}

func GetFlavors(region string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.ManageListRequestNoExpand(instanceReferenceV2Endpoint(projectID, "flavor", region),
		[]string{"id", "name", "location.region region", "vcpus", "ram", "disk"}, flags.GenericFilters)
}

// imageOSTypes maps the --os-type values to the osType of API v2 images.
var imageOSTypes = map[string]string{
	"linux":           "LINUX",
	"windows":         "WINDOWS",
	"baremetal-linux": "BAREMETAL_LINUX",
}

func GetImages(region, osType string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	filters := flags.GenericFilters
	if osType != "" {
		apiOSType, ok := imageOSTypes[osType]
		if !ok {
			display.OutputError(&flags.OutputFormatConfig, "invalid OS type %q: use linux, windows or baremetal-linux", osType)
			return
		}
		// API v2 has no osType query parameter: filter the listing
		filters = append(slices.Clone(filters), fmt.Sprintf("osType==%q", apiOSType))
	}

	common.ManageListRequestNoExpand(instanceReferenceV2Endpoint(projectID, "image", region),
		[]string{"id", "name", "location.region region", "osType", "visibility", "status"}, filters)
}

func ListContainerRegistryPlans(_ *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}
	path := fmt.Sprintf("/v1/cloud/project/%s/capabilities/containerRegistry", projectID)

	var body []map[string]any
	if err := httpLib.Client.Get(path, &body); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch container registry plans: %s", err)
		return
	}

	// Deduplicate plans by ID and format storage with bytefmt
	seenPlans := make(map[string]bool)
	var updatedBody []map[string]any
	for _, item := range body {
		for _, plan := range item["plans"].([]any) {
			planMap := plan.(map[string]any)
			planID := planMap["id"].(string)

			// Skip if we've already seen this plan
			if seenPlans[planID] {
				continue
			}
			seenPlans[planID] = true

			formatContainerRegistryPlans(planMap)

			updatedBody = append(updatedBody, planMap)
		}
	}

	updatedBody, err = filtersLib.FilterLines(updatedBody, flags.GenericFilters)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to filter results: %s", err)
		return
	}

	display.RenderTable(updatedBody, []string{"id", "name", "vulnerability", "imageStorage", "parallelRequest"}, &flags.OutputFormatConfig)
}

func ListContainerRegistryRegions(_ *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}
	path := fmt.Sprintf("/v1/cloud/project/%s/capabilities/containerRegistry", projectID)

	var body []map[string]any
	if err := httpLib.Client.Get(path, &body); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch container registry regions: %s", err)
		return
	}

	var regions []map[string]any
	for _, item := range body {
		regionName := item["regionName"].(string)
		regionType := item["regionType"].(string)

		// Remove "REGION-" prefix from regionType
		regionType = strings.TrimPrefix(regionType, "REGION-")

		regions = append(regions, map[string]any{
			"name": regionName,
			"type": regionType,
		})
	}

	regions, err = filtersLib.FilterLines(regions, flags.GenericFilters)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to filter results: %s", err)
		return
	}

	display.RenderTable(regions, []string{"name", "type"}, &flags.OutputFormatConfig)
}

func ListRancherAvailableVersions(cmd *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	serviceID, err := cmd.Flags().GetString("rancher-id")
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to get 'rancher-id' flag: %s", err)
		return
	}

	endpoint := fmt.Sprintf("/v2/publicCloud/project/%s/reference/rancher/version", projectID)
	if serviceID != "" {
		endpoint = fmt.Sprintf("/v2/publicCloud/project/%s/rancher/%s/capabilities/version", projectID, serviceID)
	}

	common.ManageListRequestNoExpand(endpoint, []string{"name", "status", "message"}, flags.GenericFilters)
}

func ListRancherAvailablePlans(cmd *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	serviceID, err := cmd.Flags().GetString("rancher-id")
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to get 'rancher-id' flag: %s", err)
		return
	}

	endpoint := fmt.Sprintf("/v2/publicCloud/project/%s/reference/rancher/plan", projectID)
	if serviceID != "" {
		endpoint = fmt.Sprintf("/v2/publicCloud/project/%s/rancher/%s/capabilities/plan", projectID, serviceID)
	}

	common.ManageListRequestNoExpand(endpoint, []string{"name", "status", "message"}, flags.GenericFilters)
}

func ListManagedDatabasePlans(_ *cobra.Command, _ []string) {
	listDatabaseAnalyticsPlans(managedDatabaseCategory)
}

func ListManagedDatabaseNodeFlavors(_ *cobra.Command, _ []string) {
	listDatabaseAnalyticsFlavors(managedDatabaseCategory)
}

func ListManagedDatabaseEngines(_ *cobra.Command, _ []string) {
	listDatabaseAnalyticsEngines(managedDatabaseCategory, "DATABASE")
}

func ListManagedAnalyticsPlans(_ *cobra.Command, _ []string) {
	listDatabaseAnalyticsPlans(managedAnalyticsCategory)
}

func ListManagedAnalyticsNodeFlavors(_ *cobra.Command, _ []string) {
	listDatabaseAnalyticsFlavors(managedAnalyticsCategory)
}

func ListManagedAnalyticsEngines(_ *cobra.Command, _ []string) {
	listDatabaseAnalyticsEngines(managedAnalyticsCategory, "ANALYTICS")
}

func listDatabaseAnalyticsPlans(category string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.ManageListRequestNoExpand(
		databaseAnalyticsEndpoint(projectID, category, "capabilities", "plans"),
		[]string{"name", "description", "lifecycle.status status", "backupRetention"},
		flags.GenericFilters,
	)
}

func listDatabaseAnalyticsFlavors(category string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	body, err := httpLib.FetchArray(databaseAnalyticsEndpoint(projectID, category, "capabilities", "flavors"), "")
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch node flavors: %s", err)
		return
	}

	var flavors []map[string]any
	for _, flavor := range body {
		flavorMap := flavor.(map[string]any)

		// Transform memory specification into a human readable string
		memorySpec := flavorMap["specifications"].(map[string]any)["memory"].(map[string]any)
		flavorMap["memory"] = fmt.Sprintf("%s %s", memorySpec["value"], memorySpec["unit"])

		flavors = append(flavors, flavorMap)
	}

	flavors, err = filtersLib.FilterLines(flavors, flags.GenericFilters)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to filter results: %s", err)
		return
	}

	display.RenderTable(flavors, []string{"name", "specifications.vcpus vcpus", "memory"}, &flags.OutputFormatConfig)
}

// listDatabaseAnalyticsEngines lists the engines of the given category (DATABASE
// or ANALYTICS): the capabilities of both categories list every engine.
func listDatabaseAnalyticsEngines(category, engineCategory string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	body, err := httpLib.FetchArray(databaseAnalyticsEndpoint(projectID, category, "capabilities", "engines"), "")
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch engines: %s", err)
		return
	}

	var engines []map[string]any
	for _, engine := range body {
		engineMap := engine.(map[string]any)
		if engineMap["category"] != engineCategory {
			continue
		}

		// Reformat description
		engineMap["description"] = cases.Title(language.English).String(engineMap["description"].(string))

		// Transform versions array into human readable string
		var versions []string
		for _, v := range engineMap["versions"].([]any) {
			versions = append(versions, v.(map[string]any)["name"].(string))
		}
		engineMap["versions"] = strings.Join(versions, " | ")

		engines = append(engines, engineMap)
	}

	engines, err = filtersLib.FilterLines(engines, flags.GenericFilters)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to filter results: %s", err)
		return
	}

	display.RenderTable(engines, []string{"name", "description", "category", "versions", "defaultVersion"}, &flags.OutputFormatConfig)
}

func ListLoadbalancerFlavors(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	// API v2 lists the flavors of every region, with an optional region filter
	endpoint := fmt.Sprintf("/v2/publicCloud/project/%s/reference/loadbalancer/flavor", projectID)
	if len(args) > 0 && args[0] != "" {
		endpoint += "?region=" + url.QueryEscape(args[0])
	}

	common.ManageListRequestNoExpand(endpoint, loadbalancerFlavorColumns, flags.GenericFilters)
}

// loadbalancerFlavorColumns are the columns displayed for the loadbalancer flavors.
var loadbalancerFlavorColumns = []string{"name", "location.region region", "location.regionActivated regionActivated", "bandwidth", "connectionsPerSecond", "requestsPerSecond"}

// GetLoadbalancerFlavor displays the loadbalancer flavor with the given name.
// API v2 has no route to get a single flavor, so the list is fetched, restricted
// to the given region if any, and filtered on the name. Without a region, a flavor
// sold in several regions is displayed once per region.
func GetLoadbalancerFlavor(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	// Arguments: [region] <flavor_name>
	region, name := "", args[0]
	if len(args) == 2 {
		region, name = args[0], args[1]
	}

	endpoint := fmt.Sprintf("/v2/publicCloud/project/%s/reference/loadbalancer/flavor", projectID)
	if region != "" {
		endpoint += "?region=" + url.QueryEscape(region)
	}

	body, err := httpLib.FetchArray(endpoint, "")
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch loadbalancer flavors: %s", err)
		return
	}

	var matches []map[string]any
	for _, item := range body {
		flavor, ok := item.(map[string]any)
		if !ok {
			continue
		}
		if flavorName, _ := flavor["name"].(string); strings.EqualFold(flavorName, name) {
			matches = append(matches, flavor)
		}
	}

	switch len(matches) {
	case 0:
		if region != "" {
			display.OutputError(&flags.OutputFormatConfig, "loadbalancer flavor %q not found in region %s", name, region)
		} else {
			display.OutputError(&flags.OutputFormatConfig, "loadbalancer flavor %q not found", name)
		}
	case 1:
		display.OutputObject(matches[0], name, "", &flags.OutputFormatConfig)
	default:
		display.RenderTable(matches, loadbalancerFlavorColumns, &flags.OutputFormatConfig)
	}
}
