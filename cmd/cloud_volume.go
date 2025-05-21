package cmd

import (
	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/services/cloud"
)

func initCloudVolumeCommand(cloudCmd *cobra.Command) {
	volumeCmd := &cobra.Command{
		Use:   "volume",
		Short: "Manage volumes in the given cloud project",
	}
	volumeCmd.PersistentFlags().StringVar(&cloud.CloudProject, "cloud-project", "", "Cloud project ID")

	volumeListCmd := &cobra.Command{
		Use:   "list",
		Short: "List volumes",
		Run:   cloud.ListCloudVolumes,
	}
	volumeCmd.AddCommand(withFilterFlag(volumeListCmd))

	volumeCmd.AddCommand(&cobra.Command{
		Use:   "get <volume_id>",
		Short: "Get a specific volume",
		Run:   cloud.GetVolume,
		Args:  cobra.ExactArgs(1),
	})

	cloudCmd.AddCommand(volumeCmd)
}
