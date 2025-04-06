
package cmd

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal"
)

var (
	storagenetappColumnsToDisplay = []string{ "id","name","region","status" }
)

func listStorageNetApp(_ *cobra.Command, _ []string) {
	req, err := client.NewRequest(http.MethodGet, "/storage/netapp", nil, true)
	if err != nil {
		fmt.Printf("error crafting request: %s\n", err)
		return
	}

	req.Header.Set("X-Pagination-Mode", "CachedObjectList-Pages")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("error fetching /storage/netapp: %s\n", err)
		return
	}

	var unmarshalled []map[string]any
	if err := client.UnmarshalResponse(resp, &unmarshalled); err != nil {
		fmt.Printf("error unmarshalling response: %s\n", err)
		return
	}

	internal.RenderTable(unmarshalled, storagenetappColumnsToDisplay)
}

func getStorageNetApp(_ *cobra.Command, args []string) {
	var object map[string]any
	if err := client.Get(fmt.Sprintf("/storage/netapp/%s", url.PathEscape(args[0])), &object); err != nil {
		fmt.Printf("error fetching StorageNetApp: %s\n", err)
		return
	}

	internal.RenderObject(object, storagenetappColumnsToDisplay[0])
}

func init() {
	storagenetappCmd := &cobra.Command{
		Use:   "storagenetapp",
		Short: "Retrieve information and manage your StorageNetApp services",
	}

	// Command to list StorageNetApp services
	storagenetappCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List your StorageNetApp services",
		Run:   listStorageNetApp,
	})

	// Command to get a single StorageNetApp
	storagenetappCmd.AddCommand(&cobra.Command{
		Use:        "get",
		Short:      "Retrieve information of a specific StorageNetApp",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"service_name"},
		Run:        getStorageNetApp,
	})

	rootCmd.AddCommand(storagenetappCmd)
}
