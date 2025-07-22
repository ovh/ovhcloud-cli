package hostingprivatedatabase

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
	hostingprivatedatabaseColumnsToDisplay = []string{"serviceName", "displayName", "type", "version", "state"}

	//go:embed templates/hostingprivatedatabase.tmpl
	hostingprivatedatabaseTemplate string

	// HostingPrivateDatabaseDisplayName is the display name of the HostingPrivateDatabase
	HostingPrivateDatabaseDisplayName string
)

func ListHostingPrivateDatabase(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/hosting/privateDatabase", "", hostingprivatedatabaseColumnsToDisplay, flags.GenericFilters)
}

func GetHostingPrivateDatabase(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/hosting/privateDatabase", args[0], hostingprivatedatabaseTemplate)
}

func EditHostingPrivateDatabase(cmd *cobra.Command, args []string) {
	if err := common.EditResource(
		cmd,
		"/hosting/privateDatabase/{serviceName}",
		fmt.Sprintf("/hosting/privateDatabase/%s", url.PathEscape(args[0])),
		map[string]any{
			"displayName": HostingPrivateDatabaseDisplayName,
		},
		assets.HostingprivatedatabaseOpenapiSchema,
	); err != nil {
		display.ExitError(err.Error())
		return
	}
}
