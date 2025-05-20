package cmd

import (
	_ "embed"

	"github.com/spf13/cobra"
)

var (
	overtheboxColumnsToDisplay = []string{"serviceName", "offer", "status", "bandwidth"}

	//go:embed templates/overthebox.tmpl
	overtheboxTemplate string
)

func listOverTheBox(_ *cobra.Command, _ []string) {
	manageListRequest("/overTheBox", "", overtheboxColumnsToDisplay, genericFilters)
}

func getOverTheBox(_ *cobra.Command, args []string) {
	manageObjectRequest("/overTheBox", args[0], overtheboxTemplate)
}

func init() {
	overtheboxCmd := &cobra.Command{
		Use:   "overthebox",
		Short: "Retrieve information and manage your OverTheBox services",
	}

	// Command to list OverTheBox services
	overtheboxListCmd := &cobra.Command{
		Use:   "list",
		Short: "List your OverTheBox services",
		Run:   listOverTheBox,
	}
	overtheboxCmd.AddCommand(withFilterFlag(overtheboxListCmd))

	// Command to get a single OverTheBox
	overtheboxCmd.AddCommand(&cobra.Command{
		Use:   "get <service_name>",
		Short: "Retrieve information of a specific OverTheBox",
		Args:  cobra.ExactArgs(1),
		Run:   getOverTheBox,
	})

	rootCmd.AddCommand(overtheboxCmd)
}
