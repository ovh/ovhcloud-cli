package vrack

import (
	_ "embed"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/flags"
	"stash.ovh.net/api/ovh-cli/internal/services/common"
)

var (
	vrackColumnsToDisplay = []string{"serviceName", "name", "description"}

	//go:embed templates/vrack.tmpl
	vrackTemplate string
)

func ListVrack(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/vrack", "", vrackColumnsToDisplay, flags.GenericFilters)
}

func GetVrack(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/vrack", args[0], vrackTemplate)
}
