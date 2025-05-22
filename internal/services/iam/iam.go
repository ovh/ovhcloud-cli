package iam

import (
	_ "embed"
	"fmt"
	"net/url"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/display"
	"stash.ovh.net/api/ovh-cli/internal/editor"
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
)

func ListIAMPolicies(_ *cobra.Command, _ []string) {
	common.ManageListRequestNoExpand("/v2/iam/policy", iamPolicyColumnsToDisplay, flags.GenericFilters)
}

func GetIAMPolicy(_ *cobra.Command, args []string) {
	path := fmt.Sprintf("/v2/iam/policy/%s?details=true", url.PathEscape(args[0]))

	var object map[string]any
	if err := httpLib.Client.Get(path, &object); err != nil {
		display.ExitError("error fetching IAM policy %s: %s", args[0], err)
	}

	display.OutputObject(object, args[0], iamPolicyTemplate, &flags.OutputFormatConfig)
}

func EditIAMPolicy(_ *cobra.Command, args []string) {
	url := fmt.Sprintf("/v2/iam/policy/%s", url.PathEscape(args[0]))
	editor.EditResource(httpLib.Client, "/iam/policy/{policyId}", url, iamOpenapiSchema)
}

func ListIAMPermissionsGroups(_ *cobra.Command, _ []string) {
	common.ManageListRequestNoExpand("/v2/iam/permissionsGroup", iamPermissionsGroupColumnsToDisplay, flags.GenericFilters)
}

func GetIAMPermissionsGroup(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/v2/iam/permissionsGroup", args[0], iamPermissionsGroupTemplate)
}

func EditIAMPermissionsGroup(_ *cobra.Command, args []string) {
	url := fmt.Sprintf("/v2/iam/permissionsGroup/%s", url.PathEscape(args[0]))
	editor.EditResource(httpLib.Client, "/iam/permissionsGroup/{permissionsGroupURN}", url, iamOpenapiSchema)
}

func ListIAMResources(_ *cobra.Command, _ []string) {
	common.ManageListRequestNoExpand("/v2/iam/resource", iamResourceColumnsToDisplay, flags.GenericFilters)
}

func GetIAMResource(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/v2/iam/resource", args[0], iamResourceTemplate)
}

func EditIAMResource(_ *cobra.Command, args []string) {
	url := fmt.Sprintf("/v2/iam/resource/%s", url.PathEscape(args[0]))
	editor.EditResource(httpLib.Client, "/iam/resource/{resourceURN}", url, iamOpenapiSchema)
}

func ListIAMResourceGroups(_ *cobra.Command, _ []string) {
	common.ManageListRequestNoExpand("/v2/iam/resourceGroup", iamResourceGroupColumnsToDisplay, flags.GenericFilters)
}

func GetIAMResourceGroup(_ *cobra.Command, args []string) {
	path := fmt.Sprintf("/v2/iam/resourceGroup/%s?details=true", url.PathEscape(args[0]))

	var object map[string]any
	if err := httpLib.Client.Get(path, &object); err != nil {
		display.ExitError("error fetching IAM resource group %s: %s", args[0], err)
	}

	display.OutputObject(object, args[0], iamResourceGroupTemplate, &flags.OutputFormatConfig)
}

func EditIAMResourceGroup(_ *cobra.Command, args []string) {
	url := fmt.Sprintf("/v2/iam/resourceGroup/%s", url.PathEscape(args[0]))
	editor.EditResource(httpLib.Client, "/iam/resourceGroup/{groupId}", url, iamOpenapiSchema)
}
