package ovhcloudconnect

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
	ovhcloudconnectColumnsToDisplay = []string{"uuid", "provider", "status", "description"}

	//go:embed templates/ovhcloudconnect.tmpl
	ovhcloudconnectTemplate string

	//go:embed api-schemas/ovhcloudconnect.json
	ovhcloudconnectOpenapiSchema []byte
)

func ListOvhCloudConnect(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/ovhCloudConnect", "", ovhcloudconnectColumnsToDisplay, flags.GenericFilters)
}

func GetOvhCloudConnect(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/ovhCloudConnect", args[0], ovhcloudconnectTemplate)
}

func EditOvhCloudConnect(_ *cobra.Command, args []string) {
	url := fmt.Sprintf("/ovhCloudConnect/%s", url.PathEscape(args[0]))
	editor.EditResource(httpLib.Client, "/ovhCloudConnect/{serviceName}", url, ovhcloudconnectOpenapiSchema)
}
