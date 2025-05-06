package cmd

import (
	_ "embed"

	"github.com/spf13/cobra"
)

var (
	iploadbalancingColumnsToDisplay = []string{"serviceName", "displayName", "zone", "state"}

	//go:embed templates/iploadbalancing.tmpl
	iploadbalancingTemplate string
)

func listIpLoadbalancing(_ *cobra.Command, _ []string) {
	manageListRequest("/ipLoadbalancing", "", iploadbalancingColumnsToDisplay, genericFilters)
}

func getIpLoadbalancing(_ *cobra.Command, args []string) {
	manageObjectRequest("/ipLoadbalancing", args[0], iploadbalancingTemplate)
}

func init() {
	iploadbalancingCmd := &cobra.Command{
		Use:   "iploadbalancing",
		Short: "Retrieve information and manage your IpLoadbalancing services",
	}

	// Command to list IpLoadbalancing services
	iploadbalancingListCmd := &cobra.Command{
		Use:   "list",
		Short: "List your IpLoadbalancing services",
		Run:   listIpLoadbalancing,
	}
	iploadbalancingCmd.AddCommand(withFilterFlag(iploadbalancingListCmd))

	// Command to get a single IpLoadbalancing
	iploadbalancingCmd.AddCommand(&cobra.Command{
		Use:        "get",
		Short:      "Retrieve information of a specific IpLoadbalancing",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"service_name"},
		Run:        getIpLoadbalancing,
	})

	rootCmd.AddCommand(iploadbalancingCmd)
}
