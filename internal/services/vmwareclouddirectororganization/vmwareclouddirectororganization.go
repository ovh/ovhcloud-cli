package vmwareclouddirectororganization

import (
	_ "embed"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/flags"
	"stash.ovh.net/api/ovh-cli/internal/services/common"
)

var (
	vmwareclouddirectororganizationColumnsToDisplay = []string{"id", "currentState.fullName", "currentState.region", "resourceStatus"}

	//go:embed templates/vmwareclouddirectororganization.tmpl
	vmwareclouddirectororganizationTemplate string
)

func ListVmwareCloudDirectorOrganization(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/v2/vmwareCloudDirector/organization", "id", vmwareclouddirectororganizationColumnsToDisplay, flags.GenericFilters)
}

func GetVmwareCloudDirectorOrganization(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/v2/vmwareCloudDirector/organization", args[0], vmwareclouddirectororganizationTemplate)
}
