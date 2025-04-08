
package cmd

import (
	"github.com/spf13/cobra"
)

var (
	cdndedicatedColumnsToDisplay = []string{ "service","offer","anycast" }
)

func listCdnDedicated(_ *cobra.Command, _ []string) {
	manageListRequest("/cdn/dedicated", cdndedicatedColumnsToDisplay)
}

func getCdnDedicated(_ *cobra.Command, args []string) {
	manageObjectRequest("/cdn/dedicated", args[0], cdndedicatedColumnsToDisplay[0])
}

func init() {
	cdndedicatedCmd := &cobra.Command{
		Use:   "cdndedicated",
		Short: "Retrieve information and manage your CdnDedicated services",
	}

	// Command to list CdnDedicated services
	cdndedicatedCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List your CdnDedicated services",
		Run:   listCdnDedicated,
	})

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
