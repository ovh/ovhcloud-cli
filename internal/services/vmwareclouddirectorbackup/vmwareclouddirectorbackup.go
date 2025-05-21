package vmwareclouddirectorbackup

import (
	_ "embed"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/flags"
	"stash.ovh.net/api/ovh-cli/internal/services/common"
)

var (
	vmwareclouddirectorbackupColumnsToDisplay = []string{"id", "iam.displayName", "currentState.azName", "resourceStatus"}

	//go:embed templates/vmwareclouddirectorbackup.tmpl
	vmwareclouddirectorbackupTemplate string
)

func ListVmwareCloudDirectorBackup(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/v2/vmwareCloudDirector/backup", "id", vmwareclouddirectorbackupColumnsToDisplay, flags.GenericFilters)
}

func GetVmwareCloudDirectorBackup(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/v2/vmwareCloudDirector/backup", args[0], vmwareclouddirectorbackupTemplate)
}
