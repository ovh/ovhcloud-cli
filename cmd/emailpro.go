
package cmd

import (
	"fmt"
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
		fmt.Printf("error crafting request: %s\n", err)
		return
	}

	req.Header.Set("X-Pagination-Mode", "CachedObjectList-Pages")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("error fetching /email/pro: %s\n", err)
		return
	}

	var unmarshalled []map[string]any
	if err := client.UnmarshalResponse(resp, &unmarshalled); err != nil {
		fmt.Printf("error unmarshalling response: %s\n", err)
		return
	}

	internal.RenderTable(unmarshalled, emailproColumnsToDisplay)
}

func getEmailPro(_ *cobra.Command, args []string) {
	var object map[string]any
	if err := client.Get(fmt.Sprintf("/email/pro/%s", url.PathEscape(args[0])), &object); err != nil {
		fmt.Printf("error fetching EmailPro: %s\n", err)
		return
	}

	internal.RenderObject(object, emailproColumnsToDisplay[0])
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
