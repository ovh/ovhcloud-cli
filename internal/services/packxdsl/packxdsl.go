package packxdsl

import (
	_ "embed"
	"fmt"
	"net/url"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/display"
	"stash.ovh.net/api/ovh-cli/internal/flags"
	"stash.ovh.net/api/ovh-cli/internal/services/common"
)

var (
	packxdslColumnsToDisplay = []string{"packName", "description"}

	//go:embed templates/packxdsl.tmpl
	packxdslTemplate string

	//go:embed api-schemas/packxdsl.json
	packxdslOpenapiSchema []byte

	PackXDSLSpec struct {
		Description string `json:"description,omitempty"`
	}
)

func ListPackXDSL(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/pack/xdsl", "", packxdslColumnsToDisplay, flags.GenericFilters)
}

func GetPackXDSL(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/pack/xdsl", args[0], packxdslTemplate)
}

func EditPackXDSL(cmd *cobra.Command, args []string) {
	if err := common.EditResource(
		cmd,
		"/pack/xdsl/{packName}",
		fmt.Sprintf("/pack/xdsl/%s", url.PathEscape(args[0])),
		PackXDSLSpec,
		packxdslOpenapiSchema,
	); err != nil {
		display.ExitError(err.Error())
		return
	}
}
