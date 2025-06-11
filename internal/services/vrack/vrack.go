package vrack

import (
	_ "embed"
	"fmt"
	"net/url"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/display"
	"stash.ovh.net/api/ovh-cli/internal/editor"
	"stash.ovh.net/api/ovh-cli/internal/flags"
	httpLib "stash.ovh.net/api/ovh-cli/internal/http"
	"stash.ovh.net/api/ovh-cli/internal/services/common"
)

var (
	vrackColumnsToDisplay = []string{"serviceName", "name", "description"}

	//go:embed templates/vrack.tmpl
	vrackTemplate string

	//go:embed api-schemas/vrack.json
	vrackOpenapiSchema []byte
)

func ListVrack(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/vrack", "", vrackColumnsToDisplay, flags.GenericFilters)
}

func GetVrack(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/vrack", args[0], vrackTemplate)
}

func EditVrack(_ *cobra.Command, args []string) {
	url := fmt.Sprintf("/vrack/%s", url.PathEscape(args[0]))
	if err := editor.EditResource(httpLib.Client, "/vrack/{serviceName}", url, vrackOpenapiSchema); err != nil {
		display.ExitError(err.Error())
	}
}
