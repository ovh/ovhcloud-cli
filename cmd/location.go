
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
	locationColumnsToDisplay = []string{ "name","type","specificType","location" }
)

func listLocation(_ *cobra.Command, _ []string) {
	req, err := client.NewRequest(http.MethodGet, "/v2/location", nil, true)
	if err != nil {
		log.Fatalf("error crafting request: %s\n", err)
	}

	req.Header.Set("X-Pagination-Mode", "CachedObjectList-Pages")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("error fetching /v2/location: %s\n", err)
	}

	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("error reading response body: %s", err)
	}

	internal.OutputTable(bodyBytes, locationColumnsToDisplay, jsonOutput, yamlOutput)
}

func getLocation(_ *cobra.Command, args []string) {
	var object map[string]any
	if err := client.Get(fmt.Sprintf("/v2/location/%s", url.PathEscape(args[0])), &object); err != nil {
		log.Fatalf("error fetching Location: %s\n", err)
	}

	internal.OutputObject(object, locationColumnsToDisplay[0], jsonOutput, yamlOutput)
}

func init() {
	locationCmd := &cobra.Command{
		Use:   "location",
		Short: "Retrieve information and manage your Location services",
	}

	// Command to list Location services
	locationCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List your Location services",
		Run:   listLocation,
	})

	// Command to get a single Location
	locationCmd.AddCommand(&cobra.Command{
		Use:        "get",
		Short:      "Retrieve information of a specific Location",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"service_name"},
		Run:        getLocation,
	})

	rootCmd.AddCommand(locationCmd)
}
