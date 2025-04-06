
package cmd

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal"
)

var (
	xdslColumnsToDisplay = []string{ "accessName","accessType","provider","role","status" }
)

func listXdsl(_ *cobra.Command, _ []string) {
	req, err := client.NewRequest(http.MethodGet, "/xdsl", nil, true)
	if err != nil {
		fmt.Printf("error crafting request: %s\n", err)
		return
	}

	req.Header.Set("X-Pagination-Mode", "CachedObjectList-Pages")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("error fetching /xdsl: %s\n", err)
		return
	}

	var unmarshalled []map[string]any
	if err := client.UnmarshalResponse(resp, &unmarshalled); err != nil {
		fmt.Printf("error unmarshalling response: %s\n", err)
		return
	}

	internal.RenderTable(unmarshalled, xdslColumnsToDisplay)
}

func getXdsl(_ *cobra.Command, args []string) {
	var object map[string]any
	if err := client.Get(fmt.Sprintf("/xdsl/%s", url.PathEscape(args[0])), &object); err != nil {
		fmt.Printf("error fetching Xdsl: %s\n", err)
		return
	}

	internal.RenderObject(object, xdslColumnsToDisplay[0])
}

func init() {
	xdslCmd := &cobra.Command{
		Use:   "xdsl",
		Short: "Retrieve information and manage your Xdsl services",
	}

	// Command to list Xdsl services
	xdslCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List your Xdsl services",
		Run:   listXdsl,
	})

	// Command to get a single Xdsl
	xdslCmd.AddCommand(&cobra.Command{
		Use:        "get",
		Short:      "Retrieve information of a specific Xdsl",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"service_name"},
		Run:        getXdsl,
	})

	rootCmd.AddCommand(xdslCmd)
}
