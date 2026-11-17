// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cloud

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/ovh/ovhcloud-cli/internal/assets"
	"github.com/ovh/ovhcloud-cli/internal/display"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
	"github.com/ovh/ovhcloud-cli/internal/services/common"
	"github.com/spf13/cobra"
)

var (
	cloudprojectLoadbalancerColumnsToDisplay = []string{"id", "currentState.name name", "currentState.location.region region", "currentState.provisioningStatus provisioningStatus", "currentState.operatingStatus operatingStatus", "resourceStatus"}

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
		Size       string `json:"-"`
		TargetSpec struct {
			Description string `json:"description,omitempty"`
			Flavor      struct {
				Name string `json:"name,omitempty"`
			} `json:"flavor,omitzero"`
			Name string `json:"name,omitempty"`
		} `json:"targetSpec"`
	}

	CloudLoadbalancerCreateSpec struct {
		Size       string `json:"-"`
		TargetSpec struct {
			Description string `json:"description,omitempty"`
			Flavor      struct {
				Name string `json:"name,omitempty"`
			} `json:"flavor,omitzero"`
			Location struct {
				Region string `json:"region,omitempty"`
			} `json:"location,omitzero"`
			Name    string `json:"name,omitempty"`
			Network struct {
				Id       string `json:"id,omitempty"`
				Ip       string `json:"ip,omitempty"`
				SubnetId string `json:"subnetId,omitempty"`
			} `json:"network,omitzero"`
		} `json:"targetSpec"`
	}

	CloudLoadbalancerAssociateFloatingIpSpec struct {
		FloatingIpId string `json:"floatingIpId,omitempty"`
		Ip           string `json:"ip,omitempty"`
	}

	CloudLoadbalancerCreateFloatingIpSpec struct {
		Ip string `json:"ip,omitempty"`
	}

	// Listener
	listenerColumnsToDisplay = []string{"id", "currentState.name name", "currentState.protocol protocol", "currentState.protocolPort port", "currentState.operatingStatus operatingStatus", "currentState.provisioningStatus provisioningStatus", "resourceStatus"}

	//go:embed templates/cloud_loadbalancer_listener.tmpl
	cloudLoadbalancerListenerTemplate string

	//go:embed parameter-samples/loadbalancer-listener-create.json
	LoadbalancerListenerCreationExample string

	CloudLoadbalancerListenerUpdateSpec struct {
		TargetSpec struct {
			DefaultPool struct {
				Id string `json:"id,omitempty"`
			} `json:"defaultPool,omitzero"`
			DefaultTlsContainerRef string `json:"defaultTlsContainerRef,omitempty"`
			Description            string `json:"description,omitempty"`
			Name                   string `json:"name,omitempty"`
		} `json:"targetSpec"`
	}

	CloudLoadbalancerListenerCreateSpec struct {
		TargetSpec struct {
			DefaultPool struct {
				Id string `json:"id,omitempty"`
			} `json:"defaultPool,omitzero"`
			DefaultTlsContainerRef string `json:"defaultTlsContainerRef,omitempty"`
			Description            string `json:"description,omitempty"`
			Name                   string `json:"name,omitempty"`
			Protocol               string `json:"protocol,omitempty"`
			ProtocolPort           int    `json:"protocolPort,omitempty"`
		} `json:"targetSpec"`
	}

	// Pool
	poolColumnsToDisplay       = []string{"id", "currentState.name name", "currentState.algorithm algorithm", "currentState.protocol protocol", "currentState.operatingStatus operatingStatus", "currentState.provisioningStatus provisioningStatus", "resourceStatus"}
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
		TargetSpec struct {
			Algorithm     string                  `json:"algorithm,omitempty"`
			Description   string                  `json:"description,omitempty"`
			HealthMonitor PoolHealthMonitorParams `json:"healthMonitor,omitzero"`
			Name          string                  `json:"name,omitempty"`
			Protocol      string                  `json:"protocol,omitempty"`
		} `json:"targetSpec"`
	}

	CloudLoadbalancerPoolUpdateSpec struct {
		TargetSpec struct {
			Algorithm     string                  `json:"algorithm,omitempty"`
			Description   string                  `json:"description,omitempty"`
			HealthMonitor PoolHealthMonitorParams `json:"healthMonitor,omitzero"`
			Name          string                  `json:"name,omitempty"`
		} `json:"targetSpec"`
	}

	CloudLoadbalancerPoolMemberUpdateSpec struct {
		TargetSpec struct {
			Backup bool   `json:"backup,omitempty"`
			Name   string `json:"name,omitempty"`
			Weight int    `json:"weight,omitempty"`
		} `json:"targetSpec"`
	}

	CloudLoadbalancerPoolMemberCreateSpec struct {
		TargetSpec struct {
			Address      string `json:"address,omitempty"`
			Backup       bool   `json:"backup,omitempty"`
			Name         string `json:"name,omitempty"`
			ProtocolPort int    `json:"protocolPort,omitempty"`
			Subnet       struct {
				Id string `json:"id,omitempty"`
			} `json:"subnet,omitzero"`
			Weight int `json:"weight,omitempty"`
		} `json:"targetSpec"`
	}

	// L7 Policy
	l7PolicyColumnsToDisplay = []string{"id", "currentState.name name", "currentState.action action", "currentState.position position", "currentState.operatingStatus operatingStatus", "currentState.provisioningStatus provisioningStatus", "resourceStatus"}

	//go:embed templates/cloud_loadbalancer_l7policy.tmpl
	cloudLoadbalancerL7PolicyTemplate string

	//go:embed parameter-samples/loadbalancer-l7policy-create.json
	LoadbalancerL7PolicyCreationExample string

	CloudLoadbalancerL7PolicyCreateSpec struct {
		TargetSpec l7PolicyTargetSpecParams `json:"targetSpec"`
	}

	// CloudLoadbalancerL7PolicyUpdateSpec carries no rule: the rules of a policy are
	// kept as is unless they are edited with --editor or --from-file.
	CloudLoadbalancerL7PolicyUpdateSpec struct {
		TargetSpec struct {
			Action           string `json:"action,omitempty"`
			Description      string `json:"description,omitempty"`
			Name             string `json:"name,omitempty"`
			Position         int    `json:"position,omitempty"`
			RedirectHttpCode int    `json:"redirectHttpCode,omitempty"`
			RedirectPool     struct {
				Id string `json:"id,omitempty"`
			} `json:"redirectPool,omitzero"`
			RedirectPrefix string `json:"redirectPrefix,omitempty"`
			RedirectUrl    string `json:"redirectUrl,omitempty"`
		} `json:"targetSpec"`
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

// PoolHealthMonitorParams is the health monitor nested in a pool target spec.
type PoolHealthMonitorParams struct {
	Delay          int    `json:"delay,omitempty"`
	ExpectedCodes  string `json:"expectedCodes,omitempty"`
	HttpMethod     string `json:"httpMethod,omitempty"`
	MaxRetries     int    `json:"maxRetries,omitempty"`
	MaxRetriesDown int    `json:"maxRetriesDown,omitempty"`
	Timeout        int    `json:"timeout,omitempty"`
	Type           string `json:"type,omitempty"`
	UrlPath        string `json:"urlPath,omitempty"`
}

// isSet reports whether at least one health monitor parameter was given.
func (hm PoolHealthMonitorParams) isSet() bool {
	return hm != PoolHealthMonitorParams{}
}

// missingRequired lists the flags required by the API for a health monitor
// that are neither set in hm nor already defined by the existing health monitor.
func (hm PoolHealthMonitorParams) missingRequired(existing map[string]any) []string {
	var missing []string
	has := func(flagValue bool, key string) bool {
		if flagValue {
			return true
		}
		v, ok := existing[key]
		return ok && v != nil && v != ""
	}
	if !has(hm.Type != "", "type") {
		missing = append(missing, "--health-monitor-type")
	}
	if !has(hm.Delay != 0, "delay") {
		missing = append(missing, "--health-monitor-delay")
	}
	if !has(hm.Timeout != 0, "timeout") {
		missing = append(missing, "--health-monitor-timeout")
	}
	if !has(hm.MaxRetries != 0, "maxRetries") {
		missing = append(missing, "--health-monitor-max-retries")
	}
	return missing
}

// healthMonitorMissingError builds the error returned when a health monitor would be incomplete.
// noneYet is true when the pool has no health monitor to complete, in a pool edit.
func healthMonitorMissingError(missing []string, noneYet bool) error {
	list := strings.Join(missing, ", ")
	if noneYet {
		return fmt.Errorf("this pool has no health monitor yet, %s must all be given to create one", list)
	}
	return fmt.Errorf("incomplete health monitor, the following parameters are required: %s", list)
}

// L7PolicyRuleParams is the single rule that can be given with flags when creating an L7 policy.
type L7PolicyRuleParams struct {
	CompareType string `json:"compareType,omitempty"`
	Invert      bool   `json:"invert,omitempty"`
	Key         string `json:"key,omitempty"`
	Type        string `json:"type,omitempty"`
	Value       string `json:"value,omitempty"`
}

func (r *L7PolicyRuleParams) isSet() bool {
	return *r != L7PolicyRuleParams{}
}

func (r *L7PolicyRuleParams) upperCase() {
	r.CompareType = strings.ToUpper(r.CompareType)
	r.Type = strings.ToUpper(r.Type)
}

// l7PolicyTargetSpecParams is the target spec of an L7 policy created with flags.
type l7PolicyTargetSpecParams struct {
	Action           string `json:"action,omitempty"`
	Description      string `json:"description,omitempty"`
	Name             string `json:"name,omitempty"`
	Position         int    `json:"position,omitempty"`
	RedirectHttpCode int    `json:"redirectHttpCode,omitempty"`
	RedirectPool     struct {
		Id string `json:"id,omitempty"`
	} `json:"redirectPool,omitzero"`
	RedirectPrefix string `json:"redirectPrefix,omitempty"`
	RedirectUrl    string `json:"redirectUrl,omitempty"`

	// Rule is the single rule given with flags, sent as the only element of "rules".
	Rule L7PolicyRuleParams `json:"-"`
}

// MarshalJSON sends the rule given with flags as the "rules" array expected by the API.
func (t l7PolicyTargetSpecParams) MarshalJSON() ([]byte, error) {
	type alias l7PolicyTargetSpecParams
	out := struct {
		alias
		Rules []L7PolicyRuleParams `json:"rules,omitempty"`
	}{alias: alias(t)}
	if t.Rule.isSet() {
		out.Rules = []L7PolicyRuleParams{t.Rule}
	}
	return json.Marshal(out)
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

// ---------------------------------------------------------------------------
// Loadbalancer CRUD
// ---------------------------------------------------------------------------

func ListCloudLoadbalancers(_ *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.ManageListRequestNoExpand(fmt.Sprintf("/v2/publicCloud/project/%s/loadbalancer", projectID), cloudprojectLoadbalancerColumnsToDisplay, flags.GenericFilters)
}

func GetCloudLoadbalancer(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.ManageObjectRequest(fmt.Sprintf("/v2/publicCloud/project/%s/loadbalancer", projectID), args[0], cloudLoadbalancerTemplate)
}

func EditCloudLoadbalancer(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	// API v2 references flavors by their upper case name
	CloudLoadbalancerUpdateSpec.TargetSpec.Flavor.Name = strings.ToUpper(CloudLoadbalancerUpdateSpec.Size)

	if err := common.EditResource(
		cmd,
		"/publicCloud/project/{projectId}/loadbalancer/{loadbalancerId}",
		fmt.Sprintf("/v2/publicCloud/project/%s/loadbalancer/%s", projectID, url.PathEscape(args[0])),
		CloudLoadbalancerUpdateSpec,
		assets.CloudV2OpenapiSchema,
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

	CloudLoadbalancerCreateSpec.TargetSpec.Location.Region = args[0]
	// API v2 references flavors by their upper case name
	CloudLoadbalancerCreateSpec.TargetSpec.Flavor.Name = strings.ToUpper(CloudLoadbalancerCreateSpec.Size)

	result, err := common.CreateResource(
		cmd,
		"/publicCloud/project/{projectId}/loadbalancer",
		fmt.Sprintf("/v2/publicCloud/project/%s/loadbalancer", projectID),
		LoadbalancerCreationExample,
		CloudLoadbalancerCreateSpec,
		assets.CloudV2OpenapiSchema,
		[]string{"targetSpec.name", "targetSpec.flavor.name", "targetSpec.network.subnetId"},
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

	endpoint := fmt.Sprintf("/v2/publicCloud/project/%s/loadbalancer/%s", projectID, url.PathEscape(args[0]))
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
		nil,
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

func loadbalancerListenerPath(projectID, loadbalancerID string) string {
	return fmt.Sprintf("/v2/publicCloud/project/%s/loadbalancer/%s/listener", projectID, url.PathEscape(loadbalancerID))
}

func ListCloudLoadbalancerListeners(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.ManageListRequestNoExpand(loadbalancerListenerPath(projectID, args[0]), listenerColumnsToDisplay, flags.GenericFilters)
}

func GetCloudLoadbalancerListener(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.ManageObjectRequest(loadbalancerListenerPath(projectID, args[0]), args[1], cloudLoadbalancerListenerTemplate)
}

func CreateCloudLoadbalancerListener(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	// API v2 expects upper case enum values
	CloudLoadbalancerListenerCreateSpec.TargetSpec.Protocol = strings.ToUpper(CloudLoadbalancerListenerCreateSpec.TargetSpec.Protocol)

	result, err := common.CreateResource(
		cmd,
		"/publicCloud/project/{projectId}/loadbalancer/{loadbalancerId}/listener",
		loadbalancerListenerPath(projectID, args[0]),
		LoadbalancerListenerCreationExample,
		CloudLoadbalancerListenerCreateSpec,
		assets.CloudV2OpenapiSchema,
		[]string{"targetSpec.name", "targetSpec.protocol", "targetSpec.protocolPort"},
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

	if err := common.EditResource(
		cmd,
		"/publicCloud/project/{projectId}/loadbalancer/{loadbalancerId}/listener/{listenerId}",
		fmt.Sprintf("%s/%s", loadbalancerListenerPath(projectID, args[0]), url.PathEscape(args[1])),
		CloudLoadbalancerListenerUpdateSpec,
		assets.CloudV2OpenapiSchema,
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

	endpoint := fmt.Sprintf("%s/%s", loadbalancerListenerPath(projectID, args[0]), url.PathEscape(args[1]))
	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to delete listener: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Listener %s deleted successfully", args[1])
}

// ---------------------------------------------------------------------------
// Pool
// ---------------------------------------------------------------------------

func loadbalancerPoolPath(projectID, loadbalancerID string) string {
	return fmt.Sprintf("/v2/publicCloud/project/%s/loadbalancer/%s/pool", projectID, url.PathEscape(loadbalancerID))
}

func ListCloudLoadbalancerPools(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.ManageListRequestNoExpand(loadbalancerPoolPath(projectID, args[0]), poolColumnsToDisplay, flags.GenericFilters)
}

func GetCloudLoadbalancerPool(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.ManageObjectRequest(loadbalancerPoolPath(projectID, args[0]), args[1], cloudLoadbalancerPoolTemplate)
}

func CreateCloudLoadbalancerPool(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	// API v2 expects upper case enum values
	spec := &CloudLoadbalancerPoolCreateSpec.TargetSpec
	spec.Algorithm = strings.ToUpper(spec.Algorithm)
	spec.Protocol = strings.ToUpper(spec.Protocol)
	spec.HealthMonitor.Type = strings.ToUpper(spec.HealthMonitor.Type)
	spec.HealthMonitor.HttpMethod = strings.ToUpper(spec.HealthMonitor.HttpMethod)

	// With --from-file or --editor the health monitor may be defined by the file,
	// so the flags alone cannot be checked: let the API validate the merged body.
	if spec.HealthMonitor.isSet() && flags.ParametersFile == "" && !flags.ParametersViaEditor {
		if missing := spec.HealthMonitor.missingRequired(nil); len(missing) > 0 {
			display.OutputError(&flags.OutputFormatConfig, "failed to create pool: %s", healthMonitorMissingError(missing, false))
			return
		}
	}

	result, err := common.CreateResource(
		cmd,
		"/publicCloud/project/{projectId}/loadbalancer/{loadbalancerId}/pool",
		loadbalancerPoolPath(projectID, args[0]),
		LoadbalancerPoolCreationExample,
		CloudLoadbalancerPoolCreateSpec,
		assets.CloudV2OpenapiSchema,
		[]string{"targetSpec.algorithm", "targetSpec.protocol"},
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

	// API v2 expects upper case enum values
	spec := &CloudLoadbalancerPoolUpdateSpec.TargetSpec
	spec.Algorithm = strings.ToUpper(spec.Algorithm)
	spec.HealthMonitor.Type = strings.ToUpper(spec.HealthMonitor.Type)
	spec.HealthMonitor.HttpMethod = strings.ToUpper(spec.HealthMonitor.HttpMethod)

	// Health monitor flags given: make sure the resulting health monitor is complete,
	// instead of letting the API answer with a technical 400. Skipped with --from-file
	// or --editor, where the health monitor may be defined by the file.
	if spec.HealthMonitor.isSet() && flags.ParametersFile == "" && !flags.ParametersViaEditor {
		var pool map[string]any
		poolURL := fmt.Sprintf("%s/%s", loadbalancerPoolPath(projectID, args[0]), url.PathEscape(args[1]))
		if err := httpLib.Client.Get(poolURL, &pool); err != nil {
			display.OutputError(&flags.OutputFormatConfig, "error fetching %s: %s", poolURL, err)
			return
		}

		var existing map[string]any
		if targetSpec, ok := pool["targetSpec"].(map[string]any); ok {
			existing, _ = targetSpec["healthMonitor"].(map[string]any)
		}

		if missing := spec.HealthMonitor.missingRequired(existing); len(missing) > 0 {
			display.OutputError(&flags.OutputFormatConfig, "%s", healthMonitorMissingError(missing, existing == nil))
			return
		}
	}

	if err := common.EditResource(
		cmd,
		"/publicCloud/project/{projectId}/loadbalancer/{loadbalancerId}/pool/{poolId}",
		fmt.Sprintf("%s/%s", loadbalancerPoolPath(projectID, args[0]), url.PathEscape(args[1])),
		CloudLoadbalancerPoolUpdateSpec,
		assets.CloudV2OpenapiSchema,
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

	endpoint := fmt.Sprintf("%s/%s", loadbalancerPoolPath(projectID, args[0]), url.PathEscape(args[1]))
	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to delete pool: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Pool %s deleted successfully", args[1])
}

// ---------------------------------------------------------------------------
// Pool Member
// ---------------------------------------------------------------------------

func loadbalancerPoolMemberPath(projectID, loadbalancerID, poolID string) string {
	return fmt.Sprintf("%s/%s/member", loadbalancerPoolPath(projectID, loadbalancerID), url.PathEscape(poolID))
}

func ListCloudLoadbalancerPoolMembers(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.ManageListRequestNoExpand(loadbalancerPoolMemberPath(projectID, args[0], args[1]), poolMemberColumnsToDisplay, flags.GenericFilters)
}

func GetCloudLoadbalancerPoolMember(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.ManageObjectRequest(loadbalancerPoolMemberPath(projectID, args[0], args[1]), args[2], cloudLoadbalancerPoolMemberTemplate)
}

func CreateCloudLoadbalancerPoolMember(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	result, err := common.CreateResource(
		cmd,
		"/publicCloud/project/{projectId}/loadbalancer/{loadbalancerId}/pool/{poolId}/member",
		loadbalancerPoolMemberPath(projectID, args[0], args[1]),
		LoadbalancerPoolMemberCreationExample,
		CloudLoadbalancerPoolMemberCreateSpec,
		assets.CloudV2OpenapiSchema,
		[]string{"targetSpec.address", "targetSpec.protocolPort"},
	)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to create pool member: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, result, "✅ Pool member created successfully (ID: %s)", result["id"])
}

func EditCloudLoadbalancerPoolMember(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	if err := common.EditResource(
		cmd,
		"/publicCloud/project/{projectId}/loadbalancer/{loadbalancerId}/pool/{poolId}/member/{memberId}",
		fmt.Sprintf("%s/%s", loadbalancerPoolMemberPath(projectID, args[0], args[1]), url.PathEscape(args[2])),
		CloudLoadbalancerPoolMemberUpdateSpec,
		assets.CloudV2OpenapiSchema,
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

	endpoint := fmt.Sprintf("%s/%s", loadbalancerPoolMemberPath(projectID, args[0], args[1]), url.PathEscape(args[2]))
	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to delete pool member: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Pool member %s deleted successfully", args[2])
}

// ---------------------------------------------------------------------------
// L7 Policy
// ---------------------------------------------------------------------------

func loadbalancerL7PolicyPath(projectID, loadbalancerID, listenerID string) string {
	return fmt.Sprintf("%s/%s/l7policy", loadbalancerListenerPath(projectID, loadbalancerID), url.PathEscape(listenerID))
}

func ListCloudLoadbalancerL7Policies(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.ManageListRequestNoExpand(loadbalancerL7PolicyPath(projectID, args[0], args[1]), l7PolicyColumnsToDisplay, flags.GenericFilters)
}

func GetCloudLoadbalancerL7Policy(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.ManageObjectRequest(loadbalancerL7PolicyPath(projectID, args[0], args[1]), args[2], cloudLoadbalancerL7PolicyTemplate)
}

func CreateCloudLoadbalancerL7Policy(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	// API v2 expects upper case enum values
	spec := &CloudLoadbalancerL7PolicyCreateSpec.TargetSpec
	spec.Action = strings.ToUpper(spec.Action)
	spec.Rule.upperCase()

	result, err := common.CreateResource(
		cmd,
		"/publicCloud/project/{projectId}/loadbalancer/{loadbalancerId}/listener/{listenerId}/l7policy",
		loadbalancerL7PolicyPath(projectID, args[0], args[1]),
		LoadbalancerL7PolicyCreationExample,
		CloudLoadbalancerL7PolicyCreateSpec,
		assets.CloudV2OpenapiSchema,
		[]string{"targetSpec.action"},
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

	// API v2 expects upper case enum values. Rules are only editable with --editor or --from-file.
	CloudLoadbalancerL7PolicyUpdateSpec.TargetSpec.Action = strings.ToUpper(CloudLoadbalancerL7PolicyUpdateSpec.TargetSpec.Action)

	if err := common.EditResource(
		cmd,
		"/publicCloud/project/{projectId}/loadbalancer/{loadbalancerId}/listener/{listenerId}/l7policy/{l7PolicyId}",
		fmt.Sprintf("%s/%s", loadbalancerL7PolicyPath(projectID, args[0], args[1]), url.PathEscape(args[2])),
		CloudLoadbalancerL7PolicyUpdateSpec,
		assets.CloudV2OpenapiSchema,
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

	endpoint := fmt.Sprintf("%s/%s", loadbalancerL7PolicyPath(projectID, args[0], args[1]), url.PathEscape(args[2]))
	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to delete L7 policy: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ L7 policy %s deleted successfully", args[2])
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
