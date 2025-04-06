
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
	ldpColumnsToDisplay = []string{ "serviceName","displayName","isClusterOwner","state","username" }
)

func listLdp(_ *cobra.Command, _ []string) {
	req, err := client.NewRequest(http.MethodGet, "/dbaas/logs", nil, true)
	if err != nil {
		log.Fatalf("error crafting request: %s\n", err)
	}

	req.Header.Set("X-Pagination-Mode", "CachedObjectList-Pages")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("error fetching /dbaas/logs: %s\n", err)
	}

	var unmarshalled []map[string]any
	if err := client.UnmarshalResponse(resp, &unmarshalled); err != nil {
		log.Fatalf("error unmarshalling response: %s\n", err)
	}

	internal.OutputTable(unmarshalled, ldpColumnsToDisplay, jsonOutput, yamlOutput)
}

func getLdp(_ *cobra.Command, args []string) {
	var object map[string]any
	if err := client.Get(fmt.Sprintf("/dbaas/logs/%s", url.PathEscape(args[0])), &object); err != nil {
		log.Fatalf("error fetching Ldp: %s\n", err)
	}

	internal.OutputObject(object, ldpColumnsToDisplay[0], jsonOutput, yamlOutput)
}

func init() {
	ldpCmd := &cobra.Command{
		Use:   "ldp",
		Short: "Retrieve information and manage your Ldp services",
	}

	// Command to list Ldp services
	ldpCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List your Ldp services",
		Run:   listLdp,
	})

	// Command to get a single Ldp
	ldpCmd.AddCommand(&cobra.Command{
		Use:        "get",
		Short:      "Retrieve information of a specific Ldp",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"service_name"},
		Run:        getLdp,
	})

	rootCmd.AddCommand(ldpCmd)
}
