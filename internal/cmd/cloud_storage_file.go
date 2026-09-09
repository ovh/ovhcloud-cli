// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/ovh/ovhcloud-cli/internal/assets"
	"github.com/ovh/ovhcloud-cli/internal/services/cloud"
	"github.com/spf13/cobra"
)

func initCloudStorageFileCommand(cloudCmd *cobra.Command) {
	storageFileCmd := &cobra.Command{
		Use:   "file",
		Short: "Manage file storage shares in the given cloud project",
	}
	storageFileCmd.PersistentFlags().StringVar(&cloud.CloudProject, "cloud-project", "", "Cloud project ID")
	storageFileCmd.PersistentFlags().StringVar(&cloud.ShareRegion, "region", "", "Region (skip region discovery if set)")

	networkCmd := &cobra.Command{
		Use:   "network",
		Short: "Manage file storage share networks",
	}
	storageFileCmd.AddCommand(networkCmd)

	networkListCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List share networks",
		Run:     cloud.ListShareNetworks,
	}
	networkCmd.AddCommand(withFilterFlag(networkListCmd))

	networkCmd.AddCommand(&cobra.Command{
		Use:   "get <share_network_id>",
		Short: "Get a share network",
		Run:   cloud.GetShareNetwork,
		Args:  cobra.ExactArgs(1),
	})

	networkCreateCmd := &cobra.Command{
		Use:   "create <region>",
		Short: "Create a share network",
		Run:   cloud.CreateShareNetwork,
		Args:  cobra.ExactArgs(1),
	}
	networkCreateCmd.Flags().StringVar(&cloud.ShareNetworkSpec.TargetSpec.Description, "description", "", "Share network description")
	networkCreateCmd.Flags().StringVar(&cloud.ShareNetworkSpec.TargetSpec.Name, "name", "", "Share network name")
	networkCreateCmd.Flags().StringVar(&cloud.ShareNetworkSpec.TargetSpec.Location.AvailabilityZone, "availability-zone", "", "Availability zone within the region")
	networkCreateCmd.Flags().StringVar(&cloud.ShareNetworkSpec.TargetSpec.Network.Id, "network-id", "", "Private network ID")
	networkCreateCmd.Flags().StringVar(&cloud.ShareNetworkSpec.TargetSpec.Subnet.Id, "subnet-id", "", "Private subnet ID")
	addParameterFileFlags(networkCreateCmd, false, assets.CloudV2OpenapiSchema, "/publicCloud/project/{projectId}/storage/file/network", "post", cloud.ShareNetworkCreateExample, nil)
	addInteractiveEditorFlag(networkCreateCmd)
	markFlagsMutuallyExclusive(networkCreateCmd, "from-file", "editor")
	networkCmd.AddCommand(networkCreateCmd)

	networkCmd.AddCommand(&cobra.Command{
		Use:   "delete <share_network_id>",
		Short: "Delete a share network",
		Run:   cloud.DeleteShareNetwork,
		Args:  cobra.ExactArgs(1),
	})

	// Share commands
	shareCmd := &cobra.Command{
		Use:   "share",
		Short: "Manage file storage shares",
	}
	storageFileCmd.AddCommand(shareCmd)

	shareListCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List shares",
		Run:     cloud.ListShares,
	}
	shareCmd.AddCommand(withFilterFlag(shareListCmd))

	shareCmd.AddCommand(&cobra.Command{
		Use:   "get <share_id>",
		Short: "Get a specific share",
		Run:   cloud.GetShare,
		Args:  cobra.ExactArgs(1),
	})

	shareCmd.AddCommand(getShareCreateCmd())

	shareEditCmd := &cobra.Command{
		Use:   "edit <share_id>",
		Short: "Edit the given share",
		Run:   cloud.EditShare,
		Args:  cobra.ExactArgs(1),
	}
	shareEditCmd.Flags().StringVar(&cloud.ShareEditSpec.TargetSpec.Description, "description", "", "Share description")
	shareEditCmd.Flags().StringVar(&cloud.ShareEditSpec.TargetSpec.Name, "name", "", "Share name")
	shareEditCmd.Flags().IntVar(&cloud.ShareEditSpec.TargetSpec.Size, "new-size", 0, "New share size in GB")
	addInteractiveEditorFlag(shareEditCmd)
	shareCmd.AddCommand(shareEditCmd)

	shareCmd.AddCommand(&cobra.Command{
		Use:   "delete <share_id>",
		Short: "Delete the given share",
		Run:   cloud.DeleteShare,
		Args:  cobra.ExactArgs(1),
	})

	// ACL subcommands
	aclCmd := &cobra.Command{
		Use:   "acl",
		Short: "Manage share access control lists",
	}
	shareCmd.AddCommand(aclCmd)

	aclListCmd := &cobra.Command{
		Use:     "list <share_id>",
		Aliases: []string{"ls"},
		Short:   "List ACLs for the given share",
		Run:     cloud.ListShareACLs,
		Args:    cobra.ExactArgs(1),
	}
	aclCmd.AddCommand(withFilterFlag(aclListCmd))

	aclCmd.AddCommand(&cobra.Command{
		Use:   "get <share_id> <acl_id>",
		Short: "Get a specific ACL for the given share",
		Run:   cloud.GetShareACL,
		Args:  cobra.ExactArgs(2),
	})

	aclCreateCmd := &cobra.Command{
		Use:   "create <share_id>",
		Short: "Create an ACL for the given share",
		Run:   cloud.CreateShareACL,
		Args:  cobra.ExactArgs(1),
	}
	aclCreateCmd.Flags().StringVar(&cloud.ShareACLSpec.TargetSpec.AccessLevel, "access-level", "", "Access level (READ_ONLY, READ_WRITE)")
	aclCreateCmd.Flags().StringVar(&cloud.ShareACLSpec.TargetSpec.AccessTo, "access-to", "", "Access target (IP address or CIDR)")
	aclCreateCmd.MarkFlagRequired("access-level")
	aclCreateCmd.MarkFlagRequired("access-to")
	aclCmd.AddCommand(aclCreateCmd)

	aclCmd.AddCommand(&cobra.Command{
		Use:   "delete <share_id> <acl_id>",
		Short: "Delete an ACL from the given share",
		Run:   cloud.DeleteShareACL,
		Args:  cobra.ExactArgs(2),
	})

	// Snapshot subcommands
	snapshotCmd := &cobra.Command{
		Use:   "snapshot",
		Short: "Manage share snapshots",
	}
	shareCmd.AddCommand(snapshotCmd)

	snapshotListCmd := &cobra.Command{
		Use:     "list <share_id>",
		Aliases: []string{"ls"},
		Short:   "List snapshots for the given share",
		Run:     cloud.ListShareSnapshots,
		Args:    cobra.ExactArgs(1),
	}
	snapshotCmd.AddCommand(withFilterFlag(snapshotListCmd))

	snapshotCmd.AddCommand(&cobra.Command{
		Use:   "get <share_id> <snapshot_id>",
		Short: "Get a specific snapshot for the given share",
		Run:   cloud.GetShareSnapshot,
		Args:  cobra.ExactArgs(2),
	})

	snapshotCreateCmd := &cobra.Command{
		Use:   "create <share_id>",
		Short: "Create a snapshot of the given share",
		Run:   cloud.CreateShareSnapshot,
		Args:  cobra.ExactArgs(1),
	}
	snapshotCreateCmd.Flags().StringVar(&cloud.ShareSnapshotSpec.Description, "description", "", "Snapshot description")
	snapshotCreateCmd.Flags().StringVar(&cloud.ShareSnapshotSpec.Name, "name", "", "Snapshot name")
	snapshotCmd.AddCommand(snapshotCreateCmd)

	snapshotCmd.AddCommand(&cobra.Command{
		Use:   "delete <share_id> <snapshot_id>",
		Short: "Delete a snapshot from the given share",
		Run:   cloud.DeleteShareSnapshot,
		Args:  cobra.ExactArgs(2),
	})

	cloudCmd.AddCommand(storageFileCmd)
}

func getShareCreateCmd() *cobra.Command {
	shareCreateCmd := &cobra.Command{
		Use:   "create <region>",
		Short: "Create a new share",
		Run:   cloud.CreateShare,
		Args:  cobra.ExactArgs(1),
	}
	shareCreateCmd.Flags().StringVar(&cloud.ShareSpec.TargetSpec.Location.AvailabilityZone, "availability-zone", "", "Availability zone (required in 3AZ regions)")
	shareCreateCmd.Flags().StringVar(&cloud.ShareSpec.TargetSpec.Description, "description", "", "Share description")
	shareCreateCmd.Flags().StringVar(&cloud.ShareSpec.TargetSpec.Name, "name", "", "Share name")
	shareCreateCmd.Flags().StringVar(&cloud.ShareSpec.TargetSpec.ShareNetwork.Id, "share-network-id", "", "Share network ID")
	shareCreateCmd.Flags().IntVar(&cloud.ShareSpec.TargetSpec.Size, "size", 0, "Share size in GB")
	shareCreateCmd.Flags().StringVar(&cloud.ShareSpec.TargetSpec.Protocol, "protocol", "NFS", "Share protocol")
	shareCreateCmd.Flags().StringVar(&cloud.ShareSpec.TargetSpec.ShareType, "share-type", "STANDARD_1AZ", "Share type")

	addParameterFileFlags(shareCreateCmd, false, assets.CloudV2OpenapiSchema, "/publicCloud/project/{projectId}/storage/file/share", "post", cloud.ShareCreateExample, nil)
	addInteractiveEditorFlag(shareCreateCmd)
	markFlagsMutuallyExclusive(shareCreateCmd, "from-file", "editor")

	return shareCreateCmd
}
