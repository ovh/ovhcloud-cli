// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cloud

import (
	"fmt"
	"net/url"

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
		display.ExitError(err.Error())
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
		display.ExitError(err.Error())
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
		display.ExitError(err.Error())
		return
	}
	path := fmt.Sprintf("/cloud/project/%s/capabilities/containerRegistry", projectID)

	var body []map[string]any
	if err := httpLib.Client.Get(path, &body); err != nil {
		display.ExitError("failed to fetch container registry plans: %s", err)
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
		display.ExitError("failed to filter results: %s", err)
		return
	}

	display.RenderTable(updatedBody, []string{"region", "id", "name"}, &flags.OutputFormatConfig)
}
