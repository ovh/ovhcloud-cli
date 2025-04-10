package cmd

import (
	"github.com/spf13/cobra"
)

var (
	domainzoneColumnsToDisplay = []string{ "name","dnssecSupported","hasDnsAnycast" }
)

func listDomainZone(_ *cobra.Command, _ []string) {
	manageListRequest("/domain/zone", domainzoneColumnsToDisplay, genericFilters)
}

func getDomainZone(_ *cobra.Command, args []string) {
	manageObjectRequest("/domain/zone", args[0], domainzoneColumnsToDisplay[0])
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
	domainzoneListCmd.PersistentFlags().StringArrayVar(
		&genericFilters,
		"filter",
		nil,
		"Filter results by any property using github.com/PaesslerAG/gval syntax'",
	)
	domainzoneCmd.AddCommand(domainzoneListCmd)

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
