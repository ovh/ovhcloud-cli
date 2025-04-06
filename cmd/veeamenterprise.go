
package cmd

import (
	"fmt"
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
		fmt.Printf("error crafting request: %s\n", err)
		return
	}

	req.Header.Set("X-Pagination-Mode", "CachedObjectList-Pages")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("error fetching /veeam/veeamEnterprise: %s\n", err)
		return
	}

	var unmarshalled []map[string]any
	if err := client.UnmarshalResponse(resp, &unmarshalled); err != nil {
		fmt.Printf("error unmarshalling response: %s\n", err)
		return
	}

	internal.RenderTable(unmarshalled, veeamenterpriseColumnsToDisplay)
}

func getVeeamEnterprise(_ *cobra.Command, args []string) {
	var object map[string]any
	if err := client.Get(fmt.Sprintf("/veeam/veeamEnterprise/%s", url.PathEscape(args[0])), &object); err != nil {
		fmt.Printf("error fetching VeeamEnterprise: %s\n", err)
		return
	}

	internal.RenderObject(object, veeamenterpriseColumnsToDisplay[0])
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
