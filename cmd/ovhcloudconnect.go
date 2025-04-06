
package cmd

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal"
)

var (
	ovhcloudconnectColumnsToDisplay = []string{ "uuid","provider","status","description" }
)

func listOvhCloudConnect(_ *cobra.Command, _ []string) {
	req, err := client.NewRequest(http.MethodGet, "/ovhCloudConnect", nil, true)
	if err != nil {
		fmt.Printf("error crafting request: %s\n", err)
		return
	}

	req.Header.Set("X-Pagination-Mode", "CachedObjectList-Pages")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("error fetching /ovhCloudConnect: %s\n", err)
		return
	}

	var unmarshalled []map[string]any
	if err := client.UnmarshalResponse(resp, &unmarshalled); err != nil {
		fmt.Printf("error unmarshalling response: %s\n", err)
		return
	}

	internal.RenderTable(unmarshalled, ovhcloudconnectColumnsToDisplay)
}

func getOvhCloudConnect(_ *cobra.Command, args []string) {
	var object map[string]any
	if err := client.Get(fmt.Sprintf("/ovhCloudConnect/%s", url.PathEscape(args[0])), &object); err != nil {
		fmt.Printf("error fetching OvhCloudConnect: %s\n", err)
		return
	}

	internal.RenderObject(object, ovhcloudconnectColumnsToDisplay[0])
}

func init() {
	ovhcloudconnectCmd := &cobra.Command{
		Use:   "ovhcloudconnect",
		Short: "Retrieve information and manage your OvhCloudConnect services",
	}

	// Command to list OvhCloudConnect services
	ovhcloudconnectCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List your OvhCloudConnect services",
		Run:   listOvhCloudConnect,
	})

	// Command to get a single OvhCloudConnect
	ovhcloudconnectCmd.AddCommand(&cobra.Command{
		Use:        "get",
		Short:      "Retrieve information of a specific OvhCloudConnect",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"service_name"},
		Run:        getOvhCloudConnect,
	})

	rootCmd.AddCommand(ovhcloudconnectCmd)
}
