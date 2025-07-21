package vrackservices

import (
	_ "embed"
	"fmt"
	"net/url"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/assets"
	"stash.ovh.net/api/ovh-cli/internal/display"
	"stash.ovh.net/api/ovh-cli/internal/editor"
	"stash.ovh.net/api/ovh-cli/internal/flags"
	httpLib "stash.ovh.net/api/ovh-cli/internal/http"
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

func EditVrackServices(_ *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("/v2/vrackServices/resource/%s", url.PathEscape(args[0]))
	if err := editor.EditResource(
		httpLib.Client,
		"/vrackServices/resource/{vrackServicesId}",
		endpoint,
		assets.VrackservicesOpenapiSchema,
	); err != nil {
		display.ExitError(err.Error())
	}
}
