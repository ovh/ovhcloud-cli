
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
	webhostingColumnsToDisplay = []string{ "serviceName","displayName","datacenter","state" }
)

func listWebHosting(_ *cobra.Command, _ []string) {
	req, err := client.NewRequest(http.MethodGet, "/hosting/web", nil, true)
	if err != nil {
		log.Fatalf("error crafting request: %s\n", err)
	}

	req.Header.Set("X-Pagination-Mode", "CachedObjectList-Pages")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("error fetching /hosting/web: %s\n", err)
	}

	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("error reading response body: %s", err)
	}

	internal.OutputTable(bodyBytes, webhostingColumnsToDisplay, jsonOutput, yamlOutput)
}

func getWebHosting(_ *cobra.Command, args []string) {
	var object map[string]any
	if err := client.Get(fmt.Sprintf("/hosting/web/%s", url.PathEscape(args[0])), &object); err != nil {
		log.Fatalf("error fetching WebHosting: %s\n", err)
	}

	internal.OutputObject(object, webhostingColumnsToDisplay[0], jsonOutput, yamlOutput)
}

func init() {
	webhostingCmd := &cobra.Command{
		Use:   "webhosting",
		Short: "Retrieve information and manage your WebHosting services",
	}

	// Command to list WebHosting services
	webhostingCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List your WebHosting services",
		Run:   listWebHosting,
	})

	// Command to get a single WebHosting
	webhostingCmd.AddCommand(&cobra.Command{
		Use:        "get",
		Short:      "Retrieve information of a specific WebHosting",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"service_name"},
		Run:        getWebHosting,
	})

	rootCmd.AddCommand(webhostingCmd)
}
