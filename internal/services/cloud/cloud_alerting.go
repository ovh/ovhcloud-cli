// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cloud

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/ovh/ovhcloud-cli/internal/assets"
	"github.com/ovh/ovhcloud-cli/internal/display"
	filtersLib "github.com/ovh/ovhcloud-cli/internal/filters"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
	"github.com/ovh/ovhcloud-cli/internal/openapi"
	"github.com/ovh/ovhcloud-cli/internal/services/common"
	"github.com/ovh/ovhcloud-cli/internal/utils"
	"github.com/spf13/cobra"
)

var (
	cloudprojectAlertingConfigColumnsToDisplay = []string{"id", "emails", "formattedMonthlyThreshold.text threshold", "delay"}

	//go:embed templates/cloud_alerting.tmpl
	cloudAlertingConfigTemplate string

	//go:embed templates/cloud_alerting_alert.tmpl
	cloudAlertingTriggeredAlertTemplate string

	//go:embed parameter-samples/alerting-create.json
	AlertingConfigCreateExample string

	AlertingConfigSpec struct {
		Delay            int64    `json:"delay,omitempty"`
		Emails           []string `json:"emails,omitempty"`
		MonthlyThreshold int64    `json:"monthlyThreshold,omitempty"`
		Name             string   `json:"name,omitempty"`
		Service          string   `json:"service,omitempty"`
		Status           string   `json:"status,omitempty"`
	}

	AlertingConfigEditSpec struct {
		Delay            int64    `json:"delay,omitempty"`
		Emails           []string `json:"emails,omitempty"`
		MonthlyThreshold int64    `json:"monthlyThreshold,omitempty"`
		Name             string   `json:"name,omitempty"`
		Service          string   `json:"service,omitempty"`
		Status           string   `json:"status,omitempty"`
	}

	cloudprojectAlertingTriggeredAlertColumnsToDisplay = []string{"alertId", "alertDate", "emails"}
)

// ListCloudAlertingConfigs lists all billing alert configurations for a project
func ListCloudAlertingConfigs(_ *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	path := fmt.Sprintf("/v1/cloud/project/%s/alerting", projectID)
	body, err := httpLib.FetchExpandedArray(path, "")
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch alerting configurations: %s", err)
		return
	}

	body, err = filtersLib.FilterLines(body, flags.GenericFilters)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to filter results: %s", err)
		return
	}

	display.RenderTable(body, cloudprojectAlertingConfigColumnsToDisplay, &flags.OutputFormatConfig)
}

// GetCloudAlertingConfig gets a specific billing alert configuration
func GetCloudAlertingConfig(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.ManageObjectRequest(fmt.Sprintf("/v1/cloud/project/%s/alerting", projectID), args[0], cloudAlertingConfigTemplate)
}

// CreateCloudAlertingConfig creates a new billing alert configuration
func CreateCloudAlertingConfig(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	config, err := common.CreateResource(
		cmd,
		"/cloud/project/{serviceName}/alerting",
		fmt.Sprintf("/v1/cloud/project/%s/alerting", projectID),
		AlertingConfigCreateExample,
		AlertingConfigSpec,
		assets.CloudOpenapiSchema,
		[]string{"delay", "emails", "monthlyThreshold"},
	)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to create alerting configuration: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, config, "✅ Alerting configuration created successfully (id: %s)", config["id"])
}

// EditCloudAlertingConfig edits an existing billing alert configuration
func EditCloudAlertingConfig(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	// Check if any flags were provided
	if cmd.Flags().NFlag() == 0 {
		display.OutputInfo(&flags.OutputFormatConfig, nil, "🟠 No parameters given, nothing to edit")
		return
	}

	// Create object from parameters given on command line
	jsonCliParameters, err := json.Marshal(AlertingConfigEditSpec)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to prepare arguments from command line: %s", err)
		return
	}
	var cliParameters map[string]any
	if err := json.Unmarshal(jsonCliParameters, &cliParameters); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to parse arguments from command line: %s", err)
		return
	}

	// Fetch current resource
	endpoint := fmt.Sprintf("/v1/cloud/project/%s/alerting/%s", projectID, url.PathEscape(args[0]))
	var object map[string]any
	if err := httpLib.Client.Get(endpoint, &object); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error fetching resource %s: %s", endpoint, err)
		return
	}

	// Remove the "email" field if it exists (it should never be sent in PUT requests)
	delete(object, "email")

	// Merge CLI parameters with the fetched object
	if err := utils.MergeMaps(object, cliParameters); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to merge CLI parameters: %s", err)
		return
	}

	// Filter editable fields from OpenAPI spec
	editableBody, err := openapi.FilterEditableFields(
		assets.CloudOpenapiSchema,
		"/cloud/project/{serviceName}/alerting/{id}",
		"put",
		object,
	)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to extract writable properties: %s", err)
		return
	}

	// Update the resource
	if err := httpLib.Client.Put(endpoint, editableBody, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to update resource: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Resource updated successfully")
}

// DeleteCloudAlertingConfig deletes a billing alert configuration
func DeleteCloudAlertingConfig(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	path := fmt.Sprintf("/v1/cloud/project/%s/alerting/%s", projectID, url.PathEscape(args[0]))
	if err := httpLib.Client.Delete(path, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to delete alerting configuration: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Alerting configuration %s deleted successfully", args[0])
}

// ListCloudAlertingTriggeredAlerts lists all triggered alerts for a specific alert configuration
func ListCloudAlertingTriggeredAlerts(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	path := fmt.Sprintf("/v1/cloud/project/%s/alerting/%s/alert", projectID, url.PathEscape(args[0]))
	body, err := httpLib.FetchExpandedArray(path, "")
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch triggered alerts: %s", err)
		return
	}

	body, err = filtersLib.FilterLines(body, flags.GenericFilters)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to filter results: %s", err)
		return
	}

	display.RenderTable(body, cloudprojectAlertingTriggeredAlertColumnsToDisplay, &flags.OutputFormatConfig)
}

// GetCloudAlertingTriggeredAlert gets a specific triggered alert
func GetCloudAlertingTriggeredAlert(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.ManageObjectRequest(
		fmt.Sprintf("/v1/cloud/project/%s/alerting/%s/alert", projectID, url.PathEscape(args[0])),
		args[1],
		cloudAlertingTriggeredAlertTemplate,
	)
}
