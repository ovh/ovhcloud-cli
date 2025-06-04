package cmd

import (
	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/services/storagenetapp"
)

func init() {
	storagenetappCmd := &cobra.Command{
		Use:   "storage-netapp",
		Short: "Retrieve information and manage your Storage NetApp services",
	}

	// Command to list StorageNetApp services
	storagenetappListCmd := &cobra.Command{
		Use:   "list",
		Short: "List your Storage NetApp services",
		Run:   storagenetapp.ListStorageNetApp,
	}
	storagenetappCmd.AddCommand(withFilterFlag(storagenetappListCmd))

	// Command to get a single StorageNetApp
	storagenetappCmd.AddCommand(&cobra.Command{
		Use:   "get <service_name>",
		Short: "Retrieve information of a specific StorageNetApp",
		Args:  cobra.ExactArgs(1),
		Run:   storagenetapp.GetStorageNetApp,
	})

	// Command to update a single StorageNetApp
	storagenetappCmd.AddCommand(&cobra.Command{
		Use:   "edit <service_name>",
		Short: "Edit the given StorageNetApp",
		Run:   storagenetapp.EditStorageNetApp,
	})

	rootCmd.AddCommand(storagenetappCmd)
}
