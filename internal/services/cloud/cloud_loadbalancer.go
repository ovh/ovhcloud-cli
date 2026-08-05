// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cloud

import (
	_ "embed"
	"fmt"
	"net/url"
	"strings"

	"github.com/google/uuid"
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
		Size        string `json:"-"`
		FlavorId    string `json:"flavorId,omitempty"`
	}

	CloudLoadbalancerCreateSpec struct {
		Size     string `json:"-"`
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

	// Listener
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

	// CloudLoadbalancerListenerLoadbalancerIDFilter filters listeners by loadbalancer ID
	CloudLoadbalancerListenerLoadbalancerIDFilter string

	// Pool
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

	CloudLoadbalancerPoolMemberCreateSpec cloudLoadbalancerPoolMemberCreateSpec

	// Health Monitor
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

	// L7 Policy & Rule
	l7PolicyColumnsToDisplay = []string{"id", "name", "action", "listenerId", "position", "operatingStatus"}
	l7RuleColumnsToDisplay   = []string{"id", "ruleType", "compareType", "value", "key", "invert"}

	//go:embed templates/cloud_loadbalancer_l7policy.tmpl
	cloudLoadbalancerL7PolicyTemplate string

	//go:embed templates/cloud_loadbalancer_l7rule.tmpl
	cloudLoadbalancerL7RuleTemplate string

	//go:embed parameter-samples/loadbalancer-l7policy-create.json
	LoadbalancerL7PolicyCreationExample string

	//go:embed parameter-samples/loadbalancer-l7rule-create.json
	LoadbalancerL7RuleCreationExample string

	CloudLoadbalancerL7PolicyCreateSpec struct {
		Action           string `json:"action,omitempty"`
		Description      string `json:"description,omitempty"`
		ListenerId       string `json:"listenerId,omitempty"`
		Name             string `json:"name,omitempty"`
		Position         int    `json:"position,omitempty"`
		RedirectHttpCode int    `json:"redirectHttpCode,omitempty"`
		RedirectPoolId   string `json:"redirectPoolId,omitempty"`
		RedirectPrefix   string `json:"redirectPrefix,omitempty"`
		RedirectUrl      string `json:"redirectUrl,omitempty"`
	}

	CloudLoadbalancerL7PolicyUpdateSpec struct {
		Action           string `json:"action,omitempty"`
		Description      string `json:"description,omitempty"`
		ListenerId       string `json:"listenerId,omitempty"`
		Name             string `json:"name,omitempty"`
		Position         int    `json:"position,omitempty"`
		RedirectHttpCode int    `json:"redirectHttpCode,omitempty"`
		RedirectPoolId   string `json:"redirectPoolId,omitempty"`
		RedirectPrefix   string `json:"redirectPrefix,omitempty"`
		RedirectUrl      string `json:"redirectUrl,omitempty"`
	}

	CloudLoadbalancerL7RuleCreateSpec struct {
		CompareType string `json:"compareType,omitempty"`
		Invert      bool   `json:"invert,omitempty"`
		Key         string `json:"key,omitempty"`
		RuleType    string `json:"ruleType,omitempty"`
		Value       string `json:"value,omitempty"`
	}

	CloudLoadbalancerL7RuleUpdateSpec struct {
		CompareType string `json:"compareType,omitempty"`
		Invert      bool   `json:"invert,omitempty"`
		Key         string `json:"key,omitempty"`
		RuleType    string `json:"ruleType,omitempty"`
		Value       string `json:"value,omitempty"`
	}

	// Log
	logSubscriptionColumnsToDisplay = []string{"subscriptionId", "kind", "streamId", "createdAt"}

	//go:embed templates/cloud_loadbalancer_log_subscription.tmpl
	cloudLoadbalancerLogSubscriptionTemplate string

	//go:embed parameter-samples/loadbalancer-log-subscription-create.json
	LoadbalancerLogSubscriptionCreationExample string

	CloudLoadbalancerLogSubscriptionCreateSpec struct {
		Kind     string `json:"kind,omitempty"`
		StreamId string `json:"streamId,omitempty"`
	}

	CloudLoadbalancerLogURLSpec struct {
		Kind string `json:"kind"`
	}
)

type cloudLoadbalancerPoolMemberCreateSpec struct {
	Address      string `json:"address,omitempty"`
	Name         string `json:"name,omitempty"`
	ProtocolPort int    `json:"protocolPort,omitempty"`
	Weight       int    `json:"weight,omitempty"`
}

// ---------------------------------------------------------------------------
// Loadbalancer helpers
// ---------------------------------------------------------------------------

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

// resolveLoadbalancerSize resolves a size name (e.g. "small", "medium") to its
// flavor UUID for the given region. If the value is already a UUID it is returned as-is.
func resolveLoadbalancerSize(projectID, region, size string) (string, error) {
	if size == "" {
		return "", nil
	}
	if _, err := uuid.Parse(size); err == nil {
		return size, nil
	}

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/region/%s/loadbalancing/flavor",
		projectID, url.PathEscape(region))

	var flavors []map[string]any
	if err := httpLib.Client.Get(endpoint, &flavors); err != nil {
		return "", fmt.Errorf("failed to fetch loadbalancer flavors: %w", err)
	}

	for _, f := range flavors {
		if name, ok := f["name"].(string); ok && strings.EqualFold(name, size) {
			if id, ok := f["id"].(string); ok {
				return id, nil
			}
		}
	}

	var available []string
	for _, f := range flavors {
		if name, ok := f["name"].(string); ok {
			available = append(available, name)
		}
	}

	return "", fmt.Errorf("unknown loadbalancer size %q, available sizes: %s", size, strings.Join(available, ", "))
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

// ---------------------------------------------------------------------------
// Loadbalancer CRUD
// ---------------------------------------------------------------------------

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

	// Resolve --size name to flavor UUID
	if CloudLoadbalancerUpdateSpec.Size != "" {
		flavorID, err := resolveLoadbalancerSize(projectID, region, CloudLoadbalancerUpdateSpec.Size)
		if err != nil {
			display.OutputError(&flags.OutputFormatConfig, "%s", err)
			return
		}
		CloudLoadbalancerUpdateSpec.FlavorId = flavorID
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

	// Resolve --size name to flavor UUID
	if CloudLoadbalancerCreateSpec.Size != "" {
		flavorID, err := resolveLoadbalancerSize(projectID, region, CloudLoadbalancerCreateSpec.Size)
		if err != nil {
			display.OutputError(&flags.OutputFormatConfig, "%s", err)
			return
		}
		CloudLoadbalancerCreateSpec.FlavorId = flavorID
	}

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
		[]string{"floatingIpId", "ip"},
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

// ---------------------------------------------------------------------------
// Listener
// ---------------------------------------------------------------------------

func ListCloudLoadbalancerListeners(_ *cobra.Command, _ []string) {
	if CloudLoadbalancerListenerLoadbalancerIDFilter == "" {
		listLoadbalancingResources("listener", listenerColumnsToDisplay)
		return
	}

	listLoadbalancingResources(
		"listener?loadbalancerId="+url.QueryEscape(CloudLoadbalancerListenerLoadbalancerIDFilter),
		listenerColumnsToDisplay,
	)
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

// ---------------------------------------------------------------------------
// Pool
// ---------------------------------------------------------------------------

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
		[]string{"algorithm", "protocol"},
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

// ---------------------------------------------------------------------------
// Pool Member
// ---------------------------------------------------------------------------

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

	// When using CLI flags (not --from-file or --editor), wrap the single member
	// into the members array expected by the API, and POST directly.
	// For --from-file/--editor, fall back to CreateResource which handles those input methods.
	if !flags.ParametersViaEditor && flags.ParametersFile == "" &&
		(CloudLoadbalancerPoolMemberCreateSpec.Address != "" || CloudLoadbalancerPoolMemberCreateSpec.ProtocolPort != 0) {
		body := struct {
			Members []cloudLoadbalancerPoolMemberCreateSpec `json:"members"`
		}{
			Members: []cloudLoadbalancerPoolMemberCreateSpec{CloudLoadbalancerPoolMemberCreateSpec},
		}

		var result map[string]any
		if err := httpLib.Client.Post(endpoint, body, &result); err != nil {
			display.OutputError(&flags.OutputFormatConfig, "failed to create pool member: %s", err)
			return
		}

		display.OutputInfo(&flags.OutputFormatConfig, result, "✅ Pool member(s) created successfully")
		return
	}

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

// ---------------------------------------------------------------------------
// Health Monitor
// ---------------------------------------------------------------------------

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
		[]string{"delay", "maxRetries", "monitorType", "name", "poolId", "timeout"},
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

// ---------------------------------------------------------------------------
// L7 Policy
// ---------------------------------------------------------------------------

func ListCloudLoadbalancerL7Policies(_ *cobra.Command, _ []string) {
	listLoadbalancingResources("l7Policy", l7PolicyColumnsToDisplay)
}

func GetCloudLoadbalancerL7Policy(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	_, policy, err := locateLoadbalancingResource(projectID, "l7Policy", args[0])
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	display.OutputObject(policy, args[0], cloudLoadbalancerL7PolicyTemplate, &flags.OutputFormatConfig)
}

func CreateCloudLoadbalancerL7Policy(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	region := args[0]
	endpoint := fmt.Sprintf("/v1/cloud/project/%s/region/%s/loadbalancing/l7Policy",
		projectID, url.PathEscape(region))

	result, err := common.CreateResource(
		cmd,
		"/cloud/project/{serviceName}/region/{regionName}/loadbalancing/l7Policy",
		endpoint,
		LoadbalancerL7PolicyCreationExample,
		CloudLoadbalancerL7PolicyCreateSpec,
		assets.CloudOpenapiSchema,
		[]string{"action", "listenerId"},
	)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to create L7 policy: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, result, "✅ L7 policy created successfully (ID: %s)", result["id"])
}

func EditCloudLoadbalancerL7Policy(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	region, _, err := locateLoadbalancingResource(projectID, "l7Policy", args[0])
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	if err := common.EditResource(
		cmd,
		"/cloud/project/{serviceName}/region/{regionName}/loadbalancing/l7Policy/{l7PolicyId}",
		fmt.Sprintf("/v1/cloud/project/%s/region/%s/loadbalancing/l7Policy/%s",
			projectID, url.PathEscape(region), url.PathEscape(args[0])),
		CloudLoadbalancerL7PolicyUpdateSpec,
		assets.CloudOpenapiSchema,
	); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}
}

func DeleteCloudLoadbalancerL7Policy(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	region, _, err := locateLoadbalancingResource(projectID, "l7Policy", args[0])
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/region/%s/loadbalancing/l7Policy/%s",
		projectID, url.PathEscape(region), url.PathEscape(args[0]))

	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to delete L7 policy: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ L7 policy %s deleted successfully", args[0])
}

// ---------------------------------------------------------------------------
// L7 Rule
// ---------------------------------------------------------------------------

func ListCloudLoadbalancerL7Rules(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	region, _, err := locateLoadbalancingResource(projectID, "l7Policy", args[0])
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/region/%s/loadbalancing/l7Policy/%s/l7Rule",
		projectID, url.PathEscape(region), url.PathEscape(args[0]))

	common.ManageListRequestNoExpand(endpoint, l7RuleColumnsToDisplay, flags.GenericFilters)
}

func GetCloudLoadbalancerL7Rule(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	region, _, err := locateLoadbalancingResource(projectID, "l7Policy", args[0])
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/region/%s/loadbalancing/l7Policy/%s/l7Rule/%s",
		projectID, url.PathEscape(region), url.PathEscape(args[0]), url.PathEscape(args[1]))

	var rule map[string]any
	if err := httpLib.Client.Get(endpoint, &rule); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch L7 rule: %s", err)
		return
	}

	display.OutputObject(rule, args[1], cloudLoadbalancerL7RuleTemplate, &flags.OutputFormatConfig)
}

func CreateCloudLoadbalancerL7Rule(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	region, _, err := locateLoadbalancingResource(projectID, "l7Policy", args[0])
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/region/%s/loadbalancing/l7Policy/%s/l7Rule",
		projectID, url.PathEscape(region), url.PathEscape(args[0]))

	result, err := common.CreateResource(
		cmd,
		"/cloud/project/{serviceName}/region/{regionName}/loadbalancing/l7Policy/{l7PolicyId}/l7Rule",
		endpoint,
		LoadbalancerL7RuleCreationExample,
		CloudLoadbalancerL7RuleCreateSpec,
		assets.CloudOpenapiSchema,
		[]string{"compareType", "ruleType", "value"},
	)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to create L7 rule: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, result, "✅ L7 rule created successfully (ID: %s)", result["id"])
}

func EditCloudLoadbalancerL7Rule(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	region, _, err := locateLoadbalancingResource(projectID, "l7Policy", args[0])
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	if err := common.EditResource(
		cmd,
		"/cloud/project/{serviceName}/region/{regionName}/loadbalancing/l7Policy/{l7PolicyId}/l7Rule/{l7RuleId}",
		fmt.Sprintf("/v1/cloud/project/%s/region/%s/loadbalancing/l7Policy/%s/l7Rule/%s",
			projectID, url.PathEscape(region), url.PathEscape(args[0]), url.PathEscape(args[1])),
		CloudLoadbalancerL7RuleUpdateSpec,
		assets.CloudOpenapiSchema,
	); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}
}

func DeleteCloudLoadbalancerL7Rule(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	region, _, err := locateLoadbalancingResource(projectID, "l7Policy", args[0])
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/region/%s/loadbalancing/l7Policy/%s/l7Rule/%s",
		projectID, url.PathEscape(region), url.PathEscape(args[0]), url.PathEscape(args[1]))

	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to delete L7 rule: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ L7 rule %s deleted successfully", args[1])
}

// ---------------------------------------------------------------------------
// Log Subscription
// ---------------------------------------------------------------------------

func ListCloudLoadbalancerLogSubscriptions(_ *cobra.Command, args []string) {
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

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/region/%s/loadbalancing/loadbalancer/%s/log/subscription",
		projectID, url.PathEscape(region), url.PathEscape(args[0]))

	common.ManageListRequestNoExpand(endpoint, logSubscriptionColumnsToDisplay, flags.GenericFilters)
}

func GetCloudLoadbalancerLogSubscription(_ *cobra.Command, args []string) {
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

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/region/%s/loadbalancing/loadbalancer/%s/log/subscription/%s",
		projectID, url.PathEscape(region), url.PathEscape(args[0]), url.PathEscape(args[1]))

	var subscription map[string]any
	if err := httpLib.Client.Get(endpoint, &subscription); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch log subscription: %s", err)
		return
	}

	display.OutputObject(subscription, args[1], cloudLoadbalancerLogSubscriptionTemplate, &flags.OutputFormatConfig)
}

func CreateCloudLoadbalancerLogSubscription(cmd *cobra.Command, args []string) {
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

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/region/%s/loadbalancing/loadbalancer/%s/log/subscription",
		projectID, url.PathEscape(region), url.PathEscape(args[0]))

	result, err := common.CreateResource(
		cmd,
		"/cloud/project/{serviceName}/region/{regionName}/loadbalancing/loadbalancer/{loadBalancerId}/log/subscription",
		endpoint,
		LoadbalancerLogSubscriptionCreationExample,
		CloudLoadbalancerLogSubscriptionCreateSpec,
		assets.CloudOpenapiSchema,
		[]string{"kind", "streamId"},
	)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to create log subscription: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, result, "✅ Log subscription created successfully")
}

func DeleteCloudLoadbalancerLogSubscription(_ *cobra.Command, args []string) {
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

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/region/%s/loadbalancing/loadbalancer/%s/log/subscription/%s",
		projectID, url.PathEscape(region), url.PathEscape(args[0]), url.PathEscape(args[1]))

	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to delete log subscription: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Log subscription %s deleted successfully", args[1])
}

// ---------------------------------------------------------------------------
// Log URL & Kind
// ---------------------------------------------------------------------------

func GenerateCloudLoadbalancerLogURL(_ *cobra.Command, args []string) {
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

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/region/%s/loadbalancing/loadbalancer/%s/log/url",
		projectID, url.PathEscape(region), url.PathEscape(args[0]))

	var result map[string]any
	if err := httpLib.Client.Post(endpoint, CloudLoadbalancerLogURLSpec, &result); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to generate log URL: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, result, "✅ Temporary log URL generated successfully: %s", result["url"])
}

func ListCloudLoadbalancerLogKinds(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/region/%s/loadbalancing/log/kind",
		projectID, url.PathEscape(args[0]))

	var kinds []string
	if err := httpLib.Client.Get(endpoint, &kinds); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch log kinds: %s", err)
		return
	}

	display.OutputObject(map[string]any{"kinds": kinds}, "Log kinds", "", &flags.OutputFormatConfig)
}

func GetCloudLoadbalancerLogKind(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/region/%s/loadbalancing/log/kind/%s",
		projectID, url.PathEscape(args[0]), url.PathEscape(args[1]))

	var kind map[string]any
	if err := httpLib.Client.Get(endpoint, &kind); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch log kind: %s", err)
		return
	}

	display.OutputObject(kind, args[1], "", &flags.OutputFormatConfig)
}
