
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
	smsColumnsToDisplay = []string{ "name","status" }
)

func listSms(_ *cobra.Command, _ []string) {
	req, err := client.NewRequest(http.MethodGet, "/sms", nil, true)
	if err != nil {
		log.Fatalf("error crafting request: %s\n", err)
	}

	req.Header.Set("X-Pagination-Mode", "CachedObjectList-Pages")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("error fetching /sms: %s\n", err)
	}

	var unmarshalled []map[string]any
	if err := client.UnmarshalResponse(resp, &unmarshalled); err != nil {
		log.Fatalf("error unmarshalling response: %s\n", err)
	}

	internal.OutputTable(unmarshalled, smsColumnsToDisplay, jsonOutput, yamlOutput)
}

func getSms(_ *cobra.Command, args []string) {
	var object map[string]any
	if err := client.Get(fmt.Sprintf("/sms/%s", url.PathEscape(args[0])), &object); err != nil {
		log.Fatalf("error fetching Sms: %s\n", err)
	}

	internal.OutputObject(object, smsColumnsToDisplay[0], jsonOutput, yamlOutput)
}

func init() {
	smsCmd := &cobra.Command{
		Use:   "sms",
		Short: "Retrieve information and manage your Sms services",
	}

	// Command to list Sms services
	smsCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List your Sms services",
		Run:   listSms,
	})

	// Command to get a single Sms
	smsCmd.AddCommand(&cobra.Command{
		Use:        "get",
		Short:      "Retrieve information of a specific Sms",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"service_name"},
		Run:        getSms,
	})

	rootCmd.AddCommand(smsCmd)
}
