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
	cloudprojectIPFailoverColumnsToDisplay = []string{"id", "ip", "status", "routedTo", "geoloc"}

	//go:embed templates/cloud_ip_failover.tmpl
	cloudIPFailoverTemplate string
)

func fetchCloudIPFailovers(projectID string) ([]map[string]any, error) {
	path := fmt.Sprintf("/v1/cloud/project/%s/ip/failover", projectID)

	var body []map[string]any
	if err := httpLib.Client.Get(path, &body); err != nil {
		return nil, fmt.Errorf("failed to fetch failover IPs: %w", err)
	}
	return body, nil
}

func ListCloudIPFailovers(_ *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	body, err := fetchCloudIPFailovers(projectID)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	body, err = filtersLib.FilterLines(body, flags.GenericFilters)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to filter results: %s", err)
		return
	}

	display.RenderTable(body, cloudprojectIPFailoverColumnsToDisplay, &flags.OutputFormatConfig)
}

func GetCloudIPFailover(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.ManageObjectRequest(fmt.Sprintf("/v1/cloud/project/%s/ip/failover", projectID), args[0], cloudIPFailoverTemplate)
}

func AttachCloudIPFailover(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/ip/failover/%s/attach", projectID, url.PathEscape(args[0]))

	var result map[string]any
	if err := httpLib.Client.Post(endpoint, map[string]string{"instanceId": args[1]}, &result); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to attach failover IP %q to instance %q: %s", args[0], args[1], err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Failover IP %s attached to instance %s successfully", args[0], args[1])
}
