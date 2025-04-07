
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
	ovhcloudconnectColumnsToDisplay = []string{ "uuid","provider","status","description" }
)

func listOvhCloudConnect(_ *cobra.Command, _ []string) {
	req, err := client.NewRequest(http.MethodGet, "/ovhCloudConnect", nil, true)
	if err != nil {
		log.Fatalf("error crafting request: %s\n", err)
	}

	req.Header.Set("X-Pagination-Mode", "CachedObjectList-Pages")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("error fetching /ovhCloudConnect: %s\n", err)
	}

	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("error reading response body: %s", err)
	}

	internal.OutputTable(bodyBytes, ovhcloudconnectColumnsToDisplay, jsonOutput, yamlOutput)
}

func getOvhCloudConnect(_ *cobra.Command, args []string) {
	var object map[string]any
	if err := client.Get(fmt.Sprintf("/ovhCloudConnect/%s", url.PathEscape(args[0])), &object); err != nil {
		log.Fatalf("error fetching OvhCloudConnect: %s\n", err)
	}

	internal.OutputObject(object, ovhcloudconnectColumnsToDisplay[0], jsonOutput, yamlOutput)
}

func init() {
	ovhcloudconnectCmd := &cobra.Command{
		Use:   "ovhcloudconnect",
		Short: "Retrieve information and manage your OvhCloudConnect services",
	}

	// Command to list OvhCloudConnect services
	ovhcloudconnectCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List your OvhCloudConnect services",
		Run:   listOvhCloudConnect,
	})

	// Command to get a single OvhCloudConnect
	ovhcloudconnectCmd.AddCommand(&cobra.Command{
		Use:        "get",
		Short:      "Retrieve information of a specific OvhCloudConnect",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"service_name"},
		Run:        getOvhCloudConnect,
	})

	rootCmd.AddCommand(ovhcloudconnectCmd)
}
