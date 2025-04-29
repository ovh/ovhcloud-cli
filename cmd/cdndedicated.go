package cmd

import (
	_ "embed"

	"github.com/spf13/cobra"
)

var (
	cdndedicatedColumnsToDisplay = []string{"service", "offer", "anycast"}

	//go:embed templates/cdndedicated.tmpl
	cdndedicatedTemplate string
)

func listCdnDedicated(_ *cobra.Command, _ []string) {
	manageListRequest("/cdn/dedicated", "", cdndedicatedColumnsToDisplay, genericFilters)
}

func getCdnDedicated(_ *cobra.Command, args []string) {
	manageObjectRequest("/cdn/dedicated", args[0], cdndedicatedTemplate)
}

func init() {
	cdndedicatedCmd := &cobra.Command{
		Use:   "cdndedicated",
		Short: "Retrieve information and manage your CdnDedicated services",
	}

	// Command to list CdnDedicated services
	cdndedicatedListCmd := &cobra.Command{
		Use:   "list",
		Short: "List your CdnDedicated services",
		Run:   listCdnDedicated,
	}
	cdndedicatedListCmd.PersistentFlags().StringArrayVar(
		&genericFilters,
		"filter",
		nil,
		"Filter results by any property using github.com/PaesslerAG/gval syntax'",
	)
	cdndedicatedCmd.AddCommand(cdndedicatedListCmd)

	// Command to get a single CdnDedicated
	cdndedicatedCmd.AddCommand(&cobra.Command{
		Use:        "get",
		Short:      "Retrieve information of a specific CdnDedicated",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"service_name"},
		Run:        getCdnDedicated,
	})

	rootCmd.AddCommand(cdndedicatedCmd)
}
