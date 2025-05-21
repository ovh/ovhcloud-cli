package hostingprivatedatabase

import (
	_ "embed"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/flags"
	"stash.ovh.net/api/ovh-cli/internal/services/common"
)

var (
	hostingprivatedatabaseColumnsToDisplay = []string{"serviceName", "displayName", "type", "version", "state"}

	//go:embed templates/hostingprivatedatabase.tmpl
	hostingprivatedatabaseTemplate string
)

func ListHostingPrivateDatabase(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/hosting/privateDatabase", "", hostingprivatedatabaseColumnsToDisplay, flags.GenericFilters)
}

func GetHostingPrivateDatabase(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/hosting/privateDatabase", args[0], hostingprivatedatabaseTemplate)
}
