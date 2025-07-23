package alldom

import (
	_ "embed"

	"github.com/ovh/ovhcloud-cli/internal/flags"
	"github.com/ovh/ovhcloud-cli/internal/services/common"
	"github.com/spf13/cobra"
)

var (
	alldomColumnsToDisplay = []string{"name", "type", "offer"}

	//go:embed templates/alldom.tmpl
	alldomTemplate string
)

func ListAllDom(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/allDom", "", alldomColumnsToDisplay, flags.GenericFilters)
}

func GetAllDom(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/allDom", args[0], alldomTemplate)
}
