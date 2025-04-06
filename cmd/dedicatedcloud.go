
package cmd

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal"
)

var (
	dedicatedcloudColumnsToDisplay = []string{ "serviceName","location","state","description" }
)

func listDedicatedCloud(_ *cobra.Command, _ []string) {
	req, err := client.NewRequest(http.MethodGet, "/dedicatedCloud", nil, true)
	if err != nil {
		fmt.Printf("error crafting request: %s\n", err)
		return
	}

	req.Header.Set("X-Pagination-Mode", "CachedObjectList-Pages")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("error fetching /dedicatedCloud: %s\n", err)
		return
	}

	var unmarshalled []map[string]any
	if err := client.UnmarshalResponse(resp, &unmarshalled); err != nil {
		fmt.Printf("error unmarshalling response: %s\n", err)
		return
	}

	internal.RenderTable(unmarshalled, dedicatedcloudColumnsToDisplay)
}

func getDedicatedCloud(_ *cobra.Command, args []string) {
	var object map[string]any
	if err := client.Get(fmt.Sprintf("/dedicatedCloud/%s", url.PathEscape(args[0])), &object); err != nil {
		fmt.Printf("error fetching DedicatedCloud: %s\n", err)
		return
	}

	internal.RenderObject(object, dedicatedcloudColumnsToDisplay[0])
}

func init() {
	dedicatedcloudCmd := &cobra.Command{
		Use:   "dedicatedcloud",
		Short: "Retrieve information and manage your DedicatedCloud services",
	}

	// Command to list DedicatedCloud services
	dedicatedcloudCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List your DedicatedCloud services",
		Run:   listDedicatedCloud,
	})

	// Command to get a single DedicatedCloud
	dedicatedcloudCmd.AddCommand(&cobra.Command{
		Use:        "get",
		Short:      "Retrieve information of a specific DedicatedCloud",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"service_name"},
		Run:        getDedicatedCloud,
	})

	rootCmd.AddCommand(dedicatedcloudCmd)
}
