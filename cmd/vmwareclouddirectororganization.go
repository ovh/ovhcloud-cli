package cmd

import (
	_ "embed"

	"github.com/spf13/cobra"
)

var (
	vmwareclouddirectororganizationColumnsToDisplay = []string{"id", "currentState.fullName", "currentState.region", "resourceStatus"}

	//go:embed templates/vmwareclouddirectororganization.tmpl
	vmwareclouddirectororganizationTemplate string
)

func listVmwareCloudDirectorOrganization(_ *cobra.Command, _ []string) {
	manageListRequest("/v2/vmwareCloudDirector/organization", "id", vmwareclouddirectororganizationColumnsToDisplay, genericFilters)
}

func getVmwareCloudDirectorOrganization(_ *cobra.Command, args []string) {
	manageObjectRequest("/v2/vmwareCloudDirector/organization", args[0], vmwareclouddirectororganizationTemplate)
}

func init() {
	vmwareclouddirectororganizationCmd := &cobra.Command{
		Use:   "vmwareclouddirector-organization",
		Short: "Retrieve information and manage your VmwareCloudDirector Organizations",
	}

	// Command to list VmwareCloudDirectorOrganization services
	vmwareclouddirectororganizationListCmd := &cobra.Command{
		Use:   "list",
		Short: "List your VmwareCloudDirector Organizations",
		Run:   listVmwareCloudDirectorOrganization,
	}
	vmwareclouddirectororganizationCmd.AddCommand(withFilterFlag(vmwareclouddirectororganizationListCmd))

	// Command to get a single VmwareCloudDirectorOrganization
	vmwareclouddirectororganizationCmd.AddCommand(&cobra.Command{
		Use:   "get <service_name>",
		Short: "Retrieve information of a specific VmwareCloudDirector Organization",
		Args:  cobra.ExactArgs(1),
		Run:   getVmwareCloudDirectorOrganization,
	})

	rootCmd.AddCommand(vmwareclouddirectororganizationCmd)
}
