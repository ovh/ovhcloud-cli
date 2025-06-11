package ldp

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
	ldpColumnsToDisplay = []string{"serviceName", "displayName", "isClusterOwner", "state", "username"}

	//go:embed templates/ldp.tmpl
	ldpTemplate string

	//go:embed api-schemas/ldp.json
	ldpOpenapiSchema []byte
)

func ListLdp(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/dbaas/logs", "", ldpColumnsToDisplay, flags.GenericFilters)
}

func GetLdp(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/dbaas/logs", args[0], ldpTemplate)
}

func EditLdp(_ *cobra.Command, args []string) {
	url := fmt.Sprintf("/dbaas/logs/%s", url.PathEscape(args[0]))
	if err := editor.EditResource(httpLib.Client, "/dbaas/logs/{serviceName}", url, ldpOpenapiSchema); err != nil {
		display.ExitError(err.Error())
	}
}
