package packxdsl

import (
	_ "embed"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/flags"
	"stash.ovh.net/api/ovh-cli/internal/services/common"
)

var (
	packxdslColumnsToDisplay = []string{"packName", "description"}

	//go:embed templates/packxdsl.tmpl
	packxdslTemplate string
)

func ListPackXDSL(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/pack/xdsl", "", packxdslColumnsToDisplay, flags.GenericFilters)
}

func GetPackXDSL(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/pack/xdsl", args[0], packxdslTemplate)
}
