
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
	dedicatedclusterColumnsToDisplay = []string{ "id","region","model","status" }
)

func listDedicatedCluster(_ *cobra.Command, _ []string) {
	req, err := client.NewRequest(http.MethodGet, "/dedicated/cluster", nil, true)
	if err != nil {
		log.Fatalf("error crafting request: %s\n", err)
	}

	req.Header.Set("X-Pagination-Mode", "CachedObjectList-Pages")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("error fetching /dedicated/cluster: %s\n", err)
	}

	var unmarshalled []map[string]any
	if err := client.UnmarshalResponse(resp, &unmarshalled); err != nil {
		log.Fatalf("error unmarshalling response: %s\n", err)
	}

	internal.OutputTable(unmarshalled, dedicatedclusterColumnsToDisplay, jsonOutput, yamlOutput)
}

func getDedicatedCluster(_ *cobra.Command, args []string) {
	var object map[string]any
	if err := client.Get(fmt.Sprintf("/dedicated/cluster/%s", url.PathEscape(args[0])), &object); err != nil {
		log.Fatalf("error fetching DedicatedCluster: %s\n", err)
	}

	internal.OutputObject(object, dedicatedclusterColumnsToDisplay[0], jsonOutput, yamlOutput)
}

func init() {
	dedicatedclusterCmd := &cobra.Command{
		Use:   "dedicatedcluster",
		Short: "Retrieve information and manage your DedicatedCluster services",
	}

	// Command to list DedicatedCluster services
	dedicatedclusterCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List your DedicatedCluster services",
		Run:   listDedicatedCluster,
	})

	// Command to get a single DedicatedCluster
	dedicatedclusterCmd.AddCommand(&cobra.Command{
		Use:        "get",
		Short:      "Retrieve information of a specific DedicatedCluster",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"service_name"},
		Run:        getDedicatedCluster,
	})

	rootCmd.AddCommand(dedicatedclusterCmd)
}
