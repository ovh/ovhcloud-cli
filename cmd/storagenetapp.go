
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
	storagenetappColumnsToDisplay = []string{ "id","name","region","status" }
)

func listStorageNetApp(_ *cobra.Command, _ []string) {
	req, err := client.NewRequest(http.MethodGet, "/storage/netapp", nil, true)
	if err != nil {
		log.Fatalf("error crafting request: %s\n", err)
	}

	req.Header.Set("X-Pagination-Mode", "CachedObjectList-Pages")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("error fetching /storage/netapp: %s\n", err)
	}

	var unmarshalled []map[string]any
	if err := client.UnmarshalResponse(resp, &unmarshalled); err != nil {
		log.Fatalf("error unmarshalling response: %s\n", err)
	}

	internal.OutputTable(unmarshalled, storagenetappColumnsToDisplay, jsonOutput, yamlOutput)
}

func getStorageNetApp(_ *cobra.Command, args []string) {
	var object map[string]any
	if err := client.Get(fmt.Sprintf("/storage/netapp/%s", url.PathEscape(args[0])), &object); err != nil {
		log.Fatalf("error fetching StorageNetApp: %s\n", err)
	}

	internal.OutputObject(object, storagenetappColumnsToDisplay[0], jsonOutput, yamlOutput)
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
