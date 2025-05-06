package cmd

import (
	_ "embed"
	"fmt"
	"net/url"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/display"
)

var (
	domainzoneColumnsToDisplay = []string{"name", "dnssecSupported", "hasDnsAnycast"}

	//go:embed templates/domainzone.tmpl
	domainzoneTemplate string
)

func listDomainZone(_ *cobra.Command, _ []string) {
	manageListRequest("/domain/zone", "", domainzoneColumnsToDisplay, genericFilters)
}

func getDomainZone(_ *cobra.Command, args []string) {
	path := fmt.Sprintf("/domain/zone/%s", url.PathEscape(args[0]))

	// Fetch domain zone
	var object map[string]any
	if err := client.Get(path, &object); err != nil {
		display.ExitError("error fetching %s: %s\n", path, err)
	}

	// Fetch running tasks
	path = fmt.Sprintf("/domain/zone/%s/record", url.PathEscape(args[0]))
	records, err := fetchExpandedArray(path, "")
	if err != nil {
		display.ExitError("error fetching records for %s: %s", args[0], err)
	}
	object["records"] = records

	display.OutputObject(object, args[0], domainzoneTemplate, &outputFormatConfig)
}

func init() {
	domainzoneCmd := &cobra.Command{
		Use:   "domainzone",
		Short: "Retrieve information and manage your DomainZone services",
	}

	// Command to list DomainZone services
	domainzoneListCmd := &cobra.Command{
		Use:   "list",
		Short: "List your DomainZone services",
		Run:   listDomainZone,
	}
	domainzoneCmd.AddCommand(withFilterFlag(domainzoneListCmd))

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
