package cmd

import (
	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/flags"
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

	editStorageS3Cmd := &cobra.Command{
		Use:   "edit <container_name>",
		Short: "Edit the given S3™* compatible storage container (* S3 is a trademark filed by Amazon Technologies,Inc. OVHcloud's service is not sponsored by, endorsed by, or otherwise affiliated with Amazon Technologies,Inc.)",
		Run:   cloud.EditStorageS3,
		Args:  cobra.ExactArgs(1),
	}
	editStorageS3Cmd.Flags().StringVar(&cloud.StorageS3Spec.Encryption.SSEAlgorithm, "encryption-sse-algorithm", "", "Encryption SSE Algorithm (AES256, plaintext)")
	editStorageS3Cmd.Flags().StringVar(&cloud.StorageS3Spec.ObjectLock.Rule.Mode, "object-lock-rule-mode", "", "Object lock mode (compliance, governance)")
	editStorageS3Cmd.Flags().StringVar(&cloud.StorageS3Spec.ObjectLock.Rule.Period, "object-lock-rule-period", "", "Object lock period (e.g., P3Y6M4DT12H30M5S)")
	editStorageS3Cmd.Flags().StringVar(&cloud.StorageS3Spec.ObjectLock.Status, "object-lock-status", "", "Object lock status (disabled, enabled)")
	editStorageS3Cmd.Flags().StringToStringVar(&cloud.StorageS3Spec.Tags, "tag", nil, "Container tags as key=value pairs")
	editStorageS3Cmd.Flags().StringVar(&cloud.StorageS3Spec.Versioning.Status, "versioning-status", "", "Versioning status (disabled, enabled, suspended)")
	editStorageS3Cmd.Flags().BoolVar(&flags.ParametersViaEditor, "editor", false, "Use a text editor to define parameters")
	storageS3Cmd.AddCommand(editStorageS3Cmd)

	cloudCmd.AddCommand(storageS3Cmd)
}
