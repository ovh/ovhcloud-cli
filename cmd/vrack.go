
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
	vrackColumnsToDisplay = []string{ "name","description" }
)

func listVrack(_ *cobra.Command, _ []string) {
	req, err := client.NewRequest(http.MethodGet, "/vrack", nil, true)
	if err != nil {
		log.Fatalf("error crafting request: %s\n", err)
	}

	req.Header.Set("X-Pagination-Mode", "CachedObjectList-Pages")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("error fetching /vrack: %s\n", err)
	}

	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("error reading response body: %s", err)
	}

	internal.OutputTable(bodyBytes, vrackColumnsToDisplay, jsonOutput, yamlOutput)
}

func getVrack(_ *cobra.Command, args []string) {
	var object map[string]any
	if err := client.Get(fmt.Sprintf("/vrack/%s", url.PathEscape(args[0])), &object); err != nil {
		log.Fatalf("error fetching Vrack: %s\n", err)
	}

	internal.OutputObject(object, vrackColumnsToDisplay[0], jsonOutput, yamlOutput)
}

func init() {
	vrackCmd := &cobra.Command{
		Use:   "vrack",
		Short: "Retrieve information and manage your Vrack services",
	}

	// Command to list Vrack services
	vrackCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List your Vrack services",
		Run:   listVrack,
	})

	// Command to get a single Vrack
	vrackCmd.AddCommand(&cobra.Command{
		Use:        "get",
		Short:      "Retrieve information of a specific Vrack",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"service_name"},
		Run:        getVrack,
	})

	rootCmd.AddCommand(vrackCmd)
}
