
package cmd

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal"
)

var (
	sslgatewayColumnsToDisplay = []string{ "serviceName","displayName","state","zones" }
)

func listSslGateway(_ *cobra.Command, _ []string) {
	req, err := client.NewRequest(http.MethodGet, "/sslGateway", nil, true)
	if err != nil {
		fmt.Printf("error crafting request: %s\n", err)
		return
	}

	req.Header.Set("X-Pagination-Mode", "CachedObjectList-Pages")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("error fetching /sslGateway: %s\n", err)
		return
	}

	var unmarshalled []map[string]any
	if err := client.UnmarshalResponse(resp, &unmarshalled); err != nil {
		fmt.Printf("error unmarshalling response: %s\n", err)
		return
	}

	internal.RenderTable(unmarshalled, sslgatewayColumnsToDisplay)
}

func getSslGateway(_ *cobra.Command, args []string) {
	var object map[string]any
	if err := client.Get(fmt.Sprintf("/sslGateway/%s", url.PathEscape(args[0])), &object); err != nil {
		fmt.Printf("error fetching SslGateway: %s\n", err)
		return
	}

	internal.RenderObject(object, sslgatewayColumnsToDisplay[0])
}

func init() {
	sslgatewayCmd := &cobra.Command{
		Use:   "sslgateway",
		Short: "Retrieve information and manage your SslGateway services",
	}

	// Command to list SslGateway services
	sslgatewayCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List your SslGateway services",
		Run:   listSslGateway,
	})

	// Command to get a single SslGateway
	sslgatewayCmd.AddCommand(&cobra.Command{
		Use:        "get",
		Short:      "Retrieve information of a specific SslGateway",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"service_name"},
		Run:        getSslGateway,
	})

	rootCmd.AddCommand(sslgatewayCmd)
}
