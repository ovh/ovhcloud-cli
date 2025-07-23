package cmd

import (
	"github.com/ovh/ovhcloud-cli/internal/services/cloud"
	"github.com/spf13/cobra"
)

func initCloudStorageSwiftCommand(cloudCmd *cobra.Command) {
	storageSwiftCmd := &cobra.Command{
		Use:   "storage-swift",
		Short: "Manage SWIFT storage containers in the given cloud project",
	}
	storageSwiftCmd.PersistentFlags().StringVar(&cloud.CloudProject, "cloud-project", "", "Cloud project ID")

	storageSwiftListCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List SWIFT storage containers",
		Run:     cloud.ListCloudStorageSwift,
	}
	storageSwiftCmd.AddCommand(withFilterFlag(storageSwiftListCmd))

	storageSwiftCmd.AddCommand(&cobra.Command{
		Use:   "get <container_id>",
		Short: "Get a specific SWIFT storage container",
		Run:   cloud.GetStorageSwift,
		Args:  cobra.ExactArgs(1),
	})

	editCmd := &cobra.Command{
		Use:   "edit <container_id>",
		Short: "Edit the given SWIFT storage container",
		Args:  cobra.ExactArgs(1),
		Run:   cloud.EditStorageSwift,
	}
	editCmd.Flags().StringVar(&cloud.CloudSwiftContainerType, "type", "", "Type of the SWIFT storage container (private, public, static)")
	addInteractiveEditorFlag(editCmd)
	storageSwiftCmd.AddCommand(editCmd)

	cloudCmd.AddCommand(storageSwiftCmd)
}
