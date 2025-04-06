
package cmd

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal"
)

var (
	overtheboxColumnsToDisplay = []string{ "serviceName","offer","status","bandwidth" }
)

func listOverTheBox(_ *cobra.Command, _ []string) {
	req, err := client.NewRequest(http.MethodGet, "/overTheBox", nil, true)
	if err != nil {
		fmt.Printf("error crafting request: %s\n", err)
		return
	}

	req.Header.Set("X-Pagination-Mode", "CachedObjectList-Pages")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("error fetching /overTheBox: %s\n", err)
		return
	}

	var unmarshalled []map[string]any
	if err := client.UnmarshalResponse(resp, &unmarshalled); err != nil {
		fmt.Printf("error unmarshalling response: %s\n", err)
		return
	}

	internal.RenderTable(unmarshalled, overtheboxColumnsToDisplay)
}

func getOverTheBox(_ *cobra.Command, args []string) {
	var object map[string]any
	if err := client.Get(fmt.Sprintf("/overTheBox/%s", url.PathEscape(args[0])), &object); err != nil {
		fmt.Printf("error fetching OverTheBox: %s\n", err)
		return
	}

	internal.RenderObject(object, overtheboxColumnsToDisplay[0])
}

func init() {
	overtheboxCmd := &cobra.Command{
		Use:   "overthebox",
		Short: "Retrieve information and manage your OverTheBox services",
	}

	// Command to list OverTheBox services
	overtheboxCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List your OverTheBox services",
		Run:   listOverTheBox,
	})

	// Command to get a single OverTheBox
	overtheboxCmd.AddCommand(&cobra.Command{
		Use:        "get",
		Short:      "Retrieve information of a specific OverTheBox",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"service_name"},
		Run:        getOverTheBox,
	})

	rootCmd.AddCommand(overtheboxCmd)
}
