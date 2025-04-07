
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
	baremetalColumnsToDisplay = []string{ "serverId","name","region","os" }
)

func listBaremetal(_ *cobra.Command, _ []string) {
	req, err := client.NewRequest(http.MethodGet, "/dedicated/server", nil, true)
	if err != nil {
		log.Fatalf("error crafting request: %s\n", err)
	}

	req.Header.Set("X-Pagination-Mode", "CachedObjectList-Pages")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("error fetching /dedicated/server: %s\n", err)
	}

	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("error reading response body: %s", err)
	}

	internal.OutputTable(bodyBytes, baremetalColumnsToDisplay, jsonOutput, yamlOutput)
}

func getBaremetal(_ *cobra.Command, args []string) {
	var object map[string]any
	if err := client.Get(fmt.Sprintf("/dedicated/server/%s", url.PathEscape(args[0])), &object); err != nil {
		log.Fatalf("error fetching Baremetal: %s\n", err)
	}

	internal.OutputObject(object, baremetalColumnsToDisplay[0], jsonOutput, yamlOutput)
}

func init() {
	baremetalCmd := &cobra.Command{
		Use:   "baremetal",
		Short: "Retrieve information and manage your Baremetal services",
	}

	// Command to list Baremetal services
	baremetalCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List your Baremetal services",
		Run:   listBaremetal,
	})

	// Command to get a single Baremetal
	baremetalCmd.AddCommand(&cobra.Command{
		Use:        "get",
		Short:      "Retrieve information of a specific Baremetal",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"service_name"},
		Run:        getBaremetal,
	})

	rootCmd.AddCommand(baremetalCmd)
}
