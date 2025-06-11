package vps

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
	vpsColumnsToDisplay = []string{"name", "displayName", "state", "zone"}

	//go:embed templates/vps.tmpl
	vpsTemplate string

	//go:embed api-schemas/vps.json
	vpsOpenapiSchema []byte
)

func ListVps(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/vps", "", vpsColumnsToDisplay, flags.GenericFilters)
}

func GetVps(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/vps", args[0], vpsTemplate)
}

func EditVps(_ *cobra.Command, args []string) {
	url := fmt.Sprintf("/vps/%s", url.PathEscape(args[0]))
	if err := editor.EditResource(httpLib.Client, "/vps/{serviceName}", url, vpsOpenapiSchema); err != nil {
		display.ExitError(err.Error())
	}
}
