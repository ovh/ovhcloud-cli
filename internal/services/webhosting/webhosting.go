package webhosting

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
	webhostingColumnsToDisplay = []string{"serviceName", "displayName", "datacenter", "state"}

	//go:embed templates/webhosting.tmpl
	webhostingTemplate string

	//go:embed api-schemas/webhosting.json
	webhostingOpenapiSchema []byte
)

func ListWebHosting(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/hosting/web", "", webhostingColumnsToDisplay, flags.GenericFilters)
}

func GetWebHosting(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/hosting/web", args[0], webhostingTemplate)
}

func EditWebHosting(_ *cobra.Command, args []string) {
	url := fmt.Sprintf("/hosting/web/%s", url.PathEscape(args[0]))
	editor.EditResource(httpLib.Client, "/hosting/web/{serviceName}", url, webhostingOpenapiSchema)
}
