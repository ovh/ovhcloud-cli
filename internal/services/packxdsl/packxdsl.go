package packxdsl

import (
	_ "embed"
	"fmt"
	"net/url"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/editor"
	"stash.ovh.net/api/ovh-cli/internal/flags"
	httpLib "stash.ovh.net/api/ovh-cli/internal/http"
	"stash.ovh.net/api/ovh-cli/internal/services/common"
)

var (
	packxdslColumnsToDisplay = []string{"packName", "description"}

	//go:embed templates/packxdsl.tmpl
	packxdslTemplate string

	//go:embed api-schemas/packxdsl.json
	packxdslOpenapiSchema []byte
)

func ListPackXDSL(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/pack/xdsl", "", packxdslColumnsToDisplay, flags.GenericFilters)
}

func GetPackXDSL(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/pack/xdsl", args[0], packxdslTemplate)
}

func EditPackXDSL(_ *cobra.Command, args []string) {
	url := fmt.Sprintf("/pack/xdsl/%s", url.PathEscape(args[0]))
	editor.EditResource(httpLib.Client, "/pack/xdsl/{packName}", url, packxdslOpenapiSchema)
}
