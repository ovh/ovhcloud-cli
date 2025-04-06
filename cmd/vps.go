package cmd

import (
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal"
)

var (
	vpsColumnsToDisplay = []string{"name", "displayName", "state", "zone"}
)

func listVPS(_ *cobra.Command, _ []string) {
	req, err := client.NewRequest(http.MethodGet, "/vps", nil, true)
	if err != nil {
		log.Fatalf("error crafting request: %s\n", err)
	}

	req.Header.Set("X-Pagination-Mode", "CachedObjectList-Pages")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("error fetching /vps: %s\n", err)
	}

	var unmarshalled []map[string]any
	if err := client.UnmarshalResponse(resp, &unmarshalled); err != nil {
		log.Fatalf("error unmarshalling response: %s\n", err)
	}

	internal.OutputTable(unmarshalled, vpsColumnsToDisplay, jsonOutput, yamlOutput)
}

func getVPS(_ *cobra.Command, args []string) {
	var object map[string]any
	if err := client.Get(fmt.Sprintf("/vps/%s", url.PathEscape(args[0])), &object); err != nil {
		log.Fatalf("error fetching vps: %s\n", err)
	}

	internal.OutputObject(object, vpsColumnsToDisplay[0], jsonOutput, yamlOutput)
}

func init() {
	vpsCmd := &cobra.Command{
		Use:   "vps",
		Short: "Retrieve information and manage your VPS services",
	}

	// Command to list VPS services
	vpsCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List your VPS services",
		Run:   listVPS,
	})

	// Command to get a single VPS
	vpsCmd.AddCommand(&cobra.Command{
		Use:        "get",
		Short:      "Retrieve information of a specific VPS",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"service_name"},
		Run:        getVPS,
	})

	rootCmd.AddCommand(vpsCmd)
}
