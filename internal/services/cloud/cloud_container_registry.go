// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cloud

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"strconv"

	"code.cloudfoundry.org/bytefmt"
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

	cloudprojectContainerRegistryPlanCapabilitiesColumnsToDisplay = []string{"id", "name", "vulnerability", "imageStorage", "parallelRequest"}

	cloudProjectContainerRegistryIPRestrictionsColumnsToDisplay = []string{"ipBlock", "description", "createdAt", "updatedAt"}

	//go:embed templates/cloud_container_registry.tmpl
	cloudContainerRegistryTemplate string

	//go:embed templates/cloud_container_registry_user.tmpl
	cloudContainerRegistryUserTemplate string

	//go:embed parameter-samples/container-registry-create.json
	CloudContainerRegistryCreateSample string

	//go:embed parameter-samples/container-registry-user-create.json
	CloudContainerRegistryUserCreateSample string

	//go:embed parameter-samples/container-registry-iam-enable.json
	CloudContainerRegistryIamEnableSample string

	//go:embed parameter-samples/container-registry-oidc-create.json
	CloudContainerRegistryOidcCreateSample string

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

	CloudContainerRegistryIamSpec struct {
		DeleteUsers bool `json:"deleteUsers"`
	}

	CloudContainerRegistryOidcCreateSpec struct {
		DeleteUsers bool `json:"deleteUsers,omitempty"`
		Provider    struct {
			AdminGroup   string `json:"adminGroup,omitempty"`
			AutoOnboard  bool   `json:"autoOnboard,omitempty"`
			ClientID     string `json:"clientId"`
			ClientSecret string `json:"clientSecret"`
			Endpoint     string `json:"endpoint"`
			GroupFilter  string `json:"groupFilter,omitempty"`
			GroupsClaim  string `json:"groupsClaim,omitempty"`
			Name         string `json:"name"`
			Scope        string `json:"scope"`
			UserClaim    string `json:"userClaim,omitempty"`
			VerifyCert   bool   `json:"verifyCert,omitempty"`
		} `json:"provider"`
	}

	CloudContainerRegistryOidcEditSpec struct {
		AdminGroup   string `json:"adminGroup,omitempty"`
		AutoOnboard  bool   `json:"autoOnboard,omitempty"`
		ClientID     string `json:"clientId,omitempty"`
		ClientSecret string `json:"clientSecret,omitempty"`
		Endpoint     string `json:"endpoint,omitempty"`
		GroupFilter  string `json:"groupFilter,omitempty"`
		GroupsClaim  string `json:"groupsClaim,omitempty"`
		Name         string `json:"name,omitempty"`
		Scope        string `json:"scope,omitempty"`
		UserClaim    string `json:"userClaim,omitempty"`
		VerifyCert   bool   `json:"verifyCert,omitempty"`
	}

	CloudContainerRegistryPlanUpgradeSpec struct {
		PlanID string `json:"planID"`
	}

	ContainerRegistryIPRestrictionsAddSpec struct {
		IPBlock     string
		Description string
	}
	ContainerRegistryIPRestrictionsDeleteSpec struct {
		IPBlock string
	}
)

type (
	ContainerRegistryIPRestriction struct {
		CreatedAt   string `json:"createdAt,omitempty"`
		Description string `json:"description,omitempty"`
		IPBlock     string `json:"ipBlock"`
		UpdatedAt   string `json:"updatedAt,omitempty"`
	}

	ContainerRegistryIPRestrictionInput struct {
		Description string `json:"description,omitempty"`
		IPBlock     string `json:"ipBlock"`
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

func EnableContainerRegistryIAM(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	_, err = common.CreateResource(
		cmd,
		"/cloud/project/{serviceName}/containerRegistry/{registryID}/iam",
		fmt.Sprintf("/v1/cloud/project/%s/containerRegistry/%s/iam", projectID, url.PathEscape(args[0])),
		CloudContainerRegistryIamEnableSample,
		CloudContainerRegistryIamSpec,
		assets.CloudOpenapiSchema,
		nil,
	)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Container registry IAM enabled successfully")
}

func DisableContainerRegistryIAM(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/containerRegistry/%s/iam", projectID, url.PathEscape(args[0]))
	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to disable container registry IAM: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Container registry IAM disabled successfully")
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

func GetContainerRegistryOIDC(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/containerRegistry/%s/openIdConnect", projectID, url.PathEscape(args[0]))
	var configuration map[string]any
	if err := httpLib.Client.Get(endpoint, &configuration); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error fetching %s: %s", endpoint, err)
		return
	}

	display.OutputObject(configuration, args[0], "", &flags.OutputFormatConfig)
}

func CreateContainerRegistryOIDC(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	configuration, err := common.CreateResource(
		cmd,
		"/cloud/project/{serviceName}/containerRegistry/{registryID}/openIdConnect",
		fmt.Sprintf("/v1/cloud/project/%s/containerRegistry/%s/openIdConnect", projectID, url.PathEscape(args[0])),
		CloudContainerRegistryOidcCreateSample,
		CloudContainerRegistryOidcCreateSpec,
		assets.CloudOpenapiSchema,
		[]string{"provider"}, // TODO: Add providers sub variables (name, endpoint...) when CreateResource function supports embedded map
	)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, configuration, "✅ Container registry OIDC configuration created successfully")
}

func EditContainerRegistryOIDC(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	payload := make(map[string]any)
	maybeAddString := func(flagName, key, value string) {
		if cmd.Flags().Changed(flagName) {
			payload[key] = value
		}
	}
	maybeAddBool := func(flagName, key string, value bool) {
		if cmd.Flags().Changed(flagName) {
			payload[key] = value
		}
	}

	maybeAddString("admin-group", "adminGroup", CloudContainerRegistryOidcEditSpec.AdminGroup)
	maybeAddString("client-id", "clientId", CloudContainerRegistryOidcEditSpec.ClientID)
	maybeAddString("client-secret", "clientSecret", CloudContainerRegistryOidcEditSpec.ClientSecret)
	maybeAddString("endpoint", "endpoint", CloudContainerRegistryOidcEditSpec.Endpoint)
	maybeAddString("group-filter", "groupFilter", CloudContainerRegistryOidcEditSpec.GroupFilter)
	maybeAddString("groups-claim", "groupsClaim", CloudContainerRegistryOidcEditSpec.GroupsClaim)
	maybeAddString("name", "name", CloudContainerRegistryOidcEditSpec.Name)
	maybeAddString("scope", "scope", CloudContainerRegistryOidcEditSpec.Scope)
	maybeAddString("user-claim", "userClaim", CloudContainerRegistryOidcEditSpec.UserClaim)
	maybeAddBool("auto-onboard", "autoOnboard", CloudContainerRegistryOidcEditSpec.AutoOnboard)
	maybeAddBool("verify-cert", "verifyCert", CloudContainerRegistryOidcEditSpec.VerifyCert)

	if err := common.EditResource(
		cmd,
		"/cloud/project/{serviceName}/containerRegistry/{registryID}/openIdConnect",
		fmt.Sprintf("/v1/cloud/project/%s/containerRegistry/%s/openIdConnect", projectID, url.PathEscape(args[0])),
		payload,
		assets.CloudOpenapiSchema,
	); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}
}

func DeleteContainerRegistryOIDC(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/containerRegistry/%s/openIdConnect", projectID, url.PathEscape(args[0]))
	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to delete container registry OIDC configuration: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Container registry OIDC configuration deleted successfully")
}

func ListContainerRegistryPlanCapabilities(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/containerRegistry/%s/capabilities/plan", projectID, url.PathEscape(args[0]))

	var plans []map[string]any
	if err := httpLib.Client.Get(endpoint, &plans); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch container registry plan capabilities: %s", err)
		return
	}

	for _, plan := range plans {
		formatContainerRegistryPlans(plan)
	}

	plans, err = filtersLib.FilterLines(plans, flags.GenericFilters)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to filter results: %s", err)
		return
	}

	display.RenderTable(plans, cloudprojectContainerRegistryPlanCapabilitiesColumnsToDisplay, &flags.OutputFormatConfig)
}

func UpgradeContainerRegistryPlan(_ *cobra.Command, args []string) {
	if CloudContainerRegistryPlanUpgradeSpec.PlanID == "" {
		display.OutputError(&flags.OutputFormatConfig, "plan-id flag is required")
		return
	}

	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/containerRegistry/%s/plan", projectID, url.PathEscape(args[0]))

	if err := httpLib.Client.Put(endpoint, CloudContainerRegistryPlanUpgradeSpec, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to upgrade container registry plan: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Container registry %s plan upgraded to %s", args[0], CloudContainerRegistryPlanUpgradeSpec.PlanID)
}

func formatContainerRegistryPlans(plan map[string]any) {
	// Extract and format registry limits
	if registryLimits, ok := plan["registryLimits"].(map[string]any); ok {
		if imageStorage, ok := registryLimits["imageStorage"].(json.Number); ok {
			imageStorage, err := imageStorage.Int64()
			if err != nil {
				display.OutputError(&flags.OutputFormatConfig, "%s", err)
			}

			plan["imageStorage"] = bytefmt.ByteSize(uint64(imageStorage))
		}
		if parallelRequest, ok := registryLimits["parallelRequest"].(json.Number); ok {
			plan["parallelRequest"] = parallelRequest
		}
	}

	// Extract vulnerability feature
	if features, ok := plan["features"].(map[string]any); ok {
		if vulnerability, ok := features["vulnerability"].(bool); ok {
			plan["vulnerability"] = vulnerability
		}
	}
}

// listContainerRegistryIPRestrictions lists IP restrictions for container registry (management or registry)
func listContainerRegistryIPRestrictions(_ *cobra.Command, args []string, restrictionType string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/containerRegistry/%s/ipRestrictions/%s", projectID, url.PathEscape(args[0]), restrictionType)

	var restrictions []ContainerRegistryIPRestriction
	if err := httpLib.Client.Get(endpoint, &restrictions); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch IP restrictions: %s", err)
		return
	}

	objects := make([]map[string]any, 0, len(restrictions))
	for _, restriction := range restrictions {
		objects = append(objects, map[string]any{
			"ipBlock":     restriction.IPBlock,
			"description": restriction.Description,
			"createdAt":   restriction.CreatedAt,
			"updatedAt":   restriction.UpdatedAt,
		})
	}

	ipRestrictions, err := filtersLib.FilterLines(objects, flags.GenericFilters)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to filter results: %s", err)
		return
	}

	display.RenderTable(ipRestrictions, cloudProjectContainerRegistryIPRestrictionsColumnsToDisplay, &flags.OutputFormatConfig)
}

// ListContainerRegistryIPRestrictionsManagement lists management IP restrictions for container registry
func ListContainerRegistryIPRestrictionsManagement(cmd *cobra.Command, args []string) {
	listContainerRegistryIPRestrictions(cmd, args, "management")
}

// ListContainerRegistryIPRestrictionsRegistry lists registry IP restrictions for container registry
func ListContainerRegistryIPRestrictionsRegistry(cmd *cobra.Command, args []string) {
	listContainerRegistryIPRestrictions(cmd, args, "registry")
}

// addContainerRegistryIPRestriction adds an IP restriction to container registry (management or registry)
func addContainerRegistryIPRestriction(_ *cobra.Command, args []string, restrictionType string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	// Validate IP block
	if ContainerRegistryIPRestrictionsAddSpec.IPBlock == "" {
		display.OutputError(&flags.OutputFormatConfig, "ip-block flag is required")
		return
	}

	if _, _, err := net.ParseCIDR(ContainerRegistryIPRestrictionsAddSpec.IPBlock); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "invalid CIDR notation for ip-block: %s", err)
		return
	}

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/containerRegistry/%s/ipRestrictions/%s", projectID, url.PathEscape(args[0]), restrictionType)

	// Fetch existing restrictions
	var restrictions []ContainerRegistryIPRestriction
	if err := httpLib.Client.Get(endpoint, &restrictions); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch existing IP restrictions: %s", err)
		return
	}

	// Check for duplicate IP block
	for _, restriction := range restrictions {
		if restriction.IPBlock == ContainerRegistryIPRestrictionsAddSpec.IPBlock {
			display.OutputError(&flags.OutputFormatConfig, "IP block %s already exists in IP restrictions", ContainerRegistryIPRestrictionsAddSpec.IPBlock)
			return
		}
	}

	// Append new restriction
	newRestriction := ContainerRegistryIPRestrictionInput{
		IPBlock:     ContainerRegistryIPRestrictionsAddSpec.IPBlock,
		Description: ContainerRegistryIPRestrictionsAddSpec.Description,
	}

	var inputRestrictions []ContainerRegistryIPRestrictionInput
	for _, restriction := range restrictions {
		inputRestrictions = append(inputRestrictions, ContainerRegistryIPRestrictionInput{
			IPBlock:     restriction.IPBlock,
			Description: restriction.Description,
		})
	}
	inputRestrictions = append(inputRestrictions, newRestriction)

	// PUT updated restrictions
	if err := httpLib.Client.Put(endpoint, inputRestrictions, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to update IP restrictions: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ IP restriction %s added to %s", ContainerRegistryIPRestrictionsAddSpec.IPBlock, restrictionType)
}

// AddContainerRegistryIPRestrictionsManagement adds a management IP restriction to container registry
func AddContainerRegistryIPRestrictionsManagement(cmd *cobra.Command, args []string) {
	addContainerRegistryIPRestriction(cmd, args, "management")
}

// AddContainerRegistryIPRestrictionsRegistry adds a registry IP restriction to container registry
func AddContainerRegistryIPRestrictionsRegistry(cmd *cobra.Command, args []string) {
	addContainerRegistryIPRestriction(cmd, args, "registry")
}

// deleteContainerRegistryIPRestriction deletes an IP restriction from container registry (management or registry)
func deleteContainerRegistryIPRestriction(_ *cobra.Command, args []string, restrictionType string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	// Validate IP block
	if ContainerRegistryIPRestrictionsDeleteSpec.IPBlock == "" {
		display.OutputError(&flags.OutputFormatConfig, "ip-block flag is required")
		return
	}

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/containerRegistry/%s/ipRestrictions/%s", projectID, url.PathEscape(args[0]), restrictionType)

	// Fetch existing restrictions
	var restrictions []ContainerRegistryIPRestriction
	if err := httpLib.Client.Get(endpoint, &restrictions); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch existing IP restrictions: %s", err)
		return
	}

	// Find and remove the matching IP block
	inputRestrictions := make([]ContainerRegistryIPRestrictionInput, 0)
	found := false
	for _, restriction := range restrictions {
		if restriction.IPBlock == ContainerRegistryIPRestrictionsDeleteSpec.IPBlock {
			found = true
			continue
		}
		inputRestrictions = append(inputRestrictions, ContainerRegistryIPRestrictionInput{
			IPBlock:     restriction.IPBlock,
			Description: restriction.Description,
		})
	}

	if !found {
		display.OutputError(&flags.OutputFormatConfig, "IP block %s not found in %s IP restrictions", ContainerRegistryIPRestrictionsDeleteSpec.IPBlock, restrictionType)
		return
	}

	// PUT updated restrictions
	if err := httpLib.Client.Put(endpoint, inputRestrictions, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to update IP restrictions: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ IP restriction %s deleted from %s", ContainerRegistryIPRestrictionsDeleteSpec.IPBlock, restrictionType)
}

// DeleteContainerRegistryIPRestrictionsManagement deletes a management IP restriction from container registry
func DeleteContainerRegistryIPRestrictionsManagement(cmd *cobra.Command, args []string) {
	deleteContainerRegistryIPRestriction(cmd, args, "management")
}

// DeleteContainerRegistryIPRestrictionsRegistry deletes a registry IP restriction from container registry
func DeleteContainerRegistryIPRestrictionsRegistry(cmd *cobra.Command, args []string) {
	deleteContainerRegistryIPRestriction(cmd, args, "registry")
}
