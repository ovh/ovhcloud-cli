package hostingprivatedatabase

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
	hostingprivatedatabaseColumnsToDisplay = []string{"serviceName", "displayName", "type", "version", "state"}

	//go:embed templates/hostingprivatedatabase.tmpl
	hostingprivatedatabaseTemplate string

	//go:embed api-schemas/hostingprivatedatabase.json
	hostingprivatedatabaseOpenapiSchema []byte
)

func ListHostingPrivateDatabase(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/hosting/privateDatabase", "", hostingprivatedatabaseColumnsToDisplay, flags.GenericFilters)
}

func GetHostingPrivateDatabase(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/hosting/privateDatabase", args[0], hostingprivatedatabaseTemplate)
}

func EditHostingPrivateDatabase(_ *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("/hosting/privateDatabase/%s", url.PathEscape(args[0]))
	if err := editor.EditResource(httpLib.Client, "/hosting/privateDatabase/{serviceName}", endpoint, hostingprivatedatabaseOpenapiSchema); err != nil {
		display.ExitError(err.Error())
	}
}
