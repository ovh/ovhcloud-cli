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
	poolColumnsToDisplay       = []string{"id", "name", "algorithm", "protocol", "operatingStatus", "provisioningStatus"}
	poolMemberColumnsToDisplay = []string{"id", "name", "address", "protocolPort", "weight", "operatingStatus"}

	//go:embed templates/cloud_loadbalancer_pool.tmpl
	cloudLoadbalancerPoolTemplate string

	//go:embed templates/cloud_loadbalancer_pool_member.tmpl
	cloudLoadbalancerPoolMemberTemplate string

	//go:embed parameter-samples/loadbalancer-pool-create.json
	LoadbalancerPoolCreationExample string

	//go:embed parameter-samples/loadbalancer-pool-member-create.json
	LoadbalancerPoolMemberCreationExample string

	CloudLoadbalancerPoolCreateSpec struct {
		Algorithm      string `json:"algorithm,omitempty"`
		ListenerId     string `json:"listenerId,omitempty"`
		LoadbalancerId string `json:"loadbalancerId,omitempty"`
		Name           string `json:"name,omitempty"`
		Protocol       string `json:"protocol,omitempty"`
	}

	CloudLoadbalancerPoolUpdateSpec struct {
		Algorithm string `json:"algorithm,omitempty"`
		Name      string `json:"name,omitempty"`
	}

	CloudLoadbalancerPoolMemberUpdateSpec struct {
		Name   string `json:"name,omitempty"`
		Weight int    `json:"weight,omitempty"`
	}
)

// Pool operations

func ListCloudLoadbalancerPools(_ *cobra.Command, _ []string) {
	listLoadbalancingResources("pool", poolColumnsToDisplay)
}

func GetCloudLoadbalancerPool(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	_, pool, err := locateLoadbalancingResource(projectID, "pool", args[0])
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	display.OutputObject(pool, args[0], cloudLoadbalancerPoolTemplate, &flags.OutputFormatConfig)
}

func CreateCloudLoadbalancerPool(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	region := args[0]
	endpoint := fmt.Sprintf("/v1/cloud/project/%s/region/%s/loadbalancing/pool",
		projectID, url.PathEscape(region))

	result, err := common.CreateResource(
		cmd,
		"/cloud/project/{serviceName}/region/{regionName}/loadbalancing/pool",
		endpoint,
		LoadbalancerPoolCreationExample,
		CloudLoadbalancerPoolCreateSpec,
		assets.CloudOpenapiSchema,
		nil,
	)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to create pool: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, result, "✅ Pool created successfully (ID: %s)", result["id"])
}

func EditCloudLoadbalancerPool(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	region, _, err := locateLoadbalancingResource(projectID, "pool", args[0])
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	if err := common.EditResource(
		cmd,
		"/cloud/project/{serviceName}/region/{regionName}/loadbalancing/pool/{poolId}",
		fmt.Sprintf("/v1/cloud/project/%s/region/%s/loadbalancing/pool/%s",
			projectID, url.PathEscape(region), url.PathEscape(args[0])),
		CloudLoadbalancerPoolUpdateSpec,
		assets.CloudOpenapiSchema,
	); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}
}

func DeleteCloudLoadbalancerPool(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	region, _, err := locateLoadbalancingResource(projectID, "pool", args[0])
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/region/%s/loadbalancing/pool/%s",
		projectID, url.PathEscape(region), url.PathEscape(args[0]))

	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to delete pool: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Pool %s deleted successfully", args[0])
}

// Pool Member operations

func ListCloudLoadbalancerPoolMembers(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	region, _, err := locateLoadbalancingResource(projectID, "pool", args[0])
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/region/%s/loadbalancing/pool/%s/member",
		projectID, url.PathEscape(region), url.PathEscape(args[0]))

	common.ManageListRequestNoExpand(endpoint, poolMemberColumnsToDisplay, flags.GenericFilters)
}

func GetCloudLoadbalancerPoolMember(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	region, _, err := locateLoadbalancingResource(projectID, "pool", args[0])
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/region/%s/loadbalancing/pool/%s/member/%s",
		projectID, url.PathEscape(region), url.PathEscape(args[0]), url.PathEscape(args[1]))

	var member map[string]any
	if err := httpLib.Client.Get(endpoint, &member); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch pool member: %s", err)
		return
	}

	display.OutputObject(member, args[1], cloudLoadbalancerPoolMemberTemplate, &flags.OutputFormatConfig)
}

func CreateCloudLoadbalancerPoolMember(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	region, _, err := locateLoadbalancingResource(projectID, "pool", args[0])
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/region/%s/loadbalancing/pool/%s/member",
		projectID, url.PathEscape(region), url.PathEscape(args[0]))

	result, err := common.CreateResource(
		cmd,
		"/cloud/project/{serviceName}/region/{regionName}/loadbalancing/pool/{poolId}/member",
		endpoint,
		LoadbalancerPoolMemberCreationExample,
		struct{}{},
		assets.CloudOpenapiSchema,
		nil,
	)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to create pool member: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, result, "✅ Pool member(s) created successfully")
}

func EditCloudLoadbalancerPoolMember(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	region, _, err := locateLoadbalancingResource(projectID, "pool", args[0])
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	if err := common.EditResource(
		cmd,
		"/cloud/project/{serviceName}/region/{regionName}/loadbalancing/pool/{poolId}/member/{memberId}",
		fmt.Sprintf("/v1/cloud/project/%s/region/%s/loadbalancing/pool/%s/member/%s",
			projectID, url.PathEscape(region), url.PathEscape(args[0]), url.PathEscape(args[1])),
		CloudLoadbalancerPoolMemberUpdateSpec,
		assets.CloudOpenapiSchema,
	); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}
}

func DeleteCloudLoadbalancerPoolMember(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	region, _, err := locateLoadbalancingResource(projectID, "pool", args[0])
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/region/%s/loadbalancing/pool/%s/member/%s",
		projectID, url.PathEscape(region), url.PathEscape(args[0]), url.PathEscape(args[1]))

	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to delete pool member: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Pool member %s deleted successfully", args[1])
}
