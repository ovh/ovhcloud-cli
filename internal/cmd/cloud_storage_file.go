// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/ovh/ovhcloud-cli/internal/assets"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	"github.com/ovh/ovhcloud-cli/internal/services/cloud"
	"github.com/spf13/cobra"
)

func initCloudStorageFileCommand(cloudCmd *cobra.Command) {
	storageFileCmd := &cobra.Command{
		Use:   "file",
		Short: "Manage file storage in the given cloud project",
	}
	storageFileCmd.PersistentFlags().StringVar(&cloud.CloudProject, "cloud-project", "", "Cloud project ID")

	initCloudStorageFileShareCommand(storageFileCmd)
	initCloudStorageFileSnapshotCommand(storageFileCmd)
	initCloudStorageFileNetworkCommand(storageFileCmd)

	cloudCmd.AddCommand(storageFileCmd)
}

func initCloudStorageFileShareCommand(storageFileCmd *cobra.Command) {
	shareCmd := &cobra.Command{
		Use:   "share",
		Short: "Manage file storage shares",
	}
	storageFileCmd.AddCommand(shareCmd)

	shareListCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List file storage shares",
		Run:     cloud.ListShares,
	}
	shareCmd.AddCommand(withFilterFlag(shareListCmd))

	shareCmd.AddCommand(&cobra.Command{
		Use:   "get <share_id>",
		Short: "Get a specific file storage share",
		Run:   cloud.GetShare,
		Args:  cobra.ExactArgs(1),
	})

	shareCmd.AddCommand(getShareCreateCmd())
	shareCmd.AddCommand(getShareEditCmd())

	shareCmd.AddCommand(&cobra.Command{
		Use:   "delete <share_id>",
		Short: "Delete the given file storage share",
		Run:   cloud.DeleteShare,
		Args:  cobra.ExactArgs(1),
	})
}

func getShareCreateCmd() *cobra.Command {
	shareCreateCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new file storage share",
		Run:   cloud.CreateShare,
	}
	shareCreateCmd.Flags().StringVar(&cloud.ShareSpec.TargetSpec.Name, "name", "", "Name of the file storage share")
	shareCreateCmd.Flags().StringVar(&cloud.ShareSpec.TargetSpec.Description, "description", "", "Description of the file storage share")
	shareCreateCmd.Flags().StringVar(&cloud.ShareSpec.TargetSpec.Protocol, "protocol", "", "File sharing protocol (NFS)")
	shareCreateCmd.Flags().StringVar(&cloud.ShareSpec.TargetSpec.ShareType, "share-type", "", "File storage type / performance tier (STANDARD_1AZ)")
	shareCreateCmd.Flags().IntVar(&cloud.ShareSpec.TargetSpec.Size, "size", 0, "Size of the file storage share in GB")
	shareCreateCmd.Flags().StringVar(&cloud.ShareLocationRegion, "region", "", "Region where the share is created")
	shareCreateCmd.Flags().StringVar(&cloud.ShareLocationAvailabilityZone, "availability-zone", "", "Availability zone within the region")
	shareCreateCmd.Flags().StringVar(&cloud.ShareNetworkID, "share-network-id", "", "ID of the share network to attach the share to")
	shareCreateCmd.Flags().BoolVar(&flags.WaitForTask, "wait", false, "Wait for the share to be ready before exiting")

	addParameterFileFlags(shareCreateCmd, false, assets.CloudV2OpenapiSchema, "/publicCloud/project/{projectId}/storage/file/share", "post", cloud.FileStorageShareCreateExample, nil)
	addInteractiveEditorFlag(shareCreateCmd)
	markFlagsMutuallyExclusive(shareCreateCmd, "from-file", "editor")

	return shareCreateCmd
}

func getShareEditCmd() *cobra.Command {
	shareEditCmd := &cobra.Command{
		Use:   "edit <share_id>",
		Short: "Edit the given file storage share",
		Run:   cloud.EditShare,
		Args:  cobra.ExactArgs(1),
	}
	shareEditCmd.Flags().StringVar(&cloud.ShareEditSpec.TargetSpec.Name, "name", "", "Name of the file storage share")
	shareEditCmd.Flags().StringVar(&cloud.ShareEditSpec.TargetSpec.Description, "description", "", "Description of the file storage share")
	shareEditCmd.Flags().IntVar(&cloud.ShareEditSpec.TargetSpec.Size, "size", 0, "New size of the file storage share in GB")

	addParameterFileFlags(shareEditCmd, true, assets.CloudV2OpenapiSchema, "/publicCloud/project/{projectId}/storage/file/share/{fileStorageId}", "put", "", nil)
	addInteractiveEditorFlag(shareEditCmd)
	markFlagsMutuallyExclusive(shareEditCmd, "from-file", "editor")

	return shareEditCmd
}

func initCloudStorageFileSnapshotCommand(storageFileCmd *cobra.Command) {
	snapshotCmd := &cobra.Command{
		Use:   "snapshot",
		Short: "Manage file storage snapshots",
	}
	storageFileCmd.AddCommand(snapshotCmd)

	snapshotListCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List file storage snapshots",
		Run:     cloud.ListFileStorageSnapshots,
	}
	snapshotCmd.AddCommand(withFilterFlag(snapshotListCmd))

	snapshotCmd.AddCommand(&cobra.Command{
		Use:   "get <snapshot_id>",
		Short: "Get a specific file storage snapshot",
		Run:   cloud.GetFileStorageSnapshot,
		Args:  cobra.ExactArgs(1),
	})

	snapshotCreateCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new file storage snapshot",
		Run:   cloud.CreateFileStorageSnapshot,
	}
	snapshotCreateCmd.Flags().StringVar(&cloud.SnapshotSpec.TargetSpec.Name, "name", "", "Name of the snapshot")
	snapshotCreateCmd.Flags().StringVar(&cloud.SnapshotSpec.TargetSpec.Description, "description", "", "Description of the snapshot")
	snapshotCreateCmd.Flags().StringVar(&cloud.SnapshotShareID, "share-id", "", "ID of the file storage share to snapshot")
	snapshotCreateCmd.Flags().BoolVar(&flags.WaitForTask, "wait", false, "Wait for the snapshot to be ready before exiting")
	snapshotCmd.AddCommand(snapshotCreateCmd)

	snapshotEditCmd := &cobra.Command{
		Use:   "edit <snapshot_id>",
		Short: "Edit the given file storage snapshot",
		Run:   cloud.EditFileStorageSnapshot,
		Args:  cobra.ExactArgs(1),
	}
	snapshotEditCmd.Flags().StringVar(&cloud.SnapshotEditSpec.TargetSpec.Name, "name", "", "Name of the snapshot")
	snapshotEditCmd.Flags().StringVar(&cloud.SnapshotEditSpec.TargetSpec.Description, "description", "", "Description of the snapshot")
	addInteractiveEditorFlag(snapshotEditCmd)
	snapshotCmd.AddCommand(snapshotEditCmd)

	snapshotCmd.AddCommand(&cobra.Command{
		Use:   "delete <snapshot_id>",
		Short: "Delete the given file storage snapshot",
		Run:   cloud.DeleteFileStorageSnapshot,
		Args:  cobra.ExactArgs(1),
	})
}

func initCloudStorageFileNetworkCommand(storageFileCmd *cobra.Command) {
	networkCmd := &cobra.Command{
		Use:   "network",
		Short: "Manage file storage share networks",
	}
	storageFileCmd.AddCommand(networkCmd)

	networkListCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List file storage share networks",
		Run:     cloud.ListFileStorageNetworks,
	}
	networkCmd.AddCommand(withFilterFlag(networkListCmd))

	networkCmd.AddCommand(&cobra.Command{
		Use:   "get <network_id>",
		Short: "Get a specific file storage share network",
		Run:   cloud.GetFileStorageNetwork,
		Args:  cobra.ExactArgs(1),
	})

	networkCmd.AddCommand(getFileStorageNetworkCreateCmd())

	networkCmd.AddCommand(&cobra.Command{
		Use:   "delete <network_id>",
		Short: "Delete the given file storage share network",
		Run:   cloud.DeleteFileStorageNetwork,
		Args:  cobra.ExactArgs(1),
	})
}

func getFileStorageNetworkCreateCmd() *cobra.Command {
	networkCreateCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new file storage share network",
		Run:   cloud.CreateFileStorageNetwork,
	}
	networkCreateCmd.Flags().StringVar(&cloud.NetworkSpec.TargetSpec.Name, "name", "", "Name of the share network")
	networkCreateCmd.Flags().StringVar(&cloud.NetworkSpec.TargetSpec.Description, "description", "", "Description of the share network")
	networkCreateCmd.Flags().StringVar(&cloud.NetworkLocationRegion, "region", "", "Region where the share network is created")
	networkCreateCmd.Flags().StringVar(&cloud.NetworkLocationAvailabilityZone, "availability-zone", "", "Availability zone within the region")
	networkCreateCmd.Flags().StringVar(&cloud.NetworkNetworkID, "network-id", "", "ID of the private network to back the share network")
	networkCreateCmd.Flags().StringVar(&cloud.NetworkSubnetID, "subnet-id", "", "ID of the subnet to back the share network")
	networkCreateCmd.Flags().BoolVar(&flags.WaitForTask, "wait", false, "Wait for the share network to be ready before exiting")

	addParameterFileFlags(networkCreateCmd, false, assets.CloudV2OpenapiSchema, "/publicCloud/project/{projectId}/storage/file/network", "post", cloud.FileStorageNetworkCreateExample, nil)
	addInteractiveEditorFlag(networkCreateCmd)
	markFlagsMutuallyExclusive(networkCreateCmd, "from-file", "editor")

	return networkCreateCmd
}
