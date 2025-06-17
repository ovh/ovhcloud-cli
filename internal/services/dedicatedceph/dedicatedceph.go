package dedicatedceph

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
	dedicatedcephColumnsToDisplay = []string{"serviceName", "region", "state", "status"}

	//go:embed templates/dedicatedceph.tmpl
	dedicatedcephTemplate string

	//go:embed api-schemas/dedicatedceph.json
	dedicatedcephOpenapiSchema []byte
)

func ListDedicatedCeph(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/dedicated/ceph", "", dedicatedcephColumnsToDisplay, flags.GenericFilters)
}

func GetDedicatedCeph(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/dedicated/ceph", args[0], dedicatedcephTemplate)
}

func EditDedicatedCeph(_ *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("/domain/%s", url.PathEscape(args[0]))
	if err := editor.EditResource(httpLib.Client, "/domain/{serviceName}", endpoint, dedicatedcephOpenapiSchema); err != nil {
		display.ExitError(err.Error())
	}
}
