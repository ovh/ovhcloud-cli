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
		Use:   "vmwareclouddirectororganization",
		Short: "Retrieve information and manage your VmwareCloudDirectorOrganization services",
	}

	// Command to list VmwareCloudDirectorOrganization services
	vmwareclouddirectororganizationListCmd := &cobra.Command{
		Use:   "list",
		Short: "List your VmwareCloudDirectorOrganization services",
		Run:   listVmwareCloudDirectorOrganization,
	}
	vmwareclouddirectororganizationListCmd.PersistentFlags().StringArrayVar(
		&genericFilters,
		"filter",
		nil,
		"Filter results by any property using github.com/PaesslerAG/gval syntax'",
	)
	vmwareclouddirectororganizationCmd.AddCommand(vmwareclouddirectororganizationListCmd)

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
