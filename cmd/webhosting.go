package cmd

import (
	_ "embed"

	"github.com/spf13/cobra"
)

var (
	webhostingColumnsToDisplay = []string{ "serviceName","displayName","datacenter","state" }

	//go:embed templates/webhosting.tmpl
	webhostingTemplate string
)

func listWebHosting(_ *cobra.Command, _ []string) {
	manageListRequest("/hosting/web", webhostingColumnsToDisplay, genericFilters)
}

func getWebHosting(_ *cobra.Command, args []string) {
	manageObjectRequest("/hosting/web", args[0], webhostingTemplate)
}

func init() {
	webhostingCmd := &cobra.Command{
		Use:   "webhosting",
		Short: "Retrieve information and manage your WebHosting services",
	}

	// Command to list WebHosting services
	webhostingListCmd := &cobra.Command{
		Use:   "list",
		Short: "List your WebHosting services",
		Run:   listWebHosting,
	}
	webhostingListCmd.PersistentFlags().StringArrayVar(
		&genericFilters,
		"filter",
		nil,
		"Filter results by any property using github.com/PaesslerAG/gval syntax'",
	)
	webhostingCmd.AddCommand(webhostingListCmd)

	// Command to get a single WebHosting
	webhostingCmd.AddCommand(&cobra.Command{
		Use:        "get",
		Short:      "Retrieve information of a specific WebHosting",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"service_name"},
		Run:        getWebHosting,
	})

	rootCmd.AddCommand(webhostingCmd)
}
