
package cmd

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal"
)

var (
	baremetalColumnsToDisplay = []string{ "serverId","name","region","os" }
)

func listBaremetal(_ *cobra.Command, _ []string) {
	req, err := client.NewRequest(http.MethodGet, "/dedicated/server", nil, true)
	if err != nil {
		fmt.Printf("error crafting request: %s\n", err)
		return
	}

	req.Header.Set("X-Pagination-Mode", "CachedObjectList-Pages")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("error fetching /dedicated/server: %s\n", err)
		return
	}

	var unmarshalled []map[string]any
	if err := client.UnmarshalResponse(resp, &unmarshalled); err != nil {
		fmt.Printf("error unmarshalling response: %s\n", err)
		return
	}

	internal.RenderTable(unmarshalled, baremetalColumnsToDisplay)
}

func getBaremetal(_ *cobra.Command, args []string) {
	var object map[string]any
	if err := client.Get(fmt.Sprintf("/dedicated/server/%s", url.PathEscape(args[0])), &object); err != nil {
		fmt.Printf("error fetching Baremetal: %s\n", err)
		return
	}

	internal.RenderObject(object, baremetalColumnsToDisplay[0])
}

func init() {
	baremetalCmd := &cobra.Command{
		Use:   "baremetal",
		Short: "Retrieve information and manage your Baremetal services",
	}

	// Command to list Baremetal services
	baremetalCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List your Baremetal services",
		Run:   listBaremetal,
	})

	// Command to get a single Baremetal
	baremetalCmd.AddCommand(&cobra.Command{
		Use:        "get",
		Short:      "Retrieve information of a specific Baremetal",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"service_name"},
		Run:        getBaremetal,
	})

	rootCmd.AddCommand(baremetalCmd)
}
