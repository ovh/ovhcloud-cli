
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
	xdslColumnsToDisplay = []string{ "accessName","accessType","provider","role","status" }
)

func listXdsl(_ *cobra.Command, _ []string) {
	req, err := client.NewRequest(http.MethodGet, "/xdsl", nil, true)
	if err != nil {
		log.Fatalf("error crafting request: %s\n", err)
	}

	req.Header.Set("X-Pagination-Mode", "CachedObjectList-Pages")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("error fetching /xdsl: %s\n", err)
	}

	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("error reading response body: %s", err)
	}

	internal.OutputTable(bodyBytes, xdslColumnsToDisplay, jsonOutput, yamlOutput)
}

func getXdsl(_ *cobra.Command, args []string) {
	var object map[string]any
	if err := client.Get(fmt.Sprintf("/xdsl/%s", url.PathEscape(args[0])), &object); err != nil {
		log.Fatalf("error fetching Xdsl: %s\n", err)
	}

	internal.OutputObject(object, xdslColumnsToDisplay[0], jsonOutput, yamlOutput)
}

func init() {
	xdslCmd := &cobra.Command{
		Use:   "xdsl",
		Short: "Retrieve information and manage your Xdsl services",
	}

	// Command to list Xdsl services
	xdslCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List your Xdsl services",
		Run:   listXdsl,
	})

	// Command to get a single Xdsl
	xdslCmd.AddCommand(&cobra.Command{
		Use:        "get",
		Short:      "Retrieve information of a specific Xdsl",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"service_name"},
		Run:        getXdsl,
	})

	rootCmd.AddCommand(xdslCmd)
}
