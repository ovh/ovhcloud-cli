package cloud

import (
	_ "embed"
	"fmt"
	"net/url"
	"slices"

	"github.com/spf13/cobra"

	"stash.ovh.net/api/ovh-cli/internal/config"
	"stash.ovh.net/api/ovh-cli/internal/display"
	"stash.ovh.net/api/ovh-cli/internal/editor"
	"stash.ovh.net/api/ovh-cli/internal/flags"
	httpLib "stash.ovh.net/api/ovh-cli/internal/http"
	"stash.ovh.net/api/ovh-cli/internal/services/common"
)

var (
	cloudprojectColumnsToDisplay = []string{"project_id", "projectName", "status", "description"}

	// Cloud project set by CLI flags
	CloudProject string

	//go:embed templates/cloud_project.tmpl
	cloudProjectTemplate string

	//go:embed api-schemas/cloud.json
	cloudOpenapiSchema []byte

	//go:embed api-schemas/cloud_v2.json
	cloudV2OpenapiSchema []byte
)

func ListCloudProject(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/cloud/project", "", cloudprojectColumnsToDisplay, flags.GenericFilters)
}

func GetCloudProject(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/cloud/project", args[0], cloudProjectTemplate)
}

func EditCloudProject(_ *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("/cloud/project/%s", url.PathEscape(args[0]))
	if err := editor.EditResource(httpLib.Client, "/cloud/project/{serviceName}", endpoint, cloudOpenapiSchema); err != nil {
		display.ExitError(err.Error())
	}
}

func getConfiguredCloudProject() (string, error) {
	if CloudProject != "" {
		return url.PathEscape(CloudProject), nil
	}

	projectID, err := config.GetConfigValue(flags.CliConfig, "", "default_cloud_project")
	if err != nil {
		return "", fmt.Errorf("failed to fetch default cloud project: %w", err)
	}
	if projectID == "" {
		return "", fmt.Errorf("no project ID configured, please use --cloud-project <id> or set a default cloud project in your configuration")
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
