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
)

func ListCloudProject(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/cloud/project", "", cloudprojectColumnsToDisplay, flags.GenericFilters)
}

func GetCloudProject(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/cloud/project", args[0], cloudProjectTemplate)
}

func EditCloudProject(cmd *cobra.Command, args []string) {
	if err := common.EditResource(
		cmd,
		"/cloud/project/{serviceName}",
		fmt.Sprintf("/cloud/project/%s", url.PathEscape(args[0])),
		CloudProjectSpec,
		assets.CloudOpenapiSchema,
	); err != nil {
		display.ExitError(err.Error())
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

	projectID, err := config.GetConfigValue(flags.CliConfig, "", "default_cloud_project")
	if err != nil {
		return "", fmt.Errorf("failed to fetch default cloud project: %w", err)
	}
	if projectID == "" {
		return "", fmt.Errorf("no project ID configured, please use --cloud-project <id> or set a default cloud project in your configuration. Alternatively, you can set the OVH_CLOUD_PROJECT_SERVICE environment variable")
	}

	return url.PathEscape(projectID), nil
}

func getCloudRegionsWithFeatureAvailable(projectID string, features ...string) ([]any, error) {
	url := fmt.Sprintf("/cloud/project/%s/region", projectID)

	// List regions available in the cloud project
	regions, err := httpLib.FetchExpandedArray(url, "")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch regions: %w", err)
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
