
package cmd

import (
	"github.com/spf13/cobra"
)

var (
	ldpColumnsToDisplay = []string{ "serviceName","displayName","isClusterOwner","state","username" }
)

func listLdp(_ *cobra.Command, _ []string) {
	manageListRequest("/dbaas/logs", ldpColumnsToDisplay)
}

func getLdp(_ *cobra.Command, args []string) {
	manageObjectRequest("/dbaas/logs", args[0], ldpColumnsToDisplay[0])
}

func init() {
	ldpCmd := &cobra.Command{
		Use:   "ldp",
		Short: "Retrieve information and manage your Ldp services",
	}

	// Command to list Ldp services
	ldpCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List your Ldp services",
		Run:   listLdp,
	})

	// Command to get a single Ldp
	ldpCmd.AddCommand(&cobra.Command{
		Use:        "get",
		Short:      "Retrieve information of a specific Ldp",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"service_name"},
		Run:        getLdp,
	})

	rootCmd.AddCommand(ldpCmd)
}
