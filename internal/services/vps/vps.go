package vps

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
	vpsColumnsToDisplay = []string{"name", "displayName", "state", "zone"}

	//go:embed templates/vps.tmpl
	vpsTemplate string

	//go:embed api-schemas/vps.json
	vpsOpenapiSchema []byte

	VPSSpec struct {
		DisplayName   string `json:"displayName,omitempty"`
		Keymap        string `json:"keymap,omitempty"`
		NetbootMode   string `json:"netbootMode,omitempty"`
		SlaMonitoring bool   `json:"slaMonitoring,omitempty"`
	}
)

func ListVps(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/vps", "", vpsColumnsToDisplay, flags.GenericFilters)
}

func GetVps(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/vps", args[0], vpsTemplate)
}

func EditVps(cmd *cobra.Command, args []string) {
	if err := common.EditResource(
		cmd,
		"/vps/{serviceName}",
		fmt.Sprintf("/vps/%s", url.PathEscape(args[0])),
		VPSSpec,
		vpsOpenapiSchema,
	); err != nil {
		display.ExitError(err.Error())
		return
	}
}
