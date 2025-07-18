package cmd

import (
	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/services/cloud"
)

func initCloudRegionCommand(cloudCmd *cobra.Command) {
	regionCmd := &cobra.Command{
		Use:   "region",
		Short: "Check regions in the given cloud project",
	}
	regionCmd.PersistentFlags().StringVar(&cloud.CloudProject, "cloud-project", "", "Cloud project ID")

	regionListCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List regions",
		Run:     cloud.ListCloudRegions,
	}
	regionCmd.AddCommand(withFilterFlag(regionListCmd))

	regionCmd.AddCommand(&cobra.Command{
		Use:   "get <region>",
		Short: "Get information about a region",
		Run:   cloud.GetCloudRegion,
		Args:  cobra.ExactArgs(1),
	})

	cloudCmd.AddCommand(regionCmd)
}
