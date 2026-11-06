// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cloud

import (
	_ "embed"
	"fmt"
	"net/url"
	"time"

	"github.com/ovh/ovhcloud-cli/internal/assets"
	"github.com/ovh/ovhcloud-cli/internal/display"
	filtersLib "github.com/ovh/ovhcloud-cli/internal/filters"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
	"github.com/ovh/ovhcloud-cli/internal/services/common"
	"github.com/spf13/cobra"
)

var (
	privateEndpointColumnsToDisplay                    = []string{"id", "currentState.name name", "currentState.location.region region", "currentState.service.name service", "currentState.service.tier tier", "currentState.vipAddress vip", "resourceStatus"}
	privateEndpointServiceAvailabilityColumnsToDisplay = []string{"service", "tier", "region", "transport", "port", "bandwidth", "description"}

	//go:embed templates/cloud_private_endpoint.tmpl
	privateEndpointTemplate string

	//go:embed templates/cloud_private_endpoint_service_availability.tmpl
	privateEndpointServiceAvailabilityTemplate string

	//go:embed parameter-samples/private-endpoint-create.json
	PrivateEndpointCreateExample string

	// PrivateEndpointSpec holds the parameters used to create a private endpoint.
	PrivateEndpointSpec struct {
		TargetSpec struct {
			Name             string                    `json:"name,omitempty"`
			Description      string                    `json:"description,omitempty"`
			Location         privateEndpointLocation   `json:"location,omitzero"`
			Service          privateEndpointServiceRef `json:"service,omitzero"`
			Network          privateEndpointRef        `json:"network,omitzero"`
			Subnet           privateEndpointRef        `json:"subnet,omitzero"`
			VipAddress       string                    `json:"vipAddress,omitempty"`
			PublicDnsEnabled *bool                     `json:"publicDnsEnabled,omitempty"`
		} `json:"targetSpec"`
	}

	// PrivateEndpointEditSpec holds the editable parameters of a private endpoint.
	// Description is a pointer so that an explicit empty value clears it.
	PrivateEndpointEditSpec struct {
		TargetSpec struct {
			Name        string  `json:"name,omitempty"`
			Description *string `json:"description,omitempty"`
		} `json:"targetSpec"`
	}
)

type (
	privateEndpointLocation struct {
		Region string `json:"region,omitempty"`
	}

	privateEndpointServiceRef struct {
		Name string `json:"name,omitempty"`
		Tier string `json:"tier,omitempty"`
	}

	privateEndpointRef struct {
		Id string `json:"id,omitempty"`
	}
)

func ListCloudPrivateEndpoint(_ *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.ManageListRequestNoExpand(
		fmt.Sprintf("/v2/publicCloud/project/%s/privateEndpoint", projectID),
		privateEndpointColumnsToDisplay,
		flags.GenericFilters,
	)
}

func GetCloudPrivateEndpoint(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.ManageObjectRequest(
		fmt.Sprintf("/v2/publicCloud/project/%s/privateEndpoint", projectID),
		args[0],
		privateEndpointTemplate,
	)
}

func CreateCloudPrivateEndpoint(cmd *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v2/publicCloud/project/%s/privateEndpoint", projectID)
	privateEndpoint, err := common.CreateResource(
		cmd,
		"/publicCloud/project/{projectId}/privateEndpoint",
		endpoint,
		PrivateEndpointCreateExample,
		PrivateEndpointSpec,
		assets.CloudV2OpenapiSchema,
		[]string{"targetSpec.name", "targetSpec.location.region", "targetSpec.service.name", "targetSpec.service.tier"},
	)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to create private endpoint: %s", err)
		return
	}

	privateEndpointID, _ := privateEndpoint["id"].(string)

	// Wait for the resource to be ready if --wait flag is set
	if !flags.WaitForTask {
		display.OutputInfo(&flags.OutputFormatConfig, privateEndpoint, "✅ Private endpoint creation started successfully (id: %s)", privateEndpointID)
		return
	}

	ready, err := waitForCloudResourceReady(fmt.Sprintf("%s/%s", endpoint, url.PathEscape(privateEndpointID)), 30*time.Minute)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to wait for private endpoint creation: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, ready, "✅ Private endpoint %s created successfully", privateEndpointID)
}

func EditCloudPrivateEndpoint(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	if err := common.EditResource(
		cmd,
		"/publicCloud/project/{projectId}/privateEndpoint/{privateEndpointId}",
		fmt.Sprintf("/v2/publicCloud/project/%s/privateEndpoint/%s", projectID, url.PathEscape(args[0])),
		PrivateEndpointEditSpec,
		assets.CloudV2OpenapiSchema,
	); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
	}
}

func DeleteCloudPrivateEndpoint(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v2/publicCloud/project/%s/privateEndpoint/%s", projectID, url.PathEscape(args[0]))
	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to delete private endpoint: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Private endpoint %s is being deleted…", args[0])
}

func ListCloudPrivateEndpointServiceAvailabilities(_ *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	body, err := httpLib.FetchArray(fmt.Sprintf("/v2/publicCloud/project/%s/reference/privateEndpoint/serviceAvailability", projectID), "")
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch results: %s", err)
		return
	}

	var rows []map[string]any
	for _, service := range body {
		rows = append(rows, service.(map[string]any))
	}

	// Raw output formats (-o json, -o yaml, …) get the API response untouched.
	// The default table view is flattened to one row per (service × tier)
	// before filtering, so that --filter applies to the displayed columns.
	format := &flags.OutputFormatConfig
	if !format.IsJson() && !format.IsYaml() && !format.IsInteractive() && format.CustomFormat() == "" {
		rows = flattenPrivateEndpointServices(rows)
	}

	rows, err = filtersLib.FilterLines(rows, flags.GenericFilters)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to filter results: %s", err)
		return
	}

	display.RenderTable(rows, privateEndpointServiceAvailabilityColumnsToDisplay, format)
}

// flattenPrivateEndpointServices turns the service availability list into one
// row per (service × tier) for table display.
func flattenPrivateEndpointServices(services []map[string]any) []map[string]any {
	var rows []map[string]any
	for _, service := range services {
		tiers, _ := service["tiers"].([]any)
		if len(tiers) == 0 {
			rows = append(rows, map[string]any{
				"service":     service["name"],
				"description": service["description"],
			})
			continue
		}

		for _, item := range tiers {
			tier, ok := item.(map[string]any)
			if !ok {
				continue
			}

			row := map[string]any{
				"service":     service["name"],
				"description": service["description"],
				"tier":        tier["name"],
				"transport":   tier["transport"],
				"port":        tier["port"],
				"bandwidth":   tier["bandwidth"],
			}
			if location, ok := tier["location"].(map[string]any); ok {
				row["region"] = location["region"]
			}
			rows = append(rows, row)
		}
	}
	return rows
}

func GetCloudPrivateEndpointServiceAvailability(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.ManageObjectRequest(
		fmt.Sprintf("/v2/publicCloud/project/%s/reference/privateEndpoint/serviceAvailability", projectID),
		args[0],
		privateEndpointServiceAvailabilityTemplate,
	)
}
