package cdndedicated

import (
	_ "embed"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/flags"
	"stash.ovh.net/api/ovh-cli/internal/services/common"
)

var (
	cdndedicatedColumnsToDisplay = []string{"service", "offer", "anycast"}

	//go:embed templates/cdndedicated.tmpl
	cdndedicatedTemplate string
)

func ListCdnDedicated(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/cdn/dedicated", "", cdndedicatedColumnsToDisplay, flags.GenericFilters)
}

func GetCdnDedicated(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/cdn/dedicated", args[0], cdndedicatedTemplate)
}
