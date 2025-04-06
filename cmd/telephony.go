
package cmd

import (
	"fmt"
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
		fmt.Printf("error crafting request: %s\n", err)
		return
	}

	req.Header.Set("X-Pagination-Mode", "CachedObjectList-Pages")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("error fetching /telephony: %s\n", err)
		return
	}

	var unmarshalled []map[string]any
	if err := client.UnmarshalResponse(resp, &unmarshalled); err != nil {
		fmt.Printf("error unmarshalling response: %s\n", err)
		return
	}

	internal.RenderTable(unmarshalled, telephonyColumnsToDisplay)
}

func getTelephony(_ *cobra.Command, args []string) {
	var object map[string]any
	if err := client.Get(fmt.Sprintf("/telephony/%s", url.PathEscape(args[0])), &object); err != nil {
		fmt.Printf("error fetching Telephony: %s\n", err)
		return
	}

	internal.RenderObject(object, telephonyColumnsToDisplay[0])
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
