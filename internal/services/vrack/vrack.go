package vrack

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
	vrackColumnsToDisplay = []string{"serviceName", "name", "description"}

	//go:embed templates/vrack.tmpl
	vrackTemplate string

	VrackSpec struct {
		Name        string `json:"name,omitempty"`
		Description string `json:"description,omitempty"`
	}
)

func ListVrack(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/vrack", "", vrackColumnsToDisplay, flags.GenericFilters)
}

func GetVrack(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/vrack", args[0], vrackTemplate)
}

func EditVrack(cmd *cobra.Command, args []string) {
	if err := common.EditResource(
		cmd,
		"/vrack/{serviceName}",
		fmt.Sprintf("/vrack/%s", url.PathEscape(args[0])),
		VrackSpec,
		assets.VrackOpenapiSchema,
	); err != nil {
		display.ExitError(err.Error())
		return
	}
}
