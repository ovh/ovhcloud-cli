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
		Use:   "storage-file",
		Short: "Manage file storage shares in the given cloud project",
	}
	storageFileCmd.PersistentFlags().StringVar(&cloud.CloudProject, "cloud-project", "", "Cloud project ID")
	storageFileCmd.PersistentFlags().StringVar(&cloud.ShareRegion, "region", "", "Region (skip region discovery if set)")

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
	shareEditCmd.Flags().StringVar(&cloud.ShareEditSpec.Description, "description", "", "Share description")
	shareEditCmd.Flags().StringVar(&cloud.ShareEditSpec.Name, "name", "", "Share name")
	shareEditCmd.Flags().IntVar(&cloud.ShareEditSpec.NewSize, "new-size", 0, "New share size in GB")
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
	aclCreateCmd.Flags().StringVar(&cloud.ShareACLSpec.AccessLevel, "access-level", "", "Access level (ro, rw)")
	aclCreateCmd.Flags().StringVar(&cloud.ShareACLSpec.AccessTo, "access-to", "", "Access target (IP address or CIDR)")
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
	shareCreateCmd.Flags().StringVar(&cloud.ShareSpec.AvailabilityZone, "availability-zone", "", "Availability zone (required in 3AZ regions)")
	shareCreateCmd.Flags().StringVar(&cloud.ShareSpec.Description, "description", "", "Share description")
	shareCreateCmd.Flags().StringVar(&cloud.ShareSpec.Name, "name", "", "Share name")
	shareCreateCmd.Flags().StringVar(&cloud.ShareSpec.NetworkId, "network-id", "", "Network ID")
	shareCreateCmd.Flags().IntVar(&cloud.ShareSpec.Size, "size", 0, "Share size in GB")
	shareCreateCmd.Flags().StringVar(&cloud.ShareSpec.SnapshotId, "snapshot-id", "", "Snapshot ID to create the share from")
	shareCreateCmd.Flags().StringVar(&cloud.ShareSpec.SubnetId, "subnet-id", "", "Subnet ID")
	shareCreateCmd.Flags().StringVar(&cloud.ShareSpec.Type, "type", "", "Share type")

	addParameterFileFlags(shareCreateCmd, false, assets.CloudOpenapiSchema, "/cloud/project/{serviceName}/region/{regionName}/share", "post", cloud.ShareCreateExample, nil)
	addInteractiveEditorFlag(shareCreateCmd)
	shareCreateCmd.MarkFlagsMutuallyExclusive("from-file", "editor")

	return shareCreateCmd
}
