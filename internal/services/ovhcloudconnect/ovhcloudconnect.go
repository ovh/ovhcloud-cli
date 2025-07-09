package ovhcloudconnect

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
	ovhcloudconnectColumnsToDisplay = []string{"uuid", "provider", "status", "description"}

	//go:embed templates/ovhcloudconnect.tmpl
	ovhcloudconnectTemplate string

	//go:embed api-schemas/ovhcloudconnect.json
	ovhcloudconnectOpenapiSchema []byte

	OvhCloudConnectSpec struct {
		Description string `json:"description,omitempty"`
	}
)

func ListOvhCloudConnect(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/ovhCloudConnect", "", ovhcloudconnectColumnsToDisplay, flags.GenericFilters)
}

func GetOvhCloudConnect(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/ovhCloudConnect", args[0], ovhcloudconnectTemplate)
}

func EditOvhCloudConnect(cmd *cobra.Command, args []string) {
	if err := common.EditResource(
		cmd,
		"/ovhCloudConnect/{serviceName}",
		fmt.Sprintf("/ovhCloudConnect/%s", url.PathEscape(args[0])),
		OvhCloudConnectSpec,
		ovhcloudconnectOpenapiSchema,
	); err != nil {
		display.ExitError(err.Error())
		return
	}
}
