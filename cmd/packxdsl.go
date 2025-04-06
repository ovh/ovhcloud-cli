
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
	packxdslColumnsToDisplay = []string{ "packName","description" }
)

func listPackXDSL(_ *cobra.Command, _ []string) {
	req, err := client.NewRequest(http.MethodGet, "/pack/xdsl", nil, true)
	if err != nil {
		log.Fatalf("error crafting request: %s\n", err)
	}

	req.Header.Set("X-Pagination-Mode", "CachedObjectList-Pages")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("error fetching /pack/xdsl: %s\n", err)
	}

	var unmarshalled []map[string]any
	if err := client.UnmarshalResponse(resp, &unmarshalled); err != nil {
		log.Fatalf("error unmarshalling response: %s\n", err)
	}

	internal.OutputTable(unmarshalled, packxdslColumnsToDisplay, jsonOutput, yamlOutput)
}

func getPackXDSL(_ *cobra.Command, args []string) {
	var object map[string]any
	if err := client.Get(fmt.Sprintf("/pack/xdsl/%s", url.PathEscape(args[0])), &object); err != nil {
		log.Fatalf("error fetching PackXDSL: %s\n", err)
	}

	internal.OutputObject(object, packxdslColumnsToDisplay[0], jsonOutput, yamlOutput)
}

func init() {
	packxdslCmd := &cobra.Command{
		Use:   "packxdsl",
		Short: "Retrieve information and manage your PackXDSL services",
	}

	// Command to list PackXDSL services
	packxdslCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List your PackXDSL services",
		Run:   listPackXDSL,
	})

	// Command to get a single PackXDSL
	packxdslCmd.AddCommand(&cobra.Command{
		Use:        "get",
		Short:      "Retrieve information of a specific PackXDSL",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"service_name"},
		Run:        getPackXDSL,
	})

	rootCmd.AddCommand(packxdslCmd)
}
