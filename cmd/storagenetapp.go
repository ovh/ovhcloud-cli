package cmd

import (
	_ "embed"

	"github.com/spf13/cobra"
)

var (
	storagenetappColumnsToDisplay = []string{ "id","name","region","status" }

	//go:embed templates/storagenetapp.tmpl
	storagenetappTemplate string
)

func listStorageNetApp(_ *cobra.Command, _ []string) {
	manageListRequest("/storage/netapp", storagenetappColumnsToDisplay, genericFilters)
}

func getStorageNetApp(_ *cobra.Command, args []string) {
	manageObjectRequest("/storage/netapp", args[0], storagenetappTemplate)
}

func init() {
	storagenetappCmd := &cobra.Command{
		Use:   "storagenetapp",
		Short: "Retrieve information and manage your StorageNetApp services",
	}

	// Command to list StorageNetApp services
	storagenetappListCmd := &cobra.Command{
		Use:   "list",
		Short: "List your StorageNetApp services",
		Run:   listStorageNetApp,
	}
	storagenetappListCmd.PersistentFlags().StringArrayVar(
		&genericFilters,
		"filter",
		nil,
		"Filter results by any property using github.com/PaesslerAG/gval syntax'",
	)
	storagenetappCmd.AddCommand(storagenetappListCmd)

	// Command to get a single StorageNetApp
	storagenetappCmd.AddCommand(&cobra.Command{
		Use:        "get",
		Short:      "Retrieve information of a specific StorageNetApp",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"service_name"},
		Run:        getStorageNetApp,
	})

	rootCmd.AddCommand(storagenetappCmd)
}
