package cmd

import (
	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/flags"
	"stash.ovh.net/api/ovh-cli/internal/services/cloud"
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
	volumeEditCmd.Flags().StringSliceVar(&cloud.CloudVolume.AttachedTo, "attached-to", nil, "Volume attached to instances id")
	volumeEditCmd.Flags().StringVar(&cloud.CloudVolume.AvailabilityZone, "availability-zone", "", "Availability zone of the volume")
	volumeEditCmd.Flags().BoolVar(&cloud.CloudVolume.Bootable, "bootable", false, "Volume bootable")
	volumeEditCmd.Flags().StringVar(&cloud.CloudVolume.CreationDate, "creation-date", "", "Volume creation date")
	volumeEditCmd.Flags().StringVar(&cloud.CloudVolume.Description, "description", "", "Volume description")
	volumeEditCmd.Flags().StringVar(&cloud.CloudVolume.Name, "name", "", "Volume name")
	volumeEditCmd.Flags().StringVar(&cloud.CloudVolume.PlanCode, "plan-code", "", "Order plan code")
	volumeEditCmd.Flags().StringVar(&cloud.CloudVolume.Region, "region", "", "Volume region")
	volumeEditCmd.Flags().IntVar(&cloud.CloudVolume.Size, "size", 0, "Volume size (in GB)")
	volumeEditCmd.Flags().StringVar(&cloud.CloudVolume.Status, "status", "", "Volume status (attaching, available, awaiting-transfer, backing-up, creating, deleting, detaching, downloading, error, error_backing-up, error_deleting, error_extending, error_restoring, extending, in-use, maintenance, reserved, restoring-backup, retyping, uploading)")
	volumeEditCmd.Flags().StringVar(&cloud.CloudVolume.Type, "type", "", "Volume type (classic, classic-luks, classic-multiattach, high-speed, high-speed-gen2, high-speed-gen2-luks, high-speed-luks)")
	volumeEditCmd.Flags().BoolVar(&flags.ParametersViaEditor, "editor", false, "Use a text editor to define parameters")
	volumeCmd.AddCommand(volumeEditCmd)

	cloudCmd.AddCommand(volumeCmd)
}
