package cmd

import (
	_ "embed"

	"github.com/spf13/cobra"
)

var (
	xdslColumnsToDisplay = []string{"accessName", "accessType", "provider", "role", "status"}

	//go:embed templates/xdsl.tmpl
	xdslTemplate string
)

func listXdsl(_ *cobra.Command, _ []string) {
	manageListRequest("/xdsl", "", xdslColumnsToDisplay, genericFilters)
}

func getXdsl(_ *cobra.Command, args []string) {
	manageObjectRequest("/xdsl", args[0], xdslTemplate)
}

func init() {
	xdslCmd := &cobra.Command{
		Use:   "xdsl",
		Short: "Retrieve information and manage your Xdsl services",
	}

	// Command to list Xdsl services
	xdslListCmd := &cobra.Command{
		Use:   "list",
		Short: "List your Xdsl services",
		Run:   listXdsl,
	}
	xdslCmd.AddCommand(withFilterFlag(xdslListCmd))

	// Command to get a single Xdsl
	xdslCmd.AddCommand(&cobra.Command{
		Use:        "get",
		Short:      "Retrieve information of a specific Xdsl",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"service_name"},
		Run:        getXdsl,
	})

	rootCmd.AddCommand(xdslCmd)
}
