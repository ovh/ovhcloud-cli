package xdsl

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
	xdslColumnsToDisplay = []string{"accessName", "accessType", "provider", "role", "status"}

	//go:embed templates/xdsl.tmpl
	xdslTemplate string

	//go:embed api-schemas/xdsl.json
	xdslOpenapiSchema []byte
)

func ListXdsl(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/xdsl", "", xdslColumnsToDisplay, flags.GenericFilters)
}

func GetXdsl(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/xdsl", args[0], xdslTemplate)
}

func EditXdsl(_ *cobra.Command, args []string) {
	url := fmt.Sprintf("/xdsl/%s", url.PathEscape(args[0]))
	editor.EditResource(httpLib.Client, "/xdsl/{serviceName}", url, xdslOpenapiSchema)
}
