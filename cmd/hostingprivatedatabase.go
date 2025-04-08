
package cmd

import (
	"github.com/spf13/cobra"
)

var (
	hostingprivatedatabaseColumnsToDisplay = []string{ "serviceName","displayName","type","version","state" }
)

func listHostingPrivateDatabase(_ *cobra.Command, _ []string) {
	manageListRequest("/hosting/privateDatabase", hostingprivatedatabaseColumnsToDisplay)
}

func getHostingPrivateDatabase(_ *cobra.Command, args []string) {
	manageObjectRequest("/hosting/privateDatabase", args[0], hostingprivatedatabaseColumnsToDisplay[0])
}

func init() {
	hostingprivatedatabaseCmd := &cobra.Command{
		Use:   "hostingprivatedatabase",
		Short: "Retrieve information and manage your HostingPrivateDatabase services",
	}

	// Command to list HostingPrivateDatabase services
	hostingprivatedatabaseCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List your HostingPrivateDatabase services",
		Run:   listHostingPrivateDatabase,
	})

	// Command to get a single HostingPrivateDatabase
	hostingprivatedatabaseCmd.AddCommand(&cobra.Command{
		Use:        "get",
		Short:      "Retrieve information of a specific HostingPrivateDatabase",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"service_name"},
		Run:        getHostingPrivateDatabase,
	})

	rootCmd.AddCommand(hostingprivatedatabaseCmd)
}
