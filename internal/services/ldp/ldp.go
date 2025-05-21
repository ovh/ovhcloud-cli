package ldp

import (
	_ "embed"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/flags"
	"stash.ovh.net/api/ovh-cli/internal/services/common"
)

var (
	ldpColumnsToDisplay = []string{"serviceName", "displayName", "isClusterOwner", "state", "username"}

	//go:embed templates/ldp.tmpl
	ldpTemplate string
)

func ListLdp(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/dbaas/logs", "", ldpColumnsToDisplay, flags.GenericFilters)
}

func GetLdp(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/dbaas/logs", args[0], ldpTemplate)
}
