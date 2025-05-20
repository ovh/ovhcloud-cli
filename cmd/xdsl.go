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
		Short: "Retrieve information and manage your XDSL services",
	}

	// Command to list Xdsl services
	xdslListCmd := &cobra.Command{
		Use:   "list",
		Short: "List your XDSL services",
		Run:   listXdsl,
	}
	xdslCmd.AddCommand(withFilterFlag(xdslListCmd))

	// Command to get a single Xdsl
	xdslCmd.AddCommand(&cobra.Command{
		Use:   "get <service_name>",
		Short: "Retrieve information of a specific XDSL",
		Args:  cobra.ExactArgs(1),
		Run:   getXdsl,
	})

	rootCmd.AddCommand(xdslCmd)
}
