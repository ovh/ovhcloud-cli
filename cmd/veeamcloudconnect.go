
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
	veeamcloudconnectColumnsToDisplay = []string{ "serviceName","productOffer","location","vmCount" }
)

func listVeeamCloudConnect(_ *cobra.Command, _ []string) {
	req, err := client.NewRequest(http.MethodGet, "/veeamCloudConnect", nil, true)
	if err != nil {
		log.Fatalf("error crafting request: %s\n", err)
	}

	req.Header.Set("X-Pagination-Mode", "CachedObjectList-Pages")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("error fetching /veeamCloudConnect: %s\n", err)
	}

	var unmarshalled []map[string]any
	if err := client.UnmarshalResponse(resp, &unmarshalled); err != nil {
		log.Fatalf("error unmarshalling response: %s\n", err)
	}

	internal.OutputTable(unmarshalled, veeamcloudconnectColumnsToDisplay, jsonOutput, yamlOutput)
}

func getVeeamCloudConnect(_ *cobra.Command, args []string) {
	var object map[string]any
	if err := client.Get(fmt.Sprintf("/veeamCloudConnect/%s", url.PathEscape(args[0])), &object); err != nil {
		log.Fatalf("error fetching VeeamCloudConnect: %s\n", err)
	}

	internal.OutputObject(object, veeamcloudconnectColumnsToDisplay[0], jsonOutput, yamlOutput)
}

func init() {
	veeamcloudconnectCmd := &cobra.Command{
		Use:   "veeamcloudconnect",
		Short: "Retrieve information and manage your VeeamCloudConnect services",
	}

	// Command to list VeeamCloudConnect services
	veeamcloudconnectCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List your VeeamCloudConnect services",
		Run:   listVeeamCloudConnect,
	})

	// Command to get a single VeeamCloudConnect
	veeamcloudconnectCmd.AddCommand(&cobra.Command{
		Use:        "get",
		Short:      "Retrieve information of a specific VeeamCloudConnect",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"service_name"},
		Run:        getVeeamCloudConnect,
	})

	rootCmd.AddCommand(veeamcloudconnectCmd)
}
