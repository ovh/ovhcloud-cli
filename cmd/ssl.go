
package cmd

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal"
)

var (
	sslColumnsToDisplay = []string{ "serviceName","type","authority","status" }
)

func listSsl(_ *cobra.Command, _ []string) {
	req, err := client.NewRequest(http.MethodGet, "/ssl", nil, true)
	if err != nil {
		fmt.Printf("error crafting request: %s\n", err)
		return
	}

	req.Header.Set("X-Pagination-Mode", "CachedObjectList-Pages")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("error fetching /ssl: %s\n", err)
		return
	}

	var unmarshalled []map[string]any
	if err := client.UnmarshalResponse(resp, &unmarshalled); err != nil {
		fmt.Printf("error unmarshalling response: %s\n", err)
		return
	}

	internal.RenderTable(unmarshalled, sslColumnsToDisplay)
}

func getSsl(_ *cobra.Command, args []string) {
	var object map[string]any
	if err := client.Get(fmt.Sprintf("/ssl/%s", url.PathEscape(args[0])), &object); err != nil {
		fmt.Printf("error fetching Ssl: %s\n", err)
		return
	}

	internal.RenderObject(object, sslColumnsToDisplay[0])
}

func init() {
	sslCmd := &cobra.Command{
		Use:   "ssl",
		Short: "Retrieve information and manage your Ssl services",
	}

	// Command to list Ssl services
	sslCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List your Ssl services",
		Run:   listSsl,
	})

	// Command to get a single Ssl
	sslCmd.AddCommand(&cobra.Command{
		Use:        "get",
		Short:      "Retrieve information of a specific Ssl",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"service_name"},
		Run:        getSsl,
	})

	rootCmd.AddCommand(sslCmd)
}
