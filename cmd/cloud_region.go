package cmd

import (
	_ "embed"
	"fmt"
	"net/url"

	"github.com/spf13/cobra"
)

var (
	cloudprojectRegionColumnsToDisplay = []string{"name", "type", "status"}

	//go:embed templates/cloud_region.tmpl
	cloudRegionTemplate string
)

func listCloudRegions(_ *cobra.Command, _ []string) {
	projectID := url.PathEscape(getConfiguredCloudProject())
	manageListRequest(fmt.Sprintf("/cloud/project/%s/region", projectID), "", cloudprojectRegionColumnsToDisplay, genericFilters)
}

func getCloudRegion(_ *cobra.Command, args []string) {
	projectID := url.PathEscape(getConfiguredCloudProject())
	manageObjectRequest(fmt.Sprintf("/cloud/project/%s/region", projectID), args[0], cloudRegionTemplate)
}

func initCloudRegionCommand(cloudCmd *cobra.Command) {
	regionCmd := &cobra.Command{
		Use:   "region",
		Short: "Check regions in the given cloud project",
	}
	regionCmd.PersistentFlags().StringVar(&cloudProject, "cloud-project", "", "Cloud project ID")

	regionListCmd := &cobra.Command{
		Use:   "list",
		Short: "List regions",
		Run:   listCloudRegions,
	}
	regionCmd.AddCommand(withFilterFlag(regionListCmd))

	regionCmd.AddCommand(&cobra.Command{
		Use:        "get",
		Short:      "Get information about a region",
		Run:        getCloudRegion,
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"region"},
	})

	cloudCmd.AddCommand(regionCmd)
}
