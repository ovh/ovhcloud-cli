package cmd

import (
	"github.com/spf13/cobra"
)

var (
	vrackservicesColumnsToDisplay = []string{"id", "currentState.region", "currentState.productStatus", "resourceStatus"}
)

func listVrackServices(_ *cobra.Command, _ []string) {
	manageListRequest("/v2/vrackServices/resource", vrackservicesColumnsToDisplay)
}

func getVrackServices(_ *cobra.Command, args []string) {
	manageObjectRequest("/v2/vrackServices/resource", args[0], vrackservicesColumnsToDisplay[0])
}

func init() {
	vrackservicesCmd := &cobra.Command{
		Use:   "vrackservices",
		Short: "Retrieve information and manage your VrackServices services",
	}

	// Command to list VrackServices services
	vrackservicesCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List your VrackServices services",
		Run:   listVrackServices,
	})

	// Command to get a single VrackServices
	vrackservicesCmd.AddCommand(&cobra.Command{
		Use:        "get",
		Short:      "Retrieve information of a specific VrackServices",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"service_name"},
		Run:        getVrackServices,
	})

	rootCmd.AddCommand(vrackservicesCmd)
}
