package vrackservices

import (
	_ "embed"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/flags"
	"stash.ovh.net/api/ovh-cli/internal/services/common"
)

var (
	vrackservicesColumnsToDisplay = []string{"id", "currentState.region", "currentState. productStatus", "resourceStatus"}

	//go:embed templates/vrackservices.tmpl
	vrackservicesTemplate string
)

func ListVrackServices(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/v2/vrackServices/resource", "id", vrackservicesColumnsToDisplay, flags.GenericFilters)
}

func GetVrackServices(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/v2/vrackServices/resource", args[0], vrackservicesTemplate)
}
