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
		Use:   "cdn-dedicated",
		Short: "Retrieve information and manage your dedicated CDN services",
	}

	// Command to list CdnDedicated services
	cdndedicatedListCmd := &cobra.Command{
		Use:   "list",
		Short: "List your dedicated CDN services",
		Run:   listCdnDedicated,
	}
	cdndedicatedCmd.AddCommand(withFilterFlag(cdndedicatedListCmd))

	// Command to get a single CdnDedicated
	cdndedicatedCmd.AddCommand(&cobra.Command{
		Use:   "get <service_name>",
		Short: "Retrieve information of a specific CDN",
		Args:  cobra.ExactArgs(1),
		Run:   getCdnDedicated,
	})

	rootCmd.AddCommand(cdndedicatedCmd)
}
