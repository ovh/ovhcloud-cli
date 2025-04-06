
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
	dedicatednashaColumnsToDisplay = []string{ "serviceName","customName","datacenter" }
)

func listDedicatedNasHA(_ *cobra.Command, _ []string) {
	req, err := client.NewRequest(http.MethodGet, "/dedicated/nasha", nil, true)
	if err != nil {
		log.Fatalf("error crafting request: %s\n", err)
	}

	req.Header.Set("X-Pagination-Mode", "CachedObjectList-Pages")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("error fetching /dedicated/nasha: %s\n", err)
	}

	var unmarshalled []map[string]any
	if err := client.UnmarshalResponse(resp, &unmarshalled); err != nil {
		log.Fatalf("error unmarshalling response: %s\n", err)
	}

	internal.OutputTable(unmarshalled, dedicatednashaColumnsToDisplay, jsonOutput, yamlOutput)
}

func getDedicatedNasHA(_ *cobra.Command, args []string) {
	var object map[string]any
	if err := client.Get(fmt.Sprintf("/dedicated/nasha/%s", url.PathEscape(args[0])), &object); err != nil {
		log.Fatalf("error fetching DedicatedNasHA: %s\n", err)
	}

	internal.OutputObject(object, dedicatednashaColumnsToDisplay[0], jsonOutput, yamlOutput)
}

func init() {
	dedicatednashaCmd := &cobra.Command{
		Use:   "dedicatednasha",
		Short: "Retrieve information and manage your DedicatedNasHA services",
	}

	// Command to list DedicatedNasHA services
	dedicatednashaCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List your DedicatedNasHA services",
		Run:   listDedicatedNasHA,
	})

	// Command to get a single DedicatedNasHA
	dedicatednashaCmd.AddCommand(&cobra.Command{
		Use:        "get",
		Short:      "Retrieve information of a specific DedicatedNasHA",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"service_name"},
		Run:        getDedicatedNasHA,
	})

	rootCmd.AddCommand(dedicatednashaCmd)
}
