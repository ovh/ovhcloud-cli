
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
	emaildomainColumnsToDisplay = []string{ "domain","status","offer" }
)

func listEmailDomain(_ *cobra.Command, _ []string) {
	req, err := client.NewRequest(http.MethodGet, "/email/domain", nil, true)
	if err != nil {
		log.Fatalf("error crafting request: %s\n", err)
	}

	req.Header.Set("X-Pagination-Mode", "CachedObjectList-Pages")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("error fetching /email/domain: %s\n", err)
	}

	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("error reading response body: %s", err)
	}

	internal.OutputTable(bodyBytes, emaildomainColumnsToDisplay, jsonOutput, yamlOutput)
}

func getEmailDomain(_ *cobra.Command, args []string) {
	var object map[string]any
	if err := client.Get(fmt.Sprintf("/email/domain/%s", url.PathEscape(args[0])), &object); err != nil {
		log.Fatalf("error fetching EmailDomain: %s\n", err)
	}

	internal.OutputObject(object, emaildomainColumnsToDisplay[0], jsonOutput, yamlOutput)
}

func init() {
	emaildomainCmd := &cobra.Command{
		Use:   "emaildomain",
		Short: "Retrieve information and manage your EmailDomain services",
	}

	// Command to list EmailDomain services
	emaildomainCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List your EmailDomain services",
		Run:   listEmailDomain,
	})

	// Command to get a single EmailDomain
	emaildomainCmd.AddCommand(&cobra.Command{
		Use:        "get",
		Short:      "Retrieve information of a specific EmailDomain",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"service_name"},
		Run:        getEmailDomain,
	})

	rootCmd.AddCommand(emaildomainCmd)
}
