package vmwareclouddirectororganization

import (
	_ "embed"
	"fmt"
	"net/url"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/editor"
	"stash.ovh.net/api/ovh-cli/internal/flags"
	httpLib "stash.ovh.net/api/ovh-cli/internal/http"
	"stash.ovh.net/api/ovh-cli/internal/services/common"
)

var (
	vmwareclouddirectororganizationColumnsToDisplay = []string{"id", "currentState.fullName", "currentState.region", "resourceStatus"}

	//go:embed templates/vmwareclouddirectororganization.tmpl
	vmwareclouddirectororganizationTemplate string

	//go:embed api-schemas/vmwareclouddirectororganization.json
	vmwareclouddirectororganizationOpenapiSchema []byte
)

func ListVmwareCloudDirectorOrganization(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/v2/vmwareCloudDirector/organization", "id", vmwareclouddirectororganizationColumnsToDisplay, flags.GenericFilters)
}

func GetVmwareCloudDirectorOrganization(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/v2/vmwareCloudDirector/organization", args[0], vmwareclouddirectororganizationTemplate)
}

func EditVmwareCloudDirectorOrganization(_ *cobra.Command, args []string) {
	url := fmt.Sprintf("/v2/vmwareCloudDirector/organization/%s", url.PathEscape(args[0]))
	editor.EditResource(httpLib.Client, "/vmwareCloudDirector/organization/{organizationId}", url, vmwareclouddirectororganizationOpenapiSchema)
}
