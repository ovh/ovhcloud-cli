// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package iam

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
	iamPolicyColumnsToDisplay           = []string{"id", "name", "owner", "readOnly"}
	iamPermissionsGroupColumnsToDisplay = []string{"id", "name", "description"}
	iamResourceColumnsToDisplay         = []string{"urn", "type", "displayName"}
	iamResourceGroupColumnsToDisplay    = []string{"id", "name", "owner", "readOnly"}

	//go:embed templates/iam_policy.tmpl
	iamPolicyTemplate string

	//go:embed templates/iam_permissions_group.tmpl
	iamPermissionsGroupTemplate string

	//go:embed templates/iam_resource.tmpl
	iamResourceTemplate string

	//go:embed templates/iam_resource_group.tmpl
	iamResourceGroupTemplate string

	//go:embed parameter-samples/user-create.json
	UserCreateExample string

	//go:embed parameter-samples/user-edit.json
	UserEditExample string

	//go:embed parameter-samples/token-create.json
	TokenCreateExample string

	IAMPolicySpec struct {
		Name        string   `json:"name,omitempty"`
		Description string   `json:"description,omitempty"`
		ExpiredAt   string   `json:"expiredAt,omitempty"`
		Identities  []string `json:"identities,omitempty"`
		Permissions struct {
			Allow  []iamPermission `json:"allow,omitempty"`
			Deny   []iamPermission `json:"deny,omitempty"`
			Except []iamPermission `json:"except,omitempty"`
		} `json:"permissions,omitzero"`
		PermissionsGroups []iamResourceURN `json:"permissionsGroups,omitempty"`
		Resources         []iamResourceURN `json:"resources,omitempty"`

		// Fields used for edition through the CLI
		PermissionsAllowed    []string `json:"-"`
		PermissionsDenied     []string `json:"-"`
		PermissionsExcept     []string `json:"-"`
		PermissionsGroupsURNs []string `json:"-"`
		ResourcesURNs         []string `json:"-"`
	}

	IAMResourceSpec struct {
		Tags map[string]string `json:"tags,omitempty"`
	}

	UserSpec struct {
		Description string `json:"description,omitempty"`
		Email       string `json:"email,omitempty"`
		Group       string `json:"group,omitempty"`
		Login       string `json:"login,omitempty"`
		Password    string `json:"password,omitempty"`
		Type        string `json:"type,omitempty"`
	}

	TokenSpec struct {
		Name        string `json:"name,omitempty"`
		Description string `json:"description,omitempty"`
		ExpiredAt   string `json:"expiredAt,omitempty"`
		ExpiresIn   int    `json:"expiresIn,omitempty"`
	}
)

type iamPermission struct {
	Action string `json:"action"`
}

type iamResourceURN struct {
	URN string `json:"urn"`
}

func ListIAMPolicies(_ *cobra.Command, _ []string) {
	common.ManageListRequestNoExpand("/v2/iam/policy", iamPolicyColumnsToDisplay, flags.GenericFilters)
}

func GetIAMPolicy(_ *cobra.Command, args []string) {
	path := fmt.Sprintf("/v2/iam/policy/%s?details=true", url.PathEscape(args[0]))

	var object map[string]any
	if err := httpLib.Client.Get(path, &object); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error fetching IAM policy %s: %s", args[0], err)
		return
	}

	display.OutputObject(object, args[0], iamPolicyTemplate, &flags.OutputFormatConfig)
}

func EditIAMPolicy(cmd *cobra.Command, args []string) {
	prepareIAMPermissionsFromCLI()
	if err := common.EditResource(
		cmd,
		"/iam/policy/{policyId}",
		fmt.Sprintf("/v2/iam/policy/%s", url.PathEscape(args[0])),
		IAMPolicySpec,
		assets.IamOpenapiSchema,
	); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}
}

func ListIAMPermissionsGroups(_ *cobra.Command, _ []string) {
	common.ManageListRequestNoExpand("/v2/iam/permissionsGroup", iamPermissionsGroupColumnsToDisplay, flags.GenericFilters)
}

func GetIAMPermissionsGroup(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/v2/iam/permissionsGroup", args[0], iamPermissionsGroupTemplate)
}

func EditIAMPermissionsGroup(cmd *cobra.Command, args []string) {
	prepareIAMPermissionsFromCLI()
	if err := common.EditResource(
		cmd,
		"/iam/permissionsGroup/{permissionsGroupURN}",
		fmt.Sprintf("/v2/iam/permissionsGroup/%s", url.PathEscape(args[0])),
		IAMPolicySpec,
		assets.IamOpenapiSchema,
	); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}
}

func ListIAMResources(_ *cobra.Command, _ []string) {
	common.ManageListRequestNoExpand("/v2/iam/resource", iamResourceColumnsToDisplay, flags.GenericFilters)
}

func GetIAMResource(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/v2/iam/resource", args[0], iamResourceTemplate)
}

func EditIAMResource(cmd *cobra.Command, args []string) {
	if err := common.EditResource(
		cmd,
		"/iam/resource/{resourceURN}",
		fmt.Sprintf("/v2/iam/resource/%s", url.PathEscape(args[0])),
		IAMResourceSpec,
		assets.IamOpenapiSchema,
	); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}
}

func ListIAMResourceGroups(_ *cobra.Command, _ []string) {
	common.ManageListRequestNoExpand("/v2/iam/resourceGroup", iamResourceGroupColumnsToDisplay, flags.GenericFilters)
}

func GetIAMResourceGroup(_ *cobra.Command, args []string) {
	path := fmt.Sprintf("/v2/iam/resourceGroup/%s?details=true", url.PathEscape(args[0]))

	var object map[string]any
	if err := httpLib.Client.Get(path, &object); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error fetching IAM resource group %s: %s", args[0], err)
		return
	}

	display.OutputObject(object, args[0], iamResourceGroupTemplate, &flags.OutputFormatConfig)
}

func EditIAMResourceGroup(cmd *cobra.Command, args []string) {
	prepareIAMPermissionsFromCLI()
	if err := common.EditResource(
		cmd,
		"/iam/resourceGroup/{groupId}",
		fmt.Sprintf("/v2/iam/resourceGroup/%s", url.PathEscape(args[0])),
		IAMPolicySpec,
		assets.IamOpenapiSchema,
	); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}
}

// prepareIAMPermissionsFromCLI transforms the CLI parameters into the IAMPolicySpec structure
func prepareIAMPermissionsFromCLI() {
	for _, action := range IAMPolicySpec.PermissionsAllowed {
		IAMPolicySpec.Permissions.Allow = append(IAMPolicySpec.Permissions.Allow, iamPermission{Action: action})
	}
	for _, action := range IAMPolicySpec.PermissionsDenied {
		IAMPolicySpec.Permissions.Deny = append(IAMPolicySpec.Permissions.Deny, iamPermission{Action: action})
	}
	for _, action := range IAMPolicySpec.PermissionsExcept {
		IAMPolicySpec.Permissions.Except = append(IAMPolicySpec.Permissions.Except, iamPermission{Action: action})
	}
	for _, urn := range IAMPolicySpec.PermissionsGroupsURNs {
		IAMPolicySpec.PermissionsGroups = append(IAMPolicySpec.PermissionsGroups, iamResourceURN{URN: urn})
	}
	for _, urn := range IAMPolicySpec.ResourcesURNs {
		IAMPolicySpec.Resources = append(IAMPolicySpec.Resources, iamResourceURN{URN: urn})
	}
}

func ListUsers(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/me/identity/user", "", []string{"login", "group", "description"}, flags.GenericFilters)
}

func GetUser(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/me/identity/user", args[0], "")
}

func CreateUser(cmd *cobra.Command, _ []string) {
	_, err := common.CreateResource(
		cmd,
		"/me/identity/user",
		"/me/identity/user",
		UserCreateExample,
		UserSpec,
		assets.MeOpenapiSchema,
		[]string{"login", "password", "email"})
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to create user: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ User %s created successfully", UserSpec.Login)
}

func EditUser(cmd *cobra.Command, args []string) {
	if err := common.EditResource(
		cmd,
		"/me/identity/user/{user}",
		fmt.Sprintf("/me/identity/user/%s", url.PathEscape(args[0])),
		UserSpec,
		assets.MeOpenapiSchema,
	); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}
}

func DeleteUser(_ *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("/me/identity/user/%s", url.PathEscape(args[0]))
	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to delete user %s: %s", args[0], err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ User %s deleted successfully", args[0])
}

func ListUserTokens(_ *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("/me/identity/user/%s/token", url.PathEscape(args[0]))
	common.ManageListRequest(endpoint, "", []string{"name", "description", "expiresAt"}, flags.GenericFilters)
}

func GetUserToken(_ *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("/me/identity/user/%s/token", url.PathEscape(args[0]))
	common.ManageObjectRequest(endpoint, args[1], "")
}

func CreateUserToken(cmd *cobra.Command, args []string) {
	token, err := common.CreateResource(
		cmd,
		"/me/identity/user/{user}/token",
		fmt.Sprintf("/me/identity/user/%s/token", url.PathEscape(args[0])),
		TokenCreateExample,
		TokenSpec,
		assets.MeOpenapiSchema,
		[]string{"name", "description"})
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to create token for user %s: %s", args[0], err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, token, "✅ Token %s created successfully, value: %s", token["name"], token["token"])
}

func DeleteUserToken(_ *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("/me/identity/user/%s/token/%s", url.PathEscape(args[0]), url.PathEscape(args[1]))
	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to delete token %s for user %s: %s", args[1], args[0], err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Token %s deleted successfully for user %s", args[1], args[0])
}
