
package cmd

import (
	"fmt"
	"io"
	"log"
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
		log.Fatalf("error crafting request: %s\n", err)
	}

	req.Header.Set("X-Pagination-Mode", "CachedObjectList-Pages")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("error fetching /support/tickets: %s\n", err)
	}

	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("error reading response body: %s", err)
	}

	internal.OutputTable(bodyBytes, supportticketsColumnsToDisplay, jsonOutput, yamlOutput)
}

func getSupportTickets(_ *cobra.Command, args []string) {
	var object map[string]any
	if err := client.Get(fmt.Sprintf("/support/tickets/%s", url.PathEscape(args[0])), &object); err != nil {
		log.Fatalf("error fetching SupportTickets: %s\n", err)
	}

	internal.OutputObject(object, supportticketsColumnsToDisplay[0], jsonOutput, yamlOutput)
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
