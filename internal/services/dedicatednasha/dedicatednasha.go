package dedicatednasha

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
	dedicatednashaColumnsToDisplay = []string{"serviceName", "customName", "datacenter"}

	//go:embed templates/dedicatednasha.tmpl
	dedicatednashaTemplate string

	//go:embed api-schemas/dedicatednasha.json
	dedicatednashaOpenapiSchema []byte
)

func ListDedicatedNasHA(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/dedicated/nasha", "", dedicatednashaColumnsToDisplay, flags.GenericFilters)
}

func GetDedicatedNasHA(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/dedicated/nasha", args[0], dedicatednashaTemplate)
}

func EditDedicatedNasHA(_ *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("/dedicated/nasha/%s", url.PathEscape(args[0]))
	if err := editor.EditResource(httpLib.Client, "/dedicated/nasha/{serviceName}", endpoint, dedicatednashaOpenapiSchema); err != nil {
		display.ExitError(err.Error())
	}
}
