// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cloud

import (
	_ "embed"
	"fmt"
	"net/url"

	"github.com/ovh/ovhcloud-cli/internal/display"
	filtersLib "github.com/ovh/ovhcloud-cli/internal/filters"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
	"github.com/ovh/ovhcloud-cli/internal/services/common"
	"github.com/spf13/cobra"
)

var (
	// CloudFloatingIPRegionFilter is used to filter floating IPs by region
	CloudFloatingIPRegionFilter string

	cloudFloatingIPColumnsToDisplay = []string{"id", "ip", "status", "region", "networkId", "associatedEntity"}

	//go:embed templates/cloud_floating_ip.tmpl
	cloudFloatingIPTemplate string
)

func ListFloatingIPs(_ *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	var regions []any

	if CloudFloatingIPRegionFilter != "" {
		regions = []any{CloudFloatingIPRegionFilter}
	} else {
		regions, err = getCloudRegionsWithFeatureAvailable(projectID, "network")
		if err != nil {
			display.OutputError(&flags.OutputFormatConfig, "failed to fetch regions with network feature available: %s", err)
			return
		}
	}

	baseURL := fmt.Sprintf("/v1/cloud/project/%s/region", projectID)
	floatingIPs, err := httpLib.FetchObjectsParallel[[]map[string]any](baseURL+"/%s/floatingip", regions, true)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch floating IPs: %s", err)
		return
	}

	var allFloatingIPs []map[string]any
	for i, regionFloatingIPs := range floatingIPs {
		for _, fip := range regionFloatingIPs {
			if _, ok := fip["region"]; !ok {
				fip["region"] = fmt.Sprint(regions[i])
			}
			allFloatingIPs = append(allFloatingIPs, fip)
		}
	}

	allFloatingIPs, err = filtersLib.FilterLines(allFloatingIPs, flags.GenericFilters)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to filter results: %s", err)
		return
	}

	display.RenderTable(allFloatingIPs, cloudFloatingIPColumnsToDisplay, &flags.OutputFormatConfig)
}

func GetFloatingIP(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	if CloudFloatingIPRegionFilter == "" {
		display.OutputError(&flags.OutputFormatConfig, "--region flag is required for get command")
		return
	}

	path := fmt.Sprintf("/v1/cloud/project/%s/region/%s/floatingip",
		projectID, url.PathEscape(CloudFloatingIPRegionFilter))

	common.ManageObjectRequest(path, args[0], cloudFloatingIPTemplate)
}

func DeleteFloatingIP(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	if CloudFloatingIPRegionFilter == "" {
		display.OutputError(&flags.OutputFormatConfig, "--region flag is required for delete command")
		return
	}

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/region/%s/floatingip/%s",
		projectID, url.PathEscape(CloudFloatingIPRegionFilter), url.PathEscape(args[0]))

	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to delete floating IP: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Floating IP %s deleted successfully", args[0])
}
