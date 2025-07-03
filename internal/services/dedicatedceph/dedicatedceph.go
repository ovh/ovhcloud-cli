package dedicatedceph

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
	dedicatedcephColumnsToDisplay = []string{"serviceName", "region", "state", "status"}

	//go:embed templates/dedicatedceph.tmpl
	dedicatedcephTemplate string

	//go:embed api-schemas/dedicatedceph.json
	dedicatedcephOpenapiSchema []byte

	DedicatedCephSpec struct {
		CrushTunables string `json:"crushTunables,omitempty"`
		Label         string `json:"label,omitempty"`
	}
)

func ListDedicatedCeph(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/dedicated/ceph", "", dedicatedcephColumnsToDisplay, flags.GenericFilters)
}

func GetDedicatedCeph(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/dedicated/ceph", args[0], dedicatedcephTemplate)
}

func EditDedicatedCeph(cmd *cobra.Command, args []string) {
	if err := common.EditResource(
		cmd,
		"/dedicated/ceph/{serviceName}",
		fmt.Sprintf("/dedicated/ceph/%s", url.PathEscape(args[0])),
		DedicatedCephSpec,
		dedicatedcephOpenapiSchema,
	); err != nil {
		display.ExitError(err.Error())
		return
	}
}
