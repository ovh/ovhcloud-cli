
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
	sslColumnsToDisplay = []string{ "serviceName","type","authority","status" }
)

func listSsl(_ *cobra.Command, _ []string) {
	req, err := client.NewRequest(http.MethodGet, "/ssl", nil, true)
	if err != nil {
		log.Fatalf("error crafting request: %s\n", err)
	}

	req.Header.Set("X-Pagination-Mode", "CachedObjectList-Pages")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("error fetching /ssl: %s\n", err)
	}

	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("error reading response body: %s", err)
	}

	internal.OutputTable(bodyBytes, sslColumnsToDisplay, jsonOutput, yamlOutput)
}

func getSsl(_ *cobra.Command, args []string) {
	var object map[string]any
	if err := client.Get(fmt.Sprintf("/ssl/%s", url.PathEscape(args[0])), &object); err != nil {
		log.Fatalf("error fetching Ssl: %s\n", err)
	}

	internal.OutputObject(object, sslColumnsToDisplay[0], jsonOutput, yamlOutput)
}

func init() {
	sslCmd := &cobra.Command{
		Use:   "ssl",
		Short: "Retrieve information and manage your Ssl services",
	}

	// Command to list Ssl services
	sslCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List your Ssl services",
		Run:   listSsl,
	})

	// Command to get a single Ssl
	sslCmd.AddCommand(&cobra.Command{
		Use:        "get",
		Short:      "Retrieve information of a specific Ssl",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"service_name"},
		Run:        getSsl,
	})

	rootCmd.AddCommand(sslCmd)
}
