package dedicatednasha

import (
	_ "embed"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/flags"
	"stash.ovh.net/api/ovh-cli/internal/services/common"
)

var (
	dedicatednashaColumnsToDisplay = []string{"serviceName", "customName", "datacenter"}

	//go:embed templates/dedicatednasha.tmpl
	dedicatednashaTemplate string
)

func ListDedicatedNasHA(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/dedicated/nasha", "", dedicatednashaColumnsToDisplay, flags.GenericFilters)
}

func GetDedicatedNasHA(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/dedicated/nasha", args[0], dedicatednashaTemplate)
}
