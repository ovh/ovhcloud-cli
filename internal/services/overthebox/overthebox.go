package overthebox

import (
	_ "embed"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/flags"
	"stash.ovh.net/api/ovh-cli/internal/services/common"
)

var (
	overtheboxColumnsToDisplay = []string{"serviceName", "offer", "status", "bandwidth"}

	//go:embed templates/overthebox.tmpl
	overtheboxTemplate string
)

func ListOverTheBox(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/overTheBox", "", overtheboxColumnsToDisplay, flags.GenericFilters)
}

func GetOverTheBox(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/overTheBox", args[0], overtheboxTemplate)
}
