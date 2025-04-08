
package cmd

import (
	"github.com/spf13/cobra"
)

var (
	vmwareclouddirectororganizationColumnsToDisplay = []string{ "id","currentState.fullName","currentState.region","resourceStatus" }
)

func listVmwareCloudDirectorOrganization(_ *cobra.Command, _ []string) {
	manageListRequest("/v2/vmwareCloudDirector/organization", vmwareclouddirectororganizationColumnsToDisplay)
}

func getVmwareCloudDirectorOrganization(_ *cobra.Command, args []string) {
	manageObjectRequest("/v2/vmwareCloudDirector/organization", args[0], vmwareclouddirectororganizationColumnsToDisplay[0])
}

func init() {
	vmwareclouddirectororganizationCmd := &cobra.Command{
		Use:   "vmwareclouddirectororganization",
		Short: "Retrieve information and manage your VmwareCloudDirectorOrganization services",
	}

	// Command to list VmwareCloudDirectorOrganization services
	vmwareclouddirectororganizationCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List your VmwareCloudDirectorOrganization services",
		Run:   listVmwareCloudDirectorOrganization,
	})

	// Command to get a single VmwareCloudDirectorOrganization
	vmwareclouddirectororganizationCmd.AddCommand(&cobra.Command{
		Use:        "get",
		Short:      "Retrieve information of a specific VmwareCloudDirectorOrganization",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"service_name"},
		Run:        getVmwareCloudDirectorOrganization,
	})

	rootCmd.AddCommand(vmwareclouddirectororganizationCmd)
}
