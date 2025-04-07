
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
	ipColumnsToDisplay = []string{ "ip","rir","routedTo","country","description" }
)

func listIp(_ *cobra.Command, _ []string) {
	req, err := client.NewRequest(http.MethodGet, "/ip", nil, true)
	if err != nil {
		log.Fatalf("error crafting request: %s\n", err)
	}

	req.Header.Set("X-Pagination-Mode", "CachedObjectList-Pages")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("error fetching /ip: %s\n", err)
	}

	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("error reading response body: %s", err)
	}

	internal.OutputTable(bodyBytes, ipColumnsToDisplay, jsonOutput, yamlOutput)
}

func getIp(_ *cobra.Command, args []string) {
	var object map[string]any
	if err := client.Get(fmt.Sprintf("/ip/%s", url.PathEscape(args[0])), &object); err != nil {
		log.Fatalf("error fetching Ip: %s\n", err)
	}

	internal.OutputObject(object, ipColumnsToDisplay[0], jsonOutput, yamlOutput)
}

func init() {
	ipCmd := &cobra.Command{
		Use:   "ip",
		Short: "Retrieve information and manage your Ip services",
	}

	// Command to list Ip services
	ipCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List your Ip services",
		Run:   listIp,
	})

	// Command to get a single Ip
	ipCmd.AddCommand(&cobra.Command{
		Use:        "get",
		Short:      "Retrieve information of a specific Ip",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"service_name"},
		Run:        getIp,
	})

	rootCmd.AddCommand(ipCmd)
}
