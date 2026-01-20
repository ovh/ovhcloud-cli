// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cloud

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/ovh/ovhcloud-cli/internal/assets"
	"github.com/ovh/ovhcloud-cli/internal/display"
	filtersLib "github.com/ovh/ovhcloud-cli/internal/filters"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
	"github.com/ovh/ovhcloud-cli/internal/services/common"
	"github.com/spf13/cobra"
)

var (
	cloudprojectContainerRegistryColumnsToDisplay = []string{"id", "name", "region", "plan.name plan", "deploymentMode", "version", "status"}

	cloudprojectContainerRegistryUsersColumnsToDisplay = []string{"id", "user", "email"}

	//go:embed templates/cloud_container_registry.tmpl
	cloudContainerRegistryTemplate string

	//go:embed templates/cloud_container_registry_user.tmpl
	cloudContainerRegistryUserTemplate string

	//go:embed parameter-samples/container-registry-create.json
	CloudContainerRegistryCreateSample string

	//go:embed parameter-samples/container-registry-user-create.json
	CloudContainerRegistryUserCreateSample string

	// CloudContainerRegistryName is used to edit the container registry
	CloudContainerRegistryName string

	CloudContainerRegistrySpec struct {
		Name   string `json:"name,omitempty"`
		PlanID string `json:"planID,omitempty"`
		Region string `json:"region,omitempty"`
	}

	CloudContainerRegistryUserSpec struct {
		Email string `json:"email,omitempty"`
		Login string `json:"login,omitempty"`
	}
)

func ListContainerRegistries(_ *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	// Fetch registries
	endpoint := fmt.Sprintf("/v1/cloud/project/%s/containerRegistry", projectID)
	body, err := httpLib.FetchArray(endpoint, "")
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch results: %s", err)
		return
	}

	// Fetch cloud project regions
	regions, err := fetchProjectRegions(projectID)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	var objects []map[string]any
	for _, object := range body {
		objMap := object.(map[string]any)

		// Fetch plan details for each registry
		var plan map[string]any
		if err := httpLib.Client.Get(fmt.Sprintf("%s/%s/plan", endpoint, url.PathEscape(objMap["id"].(string))), &plan); err != nil {
			display.OutputError(&flags.OutputFormatConfig, "error fetching plan details: %s", err)
			return
		}
		objMap["plan"] = plan

		// Find region deployment mode
		for _, region := range regions {
			if region["name"] == objMap["region"] {
				objMap["deploymentMode"] = region["deploymentMode"]
				break
			}
		}

		objects = append(objects, objMap)
	}

	objects, err = filtersLib.FilterLines(objects, flags.GenericFilters)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to filter results: %s", err)
		return
	}

	display.RenderTable(objects, cloudprojectContainerRegistryColumnsToDisplay, &flags.OutputFormatConfig)
}

func GetContainerRegistry(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	// Fetch registry details
	endpoint := fmt.Sprintf("/v1/cloud/project/%s/containerRegistry/%s", projectID, url.PathEscape(args[0]))
	var object map[string]any
	if err := httpLib.Client.Get(endpoint, &object); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error fetching %s: %s", endpoint, err)
		return
	}

	// Fetch plan details
	var plan map[string]any
	if err := httpLib.Client.Get(endpoint+"/plan", &plan); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error fetching plan details: %s", err)
		return
	}
	object["plan"] = plan

	// Calculate and add usage information
	planLimits := plan["registryLimits"].(map[string]any)

	usedFloat, err := object["size"].(json.Number).Float64()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error parsing used storage: %s", err)
		return
	}
	availableFloat, err := planLimits["imageStorage"].(json.Number).Float64()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error parsing available storage: %s", err)
		return
	}
	object["usage"] = map[string]any{
		"used":      usedFloat,
		"available": availableFloat,
	}

	display.OutputObject(object, args[0], cloudContainerRegistryTemplate, &flags.OutputFormatConfig)
}

func EditContainerRegistry(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	if err := common.EditResource(
		cmd,
		"/cloud/project/{serviceName}/containerRegistry/{registryID}",
		fmt.Sprintf("/v1/cloud/project/%s/containerRegistry/%s", projectID, url.PathEscape(args[0])),
		map[string]any{"name": CloudContainerRegistryName},
		assets.CloudOpenapiSchema,
	); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}
}

func CreateContainerRegistry(cmd *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	registry, err := common.CreateResource(
		cmd,
		"/cloud/project/{serviceName}/containerRegistry",
		fmt.Sprintf("/v1/cloud/project/%s/containerRegistry", projectID),
		CloudContainerRegistryCreateSample,
		CloudContainerRegistrySpec,
		assets.CloudOpenapiSchema,
		[]string{"name", "region"},
	)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, registry, "✅ Container registry '%s' created successfully", registry["id"])
}

func DeleteContainerRegistry(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/containerRegistry/%s", projectID, url.PathEscape(args[0]))
	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to delete container registry: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Container registry deleted successfully")
}

func ListContainerRegistryUsers(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.ManageListRequestNoExpand(fmt.Sprintf("/v1/cloud/project/%s/containerRegistry/%s/users", projectID, url.PathEscape(args[0])), cloudprojectContainerRegistryUsersColumnsToDisplay, flags.GenericFilters)
}

func GetContainerRegistryUser(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.ManageObjectRequest(fmt.Sprintf("/v1/cloud/project/%s/containerRegistry/%s/users", projectID, url.PathEscape(args[0])), args[1], cloudContainerRegistryUserTemplate)
}

func CreateContainerRegistryUser(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	user, err := common.CreateResource(
		cmd,
		"/cloud/project/{serviceName}/containerRegistry/{registryID}/users",
		fmt.Sprintf("/v1/cloud/project/%s/containerRegistry/%s/users", projectID, url.PathEscape(args[0])),
		CloudContainerRegistryUserCreateSample,
		CloudContainerRegistryUserSpec,
		assets.CloudOpenapiSchema,
		nil,
	)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, user, "✅ Container registry user '%s' created successfully with password '%s'", user["user"], user["password"])
}

func SetContainerRegistryUserAsAdmin(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	userID, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/containerRegistry/%s/users/%d/setAsAdmin", projectID, url.PathEscape(args[0]), userID)
	if err := httpLib.Client.Put(endpoint, nil, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to set container registry user as admin: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Container registry user successfully set as admin")
}

func DeleteContainerRegistryUser(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	userID, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/containerRegistry/%s/users/%d", projectID, url.PathEscape(args[0]), userID)
	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to delete container registry user: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Container registry user deleted successfully")
}
