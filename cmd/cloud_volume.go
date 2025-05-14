package cmd

import (
	_ "embed"
	"fmt"
	"net/url"

	"github.com/spf13/cobra"
)

var (
	cloudprojectVolumeColumnsToDisplay = []string{"id", "name", "region", "type", "status"}

	//go:embed templates/cloud_volume.tmpl
	cloudVolumeTemplate string
)

func listCloudVolumes(_ *cobra.Command, _ []string) {
	projectID := url.PathEscape(getConfiguredCloudProject())
	manageListRequestNoExpand(fmt.Sprintf("/cloud/project/%s/volume", projectID), cloudprojectVolumeColumnsToDisplay, genericFilters)
}

func getVolume(_ *cobra.Command, args []string) {
	projectID := url.PathEscape(getConfiguredCloudProject())
	manageObjectRequest(fmt.Sprintf("/cloud/project/%s/volume", projectID), args[0], cloudVolumeTemplate)
}

func initCloudVolumeCommand(cloudCmd *cobra.Command) {
	volumeCmd := &cobra.Command{
		Use:   "volume",
		Short: "Manage volumes in the given cloud project",
	}
	volumeCmd.PersistentFlags().StringVar(&cloudProject, "cloud-project", "", "Cloud project ID")

	volumeListCmd := &cobra.Command{
		Use:   "list",
		Short: "List volumes",
		Run:   listCloudVolumes,
	}
	volumeCmd.AddCommand(withFilterFlag(volumeListCmd))

	volumeCmd.AddCommand(&cobra.Command{
		Use:   "get <volume_id>",
		Short: "Get a specific volume",
		Run:   getVolume,
		Args:  cobra.ExactArgs(1),
	})

	cloudCmd.AddCommand(volumeCmd)
}
