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

// Log Subscription operations

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

// Log URL generation

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

// Log Kind operations

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
