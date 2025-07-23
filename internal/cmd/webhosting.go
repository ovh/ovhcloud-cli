package cmd

import (
	"github.com/ovh/ovhcloud-cli/internal/services/webhosting"
	"github.com/spf13/cobra"
)

func init() {
	webhostingCmd := &cobra.Command{
		Use:   "webhosting",
		Short: "Retrieve information and manage your WebHosting services",
	}

	// Command to list WebHosting services
	webhostingListCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List your WebHosting services",
		Run:     webhosting.ListWebHosting,
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
	webhostingEditCmd := &cobra.Command{
		Use:   "edit <service_name>",
		Short: "Edit the given WebHosting",
		Args:  cobra.ExactArgs(1),
		Run:   webhosting.EditWebHosting,
	}
	webhostingEditCmd.Flags().StringVar(&webhosting.WebHostingSpec.DisplayName, "display-name", "", "Display name of the WebHosting")
	addInteractiveEditorFlag(webhostingEditCmd)
	webhostingCmd.AddCommand(webhostingEditCmd)

	rootCmd.AddCommand(webhostingCmd)
}
