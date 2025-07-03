package webhosting

import (
	_ "embed"
	"fmt"
	"net/url"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/display"
	"stash.ovh.net/api/ovh-cli/internal/flags"
	"stash.ovh.net/api/ovh-cli/internal/services/common"
)

var (
	webhostingColumnsToDisplay = []string{"serviceName", "displayName", "datacenter", "state"}

	//go:embed templates/webhosting.tmpl
	webhostingTemplate string

	//go:embed api-schemas/webhosting.json
	webhostingOpenapiSchema []byte

	WebHostingSpec struct {
		DisplayName string `json:"displayName,omitempty"`
	}
)

func ListWebHosting(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/hosting/web", "", webhostingColumnsToDisplay, flags.GenericFilters)
}

func GetWebHosting(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/hosting/web", args[0], webhostingTemplate)
}

func EditWebHosting(cmd *cobra.Command, args []string) {
	if err := common.EditResource(
		cmd,
		"/hosting/web/{serviceName}",
		fmt.Sprintf("/hosting/web/%s", url.PathEscape(args[0])),
		WebHostingSpec,
		webhostingOpenapiSchema,
	); err != nil {
		display.ExitError(err.Error())
		return
	}
}
