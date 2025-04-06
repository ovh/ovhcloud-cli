
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
	domainzoneColumnsToDisplay = []string{ "name","dnssecSupported","hasDnsAnycast" }
)

func listDomainZone(_ *cobra.Command, _ []string) {
	req, err := client.NewRequest(http.MethodGet, "/domain/zone", nil, true)
	if err != nil {
		log.Fatalf("error crafting request: %s\n", err)
	}

	req.Header.Set("X-Pagination-Mode", "CachedObjectList-Pages")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("error fetching /domain/zone: %s\n", err)
	}

	var unmarshalled []map[string]any
	if err := client.UnmarshalResponse(resp, &unmarshalled); err != nil {
		log.Fatalf("error unmarshalling response: %s\n", err)
	}

	internal.OutputTable(unmarshalled, domainzoneColumnsToDisplay, jsonOutput, yamlOutput)
}

func getDomainZone(_ *cobra.Command, args []string) {
	var object map[string]any
	if err := client.Get(fmt.Sprintf("/domain/zone/%s", url.PathEscape(args[0])), &object); err != nil {
		log.Fatalf("error fetching DomainZone: %s\n", err)
	}

	internal.OutputObject(object, domainzoneColumnsToDisplay[0], jsonOutput, yamlOutput)
}

func init() {
	domainzoneCmd := &cobra.Command{
		Use:   "domainzone",
		Short: "Retrieve information and manage your DomainZone services",
	}

	// Command to list DomainZone services
	domainzoneCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List your DomainZone services",
		Run:   listDomainZone,
	})

	// Command to get a single DomainZone
	domainzoneCmd.AddCommand(&cobra.Command{
		Use:        "get",
		Short:      "Retrieve information of a specific DomainZone",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"service_name"},
		Run:        getDomainZone,
	})

	rootCmd.AddCommand(domainzoneCmd)
}
