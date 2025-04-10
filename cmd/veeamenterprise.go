package cmd

import (
	"github.com/spf13/cobra"
)

var (
	veeamenterpriseColumnsToDisplay = []string{ "serviceName","activationStatus","ip","sourceIp" }
)

func listVeeamEnterprise(_ *cobra.Command, _ []string) {
	manageListRequest("/veeam/veeamEnterprise", veeamenterpriseColumnsToDisplay, genericFilters)
}

func getVeeamEnterprise(_ *cobra.Command, args []string) {
	manageObjectRequest("/veeam/veeamEnterprise", args[0], veeamenterpriseColumnsToDisplay[0])
}

func init() {
	veeamenterpriseCmd := &cobra.Command{
		Use:   "veeamenterprise",
		Short: "Retrieve information and manage your VeeamEnterprise services",
	}

	// Command to list VeeamEnterprise services
	veeamenterpriseListCmd := &cobra.Command{
		Use:   "list",
		Short: "List your VeeamEnterprise services",
		Run:   listVeeamEnterprise,
	}
	veeamenterpriseListCmd.PersistentFlags().StringArrayVar(
		&genericFilters,
		"filter",
		nil,
		"Filter results by any property using github.com/PaesslerAG/gval syntax'",
	)
	veeamenterpriseCmd.AddCommand(veeamenterpriseListCmd)

	// Command to get a single VeeamEnterprise
	veeamenterpriseCmd.AddCommand(&cobra.Command{
		Use:        "get",
		Short:      "Retrieve information of a specific VeeamEnterprise",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"service_name"},
		Run:        getVeeamEnterprise,
	})

	rootCmd.AddCommand(veeamenterpriseCmd)
}
