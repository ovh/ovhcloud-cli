package xdsl

import (
	_ "embed"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/flags"
	"stash.ovh.net/api/ovh-cli/internal/services/common"
)

var (
	xdslColumnsToDisplay = []string{"accessName", "accessType", "provider", "role", "status"}

	//go:embed templates/xdsl.tmpl
	xdslTemplate string
)

func ListXdsl(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/xdsl", "", xdslColumnsToDisplay, flags.GenericFilters)
}

func GetXdsl(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/xdsl", args[0], xdslTemplate)
}
