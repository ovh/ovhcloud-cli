
package cmd

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal"
)

var (
	nutanixColumnsToDisplay = []string{ "serviceName","status" }
)

func listNutanix(_ *cobra.Command, _ []string) {
	req, err := client.NewRequest(http.MethodGet, "/nutanix", nil, true)
	if err != nil {
		fmt.Printf("error crafting request: %s\n", err)
		return
	}

	req.Header.Set("X-Pagination-Mode", "CachedObjectList-Pages")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("error fetching /nutanix: %s\n", err)
		return
	}

	var unmarshalled []map[string]any
	if err := client.UnmarshalResponse(resp, &unmarshalled); err != nil {
		fmt.Printf("error unmarshalling response: %s\n", err)
		return
	}

	internal.RenderTable(unmarshalled, nutanixColumnsToDisplay)
}

func getNutanix(_ *cobra.Command, args []string) {
	var object map[string]any
	if err := client.Get(fmt.Sprintf("/nutanix/%s", url.PathEscape(args[0])), &object); err != nil {
		fmt.Printf("error fetching Nutanix: %s\n", err)
		return
	}

	internal.RenderObject(object, nutanixColumnsToDisplay[0])
}

func init() {
	nutanixCmd := &cobra.Command{
		Use:   "nutanix",
		Short: "Retrieve information and manage your Nutanix services",
	}

	// Command to list Nutanix services
	nutanixCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List your Nutanix services",
		Run:   listNutanix,
	})

	// Command to get a single Nutanix
	nutanixCmd.AddCommand(&cobra.Command{
		Use:        "get",
		Short:      "Retrieve information of a specific Nutanix",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"service_name"},
		Run:        getNutanix,
	})

	rootCmd.AddCommand(nutanixCmd)
}
