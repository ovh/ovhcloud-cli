package cmd

import (
	_ "embed"

	"github.com/spf13/cobra"
)

var (
	dedicatednashaColumnsToDisplay = []string{"serviceName", "customName", "datacenter"}

	//go:embed templates/dedicatednasha.tmpl
	dedicatednashaTemplate string
)

func listDedicatedNasHA(_ *cobra.Command, _ []string) {
	manageListRequest("/dedicated/nasha", "", dedicatednashaColumnsToDisplay, genericFilters)
}

func getDedicatedNasHA(_ *cobra.Command, args []string) {
	manageObjectRequest("/dedicated/nasha", args[0], dedicatednashaTemplate)
}

func init() {
	dedicatednashaCmd := &cobra.Command{
		Use:   "dedicatednasha",
		Short: "Retrieve information and manage your DedicatedNasHA services",
	}

	// Command to list DedicatedNasHA services
	dedicatednashaListCmd := &cobra.Command{
		Use:   "list",
		Short: "List your DedicatedNasHA services",
		Run:   listDedicatedNasHA,
	}
	dedicatednashaListCmd.PersistentFlags().StringArrayVar(
		&genericFilters,
		"filter",
		nil,
		"Filter results by any property using github.com/PaesslerAG/gval syntax'",
	)
	dedicatednashaCmd.AddCommand(dedicatednashaListCmd)

	// Command to get a single DedicatedNasHA
	dedicatednashaCmd.AddCommand(&cobra.Command{
		Use:        "get",
		Short:      "Retrieve information of a specific DedicatedNasHA",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"service_name"},
		Run:        getDedicatedNasHA,
	})

	rootCmd.AddCommand(dedicatednashaCmd)
}
