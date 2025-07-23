package ldp

import (
	_ "embed"
	"fmt"
	"net/url"

	"github.com/ovh/ovhcloud-cli/internal/assets"
	"github.com/ovh/ovhcloud-cli/internal/display"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	"github.com/ovh/ovhcloud-cli/internal/services/common"
	"github.com/spf13/cobra"
)

var (
	ldpColumnsToDisplay = []string{"serviceName", "displayName", "isClusterOwner", "state", "username"}

	//go:embed templates/ldp.tmpl
	ldpTemplate string

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
		assets.LdpOpenapiSchema,
	); err != nil {
		display.ExitError(err.Error())
		return
	}
}
