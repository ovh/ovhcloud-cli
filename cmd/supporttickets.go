
package cmd

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal"
)

var (
	supportticketsColumnsToDisplay = []string{ "ticketId","serviceName","type","category","state" }
)

func listSupportTickets(_ *cobra.Command, _ []string) {
	req, err := client.NewRequest(http.MethodGet, "/support/tickets", nil, true)
	if err != nil {
		fmt.Printf("error crafting request: %s\n", err)
		return
	}

	req.Header.Set("X-Pagination-Mode", "CachedObjectList-Pages")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("error fetching /support/tickets: %s\n", err)
		return
	}

	var unmarshalled []map[string]any
	if err := client.UnmarshalResponse(resp, &unmarshalled); err != nil {
		fmt.Printf("error unmarshalling response: %s\n", err)
		return
	}

	internal.RenderTable(unmarshalled, supportticketsColumnsToDisplay)
}

func getSupportTickets(_ *cobra.Command, args []string) {
	var object map[string]any
	if err := client.Get(fmt.Sprintf("/support/tickets/%s", url.PathEscape(args[0])), &object); err != nil {
		fmt.Printf("error fetching SupportTickets: %s\n", err)
		return
	}

	internal.RenderObject(object, supportticketsColumnsToDisplay[0])
}

func init() {
	supportticketsCmd := &cobra.Command{
		Use:   "supporttickets",
		Short: "Retrieve information and manage your SupportTickets services",
	}

	// Command to list SupportTickets services
	supportticketsCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List your SupportTickets services",
		Run:   listSupportTickets,
	})

	// Command to get a single SupportTickets
	supportticketsCmd.AddCommand(&cobra.Command{
		Use:        "get",
		Short:      "Retrieve information of a specific SupportTickets",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"service_name"},
		Run:        getSupportTickets,
	})

	rootCmd.AddCommand(supportticketsCmd)
}
