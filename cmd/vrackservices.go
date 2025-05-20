package cmd

import (
	_ "embed"

	"github.com/spf13/cobra"
)

var (
	vrackservicesColumnsToDisplay = []string{"id", "currentState.region", "currentState. productStatus", "resourceStatus"}

	//go:embed templates/vrackservices.tmpl
	vrackservicesTemplate string
)

func listVrackServices(_ *cobra.Command, _ []string) {
	manageListRequest("/v2/vrackServices/resource", "id", vrackservicesColumnsToDisplay, genericFilters)
}

func getVrackServices(_ *cobra.Command, args []string) {
	manageObjectRequest("/v2/vrackServices/resource", args[0], vrackservicesTemplate)
}

func init() {
	vrackservicesCmd := &cobra.Command{
		Use:   "vrackservices",
		Short: "Retrieve information and manage your vRackServices services",
	}

	// Command to list VrackServices services
	vrackservicesListCmd := &cobra.Command{
		Use:   "list",
		Short: "List your vRackServices services",
		Run:   listVrackServices,
	}
	vrackservicesCmd.AddCommand(withFilterFlag(vrackservicesListCmd))

	// Command to get a single VrackServices
	vrackservicesCmd.AddCommand(&cobra.Command{
		Use:   "get <service_name>",
		Short: "Retrieve information of a specific vRackServices",
		Args:  cobra.ExactArgs(1),
		Run:   getVrackServices,
	})

	rootCmd.AddCommand(vrackservicesCmd)
}
