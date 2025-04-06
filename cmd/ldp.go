
package cmd

import (
	"fmt"
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
		fmt.Printf("error crafting request: %s\n", err)
		return
	}

	req.Header.Set("X-Pagination-Mode", "CachedObjectList-Pages")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("error fetching /dbaas/logs: %s\n", err)
		return
	}

	var unmarshalled []map[string]any
	if err := client.UnmarshalResponse(resp, &unmarshalled); err != nil {
		fmt.Printf("error unmarshalling response: %s\n", err)
		return
	}

	internal.RenderTable(unmarshalled, ldpColumnsToDisplay)
}

func getLdp(_ *cobra.Command, args []string) {
	var object map[string]any
	if err := client.Get(fmt.Sprintf("/dbaas/logs/%s", url.PathEscape(args[0])), &object); err != nil {
		fmt.Printf("error fetching Ldp: %s\n", err)
		return
	}

	internal.RenderObject(object, ldpColumnsToDisplay[0])
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
