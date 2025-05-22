package cmd

import (
	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/services/webhosting"
)

func init() {
	webhostingCmd := &cobra.Command{
		Use:   "webhosting",
		Short: "Retrieve information and manage your WebHosting services",
	}

	// Command to list WebHosting services
	webhostingListCmd := &cobra.Command{
		Use:   "list",
		Short: "List your WebHosting services",
		Run:   webhosting.ListWebHosting,
	}
	webhostingCmd.AddCommand(withFilterFlag(webhostingListCmd))

	// Command to get a single WebHosting
	webhostingCmd.AddCommand(&cobra.Command{
		Use:   "get <service_name>",
		Short: "Retrieve information of a specific WebHosting",
		Args:  cobra.ExactArgs(1),
		Run:   webhosting.GetWebHosting,
	})

	// Command to update a single WebHosting
	webhostingCmd.AddCommand(&cobra.Command{
		Use:   "edit <service_name>",
		Short: "Edit the given WebHosting",
		Run:   webhosting.EditWebHosting,
	})

	rootCmd.AddCommand(webhostingCmd)
}
