
package cmd

import (
	"fmt"
	"log"
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
		log.Fatalf("error crafting request: %s\n", err)
	}

	req.Header.Set("X-Pagination-Mode", "CachedObjectList-Pages")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("error fetching /nutanix: %s\n", err)
	}

	var unmarshalled []map[string]any
	if err := client.UnmarshalResponse(resp, &unmarshalled); err != nil {
		log.Fatalf("error unmarshalling response: %s\n", err)
	}

	internal.OutputTable(unmarshalled, nutanixColumnsToDisplay, jsonOutput, yamlOutput)
}

func getNutanix(_ *cobra.Command, args []string) {
	var object map[string]any
	if err := client.Get(fmt.Sprintf("/nutanix/%s", url.PathEscape(args[0])), &object); err != nil {
		log.Fatalf("error fetching Nutanix: %s\n", err)
	}

	internal.OutputObject(object, nutanixColumnsToDisplay[0], jsonOutput, yamlOutput)
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
