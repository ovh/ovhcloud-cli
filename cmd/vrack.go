
package cmd

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal"
)

var (
	vrackColumnsToDisplay = []string{ "name","description" }
)

func listVrack(_ *cobra.Command, _ []string) {
	req, err := client.NewRequest(http.MethodGet, "/vrack", nil, true)
	if err != nil {
		fmt.Printf("error crafting request: %s\n", err)
		return
	}

	req.Header.Set("X-Pagination-Mode", "CachedObjectList-Pages")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("error fetching /vrack: %s\n", err)
		return
	}

	var unmarshalled []map[string]any
	if err := client.UnmarshalResponse(resp, &unmarshalled); err != nil {
		fmt.Printf("error unmarshalling response: %s\n", err)
		return
	}

	internal.RenderTable(unmarshalled, vrackColumnsToDisplay)
}

func getVrack(_ *cobra.Command, args []string) {
	var object map[string]any
	if err := client.Get(fmt.Sprintf("/vrack/%s", url.PathEscape(args[0])), &object); err != nil {
		fmt.Printf("error fetching Vrack: %s\n", err)
		return
	}

	internal.RenderObject(object, vrackColumnsToDisplay[0])
}

func init() {
	vrackCmd := &cobra.Command{
		Use:   "vrack",
		Short: "Retrieve information and manage your Vrack services",
	}

	// Command to list Vrack services
	vrackCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List your Vrack services",
		Run:   listVrack,
	})

	// Command to get a single Vrack
	vrackCmd.AddCommand(&cobra.Command{
		Use:        "get",
		Short:      "Retrieve information of a specific Vrack",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"service_name"},
		Run:        getVrack,
	})

	rootCmd.AddCommand(vrackCmd)
}
