package cmd

import (
	_ "embed"

	"github.com/spf13/cobra"
)

var (
	vpsColumnsToDisplay = []string{"name", "displayName", "state", "zone"}

	//go:embed templates/vps.tmpl
	vpsTemplate string
)

func listVps(_ *cobra.Command, _ []string) {
	manageListRequest("/vps", "", vpsColumnsToDisplay, genericFilters)
}

func getVps(_ *cobra.Command, args []string) {
	manageObjectRequest("/vps", args[0], vpsTemplate)
}

func init() {
	vpsCmd := &cobra.Command{
		Use:   "vps",
		Short: "Retrieve information and manage your Vps services",
	}

	// Command to list Vps services
	vpsListCmd := &cobra.Command{
		Use:   "list",
		Short: "List your Vps services",
		Run:   listVps,
	}
	vpsListCmd.PersistentFlags().StringArrayVar(
		&genericFilters,
		"filter",
		nil,
		"Filter results by any property using github.com/PaesslerAG/gval syntax'",
	)
	vpsCmd.AddCommand(vpsListCmd)

	// Command to get a single Vps
	vpsCmd.AddCommand(&cobra.Command{
		Use:        "get",
		Short:      "Retrieve information of a specific Vps",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"service_name"},
		Run:        getVps,
	})

	rootCmd.AddCommand(vpsCmd)
}
