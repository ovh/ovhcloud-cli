package cmd

import (
	_ "embed"

	"github.com/spf13/cobra"
)

var (
	storagenetappColumnsToDisplay = []string{"id", "name", "region", "status"}

	//go:embed templates/storagenetapp.tmpl
	storagenetappTemplate string
)

func listStorageNetApp(_ *cobra.Command, _ []string) {
	manageListRequest("/storage/netapp", "", storagenetappColumnsToDisplay, genericFilters)
}

func getStorageNetApp(_ *cobra.Command, args []string) {
	manageObjectRequest("/storage/netapp", args[0], storagenetappTemplate)
}

func init() {
	storagenetappCmd := &cobra.Command{
		Use:   "storage-netapp",
		Short: "Retrieve information and manage your Storage NetApp services",
	}

	// Command to list StorageNetApp services
	storagenetappListCmd := &cobra.Command{
		Use:   "list",
		Short: "List your Storage NetApp services",
		Run:   listStorageNetApp,
	}
	storagenetappCmd.AddCommand(withFilterFlag(storagenetappListCmd))

	// Command to get a single StorageNetApp
	storagenetappCmd.AddCommand(&cobra.Command{
		Use:   "get <service_name>",
		Short: "Retrieve information of a specific StorageNetApp",
		Args:  cobra.ExactArgs(1),
		Run:   getStorageNetApp,
	})

	rootCmd.AddCommand(storagenetappCmd)
}
