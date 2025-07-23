package cmd

import (
	"github.com/ovh/ovhcloud-cli/internal/services/cloud"
	"github.com/spf13/cobra"
)

func initCloudReferenceCmd(cloudCmd *cobra.Command) {
	referenceCmd := &cobra.Command{
		Use:   "reference",
		Short: "Fetch reference data in the given cloud project",
	}

	referenceCmd.PersistentFlags().StringVar(&cloud.CloudProject, "cloud-project", "", "Cloud project ID")

	var region string
	flavorListCmd := withFilterFlag(&cobra.Command{
		Use:   "list-flavors",
		Short: "List available flavors in the given cloud project",
		Run: func(_ *cobra.Command, _ []string) {
			cloud.GetFlavors(region)
		},
		Args: cobra.NoArgs,
	})
	flavorListCmd.Flags().StringVarP(&region, "region", "r", "", "Region to filter flavors (e.g., GRA9, BHS5)")
	referenceCmd.AddCommand(flavorListCmd)

	var osType string
	imageListCmd := withFilterFlag(&cobra.Command{
		Use:   "list-images",
		Short: "List available images in the given cloud project",
		Run: func(_ *cobra.Command, _ []string) {
			cloud.GetImages(region, osType)
		},
		Args: cobra.NoArgs,
	})
	imageListCmd.Flags().StringVarP(&region, "region", "r", "", "Region to filter images (e.g., GRA9, BHS5)")
	imageListCmd.Flags().StringVarP(&osType, "os-type", "o", "", "OS type to filter images (baremetal-linux, bsd, linux, windows)")
	referenceCmd.AddCommand(imageListCmd)

	cloudCmd.AddCommand(referenceCmd)
}
