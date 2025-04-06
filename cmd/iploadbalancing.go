
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
	iploadbalancingColumnsToDisplay = []string{ "serviceName","displayName","zone","state" }
)

func listIpLoadbalancing(_ *cobra.Command, _ []string) {
	req, err := client.NewRequest(http.MethodGet, "/ipLoadbalancing", nil, true)
	if err != nil {
		log.Fatalf("error crafting request: %s\n", err)
	}

	req.Header.Set("X-Pagination-Mode", "CachedObjectList-Pages")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("error fetching /ipLoadbalancing: %s\n", err)
	}

	var unmarshalled []map[string]any
	if err := client.UnmarshalResponse(resp, &unmarshalled); err != nil {
		log.Fatalf("error unmarshalling response: %s\n", err)
	}

	internal.OutputTable(unmarshalled, iploadbalancingColumnsToDisplay, jsonOutput, yamlOutput)
}

func getIpLoadbalancing(_ *cobra.Command, args []string) {
	var object map[string]any
	if err := client.Get(fmt.Sprintf("/ipLoadbalancing/%s", url.PathEscape(args[0])), &object); err != nil {
		log.Fatalf("error fetching IpLoadbalancing: %s\n", err)
	}

	internal.OutputObject(object, iploadbalancingColumnsToDisplay[0], jsonOutput, yamlOutput)
}

func init() {
	iploadbalancingCmd := &cobra.Command{
		Use:   "iploadbalancing",
		Short: "Retrieve information and manage your IpLoadbalancing services",
	}

	// Command to list IpLoadbalancing services
	iploadbalancingCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List your IpLoadbalancing services",
		Run:   listIpLoadbalancing,
	})

	// Command to get a single IpLoadbalancing
	iploadbalancingCmd.AddCommand(&cobra.Command{
		Use:        "get",
		Short:      "Retrieve information of a specific IpLoadbalancing",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"service_name"},
		Run:        getIpLoadbalancing,
	})

	rootCmd.AddCommand(iploadbalancingCmd)
}
