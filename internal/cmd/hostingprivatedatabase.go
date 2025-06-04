package cmd

import (
	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/services/hostingprivatedatabase"
)

func init() {
	hostingprivatedatabaseCmd := &cobra.Command{
		Use:   "hosting-private-database",
		Short: "Retrieve information and manage your HostingPrivateDatabase services",
	}

	// Command to list HostingPrivateDatabase services
	hostingprivatedatabaseListCmd := &cobra.Command{
		Use:   "list",
		Short: "List your HostingPrivateDatabase services",
		Run:   hostingprivatedatabase.ListHostingPrivateDatabase,
	}
	hostingprivatedatabaseCmd.AddCommand(withFilterFlag(hostingprivatedatabaseListCmd))

	// Command to get a single HostingPrivateDatabase
	hostingprivatedatabaseCmd.AddCommand(&cobra.Command{
		Use:   "get <service_name>",
		Short: "Retrieve information of a specific HostingPrivateDatabase",
		Args:  cobra.ExactArgs(1),
		Run:   hostingprivatedatabase.GetHostingPrivateDatabase,
	})

	// Command to update a single HostingPrivateDatabase
	hostingprivatedatabaseCmd.AddCommand(&cobra.Command{
		Use:   "edit <service_name>",
		Short: "Edit the given HostingPrivateDatabase service",
		Run:   hostingprivatedatabase.EditHostingPrivateDatabase,
	})

	rootCmd.AddCommand(hostingprivatedatabaseCmd)
}
