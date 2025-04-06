
package cmd

import (
	"fmt"
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
		fmt.Printf("error crafting request: %s\n", err)
		return
	}

	req.Header.Set("X-Pagination-Mode", "CachedObjectList-Pages")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("error fetching /email/domain: %s\n", err)
		return
	}

	var unmarshalled []map[string]any
	if err := client.UnmarshalResponse(resp, &unmarshalled); err != nil {
		fmt.Printf("error unmarshalling response: %s\n", err)
		return
	}

	internal.RenderTable(unmarshalled, emaildomainColumnsToDisplay)
}

func getEmailDomain(_ *cobra.Command, args []string) {
	var object map[string]any
	if err := client.Get(fmt.Sprintf("/email/domain/%s", url.PathEscape(args[0])), &object); err != nil {
		fmt.Printf("error fetching EmailDomain: %s\n", err)
		return
	}

	internal.RenderObject(object, emaildomainColumnsToDisplay[0])
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
