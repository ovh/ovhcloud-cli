package vmwareclouddirectorbackup

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
	vmwareclouddirectorbackupColumnsToDisplay = []string{"id", "iam.displayName", "currentState.azName", "resourceStatus"}

	//go:embed templates/vmwareclouddirectorbackup.tmpl
	vmwareclouddirectorbackupTemplate string

	//go:embed api-schemas/vmwareclouddirectorbackup.json
	vmwareclouddirectorbackupOpenapiSchema []byte
)

func ListVmwareCloudDirectorBackup(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/v2/vmwareCloudDirector/backup", "id", vmwareclouddirectorbackupColumnsToDisplay, flags.GenericFilters)
}

func GetVmwareCloudDirectorBackup(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/v2/vmwareCloudDirector/backup", args[0], vmwareclouddirectorbackupTemplate)
}

func EditVmwareCloudDirectorBackup(_ *cobra.Command, args []string) {
	url := fmt.Sprintf("/v2/vmwareCloudDirector/backup/%s", url.PathEscape(args[0]))
	editor.EditResource(httpLib.Client, "/vmwareCloudDirector/backup/{backupId}", url, vmwareclouddirectorbackupOpenapiSchema)
}
