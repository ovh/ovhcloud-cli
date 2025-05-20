package cmd

import (
	_ "embed"

	"github.com/spf13/cobra"
)

var (
	hostingprivatedatabaseColumnsToDisplay = []string{"serviceName", "displayName", "type", "version", "state"}

	//go:embed templates/hostingprivatedatabase.tmpl
	hostingprivatedatabaseTemplate string
)

func listHostingPrivateDatabase(_ *cobra.Command, _ []string) {
	manageListRequest("/hosting/privateDatabase", "", hostingprivatedatabaseColumnsToDisplay, genericFilters)
}

func getHostingPrivateDatabase(_ *cobra.Command, args []string) {
	manageObjectRequest("/hosting/privateDatabase", args[0], hostingprivatedatabaseTemplate)
}

func init() {
	hostingprivatedatabaseCmd := &cobra.Command{
		Use:   "hosting-private-database",
		Short: "Retrieve information and manage your HostingPrivateDatabase services",
	}

	// Command to list HostingPrivateDatabase services
	hostingprivatedatabaseListCmd := &cobra.Command{
		Use:   "list",
		Short: "List your HostingPrivateDatabase services",
		Run:   listHostingPrivateDatabase,
	}
	hostingprivatedatabaseCmd.AddCommand(withFilterFlag(hostingprivatedatabaseListCmd))

	// Command to get a single HostingPrivateDatabase
	hostingprivatedatabaseCmd.AddCommand(&cobra.Command{
		Use:   "get <service_name>",
		Short: "Retrieve information of a specific HostingPrivateDatabase",
		Args:  cobra.ExactArgs(1),
		Run:   getHostingPrivateDatabase,
	})

	rootCmd.AddCommand(hostingprivatedatabaseCmd)
}
