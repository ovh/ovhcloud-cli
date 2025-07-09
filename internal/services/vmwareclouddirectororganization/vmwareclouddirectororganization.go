package vmwareclouddirectororganization

import (
	_ "embed"
	"fmt"
	"net/url"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/display"
	"stash.ovh.net/api/ovh-cli/internal/flags"
	"stash.ovh.net/api/ovh-cli/internal/services/common"
)

var (
	vmwareclouddirectororganizationColumnsToDisplay = []string{"id", "currentState.fullName", "currentState.region", "resourceStatus"}

	//go:embed templates/vmwareclouddirectororganization.tmpl
	vmwareclouddirectororganizationTemplate string

	//go:embed api-schemas/vmwareclouddirectororganization.json
	vmwareclouddirectororganizationOpenapiSchema []byte

	VmwareCloudDirectorOrganizationSpec struct {
		TargetSpec struct {
			Description string `json:"description,omitempty"`
			FullName    string `json:"fullName,omitempty"`
		} `json:"targetSpec,omitzero"`
	}
)

func ListVmwareCloudDirectorOrganization(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/v2/vmwareCloudDirector/organization", "id", vmwareclouddirectororganizationColumnsToDisplay, flags.GenericFilters)
}

func GetVmwareCloudDirectorOrganization(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/v2/vmwareCloudDirector/organization", args[0], vmwareclouddirectororganizationTemplate)
}

func EditVmwareCloudDirectorOrganization(cmd *cobra.Command, args []string) {
	if err := common.EditResource(
		cmd,
		"/vmwareCloudDirector/organization/{organizationId}",
		fmt.Sprintf("/v2/vmwareCloudDirector/organization/%s", url.PathEscape(args[0])),
		VmwareCloudDirectorOrganizationSpec,
		vmwareclouddirectororganizationOpenapiSchema,
	); err != nil {
		display.ExitError(err.Error())
		return
	}
}
