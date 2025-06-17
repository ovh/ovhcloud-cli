package emailmxplan

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
	emailmxplanColumnsToDisplay = []string{"domain", "displayName", "state", "offer"}

	//go:embed templates/emailmxplan.tmpl
	emailmxplanTemplate string

	//go:embed api-schemas/emailmxplan.json
	emailmxplanOpenapiSchema []byte
)

func ListEmailMXPlan(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/email/mxplan", "", emailmxplanColumnsToDisplay, flags.GenericFilters)
}

func GetEmailMXPlan(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/email/mxplan", args[0], emailmxplanTemplate)
}

func EditEmailMXPlan(_ *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("/email/mxplan/%s", url.PathEscape(args[0]))
	if err := editor.EditResource(httpLib.Client, "/email/mxplan/{service}", endpoint, emailmxplanOpenapiSchema); err != nil {
		display.ExitError(err.Error())
	}
}
