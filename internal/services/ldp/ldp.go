package ldp

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
	ldpColumnsToDisplay = []string{"serviceName", "displayName", "isClusterOwner", "state", "username"}

	//go:embed templates/ldp.tmpl
	ldpTemplate string

	//go:embed api-schemas/ldp.json
	ldpOpenapiSchema []byte

	LdpSpec struct {
		DisplayName string `json:"displayName,omitempty"`
		EnableIAM   bool   `json:"enableIam"`
	}
)

func ListLdp(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/dbaas/logs", "", ldpColumnsToDisplay, flags.GenericFilters)
}

func GetLdp(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/dbaas/logs", args[0], ldpTemplate)
}

func EditLdp(cmd *cobra.Command, args []string) {
	if err := common.EditResource(
		cmd,
		"/dbaas/logs/{serviceName}",
		fmt.Sprintf("/dbaas/logs/%s", url.PathEscape(args[0])),
		LdpSpec,
		ldpOpenapiSchema,
	); err != nil {
		display.ExitError(err.Error())
		return
	}
}
