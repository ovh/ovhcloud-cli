package location

import (
	_ "embed"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/flags"
	"stash.ovh.net/api/ovh-cli/internal/services/common"
)

var (
	locationColumnsToDisplay = []string{"name", "type", "specificType", "location"}

	//go:embed templates/location.tmpl
	locationTemplate string
)

func ListLocation(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/v2/location", "name", locationColumnsToDisplay, flags.GenericFilters)
}

func GetLocation(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/v2/location", args[0], locationTemplate)
}
