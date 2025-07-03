package iam

import (
	_ "embed"
	"fmt"
	"net/url"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/display"
	"stash.ovh.net/api/ovh-cli/internal/flags"
	httpLib "stash.ovh.net/api/ovh-cli/internal/http"
	"stash.ovh.net/api/ovh-cli/internal/services/common"
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

	//go:embed api-schemas/iam.json
	iamOpenapiSchema []byte

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
		display.ExitError("error fetching IAM policy %s: %s", args[0], err)
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
		iamOpenapiSchema,
	); err != nil {
		display.ExitError(err.Error())
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
		iamOpenapiSchema,
	); err != nil {
		display.ExitError(err.Error())
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
		iamOpenapiSchema,
	); err != nil {
		display.ExitError(err.Error())
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
		display.ExitError("error fetching IAM resource group %s: %s", args[0], err)
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
		iamOpenapiSchema,
	); err != nil {
		display.ExitError(err.Error())
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
