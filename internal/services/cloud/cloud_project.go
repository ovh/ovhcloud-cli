// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cloud

import (
	_ "embed"
	"fmt"
	"net/url"
	"os"
	"slices"

	"github.com/spf13/cobra"

	"github.com/ovh/ovhcloud-cli/internal/assets"
	"github.com/ovh/ovhcloud-cli/internal/config"
	"github.com/ovh/ovhcloud-cli/internal/display"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
	"github.com/ovh/ovhcloud-cli/internal/services/common"
)

var (
	cloudprojectColumnsToDisplay = []string{"project_id", "projectName", "status", "description"}

	// Cloud project set by CLI flags
	CloudProject string

	//go:embed templates/cloud_project.tmpl
	cloudProjectTemplate string

	CloudProjectSpec struct {
		Description string `json:"description,omitempty"`
		ManualQuota bool   `json:"manualQuota"`
	}

	// ChangeContactSpec contains the parameters to change project contact
	ChangeContactSpec struct {
		ContactAdmin   string `json:"contactAdmin,omitempty"`
		ContactBilling string `json:"contactBilling,omitempty"`
		ContactTech    string `json:"contactTech,omitempty"`
	}
)

func ListCloudProject(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/v1/cloud/project", "", cloudprojectColumnsToDisplay, flags.GenericFilters)
}

func GetCloudProject(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/v1/cloud/project", args[0], cloudProjectTemplate)
}

func EditCloudProject(cmd *cobra.Command, args []string) {
	if err := common.EditResource(
		cmd,
		"/cloud/project/{serviceName}",
		fmt.Sprintf("/v1/cloud/project/%s", url.PathEscape(args[0])),
		CloudProjectSpec,
		assets.CloudOpenapiSchema,
	); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}
}

func getConfiguredCloudProject() (string, error) {
	if CloudProject != "" {
		return url.PathEscape(CloudProject), nil
	}

	// If project defined in ENV, use it
	if projectID := os.Getenv("OVH_CLOUD_PROJECT_SERVICE"); projectID != "" {
		return url.PathEscape(projectID), nil
	}

	// Use OpenStack standard environment variable if set
	if projectID := os.Getenv("OS_TENANT_ID"); projectID != "" {
		return url.PathEscape(projectID), nil
	}

	projectID, err := config.GetConfigValue(flags.CliConfig, "", "default_cloud_project")
	if err != nil {
		return "", fmt.Errorf("failed to fetch default cloud project: %w", err)
	}
	if projectID == "" {
		return "", fmt.Errorf("no project ID configured, please use --cloud-project <id> or set a default cloud project in your configuration. Alternatively, you can set the OVH_CLOUD_PROJECT_SERVICE or OS_TENANT_ID environment variable")
	}

	return url.PathEscape(projectID), nil
}

func getCloudRegionsWithFeatureAvailable(projectID string, features ...string) ([]any, error) {
	regions, err := fetchProjectRegions(projectID)
	if err != nil {
		return nil, err
	}

	// Filter regions having given feature available
	var regionIDs []any
	for _, region := range regions {
		if region["status"] != "UP" {
			continue
		}

		services := region["services"].([]any)
		for _, service := range services {
			service := service.(map[string]any)

			if slices.Contains(features, service["name"].(string)) && service["status"] == "UP" {
				regionIDs = append(regionIDs, region["name"])
				break
			}
		}
	}

	return regionIDs, nil
}

func fetchProjectRegions(projectID string) ([]map[string]any, error) {
	endpoint := fmt.Sprintf("/v1/cloud/project/%s/region", projectID)

	regions, err := httpLib.FetchExpandedArray(endpoint, "")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch regions: %w", err)
	}

	// Convert region type to a more user-friendly format
	for _, region := range regions {
		switch region["type"] {
		case "region":
			region["deploymentMode"] = "1-AZ"
		case "region-3-az":
			region["deploymentMode"] = "3-AZ"
		case "localzone":
			region["deploymentMode"] = "Local Zone"
		default:
			region["deploymentMode"] = region["type"]
		}
	}

	return regions, nil
}

// GetServiceInfo gets service information for the given cloud project
func GetServiceInfo(_ *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	var serviceInfo map[string]any
	endpoint := fmt.Sprintf("/v1/cloud/project/%s/serviceInfos", projectID)
	if err := httpLib.Client.Get(endpoint, &serviceInfo); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error fetching service info: %s", err)
		return
	}

	display.OutputObject(serviceInfo, projectID, common.ServiceInfoTemplate, &flags.OutputFormatConfig)
}

// ChangeContact changes project contacts
func ChangeContact(cmd *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	// Check if any contact flags were actually set
	if ChangeContactSpec.ContactAdmin == "" && ChangeContactSpec.ContactBilling == "" && ChangeContactSpec.ContactTech == "" {
		display.OutputInfo(&flags.OutputFormatConfig, nil, "🟠 No parameters given, nothing to change")
		return
	}

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/changeContact", projectID)

	if err := httpLib.Client.Post(endpoint, &ChangeContactSpec, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to change contact: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Contact change request submitted successfully")
}

// ConfirmTermination confirms project termination
func ConfirmTermination(cmd *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	token, _ := cmd.Flags().GetString("token")
	if token == "" {
		display.OutputError(&flags.OutputFormatConfig, "termination token is required (--token)")
		return
	}

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/confirmTermination", projectID)

	params := map[string]string{"token": token}
	if err := httpLib.Client.Post(endpoint, params, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to confirm termination: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Project termination confirmed successfully")
}

// TerminateProject initiates project termination
func TerminateProject(_ *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/terminate", projectID)

	var response map[string]any
	if err := httpLib.Client.Post(endpoint, nil, &response); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to terminate project: %s", err)
		return
	}

	if token, ok := response["token"].(string); ok && token != "" {
		display.OutputInfo(&flags.OutputFormatConfig, response, "✅ Termination initiated. Use token to confirm: %s", token)
	} else {
		display.OutputInfo(&flags.OutputFormatConfig, response, "✅ Termination initiated successfully")
	}
}

// RetainProject retains a project scheduled for termination
func RetainProject(_ *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/retain", projectID)

	if err := httpLib.Client.Post(endpoint, nil, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to retain project: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Project retained successfully")
}

// UnleashProject unleashes a project
func UnleashProject(_ *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/unleash", projectID)

	if err := httpLib.Client.Post(endpoint, nil, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to unleash project: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Project unleashed successfully")
}
