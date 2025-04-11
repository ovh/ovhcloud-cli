package cmd

import (
	_ "embed"

	"github.com/spf13/cobra"
)

var (
	vrackservicesColumnsToDisplay = []string{ "id","currentState.region","currentState. productStatus","resourceStatus" }

	//go:embed templates/vrackservices.tmpl
	vrackservicesTemplate string
)

func listVrackServices(_ *cobra.Command, _ []string) {
	manageListRequest("/v2/vrackServices/resource", vrackservicesColumnsToDisplay, genericFilters)
}

func getVrackServices(_ *cobra.Command, args []string) {
	manageObjectRequest("/v2/vrackServices/resource", args[0], vrackservicesTemplate)
}

func init() {
	vrackservicesCmd := &cobra.Command{
		Use:   "vrackservices",
		Short: "Retrieve information and manage your VrackServices services",
	}

	// Command to list VrackServices services
	vrackservicesListCmd := &cobra.Command{
		Use:   "list",
		Short: "List your VrackServices services",
		Run:   listVrackServices,
	}
	vrackservicesListCmd.PersistentFlags().StringArrayVar(
		&genericFilters,
		"filter",
		nil,
		"Filter results by any property using github.com/PaesslerAG/gval syntax'",
	)
	vrackservicesCmd.AddCommand(vrackservicesListCmd)

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
