package cmd

import (
	_ "embed"

	"github.com/spf13/cobra"
)

var (
	alldomColumnsToDisplay = []string{"name", "type", "offer"}

	//go:embed templates/alldom.tmpl
	alldomTemplate string
)

func listAllDom(_ *cobra.Command, _ []string) {
	manageListRequest("/allDom", "", alldomColumnsToDisplay, genericFilters)
}

func getAllDom(_ *cobra.Command, args []string) {
	manageObjectRequest("/allDom", args[0], alldomTemplate)
}

func init() {
	alldomCmd := &cobra.Command{
		Use:   "alldom",
		Short: "Retrieve information and manage your AllDom services",
	}

	// Command to list AllDom services
	alldomListCmd := &cobra.Command{
		Use:   "list",
		Short: "List your AllDom services",
		Run:   listAllDom,
	}
	alldomCmd.AddCommand(withFilterFlag(alldomListCmd))

	// Command to get a single AllDom
	alldomCmd.AddCommand(&cobra.Command{
		Use:        "get",
		Short:      "Retrieve information of a specific AllDom",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"service_name"},
		Run:        getAllDom,
	})

	rootCmd.AddCommand(alldomCmd)
}
