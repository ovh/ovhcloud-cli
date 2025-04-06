
package cmd

import (
	"fmt"
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
		fmt.Printf("error crafting request: %s\n", err)
		return
	}

	req.Header.Set("X-Pagination-Mode", "CachedObjectList-Pages")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("error fetching /domain/zone: %s\n", err)
		return
	}

	var unmarshalled []map[string]any
	if err := client.UnmarshalResponse(resp, &unmarshalled); err != nil {
		fmt.Printf("error unmarshalling response: %s\n", err)
		return
	}

	internal.RenderTable(unmarshalled, domainzoneColumnsToDisplay)
}

func getDomainZone(_ *cobra.Command, args []string) {
	var object map[string]any
	if err := client.Get(fmt.Sprintf("/domain/zone/%s", url.PathEscape(args[0])), &object); err != nil {
		fmt.Printf("error fetching DomainZone: %s\n", err)
		return
	}

	internal.RenderObject(object, domainzoneColumnsToDisplay[0])
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
