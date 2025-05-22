package overthebox

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
	overtheboxColumnsToDisplay = []string{"serviceName", "offer", "status", "bandwidth"}

	//go:embed templates/overthebox.tmpl
	overtheboxTemplate string

	//go:embed api-schemas/overthebox.json
	overtheboxOpenapiSchema []byte
)

func ListOverTheBox(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/overTheBox", "", overtheboxColumnsToDisplay, flags.GenericFilters)
}

func GetOverTheBox(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/overTheBox", args[0], overtheboxTemplate)
}

func EditOverTheBox(_ *cobra.Command, args []string) {
	url := fmt.Sprintf("/overTheBox/%s", url.PathEscape(args[0]))
	editor.EditResource(httpLib.Client, "/overTheBox/{serviceName}", url, overtheboxOpenapiSchema)
}
