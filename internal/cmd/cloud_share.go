// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/ovh/ovhcloud-cli/internal/assets"
	"github.com/ovh/ovhcloud-cli/internal/services/cloud"
	"github.com/spf13/cobra"
)

func initCloudShareCommand(cloudCmd *cobra.Command) {
	shareCmd := &cobra.Command{
		Use:   "share",
		Short: "Manage shares in the given cloud project",
	}
	shareCmd.PersistentFlags().StringVar(&cloud.CloudProject, "cloud-project", "", "Cloud project ID")

	// Share CRUD commands
	shareListCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List shares",
		Run:     cloud.ListCloudShares,
	}
	shareCmd.AddCommand(withFilterFlag(shareListCmd))

	shareCmd.AddCommand(&cobra.Command{
		Use:   "get <share_id>",
		Short: "Get a specific share",
		Run:   cloud.GetShare,
		Args:  cobra.ExactArgs(1),
	})

	shareEditCmd := &cobra.Command{
		Use:   "edit <share_id>",
		Short: "Edit the given share",
		Run:   cloud.EditShare,
		Args:  cobra.ExactArgs(1),
	}
	shareEditCmd.Flags().StringVar(&cloud.ShareSpec.Description, "description", "", "Share description")
	shareEditCmd.Flags().StringVar(&cloud.ShareSpec.Name, "name", "", "Share name")
	shareEditCmd.Flags().IntVar(&cloud.ShareSpec.NewSize, "new-size", 0, "New share size (in GB)")
	addInteractiveEditorFlag(shareEditCmd)
	shareCmd.AddCommand(shareEditCmd)

	shareCmd.AddCommand(getShareCreateCmd())

	shareCmd.AddCommand(&cobra.Command{
		Use:   "delete <share_id>",
		Short: "Delete the given share",
		Run:   cloud.DeleteShare,
		Args:  cobra.ExactArgs(1),
	})

	// Share ACL commands
	shareAclCmd := &cobra.Command{
		Use:   "acl",
		Short: "Manage ACLs of the given share",
	}
	shareCmd.AddCommand(shareAclCmd)

	shareAclListCmd := &cobra.Command{
		Use:     "list <share_id>",
		Aliases: []string{"ls"},
		Short:   "List ACLs of the given share",
		Run:     cloud.ListShareAcls,
		Args:    cobra.ExactArgs(1),
	}
	shareAclCmd.AddCommand(withFilterFlag(shareAclListCmd))

	shareAclCmd.AddCommand(&cobra.Command{
		Use:   "get <share_id> <acl_id>",
		Short: "Get a specific share ACL",
		Run:   cloud.GetShareAcl,
		Args:  cobra.ExactArgs(2),
	})

	shareAclCmd.AddCommand(getShareAclCreateCmd())

	shareAclCmd.AddCommand(&cobra.Command{
		Use:   "delete <share_id> <acl_id>",
		Short: "Delete the given share ACL",
		Run:   cloud.DeleteShareAcl,
		Args:  cobra.ExactArgs(2),
	})

	// Share Snapshot commands
	shareSnapshotCmd := &cobra.Command{
		Use:   "snapshot",
		Short: "Manage snapshots of the given share",
	}
	shareCmd.AddCommand(shareSnapshotCmd)

	shareSnapshotListCmd := &cobra.Command{
		Use:     "list <share_id>",
		Aliases: []string{"ls"},
		Short:   "List snapshots of the given share",
		Run:     cloud.ListShareSnapshots,
		Args:    cobra.ExactArgs(1),
	}
	shareSnapshotCmd.AddCommand(withFilterFlag(shareSnapshotListCmd))

	shareSnapshotCmd.AddCommand(&cobra.Command{
		Use:   "get <share_id> <snapshot_id>",
		Short: "Get a specific share snapshot",
		Run:   cloud.GetShareSnapshot,
		Args:  cobra.ExactArgs(2),
	})

	shareSnapshotCmd.AddCommand(getShareSnapshotCreateCmd())

	shareSnapshotCmd.AddCommand(&cobra.Command{
		Use:   "delete <share_id> <snapshot_id>",
		Short: "Delete the given share snapshot",
		Run:   cloud.DeleteShareSnapshot,
		Args:  cobra.ExactArgs(2),
	})

	cloudCmd.AddCommand(shareCmd)
}

func getShareCreateCmd() *cobra.Command {
	shareCreateCmd := &cobra.Command{
		Use:   "create <region>",
		Short: "Create a new share",
		Run:   cloud.CreateShare,
		Args:  cobra.ExactArgs(1),
	}
	shareCreateCmd.Flags().StringVar(&cloud.ShareSpec.Description, "description", "", "Share description")
	shareCreateCmd.Flags().StringVar(&cloud.ShareSpec.Name, "name", "", "Share name")
	shareCreateCmd.Flags().StringVar(&cloud.ShareSpec.NetworkId, "network-id", "", "Network ID")
	shareCreateCmd.Flags().IntVar(&cloud.ShareSpec.Size, "size", 0, "Share size (in GB)")
	shareCreateCmd.Flags().StringVar(&cloud.ShareSpec.SnapshotId, "snapshot-id", "", "Snapshot ID to create the share from")
	shareCreateCmd.Flags().StringVar(&cloud.ShareSpec.SubnetId, "subnet-id", "", "Subnet ID")
	shareCreateCmd.Flags().StringVar(&cloud.ShareSpec.Type, "type", "", "Share type (standard-1az)")

	addInitParameterFileFlag(shareCreateCmd, assets.CloudOpenapiSchema, "/cloud/project/{serviceName}/region/{regionName}/share", "post", cloud.ShareCreateExample, nil)
	addInteractiveEditorFlag(shareCreateCmd)
	addFromFileFlag(shareCreateCmd)
	shareCreateCmd.MarkFlagsMutuallyExclusive("from-file", "editor")

	return shareCreateCmd
}

func getShareAclCreateCmd() *cobra.Command {
	shareAclCreateCmd := &cobra.Command{
		Use:   "create <share_id>",
		Short: "Create a new ACL for the given share",
		Run:   cloud.CreateShareAcl,
		Args:  cobra.ExactArgs(1),
	}
	shareAclCreateCmd.Flags().StringVar(&cloud.ShareAclSpec.AccessTo, "access-to", "", "Access to (e.g., IP address or CIDR)")
	shareAclCreateCmd.Flags().StringVar(&cloud.ShareAclSpec.AccessLevel, "access-level", "", "Access level (ro, rw)")

	addInitParameterFileFlag(shareAclCreateCmd, assets.CloudOpenapiSchema, "/cloud/project/{serviceName}/region/{regionName}/share/{shareId}/acl", "post", cloud.ShareAclCreateExample, nil)
	addInteractiveEditorFlag(shareAclCreateCmd)
	addFromFileFlag(shareAclCreateCmd)
	shareAclCreateCmd.MarkFlagsMutuallyExclusive("from-file", "editor")

	return shareAclCreateCmd
}

func getShareSnapshotCreateCmd() *cobra.Command {
	shareSnapshotCreateCmd := &cobra.Command{
		Use:   "create <share_id>",
		Short: "Create a snapshot of the given share",
		Run:   cloud.CreateShareSnapshot,
		Args:  cobra.ExactArgs(1),
	}
	shareSnapshotCreateCmd.Flags().StringVar(&cloud.ShareSnapshotSpec.Description, "description", "", "Snapshot description")
	shareSnapshotCreateCmd.Flags().StringVar(&cloud.ShareSnapshotSpec.Name, "name", "", "Snapshot name")

	addInitParameterFileFlag(shareSnapshotCreateCmd, assets.CloudOpenapiSchema, "/cloud/project/{serviceName}/region/{regionName}/share/{shareId}/snapshot", "post", cloud.ShareSnapshotCreateExample, nil)
	addInteractiveEditorFlag(shareSnapshotCreateCmd)
	addFromFileFlag(shareSnapshotCreateCmd)
	shareSnapshotCreateCmd.MarkFlagsMutuallyExclusive("from-file", "editor")

	return shareSnapshotCreateCmd
}
