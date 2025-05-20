package cmd

import (
	_ "embed"

	"github.com/spf13/cobra"
)

var (
	vrackColumnsToDisplay = []string{"serviceName", "name", "description"}

	//go:embed templates/vrack.tmpl
	vrackTemplate string
)

func listVrack(_ *cobra.Command, _ []string) {
	manageListRequest("/vrack", "", vrackColumnsToDisplay, genericFilters)
}

func getVrack(_ *cobra.Command, args []string) {
	manageObjectRequest("/vrack", args[0], vrackTemplate)
}

func init() {
	vrackCmd := &cobra.Command{
		Use:   "vrack",
		Short: "Retrieve information and manage your vRack services",
	}

	// Command to list Vrack services
	vrackListCmd := &cobra.Command{
		Use:   "list",
		Short: "List your vRackservices",
		Run:   listVrack,
	}
	vrackCmd.AddCommand(withFilterFlag(vrackListCmd))

	// Command to get a single Vrack
	vrackCmd.AddCommand(&cobra.Command{
		Use:   "get <service_name>",
		Short: "Retrieve information of a specific vRack",
		Args:  cobra.ExactArgs(1),
		Run:   getVrack,
	})

	rootCmd.AddCommand(vrackCmd)
}
