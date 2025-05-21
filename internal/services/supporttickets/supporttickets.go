package supporttickets

import (
	_ "embed"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/flags"
	"stash.ovh.net/api/ovh-cli/internal/services/common"
)

var (
	supportticketsColumnsToDisplay = []string{"ticketId", "serviceName", "type", "category", "state"}

	//go:embed templates/supporttickets.tmpl
	supportticketsTemplate string
)

func ListSupportTickets(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/support/tickets", "", supportticketsColumnsToDisplay, flags.GenericFilters)
}

func GetSupportTickets(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/support/tickets", args[0], supportticketsTemplate)
}
