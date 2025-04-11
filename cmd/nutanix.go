package cmd

import (
	_ "embed"

	"github.com/spf13/cobra"
)

var (
	nutanixColumnsToDisplay = []string{ "serviceName","status" }

	//go:embed templates/nutanix.tmpl
	nutanixTemplate string
)

func listNutanix(_ *cobra.Command, _ []string) {
	manageListRequest("/nutanix", nutanixColumnsToDisplay, genericFilters)
}

func getNutanix(_ *cobra.Command, args []string) {
	manageObjectRequest("/nutanix", args[0], nutanixTemplate)
}

func init() {
	nutanixCmd := &cobra.Command{
		Use:   "nutanix",
		Short: "Retrieve information and manage your Nutanix services",
	}

	// Command to list Nutanix services
	nutanixListCmd := &cobra.Command{
		Use:   "list",
		Short: "List your Nutanix services",
		Run:   listNutanix,
	}
	nutanixListCmd.PersistentFlags().StringArrayVar(
		&genericFilters,
		"filter",
		nil,
		"Filter results by any property using github.com/PaesslerAG/gval syntax'",
	)
	nutanixCmd.AddCommand(nutanixListCmd)

	// Command to get a single Nutanix
	nutanixCmd.AddCommand(&cobra.Command{
		Use:        "get",
		Short:      "Retrieve information of a specific Nutanix",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"service_name"},
		Run:        getNutanix,
	})

	rootCmd.AddCommand(nutanixCmd)
}
