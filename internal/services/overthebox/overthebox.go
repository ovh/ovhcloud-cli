package overthebox

import (
	_ "embed"
	"fmt"
	"net/url"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/assets"
	"stash.ovh.net/api/ovh-cli/internal/display"
	"stash.ovh.net/api/ovh-cli/internal/flags"
	"stash.ovh.net/api/ovh-cli/internal/services/common"
)

var (
	overtheboxColumnsToDisplay = []string{"serviceName", "offer", "status", "bandwidth"}

	//go:embed templates/overthebox.tmpl
	overtheboxTemplate string

	OverTheBoxSpec struct {
		AutoUpgrade         bool   `json:"autoUpgrade,omitempty"`
		CustomerDescription string `json:"customerDescription,omitempty"`
		ReleaseChannel      string `json:"releaseChannel,omitempty"`
	}
)

func ListOverTheBox(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/overTheBox", "", overtheboxColumnsToDisplay, flags.GenericFilters)
}

func GetOverTheBox(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/overTheBox", args[0], overtheboxTemplate)
}

func EditOverTheBox(cmd *cobra.Command, args []string) {
	if err := common.EditResource(
		cmd,
		"/overTheBox/{serviceName}",
		fmt.Sprintf("/overTheBox/%s", url.PathEscape(args[0])),
		OverTheBoxSpec,
		assets.OvertheboxOpenapiSchema,
	); err != nil {
		display.ExitError(err.Error())
		return
	}
}
