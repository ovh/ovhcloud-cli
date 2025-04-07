
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
	hostingprivatedatabaseColumnsToDisplay = []string{ "serviceName","displayName","type","version","state" }
)

func listHostingPrivateDatabase(_ *cobra.Command, _ []string) {
	req, err := client.NewRequest(http.MethodGet, "/hosting/privateDatabase", nil, true)
	if err != nil {
		log.Fatalf("error crafting request: %s\n", err)
	}

	req.Header.Set("X-Pagination-Mode", "CachedObjectList-Pages")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("error fetching /hosting/privateDatabase: %s\n", err)
	}

	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("error reading response body: %s", err)
	}

	internal.OutputTable(bodyBytes, hostingprivatedatabaseColumnsToDisplay, jsonOutput, yamlOutput)
}

func getHostingPrivateDatabase(_ *cobra.Command, args []string) {
	var object map[string]any
	if err := client.Get(fmt.Sprintf("/hosting/privateDatabase/%s", url.PathEscape(args[0])), &object); err != nil {
		log.Fatalf("error fetching HostingPrivateDatabase: %s\n", err)
	}

	internal.OutputObject(object, hostingprivatedatabaseColumnsToDisplay[0], jsonOutput, yamlOutput)
}

func init() {
	hostingprivatedatabaseCmd := &cobra.Command{
		Use:   "hostingprivatedatabase",
		Short: "Retrieve information and manage your HostingPrivateDatabase services",
	}

	// Command to list HostingPrivateDatabase services
	hostingprivatedatabaseCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List your HostingPrivateDatabase services",
		Run:   listHostingPrivateDatabase,
	})

	// Command to get a single HostingPrivateDatabase
	hostingprivatedatabaseCmd.AddCommand(&cobra.Command{
		Use:        "get",
		Short:      "Retrieve information of a specific HostingPrivateDatabase",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"service_name"},
		Run:        getHostingPrivateDatabase,
	})

	rootCmd.AddCommand(hostingprivatedatabaseCmd)
}
