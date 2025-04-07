
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
	sslgatewayColumnsToDisplay = []string{ "serviceName","displayName","state","zones" }
)

func listSslGateway(_ *cobra.Command, _ []string) {
	req, err := client.NewRequest(http.MethodGet, "/sslGateway", nil, true)
	if err != nil {
		log.Fatalf("error crafting request: %s\n", err)
	}

	req.Header.Set("X-Pagination-Mode", "CachedObjectList-Pages")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("error fetching /sslGateway: %s\n", err)
	}

	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("error reading response body: %s", err)
	}

	internal.OutputTable(bodyBytes, sslgatewayColumnsToDisplay, jsonOutput, yamlOutput)
}

func getSslGateway(_ *cobra.Command, args []string) {
	var object map[string]any
	if err := client.Get(fmt.Sprintf("/sslGateway/%s", url.PathEscape(args[0])), &object); err != nil {
		log.Fatalf("error fetching SslGateway: %s\n", err)
	}

	internal.OutputObject(object, sslgatewayColumnsToDisplay[0], jsonOutput, yamlOutput)
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
