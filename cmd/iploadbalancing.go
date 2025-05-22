package cmd

import (
	_ "embed"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/services/iploadbalancing"
)

func init() {
	iploadbalancingCmd := &cobra.Command{
		Use:   "iploadbalancing",
		Short: "Retrieve information and manage your IpLoadbalancing services",
	}

	// Command to list IpLoadbalancing services
	iploadbalancingListCmd := &cobra.Command{
		Use:   "list",
		Short: "List your IpLoadbalancing services",
		Run:   iploadbalancing.ListIpLoadbalancing,
	}
	iploadbalancingCmd.AddCommand(withFilterFlag(iploadbalancingListCmd))

	// Command to get a single IpLoadbalancing
	iploadbalancingCmd.AddCommand(&cobra.Command{
		Use:   "get <service_name>",
		Short: "Retrieve information of a specific IpLoadbalancing",
		Args:  cobra.ExactArgs(1),
		Run:   iploadbalancing.GetIpLoadbalancing,
	})

	// Command to update a single IpLoadbalancing
	iploadbalancingCmd.AddCommand(&cobra.Command{
		Use:   "edit <service_name>",
		Short: "Edit the given IpLoadbalancing",
		Run:   iploadbalancing.EditIpLoadbalancing,
	})

	rootCmd.AddCommand(iploadbalancingCmd)
}
