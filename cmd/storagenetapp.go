
package cmd

import (
	"github.com/spf13/cobra"
)

var (
	storagenetappColumnsToDisplay = []string{ "id","name","region","status" }
)

func listStorageNetApp(_ *cobra.Command, _ []string) {
	manageListRequest("/storage/netapp", storagenetappColumnsToDisplay)
}

func getStorageNetApp(_ *cobra.Command, args []string) {
	manageObjectRequest("/storage/netapp", args[0], storagenetappColumnsToDisplay[0])
}

func init() {
	storagenetappCmd := &cobra.Command{
		Use:   "storagenetapp",
		Short: "Retrieve information and manage your StorageNetApp services",
	}

	// Command to list StorageNetApp services
	storagenetappCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List your StorageNetApp services",
		Run:   listStorageNetApp,
	})

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
