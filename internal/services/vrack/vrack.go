package vrack

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
