
package cmd

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal"
)

var (
	alldomColumnsToDisplay = []string{ "name","type","offer" }
)

func listAllDom(_ *cobra.Command, _ []string) {
	req, err := client.NewRequest(http.MethodGet, "/allDom", nil, true)
	if err != nil {
		log.Fatalf("error crafting request: %s\n", err)
	}

	req.Header.Set("X-Pagination-Mode", "CachedObjectList-Pages")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("error fetching /allDom: %s\n", err)
	}

	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("error reading response body: %s", err)
	}

	internal.OutputTable(bodyBytes, alldomColumnsToDisplay, jsonOutput, yamlOutput)
}

func getAllDom(_ *cobra.Command, args []string) {
	var object map[string]any
	if err := client.Get(fmt.Sprintf("/allDom/%s", url.PathEscape(args[0])), &object); err != nil {
		log.Fatalf("error fetching AllDom: %s\n", err)
	}

	internal.OutputObject(object, alldomColumnsToDisplay[0], jsonOutput, yamlOutput)
}

func init() {
	alldomCmd := &cobra.Command{
		Use:   "alldom",
		Short: "Retrieve information and manage your AllDom services",
	}

	// Command to list AllDom services
	alldomCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List your AllDom services",
		Run:   listAllDom,
	})

	// Command to get a single AllDom
	alldomCmd.AddCommand(&cobra.Command{
		Use:        "get",
		Short:      "Retrieve information of a specific AllDom",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"service_name"},
		Run:        getAllDom,
	})

	rootCmd.AddCommand(alldomCmd)
}
