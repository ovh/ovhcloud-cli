// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/ovh/ovhcloud-cli/internal/assets"
	"github.com/ovh/ovhcloud-cli/internal/completion"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	"github.com/ovh/ovhcloud-cli/internal/services/cloud"
	"github.com/spf13/cobra"
)

func initCloudVolumeCommand(cloudCmd *cobra.Command) {
	storageBlockCmd := &cobra.Command{
		Use:   "block",
		Short: "Manage block storage volumes in the given cloud project",
	}
	storageBlockCmd.PersistentFlags().StringVar(&cloud.CloudProject, "cloud-project", "", "Cloud project ID")

	volumeListCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List volumes",
		Run:     cloud.ListCloudVolumes,
	}
	storageBlockCmd.AddCommand(withFilterFlag(volumeListCmd))

	storageBlockCmd.AddCommand(&cobra.Command{
		Use:               "get <volume_id>",
		Short:             "Get a specific volume",
		ValidArgsFunction: completion.CloudResources("/v1/cloud/project/%s/volume"),
		Run:               cloud.GetVolume,
		Args:              cobra.ExactArgs(1),
	})

	volumeEditCmd := &cobra.Command{
		Use:   "edit <volume_id>",
		Short: "Edit the given volume",
		Example: `  # Rename a volume and grow it to 40 GB
  ovhcloud cloud storage block edit <volume_id> --cloud-project <project_id> --name backups --size 40`,
		ValidArgsFunction: completion.CloudResources("/v1/cloud/project/%s/volume"),
		Run:               cloud.EditVolume,
		Args:              cobra.ExactArgs(1),
	}
	volumeEditCmd.Flags().StringVar(&cloud.VolumeEditSpec.TargetSpec.Name, "name", "", "Volume name")
	volumeEditCmd.Flags().IntVar(&cloud.VolumeEditSpec.TargetSpec.Size, "size", 0, "Volume size (in GB, can only be increased)")
	volumeEditCmd.Flags().StringVar(&cloud.VolumeEditSpec.TargetSpec.VolumeType, "type", "", "Volume type (CLASSIC, HIGH_SPEED, HIGH_SPEED_GEN2)")
	registerFlagValueCompletion(volumeEditCmd, "type", "CLASSIC", "HIGH_SPEED", "HIGH_SPEED_GEN2")
	volumeEditCmd.Flags().BoolVar(&flags.WaitForTask, "wait", false, "Wait for the volume to be READY before exiting")
	addInteractiveEditorFlag(volumeEditCmd)
	storageBlockCmd.AddCommand(volumeEditCmd)

	storageBlockCmd.AddCommand(getVolumeCreateCmd())

	storageBlockCmd.AddCommand(&cobra.Command{
		Use:               "delete <volume_id>",
		Short:             "Delete the given volume",
		ValidArgsFunction: completion.CloudResources("/v1/cloud/project/%s/volume"),
		Run:               cloud.DeleteVolume,
		Args:              cobra.ExactArgs(1),
	})

	// Volume action commands
	storageBlockCmd.AddCommand(&cobra.Command{
		Use:               "attach <volume_id> <instance_id>",
		Short:             "Attach the given volume to the given instance",
		ValidArgsFunction: completion.CloudResources("/v1/cloud/project/%s/volume"),
		Run:               cloud.AttachVolumeToInstance,
		Args:              cobra.ExactArgs(2),
	})

	storageBlockCmd.AddCommand(&cobra.Command{
		Use:               "detach <volume_id> <instance_id>",
		Short:             "Detach the given volume from the given instance",
		ValidArgsFunction: completion.CloudResources("/v1/cloud/project/%s/volume"),
		Run:               cloud.DetachVolumeFromInstance,
		Args:              cobra.ExactArgs(2),
	})

	// Volume snapshot commands
	volumeSnapshotCmd := &cobra.Command{
		Use:   "snapshot",
		Short: "Manage snapshots of the given volume",
	}
	storageBlockCmd.AddCommand(volumeSnapshotCmd)

	volumeSnapshotCreateCmd := &cobra.Command{
		Use:               "create <volume_id>",
		Short:             "Create a snapshot of the given volume",
		ValidArgsFunction: completion.CloudResources("/v1/cloud/project/%s/volume"),
		Run:               cloud.CreateVolumeSnapshot,
		Args:              cobra.ExactArgs(1),
	}
	volumeSnapshotCreateCmd.Flags().StringVar(&cloud.VolumeSnapShotSpec.Description, "description", "", "Snapshot description")
	volumeSnapshotCreateCmd.Flags().StringVar(&cloud.VolumeSnapShotSpec.Name, "name", "", "Snapshot name")
	volumeSnapshotCreateCmd.Flags().BoolVar(&flags.WaitForTask, "wait", false, "Wait for the snapshot to be READY before exiting")
	volumeSnapshotCmd.AddCommand(volumeSnapshotCreateCmd)

	volumeSnapshotListCmd := &cobra.Command{
		Use:     "list <volume_id (optional, list all snapshots if omitted)>",
		Short:   "List snapshots of the given volume",
		Aliases: []string{"ls"},
		Run:     cloud.ListVolumeSnapshots,
		Args:    cobra.NoArgs,
	}
	volumeSnapshotListCmd.Flags().String("volume-id", "", "Volume ID to filter snapshots by")
	volumeSnapshotCmd.AddCommand(volumeSnapshotListCmd)

	volumeSnapshotCmd.AddCommand(&cobra.Command{
		Use:   "delete <snapshot_id>",
		Short: "Delete the given snapshot",
		Run:   cloud.DeleteVolumeSnapshot,
		Args:  cobra.ExactArgs(1),
	})

	// Volume backup commands
	volumeBackupCmd := &cobra.Command{
		Use:   "backup",
		Short: "Manage volume backups in the given cloud project",
	}
	storageBlockCmd.AddCommand(volumeBackupCmd)

	volumeBackupCmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List volume backups",
		Run:     cloud.ListVolumeBackups,
	}))

	volumeBackupCmd.AddCommand(&cobra.Command{
		Use:   "get <backup_id>",
		Short: "Get a specific volume backup",
		Run:   cloud.GetVolumeBackup,
		Args:  cobra.ExactArgs(1),
	})

	volumeBackupCreateCmd := &cobra.Command{
		Use:               "create <volume_id> <backup_name>",
		Short:             "Create a backup of the given volume",
		ValidArgsFunction: completion.CloudResources("/v1/cloud/project/%s/volume"),
		Run:               cloud.CreateVolumeBackup,
		Args:              cobra.ExactArgs(2),
	}
	volumeBackupCreateCmd.Flags().BoolVar(&flags.WaitForTask, "wait", false, "Wait for the backup to be READY before exiting")
	volumeBackupCmd.AddCommand(volumeBackupCreateCmd)

	volumeBackupCmd.AddCommand(&cobra.Command{
		Use:   "delete <backup_id>",
		Short: "Delete the given volume backup",
		Run:   cloud.DeleteVolumeBackup,
		Args:  cobra.ExactArgs(1),
	})

	volumeBackupCmd.AddCommand(&cobra.Command{
		Use:   "restore <backup_id> <volume_id>",
		Short: "Restore a volume from the given backup",
		Run:   cloud.RestoreVolumeBackup,
		Args:  cobra.ExactArgs(2),
	})

	cloudCmd.AddCommand(storageBlockCmd)
}

func getVolumeCreateCmd() *cobra.Command {
	volumeCreateCmd := &cobra.Command{
		Use:   "create <region>",
		Short: "Create a new volume",
		Example: `  # Create a 20 GB high-speed volume in GRA11 and wait until it is ready
  ovhcloud cloud storage block create GRA11 --cloud-project <project_id> --name data --size 20 --type HIGH_SPEED --wait`,
		Run:  cloud.CreateVolume,
		Args: cobra.ExactArgs(1),
	}
	volumeCreateCmd.Flags().StringVar(&cloud.VolumeSpec.TargetSpec.Location.AvailabilityZone, "availability-zone", "", "Availability zone of the volume")
	volumeCreateCmd.Flags().StringVar(&cloud.VolumeSpec.TargetSpec.CreateFrom.BackupId, "backup-id", "", "Backup ID to create the volume from")
	volumeCreateCmd.Flags().StringVar(&cloud.VolumeSpec.TargetSpec.CreateFrom.ImageId, "image-id", "", "Image ID to create the volume from")
	volumeCreateCmd.Flags().StringVar(&cloud.VolumeSpec.TargetSpec.Name, "name", "", "Volume name")
	volumeCreateCmd.Flags().IntVar(&cloud.VolumeSpec.TargetSpec.Size, "size", 0, "Volume size (in GB)")
	volumeCreateCmd.Flags().StringVar(&cloud.VolumeSpec.TargetSpec.CreateFrom.SnapshotId, "snapshot-id", "", "Snapshot ID to create the volume from")
	volumeCreateCmd.Flags().StringVar(&cloud.VolumeSpec.TargetSpec.VolumeType, "type", "", "Volume type (CLASSIC, HIGH_SPEED, HIGH_SPEED_GEN2)")
	registerFlagValueCompletion(volumeCreateCmd, "type", "CLASSIC", "HIGH_SPEED", "HIGH_SPEED_GEN2")

	addParameterFileFlags(volumeCreateCmd, false, assets.CloudV2OpenapiSchema, "/publicCloud/project/{projectId}/storage/block/volume", "post", cloud.VolumeCreateExample, nil)
	addInteractiveEditorFlag(volumeCreateCmd)
	volumeCreateCmd.Flags().BoolVar(&flags.WaitForTask, "wait", false, "Wait for volume creation to be done before exiting")
	markFlagsMutuallyExclusive(volumeCreateCmd, "from-file", "editor")

	return volumeCreateCmd
}
