// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cloud

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/ovh/ovhcloud-cli/internal/display"
	filtersLib "github.com/ovh/ovhcloud-cli/internal/filters"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
	"github.com/ovh/ovhcloud-cli/internal/services/common"
	"github.com/spf13/cobra"
)

func GetFlavors(region string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/cloud/project/%s/flavor", projectID)
	if region != "" {
		endpoint += "?region=" + url.QueryEscape(region)
	}

	common.ManageListRequestNoExpand(endpoint, []string{"id", "name", "region", "osType", "available"}, flags.GenericFilters)
}

func GetImages(region, osType string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	query := url.Values{}
	if region != "" {
		query.Add("region", region)
	}
	if osType != "" {
		query.Add("osType", osType)
	}

	endpoint := fmt.Sprintf("/cloud/project/%s/image?%s", projectID, query.Encode())

	common.ManageListRequestNoExpand(endpoint, []string{"id", "name", "region", "type", "status"}, flags.GenericFilters)
}

func ListContainerRegistryPlans(_ *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}
	path := fmt.Sprintf("/cloud/project/%s/capabilities/containerRegistry", projectID)

	var body []map[string]any
	if err := httpLib.Client.Get(path, &body); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch container registry plans: %s", err)
		return
	}

	var updatedBody []map[string]any
	for _, item := range body {
		for _, plan := range item["plans"].([]any) {
			planMap := plan.(map[string]any)
			planMap["region"] = item["regionName"]
			updatedBody = append(updatedBody, planMap)
		}
	}

	updatedBody, err = filtersLib.FilterLines(updatedBody, flags.GenericFilters)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to filter results: %s", err)
		return
	}

	display.RenderTable(updatedBody, []string{"region", "id", "name"}, &flags.OutputFormatConfig)
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

func ListDatabasesPlans(_ *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/cloud/project/%s/database/capabilities", projectID)
	var body map[string]any
	if err := httpLib.Client.Get(endpoint, &body); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch database plans: %s", err)
		return
	}

	var plans []map[string]any
	for _, plan := range body["plans"].([]any) {
		plans = append(plans, plan.(map[string]any))
	}

	plans, err = filtersLib.FilterLines(plans, flags.GenericFilters)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to filter results: %s", err)
		return
	}

	display.RenderTable(plans, []string{"name", "description", "lifecycle.status status", "backupRetention"}, &flags.OutputFormatConfig)
}

func ListDatabasesNodeFlavors(_ *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/cloud/project/%s/database/capabilities", projectID)
	var body map[string]any
	if err := httpLib.Client.Get(endpoint, &body); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch database plans: %s", err)
		return
	}

	var flavors []map[string]any
	for _, flavor := range body["flavors"].([]any) {
		flavorMap := flavor.(map[string]any)

		// Transform specifications map into human readable strings
		specs := flavorMap["specifications"].(map[string]any)
		memorySpec := specs["memory"].(map[string]any)
		storageSpec := specs["storage"].(map[string]any)
		flavorMap["memory"] = fmt.Sprintf("%s %s", memorySpec["value"], memorySpec["unit"])
		flavorMap["storage"] = fmt.Sprintf("%s %s", storageSpec["value"], storageSpec["unit"])

		flavors = append(flavors, flavorMap)
	}

	flavors, err = filtersLib.FilterLines(flavors, flags.GenericFilters)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to filter results: %s", err)
		return
	}

	display.RenderTable(flavors, []string{"name", "core", "memory", "storage"}, &flags.OutputFormatConfig)
}

func ListDatabaseEngines(_ *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/cloud/project/%s/database/capabilities", projectID)
	var body map[string]any
	if err := httpLib.Client.Get(endpoint, &body); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch database engines: %s", err)
		return
	}

	var engines []map[string]any
	for _, engine := range body["engines"].([]any) {
		engineMap := engine.(map[string]any)

		// Reformat description
		engineMap["description"] = strings.Title(engineMap["description"].(string))

		// Transform versions array into human readable string
		var versions []string
		for _, v := range engineMap["versions"].([]any) {
			versions = append(versions, v.(string))
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

	endpoint := fmt.Sprintf("/cloud/project/%s/region/%s/loadbalancing/flavor", projectID, url.PathEscape(args[0]))
	common.ManageListRequestNoExpand(endpoint, []string{"id", "name", "region"}, flags.GenericFilters)
}
