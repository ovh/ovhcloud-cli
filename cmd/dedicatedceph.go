
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
	dedicatedcephColumnsToDisplay = []string{ "serviceName","region","state","status" }
)

func listDedicatedCeph(_ *cobra.Command, _ []string) {
	req, err := client.NewRequest(http.MethodGet, "/dedicated/ceph", nil, true)
	if err != nil {
		log.Fatalf("error crafting request: %s\n", err)
	}

	req.Header.Set("X-Pagination-Mode", "CachedObjectList-Pages")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("error fetching /dedicated/ceph: %s\n", err)
	}

	var unmarshalled []map[string]any
	if err := client.UnmarshalResponse(resp, &unmarshalled); err != nil {
		log.Fatalf("error unmarshalling response: %s\n", err)
	}

	internal.OutputTable(unmarshalled, dedicatedcephColumnsToDisplay, jsonOutput, yamlOutput)
}

func getDedicatedCeph(_ *cobra.Command, args []string) {
	var object map[string]any
	if err := client.Get(fmt.Sprintf("/dedicated/ceph/%s", url.PathEscape(args[0])), &object); err != nil {
		log.Fatalf("error fetching DedicatedCeph: %s\n", err)
	}

	internal.OutputObject(object, dedicatedcephColumnsToDisplay[0], jsonOutput, yamlOutput)
}

func init() {
	dedicatedcephCmd := &cobra.Command{
		Use:   "dedicatedceph",
		Short: "Retrieve information and manage your DedicatedCeph services",
	}

	// Command to list DedicatedCeph services
	dedicatedcephCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List your DedicatedCeph services",
		Run:   listDedicatedCeph,
	})

	// Command to get a single DedicatedCeph
	dedicatedcephCmd.AddCommand(&cobra.Command{
		Use:        "get",
		Short:      "Retrieve information of a specific DedicatedCeph",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"service_name"},
		Run:        getDedicatedCeph,
	})

	rootCmd.AddCommand(dedicatedcephCmd)
}
