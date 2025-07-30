package cmd

import (
	"github.com/ovh/ovhcloud-cli/internal/assets"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	"github.com/ovh/ovhcloud-cli/internal/services/cloud"
	"github.com/spf13/cobra"
)

func initCloudVolumeCommand(cloudCmd *cobra.Command) {
	volumeCmd := &cobra.Command{
		Use:   "volume",
		Short: "Manage volumes in the given cloud project",
	}
	volumeCmd.PersistentFlags().StringVar(&cloud.CloudProject, "cloud-project", "", "Cloud project ID")

	volumeListCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List volumes",
		Run:     cloud.ListCloudVolumes,
	}
	volumeCmd.AddCommand(withFilterFlag(volumeListCmd))

	volumeCmd.AddCommand(&cobra.Command{
		Use:   "get <volume_id>",
		Short: "Get a specific volume",
		Run:   cloud.GetVolume,
		Args:  cobra.ExactArgs(1),
	})

	volumeEditCmd := &cobra.Command{
		Use:   "edit <volume_id>",
		Short: "Edit the given volume",
		Run:   cloud.EditVolume,
		Args:  cobra.ExactArgs(1),
	}
	volumeEditCmd.Flags().StringVar(&cloud.VolumeSpec.Description, "description", "", "Volume description")
	volumeEditCmd.Flags().StringVar(&cloud.VolumeSpec.Name, "name", "", "Volume name")
	addInteractiveEditorFlag(volumeEditCmd)
	volumeCmd.AddCommand(volumeEditCmd)

	volumeCmd.AddCommand(getVolumeCreateCmd())

	volumeCmd.AddCommand(&cobra.Command{
		Use:   "delete <volume_id>",
		Short: "Delete the given volume",
		Run:   cloud.DeleteVolume,
		Args:  cobra.ExactArgs(1),
	})

	// Volume action commands
	volumeCmd.AddCommand(&cobra.Command{
		Use:   "attach <volume_id> <instance_id>",
		Short: "Attach the given volume to the given instance",
		Run:   cloud.AttachVolumeToInstance,
		Args:  cobra.ExactArgs(2),
	})

	volumeCmd.AddCommand(&cobra.Command{
		Use:   "detach <volume_id> <instance_id>",
		Short: "Detach the given volume from the given instance",
		Run:   cloud.DetachVolumeFromInstance,
		Args:  cobra.ExactArgs(2),
	})

	volumeCmd.AddCommand(&cobra.Command{
		Use:   "upsize <volume_id> <new_size (GB)>",
		Short: "Upsize the given volume",
		Run:   cloud.UpsizeVolume,
		Args:  cobra.ExactArgs(2),
	})

	// Volume snapshot commands
	volumeSnapshotCmd := &cobra.Command{
		Use:   "snapshot",
		Short: "Manage snapshots of the given volume",
	}
	volumeCmd.AddCommand(volumeSnapshotCmd)

	volumeSnapshotCreateCmd := &cobra.Command{
		Use:   "create <volume_id>",
		Short: "Create a snapshot of the given volume",
		Run:   cloud.CreateVolumeSnapshot,
		Args:  cobra.ExactArgs(1),
	}
	volumeSnapshotCreateCmd.Flags().StringVar(&cloud.VolumeSnapShotSpec.Description, "description", "", "Snapshot description")
	volumeSnapshotCreateCmd.Flags().StringVar(&cloud.VolumeSnapShotSpec.Name, "name", "", "Snapshot name")
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
	volumeCmd.AddCommand(volumeBackupCmd)

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

	volumeBackupCmd.AddCommand(&cobra.Command{
		Use:   "create <volume_id> <backup_name>",
		Short: "Create a backup of the given volume",
		Run:   cloud.CreateVolumeBackup,
		Args:  cobra.ExactArgs(2),
	})

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

	volumeCmd.AddCommand(&cobra.Command{
		Use:   "create-from-backup <backup_id> <volume_name>",
		Short: "Create a volume from the given backup",
		Run:   cloud.CreateVolumeFromBackup,
		Args:  cobra.ExactArgs(2),
	})

	cloudCmd.AddCommand(volumeCmd)
}

func getVolumeCreateCmd() *cobra.Command {
	volumeCreateCmd := &cobra.Command{
		Use:   "create <region>",
		Short: "Create a new volume",
		Run:   cloud.CreateVolume,
		Args:  cobra.ExactArgs(1),
	}
	volumeCreateCmd.Flags().StringVar(&cloud.VolumeSpec.AvailabilityZone, "availability-zone", "", "Availability zone of the volume")
	volumeCreateCmd.Flags().StringVar(&cloud.VolumeSpec.BackupId, "backup-id", "", "Backup ID")
	volumeCreateCmd.Flags().StringVar(&cloud.VolumeSpec.Description, "description", "", "Volume description")
	volumeCreateCmd.Flags().StringVar(&cloud.VolumeSpec.ImageId, "image-id", "", "Image ID to create the volume from")
	volumeCreateCmd.Flags().StringVar(&cloud.VolumeSpec.InstanceId, "instance-id", "", "Instance ID to attach the volume to")
	volumeCreateCmd.Flags().StringVar(&cloud.VolumeSpec.Name, "name", "", "Volume name")
	volumeCreateCmd.Flags().IntVar(&cloud.VolumeSpec.Size, "size", 0, "Volume size (in GB)")
	volumeCreateCmd.Flags().StringVar(&cloud.VolumeSpec.SnapshotId, "snapshot-id", "", "Snapshot ID to create the volume from")
	volumeCreateCmd.Flags().StringVar(&cloud.VolumeSpec.Type, "type", "", "Volume type (classic, classic-luks, classic-multiattach, high-speed, high-speed-gen2, high-speed-gen2-luks, high-speed-luks)")

	addInitParameterFileFlag(volumeCreateCmd, assets.CloudOpenapiSchema, "/cloud/project/{serviceName}/region/{regionName}/volume", "post", cloud.VolumeCreateExample, nil)
	addInteractiveEditorFlag(volumeCreateCmd)
	addFromFileFlag(volumeCreateCmd)
	volumeCreateCmd.Flags().BoolVar(&flags.WaitForTask, "wait", false, "Wait for volume creation to be done before exiting")
	volumeCreateCmd.MarkFlagsMutuallyExclusive("from-file", "editor")

	return volumeCreateCmd
}
