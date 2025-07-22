package xdsl

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
	xdslColumnsToDisplay = []string{"accessName", "accessType", "provider", "role", "status"}

	//go:embed templates/xdsl.tmpl
	xdslTemplate string

	XdslSpec struct {
		Description  string `json:"description,omitempty"`
		LnsRateLimit int    `json:"lnsRateLimit,omitempty"`
		Monitoring   bool   `json:"monitoring,omitempty"`
	}
)

func ListXdsl(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/xdsl", "", xdslColumnsToDisplay, flags.GenericFilters)
}

func GetXdsl(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/xdsl", args[0], xdslTemplate)
}

func EditXdsl(cmd *cobra.Command, args []string) {
	if err := common.EditResource(
		cmd,
		"/xdsl/{serviceName}",
		fmt.Sprintf("/xdsl/%s", url.PathEscape(args[0])),
		XdslSpec,
		assets.XdslOpenapiSchema,
	); err != nil {
		display.ExitError(err.Error())
		return
	}
}
