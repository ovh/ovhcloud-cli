package telephony

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
	telephonyColumnsToDisplay = []string{"billingAccount", "description", "status"}

	//go:embed templates/telephony.tmpl
	telephonyTemplate string

	//go:embed api-schemas/telephony.json
	telephonyOpenapiSchema []byte
)

func ListTelephony(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/telephony", "", telephonyColumnsToDisplay, flags.GenericFilters)
}

func GetTelephony(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/telephony", args[0], telephonyTemplate)
}

func EditTelephony(_ *cobra.Command, args []string) {
	url := fmt.Sprintf("/telephony/%s", url.PathEscape(args[0]))
	if err := editor.EditResource(httpLib.Client, "/telephony/{billingAccount}", url, telephonyOpenapiSchema); err != nil {
		display.ExitError(err.Error())
	}
}
