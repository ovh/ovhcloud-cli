
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
	veeamenterpriseColumnsToDisplay = []string{ "serviceName","activationStatus","ip","sourceIp" }
)

func listVeeamEnterprise(_ *cobra.Command, _ []string) {
	req, err := client.NewRequest(http.MethodGet, "/veeam/veeamEnterprise", nil, true)
	if err != nil {
		log.Fatalf("error crafting request: %s\n", err)
	}

	req.Header.Set("X-Pagination-Mode", "CachedObjectList-Pages")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("error fetching /veeam/veeamEnterprise: %s\n", err)
	}

	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("error reading response body: %s", err)
	}

	internal.OutputTable(bodyBytes, veeamenterpriseColumnsToDisplay, jsonOutput, yamlOutput)
}

func getVeeamEnterprise(_ *cobra.Command, args []string) {
	var object map[string]any
	if err := client.Get(fmt.Sprintf("/veeam/veeamEnterprise/%s", url.PathEscape(args[0])), &object); err != nil {
		log.Fatalf("error fetching VeeamEnterprise: %s\n", err)
	}

	internal.OutputObject(object, veeamenterpriseColumnsToDisplay[0], jsonOutput, yamlOutput)
}

func init() {
	veeamenterpriseCmd := &cobra.Command{
		Use:   "veeamenterprise",
		Short: "Retrieve information and manage your VeeamEnterprise services",
	}

	// Command to list VeeamEnterprise services
	veeamenterpriseCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List your VeeamEnterprise services",
		Run:   listVeeamEnterprise,
	})

	// Command to get a single VeeamEnterprise
	veeamenterpriseCmd.AddCommand(&cobra.Command{
		Use:        "get",
		Short:      "Retrieve information of a specific VeeamEnterprise",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"service_name"},
		Run:        getVeeamEnterprise,
	})

	rootCmd.AddCommand(veeamenterpriseCmd)
}
