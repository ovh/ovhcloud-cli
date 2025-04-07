
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
	overtheboxColumnsToDisplay = []string{ "serviceName","offer","status","bandwidth" }
)

func listOverTheBox(_ *cobra.Command, _ []string) {
	req, err := client.NewRequest(http.MethodGet, "/overTheBox", nil, true)
	if err != nil {
		log.Fatalf("error crafting request: %s\n", err)
	}

	req.Header.Set("X-Pagination-Mode", "CachedObjectList-Pages")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("error fetching /overTheBox: %s\n", err)
	}

	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("error reading response body: %s", err)
	}

	internal.OutputTable(bodyBytes, overtheboxColumnsToDisplay, jsonOutput, yamlOutput)
}

func getOverTheBox(_ *cobra.Command, args []string) {
	var object map[string]any
	if err := client.Get(fmt.Sprintf("/overTheBox/%s", url.PathEscape(args[0])), &object); err != nil {
		log.Fatalf("error fetching OverTheBox: %s\n", err)
	}

	internal.OutputObject(object, overtheboxColumnsToDisplay[0], jsonOutput, yamlOutput)
}

func init() {
	overtheboxCmd := &cobra.Command{
		Use:   "overthebox",
		Short: "Retrieve information and manage your OverTheBox services",
	}

	// Command to list OverTheBox services
	overtheboxCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List your OverTheBox services",
		Run:   listOverTheBox,
	})

	// Command to get a single OverTheBox
	overtheboxCmd.AddCommand(&cobra.Command{
		Use:        "get",
		Short:      "Retrieve information of a specific OverTheBox",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"service_name"},
		Run:        getOverTheBox,
	})

	rootCmd.AddCommand(overtheboxCmd)
}
