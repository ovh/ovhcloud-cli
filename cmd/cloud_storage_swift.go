package cmd

import (
	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/services/cloud"
)

func initCloudStorageSwiftCommand(cloudCmd *cobra.Command) {
	storageSwiftCmd := &cobra.Command{
		Use:   "storage-swift",
		Short: "Manage SWIFT storage containers in the given cloud project",
	}
	storageSwiftCmd.PersistentFlags().StringVar(&cloud.CloudProject, "cloud-project", "", "Cloud project ID")

	storageSwiftListCmd := &cobra.Command{
		Use:   "list",
		Short: "List SWIFT storage containers",
		Run:   cloud.ListCloudStorageSwift,
	}
	storageSwiftCmd.AddCommand(withFilterFlag(storageSwiftListCmd))

	storageSwiftCmd.AddCommand(&cobra.Command{
		Use:   "get <container_id>",
		Short: "Get a specific SWIFT storage container",
		Run:   cloud.GetStorageSwift,
		Args:  cobra.ExactArgs(1),
	})

	storageSwiftCmd.AddCommand(&cobra.Command{
		Use:   "edit <container_id>",
		Short: "Edit the given SWIFT storage container",
		Run:   cloud.EditStorageSwift,
	})

	cloudCmd.AddCommand(storageSwiftCmd)
}
