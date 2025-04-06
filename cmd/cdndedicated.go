
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
	cdndedicatedColumnsToDisplay = []string{ "service","offer","anycast" }
)

func listCdnDedicated(_ *cobra.Command, _ []string) {
	req, err := client.NewRequest(http.MethodGet, "/cdn/dedicated", nil, true)
	if err != nil {
		log.Fatalf("error crafting request: %s\n", err)
	}

	req.Header.Set("X-Pagination-Mode", "CachedObjectList-Pages")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("error fetching /cdn/dedicated: %s\n", err)
	}

	var unmarshalled []map[string]any
	if err := client.UnmarshalResponse(resp, &unmarshalled); err != nil {
		log.Fatalf("error unmarshalling response: %s\n", err)
	}

	internal.OutputTable(unmarshalled, cdndedicatedColumnsToDisplay, jsonOutput, yamlOutput)
}

func getCdnDedicated(_ *cobra.Command, args []string) {
	var object map[string]any
	if err := client.Get(fmt.Sprintf("/cdn/dedicated/%s", url.PathEscape(args[0])), &object); err != nil {
		log.Fatalf("error fetching CdnDedicated: %s\n", err)
	}

	internal.OutputObject(object, cdndedicatedColumnsToDisplay[0], jsonOutput, yamlOutput)
}

func init() {
	cdndedicatedCmd := &cobra.Command{
		Use:   "cdndedicated",
		Short: "Retrieve information and manage your CdnDedicated services",
	}

	// Command to list CdnDedicated services
	cdndedicatedCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List your CdnDedicated services",
		Run:   listCdnDedicated,
	})

	// Command to get a single CdnDedicated
	cdndedicatedCmd.AddCommand(&cobra.Command{
		Use:        "get",
		Short:      "Retrieve information of a specific CdnDedicated",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"service_name"},
		Run:        getCdnDedicated,
	})

	rootCmd.AddCommand(cdndedicatedCmd)
}
