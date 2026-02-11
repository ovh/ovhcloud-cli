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
)

// L7 Policy operations

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

// L7 Rule operations

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
		nil,
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
