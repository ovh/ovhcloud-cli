package cmd

import (
	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/services/cloud"
)

func initCloudStorageS3Command(cloudCmd *cobra.Command) {
	storageS3Cmd := &cobra.Command{
		Use:   "storage-s3",
		Short: "Manage S3™* compatible storage containers in the given cloud project (* S3 is a trademark filed by Amazon Technologies,Inc. OVHcloud's service is not sponsored by, endorsed by, or otherwise affiliated with Amazon Technologies,Inc.)",
	}
	storageS3Cmd.PersistentFlags().StringVar(&cloud.CloudProject, "cloud-project", "", "Cloud project ID")

	storageS3ListCmd := &cobra.Command{
		Use:   "list",
		Short: "List S3™* compatible storage containers (* S3 is a trademark filed by Amazon Technologies,Inc. OVHcloud's service is not sponsored by, endorsed by, or otherwise affiliated with Amazon Technologies,Inc.)",
		Run:   cloud.ListCloudStorageS3,
	}
	storageS3Cmd.AddCommand(withFilterFlag(storageS3ListCmd))

	storageS3Cmd.AddCommand(&cobra.Command{
		Use:   "get <container_name>",
		Short: "Get a specific S3™* compatible storage container (* S3 is a trademark filed by Amazon Technologies,Inc. OVHcloud's service is not sponsored by, endorsed by, or otherwise affiliated with Amazon Technologies,Inc.)",
		Run:   cloud.GetStorageS3,
		Args:  cobra.ExactArgs(1),
	})

	storageS3Cmd.AddCommand(&cobra.Command{
		Use:   "edit <container_name>",
		Short: "Edit the given S3™* compatible storage container (* S3 is a trademark filed by Amazon Technologies,Inc. OVHcloud's service is not sponsored by, endorsed by, or otherwise affiliated with Amazon Technologies,Inc.)",
		Run:   cloud.EditStorageS3,
	})

	cloudCmd.AddCommand(storageS3Cmd)
}
