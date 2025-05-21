package webhosting

import (
	_ "embed"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/flags"
	"stash.ovh.net/api/ovh-cli/internal/services/common"
)

var (
	webhostingColumnsToDisplay = []string{"serviceName", "displayName", "datacenter", "state"}

	//go:embed templates/webhosting.tmpl
	webhostingTemplate string
)

func ListWebHosting(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/hosting/web", "", webhostingColumnsToDisplay, flags.GenericFilters)
}

func GetWebHosting(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/hosting/web", args[0], webhostingTemplate)
}
