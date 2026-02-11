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
	"github.com/ovh/ovhcloud-cli/internal/flags"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
	"github.com/ovh/ovhcloud-cli/internal/services/common"
	"github.com/spf13/cobra"
)

var (
	healthMonitorColumnsToDisplay = []string{"id", "name", "monitorType", "poolId", "operatingStatus", "provisioningStatus"}

	//go:embed templates/cloud_loadbalancer_health_monitor.tmpl
	cloudLoadbalancerHealthMonitorTemplate string

	//go:embed parameter-samples/loadbalancer-health-monitor-create.json
	LoadbalancerHealthMonitorCreationExample string

	CloudLoadbalancerHealthMonitorCreateSpec struct {
		Delay          int    `json:"delay,omitempty"`
		MaxRetries     int    `json:"maxRetries,omitempty"`
		MaxRetriesDown int    `json:"maxRetriesDown,omitempty"`
		MonitorType    string `json:"monitorType,omitempty"`
		Name           string `json:"name,omitempty"`
		PoolId         string `json:"poolId,omitempty"`
		Timeout        int    `json:"timeout,omitempty"`
	}

	CloudLoadbalancerHealthMonitorUpdateSpec struct {
		Delay          int    `json:"delay,omitempty"`
		MaxRetries     int    `json:"maxRetries,omitempty"`
		MaxRetriesDown int    `json:"maxRetriesDown,omitempty"`
		Name           string `json:"name,omitempty"`
		Timeout        int    `json:"timeout,omitempty"`
	}
)

func ListCloudLoadbalancerHealthMonitors(_ *cobra.Command, _ []string) {
	listLoadbalancingResources("healthMonitor", healthMonitorColumnsToDisplay)
}

func GetCloudLoadbalancerHealthMonitor(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	_, hm, err := locateLoadbalancingResource(projectID, "healthMonitor", args[0])
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	display.OutputObject(hm, args[0], cloudLoadbalancerHealthMonitorTemplate, &flags.OutputFormatConfig)
}

func CreateCloudLoadbalancerHealthMonitor(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	region := args[0]
	endpoint := fmt.Sprintf("/v1/cloud/project/%s/region/%s/loadbalancing/healthMonitor",
		projectID, url.PathEscape(region))

	result, err := common.CreateResource(
		cmd,
		"/cloud/project/{serviceName}/region/{regionName}/loadbalancing/healthMonitor",
		endpoint,
		LoadbalancerHealthMonitorCreationExample,
		CloudLoadbalancerHealthMonitorCreateSpec,
		assets.CloudOpenapiSchema,
		nil,
	)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to create health monitor: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, result, "✅ Health monitor created successfully (ID: %s)", result["id"])
}

func EditCloudLoadbalancerHealthMonitor(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	region, _, err := locateLoadbalancingResource(projectID, "healthMonitor", args[0])
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	if err := common.EditResource(
		cmd,
		"/cloud/project/{serviceName}/region/{regionName}/loadbalancing/healthMonitor/{healthMonitorId}",
		fmt.Sprintf("/v1/cloud/project/%s/region/%s/loadbalancing/healthMonitor/%s",
			projectID, url.PathEscape(region), url.PathEscape(args[0])),
		CloudLoadbalancerHealthMonitorUpdateSpec,
		assets.CloudOpenapiSchema,
	); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}
}

func DeleteCloudLoadbalancerHealthMonitor(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	region, _, err := locateLoadbalancingResource(projectID, "healthMonitor", args[0])
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/region/%s/loadbalancing/healthMonitor/%s",
		projectID, url.PathEscape(region), url.PathEscape(args[0]))

	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to delete health monitor: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Health monitor %s deleted successfully", args[0])
}
