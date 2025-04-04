package cmd

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal"
)

var (
	vpsIDKey            = "name"
	vpsColumnsToDisplay = []string{vpsIDKey, "displayName", "state", "zone"}
)

func listVPS(_ *cobra.Command, _ []string) {
	req, err := client.NewRequest(http.MethodGet, "/vps", nil, true)
	if err != nil {
		fmt.Printf("error crafting request: %s\n", err)
		return
	}

	req.Header.Set("X-Pagination-Mode", "CachedObjectList-Pages")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("error fetching /vps: %s\n", err)
		return
	}

	var unmarshalled []map[string]any
	if err := client.UnmarshalResponse(resp, &unmarshalled); err != nil {
		fmt.Printf("error unmarshalling response: %s\n", err)
		return
	}

	internal.RenderTable(unmarshalled, vpsColumnsToDisplay)
}

func getVPS(_ *cobra.Command, args []string) {
	var object map[string]any
	if err := client.Get(fmt.Sprintf("/vps/%s", url.PathEscape(args[0])), &object); err != nil {
		fmt.Printf("error fetching vps: %s\n", err)
		return
	}

	internal.RenderObject(object, vpsIDKey)
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
