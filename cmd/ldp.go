package cmd

import (
	_ "embed"

	"github.com/spf13/cobra"
)

var (
	ldpColumnsToDisplay = []string{ "serviceName","displayName","isClusterOwner","state","username" }

	//go:embed templates/ldp.tmpl
	ldpTemplate string
)

func listLdp(_ *cobra.Command, _ []string) {
	manageListRequest("/dbaas/logs", ldpColumnsToDisplay, genericFilters)
}

func getLdp(_ *cobra.Command, args []string) {
	manageObjectRequest("/dbaas/logs", args[0], ldpTemplate)
}

func init() {
	ldpCmd := &cobra.Command{
		Use:   "ldp",
		Short: "Retrieve information and manage your Ldp services",
	}

	// Command to list Ldp services
	ldpListCmd := &cobra.Command{
		Use:   "list",
		Short: "List your Ldp services",
		Run:   listLdp,
	}
	ldpListCmd.PersistentFlags().StringArrayVar(
		&genericFilters,
		"filter",
		nil,
		"Filter results by any property using github.com/PaesslerAG/gval syntax'",
	)
	ldpCmd.AddCommand(ldpListCmd)

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
