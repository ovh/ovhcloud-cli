package cmd

import (
	"github.com/ovh/ovhcloud-cli/internal/services/storagenetapp"
	"github.com/spf13/cobra"
)

func init() {
	storagenetappCmd := &cobra.Command{
		Use:   "storage-netapp",
		Short: "Retrieve information and manage your Storage NetApp services",
	}

	// Command to list StorageNetApp services
	storagenetappListCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List your Storage NetApp services",
		Run:     storagenetapp.ListStorageNetApp,
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
	storagenetappEditCmd := &cobra.Command{
		Use:   "edit <service_name>",
		Short: "Edit the given StorageNetApp",
		Args:  cobra.ExactArgs(1),
		Run:   storagenetapp.EditStorageNetApp,
	}
	storagenetappEditCmd.Flags().StringVar(&storagenetapp.StorageNetAppSpec.Name, "name", "", "Name of the Storage NetApp")
	addInteractiveEditorFlag(storagenetappEditCmd)
	storagenetappCmd.AddCommand(storagenetappEditCmd)

	rootCmd.AddCommand(storagenetappCmd)
}
