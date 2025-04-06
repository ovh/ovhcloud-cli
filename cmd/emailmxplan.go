
package cmd

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal"
)

var (
	emailmxplanColumnsToDisplay = []string{ "domain","displayName","state","offer" }
)

func listEmailMXPlan(_ *cobra.Command, _ []string) {
	req, err := client.NewRequest(http.MethodGet, "/email/mxplan", nil, true)
	if err != nil {
		fmt.Printf("error crafting request: %s\n", err)
		return
	}

	req.Header.Set("X-Pagination-Mode", "CachedObjectList-Pages")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("error fetching /email/mxplan: %s\n", err)
		return
	}

	var unmarshalled []map[string]any
	if err := client.UnmarshalResponse(resp, &unmarshalled); err != nil {
		fmt.Printf("error unmarshalling response: %s\n", err)
		return
	}

	internal.RenderTable(unmarshalled, emailmxplanColumnsToDisplay)
}

func getEmailMXPlan(_ *cobra.Command, args []string) {
	var object map[string]any
	if err := client.Get(fmt.Sprintf("/email/mxplan/%s", url.PathEscape(args[0])), &object); err != nil {
		fmt.Printf("error fetching EmailMXPlan: %s\n", err)
		return
	}

	internal.RenderObject(object, emailmxplanColumnsToDisplay[0])
}

func init() {
	emailmxplanCmd := &cobra.Command{
		Use:   "emailmxplan",
		Short: "Retrieve information and manage your EmailMXPlan services",
	}

	// Command to list EmailMXPlan services
	emailmxplanCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List your EmailMXPlan services",
		Run:   listEmailMXPlan,
	})

	// Command to get a single EmailMXPlan
	emailmxplanCmd.AddCommand(&cobra.Command{
		Use:        "get",
		Short:      "Retrieve information of a specific EmailMXPlan",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"service_name"},
		Run:        getEmailMXPlan,
	})

	rootCmd.AddCommand(emailmxplanCmd)
}
