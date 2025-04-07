
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
	vpsColumnsToDisplay = []string{ "name","displayName","state","zone" }
)

func listVps(_ *cobra.Command, _ []string) {
	req, err := client.NewRequest(http.MethodGet, "/vps", nil, true)
	if err != nil {
		log.Fatalf("error crafting request: %s\n", err)
	}

	req.Header.Set("X-Pagination-Mode", "CachedObjectList-Pages")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("error fetching /vps: %s\n", err)
	}

	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("error reading response body: %s", err)
	}

	internal.OutputTable(bodyBytes, vpsColumnsToDisplay, jsonOutput, yamlOutput)
}

func getVps(_ *cobra.Command, args []string) {
	var object map[string]any
	if err := client.Get(fmt.Sprintf("/vps/%s", url.PathEscape(args[0])), &object); err != nil {
		log.Fatalf("error fetching Vps: %s\n", err)
	}

	internal.OutputObject(object, vpsColumnsToDisplay[0], jsonOutput, yamlOutput)
}

func init() {
	vpsCmd := &cobra.Command{
		Use:   "vps",
		Short: "Retrieve information and manage your Vps services",
	}

	// Command to list Vps services
	vpsCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List your Vps services",
		Run:   listVps,
	})

	// Command to get a single Vps
	vpsCmd.AddCommand(&cobra.Command{
		Use:        "get",
		Short:      "Retrieve information of a specific Vps",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"service_name"},
		Run:        getVps,
	})

	rootCmd.AddCommand(vpsCmd)
}
