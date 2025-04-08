
package cmd

import (
	"github.com/spf13/cobra"
)

var (
	iploadbalancingColumnsToDisplay = []string{ "serviceName","displayName","zone","state" }
)

func listIpLoadbalancing(_ *cobra.Command, _ []string) {
	manageListRequest("/ipLoadbalancing", iploadbalancingColumnsToDisplay)
}

func getIpLoadbalancing(_ *cobra.Command, args []string) {
	manageObjectRequest("/ipLoadbalancing", args[0], iploadbalancingColumnsToDisplay[0])
}

func init() {
	iploadbalancingCmd := &cobra.Command{
		Use:   "iploadbalancing",
		Short: "Retrieve information and manage your IpLoadbalancing services",
	}

	// Command to list IpLoadbalancing services
	iploadbalancingCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List your IpLoadbalancing services",
		Run:   listIpLoadbalancing,
	})

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
