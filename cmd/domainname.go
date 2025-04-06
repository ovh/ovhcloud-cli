
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
	domainnameColumnsToDisplay = []string{ "domain","state","whoisOwner","expirationDate","renewalDate" }
)

func listDomainName(_ *cobra.Command, _ []string) {
	req, err := client.NewRequest(http.MethodGet, "/domain", nil, true)
	if err != nil {
		log.Fatalf("error crafting request: %s\n", err)
	}

	req.Header.Set("X-Pagination-Mode", "CachedObjectList-Pages")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("error fetching /domain: %s\n", err)
	}

	var unmarshalled []map[string]any
	if err := client.UnmarshalResponse(resp, &unmarshalled); err != nil {
		log.Fatalf("error unmarshalling response: %s\n", err)
	}

	internal.OutputTable(unmarshalled, domainnameColumnsToDisplay, jsonOutput, yamlOutput)
}

func getDomainName(_ *cobra.Command, args []string) {
	var object map[string]any
	if err := client.Get(fmt.Sprintf("/domain/%s", url.PathEscape(args[0])), &object); err != nil {
		log.Fatalf("error fetching DomainName: %s\n", err)
	}

	internal.OutputObject(object, domainnameColumnsToDisplay[0], jsonOutput, yamlOutput)
}

func init() {
	domainnameCmd := &cobra.Command{
		Use:   "domainname",
		Short: "Retrieve information and manage your DomainName services",
	}

	// Command to list DomainName services
	domainnameCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List your DomainName services",
		Run:   listDomainName,
	})

	// Command to get a single DomainName
	domainnameCmd.AddCommand(&cobra.Command{
		Use:        "get",
		Short:      "Retrieve information of a specific DomainName",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"service_name"},
		Run:        getDomainName,
	})

	rootCmd.AddCommand(domainnameCmd)
}
