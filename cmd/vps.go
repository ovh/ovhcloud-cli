package cmd

import (
	_ "embed"

	"github.com/spf13/cobra"
)

var (
	vpsColumnsToDisplay = []string{"name", "displayName", "state", "zone"}

	//go:embed templates/vps.tmpl
	vpsTemplate string
)

func listVps(_ *cobra.Command, _ []string) {
	manageListRequest("/vps", "", vpsColumnsToDisplay, genericFilters)
}

func getVps(_ *cobra.Command, args []string) {
	manageObjectRequest("/vps", args[0], vpsTemplate)
}

func init() {
	vpsCmd := &cobra.Command{
		Use:   "vps",
		Short: "Retrieve information and manage your VPS services",
	}

	// Command to list Vps services
	vpsListCmd := &cobra.Command{
		Use:   "list",
		Short: "List your VPS services",
		Run:   listVps,
	}
	vpsCmd.AddCommand(withFilterFlag(vpsListCmd))

	// Command to get a single Vps
	vpsCmd.AddCommand(&cobra.Command{
		Use:   "get <service_name>",
		Short: "Retrieve information of a specific VPS",
		Args:  cobra.ExactArgs(1),
		Run:   getVps,
	})

	rootCmd.AddCommand(vpsCmd)
}
