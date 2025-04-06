
package cmd

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal"
)

var (
	alldomColumnsToDisplay = []string{ "name","type","offer" }
)

func listAllDom(_ *cobra.Command, _ []string) {
	req, err := client.NewRequest(http.MethodGet, "/allDom", nil, true)
	if err != nil {
		fmt.Printf("error crafting request: %s\n", err)
		return
	}

	req.Header.Set("X-Pagination-Mode", "CachedObjectList-Pages")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("error fetching /allDom: %s\n", err)
		return
	}

	var unmarshalled []map[string]any
	if err := client.UnmarshalResponse(resp, &unmarshalled); err != nil {
		fmt.Printf("error unmarshalling response: %s\n", err)
		return
	}

	internal.RenderTable(unmarshalled, alldomColumnsToDisplay)
}

func getAllDom(_ *cobra.Command, args []string) {
	var object map[string]any
	if err := client.Get(fmt.Sprintf("/allDom/%s", url.PathEscape(args[0])), &object); err != nil {
		fmt.Printf("error fetching AllDom: %s\n", err)
		return
	}

	internal.RenderObject(object, alldomColumnsToDisplay[0])
}

func init() {
	alldomCmd := &cobra.Command{
		Use:   "alldom",
		Short: "Retrieve information and manage your AllDom services",
	}

	// Command to list AllDom services
	alldomCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List your AllDom services",
		Run:   listAllDom,
	})

	// Command to get a single AllDom
	alldomCmd.AddCommand(&cobra.Command{
		Use:        "get",
		Short:      "Retrieve information of a specific AllDom",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"service_name"},
		Run:        getAllDom,
	})

	rootCmd.AddCommand(alldomCmd)
}
