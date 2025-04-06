
package cmd

import (
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal"
)

var (
	telephonyColumnsToDisplay = []string{ "billingAccount","description","status" }
)

func listTelephony(_ *cobra.Command, _ []string) {
	req, err := client.NewRequest(http.MethodGet, "/telephony", nil, true)
	if err != nil {
		log.Fatalf("error crafting request: %s\n", err)
	}

	req.Header.Set("X-Pagination-Mode", "CachedObjectList-Pages")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("error fetching /telephony: %s\n", err)
	}

	var unmarshalled []map[string]any
	if err := client.UnmarshalResponse(resp, &unmarshalled); err != nil {
		log.Fatalf("error unmarshalling response: %s\n", err)
	}

	internal.OutputTable(unmarshalled, telephonyColumnsToDisplay, jsonOutput, yamlOutput)
}

func getTelephony(_ *cobra.Command, args []string) {
	var object map[string]any
	if err := client.Get(fmt.Sprintf("/telephony/%s", url.PathEscape(args[0])), &object); err != nil {
		log.Fatalf("error fetching Telephony: %s\n", err)
	}

	internal.OutputObject(object, telephonyColumnsToDisplay[0], jsonOutput, yamlOutput)
}

func init() {
	telephonyCmd := &cobra.Command{
		Use:   "telephony",
		Short: "Retrieve information and manage your Telephony services",
	}

	// Command to list Telephony services
	telephonyCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List your Telephony services",
		Run:   listTelephony,
	})

	// Command to get a single Telephony
	telephonyCmd.AddCommand(&cobra.Command{
		Use:        "get",
		Short:      "Retrieve information of a specific Telephony",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"service_name"},
		Run:        getTelephony,
	})

	rootCmd.AddCommand(telephonyCmd)
}
