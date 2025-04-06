
package cmd

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal"
)

var (
	ipColumnsToDisplay = []string{ "ip","rir","routedTo","country","description" }
)

func listIp(_ *cobra.Command, _ []string) {
	req, err := client.NewRequest(http.MethodGet, "/ip", nil, true)
	if err != nil {
		fmt.Printf("error crafting request: %s\n", err)
		return
	}

	req.Header.Set("X-Pagination-Mode", "CachedObjectList-Pages")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("error fetching /ip: %s\n", err)
		return
	}

	var unmarshalled []map[string]any
	if err := client.UnmarshalResponse(resp, &unmarshalled); err != nil {
		fmt.Printf("error unmarshalling response: %s\n", err)
		return
	}

	internal.RenderTable(unmarshalled, ipColumnsToDisplay)
}

func getIp(_ *cobra.Command, args []string) {
	var object map[string]any
	if err := client.Get(fmt.Sprintf("/ip/%s", url.PathEscape(args[0])), &object); err != nil {
		fmt.Printf("error fetching Ip: %s\n", err)
		return
	}

	internal.RenderObject(object, ipColumnsToDisplay[0])
}

func init() {
	ipCmd := &cobra.Command{
		Use:   "ip",
		Short: "Retrieve information and manage your Ip services",
	}

	// Command to list Ip services
	ipCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List your Ip services",
		Run:   listIp,
	})

	// Command to get a single Ip
	ipCmd.AddCommand(&cobra.Command{
		Use:        "get",
		Short:      "Retrieve information of a specific Ip",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"service_name"},
		Run:        getIp,
	})

	rootCmd.AddCommand(ipCmd)
}
