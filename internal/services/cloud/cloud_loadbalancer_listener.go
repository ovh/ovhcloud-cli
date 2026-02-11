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
	listenerColumnsToDisplay = []string{"id", "name", "protocol", "port", "operatingStatus", "provisioningStatus"}

	//go:embed templates/cloud_loadbalancer_listener.tmpl
	cloudLoadbalancerListenerTemplate string

	//go:embed parameter-samples/loadbalancer-listener-create.json
	LoadbalancerListenerCreationExample string

	CloudLoadbalancerListenerUpdateSpec struct {
		AllowedCidrs  []string `json:"allowedCidrs,omitempty"`
		CertificateId string   `json:"certificateId,omitempty"`
		DefaultPoolId string   `json:"defaultPoolId,omitempty"`
		Description   string   `json:"description,omitempty"`
		Name          string   `json:"name,omitempty"`
	}

	CloudLoadbalancerListenerCreateSpec struct {
		LoadbalancerId string `json:"loadbalancerId,omitempty"`
		Name           string `json:"name,omitempty"`
		Port           int    `json:"port,omitempty"`
		Protocol       string `json:"protocol,omitempty"`
	}
)

func ListCloudLoadbalancerListeners(_ *cobra.Command, _ []string) {
	listLoadbalancingResources("listener", listenerColumnsToDisplay)
}

func GetCloudLoadbalancerListener(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	_, listener, err := locateLoadbalancingResource(projectID, "listener", args[0])
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	display.OutputObject(listener, args[0], cloudLoadbalancerListenerTemplate, &flags.OutputFormatConfig)
}

func CreateCloudLoadbalancerListener(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	region := args[0]
	endpoint := fmt.Sprintf("/v1/cloud/project/%s/region/%s/loadbalancing/listener",
		projectID, url.PathEscape(region))

	result, err := common.CreateResource(
		cmd,
		"/cloud/project/{serviceName}/region/{regionName}/loadbalancing/listener",
		endpoint,
		LoadbalancerListenerCreationExample,
		CloudLoadbalancerListenerCreateSpec,
		assets.CloudOpenapiSchema,
		[]string{"loadbalancerId", "name", "port", "protocol"},
	)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to create listener: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, result, "✅ Listener created successfully (ID: %s)", result["id"])
}

func EditCloudLoadbalancerListener(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	region, _, err := locateLoadbalancingResource(projectID, "listener", args[0])
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	if err := common.EditResource(
		cmd,
		"/cloud/project/{serviceName}/region/{regionName}/loadbalancing/listener/{listenerId}",
		fmt.Sprintf("/v1/cloud/project/%s/region/%s/loadbalancing/listener/%s",
			projectID, url.PathEscape(region), url.PathEscape(args[0])),
		CloudLoadbalancerListenerUpdateSpec,
		assets.CloudOpenapiSchema,
	); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}
}

func DeleteCloudLoadbalancerListener(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	region, _, err := locateLoadbalancingResource(projectID, "listener", args[0])
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/region/%s/loadbalancing/listener/%s",
		projectID, url.PathEscape(region), url.PathEscape(args[0]))

	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to delete listener: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Listener %s deleted successfully", args[0])
}
