package cmd

import (
	_ "embed"
	"fmt"
	"net/url"

	"github.com/spf13/cobra"
)

var (
	cloudprojectLoadbalancerColumnsToDisplay = []string{"id", "openstackRegion", "size", "status"}

	//go:embed templates/cloud_loadbalancer.tmpl
	cloudLoadbalancerTemplate string
)

func listCloudLoadbalancers(_ *cobra.Command, _ []string) {
	projectID := url.PathEscape(getConfiguredCloudProject())
	manageListRequest(fmt.Sprintf("/cloud/project/%s/loadbalancer", projectID), "", cloudprojectLoadbalancerColumnsToDisplay, genericFilters)
}

func getCloudLoadbalancer(_ *cobra.Command, args []string) {
	projectID := url.PathEscape(getConfiguredCloudProject())
	manageObjectRequest(fmt.Sprintf("/cloud/project/%s/loadbalancer", projectID), args[0], cloudLoadbalancerTemplate)
}

func initCloudLoadbalancerCommand(cloudCmd *cobra.Command) {
	loadbalancerCmd := &cobra.Command{
		Use:   "loadbalancer",
		Short: "Manage loadbalancers in the given cloud project",
	}
	loadbalancerCmd.PersistentFlags().StringVar(&cloudProject, "cloud-project", "", "Cloud project ID")

	loadbalancerListCmd := &cobra.Command{
		Use:   "list",
		Short: "List your loadbalancers",
		Run:   listCloudLoadbalancers,
	}
	loadbalancerCmd.AddCommand(withFilterFlag(loadbalancerListCmd))

	loadbalancerCmd.AddCommand(&cobra.Command{
		Use:        "get",
		Short:      "Get a specific loadbalancer",
		Run:        getCloudLoadbalancer,
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"loadbalancer_id"},
	})

	cloudCmd.AddCommand(loadbalancerCmd)
}
