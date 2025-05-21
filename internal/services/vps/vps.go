package vps

import (
	_ "embed"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/flags"
	"stash.ovh.net/api/ovh-cli/internal/services/common"
)

var (
	vpsColumnsToDisplay = []string{"name", "displayName", "state", "zone"}

	//go:embed templates/vps.tmpl
	vpsTemplate string
)

func ListVps(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/vps", "", vpsColumnsToDisplay, flags.GenericFilters)
}

func GetVps(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/vps", args[0], vpsTemplate)
}
