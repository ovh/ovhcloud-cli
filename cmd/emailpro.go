
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
	emailproColumnsToDisplay = []string{ "domain","displayName","state","offer" }
)

func listEmailPro(_ *cobra.Command, _ []string) {
	req, err := client.NewRequest(http.MethodGet, "/email/pro", nil, true)
	if err != nil {
		log.Fatalf("error crafting request: %s\n", err)
	}

	req.Header.Set("X-Pagination-Mode", "CachedObjectList-Pages")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("error fetching /email/pro: %s\n", err)
	}

	var unmarshalled []map[string]any
	if err := client.UnmarshalResponse(resp, &unmarshalled); err != nil {
		log.Fatalf("error unmarshalling response: %s\n", err)
	}

	internal.OutputTable(unmarshalled, emailproColumnsToDisplay, jsonOutput, yamlOutput)
}

func getEmailPro(_ *cobra.Command, args []string) {
	var object map[string]any
	if err := client.Get(fmt.Sprintf("/email/pro/%s", url.PathEscape(args[0])), &object); err != nil {
		log.Fatalf("error fetching EmailPro: %s\n", err)
	}

	internal.OutputObject(object, emailproColumnsToDisplay[0], jsonOutput, yamlOutput)
}

func init() {
	emailproCmd := &cobra.Command{
		Use:   "emailpro",
		Short: "Retrieve information and manage your EmailPro services",
	}

	// Command to list EmailPro services
	emailproCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List your EmailPro services",
		Run:   listEmailPro,
	})

	// Command to get a single EmailPro
	emailproCmd.AddCommand(&cobra.Command{
		Use:        "get",
		Short:      "Retrieve information of a specific EmailPro",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"service_name"},
		Run:        getEmailPro,
	})

	rootCmd.AddCommand(emailproCmd)
}
