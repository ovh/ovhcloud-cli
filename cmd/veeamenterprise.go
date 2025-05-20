package cmd

import (
	_ "embed"

	"github.com/spf13/cobra"
)

var (
	veeamenterpriseColumnsToDisplay = []string{"serviceName", "activationStatus", "ip", "sourceIp"}

	//go:embed templates/veeamenterprise.tmpl
	veeamenterpriseTemplate string
)

func listVeeamEnterprise(_ *cobra.Command, _ []string) {
	manageListRequest("/veeam/veeamEnterprise", "", veeamenterpriseColumnsToDisplay, genericFilters)
}

func getVeeamEnterprise(_ *cobra.Command, args []string) {
	manageObjectRequest("/veeam/veeamEnterprise", args[0], veeamenterpriseTemplate)
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
	veeamenterpriseCmd.AddCommand(withFilterFlag(veeamenterpriseListCmd))

	// Command to get a single VeeamEnterprise
	veeamenterpriseCmd.AddCommand(&cobra.Command{
		Use:   "get <service_name>",
		Short: "Retrieve information of a specific VeeamEnterprise",
		Args:  cobra.ExactArgs(1),
		Run:   getVeeamEnterprise,
	})

	rootCmd.AddCommand(veeamenterpriseCmd)
}
