
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
	okmsColumnsToDisplay = []string{ "id","region" }
)

func listOkms(_ *cobra.Command, _ []string) {
	req, err := client.NewRequest(http.MethodGet, "/v2/okms/resource", nil, true)
	if err != nil {
		log.Fatalf("error crafting request: %s\n", err)
	}

	req.Header.Set("X-Pagination-Mode", "CachedObjectList-Pages")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("error fetching /v2/okms/resource: %s\n", err)
	}

	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("error reading response body: %s", err)
	}

	internal.OutputTable(bodyBytes, okmsColumnsToDisplay, jsonOutput, yamlOutput)
}

func getOkms(_ *cobra.Command, args []string) {
	var object map[string]any
	if err := client.Get(fmt.Sprintf("/v2/okms/resource/%s", url.PathEscape(args[0])), &object); err != nil {
		log.Fatalf("error fetching Okms: %s\n", err)
	}

	internal.OutputObject(object, okmsColumnsToDisplay[0], jsonOutput, yamlOutput)
}

func init() {
	okmsCmd := &cobra.Command{
		Use:   "okms",
		Short: "Retrieve information and manage your Okms services",
	}

	// Command to list Okms services
	okmsCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List your Okms services",
		Run:   listOkms,
	})

	// Command to get a single Okms
	okmsCmd.AddCommand(&cobra.Command{
		Use:        "get",
		Short:      "Retrieve information of a specific Okms",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"service_name"},
		Run:        getOkms,
	})

	rootCmd.AddCommand(okmsCmd)
}
