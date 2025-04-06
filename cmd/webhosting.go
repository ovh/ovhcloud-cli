
package cmd

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal"
)

var (
	webhostingColumnsToDisplay = []string{ "serviceName","displayName","datacenter","state" }
)

func listWebHosting(_ *cobra.Command, _ []string) {
	req, err := client.NewRequest(http.MethodGet, "/hosting/web", nil, true)
	if err != nil {
		fmt.Printf("error crafting request: %s\n", err)
		return
	}

	req.Header.Set("X-Pagination-Mode", "CachedObjectList-Pages")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("error fetching /hosting/web: %s\n", err)
		return
	}

	var unmarshalled []map[string]any
	if err := client.UnmarshalResponse(resp, &unmarshalled); err != nil {
		fmt.Printf("error unmarshalling response: %s\n", err)
		return
	}

	internal.RenderTable(unmarshalled, webhostingColumnsToDisplay)
}

func getWebHosting(_ *cobra.Command, args []string) {
	var object map[string]any
	if err := client.Get(fmt.Sprintf("/hosting/web/%s", url.PathEscape(args[0])), &object); err != nil {
		fmt.Printf("error fetching WebHosting: %s\n", err)
		return
	}

	internal.RenderObject(object, webhostingColumnsToDisplay[0])
}

func init() {
	webhostingCmd := &cobra.Command{
		Use:   "webhosting",
		Short: "Retrieve information and manage your WebHosting services",
	}

	// Command to list WebHosting services
	webhostingCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List your WebHosting services",
		Run:   listWebHosting,
	})

	// Command to get a single WebHosting
	webhostingCmd.AddCommand(&cobra.Command{
		Use:        "get",
		Short:      "Retrieve information of a specific WebHosting",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"service_name"},
		Run:        getWebHosting,
	})

	rootCmd.AddCommand(webhostingCmd)
}
