package cmd

import (
	_ "embed"
	"fmt"
	"net/url"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/display"
	"stash.ovh.net/api/ovh-cli/internal/editor"
)

var (
	iamPolicyColumnsToDisplay           = []string{"id", "name", "owner", "readOnly"}
	iamPermissionsGroupColumnsToDisplay = []string{"id", "name", "description"}
	iamResourceColumnsToDisplay         = []string{"urn", "type", "displayName"}

	//go:embed templates/iam_policy.tmpl
	iamPolicyTemplate string

	//go:embed templates/iam_permissions_group.tmpl
	iamPermissionsGroupTemplate string

	//go:embed templates/iam_resource.tmpl
	iamResourceTemplate string

	//go:embed api-schemas/iam.json
	iamOpenapiSchema []byte
)

func listIAMPolicies(_ *cobra.Command, _ []string) {
	manageListRequestNoExpand("/v2/iam/policy", iamPolicyColumnsToDisplay, genericFilters)
}

func getIAMPolicy(_ *cobra.Command, args []string) {
	path := fmt.Sprintf("/v2/iam/policy/%s?details=true", url.PathEscape(args[0]))

	var object map[string]any
	if err := client.Get(path, &object); err != nil {
		display.ExitError("error fetching IAM policy %s: %s", args[0], err)
	}

	display.OutputObject(object, args[0], iamPolicyTemplate, &outputFormatConfig)
}

func editIAMPolicy(_ *cobra.Command, args []string) {
	url := fmt.Sprintf("/v2/iam/policy/%s", url.PathEscape(args[0]))
	editor.EditResource(client, "/iam/policy/{policyId}", url, iamOpenapiSchema)
}

func listIAMPermissionsGroups(_ *cobra.Command, _ []string) {
	manageListRequestNoExpand("/v2/iam/permissionsGroup", iamPermissionsGroupColumnsToDisplay, genericFilters)
}

func getIAMPermissionsGroup(_ *cobra.Command, args []string) {
	manageObjectRequest("/v2/iam/permissionsGroup", args[0], iamPermissionsGroupTemplate)
}

func listIAMResources(_ *cobra.Command, _ []string) {
	manageListRequestNoExpand("/v2/iam/resource", iamResourceColumnsToDisplay, genericFilters)
}

func getIAMResource(_ *cobra.Command, args []string) {
	manageObjectRequest("/v2/iam/resource", args[0], iamResourceTemplate)
}

func init() {
	iamCmd := &cobra.Command{
		Use:   "iam",
		Short: "Manage IAM resources, permissions and policies",
	}

	iamPolicyCmd := &cobra.Command{
		Use:   "policy",
		Short: "Manage IAM policies",
	}
	iamCmd.AddCommand(iamPolicyCmd)

	iamPolicyListCmd := withFilterFlag(&cobra.Command{
		Use:   "list",
		Short: "List IAM policies",
		Run:   listIAMPolicies,
	})
	iamPolicyCmd.AddCommand(iamPolicyListCmd)

	iamPolicyCmd.AddCommand(&cobra.Command{
		Use:   "get <policy_id>",
		Short: "Get a specific IAM policy",
		Run:   getIAMPolicy,
	})

	iamPolicyCmd.AddCommand(&cobra.Command{
		Use:   "edit <policy_id>",
		Short: "Edit specific IAM policy",
		Run:   editIAMPolicy,
	})

	iamPermissionsGroupCmd := &cobra.Command{
		Use:   "permissions-group",
		Short: "Manage IAM permissions groups",
	}
	iamCmd.AddCommand(iamPermissionsGroupCmd)

	iamPermissionsGroupCmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:   "list",
		Short: "List IAM permissions groups",
		Run:   listIAMPermissionsGroups,
	}))

	iamPermissionsGroupCmd.AddCommand(&cobra.Command{
		Use:   "get <permissions_group_id>",
		Short: "Get a specific IAM permissions group",
		Run:   getIAMPermissionsGroup,
	})

	iamResourceCmd := &cobra.Command{
		Use:   "resource",
		Short: "Manage IAM resources",
	}
	iamCmd.AddCommand(iamResourceCmd)

	iamResourceCmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:   "list",
		Short: "List IAM resources",
		Run:   listIAMResources,
	}))

	iamResourceCmd.AddCommand(&cobra.Command{
		Use:   "get <resource_urn>",
		Short: "Get a specific IAM resource",
		Run:   getIAMResource,
	})

	rootCmd.AddCommand(iamCmd)
}
